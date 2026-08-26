// Package auth handles registration, login, sessions and the middleware that
// puts the current user on the request context.
//
// Sessions are opaque random tokens kept in a database table, not signed JWTs.
// That makes revocation a DELETE, which is what "session revocation" in the
// spec needs and what a JWT cannot do without extra machinery.
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
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = users.ErrEmailTaken
)

// bcryptCost of 12 is roughly 250ms on current hardware: slow enough to make
// offline cracking expensive, fast enough for a login request.
const bcryptCost = 12

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
}

func NewService(db *pgxpool.Pool, userRepo *users.Repository, referrals ReferralsService, ttl time.Duration, log *slog.Logger) *Service {
	return &Service{
		db:        db,
		users:     userRepo,
		referrals: referrals,
		sessions:  &sessionRepository{db: db},
		ttl:       ttl,
		log:       log,
	}
}

type RegisterParams struct {
	Email        string
	Password     string
	Name         string
	NepalRegion  string
	Timezone     string
	ReferralCode string
}

// Register creates an account and returns it with a fresh session token.
func (s *Service) Register(ctx context.Context, p RegisterParams) (models.User, string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcryptCost)
	if err != nil {
		return models.User{}, "", fmt.Errorf("hash password: %w", err)
	}

	region := strings.TrimSpace(p.NepalRegion)
	if region == "" {
		region = "Kathmandu"
	}
	zone := strings.TrimSpace(p.Timezone)
	if zone == "" {
		zone = "Asia/Kathmandu"
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.User{}, "", fmt.Errorf("begin register tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Generate a unique referral code for this new user
	myReferralCode, err := referrals.GenerateCode()
	if err != nil {
		return models.User{}, "", fmt.Errorf("generate referral code: %w", err)
	}

	user, err := s.users.CreateTx(ctx, tx, users.CreateParams{
		Email:        strings.ToLower(strings.TrimSpace(p.Email)),
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(p.Name),
		NepalRegion:  region,
		Timezone:     zone,
		ReferralCode: myReferralCode,
	})
	if err != nil {
		return models.User{}, "", err
	}

	// If a referral code was provided, link the pending referral
	if s.referrals != nil && strings.TrimSpace(p.ReferralCode) != "" {
		if _, err := s.referrals.LinkReferralOnRegister(ctx, tx, user.ID, p.ReferralCode); err != nil {
			// If self referral or invalid code, fail registration or log
			s.log.Warn("referral linking notice during registration", "error", err, "code", p.ReferralCode)
			if errors.Is(err, users.ErrNotFound) || errors.Is(err, errors.New("cannot refer yourself")) {
				return models.User{}, "", err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, "", fmt.Errorf("commit register tx: %w", err)
	}

	token, err := s.startSession(ctx, user.ID)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}

// Login verifies a password and starts a session.
func (s *Service) Login(ctx context.Context, email, password string) (models.User, string, error) {
	user, err := s.users.ByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			// Hash anyway so a missing account and a wrong password take about
			// the same time. Otherwise response timing reveals which emails
			// are registered.
			_, _ = bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
			return models.User{}, "", ErrInvalidCredentials
		}
		return models.User{}, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return models.User{}, "", ErrInvalidCredentials
	}

	token, err := s.startSession(ctx, user.ID)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
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
			// The account was deleted but the session row survived; treat the
			// session as gone.
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

// DeleteAccount removes the user and, by cascade, all of their data.
func (s *Service) DeleteAccount(ctx context.Context, userID, password string) error {
	user, err := s.users.ByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return s.users.Delete(ctx, userID)
}

// PurgeExpiredSessions is run periodically by the background cleaner in main.
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

// newToken returns 256 bits of randomness from the OS. rand.Read never returns
// a short read; an error here means the system entropy source failed, which is
// not something to paper over.
func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// hashToken is what gets stored. SHA-256 is right here rather than bcrypt:
// the input is already 256 random bits, so there is nothing to brute-force and
// lookups stay fast.
func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
