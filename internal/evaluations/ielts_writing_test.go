package evaluations

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func TestIELTSWordCountCountsHyphenatedWordsOnce(t *testing.T) {
	if got := IELTSWordCount("A well-known check-in desk, isn't it?", ""); got != 6 {
		t.Fatalf("count = %d, want 6 (hyphenated words and contractions are one word each)", got)
	}
}

func TestIELTSWordCountDiscountsCopiedRubric(t *testing.T) {
	prompt := "Some people believe that school uniforms should be abolished because they restrict students' individuality. To what extent do you agree or disagree?"
	response := "Some people believe that school uniforms should be abolished. I disagree strongly with this view."
	// The first nine words repeat the prompt verbatim; six are the learner's own.
	if got := IELTSWordCount(response, prompt); got != 6 {
		t.Fatalf("count = %d, want 6 once the copied prompt is set aside", got)
	}
	// A short shared phrase is coincidence, not copying.
	if got := IELTSWordCount("I agree that uniforms help.", prompt); got != 5 {
		t.Fatalf("count = %d, want 5: a phrase shorter than the copy threshold still counts", got)
	}
}

func TestIELTSShortWritingIsBandOneOnEveryCriterion(t *testing.T) {
	for _, tc := range []struct {
		typeID, first string
	}{
		{"ielts-writing-task2-opinion", "Task Response"},
		{"ielts-writing-task1-figure", "Task Achievement"},
		{"ielts-writing-task1-letter", "Task Achievement"},
	} {
		e := ieltsShortWritingEvaluation(models.Question{Exam: models.ExamIELTS, TypeID: tc.typeID}, 12)
		if e.EstimatedScore == nil || *e.EstimatedScore != 1 {
			t.Fatalf("%s: estimate = %v, want 1", tc.typeID, e.EstimatedScore)
		}
		if len(e.Criteria) != 4 || e.Criteria[0].Name != tc.first {
			t.Fatalf("%s: criteria = %+v, want four starting with %s", tc.typeID, e.Criteria, tc.first)
		}
		for _, c := range e.Criteria {
			if c.Score != 1 || c.MaxScore != 9 {
				t.Fatalf("%s: criterion %s = %v/%v, want 1/9", tc.typeID, c.Name, c.Score, c.MaxScore)
			}
		}
		if !strings.Contains(e.Summary, "12 words") {
			t.Fatalf("%s: summary %q should report the count", tc.typeID, e.Summary)
		}
	}
}

func TestIELTSTooFewWordsCarriesNoBand(t *testing.T) {
	q := models.Question{Exam: models.ExamIELTS, TypeName: "Speaking Part 2 (Cue Card)"}
	e := tooFewWordsEvaluation(q, models.ExamVersion{MinScore: 0, MaxScore: 9}, "the museum", 2)
	if e.EstimatedScore != nil {
		t.Fatalf("estimate = %v, want none: IELTS Band 0 means the candidate did not attend", *e.EstimatedScore)
	}
	if len(e.Criteria) != 0 {
		t.Fatalf("criteria = %+v, want none rather than a non-IELTS Content criterion", e.Criteria)
	}
	if !strings.HasPrefix(e.Summary, "Only 2 words could be made out") {
		t.Fatalf("summary = %q", e.Summary)
	}
}
