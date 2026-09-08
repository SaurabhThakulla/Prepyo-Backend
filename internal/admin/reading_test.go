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

	for _, field := range []string{"exam", "title", "body", "groups", "difficulty"} {
		if _, ok := problems[field]; !ok {
			t.Errorf("expected a problem for %q, got %v", field, problems)
		}
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
