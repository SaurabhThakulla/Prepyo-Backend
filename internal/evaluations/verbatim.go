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
const verbatimScoringVersion = "verbatim.v9"

// assumedScore stands in for a trait Prepyo could not measure: the competent baseline of Pearson's 10-90 scale.
const assumedScore = 65.0

// verbatimEvaluation marks Read Aloud and Repeat Sentence on Pearson's three
// traits (Score Guide): Content from the alignment (a word the audio rating
// heard as bad counts as not said), Pronunciation from the audio rating's
// good/average/bad word grades, Oral Fluency from the measured pauses and pace
// averaged with the rating's own fluency (which hears flat, word-by-word reading).
func verbatimEvaluation(
	question models.Question,
	version models.ExamVersion,
	expected, spoken string,
	delivery scoring.Delivery,
	rating *ai.DeliveryRating,
) models.Evaluation {
	alignment := scoring.ScoreVerbatim(expected, spoken)
	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}
	misheard := heardAsOther(alignment)
	var grades wordGrades
	if rating != nil {
		grades = applyWordGrades(&alignment, rating)
	}
	content, contentMax := verbatimContentPoints(question.TypeID, alignment)

	// Traits are shown on the exam's 10-90 scale.
	onScale := func(points, outOf int) float64 {
		if outOf <= 0 || points <= 0 {
			return version.MinScore
		}
		return math.Round(version.MinScore + (version.MaxScore-version.MinScore)*float64(points)/float64(outOf))
	}

	contentScore := onScale(content, contentMax)
	if !strings.Contains(question.TypeID, "repeat-sentence") && content > 0 {
		// Read Aloud: the straight share of words without an error, not rounded to 0-5 first.
		contentScore = math.Round(version.MinScore + (version.MaxScore-version.MinScore)*readAloudShare(alignment))
	}

	fluencyScore, fluencyNote := assumedScore, "Your pauses and pace could not be measured on this recording, so this counts as 65 of 90, a baseline estimate."
	if delivery.Measured() {
		fluencyScore, fluencyNote = delivery.FluencyScore(), delivery.Feedback()+" "+fluencyTip(delivery)
	}
	pronunciationScore, pronunciationFeedback := assumedScore, pronunciationNote(alignment)
	confidence := "low"
	if rating != nil {
		// Word grades and the overall rating are each noisy alone; the mean of both is steadier.
		pronunciationScore = float64(rating.Pronunciation)
		if score, ok := grades.scorerScore(misheard); ok && rating.WordScored {
			pronunciationScore = score // the scorer's per-sound grades stand on their own
		} else if score, ok := grades.score(misheard); ok {
			pronunciationScore = math.Round((score + pronunciationScore) / 2)
		}
		pronunciationFeedback = rating.PronunciationFeedback
		if !delivery.Measured() && rating.Fluency > 0 {
			fluencyScore = float64(rating.Fluency)
			fluencyNote = rating.FluencyFeedback
		}
		confidence = "medium"
	}
	// A changed, added or badly said word is a stumble the timings cannot see;
	// a word left out is a Content error only. On the scorer path an unclear word counts half: nothing else hears rhythm.
	stumbles := alignment.Inserted + grades.bad
	weight := float64(stumbles)
	if rating != nil && rating.WordScored {
		stumbles += grades.average
		weight += 0.5 * float64(grades.average)
	}
	if stumbles > 0 && alignment.Expected > 0 {
		kept := math.Max(fluencyKeptFloor, 1-fluencyErrorWeight*weight/float64(alignment.Expected))
		cut := math.Round(fluencyScore * (1 - kept))
		fluencyScore = math.Max(version.MinScore, fluencyScore-cut)
		fluencyNote += fmt.Sprintf(" %d words said wrongly or unclearly broke the flow (−%.0f).", stumbles, cut)
	}
	// Pauses and pace cannot hear flat, word-by-word reading; the rating can, so it may only lower the mark.
	if rating != nil && delivery.Measured() && rating.Fluency > 0 && float64(rating.Fluency) < fluencyScore {
		fluencyScore = math.Round((fluencyScore + float64(rating.Fluency)) / 2)
		fluencyNote += " " + rating.FluencyFeedback
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
	// Read Aloud counts errors against the text's length; 0-5 here only decides the zero gate.
	return int(math.Round(5 * readAloudShare(a))), 5
}

