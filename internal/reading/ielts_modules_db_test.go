package reading

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
)

func setModule(t *testing.T, pool *pgxpool.Pool, user *models.User, module string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(),
		`UPDATE users SET target_module = $2 WHERE id = $1`, user.ID, module); err != nil {
		t.Fatalf("set module: %v", err)
	}
	user.TargetModule = module
}

type passageFacts struct {
	words   int
	tags    []string
	modules []string
}

func factsOf(t *testing.T, pool *pgxpool.Pool, id string) passageFacts {
	t.Helper()
	var f passageFacts
	if err := pool.QueryRow(context.Background(),
		`SELECT word_count, tags, modules FROM reading_passages WHERE id = $1`, id).
		Scan(&f.words, &f.tags, &f.modules); err != nil {
		t.Fatalf("read passage %s: %v", id, err)
	}
	return f
}

// A General Training paper is everyday texts, then workplace texts, then a
// longer general-interest article, 40 questions and 2,150-2,375 words.
func TestGeneralTrainingPaperComposes(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	ctx := context.Background()

	user := newLearner(t, pool, models.ExamIELTS)
	setModule(t, pool, &user, models.ModuleGeneralTraining)

	blueprint, err := svc.repo.GeneratedBlueprint(ctx, models.ExamIELTS, models.ModuleGeneralTraining)
	if err != nil {
		t.Fatalf("General Training blueprint: %v", err)
	}
	for range 5 {
		composed, err := svc.compose(ctx, user.ID, models.ExamIELTS, blueprint)
		if err != nil {
			t.Fatalf("compose General Training paper: %v", err)
		}
		if len(composed.QuestionIDs) != 40 || len(composed.PassageIDs) != 3 {
			t.Fatalf("paper = %d questions / %d passages, want 40 / 3", len(composed.QuestionIDs), len(composed.PassageIDs))
		}
		total := 0
		for i, id := range composed.PassageIDs {
			f := factsOf(t, pool, id)
			total += f.words
			if !slices.Contains(f.modules, models.ModuleGeneralTraining) {
				t.Errorf("section %d passage %s is not a General Training text", i+1, id)
			}
			if want := []string{"GT Section 1", "GT Section 2"}; i < 2 && !slices.Contains(f.tags, want[i]) {
				t.Errorf("section %d passage %s lacks tag %q", i+1, id, want[i])
			}
		}
		if total < 2150 || total > 2375 {
			t.Errorf("paper is %d words, want the official General Training range 2,150-2,375", total)
		}
	}
}

// An Academic paper never deals everyday or workplace General Training texts,
// and stays in the official 2,150-2,750 word range.
func TestAcademicPaperKeepsToAcademicTexts(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	ctx := context.Background()

	user := newLearner(t, pool, models.ExamIELTS)
	blueprint, err := svc.repo.GeneratedBlueprint(ctx, models.ExamIELTS, models.ModuleAcademic)
	if err != nil {
		t.Fatalf("Academic blueprint: %v", err)
	}
	for range 5 {
		composed, err := svc.compose(ctx, user.ID, models.ExamIELTS, blueprint)
		if err != nil {
			t.Fatalf("compose Academic paper: %v", err)
		}
		total := 0
		for _, id := range composed.PassageIDs {
			f := factsOf(t, pool, id)
			total += f.words
			if !slices.Contains(f.modules, models.ModuleAcademic) {
				t.Errorf("Academic paper dealt %s, a General Training-only text", id)
			}
		}
		if total < 2150 || total > 2750 {
			t.Errorf("paper is %d words, want the official Academic range 2,150-2,750", total)
		}
	}

	group, err := svc.repo.PickPracticeGroup(ctx, user.ID, models.ExamIELTS, models.ModuleAcademic,
		[]string{TypeMatchingInformation})
	if err != nil {
		t.Fatalf("pick practice: %v", err)
	}
	if f := factsOf(t, pool, group.PassageID); !slices.Contains(f.modules, models.ModuleAcademic) {
		t.Errorf("Academic practice dealt %s, a General Training-only text", group.PassageID)
	}
}

