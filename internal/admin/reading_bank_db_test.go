package admin

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// The authoring lifecycle, end to end: a passage is written first as a draft,
// task sets are added to it afterwards, and only publishing makes it visible to
// a learner.
//
// Run with:
//   TEST_DATABASE_URL=postgres://postgres@localhost:5432/prepyo go test ./internal/admin/

func TestPassageLifecycle(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()

	// 1. A passage with no questions at all is a legitimate draft.
	req := newPassage{
		Exam:  "IELTS",
		Title: "Lifecycle check",
		Body:  bodyOf(minPassageWords + 10),
	}
	paragraphs, problems := req.normalise()
	if len(problems) > 0 {
		t.Fatalf("an empty draft should be allowed: %v", problems)
	}

	passageID, err := h.insertPassage(ctx, req, paragraphs)
	if err != nil {
		t.Fatalf("insert passage: %v", err)
	}
	t.Cleanup(func() {
		_, _ = h.db.Exec(ctx, `DELETE FROM reading_passages WHERE id = $1`, passageID)
	})

	var published bool
	var groups int
	if err := h.db.QueryRow(ctx, `
		SELECT p.is_published, (SELECT count(*) FROM reading_question_groups WHERE passage_id = p.id)
		FROM reading_passages p WHERE p.id = $1`, passageID).Scan(&published, &groups); err != nil {
		t.Fatalf("read draft: %v", err)
	}
	if published {
		t.Error("a new passage should start as a draft")
	}
	if groups != 0 {
		t.Errorf("got %d sets on a fresh draft, want none", groups)
	}

	// 2. Sets are added afterwards, and their numbering continues.
	first, err := h.insertGroups(ctx, passageID, []newGroup{{
		TypeID:       "reading-true-false",
		Instructions: "Do the following statements agree with the information given in the passage?",
		Questions: []newQuestion{
			{Prompt: "Dreaming helps memory.", CorrectAnswers: []string{"True"}},
			{Prompt: "Dreams are always recalled.", CorrectAnswers: []string{"Not Given"}},
		},
	}})
	if err != nil {
		t.Fatalf("add first set: %v", err)
	}
	if first != 2 {
		t.Errorf("wrote %d questions, want 2", first)
	}

	// A summary completed from a list: the list belongs to the set, and every
	// gap is answered from it. Dropping it used to be silent, which left the
	// learner typing an exact phrase instead of choosing a letter.
	if _, err := h.insertGroups(ctx, passageID, []newGroup{{
		TypeID:       "reading-sentence-completion",
		Instructions: "Complete the summary using the list of words below.",
		BoxTitle:     "Archaeological discoveries",
		WordBank:     []string{"wealth", "poverty", "isolation"},
		Questions:    []newQuestion{{Prompt: "The community's ______ is clear.", CorrectAnswers: []string{"wealth"}}},
	}}); err != nil {
		t.Fatalf("add second set: %v", err)
	}

	var listEntries, gapOptions int
	if err := h.db.QueryRow(ctx, `
		SELECT jsonb_array_length(g.resources), jsonb_array_length(q.options)
		FROM reading_question_groups g
		JOIN questions q ON q.group_id = g.id
		WHERE g.passage_id = $1 AND g.type_id = 'reading-sentence-completion'`, passageID).
		Scan(&listEntries, &gapOptions); err != nil {
		t.Fatalf("read the list: %v", err)
	}
	if listEntries != 3 {
		t.Errorf("the set kept %d list entries, want 3", listEntries)
	}
	if gapOptions != 3 {
		t.Errorf("the gap offers %d choices, want the whole list", gapOptions)
	}

	var positions []int
	rows, err := h.db.Query(ctx,
		`SELECT position FROM reading_question_groups WHERE passage_id = $1 ORDER BY position`, passageID)
	if err != nil {
		t.Fatalf("read positions: %v", err)
	}
	for rows.Next() {
		var position int
		if err := rows.Scan(&position); err != nil {
			t.Fatalf("scan position: %v", err)
		}
		positions = append(positions, position)
	}
	rows.Close()
	// A second batch restarting at 1 would collide with the first: the table
	// holds (passage_id, position) unique.
	if len(positions) != 2 || positions[0] != 1 || positions[1] != 2 {
		t.Errorf("positions = %v, want 1 then 2", positions)
	}

	// 3. Questions inherit the passage's draft state, so nothing leaks early.
	var unpublishedQuestions int
	if err := h.db.QueryRow(ctx, `
		SELECT count(*) FROM questions q
		JOIN reading_question_groups g ON g.id = q.group_id
		WHERE g.passage_id = $1 AND NOT q.is_published`, passageID).Scan(&unpublishedQuestions); err != nil {
		t.Fatalf("count unpublished: %v", err)
	}
	if unpublishedQuestions != 3 {
		t.Errorf("%d of 3 questions are drafts, want all of them", unpublishedQuestions)
	}

	// 4. Addressing sets to a passage that is not there is a 404, not a crash.
	if _, err := h.insertGroups(ctx, "passage-does-not-exist", []newGroup{{
		TypeID:    "reading-true-false",
		Questions: []newQuestion{{Prompt: "x", CorrectAnswers: []string{"True"}}},
	}}); err != errNoPassage {
		t.Errorf("adding to a missing passage returned %v, want errNoPassage", err)
	}
}

