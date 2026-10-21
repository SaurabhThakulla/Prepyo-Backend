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
	// Defaults to LocalProxyCIDRs.
	TrustedProxyCIDRs []string
	WebAppURL         string

	DatabaseURL string
	AutoMigrate bool

	SessionTTL time.Duration
	// SecureCookies must be true anywhere the app is served over HTTPS.
	SecureCookies bool

	// AI providers.
	AIBaseURL      string
	AIAPIKey       string
	AIAudioBaseURL string
	AIAudioAPIKey  string
	// SpeakingEvaluated reports whether a recording can actually be scored.
	// False means the product still records and plays back answers; it just
	// does not pretend to grade them.
	SpeakingEvaluated bool

	// AITTSModel and AITTSVoices drive server-side speech on the audio
	// provider (Orpheus on OpenRouter by default), used when a browser cannot speak a
	// listening script itself. Voices are handed out one per speaker.
	AITTSModel  string
	AITTSVoices []string

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

	// BotAPIKey opens /api/v1/bot: the payment and report queues, and nothing
	// else from the admin area, for the WhatsApp admin bot. Empty keeps those
	// routes off entirely.
	BotAPIKey string
	// BotActorEmail is the admin account the bot's reviews and replies are
	// recorded under. Defaults to AdminEmail.
	BotActorEmail string
	// BotAllowedCIDRs, when set, are the only addresses the bot routes answer.
	BotAllowedCIDRs []string

	// WebDistDir is the optional directory path to frontend production build assets.
	WebDistDir string

	// S3* point at the private Backblaze B2 bucket that holds learners' media.
	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string
	// MediaDir is a local folder used instead of B2, for development.
	MediaDir string
}

// S3Enabled reports whether an S3-compatible bucket is fully configured.
func (c Config) S3Enabled() bool {
	return c.S3Endpoint != "" && c.S3Bucket != "" && c.S3AccessKeyID != "" && c.S3SecretAccessKey != ""
}

// BotAPIEnabled reports whether the bot routes should be mounted at all.
func (c Config) BotAPIEnabled() bool { return c.BotAPIKey != "" }

// AdminLoginEnabled reports whether the admin password endpoint has a
// credential to check against. When false it returns 503 rather than comparing
// against an empty password and letting anyone in.
func (c Config) AdminLoginEnabled() bool { return c.AdminEmail != "" && c.AdminPassword != "" }

// DefaultAIModel and DefaultSpeechModel are the OpenRouter models used when
// AI_MODEL_* is not set: Haiku for every text task, Gemini for recordings.
const (
	DefaultAIModel     = "anthropic/claude-haiku-5.5"
	// DefaultSpeechModel transcribes recordings; Whisper is no longer used.
	DefaultSpeechModel = "google/gemini-2.5-flash-lite"
	// DefaultDeliveryModel hears Read Aloud / Repeat Sentence audio for Pronunciation and Oral Fluency.
	// Gemini 3.7 Flash separated good and weak readings best of 6 audio models (2026-10-21), at 2.5 Flash's cost.
	DefaultDeliveryModel = "google/gemini-3.7-flash"
)

// AIModels defines model names for each AI task capability.
type AIModels struct {
	Writing  string
	Speaking string
	Tutoring string
	// Delivery rates pronunciation and fluency from audio; "off" turns it off.
	Delivery string
}

func (c Config) IsProduction() bool { return c.Env == "production" }

// AIEnabled reports whether the AI provider is configured.
func (c Config) AIEnabled() bool { return c.AIAPIKey != "" }

// SpeakingEnabled reports whether a recording can be sent to a provider that
// accepts audio. Transcript-based speaking needs only the text provider, so it
// works even when this is false.
func (c Config) SpeakingEnabled() bool { return c.SpeakingEvaluated && c.AIAudioAPIKey != "" }

// WorkableAIMaxTokens is the smallest completion cap a full evaluation fits in.
//
// One IELTS evaluation is a summary, three or four criteria with feedback,
// strengths, weaknesses and sentence-level corrections: measured live, between
// 2,500 and 4,000 completion tokens on reasoning models. Capped below that, the
// provider returns a truncated reply, the gateway rejects it as invalid, and
// every evaluation fails with nothing stored - which looks like a broken
// product rather than a one-line setting.
const WorkableAIMaxTokens = 4000

// AIMaxTokensTooLow reports whether the configured cap will truncate ordinary
// evaluations.
func (c Config) AIMaxTokensTooLow() bool { return c.AIMaxTokens < WorkableAIMaxTokens }

// LocalProxyCIDRs are the addresses a reverse proxy on the same machine or
// private network connects from: nginx on the host reaches the app through
// loopback or Docker's bridge. Nothing on the internet can send from these, so
// trusting them by default lets the app see each visitor's real IP behind such
// a proxy with no configuration, while a visitor connecting directly still
// cannot forge X-Forwarded-For. Set TRUSTED_PROXY_CIDRS to add a CDN or an
// external load balancer.
var LocalProxyCIDRs = []string{
	"127.0.0.0/8",
	"::1/128",
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"fc00::/7",
}

// MinBotAPIKeyLength is the shortest BOT_API_KEY the server will start with.
const MinBotAPIKeyLength = 32

