package billing

import (
	"context"
	"errors"
	"github.com/prepyo/backend/internal/testdb"
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
	url := testdb.URL(t)
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
		SELECT q.group_id, q.id, q.exam, q.exam_version_id, q.skill
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
		if err := rows.Scan(&q.GroupID, &q.ID, &q.Exam, &q.ExamVersionID, &q.Skill); err != nil {
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

// One start pays for a task set however many answers are submitted within it.
func TestTaskSetCountsOnceHoweverManyQuestions(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	user := newLearner(t, pool)
	sets := taskSets(t, pool, 1)

	if got := usage(t, svc, pool, user); got != 0 {
		t.Fatalf("a new learner has used %d sub-tests, want 0", got)
	}

	startSession(t, pool, user, SubTestKeyForQuestion(sets[0][0]))
	answer(t, pool, user, sets[0][0])
	if got := usage(t, svc, pool, user); got != 1 {
		t.Fatalf("after start and first answer usage = %d, want 1", got)
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
		q := sets[i][0]
		if _, err := svc.RecordSessionStart(ctx, pool, user, string(q.Exam), string(q.Skill), SubTestKeyForQuestion(q)); err != nil {
			t.Fatal(err)
		}
	}
	if got := usage(t, svc, pool, user); got != limit {
		t.Fatalf("usage = %d after opening %d sets, want %d", got, limit, limit)
	}

	for i := 0; i < limit; i++ {
		q := sets[i][1]
		if err := svc.RequireStartedSubTest(ctx, pool, user, q, q.Exam); err != nil {
			t.Errorf("continuing set %d was refused at the limit: %v", i, err)
		}
	}

	fresh := SubTestKeyForQuestion(sets[limit][0])
	_, err := svc.CheckSubTestAllowance(ctx, pool, user, fresh)
	if !errors.Is(err, ErrLimitReached) {
		t.Errorf("starting a new set at the limit returned %v, want ErrLimitReached", err)
	}
}

// startSession mirrors billing.RecordSessionStart without pulling the handler
// layer in: what is under test is the counting, not the insert.
func startSession(t *testing.T, pool *pgxpool.Pool, user models.User, itemID string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO practice_sessions (user_id, exam, skill, item_id, status)
		VALUES ($1, 'IELTS', 'reading', $2, 'active')`, user.ID, itemID)
	if err != nil {
		t.Fatalf("start session: %v", err)
	}
}

// A session row is what the start of a test writes, and RecordSessionStart
// stores the task-set key, COALESCE(group_id, id), as item_id. Submitting
// answers for that set must not charge a second sub-test: for grouped tasks the
// attempt's question id never equals the session's item_id, so an exclusion
// that compared them directly never matched and every grouped task cost one
// credit at start plus another at submit.
func TestSessionThenSubmitCountsOnceForGroupedTasks(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	user := newLearner(t, pool)
	sets := taskSets(t, pool, 1)

	key := SubTestKeyForQuestion(sets[0][0])
	startSession(t, pool, user, key)

	if got := usage(t, svc, pool, user); got != 1 {
		t.Fatalf("after starting the set usage = %d, want 1", got)
	}

	answer(t, pool, user, sets[0][0])
	if got := usage(t, svc, pool, user); got != 1 {
		t.Errorf("after starting and answering the set usage = %d, want 1 — the submit charged a second credit", got)
	}

	// Finishing the rest of the set changes nothing: the set was paid for once.
	for _, q := range sets[0][1:] {
		answer(t, pool, user, q)
	}
	if got := usage(t, svc, pool, user); got != 1 {
		t.Errorf("after answering the whole set usage = %d, want 1", got)
	}
}

// Starting a set and never submitting it still costs the one credit the start
// spent: nothing extra, nothing refunded.
func TestAbandonedSessionCountsOnce(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	user := newLearner(t, pool)
	sets := taskSets(t, pool, 1)

	startSession(t, pool, user, SubTestKeyForQuestion(sets[0][0]))

	if got := usage(t, svc, pool, user); got != 1 {
		t.Errorf("an abandoned started session gave usage = %d, want 1", got)
	}
}

func TestEvaluationDoesNotSpendAnotherCredit(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	for _, skill := range []models.SkillType{models.SkillWriting, models.SkillSpeaking} {
		t.Run(string(skill), func(t *testing.T) {
			user := newLearner(t, pool)
			var q models.Question
			if err := pool.QueryRow(ctx, `SELECT id, exam, skill FROM questions WHERE skill = $1 LIMIT 1`, skill).Scan(&q.ID, &q.Exam, &q.Skill); err != nil {
				t.Fatal(err)
			}
			if _, err := svc.RecordSessionStart(ctx, pool, user, string(q.Exam), string(skill), q.ID); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				_, err := pool.Exec(ctx, `INSERT INTO ai_evaluations
					(user_id, question_id, exam, skill, evaluation_version, request_fingerprint,
					 score_confidence, provider, model, prompt_version)
					VALUES ($1, $2, $3, $4, 'test', $5, 'low', 'test', 'test', 'test')`,
					user.ID, q.ID, q.Exam, skill, []byte{byte(i)})
				if err != nil {
					t.Fatal(err)
				}
				if got := usage(t, svc, pool, user); got != 1 {
					t.Fatalf("usage after evaluation = %d, want 1", got)
				}
			}
		})
	}
}

func TestSubmissionRequiresMatchingStartedSession(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)
	other := newLearner(t, pool)
	q := taskSets(t, pool, 1)[0][0]
	if err := svc.RequireStartedSubTest(ctx, pool, user, q, q.Exam); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("without start: %v", err)
	}
	id, err := svc.RecordSessionStart(ctx, pool, user, string(q.Exam), string(q.Skill), SubTestKeyForQuestion(q))
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.RequireStartedSubTest(ctx, pool, other, q, q.Exam); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("other learner's session: %v", err)
	}
	wrongTask := q
	wrongTask.GroupID = "not-started"
	if err := svc.RequireStartedSubTest(ctx, pool, user, wrongTask, q.Exam); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("unstarted task: %v", err)
	}
	wrongExam := models.ExamIELTS
	if q.Exam == wrongExam {
		wrongExam = models.ExamPTE
	}
	if err := svc.RequireStartedSubTest(ctx, pool, user, q, wrongExam); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("wrong exam: %v", err)
	}
	wrongSkill := q
	wrongSkill.Skill = models.SkillSpeaking
	if q.Skill == wrongSkill.Skill {
		wrongSkill.Skill = models.SkillReading
	}
	if err := svc.RequireStartedSubTest(ctx, pool, user, wrongSkill, q.Exam); !errors.Is(err, ErrSessionRequired) {
		t.Fatalf("wrong skill: %v", err)
	}
	if err := svc.RecordSessionStop(ctx, pool, user, id); err != nil {
		t.Fatal(err)
	}
	if err := svc.RequireStartedSubTest(ctx, pool, user, q, q.Exam); err != nil {
		t.Fatalf("submit after stopping timer: %v", err)
	}
	if got := usage(t, svc, pool, user); got != 1 {
		t.Fatalf("stopping changed usage to %d", got)
	}
}

func TestMocksDoNotConsumeSubTests(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	before := usage(t, svc, pool, user)

	var mockID, examVersionID string
	err := pool.QueryRow(ctx, `SELECT id, exam_version_id FROM mocks WHERE NOT is_diagnostic AND NOT is_generated LIMIT 1`).
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

func TestConcurrentStartsSpendOnlyRemainingCredit(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	limit := 5
	sets := taskSets(t, pool, limit+4)

	// Spend all but one.
	for i := 0; i < limit-1; i++ {
		startSession(t, pool, user, SubTestKeyForQuestion(sets[i][0]))
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
			if _, err := svc.RecordSessionStart(ctx, tx, user, string(q.Exam), string(q.Skill), SubTestKeyForQuestion(q)); err != nil {
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
		INSERT INTO practice_sessions (user_id, item_id, exam, skill, created_at)
		VALUES ($1, $2, $3, $4, $5)`, user.ID, SubTestKeyForQuestion(q), q.Exam, q.Skill, at)
	if err != nil {
		t.Fatalf("record attempt at %s: %v", at, err)
	}
}

// Udaan has no daily sub-test limit and no mock limit: a learner far past the
// numbers the plan row still carries is never refused.
func TestUnlimitedPlanIsNeverRefused(t *testing.T) {
	pool := testPool(t)
	svc := newService(pool)
	ctx := context.Background()
	user := newLearner(t, pool)

	if _, err := pool.Exec(ctx, `
		UPDATE users SET plan_id = 'elite', role = 'udaan',
		       plan_valid_until = CURRENT_DATE + 30 WHERE id = $1`, user.ID); err != nil {
		t.Fatal(err)
	}
	valid := time.Now().AddDate(0, 0, 30)
	user.PlanID, user.PlanValidUntil = "elite", &valid

	// Well past the 60 sub-tests a day the row still lists.
	if _, err := pool.Exec(ctx, `
		INSERT INTO practice_sessions (user_id, exam, skill, item_id, status, credits)
		SELECT $1, 'PTE', 'reading', 'item-' || g, 'stopped', 1 FROM generate_series(1, 75) g`, user.ID); err != nil {
		t.Fatal(err)
	}

	state, err := svc.CheckSubTestCredits(ctx, pool, user, 5)
	if err != nil {
		t.Fatalf("unlimited plan refused a sub-test: %v", err)
	}
	if !state.Unlimited || state.DailySubTestsUsed != 75 {
		t.Fatalf("state = %+v, want unlimited with 75 used", state)
	}
	if _, err := svc.CheckMockAllowance(ctx, pool, user); err != nil {
		t.Fatalf("unlimited plan refused a mock: %v", err)
	}

	plan, err := NewRepository(pool).Plan(ctx, "elite")
	if err != nil {
		t.Fatal(err)
	}
	if plan.DurationDays != 30 || plan.BonusDays != 3 || !plan.Unlimited {
		t.Fatalf("elite plan = %+v, want 30 + 3 days, unlimited", plan)
	}
}
