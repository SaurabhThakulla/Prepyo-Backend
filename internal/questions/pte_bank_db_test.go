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

func TestPTEExpansionBank(t *testing.T) {
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
	for _, table := range []string{"reading_passages", "reading_question_groups", "reading_reorder_items", "questions"} {
		if _, err = tx.Exec(ctx, "CREATE TEMP TABLE "+table+" (LIKE public."+table+" INCLUDING ALL) ON COMMIT DROP"); err != nil {
			t.Fatal(err)
		}
	}
	body, err := migrations.Files.ReadFile("000053_pte_practice_expansion.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRepository(tx)
	bank, total, err := repo.List(ctx, ListParams{Exam: models.ExamPTE, Limit: 200, IncludePassageQuestions: true})
	if err != nil {
		t.Fatal(err)
	}
	if total != 115 || len(bank) != 115 {
		t.Fatalf("total=%d len=%d", total, len(bank))
	}
	counts := map[string]int{}
	skills := map[models.SkillType]int{}
	for _, q := range bank {
		counts[q.TypeID]++
		skills[q.Skill]++
		if !q.SupportsExam(models.ExamPTE) || q.SupportsExam(models.ExamIELTS) {
			t.Errorf("eligibility: %s", q.ID)
		}
		if q.Explanation == "" {
			t.Errorf("missing explanation: %s", q.ID)
		}
		for _, b := range q.Blanks {
			if !strings.Contains(q.ContextPassage, "[["+b.ID+"]]") {
				t.Errorf("missing blank: %s/%s", q.ID, b.ID)
			}
		}
		if q.Skill == models.SkillListening && q.AudioTranscript == "" {
			t.Errorf("missing audio script: %s", q.ID)
		}
		if q.TypeID == "pte-describe-image" && (!strings.HasPrefix(q.ImageURL, "data:image/svg+xml,") || q.FigureData == "") {
			t.Errorf("missing figure: %s", q.ID)
		}
		if q.TypeID == "summarize-spoken-text" {
			n := len(strings.Fields(q.ModelAnswer))
			if n < 50 || n > 70 {
				t.Errorf("summary length: %s %d", q.ID, n)
			}
		}
		if q.Skill != models.SkillReading && q.Skill != models.SkillListening {
			continue
		}
		sub := models.AnswerSubmission{QuestionID: q.ID, SelectedOptions: q.CorrectAnswers, BlankResponses: map[string]string{}}
		for _, b := range q.Blanks {
			sub.BlankResponses[b.ID] = b.CorrectAnswer
		}
		if q.TypeID == "write-from-dictation" {
			sub.TextResponse = q.AudioTranscript
		}
		if q.TypeID == "summarize-spoken-text" {
			// Heuristic scores are not proof of linguistic validity; test the model separately from exact keys.
			sub.TextResponse = q.ModelAnswer
		}
		result, ok := scoring.Grade(q, sub)
		if !ok {
			t.Errorf("ungradable: %s", q.ID)
		}
		if q.TypeID != "summarize-spoken-text" && result.AccuracyPercentage != 100 {
			t.Errorf("key rejected: %s %+v", q.ID, result)
		}
		empty, ok := scoring.Grade(q, models.AnswerSubmission{})
		if !ok || empty.Score != 0 {
			t.Errorf("empty response earns marks: %s %+v", q.ID, empty)
		}
	}
	if len(counts) != 23 {
		t.Fatalf("type counts: %v", counts)
	}
	for id, n := range counts {
		if n != 5 {
			t.Errorf("%s: %d", id, n)
		}
	}
	for skill, want := range map[models.SkillType]int{models.SkillReading: 25, models.SkillListening: 40, models.SkillWriting: 15, models.SkillSpeaking: 35} {
		if skills[skill] != want {
			t.Errorf("%s=%d want %d", skill, skills[skill], want)
		}
	}
	var linked, ordered int
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM questions q JOIN reading_question_groups g ON g.id=q.group_id AND g.passage_id=q.passage_id JOIN reading_passages p ON p.id=q.passage_id WHERE q.skill='reading' AND p.is_published AND g.type_id=q.type_id`).Scan(&linked); err != nil {
		t.Fatal(err)
	}
	if err = tx.QueryRow(ctx, `SELECT count(*) FROM questions q JOIN reading_reorder_items i ON i.id=q.reorder_item_id WHERE q.type_id='reorder-paragraphs' AND i.is_published`).Scan(&ordered); err != nil {
		t.Fatal(err)
	}
	if linked != 20 || ordered != 5 {
		t.Fatalf("linked=%d ordered=%d", linked, ordered)
	}
	if _, err = tx.Exec(ctx, `UPDATE questions SET title='Editorial revision' WHERE id='pte-expansion-001'`); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
		t.Fatal(err)
	}
	q, err := repo.ByID(ctx, "pte-expansion-001")
	if err != nil || q.Title != "Editorial revision" {
		t.Fatalf("replay overwrites edits: %v", err)
	}
}
