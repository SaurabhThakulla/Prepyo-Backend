package ptemock

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/reading"
	"github.com/prepyo/backend/internal/scoring"
	"github.com/prepyo/backend/internal/testdb"
)

// fakeEvaluator rates every answer three quarters of the marks, and counts
// the calls, so the tests can see which items reached an evaluator.
type fakeEvaluator struct{ spoken, written int }

func (f *fakeEvaluator) rating() models.Evaluation {
	return models.Evaluation{Summary: "Rated.", Criteria: []models.EvaluationCriterion{
		{Name: "Content", Score: 2, MaxScore: 2}, {Name: "Form", Score: 1, MaxScore: 2},
	}}
}

func (f *fakeEvaluator) EvaluatePTEMockSpeaking(context.Context, string, models.Question, string, int, scoring.Delivery, string) (models.Evaluation, error) {
	f.spoken++
	return f.rating(), nil
}

func (f *fakeEvaluator) EvaluatePTEMockWriting(context.Context, string, models.Question, string, string) (models.Evaluation, error) {
	f.written++
	return f.rating(), nil
}

// fakeTranscriber stands in for Whisper on the audio provider.
type fakeTranscriber struct {
	available bool
	text      string
	fail      bool
	calls     int
}

func (f *fakeTranscriber) SpeakingAvailable() bool { return f.available }

func (f *fakeTranscriber) Transcribe(context.Context, []byte, string) (string, ai.Usage, error) {
	f.calls++
	if f.fail {
		return "", ai.Usage{}, ai.ErrUnavailable
	}
	return f.text, ai.Usage{}, nil
}

type fixture struct {
	svc     *Service
	billing *billing.Service
	pool    *pgxpool.Pool
	user    models.User
	eval    *fakeEvaluator
	ear     *fakeTranscriber
	prefix  string
}

func setup(t *testing.T) fixture {
	t.Helper()
	ctx := context.Background()
	// The same statement mode as the API's pool (database.Connect), which is
	// stricter about parameter types than pgx's default.
	cfg, err := pgxpool.ParseConfig(testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := database.Migrate(ctx, pool, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var user models.User
	err = pool.QueryRow(ctx, `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, referral_code)
		VALUES ('ptemock-' || gen_random_uuid() || '@test.local', 'PTE Mock Test', 'free', 'Asia/Kathmandu',
		        'PTE', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam, bonus_mock_tests`).
		Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam, &user.BonusMockTests)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	prefix := "ptemock-test-" + strings.ReplaceAll(user.ID[:8], "-", "") + "-"
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID)
		_, _ = pool.Exec(context.Background(), `DELETE FROM questions WHERE id LIKE $1`, prefix+"%")
	})

	eval := &fakeEvaluator{}
	ear := &fakeTranscriber{available: true, text: "the library opens at nine every morning"}
	billingService := billing.NewService(billing.NewRepository(pool), nil)
	svc := NewService(pool, questions.NewRepository(pool), reading.NewRepository(pool), mocks.NewRepository(pool),
		billingService, gamification.NewService(), eval, ear, nil)
	svc.manualScoring = true
	f := fixture{svc: svc, billing: billingService, pool: pool, user: user, eval: eval, ear: ear, prefix: prefix}
	f.seed(t)
	return f
}

// seed adds one item of several tasks, so every kind can be dealt whatever
// else the test database holds.
func (f fixture) seed(t *testing.T) {
	t.Helper()
	rows := []struct {
		id, skill, typeID, passage, options, correct, blanks string
	}{
		{"ra", "speaking", "read-aloud", "The library opens at nine every morning.", "", "", ""},
		{"swt", "writing", "summarize-written-text", "A passage to summarise.", "", "", ""},
		{"we", "writing", "pte-write-essay", "", "", "", ""},
		{"rwfib", "reading", "fill-in-blanks-rw", "The [[b1]] sat on the mat.",
			"", "", `[{"id":"b1","options":["cat","dog"],"correctAnswer":"cat"}]`},
		{"ro", "reading", "reorder-paragraphs", "",
			`[{"id":"A","text":"First."},{"id":"B","text":"Second."},{"id":"C","text":"Third."}]`, `["A","B","C"]`, ""},
		{"wfd", "listening", "write-from-dictation", "", "", `["the library opens at nine"]`, ""},
	}
	for _, r := range rows {
		_, err := f.pool.Exec(context.Background(), `
			INSERT INTO questions (id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
			                       context_passage, audio_transcript, time_limit_seconds, options, correct_answers,
			                       blanks, supported_exams)
			VALUES ($1, 'pte-2026-01', 'PTE', $2, $3, $3, 'Test item', 'Answer this.', NULLIF($4, ''),
			        CASE WHEN $2 = 'listening' THEN 'The library opens at nine.' END, 60,
			        NULLIF($5, '')::jsonb, NULLIF($6, '')::jsonb, NULLIF($7, '')::jsonb, ARRAY['PTE'])`,
			f.prefix+r.id, r.skill, r.typeID, r.passage, r.options, r.correct, r.blanks)
		if err != nil {
			t.Fatalf("seed %s: %v", r.id, err)
		}
	}
}

