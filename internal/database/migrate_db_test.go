package database

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/testdb"
)

func TestMigratorAppliesAllCleanly(t *testing.T) {
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := fmt.Sprintf("prepyo_restore_%d_test", time.Now().UnixNano())
	quoted := pgx.Identifier{name}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(ctx, "DROP DATABASE "+quoted+" WITH (FORCE)"); err != nil {
			t.Error(err)
		}
	}()
	cfg := admin.Config()
	cfg.ConnConfig.Database = name
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err = Migrate(ctx, pool, log); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, pool, log); err != nil {
		t.Fatalf("second migration run: %v", err)
	}

	var questionsCount int
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM questions`).Scan(&questionsCount); err != nil {
		t.Fatal(err)
	}
	if questionsCount != 0 {
		t.Fatalf("expected 0 seeded questions, got %d", questionsCount)
	}

	if err = CheckSeedIntegrity(ctx, pool); err != nil {
		t.Fatalf("seed integrity check failed: %v", err)
	}

	// Delete a core plan and verify integrity check fails
	if _, err = pool.Exec(ctx, `DELETE FROM plans WHERE id='free'`); err != nil {
		t.Fatal(err)
	}
	if err = CheckSeedIntegrity(ctx, pool); err == nil {
		t.Fatal("missing plan was not detected by CheckSeedIntegrity")
	}
}
