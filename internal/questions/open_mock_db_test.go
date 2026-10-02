package questions

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/testdb"
)

// A question's script is withheld only while it sits in a mock the same
// learner is taking now: not in someone else's, a finished one or an expired
// one, and never for plain practice.
func TestInOpenMock(t *testing.T) {
	url := testdb.URL(t)
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	learner := mockLearner(t, tx, "open-mock")
	other := mockLearner(t, tx, "other-learner")
	// A learner has at most one listening mock in progress, so the expired one
	// belongs to someone who walked away from theirs.
	lapsed := mockLearner(t, tx, "lapsed-learner")
	var pteQuestion, testID, pteMock, pteVersion string
	if err := tx.QueryRow(ctx, `SELECT id FROM questions WHERE exam = 'PTE' AND skill = 'listening' LIMIT 1`).Scan(&pteQuestion); err != nil {
		t.Fatalf("a PTE listening question: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT id FROM listening_tests LIMIT 1`).Scan(&testID); err != nil {
		t.Fatalf("a listening test: %v", err)
	}
	if err := tx.QueryRow(ctx, `SELECT id, exam_version_id FROM mocks WHERE exam = 'PTE' LIMIT 1`).Scan(&pteMock, &pteVersion); err != nil {
		t.Fatalf("a PTE mock: %v", err)
	}

	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	listening := func(user, status, expires string, ids ...string) {
		t.Helper()
		exec(`INSERT INTO listening_mock_sessions (user_id, test_id, question_ids, status, duration_minutes, expires_at)
		      VALUES ($1, $2, $3, $4, 30, now() + $5::interval)`, user, testID, ids, status, expires)
	}
	listening(learner, "in_progress", "1 hour", "ielts-open")
	listening(learner, "submitted", "1 hour", "ielts-finished")
	listening(lapsed, "in_progress", "-1 hour", "ielts-expired")
	listening(other, "in_progress", "1 hour", "ielts-someone-else")

	var session string
	if err := tx.QueryRow(ctx, `
		INSERT INTO pte_mock_sessions (user_id, kind, mock_id, exam_version_id, total_items, status)
		VALUES ($1, 'listening', $2, $3, 1, 'in_progress') RETURNING id`, learner, pteMock, pteVersion).Scan(&session); err != nil {
		t.Fatal(err)
	}
	exec(`INSERT INTO pte_mock_items (session_id, position, question_id, task, part) VALUES ($1, 1, $2, 'wfd', 'listening')`, session, pteQuestion)

	repo := NewRepository(tx)
	cases := []struct {
		name, user, question string
		want                 bool
	}{
		{"in an open listening mock", learner, "ielts-open", true},
		{"in an open PTE mock", learner, pteQuestion, true},
		{"in a submitted mock", learner, "ielts-finished", false},
		{"in an expired mock", lapsed, "ielts-expired", false},
		{"in another learner's mock", learner, "ielts-someone-else", false},
		{"only practised", learner, "practice-only", false},
		{"the PTE item, for another learner", other, pteQuestion, false},
	}
	for _, c := range cases {
		got, err := repo.InOpenMock(ctx, c.user, c.question)
		if err != nil {
			t.Fatal(err)
		}
		if got != c.want {
			t.Errorf("%s: InOpenMock = %v, want %v", c.name, got, c.want)
		}
	}

	// Once the PTE mock is completed, its script is available again.
	exec(`UPDATE pte_mock_sessions SET status = 'completed' WHERE id = $1`, session)
	if got, err := repo.InOpenMock(ctx, learner, pteQuestion); err != nil || got {
		t.Fatalf("after completing the PTE mock: InOpenMock = %v, %v", got, err)
	}
}

func mockLearner(t *testing.T, tx pgx.Tx, label string) string {
	t.Helper()
	var id string
	if err := tx.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, referral_code)
		VALUES ($1 || '-' || gen_random_uuid() || '@test.local', 'Mock Test', 'free', 'Asia/Kathmandu',
		        'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id`, label).Scan(&id); err != nil {
		t.Fatalf("create learner: %v", err)
	}
	return id
}
