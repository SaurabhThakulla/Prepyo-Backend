package database

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/testdb"
	"github.com/prepyo/backend/migrations"
)

func TestRestoreSeedContent(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `CREATE TEMP TABLE questions (LIKE public.questions INCLUDING ALL) ON COMMIT DROP`); err != nil {
		t.Fatal(err)
	}
	apply := func(name string) {
		t.Helper()
		body, err := migrations.Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
			t.Fatal(err)
		}
	}
	if err = CheckSeedIntegrity(ctx, tx); err == nil {
		t.Fatal("empty bank passed integrity check")
	}
	apply("000049_writing_content_repairs.up.sql")
	apply("000050_restore_seed_content.up.sql")
	if err = CheckSeedIntegrity(ctx, tx); err != nil {
		t.Fatal(err)
	}
	var writing, speaking int
	if err = tx.QueryRow(ctx, `SELECT count(*) FILTER (WHERE skill='writing'), count(*) FILTER (WHERE skill='speaking') FROM questions`).Scan(&writing, &speaking); err != nil {
		t.Fatal(err)
	}
	if writing != 216 || speaking != 2 {
		t.Fatalf("writing=%d speaking=%d, want 216/2", writing, speaking)
	}
	for _, tc := range []struct{ id, column, want string }{
		{"ielts-wrt-fg-001", "figure_data", "Mexico has the largest absolute increase"},
		{"ielts-wrt-fg-003", "figure_data", "become the largest source"},
		{"ielts-wrt-fg-004", "figure_data", "under a third of its 2015 figure"},
		{"ielts-wrt-fg-009", "figure_data", "no income data"},
		{"pte-wrt-001", "model_answer", "permeable surfaces"},
		{"ielts-wrt-001", "type_id", "ielts-writing-task2-opinion"},
	} {
		var value string
		if err = tx.QueryRow(ctx, "SELECT "+tc.column+" FROM questions WHERE id=$1", tc.id).Scan(&value); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(value, tc.want) {
			t.Errorf("%s: missing correction %q", tc.id, tc.want)
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE questions SET title='Editor preserved', figure_data='Editor guidance' WHERE id='ielts-wrt-fg-003'`); err != nil {
		t.Fatal(err)
	}
	snapshot := func() string {
		t.Helper()
		var value string
		if err := tx.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(q) ORDER BY id)::text FROM questions q`).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := snapshot()
	apply("000050_restore_seed_content.up.sql")
	apply("000050_restore_seed_content.down.sql")
	apply("000050_restore_seed_content.up.sql")
	if snapshot() != before {
		t.Fatal("replay or down changed pre-existing rows")
	}
	if _, err = tx.Exec(ctx, `DELETE FROM questions WHERE id='pte-spk-001'`); err != nil {
		t.Fatal(err)
	}
	if err = CheckSeedIntegrity(ctx, tx); err == nil || !strings.Contains(err.Error(), "pte-spk-001") {
		t.Fatalf("missing ID not detected: %v", err)
	}
	apply("000050_restore_seed_content.up.sql")
	if err = CheckSeedIntegrity(ctx, tx); err != nil {
		t.Fatal(err)
	}
}
