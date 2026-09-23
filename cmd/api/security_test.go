package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/prepyo/backend/internal/auth"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
)

func TestRateLimitIgnoresSpoofedForwardingHeaders(t *testing.T) {
	handler := trustedClientIP(nil)(rateLimit(2, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "198.51.100.10:1234"
		req.Header.Set("X-Forwarded-For", "10.0.0."+string(rune('1'+i)))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if i < 2 && rec.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want 204", i, rec.Code)
		}
		if i == 2 && rec.Code != http.StatusTooManyRequests {
			t.Fatalf("spoofed forwarding header bypassed limiter: status=%d", rec.Code)
		}
	}
}

func TestTrustedProxyResolution(t *testing.T) {
	var got string
	handler := trustedClientIP([]string{"10.0.0.0/8"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = middleware.GetClientIP(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 10.0.0.4")
	handler.ServeHTTP(httptest.NewRecorder(), req)
	if got != "203.0.113.7" {
		t.Fatalf("resolved client IP = %q, want 203.0.113.7", got)
	}
}

func TestAdminCORSAllowsPUT(t *testing.T) {
	handler := cors.Handler(apiCORSOptions([]string{"https://admin.example"}))(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://admin.example")
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusOK {
		t.Fatalf("preflight status = %d, want 200 or 204", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatal("preflight did not return allowed methods")
	}
	if !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "PUT") {
		t.Fatalf("allowed methods = %q, want PUT", rec.Header().Get("Access-Control-Allow-Methods"))
	}
}

func TestHealthPayloadHasNoProviderDiagnostics(t *testing.T) {
	payload := healthPayload()
	if len(payload) != 2 || payload["status"] != "healthy" || payload["version"] == "" {
		t.Fatalf("unexpected health payload: %#v", payload)
	}
}

func TestSecurityHeaders(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })

	rec := httptest.NewRecorder()
	securityHeaders(false)(ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	for header, want := range map[string]string{
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "frame-ancestors 'none'",
		"X-Content-Type-Options":  "nosniff",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
	if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS sent without HTTPS cookies: %q", got)
	}

	rec = httptest.NewRecorder()
	securityHeaders(true)(ok).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Error("HSTS missing when served over HTTPS")
	}
}

// Learners behind one address (a classroom's Wi-Fi, a carrier's shared IP)
// each get their own allowance: one device using its limit up blocks nobody
// else. Signed in, the device is the session; signed out, the device cookie.
func TestRateLimitIsPerDevice(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := trustedClientIP(nil)(deviceCookie(false)(rateLimitByDevice(2, time.Minute)(ok)))

	send := func(session, device string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.7:1234"
		if session != "" {
			req.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: session})
			req = req.WithContext(reqctx.WithUser(req.Context(), models.User{ID: "learner-" + session}))
		}
		if device != "" {
			req.AddCookie(&http.Cookie{Name: deviceCookieName, Value: device})
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	// Signed in: one session runs out, a classmate's session on the same IP does not.
	send("busy-session", "")
	send("busy-session", "")
	if code := send("busy-session", ""); code != http.StatusTooManyRequests {
		t.Fatalf("busy device over its limit = %d, want 429", code)
	}
	if code := send("classmate-session", ""); code != http.StatusNoContent {
		t.Fatalf("classmate on the same IP was blocked: %d", code)
	}

	// Signed out: each browser's device cookie is its own bucket.
	phone := newDeviceID()
	laptop := newDeviceID()
	send("", phone)
	send("", phone)
	if code := send("", phone); code != http.StatusTooManyRequests {
		t.Fatalf("signed-out device over its limit = %d, want 429", code)
	}
	if code := send("", laptop); code != http.StatusNoContent {
		t.Fatalf("another signed-out device on the same IP was blocked: %d", code)
	}
}

// A browser without a device id is given one, and a malformed id is replaced
// rather than trusted as a bucket of its own.
func TestDeviceCookieIsIssued(t *testing.T) {
	handler := deviceCookie(true)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))

	for _, existing := range []string{"", "not a real id"} {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if existing != "" {
			req.AddCookie(&http.Cookie{Name: deviceCookieName, Value: existing})
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		cookies := rec.Result().Cookies()
		if len(cookies) != 1 || !deviceIDPattern.MatchString(cookies[0].Value) || !cookies[0].HttpOnly || !cookies[0].Secure {
			t.Fatalf("existing=%q: got cookies %+v, want one fresh HttpOnly, Secure device id", existing, cookies)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: deviceCookieName, Value: newDeviceID()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if len(rec.Result().Cookies()) != 0 {
		t.Fatal("a valid device id was replaced")
	}
}
