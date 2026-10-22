package ai

import (
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// A scored IELTS practice response requires evidence for all four criteria.
// A null score remains valid for silence or insufficient evidence. It returns
// the estimate worked out from the criteria, so the model's own arithmetic
// never decides the band.
func validateIELTSCriteria(criteria []models.EvaluationCriterion, spec feedbackSpec) (float64, error) {
	required := map[string]bool{"lexical resource": false, "grammatical range and accuracy": false}
	if spec.Skill == models.SkillSpeaking {
		required["fluency and coherence"] = false
		// A transcript cannot show pronunciation, so feedback built from one
		// is three criteria, not four, and must not invent the fourth.
		if !spec.WithoutPronunciation {
			required["pronunciation"] = false
		}
	} else {
		required["coherence and cohesion"] = false
		task := "task response"
		if ieltsWritingTask(spec.TypeID, spec.TaskName) != ieltsTask2 {
			task = "task achievement"
		}
		required[task] = false
	}
	wanted := len(required)
	if len(criteria) != wanted {
		return 0, fmt.Errorf("IELTS score requires exactly %d assessment criteria", wanted)
	}
	sum := 0.0
	for _, c := range criteria {
		key := strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(c.Name, "&", "and")), " "))
		seen, known := required[key]
		if !known || seen || c.MaxScore != 9 || math.IsNaN(c.Score) || math.IsInf(c.Score, 0) || c.Score < 0 || c.Score > 9 || strings.TrimSpace(c.Feedback) == "" {
			return 0, fmt.Errorf("invalid or duplicate IELTS criterion %q", c.Name)
		}
		// The descriptors define whole-band levels only: a criterion is rated
		// at the band whose positive features the response fully fits.
		if c.Score != math.Trunc(c.Score) {
			return 0, fmt.Errorf("IELTS criterion %q must be a whole band, got %.2f", c.Name, c.Score)
		}
		required[key] = true
		sum += c.Score
	}
	// IELTS publishes no per-task rule, so the criteria mean is rounded down to
	// the half band (6.75 is 6.5): the model's whole bands run generous.
	return math.Floor(2*sum/float64(wanted)) / 2, nil
}
