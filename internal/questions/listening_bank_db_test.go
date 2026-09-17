package questions

import (
	"context"
	"fmt"
	"github.com/prepyo/backend/internal/testdb"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
	"github.com/prepyo/backend/migrations"
)

func TestPTEListeningBankMigration(t *testing.T) {
	url := testdb.URL(t)
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed question tests")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	// Shadow the real bank, including checks and indexes (LIKE does not copy
	// foreign keys), so tests work before/after deployment without changing data.
	if _, err := tx.Exec(ctx, `SET LOCAL statement_timeout = '10s';
        CREATE TEMP TABLE questions (LIKE public.questions INCLUDING ALL) ON COMMIT DROP;
        INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill,
            type_id, type_name, title, prompt, time_limit_seconds)
        VALUES ('bank-sentinel', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening',
            'write-from-dictation', 'Write from Dictation', 'Existing row', 'Existing prompt', 60)`); err != nil {
		t.Fatal(err)
	}
	apply := func(direction string) {
		t.Helper()
		body, err := migrations.Files.ReadFile("000047_pte_listening_bank." + direction + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			t.Fatal(err)
		}
	}
	snapshot := func() string {
		t.Helper()
		var data string
		if err := tx.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(q) ORDER BY id)::text FROM questions q`).Scan(&data); err != nil {
			t.Fatal(err)
		}
		return data
	}
	before := snapshot()
	apply("up")
	bank := NewRepository(tx)
	ids := make([]string, 40)
	for i := range ids {
		ids[i] = fmt.Sprintf("pte-lis-%03d", i+20)
	}
	found, err := bank.ByIDs(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(found) != 40 {
		t.Fatalf("got %d new questions, want 40", len(found))
	}
	counts := map[string]int{}
	transcripts := map[string]bool{}
	for _, id := range ids {
		q := found[id]
		counts[q.TypeID]++
		if transcripts[q.AudioTranscript] {
			t.Errorf("%s duplicates a transcript", id)
		}
		transcripts[q.AudioTranscript] = true
		t.Run(id, func(t *testing.T) { validateListeningSeed(t, q) })
	}
	types := []string{"write-from-dictation", "summarize-spoken-text", "pte-listening-mcma",
		"pte-listening-fib", "pte-highlight-correct-summary", "pte-listening-mcsa",
		"pte-select-missing-word", "pte-highlight-incorrect-word"}
	if len(counts) != len(types) {
		t.Fatalf("unexpected subtask counts: %v", counts)
	}
	for _, typeID := range types {
		if counts[typeID] != 5 {
			t.Errorf("%s has %d new questions, want 5", typeID, counts[typeID])
		}
		list, _, err := bank.List(ctx, ListParams{Exam: models.ExamPTE, Skill: models.SkillListening, TypeID: typeID, Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		visible := 0
		for _, q := range list {
			if _, ok := found[q.ID]; ok {
				visible++
			}
		}
		if visible != 5 {
			t.Errorf("%s exposes %d new questions, want 5", typeID, visible)
		}
	}
	seeded := snapshot()
	apply("down")
	if snapshot() != before {
		t.Fatal("rollback changed existing rows or left new rows behind")
	}
	apply("up")
	if snapshot() != seeded {
		t.Fatal("reapplying after rollback produced different data")
	}
}

// Validate stored keys independently of SQL generation, then exercise the same
// grader used for learner submissions, including an empty negative control.
func validateListeningSeed(t *testing.T, q models.Question) {
	t.Helper()
	if q.Exam != models.ExamPTE || q.Skill != models.SkillListening || !q.SupportsExam(models.ExamPTE) {
		t.Fatalf("wrong exam/skill: %+v", q)
	}
	if q.AudioTranscript == "" || q.AudioURL != "" || q.Title == "" || q.Prompt == "" || q.Explanation == "" || q.Points <= 0 || q.TimeLimitSeconds <= 0 {
		t.Fatal("missing content, timing or points, or unexpected recording")
	}
	public := q.PublicQuestion()
	if public.AudioTranscript != q.AudioTranscript || public.ContextPassage != q.ContextPassage || len(public.CorrectAnswers) != 0 || public.ModelAnswer != "" {
		t.Fatal("public question must retain TTS/display content but hide answer keys")
	}
	for _, b := range public.Blanks {
		if b.CorrectAnswer != "" {
			t.Fatal("public blank exposes answer")
		}
	}
	sub := models.AnswerSubmission{SelectedOptions: q.CorrectAnswers, BlankResponses: map[string]string{}}
	switch q.TypeID {
	case "write-from-dictation":
		if len(q.Blanks) != 0 || len(q.CorrectAnswers) != 1 || q.CorrectAnswers[0] != q.AudioTranscript {
			t.Fatal("dictation key must equal transcript and must not use blanks")
		}
		sub.TextResponse = q.AudioTranscript
	case "summarize-spoken-text":
		words := regexp.MustCompile(`[^\p{L}\p{N}\s']+`).ReplaceAllString(q.ModelAnswer, " ")
		if n := len(strings.Fields(words)); n < 50 || n > 70 {
			t.Fatalf("model summary has %d words, want 50-70", n)
		}
		if len(q.CorrectAnswers) == 0 {
			t.Fatal("summary needs keywords")
		}
		for _, key := range q.CorrectAnswers {
			if !strings.Contains(strings.ToLower(q.ModelAnswer), strings.ToLower(key)) {
				t.Errorf("model summary missing keyword %q", key)
			}
		}
		sub.TextResponse = q.ModelAnswer
	case "pte-listening-fib":
		if len(q.Blanks) != 3 {
			t.Fatal("expected three blanks")
		}
		reconstructed := q.ContextPassage
		for _, b := range q.Blanks {
			marker := "[[" + b.ID + "]]"
			if b.CorrectAnswer == "" || strings.Count(reconstructed, marker) != 1 {
				t.Fatalf("invalid blank %+v", b)
			}
			reconstructed = strings.ReplaceAll(reconstructed, marker, b.CorrectAnswer)
			sub.BlankResponses[b.ID] = b.CorrectAnswer
		}
		if reconstructed != q.AudioTranscript {
			t.Fatal("filled passage differs from spoken script")
		}
	default:
		options := map[string]bool{}
		for _, o := range q.Options {
			if o.ID == "" || o.Text == "" || options[o.ID] {
				t.Fatal("invalid or duplicate option")
			}
			options[o.ID] = true
		}
		seen := map[string]bool{}
		for _, key := range q.CorrectAnswers {
			if !options[key] || seen[key] {
				t.Fatalf("invalid or duplicate answer ID %q", key)
			}
			seen[key] = true
		}
		want := 1
		if q.TypeID == "pte-listening-mcma" {
			want = 2
		}
		if q.TypeID == "pte-highlight-incorrect-word" {
			want = 3
		}
		if len(q.CorrectAnswers) != want {
			t.Fatalf("got %d answers, want %d", len(q.CorrectAnswers), want)
		}
		if q.TypeID == "pte-highlight-incorrect-word" {
			displayed, spoken := strings.Fields(q.ContextPassage), strings.Fields(q.AudioTranscript)
			if len(displayed) != len(spoken) || len(q.Options) != len(displayed) {
				t.Fatal("word alignment differs")
			}
			var differences []string
			for i, word := range displayed {
				id := fmt.Sprintf("w%d", i+1)
				if q.Options[i].ID != id || q.Options[i].Text != word {
					t.Fatal("word option position mismatch")
				}
				if word != spoken[i] {
					differences = append(differences, id)
				}
			}
			if !reflect.DeepEqual(differences, q.CorrectAnswers) {
				t.Fatalf("differences %v do not match key %v", differences, q.CorrectAnswers)
			}
		} else if len(q.Options) != 4 {
			t.Fatal("expected four choices")
		}
		if q.TypeID == "pte-select-missing-word" {
			if !strings.HasSuffix(q.AudioTranscript, "[beep]") || strings.Count(q.AudioTranscript, "[beep]") != 1 {
				t.Fatal("missing-word script must end in exactly one beep")
			}
		}
	}
	result, ok := scoring.Grade(q, sub)
	if !ok || !result.IsCorrect || result.Score != result.MaxScore {
		t.Fatalf("reference answer did not receive full credit: %+v (graded=%v)", result, ok)
	}
	empty, ok := scoring.Grade(q, models.AnswerSubmission{})
	if !ok || empty.Score != 0 || empty.IsCorrect {
		t.Fatalf("empty answer received credit: %+v", empty)
	}
}
