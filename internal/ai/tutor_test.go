package ai

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

// The coach is an exam tutor: it must turn down requests outside exam
// preparation, such as writing code, instead of answering them.
func TestTutorPromptKeepsTheCoachOnTopic(t *testing.T) {
	prompt := tutorSystemPrompt(TutorRequest{Exam: models.ExamPTE})
	for _, want := range []string{
		"You help only with preparing for PTE",
		"programming code",
		"do not answer it, not even in part",
		"you can only help with PTE preparation",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("tutor prompt is missing %q", want)
		}
	}
}
