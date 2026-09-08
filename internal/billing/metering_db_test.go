package billing

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
)

// Database-backed integration tests for sub-test quota metering.
// Run with:
//   TEST_DATABASE_URL=postgres://postgres@localhost:5432/prepyo?sslmode=disable go test ./internal/billing/

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed metering tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func newLearner(t *testing.T, pool *pgxpool.Pool) models.User {
	t.Helper()
	ctx := context.Background()

	var user models.User
	err := pool.QueryRow(ctx, `
		INSERT INTO users (email, name, plan_id, timezone, referral_code)
		VALUES ('metering-' || gen_random_uuid() || '@test.local', 'Metering Test', 'free',
		        'Asia/Kathmandu', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, bonus_mock_tests`).
		Scan(&user.ID, &user.PlanID, &user.Timezone, &user.BonusMockTests)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}

	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID); err != nil {
			t.Errorf("cleanup learner: %v", err)
		}
	})
	return user
}

func taskSets(t *testing.T, pool *pgxpool.Pool, n int) [][]models.Question {
	t.Helper()
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT q.group_id, q.id, q.exam, q.exam_version_id
		  FROM questions q
		 WHERE q.group_id IN (
			   SELECT group_id FROM questions
			    WHERE group_id IS NOT NULL AND is_published
			    GROUP BY group_id HAVING count(*) >= 2
			    ORDER BY group_id
			    LIMIT $1)
		 ORDER BY q.group_id, q.id`, n)
	if err != nil {
		t.Fatalf("load task sets: %v", err)
	}
	defer rows.Close()

	bySet := map[string][]models.Question{}
	var order []string
	for rows.Next() {
		var q models.Question
		if err := rows.Scan(&q.GroupID, &q.ID, &q.Exam, &q.ExamVersionID); err != nil {
			t.Fatalf("scan question: %v", err)
		}
		if _, seen := bySet[q.GroupID]; !seen {
			order = append(order, q.GroupID)
		}
		bySet[q.GroupID] = append(bySet[q.GroupID], q)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("load task sets: %v", err)
	}
	if len(order) < n {
		t.Skipf("need %d seeded task sets with 2+ questions, found %d", n, len(order))
	}

	sets := make([][]models.Question, 0, n)
	for _, id := range order {
		sets = append(sets, bySet[id])
	}
	return sets
}

// answer records an attempt the way practice.Repository.Save does. It writes the
// row directly: what is under test is the counting, not the grading.
func answer(t *testing.T, pool *pgxpool.Pool, user models.User, q models.Question) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO practice_attempts
			(user_id, question_id, exam, exam_version_id, is_correct, score, max_score, accuracy_percentage)
		VALUES ($1, $2, $3, $4, TRUE, 1, 1, 100)`, user.ID, q.ID, q.Exam, q.ExamVersionID)
	if err != nil {
		t.Fatalf("record attempt for %s: %v", q.ID, err)
	}
}

func usage(t *testing.T, svc *Service, pool *pgxpool.Pool, user models.User) int {
	t.Helper()
	state, err := svc.State(context.Background(), pool, user)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	return state.DailySubTestsUsed
}

func newService(pool *pgxpool.Pool) *Service {
	return NewService(NewRepository(pool), nil)
}

// A task set is one sub-test however many questions are inside it. This is the
// whole point of COALESCE(group_id, id): counting rows instead would make a
// six-statement True/False set cost six.
func TestTaskSetCountsOnceHoweverManyQuestions(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	user := newLearner(t, pool)
	sets := taskSets(t, pool, 1)

	if got := usage(t, svc, pool, user); got != 0 {
		t.Fatalf("a new learner has used %d sub-tests, want 0", got)
	}

	answer(t, pool, user, sets[0][0])
	if got := usage(t, svc, pool, user); got != 1 {
		t.Fatalf("after the first answer usage = %d, want 1", got)
	}

	for _, q := range sets[0][1:] {
		answer(t, pool, user, q)
	}
	if got := usage(t, svc, pool, user); got != 1 {
		t.Errorf("after answering the whole set usage = %d, want 1 — the set is being counted per question", got)
	}
}

func TestSetAlreadyStartedStaysAnswerableAtTheLimit(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	limit := 5 // the free plan
	sets := taskSets(t, pool, limit+1)

	for i := 0; i < limit; i++ {
		answer(t, pool, user, sets[i][0])
	}
	if got := usage(t, svc, pool, user); got != limit {
		t.Fatalf("usage = %d after opening %d sets, want %d", got, limit, limit)
	}

	for i := 0; i < limit; i++ {
		key := SubTestKeyForQuestion(sets[i][1])
		if _, err := svc.CheckSubTestAllowance(ctx, pool, user, key); err != nil {
			t.Errorf("continuing set %d was refused at the limit: %v", i, err)
		}
	}

	fresh := SubTestKeyForQuestion(sets[limit][0])
	_, err := svc.CheckSubTestAllowance(ctx, pool, user, fresh)
	if !errors.Is(err, ErrLimitReached) {
		t.Errorf("starting a new set at the limit returned %v, want ErrLimitReached", err)
	}
}