// answerAll answers every item in order with what answer gives it.
func (f fixture) answerAll(t *testing.T, view View, answer func(ItemView) Answer) View {
	t.Helper()
	ctx := context.Background()
	for view.Status == StatusInProgress && view.Item != nil {
		next, err := f.svc.Submit(ctx, f.user, view.ID, view.Item.Position, answer(*view.Item))
		if err != nil {
			t.Fatalf("submit %d: %v", view.Item.Position, err)
		}
		if next.Current <= view.Current && next.Status == StatusInProgress {
			t.Fatalf("the test did not move on from %d", view.Current)
		}
		view = next
	}
	return view
}

func TestSectionalReadingRunsForwardAndScores(t *testing.T) {
	ctx := context.Background()
	f := setup(t)

	view, err := f.svc.Start(ctx, f.user, KindReading)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if view.Item == nil || view.Item.Position != 1 || view.Clock == nil || view.Clock.Scope != "part" {
		t.Fatalf("first item on the reading clock: %+v", view)
	}
	if view.Item.Question.Blanks != nil {
		for _, b := range view.Item.Question.Blanks {
			if b.CorrectAnswer != "" {
				t.Fatal("the answer key reached the screen")
			}
		}
	}

	again, err := f.svc.Start(ctx, f.user, KindReading)
	if err != nil || again.ID != view.ID {
		t.Fatalf("a second start resumes the open paper: %s %v", again.ID, err)
	}
	state, err := f.billing.State(ctx, f.pool, f.user)
	if err != nil {
		t.Fatal(err)
	}
	if state.DailySubTestsUsed != billing.SectionMockSubTests {
		t.Fatalf("sub-tests used %d, want %d charged once", state.DailySubTestsUsed, billing.SectionMockSubTests)
	}

	if view.TotalItems > 1 {
		if _, err := f.svc.Submit(ctx, f.user, view.ID, 2, Answer{}); !errors.Is(err, ErrOutOfOrder) {
			t.Fatalf("answering ahead: %v", err)
		}
	}

	view = f.answerAll(t, view, func(item ItemView) Answer {
		switch item.Question.ID {
		case f.prefix + "rwfib":
			return Answer{Response: &models.AnswerSubmission{BlankResponses: map[string]string{"b1": "cat"}}}
		case f.prefix + "ro":
			return Answer{Response: &models.AnswerSubmission{SelectedOptions: []string{"A", "B", "C"}}}
		}
		return Answer{}
	})
	if view.Status != StatusScoring {
		t.Fatalf("status %s after the last item", view.Status)
	}
	if _, err := f.svc.Submit(ctx, f.user, view.ID, 1, Answer{}); !errors.Is(err, ErrNotInProgress) {
		t.Fatalf("answering after the end: %v", err)
	}

	if err := f.svc.ScorePaper(ctx, f.user, view.ID); err != nil {
		t.Fatalf("score: %v", err)
	}
	report, err := f.svc.Report(ctx, f.user, view.ID)
	if err != nil {
		t.Fatalf("report: %v", err)
	}
	// Other packages' tests share the database and may delete a question they
	// made while it sits on this paper; its item goes with it, so the review can
	// be shorter than the paper, never longer.
	if report.Status != StatusCompleted || report.Result == nil || len(report.Review) == 0 || len(report.Review) > report.TotalItems {
		t.Fatalf("report %+v", report.View)
	}
	score, ok := report.Result.Skills[models.SkillReading]
	if !ok || score < MinScore || score > MaxScore {
		t.Fatalf("reading score %d %v", score, ok)
	}
	for _, r := range report.Review {
		if r.Question.ID == f.prefix+"rwfib" || r.Question.ID == f.prefix+"ro" {
			if r.Percent == nil || *r.Percent != 100 {
				t.Fatalf("%s marked %v", r.Question.ID, r.Percent)
			}
		}
	}
	var attempts int
	if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM mock_attempts WHERE user_id = $1 AND mock_id = 'mock-pte-reading'`,
		f.user.ID).Scan(&attempts); err != nil || attempts != 1 {
		t.Fatalf("attempts %d %v", attempts, err)
	}

	// Scoring again changes nothing.
	if err := f.svc.ScorePaper(ctx, f.user, view.ID); err != nil {
		t.Fatal(err)
	}
	if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM mock_attempts WHERE user_id = $1 AND mock_id = 'mock-pte-reading'`,
		f.user.ID).Scan(&attempts); err != nil || attempts != 1 {
		t.Fatalf("attempts after a second run %d %v", attempts, err)
	}
}

