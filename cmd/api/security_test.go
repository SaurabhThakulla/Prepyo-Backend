package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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
