package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

// newAdminTestService builds a service with no database. Every case here is
// rejected before the user lookup, so the nil pool is never reached.
func newAdminTestService(email, password string) *Service {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(nil, nil, nil, time.Hour, log, nil, email, password)
}

func TestAdminLoginConfigured(t *testing.T) {
	cases := []struct {
		name     string
		email    string
		password string
		want     bool
	}{
		{"both set", "admin@prepyo.online", "a-long-enough-password", true},
		{"no password", "admin@prepyo.online", "", false},
		{"no email", "", "a-long-enough-password", false},
		{"neither", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := newAdminTestService(tc.email, tc.password).AdminLoginConfigured(); got != tc.want {
				t.Fatalf("AdminLoginConfigured() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSignInAsAdminRejectsBadCredentials(t *testing.T) {
	const (
		adminEmail = "admin@prepyo.online"
		adminPass  = "correct-horse-battery"
	)

	cases := []struct {
		name     string
		email    string
		password string
		want     error
	}{
		{"wrong password", adminEmail, "not-the-password", ErrAdminCredentials},
		{"wrong email", "someone@example.com", adminPass, ErrAdminCredentials},
		{"empty password", adminEmail, "", ErrAdminCredentials},
		{"password in the email field", adminPass, adminPass, ErrAdminCredentials},
		{"password as a prefix", adminEmail, "correct-horse", ErrAdminCredentials},
	}

	svc := newAdminTestService(adminEmail, adminPass)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := svc.SignInAsAdmin(context.Background(), tc.email, tc.password)
			if !errors.Is(err, tc.want) {
				t.Fatalf("SignInAsAdmin() error = %v, want %v", err, tc.want)
			}
		})
	}
}

// The email is normalised the same way the users table stores it, so casing and
// stray whitespace from a form field still reach the account.
func TestSignInAsAdminNormalisesEmail(t *testing.T) {
	svc := newAdminTestService("admin@prepyo.online", "correct-horse-battery")

	// A nil user repository panics once the credential check passes, which is
	// how this test tells "matched" from "rejected" without a database.
	defer func() {
		if recover() == nil {
			t.Fatal("expected the lookup to be reached, but the credentials were rejected")
		}
	}()

	_, _, _ = svc.SignInAsAdmin(context.Background(), "  ADMIN@Prepyo.Online  ", "correct-horse-battery")
}

func TestSignInAsAdminNotConfigured(t *testing.T) {
	svc := newAdminTestService("admin@prepyo.online", "")

	_, _, err := svc.SignInAsAdmin(context.Background(), "admin@prepyo.online", "anything")
	if !errors.Is(err, ErrAdminLoginNotConfigured) {
		t.Fatalf("SignInAsAdmin() error = %v, want %v", err, ErrAdminLoginNotConfigured)
	}
}