func TestWritingFormRulesAndRatings(t *testing.T) {
	ctx := context.Background()
	f := setup(t)

	view, err := f.svc.Start(ctx, f.user, KindWriting)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if view.Clock == nil || view.Clock.Scope != "item" || view.Clock.TotalSeconds != 600 {
		t.Fatalf("Summarize Written Text has its own ten minutes: %+v", view.Clock)
	}

	// A draft survives a reload.
	draft := &models.AnswerSubmission{TextResponse: "a draft"}
	if _, err := f.svc.SaveDraft(ctx, f.user, view.ID, view.Current, draft); err != nil {
		t.Fatalf("draft: %v", err)
	}
	reloaded, err := f.svc.Get(ctx, f.user, view.ID)
	if err != nil || reloaded.Item.Draft == nil || reloaded.Item.Draft.TextResponse != "a draft" {
		t.Fatalf("draft after reload: %+v %v", reloaded.Item, err)
	}

	sentence := "The passage argues that libraries remain vital community spaces because they offer free access to knowledge and quiet study."
	view = f.answerAll(t, reloaded, func(item ItemView) Answer {
		if item.Task.Code == "SWT" {
			return Answer{Response: &models.AnswerSubmission{TextResponse: sentence}}
		}
		// Far under the essay's 120 words: PTE scores it zero on form.
		return Answer{Response: &models.AnswerSubmission{TextResponse: "Too short to be an essay."}}
	})
	if err := f.svc.ScorePaper(ctx, f.user, view.ID); err != nil {
		t.Fatalf("score: %v", err)
	}
	report, err := f.svc.Report(ctx, f.user, view.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range report.Review {
		switch r.Task {
		case "SWT":
			if r.Percent == nil || *r.Percent != 75 {
				t.Fatalf("summary rated %v", r.Percent)
			}
		case "WE":
			if r.Percent == nil || *r.Percent != 0 || !strings.Contains(r.Feedback, "120") {
				t.Fatalf("short essay %v %q", r.Percent, r.Feedback)
			}
		}
	}
	if f.eval.written == 0 {
		t.Fatal("the summary was not rated")
	}
	if _, ok := report.Result.Skills[models.SkillWriting]; !ok {
		t.Fatalf("no writing score: %+v", report.Result)
	}
}

func TestFullMockSpendsTheAllowanceAtStart(t *testing.T) {
	ctx := context.Background()
	f := setup(t)

	view, err := f.svc.Start(ctx, f.user, KindFull)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if len(view.Parts) != 3 {
		t.Fatalf("parts %+v", view.Parts)
	}
	state, err := f.billing.State(ctx, f.pool, f.user)
	if err != nil {
		t.Fatal(err)
	}
	if state.MockTestsUsed != 1 {
		t.Fatalf("mock tests used %d, want 1", state.MockTestsUsed)
	}

	// Ending early skips what is left and scores the paper.
	view, err = f.svc.Finish(ctx, f.user, view.ID)
	if err != nil || view.Status != StatusScoring {
		t.Fatalf("finish %s %v", view.Status, err)
	}
	if err := f.svc.ScorePaper(ctx, f.user, view.ID); err != nil {
		t.Fatalf("score: %v", err)
	}
	report, err := f.svc.Report(ctx, f.user, view.ID)
	if err != nil {
		t.Fatal(err)
	}
	if report.Result == nil || report.Result.Overall == nil || *report.Result.Overall != MinScore {
		t.Fatalf("an unanswered full test scores the bottom of the scale: %+v", report.Result)
	}
	state, err = f.billing.State(ctx, f.pool, f.user)
	if err != nil {
		t.Fatal(err)
	}
	if state.MockTestsUsed != 1 {
		t.Fatalf("mock tests used %d after finishing, want still 1", state.MockTestsUsed)
	}
}

func TestPartClockRunsOutWhileAway(t *testing.T) {
	ctx := context.Background()
	f := setup(t)

	view, err := f.svc.Start(ctx, f.user, KindReading)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	later := time.Now().Add(time.Duration(view.Clock.TotalSeconds)*time.Second + SubmitGrace + time.Minute)
	f.svc.now = func() time.Time { return later }

	view, err = f.svc.Get(ctx, f.user, view.ID)
	if err != nil {
		t.Fatal(err)
	}
	if view.Status != StatusScoring {
		t.Fatalf("a reading paper out of time goes to scoring, got %s", view.Status)
	}
	var timedOut int
	if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM pte_mock_items WHERE session_id = $1 AND status = 'timed_out'`,
		view.ID).Scan(&timedOut); err != nil || timedOut != view.TotalItems {
		t.Fatalf("timed out %d of %d (%v)", timedOut, view.TotalItems, err)
	}
}

func TestOtherExamsCannotStart(t *testing.T) {
	f := setup(t)
	ielts := f.user
	ielts.TargetExam = models.ExamIELTS
	if _, err := f.svc.Start(context.Background(), ielts, KindFull); !errors.Is(err, ErrNotPTE) {
		t.Fatalf("IELTS learner: %v", err)
	}
	if _, err := f.svc.Start(context.Background(), f.user, Kind("mini")); !errors.Is(err, ErrUnknownKind) {
		t.Fatalf("unknown kind: %v", err)
	}
}

// Spoken answers are transcribed as the IELTS speaking mock transcribes them:
// the server's transcript wherever it has one, the browser's when it fails.
func TestServerTranscribesSpokenAnswersLikeTheIELTSMock(t *testing.T) {
	ctx := context.Background()
	f := setup(t)

	view, err := f.svc.Start(ctx, f.user, KindSpeaking)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	recording := base64.StdEncoding.EncodeToString([]byte("fake mp3 bytes"))
	stored := func(position int) (string, string) {
		t.Helper()
		var transcript, source string
		if err := f.pool.QueryRow(ctx, `
			SELECT COALESCE(transcript, ''), COALESCE(transcript_source, '') FROM pte_mock_items
			 WHERE session_id = $1 AND position = $2`, view.ID, position).Scan(&transcript, &source); err != nil {
			t.Fatal(err)
		}
		return transcript, source
	}

	// Whisper's transcript replaces what the browser heard.
	first := view.Item.Position
	view, err = f.svc.Submit(ctx, f.user, view.ID, first,
		Answer{Transcript: "the library open at nine", DurationSeconds: 12, Audio: recording, Format: "mp3"})
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if transcript, source := stored(first); transcript != f.ear.text || source != SourceServer {
		t.Fatalf("stored %q from %q, want Whisper's", transcript, source)
	}
	if view.Item == nil {
		return
	}

	// Whisper fails: the browser's transcript is kept, and nothing is lost.
	f.ear.fail = true
	second := view.Item.Position
	if _, err := f.svc.Submit(ctx, f.user, view.ID, second,
		Answer{Transcript: "my own words", DurationSeconds: 5, Audio: recording, Format: "mp3"}); err != nil {
		t.Fatalf("submit: %v", err)
	}
	if transcript, source := stored(second); transcript != "my own words" || source != SourceBrowser {
		t.Fatalf("stored %q from %q, want the browser's", transcript, source)
	}
	if f.ear.calls != 2 {
		t.Fatalf("whisper calls %d, want one per answer", f.ear.calls)
	}
}

func TestTranscriptionOnlyForTheSpokenItemOnScreen(t *testing.T) {
	ctx := context.Background()
	f := setup(t)
	recording := base64.StdEncoding.EncodeToString([]byte("fake"))
	// Two sectional tests in a day take more sub-tests than the free plan has.
	validUntil := time.Now().Add(30 * 24 * time.Hour)
	if _, err := f.pool.Exec(ctx, `UPDATE users SET plan_id = 'pro', plan_valid_until = $2 WHERE id = $1`,
		f.user.ID, validUntil); err != nil {
		t.Fatal(err)
	}
	f.user.PlanID, f.user.PlanValidUntil = "pro", &validUntil

	speaking, err := f.svc.Start(ctx, f.user, KindSpeaking)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, ok := f.svc.transcribe(ctx, f.user.ID, speaking.ID, speaking.Item.Position+1, Answer{Audio: recording, Format: "mp3"}); ok {
		t.Fatal("transcribed an item that is not on screen")
	}
	if _, ok := f.svc.transcribe(ctx, f.user.ID, speaking.ID, speaking.Item.Position, Answer{Audio: recording, Format: "ogg"}); ok {
		t.Fatal("transcribed a format the provider does not take")
	}

	reading, err := f.svc.Start(ctx, f.user, KindReading)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, ok := f.svc.transcribe(ctx, f.user.ID, reading.ID, reading.Item.Position, Answer{Audio: recording, Format: "mp3"}); ok {
		t.Fatal("transcribed audio sent for a reading item")
	}

	f.ear.available = false
	if _, ok := f.svc.transcribe(ctx, f.user.ID, speaking.ID, speaking.Item.Position, Answer{Audio: recording, Format: "mp3"}); ok {
		t.Fatal("transcribed with no audio provider configured")
	}
	if f.ear.calls != 0 {
		t.Fatalf("whisper calls %d, want none", f.ear.calls)
	}
}
