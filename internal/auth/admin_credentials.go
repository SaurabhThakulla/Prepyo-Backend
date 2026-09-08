package auth

import (
	"crypto/subtle"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// The operations account's built-in credentials.
//
// These live in the binary so the admin login works on a fresh deploy with no
// environment set up at all. ADMIN_EMAIL and ADMIN_PASSWORD still override
// them, and production should use that override — see adminUsesBuiltIn, which
// is what makes cmd/api warn about it at boot.
//
// The password is stored as a bcrypt hash rather than plaintext: the hash going
// into version control cannot be read back out, whereas a plaintext password
// committed once stays in the history for good, even after it is changed.
//
// To change the built-in password, generate a new hash and replace the constant:
//
//	bcrypt.GenerateFromPassword([]byte("the new password"), bcrypt.DefaultCost)
const (
	defaultAdminEmail = "admin@prepyo.online"
	// bcrypt cost 10. Verifying takes a few milliseconds, which is the point:
	// it puts a floor under how fast the hash can be attacked offline.
	defaultAdminPasswordHash = "$2a$10$cY7xHl4gLtdDzC8iI8T37ueXR22Q6RkjHoeijrXUvPZTV4iZiMdIS"
)

// adminCredentials is how the service checks the one password login in the
// product. Either an environment override is in force, or the built-in hash is.
type adminCredentials struct {
	email string
	// password is the plaintext override from ADMIN_PASSWORD. Empty means the
	// built-in hash is in use.
	password     string
	passwordHash string
}

func newAdminCredentials(email, password string) adminCredentials {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		email = defaultAdminEmail
	}

	creds := adminCredentials{email: email, password: password}
	if password == "" {
		creds.passwordHash = defaultAdminPasswordHash
	}
	return creds
}

// usesBuiltIn reports whether the login is running on the credentials compiled
// into the binary, which anyone with the source can read.
func (c adminCredentials) usesBuiltIn() bool { return c.password == "" }

func (c adminCredentials) configured() bool {
	return c.email != "" && (c.password != "" || c.passwordHash != "")
}

// matches checks one sign-in attempt.
//
// Both halves are always evaluated and combined without a short circuit, so the
// time taken does not reveal whether the email alone was right.
func (c adminCredentials) matches(email, password string) bool {
	emailOK := subtle.ConstantTimeCompare(
		[]byte(strings.ToLower(strings.TrimSpace(email))), []byte(c.email)) == 1

	var passwordOK bool
	if c.password != "" {
		passwordOK = subtle.ConstantTimeCompare([]byte(password), []byte(c.password)) == 1
	} else {
		// bcrypt compares in constant time for a given hash and reports every
		// failure the same way, so a wrong password and a malformed hash are
		// indistinguishable here by design.
		passwordOK = bcrypt.CompareHashAndPassword([]byte(c.passwordHash), []byte(password)) == nil
	}

	return emailOK && passwordOK
}
