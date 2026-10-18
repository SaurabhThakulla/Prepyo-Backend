package evaluations

import (
	"math"

	"github.com/prepyo/backend/internal/models"
)

// pteTraitMax is the scale every PTE trait is shown on, as the 10-90 overall score is.
const pteTraitMax = 90

// onPTEScale shows a PTE result saved with trait points (e.g. 3 of 5) on the 10-90 scale.
func onPTEScale(e models.Evaluation) models.Evaluation {
	if e.Exam != models.ExamPTE {
		return e
	}
	for i, c := range e.Criteria {
		if c.MaxScore > 0 && c.MaxScore != pteTraitMax {
			e.Criteria[i].Score = math.Round(10 + 80*math.Min(c.Score, c.MaxScore)/c.MaxScore)
			e.Criteria[i].MaxScore = pteTraitMax
		}
	}
	return e
}
