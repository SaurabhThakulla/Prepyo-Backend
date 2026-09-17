// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env            string
	Port           string
	AllowedOrigins []string
	// TrustedProxyCIDRs enumerates reverse proxies allowed to supply X-Forwarded-For.
	TrustedProxyCIDRs []string
	WebAppURL         string

	DatabaseURL string
	RedisURL    string

	SessionTTL time.Duration
	// SecureCookies must be true anywhere the app is served over HTTPS.
	SecureCookies bool

	// AI providers.
	AIBaseURL      string
	AIAPIKey       string
	AIAudioBaseURL string
	AIAudioAPIKey  string

	AIModels         AIModels
	AIRequestTimeout time.Duration
	// AIMaxTokens caps total completion tokens per request.
	AIMaxTokens int

	GoogleClientID string

	// AdminEmail / AdminPassword back the one password login the product has:
	// POST /auth/admin-login for the operations account. Learner accounts are
	// Google-only and no password is ever stored for them.
	AdminEmail    string
	AdminPassword string

	// WebDistDir is the optional directory path to frontend production build assets.
	WebDistDir string
}

// AdminLoginEnabled reports whether the admin password endpoint has a
// credential to check against. When false it returns 503 rather than comparing
// against an empty password and letting anyone in.
func (c Config) AdminLoginEnabled() bool { return c.AdminEmail != "" && c.AdminPassword != "" }

// AIModels defines model names for each AI task capability.
type AIModels struct {
	Writing  string
	Speaking string
	Tutoring string
}

func (c Config) IsProduction() bool { return c.Env == "production" }

// AIEnabled reports whether the AI provider is configured.
func (c Config) AIEnabled() bool { return c.AIAPIKey != "" }

// SpeakingEnabled reports whether the audio provider is configured.
func (c Config) SpeakingEnabled() bool { return c.AIAudioAPIKey != "" }

func Load() (*Config, error) {
	loadDotEnv()

	var problems []string

	env := stringOr("APP_ENV", "development")
	if env != "development" && env != "staging" && env != "production" {
		problems = append(problems, `APP_ENV must be development, staging or production`)
	}
	isProd := env == "production"

	aiBaseURL := strings.TrimRight(stringOr("AI_BASE_URL", "https://codecraftapi.com/v1"), "/")
	aiKey := firstOf("AI_API_KEY", "CODE_CRAFT")
	aiAudioKey := stringOr("AI_AUDIO_API_KEY", aiKey)

	cfg := &Config{
		Env:               env,
		Port:              stringOr("PORT", "8080"),
		AllowedOrigins:    listOr("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		TrustedProxyCIDRs: listOr("TRUSTED_PROXY_CIDRS", nil),
		WebAppURL:         stringOr("WEB_APP_URL", "http://localhost:3000"),
		RedisURL:          os.Getenv("REDIS_URL"),

		AIBaseURL: aiBaseURL,
		AIAPIKey:  aiKey,

		AIAudioBaseURL: strings.TrimRight(stringOr("AI_AUDIO_BASE_URL", aiBaseURL), "/"),
		AIAudioAPIKey:  aiAudioKey,

		AIModels: AIModels{
			Writing:  cleanAIModel(stringOr("AI_MODEL_WRITING", "gpt-5.6-luna"), "gpt-5.6-luna"),
			Speaking: cleanAIModel(stringOr("AI_MODEL_SPEAKING", "gemini-3.7-flash"), "gemini-3.7-flash"),
			Tutoring: cleanAIModel(stringOr("AI_MODEL_TUTORING", "gpt-5.6-luna"), "gpt-5.6-luna"),
		},
		SecureCookies: boolOr("SECURE_COOKIES", isProd),

		GoogleClientID: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),

		AdminEmail:    strings.ToLower(stringOr("ADMIN_EMAIL", "admin@prepyo.online")),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),

		WebDistDir: os.Getenv("WEB_DIST_DIR"),
	}

	if cfg.WebDistDir != "" {
		if _, err := os.Stat(filepath.Join(cfg.WebDistDir, "index.html")); err != nil {
			problems = append(problems, "WEB_DIST_DIR is set but has no index.html in it: "+cfg.WebDistDir)
		}
	}
	for _, prefix := range cfg.TrustedProxyCIDRs {
		if _, _, err := net.ParseCIDR(prefix); err != nil {
			problems = append(problems, "TRUSTED_PROXY_CIDRS contains an invalid CIDR: "+prefix)
		}
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	}

	ttl, err := durationOr("SESSION_TTL", 14*24*time.Hour)
	if err != nil {
		problems = append(problems, "SESSION_TTL "+err.Error())
	}
	cfg.SessionTTL = ttl

	timeout, err := durationOr("AI_REQUEST_TIMEOUT", 45*time.Second)
	if err != nil {
		problems = append(problems, "AI_REQUEST_TIMEOUT "+err.Error())
	}
	cfg.AIRequestTimeout = timeout

	maxTokens, err := intOr("AI_MAX_TOKENS", 8000)
	if err != nil {
		problems = append(problems, "AI_MAX_TOKENS "+err.Error())
	}
	cfg.AIMaxTokens = maxTokens

	if cfg.AdminPassword != "" && len(cfg.AdminPassword) < 12 {
		problems = append(problems, "ADMIN_PASSWORD must be at least 12 characters")
	}

	if isProd {
		if cfg.AIAPIKey == "" {
			problems = append(problems, "AI_API_KEY is required in production")
		}
		if cfg.GoogleClientID == "" {
			problems = append(problems, "GOOGLE_CLIENT_ID is required in production")
		}
		for _, origin := range cfg.AllowedOrigins {
			if isLocalhostOrigin(origin) {
				problems = append(problems, "ALLOWED_ORIGINS must not contain localhost in production: "+origin)
				break
			}
		}
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}

