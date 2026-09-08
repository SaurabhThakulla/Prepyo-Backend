package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/referrals"
	"github.com/prepyo/backend/internal/users"
)

var (
	ErrEmailTaken = users.ErrEmailTaken
)

type ReferralsService interface {
	LinkReferralOnRegister(ctx context.Context, db database.DB, refereeID, code string) (*models.Referral, error)
}

type Service struct {
	db        *pgxpool.Pool
	users     *users.Repository
	referrals ReferralsService
	sessions  *sessionRepository
	ttl       time.Duration
	log       *slog.Logger
	google    *GoogleVerifier
}

func NewService(db *pgxpool.Pool, userRepo *users.Repository, referrals ReferralsService, ttl time.Duration, log *slog.Logger, google *GoogleVerifier) *Service {
	return &Service{
		db:        db,
		users:     userRepo,
		referrals: referrals,
		sessions:  &sessionRepository{db: db},
		ttl:       ttl,
		log:       log,
		google:    google,
	}
}

// Authenticate resolves a session token to its user.
func (s *Service) Authenticate(ctx context.Context, token string) (models.User, error) {
	hash := hashToken(token)

	userID, err := s.sessions.userIDByToken(ctx, hash)
	if err != nil {
		return models.User{}, err
	}

	user, err := s.users.ByID(ctx, userID)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			return models.User{}, ErrSessionNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.delete(ctx, hashToken(token))
}

func (s *Service) LogoutEverywhere(ctx context.Context, userID string) error {
	return s.sessions.deleteAllForUser(ctx, userID)
}

func (s *Service) DeleteAccount(ctx context.Context, userID string) error {
	if _, err := s.users.ByID(ctx, userID); err != nil {
		return err
	}
	return s.users.Delete(ctx, userID)
}

func (s *Service) PurgeExpiredSessions(ctx context.Context) {
	removed, err := s.sessions.deleteExpired(ctx)
	if err != nil {
		s.log.Error("session cleanup failed", "error", err)
		return
	}
	if removed > 0 {
		s.log.Info("removed expired sessions", "count", removed)
	}
}

func (s *Service) startSession(ctx context.Context, userID string) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}
	if err := s.sessions.create(ctx, hashToken(token), userID, time.Now().Add(s.ttl)); err != nil {
		return "", err
	}
	return token, nil
}

func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// SignInWithGoogle exchanges a Google ID token for a session, creating the
// account on first sign-in.
func (s *Service) SignInWithGoogle(ctx context.Context, idToken, referralCode string) (models.User, string, error) {
	identity, err := s.google.Verify(ctx, idToken)
	if err != nil {
		return models.User{}, "", err
	}

	user, err := s.users.ByGoogleSub(ctx, identity.Subject)
	if err == nil {
		token, err := s.startSession(ctx, user.ID)
		if err != nil {
			return models.User{}, "", err
		}
		return user, token, nil
	}
	if !errors.Is(err, users.ErrNotFound) {
		return models.User{}, "", err
	}

	user, err = s.users.ByEmail(ctx, identity.Email)
	if err == nil {
		linked, err := s.users.LinkGoogleSub(ctx, user.ID, identity.Subject)
		if err != nil {
			return models.User{}, "", err
		}
		token, err := s.startSession(ctx, linked.ID)
		if err != nil {
			return models.User{}, "", err
		}
		return linked, token, nil
	}
	if !errors.Is(err, users.ErrNotFound) {
		return models.User{}, "", err
	}

	return s.registerWithGoogle(ctx, identity, referralCode)
}

func (s *Service) registerWithGoogle(ctx context.Context, identity GoogleIdentity, referralCode string) (models.User, string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.User{}, "", fmt.Errorf("begin google register tx: %w", err)
	}
	defer tx.Rollback(ctx)

	myReferralCode, err := referrals.GenerateCode()
	if err != nil {
		return models.User{}, "", fmt.Errorf("generate referral code: %w", err)
	}

	user, err := s.users.CreateTx(ctx, tx, users.CreateParams{
		Email:        identity.Email,
		GoogleSub:    identity.Subject,
		Name:         identity.Name,
		NepalRegion:  "Kathmandu",
		Timezone:     "Asia/Kathmandu",
		ReferralCode: myReferralCode,
	})
	if err != nil {
		return models.User{}, "", err
	}

	if s.referrals != nil && strings.TrimSpace(referralCode) != "" {
		if _, err := s.referrals.LinkReferralOnRegister(ctx, tx, user.ID, referralCode); err != nil {
			s.log.Warn("referral linking notice during google registration", "error", err, "code", referralCode)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, "", fmt.Errorf("commit google register tx: %w", err)
	}

	token, err := s.startSession(ctx, user.ID)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}
