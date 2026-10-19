package institution

import (
	"fmt"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword turns a password into the bcrypt hash that is stored.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// passwordMatches checks a password against a stored hash.
// With an empty hash (no such account) it still does the same work, so the
// time a sign-in takes does not reveal which emails have an account.
func passwordMatches(hash, password string) bool {
	if hash == "" {
		_ = bcrypt.CompareHashAndPassword(placeholderHash(), []byte(password))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

var (
	placeholderOnce  sync.Once
	placeholderValue []byte
)

// placeholderHash is a real hash that belongs to no account, made once.
func placeholderHash() []byte {
	placeholderOnce.Do(func() {
		placeholderValue, _ = bcrypt.GenerateFromPassword([]byte("no account has this password"), bcrypt.DefaultCost)
	})
	return placeholderValue
}