// readAloudShare is Pearson's Read Aloud content: the share of the text's
// words left after each replaced, left-out or added word counts one error.
func readAloudShare(a scoring.VerbatimResult) float64 {
	return math.Max(0, float64(a.Expected-a.Errors())/float64(a.Expected))
}

// fluencyErrorWeight scales Oral Fluency down by twice the share of words in
// error: Pearson's fluency counts hesitations, repetitions and false starts,
// which a wrong word usually is.
const fluencyErrorWeight = 2.0

// fluencyKeptFloor is the least of the fluency mark word errors leave: a
// misread passage is a Content problem first.
const fluencyKeptFloor = 0.4

// Fewest words graded for the grades to stand for Pronunciation.
const minWordsForClearShare = 3

// Share of word points (good 1, average 1/2) at which Pronunciation bottoms out
// at 10; from there it rises linearly to 90 when every word was good.
const gradeShareFloor = 0.3

// heardAsOther counts given words recognised as a different word (a missed
// word paired with an added one), usually a dropped ending or a blurred sound.
func heardAsOther(a scoring.VerbatimResult) int {
	return min(a.Expected-a.Matched, a.Inserted)
}

// maxShownAverageShare is the most of the said words shown as unclear; above it
// the rating has marked words wholesale rather than picked the weak ones.
const maxShownAverageShare = 0.4

// wordGrades counts the audio rating's grades over the words that were said.
type wordGrades struct{ good, average, bad int }

// applyWordGrades marks average words unclear and excludes severely mispronounced
// words from what was said, so they count against Content.
func applyWordGrades(a *scoring.VerbatimResult, r *ai.DeliveryRating) wordGrades {
	var g wordGrades
	grade := map[int]string{}
	for _, n := range r.Average {
		grade[n-1] = "average"
	}
	for _, n := range r.Bad {
		grade[n-1] = "bad"
	}
	for i, m := range a.Marks {
		if !m.Said {
			continue
		}
		switch grade[i] {
		case "bad":
			g.bad++
			a.Marks[i].Said = false
			a.Matched = max(0, a.Matched-1)
			if len(a.Missed) < 8 {
				a.Missed = append(a.Missed, strings.Trim(m.Text, ".,;:!?\"'()"))
			}
		case "average":
			g.average++
			a.Marks[i].Unclear = true
		default:
			g.good++
		}
	}
	// Grading nearly every word average singles nothing out: keep it in the score, not the colours.
	if said := g.good + g.average; said > 0 && float64(g.average) > maxShownAverageShare*float64(said) {
		for i := range a.Marks {
			a.Marks[i].Unclear = false
		}
	}
	if a.Expected > 0 {
		a.Accuracy = float64(a.Matched) / float64(a.Expected)
	}
	return g
}

// score is Pronunciation on 10-90 from the grades: good 1, average 1/2, bad 0,
// and a word heard as another word 0. Words simply left out stay with Content.
func (g wordGrades) score(misheard int) (float64, bool) {
	words := g.good + g.average + g.bad + misheard
	if words < minWordsForClearShare {
		return 0, false
	}
	share := (float64(g.good) + 0.5*float64(g.average)) / float64(words)
	return math.Round(10 + 80*math.Min(1, math.Max(0, (share-gradeShareFloor)/(1-gradeShareFloor)))), true
}

// scorerGoodFloor is the share of good words at which the scorer's Pronunciation bottoms out at 10.
// The scorer accepts weak forms and accented vowels, so only "good" words earn credit.
const scorerGoodFloor = 0.4

// scorerScore is Pronunciation on 10-90 from the pronunciation scorer's grades: the share of good words.
func (g wordGrades) scorerScore(misheard int) (float64, bool) {
	words := g.good + g.average + g.bad + misheard
	if words < minWordsForClearShare {
		return 0, false
	}
	share := float64(g.good) / float64(words)
	return math.Round(10 + 80*math.Min(1, math.Max(0, (share-scorerGoodFloor)/(1-scorerGoodFloor)))), true
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
