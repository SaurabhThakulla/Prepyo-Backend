package listeningmock

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/testdb"
)

func setup(t *testing.T) (*pgxpool.Pool, *Service, models.User) {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	var user models.User
	err = pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, referral_code)
		VALUES ('listening-' || gen_random_uuid() || '@test.local', 'Listening Test', 'free', 'Asia/Kathmandu',
		        'IELTS', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam`).Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID) })
	svc := NewService(pool, questions.NewRepository(pool), mocks.NewRepository(pool), nil, gamification.NewService())
	return pool, svc, user
}

// Four parts of ten, numbered 1-40, with no script anywhere in the paper.
func TestListeningPaperShape(t *testing.T) {
	_, svc, user := setup(t)
	session, err := svc.Start(context.Background(), user, StartOptions{})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if len(session.Parts) != 4 {
		t.Fatalf("parts = %d, want 4", len(session.Parts))
	}
	number := 0
	for i, part := range session.Parts {
		if part.PartNo != i+1 || part.FirstNumber != i*10+1 || part.LastNumber != i*10+10 {
			t.Errorf("part %d numbered %d-%d", part.PartNo, part.FirstNumber, part.LastNumber)
		}
		for _, g := range part.Groups {
			for _, q := range g.Questions {
				number++
				if q.AudioTranscript != "" || len(q.CorrectAnswers) != 0 {
					t.Fatalf("question %s leaks its script or key", q.ID)
				}
			}
		}
	}
	if number != 40 {
		t.Fatalf("questions = %d, want 40", number)
	}
	body, _ := json.Marshal(session)
	if strings.Contains(string(body), "Receptionist:") {
		t.Fatal("the paper carries a recording script")
	}
	script, err := svc.Script(context.Background(), user, session.ID, 1)
	if err != nil || !strings.HasPrefix(script, "Receptionist:") {
		t.Fatalf("part 1 script = %.30q (%v)", script, err)
	}
}

// Every key, submitted as a learner would type or pick it, scores 40/40 and
// band 9; one wrong answer costs one mark.
func TestListeningPaperMarksEachAnswer(t *testing.T) {
	pool, svc, user := setup(t)
	ctx := context.Background()
	session, err := svc.Start(ctx, user, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	bank, err := questions.NewRepository(pool).ByIDs(ctx, session.questionIDs)
	if err != nil {
		t.Fatal(err)
	}
	var answers []models.AnswerSubmission
	for i, id := range session.questionIDs {
		q := bank[id]
		a := models.AnswerSubmission{QuestionID: id}
		if len(q.Options) > 0 {
			a.SelectedOptions = q.CorrectAnswers
		} else {
			a.TextResponse = strings.ToUpper(q.CorrectAnswers[0]) + "."
		}
		if i == 0 {
			a.TextResponse = "Hartly" // misspelt, as IELTS marks it wrong
		}
		answers = append(answers, a)
	}
	if _, err := svc.Submit(ctx, user, session.ID, nil); !errors.Is(err, ErrNotAttempted) {
		t.Fatalf("blank submit = %v, want ErrNotAttempted", err)
	}
	result, err := svc.Submit(ctx, user, session.ID, answers)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if result.Attempt.TotalCorrect != 39 || result.Attempt.TotalQuestions != 40 {
		t.Fatalf("marks = %d/%d, want 39/40", result.Attempt.TotalCorrect, result.Attempt.TotalQuestions)
	}
	if result.Attempt.UserScore != 9 || result.Attempt.SkillScores[models.SkillListening] != 9 {
		t.Fatalf("band = %v, want 9 for 39/40", result.Attempt.UserScore)
	}
	if len(result.Review) != 40 || len(result.Scripts) != 4 {
		t.Fatalf("review = %d, scripts = %d", len(result.Review), len(result.Scripts))
	}
}

// Mock questions belong to their recordings and never appear as standalone
// practice items, which would have no audio.
func TestListeningMockItemsStayOutOfPractice(t *testing.T) {
	pool, _, _ := setup(t)
	list, _, err := questions.NewRepository(pool).List(context.Background(), questions.ListParams{
		Exam: models.ExamIELTS, Skill: models.SkillListening, Limit: 500,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range list {
		if strings.HasPrefix(q.ID, "lt-") {
			t.Fatalf("practice list includes mock item %s", q.ID)
		}
	}
}

// A full mock's paper is dealt fresh beside a listening mock the learner has
// open, and a later section-mock start still finds the learner's own paper.
func TestFullMockPaperSitsBesideAnOpenSectionMock(t *testing.T) {
	ctx := context.Background()
	pool, svc, user := setup(t)
	own, err := svc.Start(ctx, user, StartOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// Out of time and never submitted: a full mock must not pick it up.
	if _, err := pool.Exec(ctx, `UPDATE listening_mock_sessions SET expires_at = now() - interval '1 hour' WHERE id = $1`, own.ID); err != nil {
		t.Fatal(err)
	}
	full, err := svc.Start(ctx, user, StartOptions{FullMock: true})
	if err != nil {
		t.Fatalf("full mock paper: %v", err)
	}
	if full.ID == own.ID || full.SecondsRemaining <= 0 {
		t.Fatalf("full mock paper = %s with %ds left, want a fresh paper beside %s", full.ID, full.SecondsRemaining, own.ID)
	}
	again, err := svc.Start(ctx, user, StartOptions{})
	if err != nil || again.ID != own.ID {
		t.Fatalf("section mock start = %s (%v), want the learner's own paper %s", again.ID, err, own.ID)
	}
}
