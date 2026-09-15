package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/referrals"
	"github.com/prepyo/backend/internal/users"
)

var (
	ErrEmailTaken = users.ErrEmailTaken

	ErrAdminLoginNotConfigured = errors.New("admin password login is not configured")
	// ErrAdminCredentials covers a wrong email, a wrong password and an account
	// that is not an admin. They are one error on purpose: the response must not
	// tell a guesser which half they got right.
	ErrAdminCredentials      = errors.New("admin credentials are not valid")
	ErrAdminLoginRateLimited = errors.New("admin login rate limited")
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

	// The operations account's credential, held in memory from the
	// environment. No password is stored in the database for anyone.
	adminEmail    string
	adminPassword string
	adminMu       sync.Mutex
	adminAttempts map[string][]time.Time
}

func NewService(db *pgxpool.Pool, userRepo *users.Repository, referrals ReferralsService, ttl time.Duration, log *slog.Logger, google *GoogleVerifier, adminEmail, adminPassword string) *Service {
	return &Service{
		db:            db,
		users:         userRepo,
		referrals:     referrals,
		sessions:      &sessionRepository{db: db},
		ttl:           ttl,
		log:           log,
		google:        google,
		adminEmail:    strings.ToLower(strings.TrimSpace(adminEmail)),
		adminPassword: adminPassword,
		adminAttempts: make(map[string][]time.Time),
	}
}

// AdminLoginConfigured reports whether SignInAsAdmin has a credential to check.
func (s *Service) AdminLoginConfigured() bool {
	return s.adminEmail != "" && s.adminPassword != ""
}

// SignInAsAdmin exchanges the operations account's email and password for a
// session. It is the only password login in the product: every learner account
// signs in with Google, and this one exists because admin@prepyo.online has no
// Google account behind it.
func (s *Service) SignInAsAdmin(ctx context.Context, email, password string) (models.User, string, error) {
	if !s.AdminLoginConfigured() {
		return models.User{}, "", ErrAdminLoginNotConfigured
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if !s.allowAdminAttempt(email, time.Now()) {
		return models.User{}, "", ErrAdminLoginRateLimited
	}

	// Both halves are compared in constant time and combined without a short
	// circuit, so response timing does not leak which one matched.
	emailOK := subtle.ConstantTimeCompare([]byte(email), []byte(s.adminEmail))
	passwordOK := subtle.ConstantTimeCompare([]byte(password), []byte(s.adminPassword))
	if emailOK&passwordOK != 1 {
		return models.User{}, "", ErrAdminCredentials
	}

	user, err := s.users.ByEmail(ctx, s.adminEmail)
	if err != nil {
		if errors.Is(err, users.ErrNotFound) {
			// The credential is right but the seeded row is gone, so this is a
			// deployment problem rather than a bad guess. The caller still gets
			// the generic error.
			s.log.Error("admin login: no user row for the configured admin email", "email", s.adminEmail)
			return models.User{}, "", ErrAdminCredentials
		}
		return models.User{}, "", err
	}

	// The role lives in the database, so revoking admin there revokes this
	// login too, whatever the environment still says.
	if !user.IsAdmin() {
		s.log.Error("admin login: configured admin account does not have the admin role", "email", s.adminEmail)
		return models.User{}, "", ErrAdminCredentials
	}

	token, err := s.startSession(ctx, user.ID)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}

// Authenticate resolves a session token to its user.
const adminLoginWindow = 10 * time.Minute
const adminLoginMaxAttempts = 5

// allowAdminAttempt adds an account/email dimension to the outer IP limiter.
// The key is the normalized email, so changing IPs does not evade throttling.
func (s *Service) allowAdminAttempt(email string, now time.Time) bool {
	s.adminMu.Lock()
	defer s.adminMu.Unlock()

	cutoff := now.Add(-adminLoginWindow)
	attempts := s.adminAttempts[email][:0]
	for _, at := range s.adminAttempts[email] {
		if at.After(cutoff) {
			attempts = append(attempts, at)
		}
	}
	if len(attempts) >= adminLoginMaxAttempts {
		s.adminAttempts[email] = attempts
		return false
	}
	s.adminAttempts[email] = append(attempts, now)
	return true
}
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
