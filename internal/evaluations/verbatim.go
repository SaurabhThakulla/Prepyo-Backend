package evaluations

import (
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// verbatimScoringVersion is stamped on evaluations produced by alignment rather
// than by a model, so stored feedback stays traceable to how it was made.
const verbatimScoringVersion = "verbatim.v1"

// contentWeight is how much of the estimate the words carry on a task where the
// words were given. Delivery is the rest. Pearson weights content heavily on
// Read Aloud for the same reason: a learner who skips half the passage has not
// done the task, however smoothly they said the other half.
const contentWeight = 0.7

// verbatimEvaluation builds feedback for a task with one right answer, without
// asking a model anything.
//
// Two criteria, both measured: content from the alignment, oral fluency from
// what the browser timed. When the browser measured nothing, fluency is left
// out entirely and the estimate rests on content alone - an unmeasured criterion
// is not a criterion, and guessing at one is what this path exists to avoid.
func verbatimEvaluation(
	question models.Question,
	version models.ExamVersion,
	expected, spoken string,
	delivery scoring.Delivery,
) models.Evaluation {
	alignment := scoring.ScoreVerbatim(expected, spoken)
	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}

	criteria := []models.EvaluationCriterion{{
		Name:     "Content",
		Score:    scale.EstimateFromAccuracy(alignment.Accuracy),
		MaxScore: version.MaxScore,
		Feedback: alignment.Feedback(),
	}}

	accuracy := alignment.Accuracy
	if delivery.Measured() {
		fluency := delivery.FluencyAccuracy()
		criteria = append(criteria, models.EvaluationCriterion{
			Name:     "Oral Fluency",
			Score:    scale.EstimateFromAccuracy(fluency),
			MaxScore: version.MaxScore,
			Feedback: delivery.Feedback(),
		})
		accuracy = contentWeight*alignment.Accuracy + (1-contentWeight)*fluency
	}

	estimate := scale.EstimateFromAccuracy(accuracy)

	return models.Evaluation{
		Exam:              question.Exam,
		Skill:             models.SkillSpeaking,
		EvaluationVersion: ai.EvaluationVersion,
		EstimatedScore:    &estimate,
		// Measured, not estimated by a model - but still read from a device
		// transcript, which has its own errors, so never "high".
		ScoreConfidence:  "medium",
		Summary:          verbatimSummary(question, alignment, delivery),
		Criteria:         criteria,
		Strengths:        verbatimStrengths(alignment, delivery),
		Weaknesses:       verbatimWeaknesses(alignment, delivery),
		SentenceFeedback: []models.SentenceFeedback{},
		Transcript:       spoken,
	}
}

func verbatimSummary(question models.Question, alignment scoring.VerbatimResult, delivery scoring.Delivery) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is marked against the exact words you were given. ", question.TypeName)
	fmt.Fprintf(&b, "You said %d of the %d words in order.", alignment.Matched, alignment.Expected)

	if delivery.Measured() {
		fmt.Fprintf(&b, " %s", delivery.Feedback())
	}

	b.WriteString(" Pronunciation was not assessed: this score comes from the words your device heard, not from the recording itself.")
	return b.String()
}

func verbatimStrengths(alignment scoring.VerbatimResult, delivery scoring.Delivery) []string {
	strengths := []string{}

	switch {
	case alignment.Accuracy >= 0.95:
		strengths = append(strengths, "You covered the text almost word for word")
	case alignment.Accuracy >= 0.8:
		strengths = append(strengths, "You kept most of the text and its order")
	}

	if delivery.Measured() && delivery.FluencyAccuracy() >= 0.8 {
		strengths = append(strengths, "Your pace stayed steady, without long hesitations")
	}
	return strengths
}

func verbatimWeaknesses(alignment scoring.VerbatimResult, delivery scoring.Delivery) []string {
	weaknesses := []string{}

	if len(alignment.Missed) > 0 {
		weaknesses = append(weaknesses,
			"Words missed or changed: "+strings.Join(alignment.Missed, ", "))
	}
	if len(alignment.Added) > 0 {
		weaknesses = append(weaknesses,
			"Words added that were not in the text: "+strings.Join(alignment.Added, ", "))
	}

	if delivery.Measured() {
		if delivery.PauseCount > 3 {
			weaknesses = append(weaknesses,
				fmt.Sprintf("You paused %d times mid-answer, which breaks the flow an examiner listens for", delivery.PauseCount))
		}
		if rate := delivery.WordsPerMinute(); rate > 0 && rate < 100 {
			weaknesses = append(weaknesses,
				fmt.Sprintf("At about %d words a minute you were slower than the exam expects", rate))
		}
	}
	return weaknesses
}
