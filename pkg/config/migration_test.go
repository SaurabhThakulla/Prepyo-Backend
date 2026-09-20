package config

import "testing"

func TestMigrationConfiguration(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV":      "development",
		"DATABASE_URL": "postgres://localhost/prepyo",
		"AUTO_MIGRATE": "false",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AutoMigrate {
		t.Fatal("expected AutoMigrate to be false")
	}

	t.Setenv("AUTO_MIGRATE", "")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AutoMigrate {
		t.Fatal("expected default AutoMigrate to be true")
	}
}
