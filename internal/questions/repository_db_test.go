package questions

import (
	"context"
	"github.com/prepyo/backend/internal/testdb"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
)

func TestListListeningTypeFilters(t *testing.T) {
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
	// A temporary table shadows the bank on this transaction's connection only.
	// LIKE copies columns/defaults, not foreign keys or existing question rows.
	if _, err := tx.Exec(ctx, `CREATE TEMP TABLE questions (LIKE public.questions INCLUDING DEFAULTS) ON COMMIT DROP`); err != nil {
		t.Fatal(err)
	}
	fixtures := []struct {
		id, exam, skill, typeID string
		published               bool
	}{
		{"a", "PTE", "listening", "summarize-spoken-text", true},
		{"b", "PTE", "listening", "multiple-choice-multiple", true},
		{"c", "PTE", "listening", "pte-listening-mcma", true},
		{"d", "PTE", "reading", "multiple-choice-multiple", true},
		{"e", "IELTS", "listening", "multiple-choice-multiple", true},
		{"f", "PTE", "listening", "pte-listening-mcma", false},
	}
	for _, f := range fixtures {
		_, err := tx.Exec(ctx, `INSERT INTO questions
			(id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt, time_limit_seconds, is_published)
			VALUES ($1, 'test-version', $2, $3, $4, $5, 'Test', 'Test', 'Test', 60, $6)`,
			f.id, f.exam, []string{f.exam}, f.skill, f.typeID, f.published)
		if err != nil {
			t.Fatal(err)
		}
	}
	bank := NewRepository(tx)
	for _, tc := range []struct {
		filter string
		want   []string
	}{
		{"summarize-spoken-text", []string{"a"}},
		{"multiple-choice-multiple", []string{"b", "c"}},
		{"pte-listening-mcma", []string{"b", "c"}},
		{"unknown", []string{}},
	} {
		list, total, err := bank.List(ctx, ListParams{Exam: models.ExamPTE, Skill: models.SkillListening, TypeID: tc.filter, Limit: 10})
		if err != nil {
			t.Fatal(err)
		}
		got := []string{}
		for _, q := range list {
			got = append(got, q.ID)
		}
		if total != len(tc.want) || !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("filter %s: total=%d ids=%v, want %v", tc.filter, total, got, tc.want)
		}
	}
	list, total, err := bank.List(ctx, ListParams{Exam: models.ExamPTE, Skill: models.SkillListening, TypeID: "pte-listening-mcma", Limit: 1, Offset: 1})
	if err != nil || total != 2 || len(list) != 1 || list[0].ID != "c" {
		t.Fatalf("pagination: total=%d list=%v err=%v", total, list, err)
	}
}
