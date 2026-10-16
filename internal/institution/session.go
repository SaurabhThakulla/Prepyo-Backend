package institution

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prepyo/backend/pkg/httpx"
)

const (
	sessionCookie     = "prepyo_institution_session"
	sessionCookiePath = "/api/v1/institution" // sent only to the institution routes
)

// newSessionToken makes the random value that goes in the cookie.
func newSessionToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("make session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// hashToken is what the database stores in place of the token.
func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

// startSession stores a new session and gives the browser its cookie.
func (h *Handler) startSession(ctx context.Context, w http.ResponseWriter, adminID string) error {
	token, err := newSessionToken()
	if err != nil {
		return err
	}
	if err := h.store.saveSession(ctx, adminID, hashToken(token), time.Now().Add(h.sessionTTL)); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     sessionCookiePath,
		HttpOnly: true, // out of reach of page scripts
		SameSite: http.SameSiteLaxMode,
		Secure:   h.secureCookies,
		MaxAge:   int(h.sessionTTL.Seconds()),
	})
	return nil
}

// endSession deletes the session, if there is one, and removes the cookie.
func (h *Handler) endSession(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie(sessionCookie); err == nil && cookie.Value != "" {
		if err := h.store.deleteSession(r.Context(), hashToken(cookie.Value)); err != nil {
			return err
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     sessionCookiePath,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.secureCookies,
		MaxAge:   -1,
	})
	return nil
}

// requireAdmin lets a request through only when it carries a live session.
func (h *Handler) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const signInFirst = "Please sign in to continue."

		cookie, err := r.Cookie(sessionCookie)
		if err != nil || cookie.Value == "" {
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, signInFirst)
			return
		}

		account, err := h.store.accountBySession(r.Context(), hashToken(cookie.Value))
		if errors.Is(err, errNotFound) {
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, signInFirst)
			return
		}
		if err != nil {
			httpx.Internal(w, h.log, "institution.requireAdmin", err)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), accountKey{}, account)))
	})
}

type accountKey struct{}

// signedInAccount returns the account requireAdmin found. Use only behind requireAdmin.
func signedInAccount(r *http.Request) Account {
	return r.Context().Value(accountKey{}).(Account)
}