func TestMocksDoNotConsumeSubTests(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	before := usage(t, svc, pool, user)

	var mockID, examVersionID string
	err := pool.QueryRow(ctx, `SELECT id, exam_version_id FROM mocks WHERE NOT is_diagnostic LIMIT 1`).
		Scan(&mockID, &examVersionID)
	if err != nil {
		t.Skipf("no non-diagnostic mock seeded: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO mock_attempts
			(user_id, mock_id, exam_version_id, exam, user_score, skill_scores,
			 total_correct, total_questions, duration_seconds)
		VALUES ($1, $2, $3, 'IELTS', 6.5, '{}'::jsonb, 20, 40, 3600)`,
		user.ID, mockID, examVersionID)
	if err != nil {
		t.Fatalf("record mock attempt: %v", err)
	}

	if after := usage(t, svc, pool, user); after != before {
		t.Errorf("a mock changed sub-test usage from %d to %d; mocks have their own allowance", before, after)
	}

	state, err := svc.State(ctx, pool, user)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if state.MockTestsUsed != 1 {
		t.Errorf("mock allowance used = %d, want 1", state.MockTestsUsed)
	}
}

func TestConcurrentFirstAnswersSpendOneSubTest(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	limit := 5
	sets := taskSets(t, pool, limit+4)

	// Spend all but one.
	for i := 0; i < limit-1; i++ {
		answer(t, pool, user, sets[i][0])
	}

	contenders := sets[limit-1:]
	var wg sync.WaitGroup
	results := make([]error, len(contenders))

	for i, set := range contenders {
		wg.Add(1)
		go func(i int, q models.Question) {
			defer wg.Done()

			tx, err := pool.Begin(ctx)
			if err != nil {
				results[i] = err
				return
			}
			defer tx.Rollback(ctx)

			if err := LockUserForQuota(ctx, tx, user.ID); err != nil {
				results[i] = err
				return
			}
			if _, err := svc.CheckSubTestAllowance(ctx, tx, user, SubTestKeyForQuestion(q)); err != nil {
				results[i] = err
				return
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO practice_attempts
					(user_id, question_id, exam, exam_version_id, is_correct, score, max_score, accuracy_percentage)
				VALUES ($1, $2, $3, $4, TRUE, 1, 1, 100)`, user.ID, q.ID, q.Exam, q.ExamVersionID); err != nil {
				results[i] = err
				return
			}
			results[i] = tx.Commit(ctx)
		}(i, set[0])
	}
	wg.Wait()

	allowed, refused := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			allowed++
		case errors.Is(err, ErrLimitReached):
			refused++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if allowed != 1 {
		t.Errorf("%d of %d concurrent first-answers were allowed with one sub-test left, want 1", allowed, len(contenders))
	}
	if refused != len(contenders)-1 {
		t.Errorf("refused %d, want %d", refused, len(contenders)-1)
	}
	if got := usage(t, svc, pool, user); got != limit {
		t.Errorf("usage ended at %d, want exactly the limit %d", got, limit)
	}
}

func TestUsageCountsTheLearnersOwnDay(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	sets := taskSets(t, pool, 2)

	for _, tz := range []string{"Asia/Kathmandu", "not/a-zone"} {
		t.Run(tz, func(t *testing.T) {
			user := newLearner(t, pool)
			if _, err := pool.Exec(ctx, `UPDATE users SET timezone = $2 WHERE id = $1`, user.ID, tz); err != nil {
				t.Fatalf("set timezone: %v", err)
			}
			user.Timezone = tz

			start := gamification.LocalDayStart(user)

			mustInsertAt(t, pool, user, sets[0][0], start.Add(-time.Second))
			if got := usage(t, svc, pool, user); got != 0 {
				t.Errorf("an attempt one second before local midnight counted as today (usage = %d)", got)
			}

			mustInsertAt(t, pool, user, sets[1][0], start.Add(time.Second))
			if got := usage(t, svc, pool, user); got != 1 {
				t.Errorf("an attempt one second after local midnight gave usage = %d, want 1", got)
			}
		})
	}
}

func mustInsertAt(t *testing.T, pool *pgxpool.Pool, user models.User, q models.Question, at time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO practice_attempts
			(user_id, question_id, exam, exam_version_id, is_correct, score, max_score, accuracy_percentage, created_at)
		VALUES ($1, $2, $3, $4, TRUE, 1, 1, 100, $5)`, user.ID, q.ID, q.Exam, q.ExamVersionID, at)
	if err != nil {
		t.Fatalf("record attempt at %s: %v", at, err)
	}
}
