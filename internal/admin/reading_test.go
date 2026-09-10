package admin

import (
	"strings"
	"testing"
)

func TestParagraphLabel(t *testing.T) {
	cases := map[int]string{0: "A", 1: "B", 25: "Z", 26: "AA", 27: "AB"}
	for index, want := range cases {
		if got := paragraphLabel(index); got != want {
			t.Errorf("paragraphLabel(%d) = %q, want %q", index, got, want)
		}
	}
}

// Blank lines are the only thing that splits a pasted passage, so a single
// newline inside a paragraph must not start a new one.
func TestNormaliseSplitsOnBlankLines(t *testing.T) {
	req := newPassage{
		Exam:   "IELTS",
		Title:  "The Long Road of Papermaking",
		Body:   "First block\nstill first block\n\n  Second block  \n\n\n\nThird block",
		Groups: []newGroup{{TypeID: "reading-mcq-single"}},
	}

	paragraphs, problems := req.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if len(paragraphs) != 3 {
		t.Fatalf("got %d paragraphs, want 3: %+v", len(paragraphs), paragraphs)
	}
	if paragraphs[0].Text != "First block\nstill first block" {
		t.Errorf("paragraph A = %q, want the single newline kept", paragraphs[0].Text)
	}
	if paragraphs[1].Text != "Second block" {
		t.Errorf("paragraph B = %q, want it trimmed", paragraphs[1].Text)
	}
	if paragraphs[2].Label != "C" {
		t.Errorf("third label = %q, want C", paragraphs[2].Label)
	}
}

func TestNormaliseRejectsEmptyPassage(t *testing.T) {
	_, problems := newPassage{Exam: "Japanese", Difficulty: "brutal"}.normalise()

	for _, field := range []string{"exam", "title", "body", "difficulty"} {
		if _, ok := problems[field]; !ok {
			t.Errorf("expected a problem for %q, got %v", field, problems)
		}
	}

	// A passage with no task sets is a draft in progress, not a mistake: the
	// admin screen writes the text first and adds sets to it afterwards.
	if _, ok := problems["groups"]; ok {
		t.Error("a passage with no sets should be allowed to exist as a draft")
	}
}

func TestNormaliseAcceptsPreSplitParagraphs(t *testing.T) {
	req := newPassage{
		Exam:       "PTE",
		Title:      "T",
		Paragraphs: []string{"one", "two"},
		Body:       "ignored because paragraphs were given",
		Groups:     []newGroup{{TypeID: "reading-mcq-single"}},
	}

	paragraphs, problems := req.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if len(paragraphs) != 2 || paragraphs[0].Text != "one" {
		t.Errorf("got %+v, want the supplied paragraphs used as-is", paragraphs)
	}
}

