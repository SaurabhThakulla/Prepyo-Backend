package ai

import (
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// A scored IELTS practice response requires evidence for all four criteria.
// A null score remains valid for silence or insufficient evidence.
func validateIELTSCriteria(criteria []models.EvaluationCriterion, spec feedbackSpec, score float64) error {
	required := map[string]bool{"lexical resource": false, "grammatical range and accuracy": false}
	if spec.Skill == models.SkillSpeaking {
		required["fluency and coherence"] = false
		required["pronunciation"] = false
	} else {
		required["coherence and cohesion"] = false
		task := "task response"
		name := strings.ToLower(spec.TaskName)
		if strings.Contains(name, "figure") || strings.Contains(name, "task 1") {
			task = "task achievement"
		}
		required[task] = false
	}
	if len(criteria) != 4 {
		return fmt.Errorf("IELTS score requires exactly four assessment criteria")
	}
	sum := 0.0
	for _, c := range criteria {
		key := strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(c.Name, "&", "and")), " "))
		seen, known := required[key]
		if !known || seen || c.MaxScore != 9 || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) || c.Score < 0 || c.Score > 9 || strings.TrimSpace(c.Feedback) == "" {
			return fmt.Errorf("invalid or duplicate IELTS criterion %q", c.Name)
		}
		required[key] = true
		sum += c.Score
	}
	if scoring.RoundIELTSBand(sum/4) != score {
		return fmt.Errorf("IELTS estimate must equal the rounded mean of four equally weighted criteria")
	}
	return nil
}
