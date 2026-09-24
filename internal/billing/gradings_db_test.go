package billing

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

func gradingPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testPool(t)
	if err := database.Migrate(context.Background(), pool, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

// onPlan moves a learner onto a paid plan that started today.
func onPlan(t *testing.T, pool *pgxpool.Pool, user models.User, plan string) models.User {
	t.Helper()
	validUntil := time.Now().Add(30 * 24 * time.Hour)
	var started time.Time
	if err := pool.QueryRow(context.Background(), `
		UPDATE users SET plan_id = $2, plan_started_at = CURRENT_DATE, plan_valid_until = $3
		 WHERE id = $1 RETURNING plan_started_at`, user.ID, plan, validUntil).Scan(&started); err != nil {
		t.Fatalf("move to %s: %v", plan, err)
	}
	user.PlanID, user.PlanValidUntil, user.PlanStartedAt = plan, &validUntil, started
	return user
}

func TestPlansCarryTheirGradingPools(t *testing.T) {
	pool := gradingPool(t)
	want := map[string]struct {
		gradings, mocks int
		period          string
	}{
		"free": {10, 1, GradingsPerMonth}, "weekly": {110, 2, GradingsPerPlan},
		"pro": {22, 5, GradingsPerDay}, "elite": {1100, 20, GradingsPerPlan},
	}
	plans, err := NewRepository(pool).Plans(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range plans {
		w, ok := want[p.ID]
		if !ok {
			continue
		}
		if p.AIGradingsPerPeriod != w.gradings || p.MockTestsIncluded != w.mocks || p.AIGradingsPeriod != w.period {
			t.Errorf("%s: %d gradings per %s, %d mocks; want %d per %s and %d", p.ID, p.AIGradingsPerPeriod,
				p.AIGradingsPeriod, p.MockTestsIncluded, w.gradings, w.period, w.mocks)
		}
	}
}

// Every AI-marked answer spends a grading, so re-submitting the same question
// cannot be graded for free: the pool runs out.
func TestFreeGradingsRunOut(t *testing.T) {
	pool := gradingPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	for i := 0; i < 10; i++ {
		if _, err := svc.CheckAIGradings(ctx, pool, user, UnitsPerGrading); err != nil {
			t.Fatalf("grading %d refused: %v", i+1, err)
		}
		if err := svc.RecordAIGrading(ctx, pool, user.ID, UnitsPerGrading, "practice-writing", "same-question"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.CheckAIGradings(ctx, pool, user, UnitsPerGrading); !errors.Is(err, ErrGradingLimitReached) {
		t.Fatalf("the 11th grading on the free plan: %v", err)
	}
	if _, err := svc.CheckAIGradings(ctx, pool, user, WordMatchUnits); !errors.Is(err, ErrGradingLimitReached) {
		t.Fatalf("a word-matched answer with nothing left: %v", err)
	}
	state, err := svc.State(ctx, pool, user)
	if err != nil {
		t.Fatal(err)
	}
	if state.AIGradingsUsed != 10 || state.AIGradingsLimit != 10 {
		t.Fatalf("state %v of %d", state.AIGradingsUsed, state.AIGradingsLimit)
	}
	if renews, _ := time.Parse(time.DateOnly, state.AIGradingsRenewOn); renews.Day() != 1 {
		t.Fatalf("the free pool renews on the 1st, got %s", state.AIGradingsRenewOn)
	}
}

// Read Aloud and Repeat Sentence are word-matched and cost a fifth of a grading.
func TestWordMatchedAnswersCostAFifth(t *testing.T) {
	pool := gradingPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	for i := 0; i < 3; i++ {
		if err := svc.RecordAIGrading(ctx, pool, user.ID, WordMatchUnits, "practice-read-repeat", ""); err != nil {
			t.Fatal(err)
		}
	}
	state, err := svc.State(ctx, pool, user)
	if err != nil {
		t.Fatal(err)
	}
	if state.AIGradingsUsed != 0.6 {
		t.Fatalf("three word-matched answers used %v gradings, want 0.6", state.AIGradingsUsed)
	}
}

// A paid pool is the plan's own term: what was used before it started does
// not count, and it renews when the plan does.
func TestPaidPoolStartsWithThePlan(t *testing.T) {
	pool := gradingPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	if _, err := pool.Exec(ctx, `INSERT INTO ai_grading_usage (user_id, units, source, created_at)
		VALUES ($1, 500, 'before-the-plan', now() - interval '40 days')`, user.ID); err != nil {
		t.Fatal(err)
	}
	user = onPlan(t, pool, user, "weekly")
	if err := svc.RecordAIGrading(ctx, pool, user.ID, UnitsPerGrading, "practice-speaking", ""); err != nil {
		t.Fatal(err)
	}
	state, err := svc.State(ctx, pool, user)
	if err != nil {
		t.Fatal(err)
	}
	if state.AIGradingsLimit != 110 || state.AIGradingsUsed != 1 {
		t.Fatalf("weekly pool %v of %d, want 1 of 110", state.AIGradingsUsed, state.AIGradingsLimit)
	}
	if state.AIGradingsRenewOn != user.PlanValidUntil.Format(time.DateOnly) {
		t.Fatalf("renews %s, want the plan's end", state.AIGradingsRenewOn)
	}
}

// Udaan's practice is unlimited, but its full mocks are counted.
func TestUnlimitedPlanStillCountsFullMocks(t *testing.T) {
	pool := gradingPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := onPlan(t, pool, newLearner(t, pool), "elite")

	if _, err := svc.CheckSubTestCredits(ctx, pool, user, 1000); err != nil {
		t.Fatalf("Udaan practice is unlimited: %v", err)
	}
	for i := 0; i < 20; i++ {
		if _, err := pool.Exec(ctx, `INSERT INTO full_mock_sessions (user_id, module, status) VALUES ($1, 'academic', 'completed')`,
			user.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.CheckMockAllowance(ctx, pool, user); !errors.Is(err, ErrMockLimitReached) {
		t.Fatalf("the 21st full mock on Udaan: %v", err)
	}
}

// Taiyari's gradings are 22 a day and reset at the learner's midnight: what
// was used yesterday does not count against today.
func TestTaiyariGradingsResetDaily(t *testing.T) {
	pool := gradingPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := onPlan(t, pool, newLearner(t, pool), "pro")

	if _, err := pool.Exec(ctx, `INSERT INTO ai_grading_usage (user_id, units, source, created_at)
		VALUES ($1, 110, 'yesterday', now() - interval '2 days')`, user.ID); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 22; i++ {
		if _, err := svc.CheckAIGradings(ctx, pool, user, UnitsPerGrading); err != nil {
			t.Fatalf("grading %d of 22 today refused: %v", i+1, err)
		}
		if err := svc.RecordAIGrading(ctx, pool, user.ID, UnitsPerGrading, "practice-speaking", ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.CheckAIGradings(ctx, pool, user, UnitsPerGrading); !errors.Is(err, ErrGradingLimitReached) {
		t.Fatalf("the 23rd grading today: %v", err)
	}
	state, err := svc.State(ctx, pool, user)
	if err != nil {
		t.Fatal(err)
	}
	if state.AIGradingsPeriod != GradingsPerDay || state.AIGradingsUsed != 22 || state.AIGradingsLimit != 22 {
		t.Fatalf("state %v of %d per %s", state.AIGradingsUsed, state.AIGradingsLimit, state.AIGradingsPeriod)
	}
}
