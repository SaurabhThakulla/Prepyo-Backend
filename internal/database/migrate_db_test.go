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
	"github.com/prepyo/backend/migrations"
)

func TestMigratorRepairsMissingContent(t *testing.T) {
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
	if _, err = pool.Exec(ctx, `CREATE TABLE schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		t.Fatal(err)
	}
	files, err := pendingFiles(map[string]bool{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range files {
		if name >= "000050" {
			continue
		}
		body, err := migrations.Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("%s: %v", name, err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = pool.Exec(ctx, `DELETE FROM questions WHERE skill IN ('writing','speaking') AND id NOT LIKE 'pte-wrt-swt-%'`); err != nil {
		t.Fatal(err)
	}
	if err = CheckSeedIntegrity(ctx, pool); err == nil {
		t.Fatal("simulated data loss was not detected")
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err = Migrate(ctx, pool, log); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, pool, log); err != nil {
		t.Fatalf("second migration run: %v", err)
	}
	var writing, speaking, recorded int
	if err = pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE skill='writing'), count(*) FILTER(WHERE skill='speaking') FROM questions`).Scan(&writing, &speaking); err != nil {
		t.Fatal(err)
	}
	// Includes the three original IELTS Speaking pilot tasks from migration 000055.
	if writing != 241 || speaking != 55 {
		t.Fatalf("writing=%d speaking=%d", writing, speaking)
	}
	if err = pool.QueryRow(ctx, `SELECT count(*) FROM schema_migrations WHERE version='000050_restore_seed_content.up.sql'`).Scan(&recorded); err != nil || recorded != 1 {
		t.Fatalf("recorded=%d err=%v", recorded, err)
	}
	if _, err = pool.Exec(ctx, `DELETE FROM questions WHERE id='pte-spk-001'`); err != nil {
		t.Fatal(err)
	}
	if err = Migrate(ctx, pool, log); err == nil {
		t.Fatal("up-to-date migration history hid missing content")
	}
}
