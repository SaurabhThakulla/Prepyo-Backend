package ai

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

// The worked example in the system prompt has to sit on the exam's own scale.
// When it did not, the model copied the example's IELTS-shaped band into a PTE
// evaluation and validateWriting threw the whole thing away, which reached the
// learner as a 503.
func TestExampleScoreStaysOnScale(t *testing.T) {
	cases := []struct {
		name string
		min  float64
		max  float64
		want float64
	}{
		{"PTE 10-90", 10, 90, 70},
		{"IELTS 0-9", 0, 9, 7},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := exampleScore(tc.min, tc.max)
			if got != tc.want {
				t.Errorf("exampleScore(%.1f, %.1f) = %.2f, want %.2f", tc.min, tc.max, got, tc.want)
			}
			if got < tc.min || got > tc.max {
				t.Errorf("example %.2f is outside the %.1f-%.1f scale it illustrates", got, tc.min, tc.max)
			}
		})
	}
}

func TestWritingSystemPromptExampleMatchesExam(t *testing.T) {
	pte := writingSystemPrompt(WritingRequest{Exam: models.ExamPTE, MinScore: 10, MaxScore: 90})
	if !strings.Contains(pte, `"value": 70.0`) {
		t.Errorf("PTE prompt should show a PTE-scale example score:\n%s", pte)
	}
	// The old IELTS-shaped anchors are what caused the bug.
	if strings.Contains(pte, `"value": 7.0`) || strings.Contains(pte, `"maxScore": 9.0`) {
		t.Errorf("PTE prompt still contains an IELTS-scale anchor:\n%s", pte)
	}

	ielts := writingSystemPrompt(WritingRequest{Exam: models.ExamIELTS, MinScore: 0, MaxScore: 9})
	if !strings.Contains(ielts, `"value": 7.0`) {
		t.Errorf("IELTS prompt should show an IELTS-scale example score:\n%s", ielts)
	}
}

// parseWriting must reject a score from the wrong exam's scale rather than
// clamping it: a band the model could not produce correctly is not one to guess.
func TestParseWritingRejectsOffScaleScore(t *testing.T) {
	req := WritingRequest{Exam: models.ExamPTE, MinScore: 10, MaxScore: 90, LearnerText: "Cities retain heat."}
	raw := `{"summary":"Reasonable summary.","estimatedScore":{"value":8.5,"confidence":"high"},
	         "criteria":[],"strengths":[],"weaknesses":[],"sentenceFeedback":[]}`

	if _, err := parseWriting(raw, req); err == nil {
		t.Fatal("expected an 8.5 band to be rejected on the 10-90 PTE scale")
	} else if !strings.Contains(err.Error(), "outside") {
		t.Errorf("error should name the range problem, got: %v", err)
	}
}

func TestParseWritingAcceptsOnScaleScore(t *testing.T) {
	req := WritingRequest{Exam: models.ExamPTE, MinScore: 10, MaxScore: 90, LearnerText: "Cities retain heat."}
	raw := `{"summary":"Reasonable summary.","estimatedScore":{"value":79,"confidence":"high"},
	         "criteria":[{"name":"Content","score":2,"maxScore":3,"feedback":"Covers the main idea."}],
	         "strengths":["Concise"],"weaknesses":["Thin detail"],
	         "sentenceFeedback":[{"original":"Cities retain heat.","correction":"Cities retain heat well.","issueType":"style","explanation":"Adds precision."}]}`

	evaluation, err := parseWriting(raw, req)
	if err != nil {
		t.Fatalf("valid reply was rejected: %v", err)
	}
	if evaluation.EstimatedScore == nil || *evaluation.EstimatedScore != 79 {
		t.Errorf("score not carried through: %+v", evaluation.EstimatedScore)
	}
	if len(evaluation.SentenceFeedback) != 1 {
		t.Errorf("sentence feedback quoting real text should survive, got %d entries", len(evaluation.SentenceFeedback))
	}
}

// Feedback that quotes a sentence the learner never wrote is dropped, so the
// model cannot "correct" text it invented.
func TestParseWritingDropsInventedSentences(t *testing.T) {
	req := WritingRequest{Exam: models.ExamIELTS, MinScore: 0, MaxScore: 9, LearnerText: "Cities retain heat."}
	raw := `{"summary":"Reasonable summary.","estimatedScore":{"value":7,"confidence":"medium"},
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
