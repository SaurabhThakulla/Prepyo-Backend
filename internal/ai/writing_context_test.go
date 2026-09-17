package ai

import (
	"strings"
	"testing"
)

func TestWritingSourcePassageSeparatedFromResponse(t *testing.T) {
	req := WritingRequest{TaskName: "Summarize Written Text", Prompt: "Write one sentence.", ContextPassage: "Cities retain heat in dense surfaces.", LearnerText: "Reflective roofs help cool cities."}
	got := writingUserPrompt(req)
	for _, text := range []string{req.Prompt, req.ContextPassage, req.LearnerText, "Source passage (reference material, not the learner's response):"} {
		if !strings.Contains(got, text) {
			t.Fatalf("missing %q from %s", text, got)
		}
	}
	if strings.Index(got, req.ContextPassage) >= strings.Index(got, "Learner's response:") {
		t.Fatal("passage mixed into learner response")
	}
	req.ContextPassage = " \n\t"
	if strings.Contains(writingUserPrompt(req), "Source passage") {
		t.Fatal("empty passage included")
	}
}
