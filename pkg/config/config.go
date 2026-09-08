// Package config loads application configuration from environment variables.
package config

import (
	"errors"
	"fmt"
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
	WebAppURL      string

	DatabaseURL string
	RedisURL    string

	// SessionSecret signs session cookies. Rotating it logs everyone out.
	SessionSecret string
	SessionTTL    time.Duration
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

	// Issue reporting. SMTPUser doubles as the From address.
	SMTPUser      string
	SMTPPassword  string
	ReportEmailTo string

	// WebDistDir is the optional directory path to frontend production build assets.
	WebDistDir string
}

// ReportingEnabled reports whether issue reporting credentials are configured.
func (c Config) ReportingEnabled() bool { return c.SMTPUser != "" && c.SMTPPassword != "" }

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

	cfg := &Config{
		Env:              env,
		Port:             stringOr("PORT", "8080"),
		AllowedOrigins:   listOr("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		WebAppURL:        stringOr("WEB_APP_URL", "http://localhost:3000"),
		RedisURL: os.Getenv("REDIS_URL"),

		AIBaseURL: strings.TrimRight(stringOr("AI_BASE_URL", "https://codecraftapi.com/v1"), "/"),
		AIAPIKey: firstOf("AI_API_KEY", "CODE_CRAFT"),

		AIAudioBaseURL: strings.TrimRight(stringOr("AI_AUDIO_BASE_URL", "https://openrouter.ai/api/v1"), "/"),
		AIAudioAPIKey:  firstOf("AI_AUDIO_API_KEY", "OPENROUTER_API_KEY"),

		AIModels: AIModels{
			Writing: stringOr("AI_MODEL_WRITING", "gpt-5.6-luna"),
			Speaking: stringOr("AI_MODEL_SPEAKING", "google/gemini-2.5-flash"),
			Tutoring: stringOr("AI_MODEL_TUTORING", "gpt-5.6-luna"),
		},
		SecureCookies: boolOr("SECURE_COOKIES", isProd),

		GoogleClientID: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),

		AdminEmail:    strings.ToLower(stringOr("ADMIN_EMAIL", "admin@prepyo.online")),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),

		SMTPUser:      os.Getenv("GMAIL_USER"),
		SMTPPassword:  os.Getenv("GMAIL_APP_PASSWORD"),
		ReportEmailTo: stringOr("REPORT_EMAIL_TO", "sauravthakulla683@gmail.com"),

		WebDistDir: os.Getenv("WEB_DIST_DIR"),
	}

	if cfg.WebDistDir != "" {
		if _, err := os.Stat(filepath.Join(cfg.WebDistDir, "index.html")); err != nil {
			problems = append(problems, "WEB_DIST_DIR is set but has no index.html in it: "+cfg.WebDistDir)
		}
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	}

	cfg.SessionSecret = os.Getenv("SESSION_SECRET")
	switch {
	case cfg.SessionSecret == "" && isProd:
		problems = append(problems, "SESSION_SECRET is required in production")
	case cfg.SessionSecret == "":
		cfg.SessionSecret = "dev-only-insecure-session-secret"
	case len(cfg.SessionSecret) < 32:
		problems = append(problems, "SESSION_SECRET must be at least 32 characters")
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
			if strings.Contains(origin, "localhost") {
				problems = append(problems, "ALLOWED_ORIGINS must not contain localhost in production")
				break
			}
		}
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return cfg, nil
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
