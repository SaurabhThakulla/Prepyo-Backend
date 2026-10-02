package auth

import (
	"context"
	"errors"
	"fmt"
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

// testIP is the client address failures are counted against in these tests.
const testIP = "203.0.113.5"

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
			_, _, err := svc.SignInAsAdmin(context.Background(), tc.email, tc.password, testIP)
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

	_, _, _ = svc.SignInAsAdmin(context.Background(), "  ADMIN@Prepyo.Online  ", "correct-horse-battery", testIP)
}

func TestSignInAsAdminNotConfigured(t *testing.T) {
	svc := newAdminTestService("admin@prepyo.online", "")

	_, _, err := svc.SignInAsAdmin(context.Background(), "admin@prepyo.online", "anything", testIP)
	if !errors.Is(err, ErrAdminLoginNotConfigured) {
		t.Fatalf("SignInAsAdmin() error = %v, want %v", err, ErrAdminLoginNotConfigured)
	}
}

// Only wrong credentials count toward the lockout, and a correct sign-in wipes
// the slate, so an admin who signs in normally is never throttled.
func TestAdminLockoutCountsOnlyFailures(t *testing.T) {
	const email = "admin@prepyo.online"
	svc := newAdminTestService(email, "correct-horse-battery")
	now := time.Now()

	for i := 0; i < adminLoginMaxFailures-1; i++ {
		svc.recordAdminFailure(testIP, now)
	}
	if svc.adminLoginLocked(testIP, now) {
		t.Fatal("locked before reaching the failure limit")
	}

	svc.clearAdminFailures(testIP)
	for i := 0; i < adminLoginMaxFailures-1; i++ {
		svc.recordAdminFailure(testIP, now)
	}
	if svc.adminLoginLocked(testIP, now) {
		t.Fatal("failures from before a successful sign-in still counted")
	}

	svc.recordAdminFailure(testIP, now)
	if !svc.adminLoginLocked(testIP, now) {
		t.Fatal("not locked after reaching the failure limit")
	}
	if svc.adminLoginLocked(testIP, now.Add(adminLoginWindow+time.Second)) {
		t.Fatal("still locked after the window passed")
	}
}

func TestSignInAsAdminLocksAfterRepeatedFailures(t *testing.T) {
	const email = "admin@prepyo.online"
	svc := newAdminTestService(email, "correct-horse-battery")

	for i := 0; i < adminLoginMaxFailures; i++ {
		if _, _, err := svc.SignInAsAdmin(context.Background(), email, "wrong", testIP); !errors.Is(err, ErrAdminCredentials) {
			t.Fatalf("attempt %d: error = %v, want %v", i+1, err, ErrAdminCredentials)
		}
	}
	if _, _, err := svc.SignInAsAdmin(context.Background(), email, "wrong", testIP); !errors.Is(err, ErrAdminLoginRateLimited) {
		t.Fatalf("error = %v, want %v", err, ErrAdminLoginRateLimited)
	}
}

// Failing on purpose from one address must not lock the admin out elsewhere:
// the lock is per client IP, not per account.
func TestAdminLockoutIsPerClientIP(t *testing.T) {
	const email = "admin@prepyo.online"
	svc := newAdminTestService(email, "correct-horse-battery")
	for i := 0; i < adminLoginMaxFailures; i++ {
		_, _, _ = svc.SignInAsAdmin(context.Background(), email, "wrong", "198.51.100.7")
	}
	if _, _, err := svc.SignInAsAdmin(context.Background(), email, "wrong", "198.51.100.7"); !errors.Is(err, ErrAdminLoginRateLimited) {
		t.Fatalf("attacker's IP: error = %v, want %v", err, ErrAdminLoginRateLimited)
	}
	if _, _, err := svc.SignInAsAdmin(context.Background(), email, "wrong", testIP); !errors.Is(err, ErrAdminCredentials) {
		t.Fatalf("another IP: error = %v, want a normal credential check (%v)", err, ErrAdminCredentials)
	}
}

// Made-up emails can never sign in, so they are not remembered: a stream of
// them cannot grow the failure list.
func TestAdminFailuresIgnoreOtherEmails(t *testing.T) {
	svc := newAdminTestService("admin@prepyo.online", "correct-horse-battery")
	for i := 0; i < 50; i++ {
		_, _, _ = svc.SignInAsAdmin(context.Background(), "random@example.com", "wrong", testIP)
	}
	if len(svc.adminAttempts) != 0 {
		t.Fatalf("tracked %d entries for emails that are not the admin's", len(svc.adminAttempts))
	}
	if svc.adminLoginLocked(testIP, time.Now()) {
		t.Fatal("wrong emails locked the IP out")
	}
}

// The failure list never holds more than adminLoginMaxTracked IPs.
func TestAdminFailureListIsCapped(t *testing.T) {
	svc := newAdminTestService("admin@prepyo.online", "correct-horse-battery")
	now := time.Now()
	for i := 0; i < adminLoginMaxTracked+50; i++ {
		svc.recordAdminFailure(fmt.Sprintf("ip-%d", i), now.Add(time.Duration(i)*time.Millisecond))
	}
	if got := len(svc.adminAttempts); got > adminLoginMaxTracked {
		t.Fatalf("tracking %d IPs, cap is %d", got, adminLoginMaxTracked)
	}
	if _, ok := svc.adminAttempts[fmt.Sprintf("ip-%d", adminLoginMaxTracked+49)]; !ok {
		t.Fatal("the newest failure was dropped instead of the stalest")
	}
}
