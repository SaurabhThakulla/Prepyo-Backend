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

func TestGradeSentenceCompletionAcceptsAChosenOption(t *testing.T) {
	q := models.Question{
		TypeID:         "reading-sentence-completion",
		Points:         1,
		CorrectAnswers: []string{"flavour", "flavor"},
		Options: []models.QuestionOption{
			{ID: "flavour", Text: "flavour"},
			{ID: "colour", Text: "colour"},
			{ID: "aroma", Text: "aroma"},
			{ID: "texture", Text: "texture"},
		},
	}

	got, ok := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"flavour"}})
	if !ok {
		t.Fatal("Grade() returned ok=false for a sentence completion question")
	}
	if !got.IsCorrect {
		t.Errorf("choosing the right option scored %v, want correct", got)
	}

	wrong, _ := Grade(q, models.AnswerSubmission{SelectedOptions: []string{"aroma"}})
	if wrong.IsCorrect {
		t.Error("choosing a distractor was marked correct")
	}

	// Typing still works: the options are an additional route to the answer,
	// not a replacement for the one the seeded questions were written against.
	typed, _ := Grade(q, models.AnswerSubmission{TextResponse: "flavor"})
	if !typed.IsCorrect {
		t.Error("an accepted alternative spelling typed in was marked wrong")
	}
}

func TestGradeListeningSummary(t *testing.T) {
	q := models.Question{
		TypeID:         "summarize-spoken-text",
		Points:         10,
		CorrectAnswers: []string{"renewable", "grid", "storage", "battery", "stability"},
	}

	// Good summary within 50-70 words (59 words) matching keywords
	goodSummary := "The transition toward renewable energy introduces substantial intermittency issues that challenge power grid stability across the country. To address these volatility problems, electrical engineers are deploying utility scale battery storage and intelligent microgrids to regulate energy consumption effectively. These advanced solutions ensure that frequency remains constant while integrating sustainable solar and wind sources into national grids."
	got, ok := Grade(q, models.AnswerSubmission{TextResponse: goodSummary})
	if !ok {
		t.Fatal("Grade() returned ok=false for summarize-spoken-text")
	}
	if got.Score < 8 {
		t.Errorf("expected high score for good summary, got %v", got.Score)
	}

	// Empty summary
	emptyGot, _ := Grade(q, models.AnswerSubmission{TextResponse: ""})
	if emptyGot.Score != 0 || emptyGot.IsCorrect {
		t.Errorf("empty summary scored %v, want 0", emptyGot.Score)
	}
}

// ---------------------------------------------------------------------------
// IELTS Band Table Tests
// ---------------------------------------------------------------------------

func TestIELTSReadingBand(t *testing.T) {
	tests := []struct {
		correct int
		want    float64
	}{
		{40, 9.0},
		{39, 9.0},
		{38, 8.5},
		{37, 8.5},
		{36, 8.0},
		{35, 8.0},
		{34, 7.5},
		{33, 7.5},
		{32, 7.0},
		{30, 7.0},
		{29, 6.5},
		{27, 6.5},
		{26, 6.0},
		{23, 6.0},
		{22, 5.5},
		{19, 5.5},
		{18, 5.0},
		{15, 5.0},
		{14, 4.5},
		{13, 4.5},
		{12, 4.0},
		{10, 4.0},
		{9, 3.5},
		{8, 3.5},
		{7, 3.0},
		{6, 3.0},
		{5, 2.5},
		{4, 2.5},
		{3, 2.0},
		{1, 2.0},
		{0, 2.0},
	}
	for _, tc := range tests {
		got := IELTSReadingBand(tc.correct)
		if got != tc.want {
			t.Errorf("IELTSReadingBand(%d) = %.1f, want %.1f", tc.correct, got, tc.want)
		}
	}
}

func TestIELTSReadingBandClamps(t *testing.T) {
	if got := IELTSReadingBand(-5); got != 2.0 {
		t.Errorf("IELTSReadingBand(-5) = %.1f, want 2.0", got)
	}
	if got := IELTSReadingBand(50); got != 9.0 {
		t.Errorf("IELTSReadingBand(50) = %.1f, want 9.0", got)
	}
}

func TestIELTSListeningBand(t *testing.T) {
	tests := []struct {
		correct int
		want    float64
	}{
		{40, 9.0},
		{39, 9.0},
		{37, 8.5},
		{35, 8.0},
		{32, 7.5},
		{30, 7.0},
		{26, 6.5},
		{23, 6.0},
		{18, 5.5},
		{16, 5.0},
		{13, 4.5},
		{11, 4.0},
		{8, 3.5},
		{6, 3.0},
		{4, 2.5},
		{3, 2.0},
		{0, 2.0},
	}
	for _, tc := range tests {
		got := IELTSListeningBand(tc.correct)
		if got != tc.want {
			t.Errorf("IELTSListeningBand(%d) = %.1f, want %.1f", tc.correct, got, tc.want)
		}
	}
}

func TestEstimateFromRawMarksUsesTableForIELTSReading(t *testing.T) {
	scale := Scale{Min: 0, Max: 9, Step: 0.5}

	// 23/40 should give Band 6.0 via the table, not 5.0 via linear interpolation.
	got := scale.EstimateFromRawMarks("IELTS", "reading", 23, 40)
	if got != 6.0 {
		t.Errorf("EstimateFromRawMarks(IELTS, reading, 23, 40) = %.1f, want 6.0", got)
	}

	// 30/40 should give Band 7.0 via the table, not 6.5 via linear.
	got = scale.EstimateFromRawMarks("IELTS", "reading", 30, 40)
	if got != 7.0 {
		t.Errorf("EstimateFromRawMarks(IELTS, reading, 30, 40) = %.1f, want 7.0", got)
	}
}

func TestEstimateFromRawMarksFallsBackForPTE(t *testing.T) {
	scale := Scale{Min: 10, Max: 90, Step: 1}

	// PTE should still use linear interpolation.
	got := scale.EstimateFromRawMarks("PTE", "reading", 30, 40)
	want := scale.EstimateFromAccuracy(30.0 / 40.0)
	if got != want {
		t.Errorf("EstimateFromRawMarks(PTE, reading, 30, 40) = %.1f, want %.1f (linear)", got, want)
	}
}

func TestEstimateFromRawMarksFallsBackForNon40Questions(t *testing.T) {
	scale := Scale{Min: 0, Max: 9, Step: 0.5}

	// IELTS reading with 13 questions (single passage practice) should use linear.
	got := scale.EstimateFromRawMarks("IELTS", "reading", 10, 13)
	want := scale.EstimateFromAccuracy(10.0 / 13.0)
	if got != want {
		t.Errorf("EstimateFromRawMarks(IELTS, reading, 10, 13) = %.1f, want %.1f (linear)", got, want)
	}
}

