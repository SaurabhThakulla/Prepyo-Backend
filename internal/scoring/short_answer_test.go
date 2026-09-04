package scoring

import (
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func gapQuestion() models.Question {
	return models.Question{
		ID:     "q-gap",
		TypeID: "reading-sentence-completion",
		Points: 1,
		// Two spellings of one answer, not two gaps. Which is the whole reason
		// this task type is dispatched separately.
		CorrectAnswers: []string{"flavour", "flavor"},
	}
}

func TestGradeShortAnswerAcceptsEverySpelling(t *testing.T) {
	for _, given := range []string{"flavour", "flavor", "FLAVOUR", "  Flavour. ", "flavour!"} {
		result, ok := Grade(gapQuestion(), models.AnswerSubmission{TextResponse: given})
		if !ok {
			t.Fatalf("%q: no grader ran", given)
		}
		if !result.IsCorrect || result.Score != 1 {
			t.Errorf("%q: correct=%v score=%v, want correct with full marks",
				given, result.IsCorrect, result.Score)
		}
	}
}

// The bug this guards against: gradeOrderedAnswers would read the two accepted
// spellings as two separate answers and mark a right response half correct.
func TestGradeShortAnswerDoesNotTreatAlternativesAsExtraGaps(t *testing.T) {
	result, _ := Grade(gapQuestion(), models.AnswerSubmission{TextResponse: "flavour"})

	if result.AccuracyPercentage != 100 {
		t.Errorf("accuracy = %d, want 100", result.AccuracyPercentage)
	}
	if result.MaxScore != 1 {
		t.Errorf("maxScore = %v, want 1", result.MaxScore)
	}
}

func TestGradeShortAnswerRejectsAnythingElse(t *testing.T) {
	for _, given := range []string{"taste", "", "   ", "flavours"} {
		result, _ := Grade(gapQuestion(), models.AnswerSubmission{TextResponse: given})
		if result.IsCorrect || result.Score != 0 {
			t.Errorf("%q: correct=%v score=%v, want wrong with no marks",
				given, result.IsCorrect, result.Score)
		}
		if result.ErrorTag == "" {
			t.Errorf("%q: a wrong answer must carry an error tag for the mistake bank", given)
		}
	}
}

// A client may render a gap as a word bank rather than a text field.
func TestGradeShortAnswerAcceptsASingleSelectedOption(t *testing.T) {
	result, _ := Grade(gapQuestion(), models.AnswerSubmission{SelectedOptions: []string{"flavor"}})

	if !result.IsCorrect {
		t.Errorf("correct = false, want true for a single selected option")
	}
}

func TestGradeShortAnswerReportsWhatWasExpected(t *testing.T) {
	result, _ := Grade(gapQuestion(), models.AnswerSubmission{TextResponse: "taste"})

	if result.CorrectDisplay != "flavour / flavor" {
		t.Errorf("correctDisplay = %q, want both accepted spellings", result.CorrectDisplay)
	}
	if result.UserDisplay != "taste" {
		t.Errorf("userDisplay = %q, want what the learner typed", result.UserDisplay)
	}
}

// The reading task types that are answered by choosing must keep going through
// gradeChoice, not through the short-answer path added alongside it.
func TestReadingChoiceTypesStillGradeAsChoices(t *testing.T) {
	q := models.Question{
		ID:     "q-tf",
		TypeID: "reading-true-false",
		Points: 1,
		Options: []models.QuestionOption{
			{ID: "TRUE", Text: "True"},
			{ID: "FALSE", Text: "False"},
		},
		CorrectAnswers: []string{"FALSE"},
	}

	right, ok := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"FALSE"}})
	if !ok || !right.IsCorrect {
		t.Errorf("ok=%v correct=%v, want a correct grade", ok, right.IsCorrect)
	}

	wrong, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"TRUE"}})
	if wrong.IsCorrect {
		t.Error("selecting TRUE scored correct, want wrong")
	}
}
