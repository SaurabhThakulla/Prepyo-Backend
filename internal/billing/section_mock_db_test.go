package billing

import (
	"context"
	"errors"
	"testing"
)

// A section mock spends SectionMockSubTests in one go, all or nothing.
func TestSectionMockSpendsFiveSubTests(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool) // free plan: 5 a day

	if _, err := svc.CheckSubTestCredits(ctx, pool, user, SectionMockSubTests); err != nil {
		t.Fatalf("a fresh day should afford a section mock: %v", err)
	}
	if _, err := svc.RecordSessionStartCredits(ctx, pool, user, "PTE", "reading", "mock:test", SectionMockSubTests); err != nil {
		t.Fatalf("record section mock: %v", err)
	}
	if got := usage(t, svc, pool, user); got != SectionMockSubTests {
		t.Fatalf("usage after a section mock = %d, want %d", got, SectionMockSubTests)
	}
	if _, err := svc.CheckSubTestAllowance(ctx, pool, user, "any"); !errors.Is(err, ErrLimitReached) {
		t.Errorf("an ordinary task after spending all 5 should be refused, got %v", err)
	}
}

func TestSectionMockRefusedWithoutFiveLeft(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	if _, err := svc.RecordSessionStart(ctx, pool, user, "PTE", "reading", "one-task"); err != nil {
		t.Fatalf("record task: %v", err)
	}
	if _, err := svc.CheckSubTestCredits(ctx, pool, user, SectionMockSubTests); !errors.Is(err, ErrLimitReached) {
		t.Errorf("4 left should not start a 5-credit mock, got %v", err)
	}
	if _, err := svc.CheckSubTestAllowance(ctx, pool, user, "any"); err != nil {
		t.Errorf("4 left should still allow an ordinary task: %v", err)
	}
}

// Section mocks are paid for in sub-tests, so their result must not also use
// up the full-mock allowance.
func TestSectionMockLeavesFullMockAllowanceAlone(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	var mockID, examVersionID, exam string
	if err := pool.QueryRow(ctx, `SELECT id, exam_version_id, exam FROM mocks WHERE is_generated LIMIT 1`).
		Scan(&mockID, &examVersionID, &exam); err != nil {
		t.Skipf("no generated mock seeded: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO mock_attempts
			(user_id, mock_id, exam_version_id, exam, user_score, skill_scores,
			 total_correct, total_questions, duration_seconds)
		VALUES ($1, $2, $3, $4, 50, '{}'::jsonb, 10, 13, 1800)`,
		user.ID, mockID, examVersionID, exam); err != nil {
		t.Fatalf("record section mock attempt: %v", err)
	}

	state, err := svc.State(ctx, pool, user)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if state.MockTestsUsed != 0 {
		t.Errorf("full mock allowance used = %d after a section mock, want 0", state.MockTestsUsed)
	}
}
