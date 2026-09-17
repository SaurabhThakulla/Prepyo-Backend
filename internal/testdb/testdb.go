// Package testdb prevents integration tests from accidentally using application data.
package testdb

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// URL never falls back to DATABASE_URL or .env. Tests require an explicit,
// loopback-only database whose name ends in _test. No remote override exists.
func URL(t testing.TB) string {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL not set; use a disposable local *_test database")
	}
	if err := Validate(raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func Validate(raw string) error {
	cfg, err := pgxpool.ParseConfig(raw)
	if err != nil {
		return fmt.Errorf("invalid TEST_DATABASE_URL (details suppressed to protect credentials)")
	}
	local := func(host string) bool {
		ip := net.ParseIP(host)
		return host == "localhost" || (ip != nil && ip.IsLoopback())
	}
	if !local(cfg.ConnConfig.Host) || cfg.ConnConfig.Port == 6543 || !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		return fmt.Errorf("refusing unsafe test database: require loopback host, non-pooler port and database name ending in _test")
	}
	for _, fallback := range cfg.ConnConfig.Fallbacks {
		if !local(fallback.Host) || fallback.Port == 6543 {
			return fmt.Errorf("refusing remote or pooler test fallback")
		}
	}
	return nil
}
