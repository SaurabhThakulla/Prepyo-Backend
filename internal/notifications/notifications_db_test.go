package notifications

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/testdb"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), testdb.URL(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// newUser creates a user with the given role, plan and streak state, and
// removes it (and its notifications, by cascade) after the test.
func newUser(t *testing.T, pool *pgxpool.Pool, role, planID, extra string) string {
	t.Helper()
	ctx := context.Background()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, name, role, plan_id, timezone, referral_code)
		VALUES ('notify-' || gen_random_uuid() || '@test.local', 'Notify Test', $1, $2,
		        'Asia/Kathmandu', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id`, role, planID).Scan(&id)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if extra != "" {
		if _, err := pool.Exec(ctx, `UPDATE users SET `+extra+` WHERE id = $1`, id); err != nil {
			t.Fatalf("set up user: %v", err)
		}
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id); err != nil {
			t.Errorf("cleanup user: %v", err)
		}
	})
	return id
}

func countFor(t *testing.T, pool *pgxpool.Pool, userID, notifType string) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM notifications WHERE user_id = $1 AND type = $2`, userID, notifType).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestDedupeKeySendsOnce(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	user := newUser(t, pool, "suru", "free", "")

	p := CreateParams{UserID: user, Type: TypeLimit, Title: "Limit", Message: "m", DedupeKey: "limit:practice:2026-09-23"}
	first, err := Send(ctx, pool, p)
	if err != nil || first == "" {
		t.Fatalf("first send: id=%q err=%v", first, err)
	}
	second, err := Send(ctx, pool, p)
	if err != nil || second != "" {
		t.Fatalf("second send should be dropped: id=%q err=%v", second, err)
	}
	// Without a key, every notice goes out.
	p.DedupeKey = ""
	for i := 0; i < 2; i++ {
		if _, err := Send(ctx, pool, p); err != nil {
			t.Fatal(err)
		}
	}
	if got := countFor(t, pool, user, TypeLimit); got != 3 {
		t.Fatalf("got %d notices, want 3", got)
	}
}

func TestSendToAdminsReachesOnlyAdmins(t *testing.T) {
	pool := testPool(t)
	admin := newUser(t, pool, "admin", "free", "")
	learner := newUser(t, pool, "suru", "free", "")

	if err := SendToAdmins(context.Background(), pool, CreateParams{Type: TypeReport, Title: "New report", Message: "m"}); err != nil {
		t.Fatal(err)
	}
	if countFor(t, pool, admin, TypeReport) != 1 {
		t.Fatal("admin was not told")
	}
	if countFor(t, pool, learner, TypeReport) != 0 {
		t.Fatal("a learner received an admin notice")
	}
}

func TestScheduledNotices(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)

	ending := newUser(t, pool, "taiyari", "pro", "plan_valid_until = CURRENT_DATE + 2")
	farOff := newUser(t, pool, "taiyari", "pro", "plan_valid_until = CURRENT_DATE + 20")
	ended := newUser(t, pool, "taiyari", "pro", "plan_valid_until = CURRENT_DATE - 1")
	longGone := newUser(t, pool, "suru", "pro", "plan_valid_until = CURRENT_DATE - 30")
	// Practised yesterday in their own zone, not yet today.
	streaker := newUser(t, pool, "suru", "free",
		"streak_days = 4, streak_last_active_date = (now() AT TIME ZONE 'Asia/Kathmandu')::date - 1")

	for run := 0; run < 2; run++ {
		if _, err := repo.RunScheduled(ctx); err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
	}

	if countFor(t, pool, ending, TypePlan) != 1 {
		t.Error("plan ending in 2 days: want exactly one notice across two runs")
	}
	if countFor(t, pool, farOff, TypePlan) != 0 {
		t.Error("plan ending in 20 days was warned")
	}
	if countFor(t, pool, ended, TypePlan) != 1 {
		t.Error("plan that ended yesterday: want exactly one notice")
	}
	if countFor(t, pool, longGone, TypePlan) != 0 {
		t.Error("plan that ended a month ago was notified")
	}

	// The reminder only goes out from 6pm in the learner's zone.
	var evening bool
	if err := pool.QueryRow(ctx, `SELECT (now() AT TIME ZONE 'Asia/Kathmandu')::time >= '18:00'`).Scan(&evening); err != nil {
		t.Fatal(err)
	}
	want := 0
	if evening {
		want = 1
	}
	if got := countFor(t, pool, streaker, TypeStreak); got != want {
		t.Errorf("streak reminders = %d, want %d (evening=%v)", got, want, evening)
	}
}

func TestCleanupKeepsUnreadAndRecent(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	user := newUser(t, pool, "suru", "free", "")

	if _, err := pool.Exec(ctx, `
		INSERT INTO notifications (user_id, title, message, type, read, created_at) VALUES
		($1, 'old read', 'm', 'system', TRUE, now() - interval '100 days'),
		($1, 'old unread', 'm', 'system', FALSE, now() - interval '100 days'),
		($1, 'new read', 'm', 'system', TRUE, now())`, user); err != nil {
		t.Fatal(err)
	}
	if _, err := NewRepository(pool).RunScheduled(ctx); err != nil {
		t.Fatal(err)
	}
	if got := countFor(t, pool, user, TypeSystem); got != 2 {
		t.Fatalf("got %d left, want 2 (only the old read one removed)", got)
	}
}

// Deleting removes a notice from the inbox for good, only the owner can do it,
// and clearing a once-only notice does not let it be sent a second time.
func TestDeleteAndClear(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)
	owner := newUser(t, pool, "suru", "free", "")
	stranger := newUser(t, pool, "suru", "free", "")

	plain, err := Send(ctx, pool, CreateParams{UserID: owner, Type: TypeSystem, Title: "Plain", Message: "m"})
	if err != nil {
		t.Fatal(err)
	}
	limit := CreateParams{UserID: owner, Type: TypeLimit, Title: "Limit", Message: "m", DedupeKey: "limit:practice:today"}
	if _, err := Send(ctx, pool, limit); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(ctx, stranger, plain); err != ErrNotFound {
		t.Fatalf("another user deleted the notice: err=%v", err)
	}
	if err := repo.Delete(ctx, owner, "not-a-uuid"); err != ErrNotFound {
		t.Fatalf("malformed id: err=%v, want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, owner, plain); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, owner, plain); err != ErrNotFound {
		t.Fatalf("second delete: err=%v, want ErrNotFound", err)
	}

	removed, err := repo.DeleteAll(ctx, owner)
	if err != nil || removed != 1 {
		t.Fatalf("clear all: removed=%d err=%v, want 1", removed, err)
	}
	list, total, unread, err := repo.List(ctx, owner, 10, 0)
	if err != nil || len(list) != 0 || total != 0 || unread != 0 {
		t.Fatalf("inbox after clearing: %d listed, total=%d unread=%d err=%v", len(list), total, unread, err)
	}

	// The cleared limit notice keeps its key, so the same day's notice is not sent again.
	if id, err := Send(ctx, pool, limit); err != nil || id != "" {
		t.Fatalf("cleared once-only notice was sent again: id=%q err=%v", id, err)
	}
}
