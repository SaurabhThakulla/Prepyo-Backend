package admin

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/scoring"
	"github.com/prepyo/backend/migrations"
)

func TestListeningTypeMigrationAndAliasFilters(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	// Everything, including both migration directions, is rolled back.
	if _, err := tx.Exec(ctx, `SET LOCAL statement_timeout = '10s'`); err != nil {
		t.Fatal(err)
	}
	var before string
	const snapshot = `SELECT coalesce(jsonb_agg(to_jsonb(q) ORDER BY id), '[]'::jsonb)::text FROM questions q WHERE exam = 'PTE' AND skill = 'listening'`
	if err := tx.QueryRow(ctx, snapshot).Scan(&before); err != nil {
		t.Fatal(err)
	}
	apply := func(name string) {
		t.Helper()
		data, err := migrations.Files.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, string(data)); err != nil {
			t.Fatal(err)
		}
	}
	apply("000046_pte_listening_types.down.sql")
	var legacy string
	if err := tx.QueryRow(ctx, snapshot).Scan(&legacy); err != nil {
		t.Fatal(err)
	}
	apply("000046_pte_listening_types.up.sql")
	bank := questions.NewRepository(tx)
	for _, pair := range [][2]string{
		{"multiple-choice-multiple", "pte-listening-mcma"},
		{"fill-in-the-blanks", "pte-listening-fib"},
		{"highlight-correct-summary", "pte-highlight-correct-summary"},
		{"select-missing-word", "pte-select-missing-word"},
	} {
		var ids []string
		for _, filter := range pair {
			list, total, err := bank.List(ctx, questions.ListParams{Exam: models.ExamPTE, Skill: models.SkillListening, TypeID: filter, Limit: 100})
			if err != nil {
				t.Fatal(err)
			}
			if total < 2 {
				t.Fatalf("%s returned %d questions; expected migrated seed", filter, total)
			}
			got := []string{}
			for _, q := range list {
				got = append(got, q.ID)
			}
			if ids != nil && !reflect.DeepEqual(ids, got) {
				t.Fatalf("aliases differ: %v vs %v", ids, got)
			}
			ids = got
		}
	}
	apply("000046_pte_listening_types.down.sql")
	var after string
	if err := tx.QueryRow(ctx, snapshot).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != legacy {
		t.Fatal("up/down changed content beyond type metadata")
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if err := h.db.QueryRow(ctx, snapshot).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatal("migration verification changed the database")
	}
}

func TestListeningSummaryAndWordsAuthoringRoundTrip(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()
	for _, req := range []newAuthoredQuestion{
		{Exam: "PTE", TypeID: "summarize-spoken-text", Title: "Summary round trip", AudioTranscript: "Battery storage helps the grid.", ModelAnswer: "Battery storage helps the grid.", CorrectAnswers: []string{"battery", "grid"}},
		{Exam: "PTE", TypeID: "pte-highlight-incorrect-word", Title: "Word round trip", AudioTranscript: "The library and the museum are open.", ContextPassage: "The library and the library are open.", CorrectAnswers: []string{"w5"}},
	} {
		t.Run(req.TypeID, func(t *testing.T) {
			q, problems := req.normalise()
			if len(problems) != 0 {
				t.Fatal(problems)
			}
			id, err := h.insertAuthoredQuestion(ctx, q)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _, _ = h.db.Exec(context.Background(), `DELETE FROM questions WHERE id = $1`, id) })
			r := httptest.NewRequest("GET", "/", nil)
			route := chi.NewRouteContext()
			route.URLParams.Add("id", id)
			r = r.WithContext(context.WithValue(ctx, chi.RouteCtxKey, route))
			w := httptest.NewRecorder()
			h.authoredQuestion(w, r)
			if w.Code != 200 {
				t.Fatalf("detail: %d %s", w.Code, w.Body.String())
			}
			var detail struct {
				Question newAuthoredQuestion `json:"question"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &detail); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(detail.Question.CorrectAnswers, q.correctAnswers) {
				t.Fatalf("detail changed keys: %v", detail.Question.CorrectAnswers)
			}
			updated, problems := detail.Question.normalise()
			if len(problems) != 0 {
				t.Fatal(problems)
			}
			if err := h.saveAuthoredQuestion(ctx, id, updated); err != nil {
				t.Fatal(err)
			}
			found, err := questions.NewRepository(h.db).ByIDs(ctx, []string{id})
			if err != nil {
				t.Fatal(err)
			}
			stored := found[id]
			if stored.ContextPassage != q.contextPassage || stored.ModelAnswer != q.modelAnswer {
				t.Fatal("display or model answer lost")
			}
			answer := models.AnswerSubmission{SelectedOptions: []string{"w5"}, TextResponse: "battery grid " + strings.Repeat("energy ", 48)}
			result, ok := scoring.Grade(stored, answer)
			if !ok || !result.IsCorrect {
				t.Fatalf("round-trip grading: %+v", result)
			}
		})
	}
}