// isLocalhostOrigin reports whether an origin points at the machine itself
// rather than the public internet: loopback IPs (127.0.0.0/8, ::1) plus the
// unspecified address, and .localhost names, which resolve to loopback.
//
// The host is compared exactly, so a real domain that merely contains the
// word "localhost" (https://my-localhost-blog.com) is allowed through — the
// old substring check would have rejected it and, worse, waved through
// nothing an attacker actually needed. The point of the check is to stop a
// production deploy from trusting a browser on the same box.
func isLocalhostOrigin(origin string) bool {
	host := originHost(origin)
	if host == "" {
		return false
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback() || ip.IsUnspecified()
	}
	return strings.EqualFold(host, "localhost") ||
		strings.HasSuffix(strings.ToLower(host), ".localhost")
}

// originHost extracts just the host from an origin such as
// "https://sub.example.com:8443", dropping scheme, port, path and userinfo.
// A bare "host:port" or host entry is accepted too, so a miswritten
// ALLOWED_ORIGINS is still judged on its host rather than skipped.
func originHost(origin string) string {
	trimmed := strings.TrimSpace(origin)
	if u, err := url.Parse(trimmed); err == nil && u.Hostname() != "" {
		return u.Hostname()
	}
	// Not a URL the parser recognises; fall back to treating the whole value
	// as a host, with or without a port.
	host := strings.Trim(trimmed, "[]")
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

// firstOf returns the first non-empty environment variable value from the provided keys.
func firstOf(keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return ""
}

func stringOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func listOr(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	var out []string
	for _, part := range strings.Split(raw, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func intOr(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback, errors.New("must be a whole number")
	}
	if v <= 0 {
		return fallback, errors.New("must be greater than zero")
	}
	return v, nil
}

func boolOr(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}

func durationOr(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback, errors.New(`must be a duration such as "45s" or "336h"`)
	}
	if d <= 0 {
		return fallback, errors.New("must be greater than zero")
	}
	return d, nil
}

func cleanAIModel(val, fallback string) string {
	val = strings.TrimSpace(val)
	if val == "" || strings.Contains(strings.ToLower(val), "gemini-3.8") {
		return fallback
	}
	return val
}

func loadDotEnv() {
	if len(os.Args) > 0 && (strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") || strings.Contains(os.Args[0], "go_build") || strings.Contains(os.Args[0], "__debug_bin")) {
		return
	}
	paths := []string{".env", "../.env", "../../.env"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
		break
	}
}

func (c Config) GoogleSignInEnabled() bool { return c.GoogleClientID != "" }