func TestPassageDetailNamesTheExamsEachSetServes(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()

	req := newPassage{
		Exam:  "IELTS",
		Title: "Detail exams check",
		Body:  bodyOf(minPassageWords + 10),
	}
	paragraphs, problems := req.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected validation problems: %v", problems)
	}

	passageID, err := h.insertPassage(ctx, req, paragraphs)
	if err != nil {
		t.Fatalf("insert passage: %v", err)
	}
	t.Cleanup(func() {
		_, _ = h.db.Exec(ctx, `DELETE FROM reading_passages WHERE id = $1`, passageID)
	})

	if _, err := h.insertGroups(ctx, passageID, []newGroup{
		{
			TypeID:       "reading-mcq-single",
			Exams:        []string{"PTE", "IELTS"},
			Instructions: "Choose the correct letter.",
			Questions: []newQuestion{{
				Prompt:         "What is the passage about?",
				Options:        []string{"Trade", "Weather"},
				CorrectAnswers: []string{"Trade"},
			}},
		},
		{
			TypeID:       "reading-yes-no-not-given",
			Exams:        []string{"IELTS"},
			Instructions: "Do the statements agree with the views of the writer?",
			Questions:    []newQuestion{{Prompt: "The writer approves.", CorrectAnswers: []string{"Yes"}}},
		},
	}); err != nil {
		t.Fatalf("add sets: %v", err)
	}

	r := httptest.NewRequest("GET", "/admin/reading/passages/"+passageID, nil)
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("id", passageID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, routeCtx))

	w := httptest.NewRecorder()
	h.passage(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	var body struct {
		Passage struct {
			Exams []string `json:"exams"`
			Sets  []struct {
				TypeID string   `json:"typeId"`
				Exams  []string `json:"exams"`
			} `json:"sets"`
		} `json:"passage"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode detail: %v", err)
	}

	if len(body.Passage.Exams) != 2 {
		t.Errorf("passage exams = %v, want both papers", body.Passage.Exams)
	}
	byType := map[string][]string{}
	for _, set := range body.Passage.Sets {
		byType[set.TypeID] = set.Exams
	}
	if got := byType["reading-mcq-single"]; len(got) != 2 {
		t.Errorf("multiple choice set exams = %v, want both papers", got)
	}
	if got := byType["reading-yes-no-not-given"]; len(got) != 1 || got[0] != "IELTS" {
		t.Errorf("yes/no set exams = %v, want [IELTS]", got)
	}
}
