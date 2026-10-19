package evaluations

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

const passage = "Every flat map distorts the globe it represents."

func pteQuestion() models.Question {
	return models.Question{
		Exam:           models.ExamPTE,
		Skill:          models.SkillSpeaking,
		TypeID:         "read-aloud",
		TypeName:       "Read Aloud",
		ContextPassage: passage,
	}
}

func pteVersion() models.ExamVersion {
	return models.ExamVersion{Exam: models.ExamPTE, MinScore: 10, MaxScore: 90, ScoreStep: 1}
}

func TestVerbatimEvaluationScoresOnPearsonTraits(t *testing.T) {
	delivery := scoring.Delivery{
		DurationSeconds: 4, WordsSpoken: 8, PauseCount: 1, LongestPauseSeconds: 0.8, SpeakingRatio: 0.8,
	}
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, delivery, nil)

	if evaluation.EstimatedScore == nil {
		t.Fatal("a perfect read produced no score")
	}
	// Content 90 + Pronunciation 65 (assumed) + Fluency 76 (one pause in a short read): (90 + 65 + 76)/3 = 77.
	if *evaluation.EstimatedScore != 77 {
		t.Errorf("perfect read with one pause scored %.0f, want 77", *evaluation.EstimatedScore)
	}
	names := []string{}
	for _, c := range evaluation.Criteria {
		names = append(names, c.Name)
	}
	if strings.Join(names, ",") != "Content,Pronunciation,Oral Fluency" {
		t.Fatalf("criteria = %v", names)
	}
	if evaluation.Criteria[1].Score != 65 || evaluation.Criteria[1].MaxScore != 90 {
		t.Errorf("pronunciation = %+v, want the assumed 65 of 90", evaluation.Criteria[1])
	}
	if !strings.Contains(evaluation.Summary, "Pronunciation was not assessed") {
		t.Errorf("summary does not say pronunciation went unassessed: %q", evaluation.Summary)
	}
	if len(evaluation.Words) != len(strings.Fields(passage)) {
		t.Errorf("words = %d, want one mark per word of the passage", len(evaluation.Words))
	}
}

func TestVerbatimEvaluationMarksDownAMissedRead(t *testing.T) {
	full := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, scoring.Delivery{}, nil)
	partial := verbatimEvaluation(pteQuestion(), pteVersion(), passage, "every flat map", scoring.Delivery{}, nil)

	if *partial.EstimatedScore >= *full.EstimatedScore {
		t.Errorf("half a passage scored %.0f against %.0f for the whole thing",
			*partial.EstimatedScore, *full.EstimatedScore)
	}
	if len(partial.Weaknesses) == 0 {
		t.Error("no weakness listed for a passage that was mostly skipped")
	}
}

func TestVerbatimEvaluationNothingSaidScoresTheFloor(t *testing.T) {
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, "", scoring.Delivery{}, nil)
	if *evaluation.EstimatedScore != 10 {
		t.Errorf("an empty read scored %.0f, want 10", *evaluation.EstimatedScore)
	}
}

// The recording behind the 81 that APEUni-style marking should bring down.
func TestVerbatimEvaluationPhytoplanktonRecording(t *testing.T) {
	given := "Marine phytoplankton play an essential role in regulating Earth's climate by absorbing carbon dioxide through photosynthesis. When these microscopic organisms die, a significant portion of the organic carbon sinks into the deep ocean, sequestering greenhouse gases for centuries."
	heard := "Marine psychopathians play an essential role in regulating Earth's climate by absorbing carbon dioxide through photosynthesis. When this microscopic organism dies, a significant portion of the organic carbon sinks into the deep ocean, squirting greenhouse gases for centuries."
	delivery := scoring.Delivery{DurationSeconds: 23, WordsSpoken: 38, PauseCount: 2, LongestPauseSeconds: 0.9, SpeakingRatio: 0.8}

	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), given, heard, delivery, nil)
	// Content 74 (4 of 5), Pronunciation 65 (assumed), Fluency 38 (99 wpm & 2 pauses): (74 + 65 + 38)/3 = 59.
	if *evaluation.EstimatedScore != 59 {
		t.Errorf("scored %.0f, want 59", *evaluation.EstimatedScore)
	}
	missed := 0
	for _, w := range evaluation.Words {
		if !w.Said {
			missed++
		}
	}
	if missed != 5 {
		t.Errorf("%d words marked missed, want 5", missed)
	}
}

func TestVerbatimEvaluationUsesTheAudioRating(t *testing.T) {
	rating := &ai.DeliveryRating{Pronunciation: 75, PronunciationFeedback: "Stress in phytoplankton.", Fluency: 80, FluencyFeedback: "Word by word.", Unclear: []string{"flat"}}
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, scoring.Delivery{}, rating)

	// Pronunciation from the words heard clearly: 7 of 8 said words clear is 65.
	// Content 90 + Pronunciation 65 + Fluency 80 (the model's, nothing measured): (90 + 65 + 80)/3 = 78.
	if *evaluation.EstimatedScore != 78 || evaluation.Criteria[1].Score != 65 {
		t.Errorf("scored %.0f with pronunciation %.0f, want 78 and 65", *evaluation.EstimatedScore, evaluation.Criteria[1].Score)
	}
	if evaluation.Criteria[1].Feedback != "Stress in phytoplankton." || evaluation.ScoreConfidence != "medium" {
		t.Errorf("rating not used: %+v, confidence %s", evaluation.Criteria[1], evaluation.ScoreConfidence)
	}
	if !evaluation.Words[1].Unclear || evaluation.Words[0].Unclear {
		t.Errorf("unclear marks = %+v", evaluation.Words[:2])
	}
}

func TestVerbatimEvaluationPrefersMeasuredFluency(t *testing.T) {
	rating := &ai.DeliveryRating{Pronunciation: 58, PronunciationFeedback: "Clear.", Fluency: 58, FluencyFeedback: "Uneven.", Unclear: nil}
	smooth := scoring.Delivery{DurationSeconds: 4, WordsSpoken: 8, SpeakingRatio: 0.9}
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, smooth, rating)

	// No pauses (5 of 5) beats the model's 58, and every word heard clearly is 90.
	if evaluation.Criteria[1].Score != 90 || evaluation.Criteria[2].Score != 90 || *evaluation.EstimatedScore != 90 {
		t.Errorf("criteria %+v, estimate %.0f, want 90 across", evaluation.Criteria, *evaluation.EstimatedScore)
	}

	halting := scoring.Delivery{DurationSeconds: 4, WordsSpoken: 8, SpeakingRatio: 0.9, PauseCount: 4}
	if got := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, halting, rating).Criteria[2].Score; got != 34 {
		t.Errorf("fluency with four pauses = %.0f, want 34", got)
	}
}

func TestScoreVerbatimPlacesPauses(t *testing.T) {
	r := scoring.ScoreVerbatim(passage, "every flat map"+scoring.PauseMark+"distorts the globe it represents")
	for i, w := range r.Marks {
		if w.PauseAfter != (i == 2) {
			t.Errorf("word %d %q pauseAfter = %v", i, w.Text, w.PauseAfter)
		}
	}
}
