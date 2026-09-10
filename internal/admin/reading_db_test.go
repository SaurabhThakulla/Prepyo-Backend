package admin

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database-backed tests for reading authoring. Validation is covered without a
// database in reading_test.go; what needs one is the writing itself, because a
// column list that does not match the table is only wrong at run time.
//
// Run with:
//   TEST_DATABASE_URL=postgres://postgres@localhost:5432/prepyo?sslmode=disable go test ./internal/admin/

func testHandler(t *testing.T) *Handler {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed admin tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	return NewHandler(pool, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))
}

// bodyOf returns a passage long enough to clear the word-count floor.
func bodyOf(words int) string {
	body := ""
	for i := 0; i < words; i++ {
		body += "word "
	}
	return body
}

func TestInsertPassageWritesGapFills(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()

	req := newPassage{
		Exam:  "PTE",
		Title: "Gap fill authoring check",
		Body:  bodyOf(minPassageWords + 10),
		Groups: []newGroup{
			{
				TypeID:       "fill-in-blanks-rw",
				Instructions: "Select the appropriate answer choice for each blank.",
				Questions: []newQuestion{{
					Prompt:         "Cacao before Europe",
					ContextPassage: "Cacao was a [[b1]] long before it became a [[b2]].",
					Blanks: []newBlank{
						{ID: "b1", Options: []string{"drink", "coin"}, CorrectAnswer: "drink"},
						{ID: "b2", Options: []string{"confection", "currency"}, CorrectAnswer: "confection"},
					},
				}},
			},
			{
				TypeID:       "fill-in-blanks-r",
				Instructions: "Drag words from the box below into the gaps.",
				WordBank:     []string{"outlawed", "trebled", "variety", "warmer"},
				Questions: []newQuestion{{
					Prompt:         "Hives in the city",
					ContextPassage: "Keeping bees in cities was quietly [[b1]].",
					Blanks:         []newBlank{{ID: "b1", CorrectAnswer: "outlawed"}},
				}},
			},
		},
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
		if _, err := h.db.Exec(ctx, `DELETE FROM reading_passages WHERE id = $1`, passageID); err != nil {
			t.Errorf("cleanup passage: %v", err)
		}
	})

	// A gap-fill group that rendered the passage would hand the learner the
	// ungapped original.
	var displays []string
	rows, err := h.db.Query(ctx,
		`SELECT passage_display FROM reading_question_groups WHERE passage_id = $1 ORDER BY position`, passageID)
	if err != nil {
		t.Fatalf("read groups: %v", err)
	}
	for rows.Next() {
		var display string
		if err := rows.Scan(&display); err != nil {
			t.Fatalf("scan group: %v", err)
		}
		displays = append(displays, display)
	}
	rows.Close()
	if len(displays) != 2 || displays[0] != "hidden" || displays[1] != "hidden" {
		t.Errorf("passage_display = %v, want both hidden", displays)
	}

	var bank []resource
	var bankJSON []byte
	if err := h.db.QueryRow(ctx, `
		SELECT resources FROM reading_question_groups
		WHERE passage_id = $1 AND type_id = 'fill-in-blanks-r'`, passageID).Scan(&bankJSON); err != nil {
		t.Fatalf("read word bank: %v", err)
	}
	if err := json.Unmarshal(bankJSON, &bank); err != nil {
		t.Fatalf("decode word bank: %v", err)
	}
	if len(bank) != 4 || bank[0].Label != "w1" || bank[0].Text != "outlawed" {
		t.Errorf("word bank = %+v, want four labelled words", bank)
	}

	var context1 string
	var blanksJSON []byte
	if err := h.db.QueryRow(ctx, `
		SELECT context_passage, blanks FROM questions
		WHERE passage_id = $1 AND type_id = 'fill-in-blanks-rw'`, passageID).Scan(&context1, &blanksJSON); err != nil {
		t.Fatalf("read gapped question: %v", err)
	}
	var blanks []storedBlank
	if err := json.Unmarshal(blanksJSON, &blanks); err != nil {
		t.Fatalf("decode blanks: %v", err)
	}
	if len(blanks) != 2 || blanks[0].ID != "b1" || blanks[0].CorrectAnswer != "drink" {
		t.Fatalf("blanks = %+v, want b1 answered drink", blanks)
	}
	if len(blanks[0].Options) != 2 {
		t.Errorf("b1 options = %v, want the two choices stored", blanks[0].Options)
	}

	// Every marker needs a key and every key a marker — the invariant the seed
	// migrations assert over the whole bank.
	var markers, keyed int
	if err := h.db.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM regexp_matches(context_passage, '\[\[b[0-9]+\]\]', 'g')),
		       jsonb_array_length(blanks)
		FROM questions WHERE passage_id = $1 AND type_id = 'fill-in-blanks-rw'`, passageID).
		Scan(&markers, &keyed); err != nil {
		t.Fatalf("count markers: %v", err)
	}
	if markers != keyed {
		t.Errorf("%d markers but %d keys", markers, keyed)
	}
}

func TestInsertReorderItemWritesQuestion(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()

	req := newReorderItem{
		Exam:  "PTE",
		Title: "Reorder authoring check",
		Boxes: []string{"Opening box.", "Second box.", "Third box.", "Closing box."},
	}

	boxes, problems := req.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected validation problems: %v", problems)
	}

	itemID, err := h.insertReorderItem(ctx, req, boxes)
	if err != nil {
		t.Fatalf("insert reorder item: %v", err)
	}
	t.Cleanup(func() {
		if _, err := h.db.Exec(ctx, `DELETE FROM reading_reorder_items WHERE id = $1`, itemID); err != nil {
			t.Errorf("cleanup item: %v", err)
		}
	})

	// The order the boxes were written in is the answer key, and points are the
	// adjacent pairs, which is what the scorer marks.
	var orderJSON []byte
	var points int
	var exams []string
	if err := h.db.QueryRow(ctx, `
		SELECT correct_answers, points, supported_exams FROM questions WHERE reorder_item_id = $1`, itemID).
		Scan(&orderJSON, &points, &exams); err != nil {
		t.Fatalf("read reorder question: %v", err)
	}

	var order []string
	if err := json.Unmarshal(orderJSON, &order); err != nil {
		t.Fatalf("decode order: %v", err)
	}
	if len(order) != 4 || order[0] != "A" || order[3] != "D" {
		t.Errorf("order = %v, want A-D", order)
	}
	if points != 3 {
		t.Errorf("points = %d, want one per adjacent pair", points)
	}
	if len(exams) != 1 || exams[0] != "PTE" {
		t.Errorf("supported_exams = %v, want just PTE", exams)
	}
}

// An IELTS paper prints a summary set inside a titled box, and the title is
// neither the task name nor the instruction line.
func TestInsertPassageKeepsTheBoxTitle(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()

	req := newPassage{
		Exam:  "IELTS",
		Title: "Box title authoring check",
		Body:  bodyOf(minPassageWords + 10),
		Groups: []newGroup{{
			TypeID:       "reading-sentence-completion",
			Instructions: "Complete the summary below. Write ONE WORD ONLY from the passage in each gap.",
			BoxTitle:     "Archaeological discoveries",
			Questions: []newQuestion{
				{Prompt: "The community's ______ is indicated by the amount of pottery.", CorrectAnswers: []string{"wealth"}},
				{Prompt: "Other finds include round ceramic ______.", CorrectAnswers: []string{"objects"}},
			},
		}},
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
		if _, err := h.db.Exec(ctx, `DELETE FROM reading_passages WHERE id = $1`, passageID); err != nil {
			t.Errorf("cleanup passage: %v", err)
		}
	})

	var title string
	var questions int
	if err := h.db.QueryRow(ctx, `
		SELECT g.box_title, count(q.id)
		FROM reading_question_groups g
		JOIN questions q ON q.group_id = g.id
		WHERE g.passage_id = $1
		GROUP BY g.box_title`, passageID).Scan(&title, &questions); err != nil {
		t.Fatalf("read group: %v", err)
	}
	if title != "Archaeological discoveries" {
		t.Errorf("box_title = %q, want the summary's own heading", title)
	}
	// Each gap is its own numbered question, which is what lets the paper mark
	// them 22, 23, 24 … and score them one by one.
	if questions != 2 {
		t.Errorf("got %d questions, want one per gap", questions)
	}
}

func TestOnePassageCarriesBothExams(t *testing.T) {
	h := testHandler(t)
	ctx := context.Background()

	req := newPassage{
		Exam:    "IELTS",
		Title:   "Shared passage authoring check",
		Body:    bodyOf(minPassageWords + 10),
		Publish: true,
		Groups: []newGroup{
			{
				TypeID:       "reading-mcq-single",
				Exams:        []string{"IELTS", "PTE"},
				Instructions: "Choose the correct letter.",
				Questions: []newQuestion{{
					Prompt:         "What does the passage mainly describe?",
					Options:        []string{"A trade route", "A chemical process"},
					CorrectAnswers: []string{"A trade route"},
				}},
			},
			{
				TypeID:       "fill-in-blanks-rw",
				Exams:        []string{"PTE"},
				Instructions: "Select the appropriate answer choice for each blank.",
				Questions: []newQuestion{{
					Prompt:         "Trade and the bean",
					ContextPassage: "The bean travelled as [[b1]] before it travelled as food.",
					Blanks:         []newBlank{{ID: "b1", Options: []string{"money", "cargo"}, CorrectAnswer: "money"}},
				}},
			},
			{
				TypeID:       "reading-arrange-passage",
				Instructions: "The boxes below are printed out of order.",
				Boxes:        []string{"The earliest use of the bean.", "The bean reaches Europe."},
				Questions: []newQuestion{
					{Prompt: "Position 1: the earliest stage.", CorrectAnswers: []string{"A"}},
					{Prompt: "Position 2: the European stage.", CorrectAnswers: []string{"B"}},
				},
			},
		},
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
		if _, err := h.db.Exec(ctx, `DELETE FROM reading_passages WHERE id = $1`, passageID); err != nil {
			t.Errorf("cleanup passage: %v", err)
		}
	})

	type row struct {
		typeID    string
		exams     []string
		versionID string
	}
	rows, err := h.db.Query(ctx, `
		SELECT DISTINCT q.type_id, q.supported_exams, q.exam_version_id
		FROM questions q WHERE q.passage_id = $1 ORDER BY q.type_id`, passageID)
	if err != nil {
		t.Fatalf("read questions: %v", err)
	}
	found := map[string]row{}
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.typeID, &r.exams, &r.versionID); err != nil {
			t.Fatalf("scan question: %v", err)
		}
		found[r.typeID] = r
	}
	rows.Close()

	shared := found["reading-mcq-single"]
	if len(shared.exams) != 2 || shared.exams[0] != "PTE" || shared.exams[1] != "IELTS" {
		t.Errorf("multiple choice supported_exams = %v, want both papers", shared.exams)
	}

	pte := found["fill-in-blanks-rw"]
	if len(pte.exams) != 1 || pte.exams[0] != "PTE" {
		t.Errorf("gap-fill supported_exams = %v, want [PTE]", pte.exams)
	}
	if pte.versionID != "pte-2026-01" {
		t.Errorf("gap-fill exam_version_id = %q, want the current PTE version", pte.versionID)
	}

	arrange := found["reading-arrange-passage"]
	if len(arrange.exams) != 1 || arrange.exams[0] != "IELTS" {
		t.Errorf("arrange supported_exams = %v, want [IELTS] inherited from the passage", arrange.exams)
	}

	for _, exam := range []string{"PTE", "IELTS"} {
		var groups int
		if err := h.db.QueryRow(ctx, `
			SELECT count(DISTINCT g.id)
			FROM reading_question_groups g
			JOIN reading_passages p ON p.id = g.passage_id AND p.is_published
			JOIN questions q ON q.group_id = g.id AND q.is_published
			WHERE g.passage_id = $1 AND $2 = ANY(q.supported_exams)`, passageID, exam).Scan(&groups); err != nil {
			t.Fatalf("count %s groups: %v", exam, err)
		}
		if groups == 0 {
			t.Errorf("%s can be dealt nothing from this passage", exam)
		}
	}

	var boxesJSON []byte
	var shuffle bool
	if err := h.db.QueryRow(ctx, `
		SELECT resources, shuffle_questions FROM reading_question_groups
		WHERE passage_id = $1 AND type_id = 'reading-arrange-passage'`, passageID).
		Scan(&boxesJSON, &shuffle); err != nil {
		t.Fatalf("read arrange group: %v", err)
	}
	var boxes []resource
	if err := json.Unmarshal(boxesJSON, &boxes); err != nil {
		t.Fatalf("decode boxes: %v", err)
	}
	if len(boxes) != 2 || boxes[0].Label != "Paragraph A" {
		t.Errorf("boxes = %+v, want the ordering material labelled A and B", boxes)
	}
	if shuffle {
		t.Error("an ordering set was marked shuffleable, which scrambles the positions")
	}
}
