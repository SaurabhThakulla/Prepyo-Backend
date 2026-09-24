package ptemock

import (
	"math"

	"github.com/prepyo/backend/internal/models"
)

// PTE Academic reports every score on a scale of 10 to 90.
const (
	MinScore = 10
	MaxScore = 90
)

// Mark is one item's mark: score out of max. An item with Scored false has no
// mark (its rating was unavailable) and is left out, not counted as zero. An
// item the learner skipped or ran out of time on is scored, at zero.
type Mark struct {
	Task     string
	Scored   bool
	Score    float64
	MaxScore float64
}

// fraction is the share of the marks earned, 0 to 1.
func (m Mark) fraction() float64 {
	if m.MaxScore <= 0 {
		return 0
	}
	return math.Max(0, math.Min(1, m.Score/m.MaxScore))
}

// TaskResult is how the learner did on one task across its items.
type TaskResult struct {
	Code    string                       `json:"code"`
	Name    string                       `json:"name"`
	Part    Part                         `json:"part"`
	Items   int                          `json:"items"`
	Scored  int                          `json:"scored"`
	Percent int                          `json:"percent"`
	Weights map[models.SkillType]float64 `json:"weights"`
}

// Result is a scored paper.
type Result struct {
	// Overall is the overall score: the mean of the four skill scores. Only a
	// full test has one.
	Overall *int `json:"overall"`
	// Skills are the communicative skill scores the paper can support: every
	// skill a scored item counts towards, for a full test; the section's own
	// skill for a sectional test.
	Skills map[models.SkillType]int `json:"skills"`
	Tasks  []TaskResult             `json:"tasks"`
	// Unscored counts items with no mark, which the scores leave out.
	Unscored int `json:"unscored"`
	// Answered counts items with any response at all.
	Answered int `json:"answered"`
	Items    int `json:"items"`
}

// ToScale turns a share of the marks into a 10-90 score. The mapping is linear:
// Pearson's own conversion is not published, so a Prepyo score is an estimate
// of the scale position, not a prediction of an official score.
func ToScale(fraction float64) int {
	fraction = math.Max(0, math.Min(1, fraction))
	return int(math.Round(MinScore + (MaxScore-MinScore)*fraction))
}

// Score aggregates item marks into task results and skill scores.
//
// Within a task each item counts equally (the mean of item fractions), so a
// long Fill in the Blanks does not outweigh a short one. A skill is the
// weighted mean of the tasks that count towards it (Task.Weights); a task with
// no scored item is left out and the remaining weights share its place.
func Score(blueprint Blueprint, marks []Mark, answered int) Result {
	type acc struct {
		items, scored int
		sum           float64
	}
	byTask := map[string]*acc{}
	order := []string{}
	unscored := 0
	for _, m := range marks {
		a, ok := byTask[m.Task]
		if !ok {
			a = &acc{}
			byTask[m.Task] = a
			order = append(order, m.Task)
		}
		a.items++
		if !m.Scored {
			unscored++
			continue
		}
		a.scored++
		a.sum += m.fraction()
	}

	result := Result{Skills: map[models.SkillType]int{}, Unscored: unscored, Answered: answered, Items: len(marks)}
	skillWeight := map[models.SkillType]float64{}
	skillSum := map[models.SkillType]float64{}
	for _, code := range order {
		task := Tasks[code]
		a := byTask[code]
		tr := TaskResult{Code: code, Name: task.Name, Part: task.Part, Items: a.items, Scored: a.scored, Weights: task.Weights}
		if a.scored > 0 {
			mean := a.sum / float64(a.scored)
			tr.Percent = int(math.Round(mean * 100))
			for skill, weight := range task.Weights {
				skillWeight[skill] += weight
				skillSum[skill] += weight * mean
			}
		}
		result.Tasks = append(result.Tasks, tr)
	}

	for _, skill := range blueprint.Skills {
		if w := skillWeight[skill]; w > 0 {
			result.Skills[skill] = ToScale(skillSum[skill] / w)
		}
	}

	if blueprint.Kind == KindFull && len(result.Skills) > 0 {
		total := 0
		for _, s := range result.Skills {
			total += s
		}
		overall := int(math.Round(float64(total) / float64(len(result.Skills))))
		result.Overall = &overall
	}
	return result
}

// Headline is the single score a paper is recorded under: the overall for the
// full test, the section's skill score for a sectional test.
func (r Result) Headline(blueprint Blueprint) (float64, bool) {
	if r.Overall != nil {
		return float64(*r.Overall), true
	}
	if len(blueprint.Skills) == 1 {
		if s, ok := r.Skills[blueprint.Skills[0]]; ok {
			return float64(s), true
		}
	}
	return 0, false
}

// FractionOfEvaluation reads a rated answer as a share of the marks: the sum of
// its trait scores out of their maxima, which is how PTE builds an item score
// from its traits, or the estimated score's position on the scale when there
// are no traits.
func FractionOfEvaluation(e models.Evaluation) (float64, bool) {
	var score, max float64
	for _, c := range e.Criteria {
		if c.MaxScore > 0 {
			score += math.Max(0, math.Min(c.Score, c.MaxScore))
			max += c.MaxScore
		}
	}
	if max > 0 {
		return score / max, true
	}
	if e.EstimatedScore != nil {
		return math.Max(0, math.Min(1, (*e.EstimatedScore-MinScore)/(MaxScore-MinScore))), true
	}
	return 0, false
}
