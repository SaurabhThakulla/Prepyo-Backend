package ai

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

// The prompts show the reply's shape without an example score: an example
// band pulls every rating towards itself.
func TestWritingSystemPromptHasNoExampleScore(t *testing.T) {
	for _, exam := range []models.ExamType{models.ExamPTE, models.ExamIELTS} {
		prompt := writingSystemPrompt(WritingRequest{Exam: exam, MinScore: 0, MaxScore: 90})
		if exampleValue.MatchString(prompt) {
			t.Errorf("%s prompt carries a numeric example score:\n%s", exam, prompt)
		}
	}
}

var exampleValue = regexp.MustCompile(`"value": ?[0-9]`)

// A PTE reply is marked on Pearson's trait points; one without them is asked
// for again rather than trusted.
func TestParseWritingRequiresPearsonTraits(t *testing.T) {
	req := WritingRequest{Exam: models.ExamPTE, MinScore: 10, MaxScore: 90, LearnerText: "Cities retain heat."}
	raw := `{"summary":"Reasonable summary.","estimatedScore":{"value":79,"confidence":"high"},
	         "criteria":[],"strengths":[],"weaknesses":[],"sentenceFeedback":[]}`

	if _, err := parseWriting(raw, req); err == nil || !strings.Contains(err.Error(), "traits") {
		t.Fatalf("a PTE reply without its traits must be rejected, got: %v", err)
	}
}

func swtReply(content, form, grammar, vocabulary int) string {
	return fmt.Sprintf(`{"summary":"Reasonable summary.","estimatedScore":{"value":79,"confidence":"high"},
	         "criteria":[{"name":"Content","score":%d,"maxScore":4,"feedback":"Main idea."},
	                     {"name":"Form","score":%d,"maxScore":1,"feedback":"One sentence."},
	                     {"name":"Grammar","score":%d,"maxScore":2,"feedback":"Sound."},
	                     {"name":"Vocabulary","score":%d,"maxScore":2,"feedback":"Fitting."}],
	         "strengths":["Concise"],"weaknesses":["Thin detail"],
	         "sentenceFeedback":[{"original":"Cities retain heat because concrete and asphalt absorb sunlight.","correction":"x","issueType":"style","explanation":"y"}]}`,
		content, form, grammar, vocabulary)
}

// The estimate is worked out from the trait points, whatever number the model
// put in estimatedScore: 7 of 9 points is 10 + 80*(7/9)^1.25 = 68.
func TestParseWritingScoresFromPearsonTraits(t *testing.T) {
	req := WritingRequest{Exam: models.ExamPTE, TypeID: "pte-summarize-written-text", MinScore: 10, MaxScore: 90,
		LearnerText: "Cities retain heat because concrete and asphalt absorb sunlight."}

	evaluation, err := parseWriting(swtReply(3, 1, 2, 1), req)
	if err != nil {
		t.Fatalf("valid reply was rejected: %v", err)
	}
	if evaluation.EstimatedScore == nil || *evaluation.EstimatedScore != 68 {
		t.Errorf("score = %v, want 68 from the trait points", evaluation.EstimatedScore)
	}
	if len(evaluation.SentenceFeedback) != 1 {
		t.Errorf("sentence feedback quoting real text should survive, got %d entries", len(evaluation.SentenceFeedback))
	}

	// Content 0 is a gate: Pearson gives no points at all.
	gated, err := parseWriting(swtReply(0, 1, 2, 2), req)
	if err != nil {
		t.Fatal(err)
	}
	if *gated.EstimatedScore != 10 {
		t.Errorf("content 0 scored %v, want the scale minimum 10", *gated.EstimatedScore)
	}

	// Form follows the length rule, whatever the model said.
	req.LearnerText = "Cities retain heat."
	short, err := parseWriting(swtReply(3, 1, 2, 2), req)
	if err != nil {
		t.Fatal(err)
	}
	if *short.EstimatedScore != 10 {
		t.Errorf("a 3-word summary scored %v, want 10 (Form 0 under 5 words)", *short.EstimatedScore)
	}
}

// Feedback that quotes a sentence the learner never wrote is dropped, so the
// model cannot "correct" text it invented.
func TestParseWritingDropsInventedSentences(t *testing.T) {
	req := WritingRequest{Exam: models.ExamPTE, MinScore: 10, MaxScore: 90, LearnerText: "Cities retain heat."}
	raw := `{"summary":"Reasonable summary.","estimatedScore":{"value":null,"confidence":"medium"},
	         "criteria":[],"strengths":[],"weaknesses":[],
	         "sentenceFeedback":[{"original":"A sentence never written.","correction":"x","issueType":"grammar","explanation":"y"}]}`

	evaluation, err := parseWriting(raw, req)
	if err != nil {
		t.Fatalf("valid reply was rejected: %v", err)
	}
	if len(evaluation.SentenceFeedback) != 0 {
		t.Errorf("invented sentence should have been dropped, got %d entries", len(evaluation.SentenceFeedback))
	}
}

func TestWritingUserPromptCarriesFigureData(t *testing.T) {
	req := WritingRequest{
		TaskName:    "Describe the figure",
		Prompt:      "Summarise the information.",
		FigureData:  "Line graph. Canada 2000 = 51%, 2020 = 94%.",
		LearnerText: "The chart shows internet use rising.",
	}

	got := writingUserPrompt(req)
	for _, want := range []string{req.Prompt, req.FigureData, req.LearnerText} {
		if !strings.Contains(got, want) {
			t.Errorf("prompt is missing %q:\n%s", want, got)
		}
	}
}

func TestWritingUserPromptOmitsAbsentFigureData(t *testing.T) {
	req := WritingRequest{
		TaskName:    "Write Essay",
		Prompt:      "Do you agree?",
		LearnerText: "I disagree because...",
	}

	if got := writingUserPrompt(req); strings.Contains(got, "figure") {
		t.Errorf("essay prompt mentions a figure it does not have:\n%s", got)
	}
}
