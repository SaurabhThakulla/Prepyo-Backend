package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func trustedClientIP(prefixes []string) func(http.Handler) http.Handler {
	direct := middleware.ClientIPFromRemoteAddr
	if len(prefixes) == 0 {
		return direct
	}
	fromProxy := middleware.ClientIPFromXFF(prefixes...)
	return func(next http.Handler) http.Handler {
		directHandler := direct(next)
		proxyHandler := fromProxy(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := r.RemoteAddr
			if parsed, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				host = parsed
			}
			peer := net.ParseIP(strings.Trim(host, "[]"))
			if peer == nil {
				directHandler.ServeHTTP(w, r)
				return
			}
			for _, prefix := range prefixes {
				_, network, err := net.ParseCIDR(prefix)
				if err == nil && network.Contains(peer) {
					proxyHandler.ServeHTTP(w, r)
					return
				}
			}
			directHandler.ServeHTTP(w, r)
		})
	}
}

func clientIPKey(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}

func healthPayload() map[string]any {
	return map[string]any{
		"status":  "healthy",
		"version": "2026-09-15",
	}
}

// securityHeaders sets the browser protections every response should carry.
//
// A page-wide script CSP is left out on purpose: Google sign-in loads its own
// scripts and frames, and a policy that blocks them locks everyone out. What is
// here costs nothing and closes the cheap attacks: framing the app to trick a
// signed-in learner into clicking (clickjacking), MIME sniffing an upload into
// something runnable, and leaking full URLs to other sites.
func securityHeaders(hsts bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Frame-Options", "DENY")
			h.Set("Content-Security-Policy", "frame-ancestors 'none'")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			// The recorder needs the microphone; nothing needs the camera or location.
			h.Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(self)")
			if hsts {
				h.Set("Strict-Transport-Security", "max-age=31536000")
			}
			next.ServeHTTP(w, r)
		})
	}
}
