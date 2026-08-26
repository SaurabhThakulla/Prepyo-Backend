// Package scoring turns a learner's answer into a result.
//
// Everything here is deterministic and runs on the server. The client sends
// what the learner typed or selected and nothing else, so a score cannot be
// forged by editing a request.
//
// Skills that need judgement (speaking, writing) are not scored here; they go
// to the AI gateway for qualitative feedback and an explicitly estimated score.
package scoring

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// Result is the outcome of one graded answer.
type Result struct {
	IsCorrect          bool
	Score              float64
	MaxScore           float64
	AccuracyPercentage int
	Feedback           string
	CorrectDisplay     string
	UserDisplay        string
	// ErrorTag groups mistakes in the mistake bank. Empty when correct.
	ErrorTag string
}

// Deterministic reports whether a skill can be graded without a model.
func Deterministic(skill models.SkillType) bool {
	return skill == models.SkillReading || skill == models.SkillListening
}

// Grade evaluates a submission against a question.
//
// An unrecognised task type returns ok=false rather than a default pass. The
// caller turns that into an error, because silently awarding full marks for a
// task nobody wrote a grader for is worse than failing loudly.
func Grade(q models.Question, sub models.AnswerSubmission) (Result, bool) {
	switch {
	case len(q.Blanks) > 0:
		return gradeBlanks(q, sub), true
	case q.TypeID == "reorder-paragraphs":
		return gradeReorder(q, sub), true
	case q.TypeID == "write-from-dictation":
		return gradeDictation(q, sub), true
	case len(q.CorrectAnswers) > 0 && len(q.Options) > 0:
		return gradeChoice(q, sub), true
	case len(q.CorrectAnswers) > 0:
		return gradeOrderedAnswers(q, sub), true
	default:
		return Result{}, false
	}
}

// gradeBlanks awards one share of the marks per correctly filled blank.
func gradeBlanks(q models.Question, sub models.AnswerSubmission) Result {
	correct := 0
	var correctParts, userParts []string

	for _, blank := range q.Blanks {
		given := normalise(sub.BlankResponses[blank.ID])
		want := normalise(blank.CorrectAnswer)
		if given != "" && given == want {
			correct++
		}
		correctParts = append(correctParts, fmt.Sprintf("%s: %s", blank.ID, blank.CorrectAnswer))
		userParts = append(userParts, fmt.Sprintf("%s: %s", blank.ID, orPlaceholder(sub.BlankResponses[blank.ID])))
	}

	return proportional(q, correct, len(q.Blanks),
		fmt.Sprintf("You filled %d of %d blanks correctly.", correct, len(q.Blanks)),
		strings.Join(correctParts, ", "), strings.Join(userParts, ", "), "Vocabulary")
}

// gradeReorder scores adjacent pairs, which is how PTE marks this task: getting
// two neighbours in the right order earns a mark even if the whole sequence is
// not perfect.
func gradeReorder(q models.Question, sub models.AnswerSubmission) Result {
	want := q.CorrectAnswers
	got := sub.SelectedOptions

	position := make(map[string]int, len(got))
	for i, id := range got {
		position[id] = i
	}

	pairs := 0
	totalPairs := len(want) - 1
	for i := 0; i < totalPairs; i++ {
		first, firstOK := position[want[i]]
		second, secondOK := position[want[i+1]]
		if firstOK && secondOK && second == first+1 {
			pairs++
		}
	}

	return proportional(q, pairs, totalPairs,
		fmt.Sprintf("You placed %d of %d adjacent pairs correctly.", pairs, totalPairs),
		strings.Join(want, " → "), orPlaceholder(strings.Join(got, " → ")), "Sequencing")
}

// gradeDictation compares the words the learner typed against the target
// sentence. Each target word can only be matched once, so repeating a word does
// not earn extra marks.
func gradeDictation(q models.Question, sub models.AnswerSubmission) Result {
	if len(q.CorrectAnswers) == 0 {
		return Result{}
	}
	target := words(q.CorrectAnswers[0])
	given := words(sub.TextResponse)

	remaining := make(map[string]int, len(given))
	for _, w := range given {
		remaining[w]++
	}

	matched := 0
	for _, w := range target {
		if remaining[w] > 0 {
			remaining[w]--
			matched++
		}
	}

	return proportional(q, matched, len(target),
		fmt.Sprintf("You matched %d of %d words.", matched, len(target)),
		q.CorrectAnswers[0], orPlaceholder(sub.TextResponse), "Dictation accuracy")
}

// gradeChoice handles single and multiple answer questions. Extra selections
// cancel out correct ones, so guessing everything scores zero rather than full
// marks.
func gradeChoice(q models.Question, sub models.AnswerSubmission) Result {
	want := make(map[string]bool, len(q.CorrectAnswers))
	for _, id := range q.CorrectAnswers {
		want[normalise(id)] = true
	}

	hits, wrong := 0, 0
	seen := map[string]bool{}
	for _, id := range sub.SelectedOptions {
		key := normalise(id)
		if seen[key] {
			continue
		}
		seen[key] = true
		if want[key] {
			hits++
		} else {
			wrong++
		}
	}

	net := hits - wrong
	if net < 0 {
		net = 0
	}

	return proportional(q, net, len(q.CorrectAnswers),
		fmt.Sprintf("You selected %d of %d correct options.", hits, len(q.CorrectAnswers)),
		strings.Join(q.CorrectAnswers, ", "), orPlaceholder(strings.Join(sub.SelectedOptions, ", ")), "Comprehension")
}

// gradeOrderedAnswers compares answers position by position, used for
// True/False/Not Given sets where each statement has its own answer.
func gradeOrderedAnswers(q models.Question, sub models.AnswerSubmission) Result {
	correct := 0
	for i, want := range q.CorrectAnswers {
		if i < len(sub.SelectedOptions) && normalise(sub.SelectedOptions[i]) == normalise(want) {
			correct++
		}
	}

	return proportional(q, correct, len(q.CorrectAnswers),
		fmt.Sprintf("You answered %d of %d statements correctly.", correct, len(q.CorrectAnswers)),
		strings.Join(q.CorrectAnswers, ", "), orPlaceholder(strings.Join(sub.SelectedOptions, ", ")), "Inference")
}

// proportional builds the result shared by every grader: marks scale with the
// share of sub-items answered correctly.
func proportional(q models.Question, correct, total int, feedback, correctDisplay, userDisplay, errorTag string) Result {
	maxScore := float64(q.Points)
	if total <= 0 {
		return Result{MaxScore: maxScore, Feedback: "This question has no answer key.", CorrectDisplay: correctDisplay, UserDisplay: userDisplay}
	}

	ratio := float64(correct) / float64(total)
	isCorrect := correct == total

	result := Result{
		IsCorrect:          isCorrect,
		Score:              math.Round(ratio*maxScore*100) / 100,
		MaxScore:           maxScore,
		AccuracyPercentage: int(math.Round(ratio * 100)),
		Feedback:           feedback,
		CorrectDisplay:     correctDisplay,
		UserDisplay:        userDisplay,
	}
	if !isCorrect {
		result.ErrorTag = errorTag
	}
	return result
}

var nonWord = regexp.MustCompile(`[^\p{L}\p{N}\s']+`)

// normalise lowercases and trims so "Debunked " and "debunked" match.
func normalise(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// words splits a sentence into comparable lowercase words, dropping
// punctuation but keeping apostrophes so "don't" stays one word.
func words(s string) []string {
	cleaned := nonWord.ReplaceAllString(strings.ToLower(s), " ")
	return strings.Fields(cleaned)
}

func orPlaceholder(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(no answer)"
	}
	return s
}
