package auth

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// newAdminTestService builds a service with no database. Every case here is
// rejected before the user lookup, so the nil pool is never reached.
func newAdminTestService(email, password string) *Service {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(nil, nil, nil, time.Hour, log, nil, email, password)
}

// The binary carries a credential, so the login is configured even with a bare
// environment. There is no "admin sign-in is switched off" state to fall into.
func TestAdminLoginConfigured(t *testing.T) {
	cases := []struct {
		name        string
		email       string
		password    string
		wantBuiltIn bool
	}{
		{"bare environment", "", "", true},
		{"email override only", "ops@prepyo.online", "", true},
		{"password override", "", "a-long-enough-password", false},
		{"both overridden", "ops@prepyo.online", "a-long-enough-password", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := newAdminTestService(tc.email, tc.password)
			if !svc.AdminLoginConfigured() {
				t.Error("AdminLoginConfigured() = false; the built-in credential should always configure it")
			}
			if got := svc.AdminUsesBuiltInPassword(); got != tc.wantBuiltIn {
				t.Errorf("AdminUsesBuiltInPassword() = %v, want %v", got, tc.wantBuiltIn)
			}
		})
	}
}

// A hash truncated or mistyped when it was pasted in would leave nobody able to
// sign in, and every attempt would look like a wrong password.
func TestDefaultAdminPasswordHashIsUsable(t *testing.T) {
	cost, err := bcrypt.Cost([]byte(defaultAdminPasswordHash))
	if err != nil {
		t.Fatalf("built-in admin password hash is not valid bcrypt: %v", err)
	}
	if cost < bcrypt.DefaultCost {
		t.Errorf("built-in admin password hash cost = %d, want at least %d", cost, bcrypt.DefaultCost)
	}
}

// An override must replace the built-in password, not sit alongside it.
func TestAdminPasswordOverrideReplacesBuiltIn(t *testing.T) {
	const override = "an-override-password"
	creds := newAdminCredentials("", override)

	if !creds.matches(defaultAdminEmail, override) {
		t.Error("the override password was rejected")
	}
	if creds.usesBuiltIn() {
		t.Error("usesBuiltIn() = true while an override is set")
	}
	if creds.matches(defaultAdminEmail, "some-other-password") {
		t.Error("a password other than the override was accepted")
	}
}

// The built-in email is used when nothing overrides it, and an override wins.
func TestAdminEmailDefaults(t *testing.T) {
	if got := newAdminCredentials("", "").email; got != defaultAdminEmail {
		t.Errorf("email = %q, want the built-in %q", got, defaultAdminEmail)
	}
	if got := newAdminCredentials("  OPS@Prepyo.Online ", "pw").email; got != "ops@prepyo.online" {
		t.Errorf("email = %q, want it trimmed and lowercased", got)
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

// The built-in credential rejects a wrong password the same way an override
// does, and reports it as a credential failure rather than a broken setup.
func TestSignInAsAdminBuiltInRejectsWrongPassword(t *testing.T) {
	svc := newAdminTestService("", "")

	_, _, err := svc.SignInAsAdmin(context.Background(), defaultAdminEmail, "not-the-password")
	if !errors.Is(err, ErrAdminCredentials) {
		t.Fatalf("SignInAsAdmin() error = %v, want %v", err, ErrAdminCredentials)
	}
}
