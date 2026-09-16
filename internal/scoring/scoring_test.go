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

func TestRoundIELTSBand(t *testing.T) {
	tests := []struct {
		raw  float64
		want float64
	}{
		// Standard Cambridge IELTS rounding examples:
		{6.0, 6.0},
		{6.125, 6.0},  // < 0.25 rounds down to .0
		{6.24, 6.0},
		{6.25, 6.5},   // >= 0.25 rounds up to .5
		{6.375, 6.5},
		{6.5, 6.5},
		{6.625, 6.5},
		{6.74, 6.5},
		{6.75, 7.0},   // >= 0.75 rounds up to whole band
		{6.875, 7.0},
		{7.0, 7.0},
		{8.75, 9.0},
		{9.0, 9.0},
		{9.5, 9.0},    // Clamped
		{-1.0, 0.0},   // Clamped
	}
	for _, tc := range tests {
		got := RoundIELTSBand(tc.raw)
		if got != tc.want {
			t.Errorf("RoundIELTSBand(%v) = %v, want %v", tc.raw, got, tc.want)
		}
	}
}

func TestPTEScaling(t *testing.T) {
	tests := []struct {
		accuracy float64
		want     float64
	}{
		{0.0, 10.0},
		{1.0, 90.0},
		{0.5, 50.0},
		{0.75, 70.0},
		{0.25, 30.0},
		{-0.1, 10.0},  // Clamped
		{1.2, 90.0},   // Clamped
	}
	for _, tc := range tests {
		got := PTEEstimateFromAccuracy(tc.accuracy)
		if got != tc.want {
			t.Errorf("PTEEstimateFromAccuracy(%v) = %v, want %v", tc.accuracy, got, tc.want)
		}
	}
}

func TestEstimateOverallIELTS(t *testing.T) {
	scale := Scale{Min: 0, Max: 9, Step: 0.5}

	// 4 skills: 6.5, 6.5, 6.0, 6.0 -> Average = 6.25 -> Officially rounds to 6.5!
	skills := map[models.SkillType]float64{
		models.SkillReading:   6.5,
		models.SkillListening: 6.5,
		models.SkillWriting:   6.0,
		models.SkillSpeaking:  6.0,
	}
	got := scale.EstimateOverall("IELTS", skills, 0, 0)
	if got != 6.5 {
		t.Errorf("EstimateOverall(IELTS, 4 skills avg 6.25) = %.1f, want 6.5", got)
	}

	// 4 skills: 7.0, 6.5, 6.5, 7.0 -> Average = 6.75 -> Officially rounds to 7.0!
	skills2 := map[models.SkillType]float64{
		models.SkillReading:   7.0,
		models.SkillListening: 6.5,
		models.SkillWriting:   6.5,
		models.SkillSpeaking:  7.0,
	}
	got2 := scale.EstimateOverall("IELTS", skills2, 0, 0)
	if got2 != 7.0 {
		t.Errorf("EstimateOverall(IELTS, 4 skills avg 6.75) = %.1f, want 7.0", got2)
	}
}

func TestEstimateOverallPTE(t *testing.T) {
	scale := Scale{Min: 10, Max: 90, Step: 1}

	// 4 skills: 65, 70, 62, 68 -> Average = 66.25 -> Rounds to 66
	skills := map[models.SkillType]float64{
		models.SkillReading:   65,
		models.SkillListening: 70,
		models.SkillWriting:   62,
		models.SkillSpeaking:  68,
	}
	got := scale.EstimateOverall("PTE", skills, 0, 0)
	if got != 66 {
		t.Errorf("EstimateOverall(PTE, avg 66.25) = %.1f, want 66", got)
	}
}

func TestConcordance(t *testing.T) {
	// Pearson Concordance: PTE 86 -> IELTS 9.0; PTE 65 -> IELTS 7.0; PTE 50 -> IELTS 6.0
	if got := PTEToIELTSConcordance(86); got != 9.0 {
		t.Errorf("PTEToIELTSConcordance(86) = %.1f, want 9.0", got)
	}
	if got := PTEToIELTSConcordance(68); got != 7.0 {
		t.Errorf("PTEToIELTSConcordance(68) = %.1f, want 7.0", got)
	}
	if got := PTEToIELTSConcordance(54); got != 6.0 {
		t.Errorf("PTEToIELTSConcordance(54) = %.1f, want 6.0", got)
	}

	if got := IELTSToPTEConcordance(7.0); got != 68.0 {
		t.Errorf("IELTSToPTEConcordance(7.0) = %.1f, want 68.0", got)
	}
}


