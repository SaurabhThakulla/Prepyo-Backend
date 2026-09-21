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

func TestVerbatimEvaluationScoresOnTheExamScale(t *testing.T) {
	delivery := scoring.Delivery{
		DurationSeconds: 20, WordsSpoken: 8, PauseCount: 1, LongestPauseSeconds: 0.8, SpeakingRatio: 0.8,
	}
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, delivery)

	if evaluation.EstimatedScore == nil {
		t.Fatal("a perfect read produced no score")
	}
	if *evaluation.EstimatedScore < 80 || *evaluation.EstimatedScore > 90 {
		t.Errorf("perfect read scored %.0f on the 10-90 scale", *evaluation.EstimatedScore)
	}
	if len(evaluation.Criteria) != 2 {
		t.Fatalf("criteria = %d, want Content and Oral Fluency", len(evaluation.Criteria))
	}
	// Pronunciation is the one thing a transcript cannot show, and inventing a
	// criterion for it is what this whole path exists to avoid.
	for _, c := range evaluation.Criteria {
		if strings.Contains(strings.ToLower(c.Name), "pronunciation") {
			t.Errorf("verbatim scoring returned a %q criterion", c.Name)
		}
	}
	if !strings.Contains(evaluation.Summary, "Pronunciation was not assessed") {
		t.Errorf("summary does not say pronunciation went unassessed: %q", evaluation.Summary)
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

func TestVerbatimEvaluationWithoutMeasurementsJudgesContentOnly(t *testing.T) {
	// A browser that measured nothing must not produce a fluency mark: an
	// unmeasured criterion is a guess wearing a number.
	evaluation := verbatimEvaluation(pteQuestion(), pteVersion(), passage, passage, scoring.Delivery{})

	if len(evaluation.Criteria) != 1 || evaluation.Criteria[0].Name != "Content" {
		t.Fatalf("criteria = %+v, want Content alone", evaluation.Criteria)
	}
}
