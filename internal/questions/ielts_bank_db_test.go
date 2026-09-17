package questions

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
	"github.com/prepyo/backend/internal/testdb"
	"github.com/prepyo/backend/migrations"
)

func TestIELTSPracticeBank(t *testing.T) {
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
	if _, err = tx.Exec(ctx, `CREATE TEMP TABLE questions (LIKE public.questions INCLUDING ALL) ON COMMIT DROP`); err != nil {
		t.Fatal(err)
	}
	body, err := migrations.Files.ReadFile("000052_ielts_practice_bank.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
			t.Fatal(err)
		}
	}
	bank, total, err := NewRepository(tx).List(ctx, ListParams{Exam: models.ExamIELTS, Limit: 200})
	if err != nil {
		t.Fatal(err)
	}
	if total != 40 || len(bank) != 40 {
		t.Fatalf("bank=%d total=%d", len(bank), total)
	}
	counts := map[string]int{}
	for _, q := range bank {
		counts[q.TypeID]++
		if !q.SupportsExam(models.ExamIELTS) || q.SupportsExam(models.ExamPTE) {
			t.Errorf("wrong eligibility: %s", q.ID)
		}
		if q.Skill == models.SkillReading {
			t.Errorf("reading content inserted: %s", q.ID)
		}
		if q.Skill == models.SkillListening {
			if q.AudioTranscript == "" {
				t.Errorf("missing script: %s", q.ID)
			}
			answer := models.AnswerSubmission{QuestionID: q.ID, SelectedOptions: q.CorrectAnswers, BlankResponses: map[string]string{}}
			for _, b := range q.Blanks {
				answer.BlankResponses[b.ID] = b.CorrectAnswer
			}
			result, ok := scoring.Grade(q, answer)
			if !ok || !result.IsCorrect || result.AccuracyPercentage != 100 {
				t.Errorf("ungradable key: %s %+v", q.ID, result)
			}
			wrong, ok := scoring.Grade(q, models.AnswerSubmission{})
			if !ok || wrong.Score != 0 {
				t.Errorf("empty answer earns marks: %s", q.ID)
			}
		}
		if q.TypeID == "ielts-listening-map" || q.TypeID == "ielts-writing-task1-figure" {
			if q.ImageURL == "" {
				t.Errorf("missing figure %s", q.ID)
			}
		}
		if q.TypeID == "ielts-speaking-part2" && (q.PrepTimeSeconds != 60 || q.TimeLimitSeconds != 120) {
			t.Errorf("cue card timing %s", q.ID)
		}
	}
	if len(counts) != 8 {
		t.Fatalf("expected eight subtasks: %v", counts)
	}
	for key, count := range counts {
		if count != 5 {
			t.Errorf("%s count=%d", key, count)
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE questions SET title='Editorial change' WHERE id='ielts-audit-001'`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
		t.Fatal(err)
	}
	q, err := NewRepository(tx).ByID(ctx, "ielts-audit-001")
	if err != nil || q.Title != "Editorial change" {
		t.Fatalf("replay overwrites content: %v", err)
	}
}
