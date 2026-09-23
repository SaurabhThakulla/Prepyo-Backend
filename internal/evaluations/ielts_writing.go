package evaluations

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
)

const ieltsShortWritingVersion = "ielts-short-writing.v1"

// ieltsBand1WordLimit is the public descriptors' rule: responses of 20 words
// or fewer are rated at Band 1 on every criterion.
const ieltsBand1WordLimit = 20

// copiedRunLength is the shortest run of consecutive words, shared with the
// task prompt, that counts as copied rubric rather than coincidence.
const copiedRunLength = 5

// ieltsWords splits text into words the way IELTS counts them: a hyphenated
// word such as "check-in" is one word, and punctuation around a word is not a
// word of its own. Words are lowercased for comparison.
func ieltsWords(text string) []string {
	fields := strings.Fields(text)
	words := make([]string, 0, len(fields))
	for _, field := range fields {
		word := strings.TrimFunc(field, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		if word != "" {
			words = append(words, strings.ToLower(word))
		}
	}
	return words
}

// IELTSWordCount counts the learner's own words: every word of the response
// except those inside a run of copiedRunLength or more consecutive words that
// also appears in the task prompt. The descriptors say copied rubric must be
// discounted, and a learner who pastes the question into their essay has not
// written those words.
func IELTSWordCount(response, prompt string) int {
	words := ieltsWords(response)
	promptWords := ieltsWords(prompt)
	if len(words) < copiedRunLength || len(promptWords) < copiedRunLength {
		return len(words)
	}

	grams := make(map[string]bool, len(promptWords))
	for i := 0; i+copiedRunLength <= len(promptWords); i++ {
		grams[strings.Join(promptWords[i:i+copiedRunLength], " ")] = true
	}

	copied := make([]bool, len(words))
	for i := 0; i+copiedRunLength <= len(words); i++ {
		if grams[strings.Join(words[i:i+copiedRunLength], " ")] {
			for j := i; j < i+copiedRunLength; j++ {
				copied[j] = true
			}
		}
	}

	count := 0
	for _, isCopied := range copied {
		if !isCopied {
			count++
		}
	}
	return count
}

// ieltsShortWritingEvaluation applies the Band 1 rule without a model call:
// the outcome is fixed by the descriptors, so asking a model could only add
// variation to it.
func ieltsShortWritingEvaluation(question models.Question, words int) models.Evaluation {
	first := "Task Response"
	if ai.IELTSMinimumWords(question.TypeID, question.TypeName) < 250 {
		first = "Task Achievement"
	}
	minimum := ai.IELTSMinimumWords(question.TypeID, question.TypeName)

	band := 1.0
	feedback := "IELTS rates responses of 20 words or fewer at Band 1 on every criterion."
	criteria := make([]models.EvaluationCriterion, 0, 4)
	for _, name := range []string{first, "Coherence and Cohesion", "Lexical Resource", "Grammatical Range and Accuracy"} {
		criteria = append(criteria, models.EvaluationCriterion{Name: name, Score: band, MaxScore: 9, Feedback: feedback})
	}

	counted := fmt.Sprintf("Your response has %d %s of your own", words, pluralWord(words))
	return models.Evaluation{
		Exam:              models.ExamIELTS,
		Skill:             models.SkillWriting,
		EvaluationVersion: ai.EvaluationVersion,
		EstimatedScore:    &band,
		ScoreConfidence:   "high",
		Summary: counted + " once wording copied from the task is set aside. " +
			"IELTS rates responses of 20 words or fewer at Band 1 on every criterion, so this estimate is Band 1. " +
			fmt.Sprintf("This task asks for at least %d words.", minimum),
		Criteria:         criteria,
		Strengths:        []string{},
		Weaknesses:       []string{fmt.Sprintf("Write a complete response of at least %d words in your own words.", minimum)},
		SentenceFeedback: []models.SentenceFeedback{},
	}
}
