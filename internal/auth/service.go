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
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/referrals"
	"github.com/prepyo/backend/internal/sms"
	"github.com/prepyo/backend/internal/users"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = users.ErrEmailTaken
	ErrPhoneTaken         = users.ErrPhoneTaken
	ErrEmailRequired      = errors.New("email is required for a new account")
	ErrEmailInvalid       = errors.New("email is not valid")
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

	sms sms.Sender
	// testCodes maps a phone number to a fixed code, for numbers that cannot
	// receive an SMS. It comes from configuration and is empty unless set, so
	// nothing is bypassed by default.
	testCodes map[string]string
}

func NewService(db *pgxpool.Pool, userRepo *users.Repository, referrals ReferralsService, ttl time.Duration, log *slog.Logger, sender sms.Sender, testCodes map[string]string) *Service {
	return &Service{
		db:        db,
		users:     userRepo,
		referrals: referrals,
		sessions:  &sessionRepository{db: db},
		ttl:       ttl,
		log:       log,
		sms:       sender,
		testCodes: testCodes,
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
// DeleteAccount confirms the request before destroying anything. An account
// created by phone has no password to check, so it confirms with a fresh code
// sent to that number instead.
func (s *Service) DeleteAccount(ctx context.Context, userID, password, code string) error {
	user, err := s.users.ByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.PasswordHash == "" {
		if user.Phone == "" {
			return ErrInvalidCredentials
		}
		if err := s.checkOTP(ctx, user.Phone, code); err != nil {
			return err
		}
		return s.users.Delete(ctx, userID)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}
	return s.users.Delete(ctx, userID)
}

// checkOTP validates and burns the newest unconsumed code for a number.
func (s *Service) checkOTP(ctx context.Context, phone, code string) error {
	pending, err := s.latestOTP(ctx, phone)
	if err != nil {
		return ErrOTPInvalid
	}
	if pending.attempts >= otpMaxAttempts {
		return ErrOTPRateLimited
	}
	if time.Now().After(pending.expiresAt) {
		return ErrOTPExpired
	}
	if !constantTimeEqual(pending.codeHash, hashOTP(phone, strings.TrimSpace(code))) {
		s.recordOTPAttempt(ctx, pending.id)
		return ErrOTPInvalid
	}
	return s.consumeOTP(ctx, pending.id)
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

// RequestOTP issues a code for a phone number and sends it.
//
// It reports whether the number already has an account, which is what lets the
// sign-in screen ask a first-time user for their name without a second round
// trip.
func (s *Service) RequestOTP(ctx context.Context, rawPhone string) (bool, error) {
	phone, err := NormalisePhone(rawPhone)
	if err != nil {
		return false, err
	}

	count, sinceLast, err := s.recentOTPCount(ctx, phone)
	if err != nil {
		return false, err
	}
	if count >= otpMaxPerHour || sinceLast < otpMinInterval {
		return false, ErrOTPRateLimited
	}

	registered := true
	if _, err := s.users.ByPhone(ctx, phone); err != nil {
		if !errors.Is(err, users.ErrNotFound) {
			return false, err
		}
		registered = false
	}

	code, ok := s.testCodes[phone]
	if !ok {
		code, err = newOTPCode()
		if err != nil {
			return false, err
		}
	}

	if err := s.storeOTP(ctx, phone, hashOTP(phone, code), time.Now().Add(otpTTL)); err != nil {
		return false, err
	}

	// A test number has a code the holder already knows, so there is nothing to
	// deliver and no reason to spend a message on it.
	if ok {
		return registered, nil
	}

	if err := s.sms.Send(ctx, phone, otpMessage(code)); err != nil {
		return false, fmt.Errorf("send otp: %w", err)
	}
	return registered, nil
}

// VerifyOTP checks a code and returns the account it belongs to, creating one
// on first use. Name is required only when the number is new.
func (s *Service) VerifyOTP(ctx context.Context, rawPhone, code, name, email, referralCode string) (models.User, string, error) {
	phone, err := NormalisePhone(rawPhone)
	if err != nil {
		return models.User{}, "", err
	}

	// Whether the account exists decides whether a name is required, and that
	// has to be settled before the code is checked: checkOTP burns the code, so
	// asking for a name afterwards would cost the learner a fresh SMS.
	user, lookupErr := s.users.ByPhone(ctx, phone)
	if lookupErr != nil && !errors.Is(lookupErr, users.ErrNotFound) {
		return models.User{}, "", lookupErr
	}
	isNew := errors.Is(lookupErr, users.ErrNotFound)
	if isNew {
		if strings.TrimSpace(name) == "" {
			return models.User{}, "", ErrNameRequired
		}
		if strings.TrimSpace(email) == "" {
			return models.User{}, "", ErrEmailRequired
		}
		if _, err := mail.ParseAddress(strings.TrimSpace(email)); err != nil {
			return models.User{}, "", ErrEmailInvalid
		}
	}

	if err := s.checkOTP(ctx, phone, code); err != nil {
		return models.User{}, "", err
	}

	if !isNew {
		token, err := s.startSession(ctx, user.ID)
		if err != nil {
			return models.User{}, "", err
		}
		return user, token, nil
	}
	return s.registerByPhone(ctx, phone, strings.TrimSpace(name), strings.TrimSpace(email), referralCode)
}

func (s *Service) registerByPhone(ctx context.Context, phone, name, email, referralCode string) (models.User, string, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.User{}, "", fmt.Errorf("begin phone register tx: %w", err)
	}
	defer tx.Rollback(ctx)

	myReferralCode, err := referrals.GenerateCode()
	if err != nil {
		return models.User{}, "", fmt.Errorf("generate referral code: %w", err)
	}

	user, err := s.users.CreateTx(ctx, tx, users.CreateParams{
		Phone:        phone,
		Email:        strings.ToLower(email),
		Name:         name,
		NepalRegion:  "Kathmandu",
		Timezone:     "Asia/Kathmandu",
		ReferralCode: myReferralCode,
	})
	if err != nil {
		return models.User{}, "", err
	}

	if s.referrals != nil && strings.TrimSpace(referralCode) != "" {
		if _, err := s.referrals.LinkReferralOnRegister(ctx, tx, user.ID, referralCode); err != nil {
			s.log.Warn("referral linking notice during phone registration", "error", err, "code", referralCode)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return models.User{}, "", fmt.Errorf("commit phone register tx: %w", err)
	}

	token, err := s.startSession(ctx, user.ID)
	if err != nil {
		return models.User{}, "", err
	}
	return user, token, nil
}
