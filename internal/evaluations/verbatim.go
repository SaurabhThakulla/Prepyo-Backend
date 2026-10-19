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
const verbatimScoringVersion = "verbatim.v3"

// assumedScore stands in for a trait Prepyo could not measure: the competent baseline of Pearson's 10-90 scale.
const assumedScore = 65.0

// verbatimEvaluation marks Read Aloud and Repeat Sentence on Pearson's three
// traits (Score Guide): Content from the alignment, Oral Fluency from
// the measured pauses and pace (or audio rating), Pronunciation from audio rating.
func verbatimEvaluation(
	question models.Question,
	version models.ExamVersion,
	expected, spoken string,
	delivery scoring.Delivery,
	rating *ai.DeliveryRating,
) models.Evaluation {
	alignment := scoring.ScoreVerbatim(expected, spoken)
	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}
	content, contentMax := verbatimContentPoints(question.TypeID, alignment)

	// Traits are shown on the exam's 10-90 scale.
	onScale := func(points, outOf int) float64 {
		if outOf <= 0 || points <= 0 {
			return version.MinScore
		}
		return math.Round(version.MinScore + (version.MaxScore-version.MinScore)*float64(points)/float64(outOf))
	}

	contentScore := onScale(content, contentMax)

	fluencyScore, fluencyNote := assumedScore, "Your pauses and pace could not be measured on this recording, so this counts as 65 of 90, a baseline estimate."
	if delivery.Measured() {
		fluencyScore, fluencyNote = onScale(delivery.FluencyPoints(), 5), delivery.Feedback()+" "+fluencyTip(delivery)
	}
	pronunciationScore, pronunciationFeedback := assumedScore, pronunciationNote(alignment)
	confidence := "low"
	if rating != nil {
		pronunciationScore = float64(rating.Pronunciation)
		pronunciationFeedback = rating.PronunciationFeedback
		fluencyScore = float64(rating.Fluency)
		fluencyNote = rating.FluencyFeedback
		confidence = "medium"
		markUnclear(alignment.Marks, rating.Unclear)
	}
	// Pearson gives no points at all when Content is 0.
	if content == 0 {
		contentScore = version.MinScore
		fluencyScore = version.MinScore
		pronunciationScore = version.MinScore
	}

	criteria := []models.EvaluationCriterion{
		{Name: "Content", Score: contentScore, MaxScore: version.MaxScore, Feedback: alignment.Feedback()},
		{Name: "Pronunciation", Score: pronunciationScore, MaxScore: version.MaxScore, Feedback: pronunciationFeedback},
		{Name: "Oral Fluency", Score: fluencyScore, MaxScore: version.MaxScore, Feedback: fluencyNote},
	}
	estimate := math.Round((contentScore + pronunciationScore + fluencyScore) / 3.0)
	if scale.Step > 0 {
		estimate = math.Round(estimate/scale.Step) * scale.Step
	}
	estimate = math.Min(math.Max(estimate, version.MinScore), version.MaxScore)

	return models.Evaluation{
		Exam:              question.Exam,
		Skill:             models.SkillSpeaking,
		EvaluationVersion: ai.EvaluationVersion,
		EstimatedScore:    &estimate,
		ScoreConfidence:   confidence,
		Summary:           verbatimSummary(question, alignment, delivery, rating != nil),
		Criteria:          criteria,
		Strengths:         verbatimStrengths(alignment, delivery),
		Weaknesses:        verbatimWeaknesses(alignment, delivery),
		SentenceFeedback:  []models.SentenceFeedback{},
		Transcript:        strings.Join(strings.Fields(strings.ReplaceAll(spoken, scoring.PauseMark, " ")), " "),
		Words:             alignment.Marks,
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
	note := "Not measured from your recording yet, so it counts as 65 of 90, a baseline estimate."
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

// markUnclear flags the given words the audio rating found hard to understand.
func markUnclear(marks []models.WordMark, unclear []string) {
	want := map[string]bool{}
	for _, w := range unclear {
		want[strings.ToLower(strings.Trim(w, ".,;:!?\"'()"))] = true
	}
	for i, m := range marks {
		if m.Said && want[strings.ToLower(strings.Trim(m.Text, ".,;:!?\"'()"))] {
			marks[i].Unclear = true
		}
	}
}

func verbatimSummary(question models.Question, alignment scoring.VerbatimResult, delivery scoring.Delivery, heard bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is marked on Content, Pronunciation and Oral Fluency, as in the exam. ", question.TypeName)
	fmt.Fprintf(&b, "You said %d of the %d words in order.", alignment.Matched, alignment.Expected)

	if delivery.Measured() && !heard {
		fmt.Fprintf(&b, " %s", delivery.Feedback())
	}

	if heard {
		b.WriteString(" Pronunciation and Oral Fluency were rated from your recording.")
	} else {
		b.WriteString(" Pronunciation was not assessed yet: it counts as 65 of 90, so this estimate is cautious.")
	}
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
