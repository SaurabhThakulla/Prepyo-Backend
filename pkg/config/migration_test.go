package config

import "testing"

func TestMigrationConfiguration(t *testing.T) {
	setEnv(t, map[string]string{
		"APP_ENV": "development", "DATABASE_URL": "postgres://localhost/prepyo",
		"MIGRATION_DATABASE_URL": "postgres://localhost/prepyo_direct",
		"AUTO_MIGRATE":           "false",
	})
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AutoMigrate || cfg.MigrationDatabaseURL != "postgres://localhost/prepyo_direct" {
		t.Fatal("separate migration configuration not loaded")
	}
	t.Setenv("AUTO_MIGRATE", "")
	t.Setenv("MIGRATION_DATABASE_URL", "")
	cfg, err = Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AutoMigrate || cfg.MigrationDatabaseURL != "" {
		t.Fatal("legacy migration defaults changed")
	}
}
