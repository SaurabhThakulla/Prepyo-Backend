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

func TestLoadDevelopmentDefaults(t *testing.T) {
	setEnv(t, map[string]string{"DATABASE_URL": "postgres://localhost/prepyo"})

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
	if cfg.SessionSecret == "" {
		t.Error("SessionSecret is empty; development should get a usable fallback")
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
			name: "missing session secret",
			env: map[string]string{
				"SESSION_SECRET":     "",
				"OPENROUTER_API_KEY": "key",
			},
			wantErr: "SESSION_SECRET",
		},
		{
			name: "short session secret",
			env: map[string]string{
				"SESSION_SECRET":     "too-short",
				"OPENROUTER_API_KEY": "key",
			},
			wantErr: "at least 32 characters",
		},
		{
			name: "missing ai key",
			env: map[string]string{
				"SESSION_SECRET":     strings.Repeat("a", 40),
				"OPENROUTER_API_KEY": "",
			},
			wantErr: "OPENROUTER_API_KEY",
		},
		{
			name: "localhost origin",
			env: map[string]string{
				"SESSION_SECRET":     strings.Repeat("a", 40),
				"OPENROUTER_API_KEY": "key",
				"ALLOWED_ORIGINS":    "https://prepyo.np,http://localhost:3000",
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
		"APP_ENV":            "production",
		"DATABASE_URL":       "postgres://db/prepyo",
		"ALLOWED_ORIGINS":    "https://prepyo.np",
		"SESSION_SECRET":     strings.Repeat("a", 40),
		"OPENROUTER_API_KEY": "key",
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
		"APP_ENV":            "production",
		"DATABASE_URL":       "",
		"SESSION_SECRET":     "",
		"OPENROUTER_API_KEY": "",
		"ALLOWED_ORIGINS":    "https://prepyo.np",
	})

	_, err := Load()
	if err == nil {
		t.Fatal("Load() succeeded, want an error")
	}
	for _, want := range []string{"DATABASE_URL", "SESSION_SECRET", "OPENROUTER_API_KEY"} {
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
