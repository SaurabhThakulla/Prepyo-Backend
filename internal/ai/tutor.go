package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

const TutorPromptVersion = "tutor.v1"

// maxTutorHistory caps how much conversation is sent upstream. Only the recent
// turns matter for a follow-up question, and a shorter prompt is cheaper and
// leaks less of the learner's history to the provider.
const maxTutorHistory = 10

type TutorMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TutorRequest struct {
	Exam     models.ExamType
	Messages []TutorMessage
	// TaskContext is the question the learner is looking at, if any. Only the
	// prompt and task name are passed, never the answer key.
	TaskContext string
}

func (g *Gateway) Tutor(ctx context.Context, req TutorRequest) (string, Usage, error) {
	messages := []chatMessage{{Role: "system", Content: tutorSystemPrompt(req)}}

	history := req.Messages
	if len(history) > maxTutorHistory {
		history = history[len(history)-maxTutorHistory:]
	}
	for _, m := range history {
		role := m.Role
		// Only these two roles come from the client; anything else could be an
		// attempt to inject a new system instruction.
		if role != "user" && role != "assistant" {
			continue
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		messages = append(messages, chatMessage{Role: role, Content: m.Content})
	}

	reply, usage, err := g.complete(ctx, g.models.Tutoring, TutorPromptVersion, messages, false)
	if err != nil {
		return "", Usage{}, err
	}
	if strings.TrimSpace(reply) == "" {
		return "", usage, ErrBadOutput
	}
	return reply, usage, nil
}

func tutorSystemPrompt(req TutorRequest) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(
		"You are a %s preparation tutor helping a learner in Nepal. Be concise, concrete and encouraging. "+
			"Use short paragraphs and examples rather than long lists.\n\n", req.Exam))
	b.WriteString("Be honest about what you do not know. Do not state official scoring weightings unless you are certain of them. " +
		"Make clear that any score you mention is a practice estimate, not an official result.\n")

	if strings.TrimSpace(req.TaskContext) != "" {
		b.WriteString("\nThe learner is currently working on this task:\n")
		b.WriteString(req.TaskContext)
		b.WriteString("\nHelp them think it through. Do not simply hand them a finished answer.\n")
	}
	return b.String()
}
