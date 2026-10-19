package scoring

import (
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// Summarize Spoken Text without a model, on Pearson's Content (0-4) and Form
// (0-2) traits. As in the exam, a summary with no content earns nothing at all:
// length alone never scores.

const (
	summaryContentMax = 4
	summaryFormMax    = 2
	// realWordShare is how much of a summary must be English words from the
	// lecture or everyday function words; below it, it is not a summary.
	realWordShare = 0.5
	// topicWordsForCredit is how many distinct lecture words earn one content
	// point when no listed key concept was used, so a paraphrase is not zeroed.
	topicWordsForCredit = 4
)

// gradeListeningSummary marks a summary of the lecture.
func gradeListeningSummary(q models.Question, sub models.AnswerSubmission) Result {
	text := strings.TrimSpace(sub.TextResponse)
	given := words(text)
	key := "50-70 words covering the key points: " + strings.Join(q.CorrectAnswers, ", ")
	if len(given) == 0 {
		return proportional(q, 0, summaryContentMax+summaryFormMax, "No summary submitted.", key, orPlaceholder(text), "Content")
	}

	lecture := wordSet(q.AudioTranscript)
	for _, kw := range q.CorrectAnswers {
		for _, w := range words(kw) {
			lecture[w] = true
		}
	}
	// Without the lecture's text the words cannot be checked; the key-point gate below still applies.
	if strings.TrimSpace(q.AudioTranscript) != "" && realShare(given, lecture) < realWordShare {
		return proportional(q, 0, summaryContentMax+summaryFormMax,
			"This is not a summary of the lecture: most of the words are not English words about the topic. It scores 0.",
			key, orPlaceholder(text), "Content")
	}

	lower := strings.ToLower(text)
	found := 0
	for _, kw := range q.CorrectAnswers {
		if strings.Contains(lower, strings.ToLower(kw)) {
			found++
		}
	}
	content := summaryContentMax
	if len(q.CorrectAnswers) > 0 {
		content = int(math.Round(float64(found) / float64(len(q.CorrectAnswers)) * summaryContentMax))
	}
	if content == 0 && topicWords(given, lecture) >= topicWordsForCredit {
		content = 1
	}
	if content == 0 {
		return proportional(q, 0, summaryContentMax+summaryFormMax,
			fmt.Sprintf("None of the lecture's key points are in your summary (%s), so it scores 0, whatever its length.",
				strings.Join(q.CorrectAnswers, ", ")),
			key, orPlaceholder(text), "Content")
	}

	form := summaryForm(len(given))
	feedback := fmt.Sprintf("Content %d/%d: you covered %d of %d key points. Form %d/%d: %d words (aim for 50-70).",
		content, summaryContentMax, found, len(q.CorrectAnswers), form, summaryFormMax, len(given))
	return proportional(q, content+form, summaryContentMax+summaryFormMax, feedback, key, orPlaceholder(text), "Content & Form")
}

// summaryForm is Pearson's Form for this task: 50-70 words 2, 40-49 or 71-100 words 1, otherwise 0.
func summaryForm(n int) int {
	switch {
	case n >= 50 && n <= 70:
		return 2
	case n >= 40 && n <= 100:
		return 1
	}
	return 0
}

// realShare is the share of words of three letters or more that are lecture words or everyday English.
func realShare(given []string, lecture map[string]bool) float64 {
	counted, real := 0, 0
	for _, w := range given {
		if len([]rune(w)) < 3 {
			continue
		}
		counted++
		if lecture[w] || commonWords[w] {
			real++
		}
	}
	if counted == 0 {
		return 0
	}
	return float64(real) / float64(counted)
}

// topicWords counts distinct lecture words of four letters or more that are not everyday English.
func topicWords(given []string, lecture map[string]bool) int {
	seen := map[string]bool{}
	for _, w := range given {
		if len([]rune(w)) >= 4 && lecture[w] && !commonWords[w] {
			seen[w] = true
		}
	}
	return len(seen)
}

func wordSet(s string) map[string]bool {
	set := map[string]bool{}
	for _, w := range words(s) {
		set[w] = true
	}
	return set
}

// commonWords are everyday English words a summary uses whatever the topic.
var commonWords = func() map[string]bool {
	set := map[string]bool{}
	for _, w := range strings.Fields(`the and for are but not you all any can had her was one our out has his how its may new now old see two way who did get him let say she too use
		that with have this will your from they know want been good much some time very when come here just like long make many more only over such take than them well were
		what which their there these those would could should about after again also because before being between both during each first into most other same then through
		under until while where why many people often used uses using called known made make makes said says shows show showed explains explained describes described lecture
		speaker talk talks talked discusses discussed summary main idea ideas point points important example examples however therefore although since thus also finally
		later today years year century centuries part parts way ways help helps helped different early modern still even became become becomes common commonly started start`) {
		set[w] = true
	}
	return set
}()
