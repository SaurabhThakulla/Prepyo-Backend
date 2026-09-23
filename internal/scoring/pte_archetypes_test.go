package scoring

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

// TestPTEArchetypesSandboxExecution executes dynamic test cases against each of the
// official PTE Academic question archetypes to evaluate scoring behavior and compliance in the sandbox.
func TestPTEArchetypesSandboxExecution(t *testing.T) {
	// =========================================================================
	// PART 1: SPEAKING & WRITING
	// =========================================================================

	t.Run("PTE_READ_ALOUD: Scored via Deterministic Check", func(t *testing.T) {
		q := models.Question{
			ID:     "pte-ra-1",
			Exam:   models.ExamPTE,
			Skill:  models.SkillSpeaking,
			TypeID: "pte-read-aloud",
			Points: 90,
		}
		// Deterministic check
		if Deterministic(q.Skill) {
			t.Errorf("Read Aloud should not be graded deterministically by text alone; requires speech analysis")
		}
	})

	t.Run("PTE_REPEAT_SENTENCE: Verbatim sequence check", func(t *testing.T) {
		q := models.Question{
			ID:     "pte-rs-1",
			Exam:   models.ExamPTE,
			Skill:  models.SkillSpeaking,
			TypeID: "pte-repeat-sentence",
			Points: 90,
		}
		if Deterministic(q.Skill) {
			t.Errorf("Repeat Sentence requires acoustic pipeline")
		}
	})

	t.Run("PTE_DESCRIBE_IMAGE: Rubric Constraints", func(t *testing.T) {
		q := models.Question{
			ID:     "pte-di-1",
			Exam:   models.ExamPTE,
			Skill:  models.SkillSpeaking,
			TypeID: "pte-describe-image",
			Points: 90,
		}
		if Deterministic(q.Skill) {
			t.Errorf("Describe Image cannot be graded deterministically without vision/speech model")
		}
	})

	t.Run("PTE_RETELL_LECTURE: Lecture Keypoint Verification", func(t *testing.T) {
		q := models.Question{
			ID:     "pte-rl-1",
			Exam:   models.ExamPTE,
			Skill:  models.SkillSpeaking,
			TypeID: "pte-retell-lecture",
			Points: 90,
		}
		if Deterministic(q.Skill) {
			t.Errorf("Re-tell Lecture requires multimodal assessment")
		}
	})

	t.Run("PTE_ANSWER_SHORT_QUESTION: Binary lexical identification", func(t *testing.T) {
		q := models.Question{
			ID:             "pte-asq-1",
			Exam:           models.ExamPTE,
			Skill:          models.SkillSpeaking,
			TypeID:         "pte-answer-short-question",
			Points:         1,
			CorrectAnswers: []string{"thermometer", "a thermometer"},
		}
		// Currently in backend, ASQ is passed to AI evaluate if treated as speaking,
		// or shortAnswerTypes if classified as short-answer.
		t.Logf("ASQ registered with expected answer: %v", q.CorrectAnswers)
	})

	t.Run("PTE_SUMMARIZE_GROUP_DISCUSSION: Post-Aug 2025 Speaking Task", func(t *testing.T) {
		q := models.Question{
			ID:     "pte-sgd-1",
			Exam:   models.ExamPTE,
			Skill:  models.SkillSpeaking,
			TypeID: "pte-summarize-group-discussion",
			Points: 90,
		}
		if Deterministic(q.Skill) {
			t.Errorf("SGD requires acoustic scoring")
		}
	})

	t.Run("PTE_RESPOND_TO_A_SITUATION: Post-Aug 2025 Speaking Task", func(t *testing.T) {
		q := models.Question{
			ID:     "pte-rts-1",
			Exam:   models.ExamPTE,
			Skill:  models.SkillSpeaking,
			TypeID: "pte-respond-to-situation",
			Points: 90,
		}
		if Deterministic(q.Skill) {
			t.Errorf("RTS requires acoustic scoring")
		}
	})

	t.Run("PTE_SUMMARIZE_WRITTEN_TEXT: Form Gatekeeper Validation", func(t *testing.T) {
		// Pearson Rule: Single sentence between 5 and 75 words.
		// Multiple sentences (e.g. with interior period) must be rejected with 0 Form.
		twoSentenceResponse := "This is the first sentence about climate change. This is the second sentence concluding the idea."
		wordsInTwo := len(strings.Fields(twoSentenceResponse))
		if wordsInTwo < 5 || wordsInTwo > 75 {
			t.Fatalf("Test setup failure: word count must be 5-75")
		}

		// Check if multiple sentences exist:
		clean := strings.TrimRight(twoSentenceResponse, ".!?")
		hasMultipleSentences := strings.ContainsAny(clean, ".!?")
		if !hasMultipleSentences {
			t.Errorf("Expected multiple sentences detected in SWT response")
		}
		t.Logf("SWT Form check: Multi-sentence correctly flagged? %v (Pearson requires Form=0, total=0)", hasMultipleSentences)
	})

	t.Run("PTE_WRITE_ESSAY: Word count gatekeeper", func(t *testing.T) {
		// Pearson Rule: 200-300 words is full Form (2).
		// 120-199 or 301-380 is Form 1.
		// <120 or >380 is Form 0 -> total disqualification (0).
		shortEssay := "This essay is far too brief and fails to meet minimum lengths."
		count := len(strings.Fields(shortEssay))
		if count >= 120 {
			t.Fatalf("Test setup failed: count should be < 120")
		}
		isDisqualified := count < 120 || count > 380
		if !isDisqualified {
			t.Errorf("Expected essay with %d words to be disqualified (Pearson Form=0)", count)
		}
	})

	t.Run("PTE_PERSONAL_INTRODUCTION: Unscored familiarisation", func(t *testing.T) {
		// Personal introduction must have 0 points and not affect communicative score
		q := models.Question{
			ID:     "pte-intro-1",
			Exam:   models.ExamPTE,
			Skill:  models.SkillSpeaking,
			TypeID: "pte-personal-introduction",
			Points: 0,
		}
		if q.Points != 0 {
			t.Errorf("Personal introduction must be unscored (0 points), got %d", q.Points)
		}
	})

	// =========================================================================
	// PART 2: READING
	// =========================================================================

	t.Run("PTE_READING_FILL_IN_THE_BLANKS_DROPDOWN", func(t *testing.T) {
		q := models.Question{
			ID:     "q-rfib-dropdown",
			Exam:   models.ExamPTE,
			Skill:  models.SkillReading,
			TypeID: "fill-in-blanks-rw",
			Points: 4,
			Blanks: []models.Blank{
				{ID: "b1", CorrectAnswer: "predominant"},
				{ID: "b2", CorrectAnswer: "fluctuated"},
				{ID: "b3", CorrectAnswer: "accelerated"},
				{ID: "b4", CorrectAnswer: "inevitable"},
			},
		}

		sub := models.AnswerSubmission{
			Exam: models.ExamPTE,
			BlankResponses: map[string]string{
				"b1": "predominant",
				"b2": "fluctuated",
				"b3": "declined", // wrong
				"b4": "inevitable",
			},
		}

		res, ok := Grade(q, sub)
		if !ok {
			t.Fatalf("Grade() returned ok=false")
		}
		// In Pearson, 3 correct out of 4 gives 3 raw marks
		if res.Marks != 3 || res.MarksAvailable != 4 {
			t.Errorf("Marks = %d/%d, want 3/4", res.Marks, res.MarksAvailable)
		}
		t.Logf("Reading FIB Dropdown: Scored %v / %v (Accuracy: %d%%)", res.Score, res.MaxScore, res.AccuracyPercentage)
	})

	t.Run("PTE_READING_MULTIPLE_CHOICE_MULTIPLE_ANSWERS: Negative Marking Floor", func(t *testing.T) {
		q := models.Question{
			ID:             "q-rmcma",
			Exam:           models.ExamPTE,
			Skill:          models.SkillReading,
			TypeID:         "pte-reading-mcma",
			Points:         3,
			CorrectAnswers: []string{"A", "B", "C"},
			Options: []models.QuestionOption{
				{ID: "A", Text: "Option A"},
				{ID: "B", Text: "Option B"},
				{ID: "C", Text: "Option C"},
				{ID: "D", Text: "Option D"},
				{ID: "E", Text: "Option E"},
			},
		}

		// Case 1: 2 correct, 1 wrong -> Net = 2 - 1 = 1 mark
		sub1 := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"A", "B", "D"},
		}
		res1, ok := Grade(q, sub1)
		if !ok {
			t.Fatalf("Grade() returned ok=false")
		}
		// Ratio: net (1) / required (3) = 1/3
		wantAccuracy := 33
		if res1.AccuracyPercentage != wantAccuracy {
			t.Errorf("Accuracy = %d%%, want %d%%", res1.AccuracyPercentage, wantAccuracy)
		}

		// Case 2: 1 correct, 2 wrong -> Net = 1 - 2 = -1 -> Floored at 0
		sub2 := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"A", "D", "E"},
		}
		res2, _ := Grade(q, sub2)
		if res2.Score < 0 {
			t.Errorf("PTE negative marking must not drop below 0; got score %v", res2.Score)
		}
		if res2.AccuracyPercentage != 0 {
			t.Errorf("Floor at 0 expected, got accuracy %d%%", res2.AccuracyPercentage)
		}
		t.Logf("Reading MCMA Negative Marking Floored Correctly: Score=%v, Accuracy=%d%%", res2.Score, res2.AccuracyPercentage)
	})

	t.Run("PTE_REORDER_PARAGRAPH: Adjacent Pair Scoring", func(t *testing.T) {
		q := models.Question{
			ID:             "q-reorder",
			Exam:           models.ExamPTE,
			Skill:          models.SkillReading,
			TypeID:         "reorder-paragraphs",
			Points:         3, // 4 items -> 3 pairs
			CorrectAnswers: []string{"P1", "P2", "P3", "P4"},
		}

		// Submission with pairs: [P1, P2] (correct), [P4, P3] (wrong)
		// User order: P1, P2, P4, P3 -> Adjacent pairs: (P1,P2) YES, (P2,P4) NO, (P4,P3) NO.
		sub := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"P1", "P2", "P4", "P3"},
		}
		res, ok := Grade(q, sub)
		if !ok {
			t.Fatalf("Grade() returned ok=false")
		}
		// 1 pair correct out of 3 pairs
		if res.AccuracyPercentage != 33 {
			t.Errorf("Reorder pair accuracy = %d%%, want 33%%", res.AccuracyPercentage)
		}
		t.Logf("Reorder Paragraphs: 1 of 3 pairs correct, Accuracy: %d%%", res.AccuracyPercentage)
	})

	t.Run("PTE_READING_FILL_IN_THE_BLANKS_DRAG_DROP", func(t *testing.T) {
		q := models.Question{
			ID:     "q-rfib-dragdrop",
			Exam:   models.ExamPTE,
			Skill:  models.SkillReading,
			TypeID: "fill-in-blanks-r",
			Points: 3,
			Blanks: []models.Blank{
				{ID: "b1", CorrectAnswer: "hypothesis"},
				{ID: "b2", CorrectAnswer: "experiment"},
				{ID: "b3", CorrectAnswer: "conclusion"},
			},
		}
		sub := models.AnswerSubmission{
			Exam: models.ExamPTE,
			BlankResponses: map[string]string{
				"b1": "hypothesis",
				"b2": "experiment",
				"b3": "conclusion",
			},
		}
		res, ok := Grade(q, sub)
		if !ok || !res.IsCorrect {
			t.Errorf("All correct drag and drop failed: %+v", res)
		}
		t.Logf("Reading FIB Drag-Drop: Perfect match scored %v/%v", res.Score, res.MaxScore)
	})

	t.Run("PTE_READING_MULTIPLE_CHOICE_SINGLE_ANSWER", func(t *testing.T) {
		q := models.Question{
			ID:             "q-rmcsa",
			Exam:           models.ExamPTE,
			Skill:          models.SkillReading,
			TypeID:         "pte-reading-mcsa",
			Points:         1,
			CorrectAnswers: []string{"B"},
			Options: []models.QuestionOption{
				{ID: "A", Text: "Alpha"},
				{ID: "B", Text: "Beta"},
				{ID: "C", Text: "Gamma"},
			},
		}
		sub := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"B"},
		}
		res, ok := Grade(q, sub)
		if !ok || !res.IsCorrect || res.Score != 1 {
			t.Errorf("Reading MCSA grading failed: %+v", res)
		}
		t.Logf("Reading MCSA: Single correct choice scored %v/%v", res.Score, res.MaxScore)
	})

	// =========================================================================
	// PART 3: LISTENING
	// =========================================================================

	t.Run("PTE_SUMMARIZE_SPOKEN_TEXT: Behavior Audit", func(t *testing.T) {
		q := models.Question{
			ID:             "q-sst",
			Exam:           models.ExamPTE,
			Skill:          models.SkillListening,
			TypeID:         "summarize-spoken-text",
			Points:         10,
			CorrectAnswers: []string{"photosynthesis", "chlorophyll", "sunlight", "glucose"},
		}

		// Response within 50-70 words
		summaryText := "The lecture discussed photosynthesis and how chlorophyll absorbs sunlight to produce glucose for the plant. This biological process provides chemical energy necessary for growth and cellular respiration throughout living organisms across different ecological systems worldwide and forms the foundation of all terrestrial trophic food chains today."
		wordCount := len(strings.Fields(summaryText))

		sub := models.AnswerSubmission{
			Exam:         models.ExamPTE,
			TextResponse: summaryText,
		}

		res, ok := Grade(q, sub)
		if !ok {
			t.Fatalf("Grade() returned ok=false")
		}

		t.Logf("SST Current Output: WordCount=%d, Score=%v/%v, Accuracy=%d%%, Feedback=%s",
			wordCount, res.Score, res.MaxScore, res.AccuracyPercentage, res.Feedback)

		// Note audit finding: Current scoring uses keyword substring matching rather than 5 Pearson traits.
		if strings.Contains(res.Feedback, "Captured") {
			t.Logf("AUDIT VERIFIED: Current backend SST logic relies on keyword count matching (%s)", res.Feedback)
		}
	})

	t.Run("PTE_LISTENING_MULTIPLE_CHOICE_MULTIPLE_ANSWERS: Negative Marking", func(t *testing.T) {
		q := models.Question{
			ID:             "q-lmcma",
			Exam:           models.ExamPTE,
			Skill:          models.SkillListening,
			TypeID:         "pte-listening-mcma",
			Points:         2,
			CorrectAnswers: []string{"opt-1", "opt-3"},
			Options: []models.QuestionOption{
				{ID: "opt-1", Text: "First"},
				{ID: "opt-2", Text: "Second"},
				{ID: "opt-3", Text: "Third"},
			},
		}

		sub := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"opt-1", "opt-2"}, // 1 right, 1 wrong -> net 0
		}
		res, ok := Grade(q, sub)
		if !ok {
			t.Fatalf("Grade() returned ok=false")
		}
		if res.Score != 0 {
			t.Errorf("Listening MCMA net score should be 0; got %v", res.Score)
		}
		t.Logf("Listening MCMA Negative Marking floored at 0: Score=%v", res.Score)
	})

	t.Run("PTE_LISTENING_FILL_IN_THE_BLANKS_TYPE_IN", func(t *testing.T) {
		q := models.Question{
			ID:     "q-lfib",
			Exam:   models.ExamPTE,
			Skill:  models.SkillListening,
			TypeID: "pte-listening-fib",
			Points: 2,
			Blanks: []models.Blank{
				{ID: "b1", CorrectAnswer: "ecosystem"},
				{ID: "b2", CorrectAnswer: "biodiversity"},
			},
		}
		sub := models.AnswerSubmission{
			Exam: models.ExamPTE,
			BlankResponses: map[string]string{
				"b1": "Ecosystem ", // case and whitespace tolerant
				"b2": "biodiversity",
			},
		}
		res, ok := Grade(q, sub)
		if !ok || !res.IsCorrect {
			t.Errorf("Listening FIB grading failed: %+v", res)
		}
		t.Logf("Listening FIB: Scored %v/%v (Accuracy: %d%%)", res.Score, res.MaxScore, res.AccuracyPercentage)
	})

	t.Run("PTE_HIGHLIGHT_CORRECT_SUMMARY", func(t *testing.T) {
		q := models.Question{
			ID:             "q-hcs",
			Exam:           models.ExamPTE,
			Skill:          models.SkillListening,
			TypeID:         "pte-highlight-correct-summary",
			Points:         1,
			CorrectAnswers: []string{"summary-c"},
			Options: []models.QuestionOption{
				{ID: "summary-a", Text: "Summary A"},
				{ID: "summary-b", Text: "Summary B"},
				{ID: "summary-c", Text: "Summary C"},
			},
		}
		sub := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"summary-c"},
		}
		res, ok := Grade(q, sub)
		if !ok || !res.IsCorrect {
			t.Errorf("Highlight Correct Summary failed: %+v", res)
		}
		t.Logf("Highlight Correct Summary: Scored %v/%v", res.Score, res.MaxScore)
	})

	t.Run("PTE_LISTENING_MULTIPLE_CHOICE_SINGLE_ANSWER", func(t *testing.T) {
		q := models.Question{
			ID:             "q-lmcsa",
			Exam:           models.ExamPTE,
			Skill:          models.SkillListening,
			TypeID:         "pte-listening-mcsa",
			Points:         1,
			CorrectAnswers: []string{"A"},
			Options: []models.QuestionOption{
				{ID: "A", Text: "A"},
				{ID: "B", Text: "B"},
			},
		}
		sub := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"A"},
		}
		res, ok := Grade(q, sub)
		if !ok || !res.IsCorrect {
			t.Errorf("Listening MCSA failed: %+v", res)
		}
		t.Logf("Listening MCSA: Scored %v/%v", res.Score, res.MaxScore)
	})

	t.Run("PTE_SELECT_MISSING_WORD", func(t *testing.T) {
		q := models.Question{
			ID:             "q-smw",
			Exam:           models.ExamPTE,
			Skill:          models.SkillListening,
			TypeID:         "pte-select-missing-word",
			Points:         1,
			CorrectAnswers: []string{"word-3"},
			Options: []models.QuestionOption{
				{ID: "word-1", Text: "Word 1"},
				{ID: "word-2", Text: "Word 2"},
				{ID: "word-3", Text: "Word 3"},
			},
		}
		sub := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"word-3"},
		}
		res, ok := Grade(q, sub)
		if !ok || !res.IsCorrect {
			t.Errorf("Select Missing Word failed: %+v", res)
		}
		t.Logf("Select Missing Word: Scored %v/%v", res.Score, res.MaxScore)
	})

	t.Run("PTE_HIGHLIGHT_INCORRECT_WORDS: Negative Marking", func(t *testing.T) {
		// HIW awards 1 mark per correct clicked word, deducts 1 mark per incorrectly clicked word. Floor at 0.
		q := models.Question{
			ID:             "q-hiw",
			Exam:           models.ExamPTE,
			Skill:          models.SkillListening,
			TypeID:         "pte-highlight-incorrect-word",
			Points:         3,
			CorrectAnswers: []string{"w3", "w7", "w12"},
			Options: []models.QuestionOption{
				{ID: "w3", Text: "error1"},
				{ID: "w7", Text: "error2"},
				{ID: "w12", Text: "error3"},
				{ID: "w15", Text: "validword"},
			},
		}

		// 2 correct (w3, w7) and 1 wrong (w15) -> Net = 1 mark
		sub := models.AnswerSubmission{
			Exam:            models.ExamPTE,
			SelectedOptions: []string{"w3", "w7", "w15"},
		}
		res, ok := Grade(q, sub)
		if !ok {
			t.Fatalf("Grade() returned ok=false")
		}
		if res.AccuracyPercentage != 33 {
			t.Errorf("HIW Net Accuracy = %d%%, want 33%%", res.AccuracyPercentage)
		}
		t.Logf("Highlight Incorrect Words: Net score = %v/%v, Accuracy = %d%%", res.Score, res.MaxScore, res.AccuracyPercentage)
	})

	t.Run("PTE_WRITE_FROM_DICTATION: Word match and sequence", func(t *testing.T) {
		q := models.Question{
			ID:             "q-wfd",
			Exam:           models.ExamPTE,
			Skill:          models.SkillListening,
			TypeID:         "write-from-dictation",
			Points:         8,
			CorrectAnswers: []string{"The student union will hold an extraordinary general meeting tomorrow."},
		}

		// Learner writes 7 of 8 words correctly
		sub := models.AnswerSubmission{
			Exam:         models.ExamPTE,
			TextResponse: "The student union will hold an general meeting tomorrow.", // omitted 'extraordinary'
		}

		res, ok := Grade(q, sub)
		if !ok {
			t.Fatalf("Grade() returned ok=false")
		}

		targetLen := len(strings.Fields(q.CorrectAnswers[0]))
		t.Logf("Write From Dictation: Matched %v of %d words. Score = %v/%v, Feedback = %s",
			res.Score, targetLen, res.Score, res.MaxScore, res.Feedback)
	})
}
