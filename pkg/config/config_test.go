package config

import (
	"strings"
	"testing"
	"time"
)

// setEnv applies a set of variables for one test and restores the environment
// afterwards, so tests can run in any order.
func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for key, value := range vars {
		t.Setenv(key, value)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	setEnv(t, map[string]string{"DATABASE_URL": ""})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() succeeded without DATABASE_URL, want an error")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("error = %q, want it to name DATABASE_URL", err)
	}
}

func TestLoadAudioKeyFallsBackToAIKey(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":           "development",
		"DATABASE_URL":      "postgres://localhost/prepyo",
		"AI_API_KEY":        "test-text-api-key",
		"CODE_CRAFT":        "",
		"AI_AUDIO_API_KEY":  "",
		"test-text-api-key": "",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cfg.AIAudioAPIKey != cfg.AIAPIKey {
		t.Error("AIAudioAPIKey does not fall back to AIAPIKey")
	}
	// The key falls back so a single-provider setup keeps working, but an
	// inherited key does not make the text endpoint able to hear a recording.
	// Treating it as though it did is what sent recordings to a provider that
	// answered "that model does not exist" after the whole upload.
	if cfg.SpeakingEnabled() {
		t.Error("SpeakingEnabled() = true with no audio provider of its own")
	}
}

func TestLoadDevelopmentDefaults(t *testing.T) {
	setEnv(t, map[string]string{"DATABASE_URL": "postgres://localhost/prepyo"})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
	if cfg.SecureCookies {
		t.Error("SecureCookies = true in development, want false so local HTTP works")
	}
	if cfg.AIEnabled() {
		t.Error("AIEnabled() = true without an API key")
	}
}

// Production must not boot with development-grade settings.
func TestLoadProductionRequirements(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name: "missing ai key",
			env: map[string]string{
				"AI_API_KEY": "",
			},
			wantErr: "AI_API_KEY",
		},
		{
			name: "localhost origin",
			env: map[string]string{
				"AI_API_KEY":      "key",
				"ALLOWED_ORIGINS": "https://prepyo.np,http://localhost:3000",
			},
			wantErr: "localhost",
		},
		{
			name: "loopback ip origin",
			env: map[string]string{
				"AI_API_KEY":      "key",
				"ALLOWED_ORIGINS": "https://prepyo.np,http://127.0.0.1:3000",
			},
			wantErr: "localhost",
		},
		{
			name: "ipv6 loopback origin",
			env: map[string]string{
				"AI_API_KEY":      "key",
				"ALLOWED_ORIGINS": "http://[::1]:8080",
			},
			wantErr: "localhost",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := map[string]string{
				"APP_ENV":         "production",
				"DATABASE_URL":    "postgres://db/prepyo",
				"ALLOWED_ORIGINS": "https://prepyo.np",
			}
			for k, v := range tc.env {
				base[k] = v
			}
			setEnv(t, base)

			_, err := Load()
			if err == nil {
				t.Fatalf("Load() succeeded, want an error mentioning %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to mention %q", err, tc.wantErr)
			}
		})
	}
}

func TestLoadProductionSucceedsWhenConfigured(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          "production",
		"DATABASE_URL":     "postgres://db/prepyo",
		"ALLOWED_ORIGINS":  "https://prepyo.np",
		"AI_API_KEY":       "key",
		"GOOGLE_CLIENT_ID": "prepyo.apps.googleusercontent.com",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if !cfg.SecureCookies {
		t.Error("SecureCookies = false in production, want true")
	}
	if !cfg.IsProduction() {
		t.Error("IsProduction() = false")
	}
}

// Every problem should be reported at once, so one boot attempt tells you
// everything that needs fixing.
func TestLoadReportsAllProblemsTogether(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":         "production",
		"DATABASE_URL":    "",
		"AI_API_KEY":      "",
		"ALLOWED_ORIGINS": "https://prepyo.np",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() succeeded, want an error")
	}
	for _, want := range []string{"DATABASE_URL", "AI_API_KEY"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q; got:\n%s", want, err)
		}
	}
}

func TestDurationParsing(t *testing.T) {
	t.Run("valid override", func(t *testing.T) {
		setEnv(t, map[string]string{
			"DATABASE_URL": "postgres://localhost/prepyo",
			"SESSION_TTL":  "1h",
		})
		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}
		if cfg.SessionTTL != time.Hour {
			t.Errorf("SessionTTL = %v, want 1h", cfg.SessionTTL)
		}
	})

	t.Run("nonsense value is rejected", func(t *testing.T) {
		setEnv(t, map[string]string{
			"DATABASE_URL": "postgres://localhost/prepyo",
			"SESSION_TTL":  "fourteen days",
		})
		if _, err := Load(); err == nil {
			t.Error("Load() succeeded with an unparseable SESSION_TTL, want an error")
		}
	})
}

func TestCleanAIModel(t *testing.T) {
	cases := []struct {
		input    string
		fallback string
		want     string
	}{
		{"gemini-3.8-flash-high", "gpt-5.6-luna", "gpt-5.6-luna"},
		{"GEMINI-3.8", "gpt-5.6-luna", "gpt-5.6-luna"},
		{"", "default-model", "default-model"},
		{"  ", "default-model", "default-model"},
		{"gpt-5.6-luna", "default-model", "gpt-5.6-luna"},
		{"gemini-3.7-flash", "default-model", "gemini-3.7-flash"},
	}

	for _, tc := range cases {
		if got := cleanAIModel(tc.input, tc.fallback); got != tc.want {
			t.Errorf("cleanAIModel(%q, %q) = %q, want %q", tc.input, tc.fallback, got, tc.want)
		}
	}
}

func TestIsLocalhostOrigin(t *testing.T) {
	cases := []struct {
		origin string
		want   bool
	}{
		{"http://localhost:3000", true},
		{"https://localhost", true},
		{"http://127.0.0.1:5173", true},
		{"http://127.0.0.1", true},
		{"http://[::1]:8080", true},
		{"http://0.0.0.0:3000", true},
		{"https://dev.localhost", true},
		{"localhost:3000", true},
		{"https://prepyo.online", false},
		{"https://localhost.attacker.com", false},
		{"http://evil-localhost.com", false},
		{"https://prepyo.np/localhost", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := isLocalhostOrigin(tc.origin); got != tc.want {
			t.Errorf("isLocalhostOrigin(%q) = %v, want %v", tc.origin, got, tc.want)
		}
	}
}

// A domain that merely contains the word "localhost" is a real origin and must
// pass the production boot check; only actual loopback hosts are rejected.
func TestLoadProductionAllowsLocalhostSubstrings(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":          "production",
		"DATABASE_URL":     "postgres://db/prepyo",
		"ALLOWED_ORIGINS":  "https://prepyo.np,https://my-localhost-blog.com",
		"AI_API_KEY":       "key",
		"GOOGLE_CLIENT_ID": "prepyo.apps.googleusercontent.com",
	})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Errorf("AllowedOrigins = %v, want both entries kept", cfg.AllowedOrigins)
	}
}
