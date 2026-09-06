package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"math/big"
	"time"
)

var (
	ErrOTPRateLimited = errors.New("too many codes requested")
	ErrOTPInvalid     = errors.New("that code is not right")
	ErrOTPExpired     = errors.New("that code has expired")
	ErrNameRequired   = errors.New("name is required for a new account")
)

const (
	otpTTL         = 5 * time.Minute
	otpMinInterval = 60 * time.Second
	otpMaxPerHour  = 5
	otpMaxAttempts = 5
	otpCodeDigits  = 6
)

func newOTPCode() (string, error) {
	limit := big.NewInt(1)
	for i := 0; i < otpCodeDigits; i++ {
		limit.Mul(limit, big.NewInt(10))
	}
	n, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", fmt.Errorf("generate otp: %w", err)
	}
	return fmt.Sprintf("%0*d", otpCodeDigits, n), nil
}

// hashOTP binds the code to the number it was issued for, so a code harvested
// for one phone cannot be replayed against another.
func hashOTP(phone, code string) []byte {
	sum := sha256.Sum256([]byte(phone + ":" + code))
	return sum[:]
}

func otpMessage(code string) string {
	return code + " is your Prepyo verification code. It expires in 5 minutes. Do not share it with anyone."
}

// recentOTPCount reports how many codes went to this number in the last hour,
// and how long ago the most recent one was sent.
func (s *Service) recentOTPCount(ctx context.Context, phone string) (int, time.Duration, error) {
	var count int
	var lastSent *time.Time

	err := s.db.QueryRow(ctx, `
		SELECT count(*), max(created_at)
		  FROM otp_codes
		 WHERE phone = $1 AND created_at > now() - interval '1 hour'`,
		phone,
	).Scan(&count, &lastSent)
	if err != nil {
		return 0, 0, fmt.Errorf("count recent otps: %w", err)
	}

	if lastSent == nil {
		return count, otpMinInterval, nil
	}
	return count, time.Since(*lastSent), nil
}

func (s *Service) storeOTP(ctx context.Context, phone string, codeHash []byte, expiresAt time.Time) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO otp_codes (phone, code_hash, expires_at) VALUES ($1, $2, $3)`,
		phone, codeHash, expiresAt)
	if err != nil {
		return fmt.Errorf("store otp: %w", err)
	}
	return nil
}

type pendingOTP struct {
	id        string
	codeHash  []byte
	expiresAt time.Time
	attempts  int
}

func (s *Service) latestOTP(ctx context.Context, phone string) (pendingOTP, error) {
	var p pendingOTP
	err := s.db.QueryRow(ctx, `
		SELECT id, code_hash, expires_at, attempts
		  FROM otp_codes
		 WHERE phone = $1 AND consumed_at IS NULL
		 ORDER BY created_at DESC
		 LIMIT 1`,
		phone,
	).Scan(&p.id, &p.codeHash, &p.expiresAt, &p.attempts)
	return p, err
}

func (s *Service) recordOTPAttempt(ctx context.Context, id string) {
	if _, err := s.db.Exec(ctx, `UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`, id); err != nil {
		s.log.Error("record otp attempt", "error", err)
	}
}

func (s *Service) consumeOTP(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, `UPDATE otp_codes SET consumed_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("consume otp: %w", err)
	}
	return nil
}

// PurgeExpiredOTPs drops codes that can no longer be used. Kept on the same
// schedule as expired sessions.
func (s *Service) PurgeExpiredOTPs(ctx context.Context) {
	if _, err := s.db.Exec(ctx,
		`DELETE FROM otp_codes WHERE created_at < now() - interval '1 day'`); err != nil {
		s.log.Error("purge expired otps", "error", err)
	}
}

func constantTimeEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
