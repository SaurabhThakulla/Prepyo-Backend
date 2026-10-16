package evaluations

import (
	"strings"
	"testing"

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
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, delivery)

	if evaluation.EstimatedScore == nil {
		t.Fatal("a perfect read produced no score")
	}
	// Content 5 + Pronunciation 3 (assumed) + Fluency 4 of 15 points.
	if *evaluation.EstimatedScore != 74 {
		t.Errorf("perfect read with one pause scored %.0f, want 74", *evaluation.EstimatedScore)
	}
	names := []string{}
	for _, c := range evaluation.Criteria {
		names = append(names, c.Name)
	}
	if strings.Join(names, ",") != "Content,Pronunciation,Oral Fluency" {
		t.Fatalf("criteria = %v", names)
	}
	if evaluation.Criteria[1].Score != 3 || evaluation.Criteria[1].MaxScore != 5 {
		t.Errorf("pronunciation = %+v, want the assumed 3 of 5", evaluation.Criteria[1])
	}
	if !strings.Contains(evaluation.Summary, "Pronunciation was not assessed") {
		t.Errorf("summary does not say pronunciation went unassessed: %q", evaluation.Summary)
	}
	if len(evaluation.Words) != len(strings.Fields(passage)) {
		t.Errorf("words = %d, want one mark per word of the passage", len(evaluation.Words))
	}
}

func TestVerbatimEvaluationMarksDownAMissedRead(t *testing.T) {
	full := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, scoring.Delivery{})
	partial := verbatimEvaluation(pteQuestion(), pteVersion(), passage, "every flat map", scoring.Delivery{})

	if *partial.EstimatedScore >= *full.EstimatedScore {
		t.Errorf("half a passage scored %.0f against %.0f for the whole thing",
			*partial.EstimatedScore, *full.EstimatedScore)
	}
	if len(partial.Weaknesses) == 0 {
		t.Error("no weakness listed for a passage that was mostly skipped")
	}
}

func TestVerbatimEvaluationNothingSaidScoresTheFloor(t *testing.T) {
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, "", scoring.Delivery{})
	if *evaluation.EstimatedScore != 10 {
		t.Errorf("an empty read scored %.0f, want 10", *evaluation.EstimatedScore)
	}
}

// The recording behind the 81 that APEUni-style marking should bring down.
func TestVerbatimEvaluationPhytoplanktonRecording(t *testing.T) {
	given := "Marine phytoplankton play an essential role in regulating Earth's climate by absorbing carbon dioxide through photosynthesis. When these microscopic organisms die, a significant portion of the organic carbon sinks into the deep ocean, sequestering greenhouse gases for centuries."
	heard := "Marine psychopathians play an essential role in regulating Earth's climate by absorbing carbon dioxide through photosynthesis. When this microscopic organism dies, a significant portion of the organic carbon sinks into the deep ocean, squirting greenhouse gases for centuries."
	delivery := scoring.Delivery{DurationSeconds: 23, WordsSpoken: 38, PauseCount: 2, LongestPauseSeconds: 0.9, SpeakingRatio: 0.8}

	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), given, heard, delivery)
	// Content 4 (5 errors in 38), Pronunciation 3, Fluency 2 (two pauses, 99 wpm): 9 of 15.
	if *evaluation.EstimatedScore != 58 {
		t.Errorf("scored %.0f, want 58 (was 81 before trait points)", *evaluation.EstimatedScore)
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

func TestScoreVerbatimPlacesPauses(t *testing.T) {
	r := scoring.ScoreVerbatim(passage, "every flat map"+scoring.PauseMark+"distorts the globe it represents")
	for i, w := range r.Marks {
		if w.PauseAfter != (i == 2) {
			t.Errorf("word %d %q pauseAfter = %v", i, w.Text, w.PauseAfter)
		}
	}
}
