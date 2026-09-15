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
