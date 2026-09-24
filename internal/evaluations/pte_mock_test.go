package evaluations

import "testing"

func TestShortAnswerTasksAreRatedNotRefused(t *testing.T) {
	for _, id := range []string{"pte-answer-short-question", "answer-short-question"} {
		if !expectsShortAnswer(id) {
			t.Errorf("%s answers in a word or two", id)
		}
	}
	for _, id := range []string{"pte-describe-image", "pte-retell-lecture", "read-aloud"} {
		if expectsShortAnswer(id) {
			t.Errorf("%s is not a short-answer task", id)
		}
	}
}
