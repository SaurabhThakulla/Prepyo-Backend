package scoring

import (
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func blanksQuestion() models.Question {
	return models.Question{
		ID:     "q-blanks",
		TypeID: "fill-in-blanks-rw",
		Points: 10,
		Blanks: []models.Blank{
			{ID: "b1", CorrectAnswer: "debunked"},
			{ID: "b2", CorrectAnswer: "strengthens"},
		},
	}
}

func TestGradeBlanks(t *testing.T) {
	tests := []struct {
		name         string
		responses    map[string]string
		wantScore    float64
		wantAccuracy int
		wantCorrect  bool
	}{
		{
			name:         "all correct",
			responses:    map[string]string{"b1": "debunked", "b2": "strengthens"},
			wantScore:    10,
			wantAccuracy: 100,
			wantCorrect:  true,
		},
		{
			name:         "half correct",
			responses:    map[string]string{"b1": "debunked", "b2": "weakens"},
			wantScore:    5,
			wantAccuracy: 50,
		},
		{
			name:         "case and spacing are ignored",
			responses:    map[string]string{"b1": "  Debunked ", "b2": "STRENGTHENS"},
			wantScore:    10,
			wantAccuracy: 100,
			wantCorrect:  true,
		},
		{
			name:         "missing answers score zero",
			responses:    nil,
			wantScore:    0,
			wantAccuracy: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Grade(blanksQuestion(), models.AnswerSubmission{BlankResponses: tc.responses})
			if !ok {
				t.Fatal("Grade() returned ok=false, want a graded result")
			}
			if got.Score != tc.wantScore {
				t.Errorf("Score = %v, want %v", got.Score, tc.wantScore)
			}
			if got.AccuracyPercentage != tc.wantAccuracy {
				t.Errorf("AccuracyPercentage = %d, want %d", got.AccuracyPercentage, tc.wantAccuracy)
			}
			if got.IsCorrect != tc.wantCorrect {
				t.Errorf("IsCorrect = %v, want %v", got.IsCorrect, tc.wantCorrect)
			}
		})
	}
}

func TestGradeBlanksTagsMistakeOnlyWhenWrong(t *testing.T) {
	perfect, _ := Grade(blanksQuestion(), models.AnswerSubmission{
		BlankResponses: map[string]string{"b1": "debunked", "b2": "strengthens"},
	})
	if perfect.ErrorTag != "" {
		t.Errorf("ErrorTag = %q on a correct answer, want empty", perfect.ErrorTag)
	}

	wrong, _ := Grade(blanksQuestion(), models.AnswerSubmission{
		BlankResponses: map[string]string{"b1": "reinforced", "b2": "strengthens"},
	})
	if wrong.ErrorTag == "" {
		t.Error("ErrorTag is empty on a wrong answer, want a tag for the mistake bank")
	}
}

func TestGradeReorderScoresAdjacentPairs(t *testing.T) {
	q := models.Question{
		TypeID:         "reorder-paragraphs",
		Points:         12,
		CorrectAnswers: []string{"p1", "p2", "p3", "p4"},
	}

	tests := []struct {
		name         string
		answer       []string
		wantAccuracy int
	}{
		{"perfect order", []string{"p1", "p2", "p3", "p4"}, 100},
		// p1-p2 and p3-p4 stay adjacent, p2-p3 does not: 2 of 3 pairs.
		{"two pairs intact", []string{"p3", "p4", "p1", "p2"}, 67},
		{"fully reversed", []string{"p4", "p3", "p2", "p1"}, 0},
		{"nothing submitted", nil, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := Grade(q, models.AnswerSubmission{SelectedOptions: tc.answer})
			if !ok {
				t.Fatal("Grade() returned ok=false")
			}
			if got.AccuracyPercentage != tc.wantAccuracy {
				t.Errorf("AccuracyPercentage = %d, want %d", got.AccuracyPercentage, tc.wantAccuracy)
			}
		})
	}
}

func TestGradeDictationCountsEachTargetWordOnce(t *testing.T) {
	q := models.Question{
		TypeID:         "write-from-dictation",
		Points:         10,
		CorrectAnswers: []string{"All submitted assignments must contain academic citations."},
	}

	t.Run("exact match", func(t *testing.T) {
		got, _ := Grade(q, models.AnswerSubmission{
			TextResponse: "All submitted assignments must contain academic citations.",
		})
		if !got.IsCorrect {
			t.Errorf("IsCorrect = false, want true (accuracy %d)", got.AccuracyPercentage)
		}
	})

	t.Run("missing plurals lose marks", func(t *testing.T) {
		got, _ := Grade(q, models.AnswerSubmission{
			TextResponse: "All submitted assignment must contain academic citation.",
		})
		if got.IsCorrect {
			t.Error("IsCorrect = true, want false: two words differ")
		}
	})

	t.Run("repeating a word does not inflate the score", func(t *testing.T) {
		got, _ := Grade(q, models.AnswerSubmission{
			TextResponse: "all all all all all all all",
		})
		// Only the single "all" in the target can be matched.
		if got.AccuracyPercentage > 20 {
			t.Errorf("AccuracyPercentage = %d, want a low score for repeated filler", got.AccuracyPercentage)
		}
	})
}

func TestGradeChoicePenalisesOverSelection(t *testing.T) {
	q := models.Question{
		TypeID: "multiple-choice-multiple",
		Points: 10,
		Options: []models.QuestionOption{
			{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"},
		},
		CorrectAnswers: []string{"a", "b"},
	}

	t.Run("exact selection scores full marks", func(t *testing.T) {
		got, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"a", "b"}})
		if got.Score != 10 {
			t.Errorf("Score = %v, want 10", got.Score)
		}
	})

	t.Run("selecting everything scores zero", func(t *testing.T) {
		got, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"a", "b", "c", "d"}})
		if got.Score != 0 {
			t.Errorf("Score = %v, want 0: two right minus two wrong", got.Score)
		}
	})

	t.Run("duplicates are ignored", func(t *testing.T) {
		got, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"a", "a", "a", "b"}})
		if got.Score != 10 {
			t.Errorf("Score = %v, want 10", got.Score)
		}
	})
}

func TestGradeOrderedAnswers(t *testing.T) {
	q := models.Question{
		TypeID:         "ielts-reading-tfng",
		Points:         12,
		CorrectAnswers: []string{"TRUE", "NOT GIVEN", "FALSE"},
	}

	got, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"true", "FALSE", "FALSE"}})
	if got.AccuracyPercentage != 67 {
		t.Errorf("AccuracyPercentage = %d, want 67 (two of three, case-insensitive)", got.AccuracyPercentage)
	}
}

// An unknown task type must not quietly award full marks.
func TestGradeRejectsUnknownTaskType(t *testing.T) {
	q := models.Question{TypeID: "some-task-nobody-implemented", Points: 10}

	if _, ok := Grade(q, models.AnswerSubmission{TextResponse: "anything"}); ok {
		t.Error("Grade() returned ok=true for an ungradable question, want ok=false")
	}
}

func TestDeterministic(t *testing.T) {
	for _, skill := range []models.SkillType{models.SkillReading, models.SkillListening} {
		if !Deterministic(skill) {
			t.Errorf("Deterministic(%q) = false, want true", skill)
		}
	}
	for _, skill := range []models.SkillType{models.SkillSpeaking, models.SkillWriting} {
		if Deterministic(skill) {
			t.Errorf("Deterministic(%q) = true, want false: needs AI evaluation", skill)
		}
	}
}
