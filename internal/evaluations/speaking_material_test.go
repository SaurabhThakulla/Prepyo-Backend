package evaluations

import (
	"testing"

	"github.com/prepyo/backend/internal/models"
)

// Only word-for-word tasks may hand the grader text to repeat. A lecture, a
// discussion or a question sent that way marks a correct answer down for every
// word it does not echo back.
func TestSpeakingMaterialSeparatesRepeatingFromResponding(t *testing.T) {
	cases := []struct {
		name     string
		question models.Question
		want     speakingMaterial
	}{
		{"read aloud repeats its passage",
			models.Question{TypeID: "read-aloud", ContextPassage: "Every map distorts."},
			speakingMaterial{expected: "Every map distorts."}},
		{"repeat sentence repeats its transcript",
			models.Question{TypeID: "pte-repeat-sentence", AudioTranscript: "Collect your lab coat."},
			speakingMaterial{expected: "Collect your lab coat."}},
		{"retell responds to the lecture",
			models.Question{TypeID: "pte-retell-lecture", AudioTranscript: "Today I want to explain GPS.", ModelAnswer: "The lecture explains GPS."},
			speakingMaterial{source: "Today I want to explain GPS.", reference: "The lecture explains GPS."}},
		{"short answer responds to the question",
			models.Question{TypeID: "pte-answer-short-question", AudioTranscript: "What measures temperature?", ModelAnswer: "A thermometer."},
			speakingMaterial{source: "What measures temperature?", reference: "A thermometer."}},
		{"situation responds to the passage",
			models.Question{TypeID: "pte-respond-to-situation", ContextPassage: "Your order is wrong.", AudioTranscript: "Your order is wrong."},
			speakingMaterial{source: "Your order is wrong."}},
		{"cue card has nothing to compare",
			models.Question{TypeID: "ielts-speaking-part2"},
			speakingMaterial{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := speakingMaterialOf(c.question); got != c.want {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}
