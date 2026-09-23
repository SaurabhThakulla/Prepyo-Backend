package speakingmock

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/testdb"
)

type fakeListener struct{ available bool }

func (f fakeListener) SpeakingAvailable() bool { return f.available }
func (f fakeListener) Transcribe(context.Context, []byte, string) (string, ai.Usage, error) {
	return "transcribed on the server", ai.Usage{}, nil
}

type fakeRater struct{ turns []ai.SpeakingTestTurn }

func (f *fakeRater) EvaluateSpeakingTest(_ context.Context, _ models.User, _ string, turns []ai.SpeakingTestTurn) (models.Evaluation, error) {
	f.turns = turns
	band := 6.5
	return models.Evaluation{Exam: models.ExamIELTS, Skill: models.SkillSpeaking, EstimatedScore: &band}, nil
}

func setup(t *testing.T, listener Transcriber) (*Service, *fakeRater, models.User) {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	var user models.User
	err = pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, referral_code)
		VALUES ('speaking-' || gen_random_uuid() || '@test.local', 'Speaking Test', 'free', 'Asia/Kathmandu',
		        'IELTS', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam`).Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID) })
	rater := &fakeRater{}
	return NewService(pool, mocks.NewRepository(pool), nil, gamification.NewService(), listener, rater), rater, user
}

// The test runs Part 1, the Part 2 long turn and Part 3, and Part 3 discusses
// the cue card's topic.
func TestSpeakingTestShape(t *testing.T) {
	svc, _, user := setup(t, fakeListener{})
	session, err := svc.Start(context.Background(), user, false)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	parts := map[int]int{}
	var cue *Step
	for i := range session.Steps {
		step := session.Steps[i]
		parts[step.Part]++
		if step.CueCard != nil {
			cue = &session.Steps[i]
		}
	}
	if parts[1] != 6 || parts[2] != 2 || parts[3] != 5 {
		t.Fatalf("steps per part = %v, want 6/2/5", parts)
	}
	if cue == nil || cue.PrepSeconds != 60 || cue.AnswerSeconds != 120 || len(cue.CueCard.Points) != 3 {
		t.Fatalf("cue card step = %+v, want 60s preparation and up to 2 minutes", cue)
	}
	var part3Lead string
	for _, step := range session.Steps {
		if step.Part == 3 && step.Lead != "" {
			part3Lead = step.Lead
		}
	}
	if !strings.Contains(part3Lead, "We've been talking about") {
		t.Fatalf("Part 3 does not link to the Part 2 topic: %q", part3Lead)
	}
}

// Answers are transcribed on the server where it can, otherwise the browser's
// transcript is kept; a re-recorded answer replaces the first take.
func TestSpeakingAnswersAreTranscribed(t *testing.T) {
	ctx := context.Background()
	audio := base64.StdEncoding.EncodeToString([]byte("RIFF-not-really-audio"))

	svc, _, user := setup(t, fakeListener{available: false})
	session, err := svc.Start(ctx, user, false)
	if err != nil {
		t.Fatal(err)
	}
	key := session.Steps[0].Key
	a, err := svc.SaveAnswer(ctx, user, session.ID, AnswerParams{Key: key, AudioBase64: audio, Format: "wav", DurationSeconds: 12, BrowserTranscript: "I live in a flat"})
	if err != nil || a.Source != "browser" || a.Transcript != "I live in a flat" {
		t.Fatalf("answer = %+v (%v), want the browser transcript", a, err)
	}

	svc2, _, user2 := setup(t, fakeListener{available: true})
	session2, err := svc2.Start(ctx, user2, false)
	if err != nil {
		t.Fatal(err)
	}
	a, err = svc2.SaveAnswer(ctx, user2, session2.ID, AnswerParams{Key: session2.Steps[0].Key, AudioBase64: audio, Format: "wav", DurationSeconds: 12, BrowserTranscript: "browser words"})
	if err != nil || a.Source != "server" || a.Transcript != "transcribed on the server" {
		t.Fatalf("answer = %+v (%v), want the server transcript", a, err)
	}
	if _, err := svc2.SaveAnswer(ctx, user2, session2.ID, AnswerParams{Key: "p9-9", BrowserTranscript: "x"}); !errors.Is(err, ErrUnknownStep) {
		t.Fatalf("unknown step error = %v", err)
	}
}

// The whole test is rated once, with every answer in test order.
func TestSpeakingTestIsRatedAsAWhole(t *testing.T) {
	ctx := context.Background()
	svc, rater, user := setup(t, fakeListener{})
	session, err := svc.Start(ctx, user, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveAnswer(ctx, user, session.ID, AnswerParams{Key: session.Steps[0].Key, BrowserTranscript: "only one answer"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Submit(ctx, user, session.ID); !errors.Is(err, ErrTooLittle) {
		t.Fatalf("submit with one answer = %v, want ErrTooLittle", err)
	}
	for _, step := range session.Steps {
		if _, err := svc.SaveAnswer(ctx, user, session.ID, AnswerParams{Key: step.Key, DurationSeconds: 20, BrowserTranscript: "answer to " + step.Key}); err != nil {
			t.Fatal(err)
		}
	}
	result, err := svc.Submit(ctx, user, session.ID)
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if len(rater.turns) != len(session.Steps) || rater.turns[0].Part != 1 || rater.turns[len(rater.turns)-1].Part != 3 {
		t.Fatalf("rated %d turns, want all %d in test order", len(rater.turns), len(session.Steps))
	}
	if result.Attempt == nil || result.Attempt.SkillScores[models.SkillSpeaking] != 6.5 {
		t.Fatalf("attempt = %+v, want a speaking score of 6.5", result.Attempt)
	}
}

// A test whose time ran out with too little said is closed rather than left
// open, so the learner can start another.
func TestExpiredSpeakingTestWithTooLittleIsClosed(t *testing.T) {
	ctx := context.Background()
	svc, _, user := setup(t, fakeListener{})
	session, err := svc.Start(ctx, user, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.db.Exec(ctx, `UPDATE speaking_mock_sessions SET expires_at = now() - interval '1 minute' WHERE id = $1`, session.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Submit(ctx, user, session.ID); !errors.Is(err, ErrExpiredTooLittle) {
		t.Fatalf("submit = %v, want ErrExpiredTooLittle", err)
	}
	next, err := svc.Start(ctx, user, false)
	if err != nil || next.ID == session.ID {
		t.Fatalf("next start = %s (%v), want a new test", next.ID, err)
	}
}
