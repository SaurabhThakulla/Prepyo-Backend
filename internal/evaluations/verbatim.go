package evaluations

import (
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// verbatimScoringVersion is stamped on evaluations produced by alignment rather
// than by a model, so stored feedback stays traceable to how it was made.
const verbatimScoringVersion = "verbatim.v2"

// assumedPoints stands in for a trait Prepyo could not measure: the middle of Pearson's 0-5 scale.
const assumedPoints = 3

// verbatimEvaluation marks Read Aloud and Repeat Sentence on Pearson's three
// traits (Score Guide July 2025): Content from the alignment, Oral Fluency from
// the measured pauses and pace, Pronunciation assumed, never perfect.
func verbatimEvaluation(
	question models.Question,
	version models.ExamVersion,
	expected, spoken string,
	delivery scoring.Delivery,
) models.Evaluation {
	alignment := scoring.ScoreVerbatim(expected, spoken)
	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}
	content, contentMax := verbatimContentPoints(question.TypeID, alignment)

	fluency, fluencyNote := assumedPoints, "Your pauses and pace could not be measured on this recording, so this counts as 3 of 5."
	if delivery.Measured() {
		fluency, fluencyNote = delivery.FluencyPoints(), delivery.Feedback()+" "+fluencyTip(delivery)
	}
	pronunciation := assumedPoints
	// Pearson gives no points at all when Content is 0.
	if content == 0 {
		fluency, pronunciation = 0, 0
	}

	criteria := []models.EvaluationCriterion{
		{Name: "Content", Score: float64(content), MaxScore: float64(contentMax), Feedback: alignment.Feedback()},
		{Name: "Pronunciation", Score: float64(pronunciation), MaxScore: 5, Feedback: pronunciationNote(alignment)},
		{Name: "Oral Fluency", Score: float64(fluency), MaxScore: 5, Feedback: fluencyNote},
	}
	estimate := scale.EstimateFromAccuracy(float64(content+pronunciation+fluency) / float64(contentMax+10))

	return models.Evaluation{
		Exam:              question.Exam,
		Skill:             models.SkillSpeaking,
		EvaluationVersion: ai.EvaluationVersion,
		EstimatedScore:    &estimate,
		// Pronunciation is assumed, so the estimate is never better than low confidence.
		ScoreConfidence:  "low",
		Summary:          verbatimSummary(question, alignment, delivery),
		Criteria:         criteria,
		Strengths:        verbatimStrengths(alignment, delivery),
		Weaknesses:       verbatimWeaknesses(alignment, delivery),
		SentenceFeedback: []models.SentenceFeedback{},
		Transcript:       strings.Join(strings.Fields(strings.ReplaceAll(spoken, scoring.PauseMark, " ")), " "),
		Words:            alignment.Marks,
	}
}

// verbatimContentPoints is Content on Pearson's scale for the task.
func verbatimContentPoints(typeID string, a scoring.VerbatimResult) (points, outOf int) {
	if a.Expected == 0 {
		return 0, 5
	}
	if strings.Contains(typeID, "repeat-sentence") {
		// Repeat Sentence: 3 every word in order, 2 at least half, 1 less than half, 0 almost nothing.
		switch {
		case a.Matched == a.Expected:
			return 3, 3
		case a.Accuracy >= 0.5:
			return 2, 3
		case a.Accuracy >= 0.1:
			return 1, 3
		}
		return 0, 3
	}
	// Read Aloud counts errors against the text's length; Prepyo puts that share on 0-5.
	share := float64(a.Expected-a.Errors()) / float64(a.Expected)
	return int(math.Round(5 * math.Max(0, share))), 5
}

func pronunciationNote(a scoring.VerbatimResult) string {
	note := "Not measured from your recording yet, so it counts as 3 of 5, the middle of Pearson's scale."
	if len(a.Missed) > 0 {
		note += " Words heard as something else are often pronunciation slips: practise " + strings.Join(a.Missed[:min(3, len(a.Missed))], ", ") + "."
	}
	return note
}

func fluencyTip(d scoring.Delivery) string {
	switch {
	case d.PauseCount > 1:
		return "Read in phrases and keep going after a slip instead of stopping."
	case d.WordsPerMinute() > 0 && d.WordsPerMinute() < 100:
		return "Aim for a natural speaking pace, about 120-160 words a minute."
	}
	return "Keep this smooth, even rhythm."
}

func verbatimSummary(question models.Question, alignment scoring.VerbatimResult, delivery scoring.Delivery) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is marked on Content, Pronunciation and Oral Fluency, as in the exam. ", question.TypeName)
	fmt.Fprintf(&b, "You said %d of the %d words in order.", alignment.Matched, alignment.Expected)

	if delivery.Measured() {
		fmt.Fprintf(&b, " %s", delivery.Feedback())
	}

	b.WriteString(" Pronunciation was not assessed yet: it counts as 3 of 5, so this estimate is cautious.")
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

	if delivery.Measured() && delivery.FluencyPoints() >= 4 {
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