// The rule worth protecting: an answer key that names something nobody can pick
// marks every attempt wrong forever, and only the learner would find out.
func TestResolveAnswersRejectsUngradableKeys(t *testing.T) {
	labels := map[string]bool{"A": true, "B": true}

	cases := []struct {
		name     string
		typeID   string
		question newQuestion
		wantPart string
	}{
		{
			name:     "answer is not one of the options",
			typeID:   "reading-mcq-single",
			question: newQuestion{Prompt: "Q", Options: []string{"one", "two"}, CorrectAnswers: []string{"three"}},
			wantPart: "not one of the options",
		},
		{
			name:     "paragraph that does not exist",
			typeID:   "reading-find-the-paragraph",
			question: newQuestion{Prompt: "Q", CorrectAnswers: []string{"Z"}},
			wantPart: "not in this passage",
		},
		{
			name:     "verdict outside the allowed set",
			typeID:   "reading-true-false",
			question: newQuestion{Prompt: "Q", CorrectAnswers: []string{"Maybe"}},
			wantPart: "answers this task allows",
		},
		{
			name:     "single-answer task given two",
			typeID:   "reading-mcq-single",
			question: newQuestion{Prompt: "Q", Options: []string{"one", "two"}, CorrectAnswers: []string{"one", "two"}},
			wantPart: "exactly one correct answer",
		},
		{
			name:     "no answer at all",
			typeID:   "reading-mcq-single",
			question: newQuestion{Prompt: "Q", Options: []string{"one", "two"}},
			wantPart: "Mark the correct answer",
		},
		{
			name:     "no prompt",
			typeID:   "reading-mcq-single",
			question: newQuestion{Options: []string{"one", "two"}, CorrectAnswers: []string{"one"}},
			wantPart: "Write the question",
		},
		{
			name:     "choice task with one option",
			typeID:   "reading-mcq-single",
			question: newQuestion{Prompt: "Q", Options: []string{"only"}, CorrectAnswers: []string{"only"}},
			wantPart: "at least two options",
		},
		{
			name:     "duplicate options",
			typeID:   "reading-mcq-single",
			question: newQuestion{Prompt: "Q", Options: []string{"one", "one"}, CorrectAnswers: []string{"one"}},
			wantPart: "listed twice",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, problem := resolveAnswers(supportedTypes[tc.typeID], tc.question, labels)
			if !strings.Contains(problem, tc.wantPart) {
				t.Errorf("problem = %q, want it to mention %q", problem, tc.wantPart)
			}
		})
	}
}

func TestResolveAnswersBuildsGradableContent(t *testing.T) {
	labels := map[string]bool{"A": true, "B": true, "C": true}

	t.Run("multiple choice keeps the author's wording as the id", func(t *testing.T) {
		options, answers, problem := resolveAnswers(
			supportedTypes["reading-mcq-single"],
			newQuestion{Prompt: "Q", Options: []string{"Baghdad", "Cairo"}, CorrectAnswers: []string{"Cairo"}},
			labels,
		)
		if problem != "" {
			t.Fatalf("unexpected problem: %s", problem)
		}
		if len(options) != 2 || options[0].ID != "Baghdad" || options[0].Text != "Baghdad" {
			t.Errorf("options = %+v, want id and text to match the wording", options)
		}
		if len(answers) != 1 || answers[0] != "Cairo" {
			t.Errorf("answers = %v, want [Cairo]", answers)
		}
	})

	t.Run("verdict tasks supply their own options", func(t *testing.T) {
		options, answers, problem := resolveAnswers(
			supportedTypes["reading-true-false"],
			newQuestion{Prompt: "Q", CorrectAnswers: []string{"not given"}},
			labels,
		)
		if problem != "" {
			t.Fatalf("unexpected problem: %s", problem)
		}
		if len(options) != 3 {
			t.Errorf("options = %+v, want the three fixed ones", options)
		}
		// Matched case-insensitively and stored as the id the grader compares.
		if len(answers) != 1 || answers[0] != "NOT_GIVEN" {
			t.Errorf("answers = %v, want [NOT_GIVEN]", answers)
		}
	})

	t.Run("paragraph tasks offer every label in order", func(t *testing.T) {
		options, answers, problem := resolveAnswers(
			supportedTypes["reading-find-the-paragraph"],
			newQuestion{Prompt: "Q", CorrectAnswers: []string{"b"}},
			labels,
		)
		if problem != "" {
			t.Fatalf("unexpected problem: %s", problem)
		}
		if len(options) != 3 || options[0].ID != "A" || options[2].ID != "C" {
			t.Errorf("options = %+v, want A, B, C in order", options)
		}
		if len(answers) != 1 || answers[0] != "B" {
			t.Errorf("answers = %v, want [B] upper-cased", answers)
		}
	})

	t.Run("typed answers accept several spellings and need no options", func(t *testing.T) {
		options, answers, problem := resolveAnswers(
			supportedTypes["reading-sentence-completion"],
			newQuestion{Prompt: "The taste is ______.", CorrectAnswers: []string{"flavour", "flavor"}},
			labels,
		)
		if problem != "" {
			t.Fatalf("unexpected problem: %s", problem)
		}
		if len(options) != 0 {
			t.Errorf("options = %+v, want none for a typed answer", options)
		}
		if len(answers) != 2 {
			t.Errorf("answers = %v, want both spellings kept", answers)
		}
	})
}

