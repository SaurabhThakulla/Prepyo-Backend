package progress

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/mistakes"
	"github.com/prepyo/backend/internal/models"
)

// Database-backed integration tests for exam-scoped progress and attempts.
// Run with:
//   TEST_DATABASE_URL=postgres://postgres@localhost:5432/prepyo?sslmode=disable go test ./internal/progress/

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed progress tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func newLearner(t *testing.T, pool *pgxpool.Pool, exam models.ExamType) models.User {
	t.Helper()

	var user models.User
	err := pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, referral_code)
		VALUES ('progress-' || gen_random_uuid() || '@test.local', 'Progress Test', 'free',
		        'Asia/Kathmandu', $1,
		        'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam`, exam).
		Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam)
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

// sharedQuestion returns a question ID supported by both exams.
func sharedQuestion(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()

	const id = "q-test-shared-eligibility"
	_, err := pool.Exec(ctx, `
		INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id,
			type_name, title, prompt, options, correct_answers, time_limit_seconds, points,
			difficulty, tags, passage_id, group_id, group_position)
		SELECT $1, exam_version_id, exam, ARRAY['IELTS', 'PTE'], skill, type_id, type_name,
		       'Shared', 'Set by both exams', options, correct_answers,
		       time_limit_seconds, points, difficulty, tags, passage_id, group_id, 98
		  FROM questions
		 WHERE type_id = 'reading-mcq-single' AND group_id IS NOT NULL
		 LIMIT 1
		ON CONFLICT (id) DO NOTHING`, id)
	if err != nil {
		t.Fatalf("create shared question: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM questions WHERE id = $1`, id); err != nil {
			t.Errorf("cleanup shared question: %v", err)
		}
	})
	return id
}

func recordAttempt(t *testing.T, pool *pgxpool.Pool, user models.User, questionID string, exam models.ExamType) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO practice_attempts
			(user_id, question_id, exam, exam_version_id, is_correct, score, max_score, accuracy_percentage)
		SELECT $1, $2, $3, v.id, TRUE, 1, 1, 100
		  FROM exam_versions v WHERE v.exam = $3 AND v.is_current`,
		user.ID, questionID, exam)
	if err != nil {
		t.Fatalf("record attempt: %v", err)
	}
}

func TestProgressFollowsTheAttemptNotTheQuestion(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	svc := NewService(exams.NewRepository(pool))

	questionID := sharedQuestion(t, pool)

	ielts := newLearner(t, pool, models.ExamIELTS)
	pte := newLearner(t, pool, models.ExamPTE)

	recordAttempt(t, pool, ielts, questionID, models.ExamIELTS)
	recordAttempt(t, pool, pte, questionID, models.ExamPTE)

	for _, user := range []models.User{ielts, pte} {
		estimate, err := svc.Estimate(ctx, pool, user)
		if err != nil {
			t.Fatalf("estimate for %s: %v", user.TargetExam, err)
		}
		if estimate.BasedOn == 0 {
			t.Errorf("%s learner has an attempt recorded but progress counts none; "+
				"the query is still reading the exam off the question", user.TargetExam)
		}
	}
}

func TestProgressIgnoresTheOtherExamsAttempts(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	svc := NewService(exams.NewRepository(pool))

	questionID := sharedQuestion(t, pool)

	// One learner, sitting IELTS, who has only ever answered under PTE.
	user := newLearner(t, pool, models.ExamIELTS)
	recordAttempt(t, pool, user, questionID, models.ExamPTE)

	estimate, err := svc.Estimate(ctx, pool, user)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if estimate.BasedOn != 0 {
		t.Errorf("IELTS progress counted %d PTE attempts", estimate.BasedOn)
	}
}

func TestExistingAttemptsStillResolveTheirQuestions(t *testing.T) {
	pool := testPool(t)

	var orphans int
	err := pool.QueryRow(context.Background(), `
		SELECT count(*) FROM practice_attempts a
		 WHERE NOT EXISTS (SELECT 1 FROM questions q WHERE q.id = a.question_id)
		    OR a.exam IS NULL`).Scan(&orphans)
	if err != nil {
		t.Fatalf("check attempts: %v", err)
	}
	if orphans != 0 {
		t.Errorf("%d practice attempts have no question or no exam", orphans)
	}
}

func TestMistakeBankFiltersOnTheAttemptsExam(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := mistakes.NewRepository(pool)

	questionID := sharedQuestion(t, pool)
	user := newLearner(t, pool, models.ExamPTE)

	if err := repo.Record(ctx, pool, mistakes.RecordParams{
		UserID:     user.ID,
		QuestionID: questionID,
		Exam:       models.ExamPTE,
		ErrorTag:   "Detail",
	}); err != nil {
		t.Fatalf("record mistake: %v", err)
	}

	pteList, _, err := repo.List(ctx, mistakes.ListParams{UserID: user.ID, Exam: models.ExamPTE, Limit: 10})
	if err != nil {
		t.Fatalf("list PTE mistakes: %v", err)
	}
	if len(pteList) != 1 {
		t.Fatalf("PTE mistakes = %d, want 1", len(pteList))
	}
	if pteList[0].Exam != models.ExamPTE {
		t.Errorf("mistake exam = %s, want PTE", pteList[0].Exam)
	}

	ieltsList, _, err := repo.List(ctx, mistakes.ListParams{UserID: user.ID, Exam: models.ExamIELTS, Limit: 10})
	if err != nil {
		t.Fatalf("list IELTS mistakes: %v", err)
	}
	if len(ieltsList) != 0 {
		t.Errorf("IELTS mistakes = %d, want 0; a PTE mistake leaked across exams", len(ieltsList))
	}
}

func TestBankSizeCountsSharedQuestionsForBothExams(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	svc := NewService(exams.NewRepository(pool))

	sharedQuestion(t, pool)

	ielts, _, err := svc.bankCoverage(ctx, pool, newLearner(t, pool, models.ExamIELTS))
	if err != nil {
		t.Fatalf("IELTS bank: %v", err)
	}
	pte, _, err := svc.bankCoverage(ctx, pool, newLearner(t, pool, models.ExamPTE))
	if err != nil {
		t.Fatalf("PTE bank: %v", err)
	}

	if ielts[models.SkillReading] == 0 || pte[models.SkillReading] == 0 {
		t.Fatalf("reading bank = %d IELTS / %d PTE, want both non-zero",
			ielts[models.SkillReading], pte[models.SkillReading])
	}
}
