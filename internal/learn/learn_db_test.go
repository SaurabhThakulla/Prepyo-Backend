package learn

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/testdb"
)

// A learner's progress is stored once, replaced on each save, and removed
// with the learner.
func TestProgressRoundTrip(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := database.Migrate(ctx, pool, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var userID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, name, plan_id, timezone, referral_code)
		VALUES ('learn-' || gen_random_uuid() || '@test.local', 'Learn Test', 'free', 'Asia/Kathmandu',
		        'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("create learner: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID) })

	repo := NewRepository(pool)
	if data, err := repo.Load(ctx, userID); err != nil || data != nil {
		t.Fatalf("before any save: %s, %v; want nothing", data, err)
	}

	for _, xp := range []int{15, 40} {
		saved, _ := json.Marshal(map[string]any{"xp": xp, "completed": map[string]any{}})
		if err := repo.Save(ctx, userID, saved); err != nil {
			t.Fatal(err)
		}
		data, err := repo.Load(ctx, userID)
		if err != nil {
			t.Fatal(err)
		}
		var got struct{ XP int }
		if err := json.Unmarshal(data, &got); err != nil || got.XP != xp {
			t.Fatalf("after saving xp %d: loaded %s (%v)", xp, data, err)
		}
	}

	if _, err := pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID); err != nil {
		t.Fatal(err)
	}
	var left int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM learn_progress WHERE user_id = $1`, userID).Scan(&left); err != nil || left != 0 {
		t.Fatalf("progress left after the learner was deleted: %d (%v)", left, err)
	}
}