// For multiple choice, sentence completion and short answers the official
// format says questions follow the text; TFNG and YNNG sets do in practice
// papers. Every IELTS set is checked against the paragraphs it cites.
func TestIELTSOrderedSetsFollowTheText(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT g.id, q.explanation
		  FROM reading_question_groups g
		  JOIN questions q ON q.group_id = g.id
		 WHERE 'IELTS' = ANY(q.supported_exams)
		   AND g.type_id IN ('reading-true-false', 'reading-yes-no-not-given',
		                     'reading-sentence-completion', 'reading-mcq-single', 'reading-mcq-multiple')
		 ORDER BY g.id, q.group_position`)
	if err != nil {
		t.Fatalf("read sets: %v", err)
	}
	defer rows.Close()

	cited := regexp.MustCompile(`[Pp]aragraphs? ([A-F])`)
	last := map[string]string{}
	for rows.Next() {
		var group, explanation string
		if err := rows.Scan(&group, &explanation); err != nil {
			t.Fatal(err)
		}
		m := cited.FindStringSubmatch(explanation)
		if m == nil {
			continue
		}
		if m[1] < last[group] {
			t.Errorf("%s: a question citing paragraph %s comes after one citing %s", group, m[1], last[group])
		}
		last[group] = m[1]
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(last) == 0 {
		t.Fatal("no ordered IELTS sets found to check")
	}
}

// correctAnswers answers every question on a paper with its key.
func correctAnswers(t *testing.T, pool *pgxpool.Pool, ids []string) []models.AnswerSubmission {
	t.Helper()
	bank, err := questions.NewRepository(pool).ByIDs(context.Background(), ids)
	if err != nil {
		t.Fatalf("load paper: %v", err)
	}
	answers := make([]models.AnswerSubmission, 0, len(ids))
	for _, id := range ids {
		q := bank[id]
		answer := models.AnswerSubmission{QuestionID: id}
		if len(q.Options) > 0 {
			answer.SelectedOptions = q.CorrectAnswers
		} else if len(q.CorrectAnswers) > 0 {
			answer.TextResponse = q.CorrectAnswers[0]
		}
		answers = append(answers, answer)
	}
	return answers
}

// The server keeps the paper's time. Answers saved in time come back on
// resume, and a submission after the deadline is graded from those answers,
// not from whatever the browser sends late.
func TestReadingMockKeepsServerTime(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	ctx := context.Background()

	user := newLearner(t, pool, models.ExamIELTS)
	session, err := svc.StartMock(ctx, user, models.ExamIELTS)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if session.SecondsRemaining < 3590 || session.SecondsRemaining > 3600 {
		t.Fatalf("seconds remaining = %d, want about 3600 for a 60-minute paper", session.SecondsRemaining)
	}

	var ids []string
	for _, set := range session.Sets {
		for _, g := range set.Groups {
			for _, q := range g.Questions {
				ids = append(ids, q.ID)
			}
		}
	}
	all := correctAnswers(t, pool, ids)

	// Blank papers are not results.
	if _, err := svc.SubmitMock(ctx, user, session.ID, nil, 0); !errors.Is(err, ErrNotAttempted) {
		t.Fatalf("blank submit error = %v, want ErrNotAttempted", err)
	}

	// Save three right answers, then reopen: they come back.
	if _, err := svc.SaveDrafts(ctx, user, session.ID, all[:3]); err != nil {
		t.Fatalf("save drafts: %v", err)
	}
	resumed, err := svc.ResumeMock(ctx, user, session.ID)
	if err != nil {
		t.Fatalf("resume: %v", err)
	}
	if len(resumed.DraftAnswers) != 3 {
		t.Fatalf("resumed with %d saved answers, want 3", len(resumed.DraftAnswers))
	}
	if resumed.SecondsRemaining > session.SecondsRemaining {
		t.Fatalf("resuming reset the clock: %d > %d", resumed.SecondsRemaining, session.SecondsRemaining)
	}

	// Time runs out. A late submission carrying all forty right answers is
	// graded from the three saved in time.
	if _, err := pool.Exec(ctx, `UPDATE reading_mock_sessions SET expires_at = now() - interval '10 minutes' WHERE id = $1`, session.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveDrafts(ctx, user, session.ID, all); !errors.Is(err, ErrPaperClosed) {
		t.Fatalf("late draft save error = %v, want ErrPaperClosed", err)
	}
	result, err := svc.SubmitMock(ctx, user, session.ID, all, 1)
	if err != nil {
		t.Fatalf("late submit: %v", err)
	}
	if result.Attempt.TotalCorrect != 3 || result.Attempt.TotalQuestions != 40 {
		t.Fatalf("late paper scored %d/%d, want 3/40 from the answers saved in time",
			result.Attempt.TotalCorrect, result.Attempt.TotalQuestions)
	}
	if result.Attempt.DurationSeconds > session.DurationMinutes*60 {
		t.Fatalf("duration %ds exceeds the paper's %d minutes", result.Attempt.DurationSeconds, session.DurationMinutes)
	}
}

// Time running out on a paper with no answers closes it without a band.
func TestExpiredBlankPaperClosesWithoutAResult(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	ctx := context.Background()

	user := newLearner(t, pool, models.ExamIELTS)
	session, err := svc.StartMock(ctx, user, models.ExamIELTS)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE reading_mock_sessions SET expires_at = now() - interval '10 minutes' WHERE id = $1`, session.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SubmitMock(ctx, user, session.ID, nil, 0); !errors.Is(err, ErrExpiredUnattempted) {
		t.Fatalf("error = %v, want ErrExpiredUnattempted", err)
	}
	closed, err := svc.repo.SessionByID(ctx, pool, user.ID, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if closed.Status != StatusAbandoned {
		t.Fatalf("status = %s, want %s", closed.Status, StatusAbandoned)
	}
}
