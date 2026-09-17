package questions

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
	"github.com/prepyo/backend/internal/testdb"
	"github.com/prepyo/backend/migrations"
)

func TestIELTSReadingSubtaskBank(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	for _, table := range []string{"reading_passages", "reading_question_groups", "questions"} {
		if _, err = tx.Exec(ctx, "CREATE TEMP TABLE "+table+" (LIKE public."+table+" INCLUDING ALL) ON COMMIT DROP"); err != nil {
			t.Fatal(err)
		}
	}
	run := func(file string) {
		t.Helper()
		body, err := migrations.Files.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
			t.Fatal(err)
		}
	}
	const up = "000054_ielts_reading_subtasks.up.sql"
	run(up)
	run(up)
	repo := NewRepository(tx)
	bank, total, err := repo.List(ctx, ListParams{Exam: models.ExamIELTS, Limit: 100, IncludePassageQuestions: true})
	if err != nil {
		t.Fatal(err)
	}
	if total != 25 || len(bank) != 25 {
		t.Fatalf("bank=%d total=%d", len(bank), total)
	}
	counts := map[string]int{}
	for _, q := range bank {
		key := q.TypeID
		if key == "reading-mcq-single" || key == "reading-mcq-multiple" {
			key = "multiple-choice"
		}
		counts[key]++
		if q.Skill != models.SkillReading || !q.SupportsExam(models.ExamIELTS) || q.SupportsExam(models.ExamPTE) || q.Explanation == "" {
			t.Errorf("invalid content: %s", q.ID)
		}
		sub := models.AnswerSubmission{QuestionID: q.ID, SelectedOptions: q.CorrectAnswers}
		if q.TypeID == "reading-sentence-completion" {
			sub.SelectedOptions = nil
			sub.TextResponse = q.CorrectAnswers[0]
			if !strings.Contains(q.Prompt, "________") {
				t.Errorf("missing gap: %s", q.ID)
			}
		} else {
			for _, answer := range q.CorrectAnswers {
				found := false
				for _, option := range q.Options {
					if option.ID == answer {
						found = true
					}
				}
				if !found {
					t.Errorf("unselectable key %s/%s", q.ID, answer)
				}
			}
		}
		result, ok := scoring.Grade(q, sub)
		if !ok || !result.IsCorrect || result.AccuracyPercentage != 100 {
			t.Errorf("ungradable %s: %+v", q.ID, result)
		}
		empty, ok := scoring.Grade(q, models.AnswerSubmission{})
		if !ok || empty.Score != 0 {
			t.Errorf("empty answer scores: %s", q.ID)
		}
	}
	for _, key := range []string{"multiple-choice", "reading-true-false", "reading-yes-no-not-given", "reading-matching-information", "reading-sentence-completion"} {
		if counts[key] != 5 {
			t.Errorf("%s: got %d, want 5", key, counts[key])
		}
	}
	if len(counts) != 5 {
		t.Errorf("unexpected subtasks: %v", counts)
	}
	var linked int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM questions q
 JOIN reading_question_groups g ON q.group_id=g.id AND q.type_id=g.type_id AND q.passage_id=g.passage_id
 JOIN reading_passages p ON p.id=q.passage_id
 WHERE q.is_published AND p.is_published AND g.passage_display='full'
 AND jsonb_array_length(p.paragraphs)=6 AND p.word_count>=700`).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if linked != 25 {
		t.Fatalf("only %d questions linked to published passage/groups", linked)
	}
	if _, err = tx.Exec(ctx, `UPDATE questions SET title='Editorial change' WHERE id='q-ir54-single-1'`); err != nil {
		t.Fatal(err)
	}
	run(up)
	q, err := repo.ByID(ctx, "q-ir54-single-1")
	if err != nil || q.Title != "Editorial change" {
		t.Fatalf("replay overwrites edit: %v", err)
	}
	run("000054_ielts_reading_subtasks.down.sql")
	for _, table := range []string{"questions", "reading_question_groups", "reading_passages"} {
		var count int
		if err = tx.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("rollback left %d rows in %s", count, table)
		}
	}
}
