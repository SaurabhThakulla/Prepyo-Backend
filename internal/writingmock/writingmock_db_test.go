package writingmock

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/testdb"
)

// fixedEvaluator rates each task with a set band and records what it was given.
// Submit rates the two tasks at once, so the record is guarded.
type fixedEvaluator struct {
	mu    sync.Mutex
	bands map[string]float64
	texts map[string]string
}

func (f *fixedEvaluator) EvaluateMockTask(_ context.Context, _ models.User, q models.Question, text, _ string) (models.Evaluation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.texts[q.TypeID] = text
	band := f.bands["task2"]
	if q.TypeID != "" && (q.TypeID == "ielts-writing-task1-figure" || q.TypeID == "ielts-writing-task1-letter") {
		band = f.bands["task1"]
	}
	return models.Evaluation{Exam: models.ExamIELTS, Skill: models.SkillWriting, EstimatedScore: &band}, nil
}

func setup(t *testing.T, module string) (*pgxpool.Pool, *Service, *fixedEvaluator, models.User) {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	var user models.User
	err = pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, target_module, referral_code)
		VALUES ('writing-' || gen_random_uuid() || '@test.local', 'Writing Test', 'free', 'Asia/Kathmandu',
		        'IELTS', $1, 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam, target_module`, module).
		Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam, &user.TargetModule)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID)
	})

	eval := &fixedEvaluator{bands: map[string]float64{"task1": 6, "task2": 7}, texts: map[string]string{}}
	svc := NewService(pool, questions.NewRepository(pool), mocks.NewRepository(pool), nil, gamification.NewService(), eval)
	return pool, svc, eval, user
}

// Task 2 counts twice as much as Task 1: (6 + 2*7) / 3 = 6.67, which rounds
// to 6.5 by the half-band rule.
func TestBandWeightsTaskTwoDouble(t *testing.T) {
	for _, tc := range []struct{ t1, t2, want float64 }{
		{6, 7, 6.5}, {7, 6, 6.5}, {6, 6, 6}, {5, 7, 6.5}, {9, 9, 9}, {0, 6, 4},
	} {
		if got := Band(tc.t1, tc.t2); got != tc.want {
			t.Errorf("Band(%v, %v) = %v, want %v", tc.t1, tc.t2, got, tc.want)
		}
	}
}

func TestWritingMockDealsTheModulesTaskOne(t *testing.T) {
	for module, want := range map[string]string{
		models.ModuleAcademic:        "ielts-writing-task1-figure",
		models.ModuleGeneralTraining: "ielts-writing-task1-letter",
	} {
		_, svc, _, user := setup(t, module)
		session, err := svc.Start(context.Background(), user)
		if err != nil {
			t.Fatalf("%s start: %v", module, err)
		}
		if session.Task1.TypeID != want {
			t.Errorf("%s task 1 = %s, want %s", module, session.Task1.TypeID, want)
		}
		if session.Task2.TypeID != "ielts-writing-task2-opinion" {
			t.Errorf("%s task 2 = %s", module, session.Task2.TypeID)
		}
		if session.SecondsRemaining < 3590 || session.DurationMinutes != 60 {
			t.Errorf("%s time = %ds of %d minutes, want about an hour", module, session.SecondsRemaining, session.DurationMinutes)
		}
		// A second start resumes the same paper rather than dealing another.
		again, err := svc.Start(context.Background(), user)
		if err != nil || again.ID != session.ID {
			t.Errorf("%s restart dealt %s (%v), want to resume %s", module, again.ID, err, session.ID)
		}
	}
}

func TestWritingMockRecordsTheWeightedBand(t *testing.T) {
	pool, svc, _, user := setup(t, models.ModuleAcademic)
	ctx := context.Background()
	session, err := svc.Start(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	result, err := svc.Submit(ctx, user, session.ID, "task one text", "task two text")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if result.WritingBand == nil || *result.WritingBand != 6.5 || result.Attempt == nil {
		t.Fatalf("band = %v attempt = %v, want 6.5 and a recorded attempt", result.WritingBand, result.Attempt)
	}
	if got := result.Attempt.SkillScores[models.SkillWriting]; got != 6.5 {
		t.Fatalf("attempt writing score = %v, want 6.5", got)
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM writing_mock_sessions WHERE id = $1`, session.ID).Scan(&status); err != nil || status != "submitted" {
		t.Fatalf("status = %q (%v), want submitted", status, err)
	}
	if _, err := svc.Submit(ctx, user, session.ID, "again", "again"); !errors.Is(err, ErrAlreadySubmitted) {
		t.Fatalf("second submit error = %v, want ErrAlreadySubmitted", err)
	}
}

// After the deadline, only what was saved in time is graded.
func TestLateWritingMockIsGradedFromDrafts(t *testing.T) {
	pool, svc, eval, user := setup(t, models.ModuleAcademic)
	ctx := context.Background()
	session, err := svc.Start(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveDrafts(ctx, user, session.ID, "saved task one", "saved task two"); err != nil {
		t.Fatalf("save drafts: %v", err)
	}
	resumed, err := svc.Resume(ctx, user, session.ID)
	if err != nil || resumed.Task1Draft != "saved task one" || resumed.Task2Draft != "saved task two" {
		t.Fatalf("resume = %+v (%v), want the saved drafts", resumed, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE writing_mock_sessions SET expires_at = now() - interval '10 minutes' WHERE id = $1`, session.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveDrafts(ctx, user, session.ID, "too late", "too late"); !errors.Is(err, ErrPaperClosed) {
		t.Fatalf("late save error = %v, want ErrPaperClosed", err)
	}
	if _, err := svc.Submit(ctx, user, session.ID, "written after time", "written after time"); err != nil {
		t.Fatalf("late submit: %v", err)
	}
	if eval.texts["ielts-writing-task1-figure"] != "saved task one" || eval.texts["ielts-writing-task2-opinion"] != "saved task two" {
		t.Fatalf("graded %v, want the drafts saved in time", eval.texts)
	}
}