func TestNewContentIDIsReadableAndUnique(t *testing.T) {
	first := newContentID("passage", "The Long Road of Papermaking")
	second := newContentID("passage", "The Long Road of Papermaking")

	if !strings.HasPrefix(first, "passage-the-long-road-of-papermaking-") {
		t.Errorf("id = %q, want a readable slug", first)
	}
	if first == second {
		t.Error("two ids from the same title collided")
	}
	if got := newContentID("passage", "!!! ???"); !strings.HasPrefix(got, "passage-passage-") {
		t.Errorf("id = %q, want a usable fallback when the title has no letters", got)
	}
}

// The grader reads blanks in the order the markers appear, so the key has to be
// rebuilt from the text rather than trusted in the order it arrived.
func TestResolveBlanksOrdersByMarker(t *testing.T) {
	spec := supportedTypes["fill-in-blanks-rw"]
	q := newQuestion{
		Prompt:         "Select the answer for each blank.",
		ContextPassage: "Cacao was [[b1]] before it was a confection, and the drink was [[b2]] bitter.",
		Blanks: []newBlank{
			{ID: "b2", Options: []string{"deliberately", "accidentally"}, CorrectAnswer: "deliberately"},
			{ID: "b1", Options: []string{"a drink", "a coin"}, CorrectAnswer: "a drink"},
		},
	}

	text, blanks, problem := resolveBlanks(spec, q, nil)
	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if text != q.ContextPassage {
		t.Errorf("text = %q, want it kept as written", text)
	}
	if len(blanks) != 2 || blanks[0].ID != "b1" || blanks[1].ID != "b2" {
		t.Fatalf("blanks = %+v, want b1 then b2", blanks)
	}
	if blanks[0].CorrectAnswer != "a drink" {
		t.Errorf("b1 answer = %q, want the one keyed to b1", blanks[0].CorrectAnswer)
	}
}

func TestResolveBlanksRejectsMismatchedKeys(t *testing.T) {
	rw := supportedTypes["fill-in-blanks-rw"]
	cases := map[string]newQuestion{
		"no prompt":       {ContextPassage: "a [[b1]] b", Blanks: []newBlank{{ID: "b1", Options: []string{"x", "y"}, CorrectAnswer: "x"}}},
		"no text":         {Prompt: "p"},
		"no markers":      {Prompt: "p", ContextPassage: "nothing to fill"},
		"repeated marker": {Prompt: "p", ContextPassage: "[[b1]] and [[b1]]", Blanks: []newBlank{{ID: "b1", Options: []string{"x", "y"}, CorrectAnswer: "x"}}},
		"unkeyed gap":     {Prompt: "p", ContextPassage: "[[b1]] and [[b2]]", Blanks: []newBlank{{ID: "b1", Options: []string{"x", "y"}, CorrectAnswer: "x"}}},
		"orphan key":      {Prompt: "p", ContextPassage: "[[b1]]", Blanks: []newBlank{{ID: "b1", Options: []string{"x", "y"}, CorrectAnswer: "x"}, {ID: "b9", CorrectAnswer: "z"}}},
		"one choice":      {Prompt: "p", ContextPassage: "[[b1]]", Blanks: []newBlank{{ID: "b1", Options: []string{"x"}, CorrectAnswer: "x"}}},
		"answer unlisted": {Prompt: "p", ContextPassage: "[[b1]]", Blanks: []newBlank{{ID: "b1", Options: []string{"x", "y"}, CorrectAnswer: "z"}}},
	}

	for name, q := range cases {
		if _, _, problem := resolveBlanks(rw, q, nil); problem == "" {
			t.Errorf("%s: expected a problem, got none", name)
		}
	}
}

