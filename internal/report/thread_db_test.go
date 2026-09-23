package report

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/testdb"
)

func newUser(t *testing.T, pool *pgxpool.Pool, role string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, role, timezone, referral_code)
		VALUES ('report-' || gen_random_uuid() || '@test.local', 'Report Test', $1,
		        'Asia/Kathmandu', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id`, role).Scan(&id)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id)
	})
	return id
}

// A report is a conversation the learner owns: they see their own, never
// anyone else's, and the team hears about it when it arrives.
func TestReportConversation(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	learner := newUser(t, pool, "suru")
	other := newUser(t, pool, "suru")
	admin := newUser(t, pool, "admin")

	h := NewHandler(pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	id, err := h.store(ctx, learner, "Asha", "/practice", "Audio will not play")
	if err != nil {
		t.Fatal(err)
	}

	var adminTold int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE user_id = $1 AND action_url LIKE '%' || $2`,
		admin, id).Scan(&adminTold); err != nil {
		t.Fatal(err)
	}
	if adminTold != 1 {
		t.Fatalf("admin notices for the new report = %d, want 1", adminTold)
	}

	if err := AddReply(ctx, pool, id, admin, true, "Which browser are you using?"); err != nil {
		t.Fatal(err)
	}
	if err := AddReply(ctx, pool, id, learner, false, "Chrome on Android"); err != nil {
		t.Fatal(err)
	}

	rep, err := Load(ctx, pool, id, learner)
	if err != nil {
		t.Fatal(err)
	}
	if rep.ReplyCount != 2 || !rep.Replies[0].FromStaff || rep.Replies[1].FromStaff {
		t.Fatalf("conversation out of order or miscounted: %+v", rep.Replies)
	}
	if rep.Replies[0].AuthorName != "Prepyo team" {
		t.Fatalf("staff reply signed %q, want the team name", rep.Replies[0].AuthorName)
	}

	if _, err := Load(ctx, pool, id, other); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another learner loaded the report: err=%v", err)
	}
}

func TestClipCountsCharactersNotBytes(t *testing.T) {
	long := ""
	for i := 0; i < MaxMessageRunes+10; i++ {
		long += "क"
	}
	if got := []rune(Clip(long)); len(got) != MaxMessageRunes {
		t.Fatalf("clipped to %d characters, want %d", len(got), MaxMessageRunes)
	}
}
