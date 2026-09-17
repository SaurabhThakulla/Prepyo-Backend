package questions

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/migrations"
)

func TestWritingElectricityGuidanceMigration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
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
	if _, err = tx.Exec(ctx, `SET LOCAL statement_timeout = '10s';
 CREATE TEMP TABLE questions (LIKE public.questions INCLUDING ALL) ON COMMIT DROP`); err != nil {
		t.Fatal(err)
	}
	// Load the actual historical writing seeds, supplying eligibility introduced later.
	seed, err := migrations.Files.ReadFile("000013_writing_content.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.ReplaceAll(string(seed), "ALTER TABLE questions ADD COLUMN IF NOT EXISTS figure_data TEXT;", "")
	sql = strings.ReplaceAll(sql, "exam, skill,", "exam, supported_exams, skill,")
	sql = strings.ReplaceAll(sql, "'IELTS', 'writing'", "'IELTS', ARRAY['IELTS'], 'writing'")
	sql = strings.ReplaceAll(sql, "'PTE', 'writing'", "'PTE', ARRAY['PTE'], 'writing'")
	if _, err = tx.Exec(ctx, sql); err != nil {
		t.Fatal(err)
	}
	snapshot := func() string {
		t.Helper()
		var data string
		if err := tx.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(q) ORDER BY id)::text FROM questions q`).Scan(&data); err != nil {
			t.Fatal(err)
		}
		return data
	}
	apply := func(direction string) {
		t.Helper()
		body, err := migrations.Files.ReadFile("000048_writing_electricity_guidance." + direction + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(ctx, string(body)); err != nil {
			t.Fatal(err)
		}
	}
	before := snapshot()
	apply("up")
	repo := NewRepository(tx)
	list, total, err := repo.List(ctx, ListParams{Exam: models.ExamIELTS, Skill: models.SkillWriting, TypeID: "ielts-writing-task1-figure", Limit: 100})
	if err != nil || total != 10 {
		t.Fatalf("list: total=%d err=%v", total, err)
	}
	found := false
	for _, q := range list {
		if q.ID != "ielts-wrt-fg-003" {
			continue
		}
		found = true
		if !strings.Contains(q.FigureData, "become the largest source") || strings.Contains(q.FigureData, "second largest") {
			t.Fatal(q.FigureData)
		}
		if q.PublicQuestion().FigureData != "" {
			t.Fatal("evaluator guidance leaked to learner")
		}
	}
	if !found {
		t.Fatal("corrected question missing from exam/type listing")
	}
	after := snapshot()
	apply("up")
	if snapshot() != after {
		t.Fatal("up is not idempotent")
	}
	apply("down")
	if snapshot() != before {
		t.Fatal("down did not restore seed")
	}
	apply("up")
	if snapshot() != after {
		t.Fatal("up/down/up differs")
	}
}

// The remaining writing repairs: guidance accuracy, one model answer, the
// legacy Task 2 type, and four new PTE summaries. Uses the actual 000013 seed
// plus fixtures carrying the affected content from 000002.
func TestWritingContentRepairsMigration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
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
	if _, err = tx.Exec(ctx, `SET LOCAL statement_timeout = '10s';
 CREATE TEMP TABLE questions (LIKE public.questions INCLUDING ALL) ON COMMIT DROP`); err != nil {
		t.Fatal(err)
	}
	// The 000013 seed predates the eligibility columns, so run it with
	// supported_exams filled in. Its multi-statement text goes over the simple
	// protocol via PgConn, which runs it as one implicit batch.
	seed, err := migrations.Files.ReadFile("000013_writing_content.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	patched := strings.ReplaceAll(string(seed), "ALTER TABLE questions ADD COLUMN IF NOT EXISTS figure_data TEXT;", "")
	patched = strings.ReplaceAll(patched, "exam, skill,", "exam, supported_exams, skill,")
	patched = strings.ReplaceAll(patched, "'IELTS', 'writing'", "'IELTS', ARRAY['IELTS'], 'writing'")
	patched = strings.ReplaceAll(patched, "'PTE', 'writing'", "'PTE', ARRAY['PTE'], 'writing'")
	if _, err = tx.Conn().PgConn().Exec(ctx, patched).ReadAll(); err != nil {
		t.Fatal(err)
	}
	// The two 000002 rows that 000049 also touches, with their seed content.
	for _, f := range []struct {
		id, exam, typeID, typeName, model, prompt, passage string
	}{
		{"ielts-wrt-001", "IELTS", "ielts-writing-task2", "Writing Task 2", "",
			"In many countries employees increasingly work from home. Do the advantages outweigh the disadvantages? Give reasons and include relevant examples. Write at least 250 words.",
			""},
		{"pte-wrt-001", "PTE", "summarize-written-text", "Summarize Written Text",
			"Because dense urban surfaces retain heat and raise both energy costs and health risks, planners are adopting reflective materials, green roofs and wider tree cover to reduce urban heat island effects.",
			"Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.",
			"Urban heat islands occur when cities replace natural land cover with dense concentrations of pavement and buildings that retain heat, increasing cooling energy costs and elevating heat-related illnesses. Municipal planners increasingly respond with reflective roofing materials, expanded tree canopies and permeable surfaces."},
	} {
		if _, err = tx.Exec(ctx, `INSERT INTO questions
			(id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
			 context_passage, prep_time_seconds, time_limit_seconds, correct_answers, blanks,
			 model_answer, explanation, difficulty, tags, points)
			VALUES ($1, 'test-version', $2, $3, 'writing', $4, $5, 'Seed title', $6,
			 NULLIF($7, ''), 0, 600, NULL, NULL, NULLIF($8, ''), 'Seed explanation', 'medium',
			 ARRAY['Seed'], 10)`,
			f.id, f.exam, []string{f.exam}, f.typeID, f.typeName, f.prompt, f.passage, f.model); err != nil {
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
	apply := func(direction string) {
		t.Helper()
		body, err := migrations.Files.ReadFile("000049_writing_content_repairs." + direction + ".sql")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(ctx, string(body)); err != nil {
			t.Fatal(err)
		}
	}
	before := snapshot()
	apply("up")
	row := func(id string) models.Question {
		t.Helper()
		q, err := NewRepository(tx).ByID(ctx, id)
		if err != nil {
			t.Fatalf("read %s: %v", id, err)
		}
		return q
	}
	if got := row("ielts-wrt-fg-001").FigureData; !strings.Contains(got, "Mexico has the largest absolute increase") || strings.Contains(got, "grows the least") {
		t.Errorf("fg-001 guidance not corrected: %s", got)
	}
	if got := row("ielts-wrt-fg-004").FigureData; !strings.Contains(got, "under a third of its 2015 figure") {
		t.Errorf("fg-004 guidance not corrected: %s", got)
	}
	if got := row("ielts-wrt-fg-009").FigureData; !strings.Contains(got, "no income data") || strings.Contains(got, "rises as national income falls") {
		t.Errorf("fg-009 guidance not corrected: %s", got)
	}
	if got := row("pte-wrt-001").ModelAnswer; !strings.Contains(got, "permeable surfaces") || strings.Contains(got, "green roofs") {
		t.Errorf("pte-wrt-001 model answer not corrected: %s", got)
	}
	if got := row("ielts-wrt-001"); got.TypeID != "ielts-writing-task2-opinion" || got.TypeName != "Opinion / Agree or Disagree" {
		t.Errorf("legacy task2 type not repaired: %s / %s", got.TypeID, got.TypeName)
	}
	// New summaries: stored fully, visible to the practice listing, one-sentence keys.
	list, total, err := NewRepository(tx).List(ctx, ListParams{Exam: models.ExamPTE, Skill: models.SkillWriting, TypeID: "summarize-written-text", Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	if total != 5 {
		t.Fatalf("PTE summary bank has %d questions, want 5", total)
	}
	seen := map[string]bool{}
	for _, q := range list {
		seen[q.ID] = true
		if q.ContextPassage == "" || q.ModelAnswer == "" || q.Points <= 0 || !q.SupportsExam(models.ExamPTE) {
			t.Errorf("%s missing content or eligibility: %+v", q.ID, q)
		}
		if words := len(strings.Fields(q.ModelAnswer)); words < 5 || words > 75 {
			t.Errorf("%s model answer has %d words, want 5-75", q.ID, words)
		}
		if q.ID != "pte-wrt-001" && (len(strings.Fields(q.ContextPassage)) < 100 || q.Explanation == "" || q.TimeLimitSeconds != 600) {
			t.Errorf("%s missing substantive passage, guidance or correct timing", q.ID)
		}
		if strings.Count(q.ModelAnswer, ".") != 1 || !strings.HasSuffix(q.ModelAnswer, ".") {
			t.Errorf("%s model answer is not one sentence: %s", q.ID, q.ModelAnswer)
		}
		public := q.PublicQuestion()
		if public.ContextPassage != q.ContextPassage || public.ModelAnswer != "" {
			t.Errorf("%s public question leaks answer key", q.ID)
		}
	}
	if !seen["pte-wrt-swt-002"] || !seen["pte-wrt-swt-003"] || !seen["pte-wrt-swt-004"] || !seen["pte-wrt-swt-005"] || !seen["pte-wrt-001"] {
		t.Fatalf("summary listing missing rows: %v", seen)
	}
	seeded := snapshot()
	apply("down")
	if snapshot() != before {
		t.Fatal("down did not restore the seeded rows")
	}
	apply("up")
	if snapshot() != seeded {
		t.Fatal("reapplying after rollback produced different data")
	}
}