// A Reading gap-fill answers from the group's word list, so an answer that is
// not in it can never be dragged into place.
func TestResolveBlanksChecksWordBank(t *testing.T) {
	spec := supportedTypes["fill-in-blanks-r"]
	bank, lookup, problem := wordBankOf([]string{"outlawed", "trebled", " variety ", "warmer"}, false)
	if problem != "" {
		t.Fatalf("unexpected word bank problem: %s", problem)
	}
	if len(bank) != 4 || bank[0].Label != "w1" || bank[2].Text != "variety" {
		t.Fatalf("bank = %+v, want w1..w4 trimmed", bank)
	}

	q := newQuestion{Prompt: "Drag a word into each gap.", ContextPassage: "keeping bees was [[b1]]"}
	q.Blanks = []newBlank{{ID: "b1", CorrectAnswer: "outlawed"}}
	if _, blanks, problem := resolveBlanks(spec, q, lookup); problem != "" {
		t.Errorf("unexpected problem: %s", problem)
	} else if len(blanks[0].Options) != 0 {
		t.Errorf("word-bank blanks should carry no options of their own: %+v", blanks[0])
	}

	q.Blanks = []newBlank{{ID: "b1", CorrectAnswer: "banned"}}
	if _, _, problem := resolveBlanks(spec, q, lookup); problem == "" {
		t.Error("expected an answer outside the word list to be refused")
	}
}

func TestWordBankRejectsDuplicates(t *testing.T) {
	if _, _, problem := wordBankOf([]string{"warmer", "Warmer"}, false); problem == "" {
		t.Error("expected a duplicate word to be refused")
	}
}

func TestReorderItemNormalise(t *testing.T) {
	boxes, problems := newReorderItem{
		Exam:  "PTE",
		Title: "Urban beekeeping",
		Boxes: []string{"First box", "  ", "Second box", "Third box"},
	}.normalise()
	if len(problems) > 0 {
		t.Fatalf("unexpected problems: %v", problems)
	}
	if len(boxes) != 3 || boxes[2].Label != "C" {
		t.Fatalf("boxes = %+v, want three labelled A-C with the blank dropped", boxes)
	}

	if _, problems := (newReorderItem{Exam: "PTE", Title: "T", Boxes: []string{"one", "two"}}).normalise(); problems["boxes"] == "" {
		t.Error("expected fewer than three boxes to be refused")
	}
}

// A summary completed from a list is chosen from by letter, so the list is
// lettered A, B, C … rather than numbered like a heap of draggable words.
func TestWordBankLetters(t *testing.T) {
	lettered, _, problem := wordBankOf([]string{"pitch boundary", "numerous disputes", "team tactics"}, true)
	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if lettered[0].Label != "A" || lettered[2].Label != "C" {
		t.Errorf("labels = %v, want A-C", []string{lettered[0].Label, lettered[1].Label, lettered[2].Label})
	}

	dragged, _, _ := wordBankOf([]string{"outlawed", "trebled"}, false)
	if dragged[0].Label != "w1" {
		t.Errorf("drag labels = %q, want w1", dragged[0].Label)
	}
}

func TestResolveExamsDefaultsToThePassage(t *testing.T) {
	exams, problem := resolveExams(supportedTypes["reading-mcq-single"], nil, "IELTS")
	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if len(exams) != 1 || exams[0] != "IELTS" {
		t.Errorf("exams = %v, want [IELTS]", exams)
	}
}

func TestResolveExamsSharesATaskBothPapersSet(t *testing.T) {
	exams, problem := resolveExams(supportedTypes["reading-mcq-single"],
		[]string{"ielts", " PTE ", "IELTS"}, "PTE")
	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if len(exams) != 2 || exams[0] != "PTE" || exams[1] != "IELTS" {
		t.Errorf("exams = %v, want [PTE IELTS] deduped and ordered", exams)
	}
}

