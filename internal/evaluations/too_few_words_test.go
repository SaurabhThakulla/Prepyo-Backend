package evaluations

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

// A barely-heard answer is scored at the bottom of the scale, not refused, and
// the learner is told why.
func TestTooFewWordsIsScoredNotRefused(t *testing.T) {
	q := models.Question{Exam: models.ExamPTE, TypeName: "Re-tell Lecture"}
	v := models.ExamVersion{MinScore: 10, MaxScore: 90}

	for _, tc := range []struct {
		transcript string
		words      int
		says       string
	}{
		{"", 0, "No words could be made out"},
		{"the museum", 2, "Only 2 words could be made out"},
	} {
		e := tooFewWordsEvaluation(q, v, tc.transcript, tc.words)
		if e.EstimatedScore == nil || *e.EstimatedScore != 10 {
			t.Fatalf("%q: score = %v, want the scale minimum 10", tc.transcript, e.EstimatedScore)
		}
		if !strings.HasPrefix(e.Summary, tc.says) {
			t.Fatalf("%q: summary %q should start %q", tc.transcript, e.Summary, tc.says)
		}
		if e.ScoreConfidence != "low" || len(e.Criteria) == 0 || e.Transcript != tc.transcript {
			t.Fatalf("%q: incomplete evaluation %+v", tc.transcript, e)
		}
	}
}
