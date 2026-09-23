package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/httprate"

	"github.com/prepyo/backend/internal/auth"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// deviceCookieName identifies one browser for rate limiting only. It carries
// no account and grants nothing.
const deviceCookieName = "prepyo_device"

// deviceIDPattern is what newDeviceID produces. Anything else in the cookie is
// ignored, so a junk value cannot become its own bucket.
var deviceIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{22}$`)

func newDeviceID() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

// deviceCookie gives every browser a random device id the first time it calls
// the API, so requests from before sign-in can be limited per device rather
// than per network.
func deviceCookie(secure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c, err := r.Cookie(deviceCookieName); err != nil || !deviceIDPattern.MatchString(c.Value) {
				id := newDeviceID()
				http.SetCookie(w, &http.Cookie{
					Name:     deviceCookieName,
					Value:    id,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					Secure:   secure,
					MaxAge:   int((365 * 24 * time.Hour).Seconds()),
				})
				// The limiter below reads the cookie from this request, so the
				// very first call is counted against the new id too.
				r.AddCookie(&http.Cookie{Name: deviceCookieName, Value: id})
			}
			next.ServeHTTP(w, r)
		})
	}
}

// deviceKey buckets a request by device.
//
// Signed in, the device is the session: every sign-in on a phone or laptop
// has its own, and it cannot be invented because RequireUser has already
// checked it. Signed out, it is the device cookie. With neither, the client IP.
//
// Keying by IP alone let one learner lock out everyone behind the same
// address: a coaching centre's Wi-Fi, a college network, or a mobile carrier
// sharing one IP across many customers.
func deviceKey(r *http.Request) (string, error) {
	if _, ok := reqctx.User(r.Context()); ok {
		if c, err := r.Cookie(auth.SessionCookieName); err == nil && c.Value != "" {
			sum := sha256.Sum256([]byte(c.Value))
			return "session:" + hex.EncodeToString(sum[:12]), nil
		}
	}
	if c, err := r.Cookie(deviceCookieName); err == nil && deviceIDPattern.MatchString(c.Value) {
		return "device:" + c.Value, nil
	}
	ip, err := clientIPKey(r)
	return "ip:" + ip, err
}

// rateLimitByDevice throttles each device on its own, so one busy learner
// never uses up the allowance of others on the same network.
func rateLimitByDevice(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(requests, window, deviceKey,
		httprate.WithLimitHandler(httpx.RateLimited))
}
