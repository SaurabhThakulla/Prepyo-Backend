// Package config loads settings from the environment.
//
// Load returns an error listing every problem it found rather than stopping at
// the first one, so a misconfigured deploy tells you everything that is wrong
// in a single boot attempt.
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

	OpenRouterAPIKey string
	AIModels         AIModels
	AIRequestTimeout time.Duration

	// Issue reporting. SMTPUser doubles as the From address.
	SMTPUser      string
	SMTPPassword  string
	ReportEmailTo string

	// WebDistDir is the frontend's build output. Set it and this binary serves
	// the app and the API from one origin; leave it empty and it is API-only,
	// which is what you want when running against `vite dev`.
	WebDistDir string
}

// ReportingEnabled reports whether /report has credentials to send with. When
// false the endpoint says so outright rather than accepting a report and
// dropping it: a learner who has typed out a bug deserves to know it went
// nowhere.
func (c Config) ReportingEnabled() bool { return c.SMTPUser != "" && c.SMTPPassword != "" }

// AIModels is the routing table the AI gateway uses. Model names are
// configuration, never constants in the calling code, so a model can be
// swapped without a redeploy of business logic.
type AIModels struct {
	Writing  string
	Speaking string
	Tutoring string
}

func (c Config) IsProduction() bool { return c.Env == "production" }

// AIEnabled reports whether the gateway has what it needs to reach a provider.
// When false the AI endpoints return a clear "unavailable" error instead of
// inventing a score.
func (c Config) AIEnabled() bool { return c.OpenRouterAPIKey != "" }

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
		RedisURL:         os.Getenv("REDIS_URL"),
		OpenRouterAPIKey: os.Getenv("OPENROUTER_API_KEY"),
		AIModels: AIModels{
			Writing: stringOr("AI_MODEL_WRITING", "deepseek/deepseek-chat"),
			// Speaking sends a recording, so this one must be a model that
			// accepts audio input. A text-only model here does not degrade
			// gracefully: it rejects the request and every speaking submission
			// comes back as "evaluation unavailable".
			Speaking: stringOr("AI_MODEL_SPEAKING", "google/gemini-2.5-flash"),
			Tutoring: stringOr("AI_MODEL_TUTORING", "deepseek/deepseek-chat"),
		},
		SecureCookies: boolOr("SECURE_COOKIES", isProd),

		// Not required anywhere, including production: reporting is a
		// convenience, and a missing app password should not stop the API from
		// booting and serving lessons.
		SMTPUser:      os.Getenv("GMAIL_USER"),
		SMTPPassword:  os.Getenv("GMAIL_APP_PASSWORD"),
		ReportEmailTo: stringOr("REPORT_EMAIL_TO", "sauravthakulla683@gmail.com"),

		WebDistDir: os.Getenv("WEB_DIST_DIR"),
	}

	// A path that is set but wrong is worth stopping the boot for: the
	// alternative is an API that looks healthy while every page request 404s.
	if cfg.WebDistDir != "" {
		if _, err := os.Stat(filepath.Join(cfg.WebDistDir, "index.html")); err != nil {
			problems = append(problems, "WEB_DIST_DIR is set but has no index.html in it: "+cfg.WebDistDir)
		}
	}

	// DATABASE_URL is required everywhere: there is no in-memory fallback, by
	// design. A missing database should stop the boot, not quietly serve
	// blank data.
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		problems = append(problems, "DATABASE_URL is required")
	}

	cfg.SessionSecret = os.Getenv("SESSION_SECRET")
	switch {
	case cfg.SessionSecret == "" && isProd:
		problems = append(problems, "SESSION_SECRET is required in production")
	case cfg.SessionSecret == "":
		// Development only. Restarting the server invalidates existing
		// sessions, which is a fair trade for not needing setup to run locally.
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

	if isProd {
		if cfg.OpenRouterAPIKey == "" {
			problems = append(problems, "OPENROUTER_API_KEY is required in production")
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