func Load() (*Config, error) {
	loadDotEnv()

	var problems []string

	env := stringOr("APP_ENV", "development")
	if env != "development" && env != "staging" && env != "production" {
		problems = append(problems, `APP_ENV must be development, staging or production`)
	}
	isProd := env == "production"

	aiBaseURL := strings.TrimRight(stringOr("AI_BASE_URL", "https://openrouter.ai/api/v1"), "/")
	aiKey := stringOr("AI_API_KEY", "")
	aiAudioKey := stringOr("AI_AUDIO_API_KEY", aiKey)

	// Scoring a recording needs a provider that takes audio, and inheriting the
	// text provider's key does not make it one. Deployed without AI_AUDIO_*, the
	// old fallback sent recordings to the text endpoint, which rejected the
	// model id after uploading the whole recording — three times, until the
	// request ran out of deadline and the learner was told we had broken.
	//
	// Set AI_SPEAKING_ENABLED explicitly when the text provider does serve an
	// audio-capable model on the same key.
	// OpenRouter serves Gemini audio models on the same key, so it counts as an audio provider.
	audioConfigured := stringOr("AI_AUDIO_BASE_URL", "") != "" || stringOr("AI_AUDIO_API_KEY", "") != "" ||
		strings.Contains(aiBaseURL, "openrouter.ai")

	cfg := &Config{
		Env:               env,
		Port:              stringOr("PORT", "8080"),
		AllowedOrigins:    listOr("ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		TrustedProxyCIDRs: listOr("TRUSTED_PROXY_CIDRS", LocalProxyCIDRs),
		WebAppURL:         stringOr("WEB_APP_URL", "http://localhost:3000"),

		AIBaseURL: aiBaseURL,
		AIAPIKey:  aiKey,

		AIAudioBaseURL:    strings.TrimRight(stringOr("AI_AUDIO_BASE_URL", aiBaseURL), "/"),
		AIAudioAPIKey:     aiAudioKey,
		SpeakingEvaluated: boolOr("AI_SPEAKING_ENABLED", audioConfigured),
		AITTSModel:        stringOr("AI_TTS_MODEL", "canopylabs/orpheus-3b-0.1-ft"),
		AITTSVoices:       listOr("AI_TTS_VOICES", []string{"tara", "leo", "leah", "dan", "jess", "zac"}),

		AIModels: AIModels{
			Writing:  cleanAIModel(stringOr("AI_MODEL_WRITING", DefaultAIModel), DefaultAIModel),
			Speaking: cleanSpeechModel(stringOr("AI_MODEL_SPEAKING", DefaultSpeechModel)),
			Tutoring: cleanAIModel(stringOr("AI_MODEL_TUTORING", DefaultAIModel), DefaultAIModel),
			Delivery: cleanAIModel(stringOr("AI_MODEL_DELIVERY", DefaultDeliveryModel), DefaultDeliveryModel),
		},
		SecureCookies: boolOr("SECURE_COOKIES", isProd),

		GoogleClientID: strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")),

		AdminEmail:    strings.ToLower(stringOr("ADMIN_EMAIL", "admin@prepyo.online")),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),

		BotAPIKey:       strings.TrimSpace(os.Getenv("BOT_API_KEY")),
		BotActorEmail:   strings.ToLower(stringOr("BOT_ACTOR_EMAIL", stringOr("ADMIN_EMAIL", "admin@prepyo.online"))),
		BotAllowedCIDRs: listOr("BOT_ALLOWED_CIDRS", nil),

		WebDistDir: os.Getenv("WEB_DIST_DIR"),

		S3Endpoint:        strings.TrimSpace(os.Getenv("S3_ENDPOINT")),
		S3Region:          strings.TrimSpace(os.Getenv("S3_REGION")),
		S3Bucket:          strings.TrimSpace(os.Getenv("S3_BUCKET")),
		S3AccessKeyID:     strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID")),
		S3SecretAccessKey: strings.TrimSpace(os.Getenv("S3_SECRET_ACCESS_KEY")),
		MediaDir:          strings.TrimSpace(os.Getenv("MEDIA_DIR")),
	}

	// A short key on routes that can approve a payment is worse than no routes.
	if cfg.BotAPIKey != "" && len(cfg.BotAPIKey) < MinBotAPIKeyLength {
		problems = append(problems, fmt.Sprintf("BOT_API_KEY must be at least %d characters", MinBotAPIKeyLength))
	}
	for _, prefix := range cfg.BotAllowedCIDRs {
		if _, _, err := net.ParseCIDR(prefix); err != nil {
			problems = append(problems, "BOT_ALLOWED_CIDRS contains an invalid CIDR: "+prefix)
		}
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
	cfg.AutoMigrate = boolOr("AUTO_MIGRATE", true)
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

// cleanSpeechModel keeps an old Whisper setting from breaking speaking: Whisper is not a chat model, so Gemini is used.
func cleanSpeechModel(val string) string {
	if strings.Contains(strings.ToLower(val), "whisper") {
		return DefaultSpeechModel
	}
	return cleanAIModel(val, DefaultSpeechModel)
}

func loadDotEnv() {
	if len(os.Args) > 0 && (strings.HasSuffix(os.Args[0], ".test") || strings.HasSuffix(os.Args[0], ".test.exe") || strings.Contains(os.Args[0], "go_build") || strings.Contains(os.Args[0], "__debug_bin")) {
		return
	}
	candidates := [][]string{
		{".env.local", "../.env.local", "../../.env.local"},
		{".env", "../.env", "../../.env"},
	}
	for _, paths := range candidates {
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
}

func (c Config) GoogleSignInEnabled() bool { return c.GoogleClientID != "" }