func TestResolveExamsRefusesATaskThatExamDoesNotSet(t *testing.T) {
	if _, problem := resolveExams(supportedTypes["fill-in-blanks-rw"], []string{"IELTS"}, "PTE"); problem == "" {
		t.Error("IELTS was accepted for a PTE-only task")
	}
	if _, problem := resolveExams(supportedTypes["reading-arrange-passage"], []string{"PTE"}, "IELTS"); problem == "" {
		t.Error("PTE was accepted for an IELTS-only task")
	}
}

func TestResolveExamsRefusesAnUnknownExam(t *testing.T) {
	if _, problem := resolveExams(supportedTypes["reading-mcq-single"], []string{"TOEFL"}, "PTE"); problem == "" {
		t.Error("TOEFL was accepted as an exam")
	}
}

func TestEverySupportedTypeNamesAnExam(t *testing.T) {
	for id, spec := range supportedTypes {
		if len(spec.exams) == 0 {
			t.Errorf("%s names no exam, so nothing can be authored under it", id)
		}
		for _, exam := range spec.exams {
			if exam != "PTE" && exam != "IELTS" {
				t.Errorf("%s names %q, which is not an exam", id, exam)
			}
		}
	}
}

func TestBoxesOfLabelsInOrder(t *testing.T) {
	boxes, labels, problem := boxesOf([]string{" first stage ", "second stage", "", "third stage"})
	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if len(boxes) != 3 {
		t.Fatalf("got %d boxes, want 3: %+v", len(boxes), boxes)
	}
	if boxes[0].Label != "Paragraph A" || boxes[0].Text != "first stage" {
		t.Errorf("first box = %+v, want Paragraph A / first stage", boxes[0])
	}
	if boxes[2].Label != "Paragraph C" {
		t.Errorf("third box label = %q, want Paragraph C", boxes[2].Label)
	}
	if !labels["A"] || !labels["C"] || labels["D"] {
		t.Errorf("labels = %v, want A B C only", labels)
	}
}

func TestBoxesOfNeedsSomethingToOrder(t *testing.T) {
	if _, _, problem := boxesOf([]string{"only one"}); problem == "" {
		t.Error("a single box was accepted as an ordering task")
	}
}

func TestArrangePassageAnswersNameABox(t *testing.T) {
	spec := supportedTypes["reading-arrange-passage"]
	_, labels, problem := boxesOf([]string{"earliest stage", "latest stage"})
	if problem != "" {
		t.Fatalf("boxes: %s", problem)
	}

	options, answers, problem := resolveAnswers(spec,
		newQuestion{Prompt: "Position 1", CorrectAnswers: []string{"b"}}, labels)
	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if len(answers) != 1 || answers[0] != "B" {
		t.Errorf("answers = %v, want [B]", answers)
	}
	if len(options) != 2 || options[0].ID != "A" || options[0].Text != "Paragraph A" {
		t.Errorf("options = %+v, want the boxes as Paragraph A and B", options)
	}

	if _, _, problem := resolveAnswers(spec,
		newQuestion{Prompt: "Position 1", CorrectAnswers: []string{"C"}}, labels); problem == "" {
		t.Error("an answer naming a box that does not exist was accepted")
	}
}

func TestResolveExamsFallsBackToTheTaskWhenThePassageDoesNotSetIt(t *testing.T) {
	exams, problem := resolveExams(supportedTypes["fill-in-blanks-rw"], nil, "IELTS")
	if problem != "" {
		t.Fatalf("unexpected problem: %s", problem)
	}
	if len(exams) != 1 || exams[0] != "PTE" {
		t.Errorf("exams = %v, want [PTE] taken from the task itself", exams)
	}
}
