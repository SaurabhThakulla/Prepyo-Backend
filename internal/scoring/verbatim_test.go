package scoring

import (
	"strings"
	"testing"
)

func TestScoreVerbatim(t *testing.T) {
	const passage = "The museum opens at nine and closes at five on weekdays."

	// Bounds rather than exact fractions: what matters is that a faithful read
	// earns full marks and a degraded one is marked down in proportion, not the
	// arithmetic of one fixture sentence.
	tests := []struct {
		name        string
		spoken      string
		minAccuracy float64
		maxAccuracy float64
		wantMissed  bool
	}{
		{name: "read exactly", spoken: passage, minAccuracy: 1, maxAccuracy: 1},
		{
			// Recognition returns no punctuation and no capitals; neither is a
			// content error, and a learner must not be marked down for them.
			name:        "punctuation and case ignored",
			spoken:      "the museum opens at nine and closes at five on weekdays",
			minAccuracy: 1, maxAccuracy: 1,
		},
		{
			name:        "two words missed",
			spoken:      "The museum opens at nine and closes on weekdays.",
			minAccuracy: 0.75, maxAccuracy: 0.85,
			wantMissed: true,
		},
		{
			// Every word, none in sequence: that is not reading a passage aloud,
			// and a bag-of-words match would have called it perfect.
			name:        "right words wrong order",
			spoken:      "weekdays on five at closes and nine at opens museum the",
			minAccuracy: 0, maxAccuracy: 0.4,
			wantMissed: true,
		},
		{name: "nothing said", spoken: "", minAccuracy: 0, maxAccuracy: 0, wantMissed: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ScoreVerbatim(passage, tc.spoken)
			if got.Accuracy < tc.minAccuracy-0.01 || got.Accuracy > tc.maxAccuracy+0.01 {
				t.Errorf("Accuracy = %.3f (matched %d of %d), want between %.2f and %.2f",
					got.Accuracy, got.Matched, got.Expected, tc.minAccuracy, tc.maxAccuracy)
			}
			if tc.wantMissed && len(got.Missed) == 0 {
				t.Error("no missed words listed for an imperfect read")
			}
			if !tc.wantMissed && len(got.Missed) > 0 {
				t.Errorf("missed words listed for a perfect read: %v", got.Missed)
			}
		})
	}
}

func TestVerbatimFeedbackNamesTheWords(t *testing.T) {
	result := ScoreVerbatim("The museum opens at nine", "the museum opens")
	feedback := result.Feedback()

	for _, want := range []string{"3 of the 5", "at", "nine"} {
		if !strings.Contains(feedback, want) {
			t.Errorf("feedback %q does not mention %q", feedback, want)
		}
	}
}

func TestIsVerbatimTask(t *testing.T) {
	for _, id := range []string{"read-aloud", "pte-read-aloud", "pte-repeat-sentence"} {
		if !IsVerbatimTask(id) {
			t.Errorf("IsVerbatimTask(%q) = false", id)
		}
	}
	// Re-tell Lecture carries the lecture's transcript, not a script for the
	// learner. Aligning against it would mark down the task's whole point.
	for _, id := range []string{"pte-retell-lecture", "pte-describe-image", "ielts-speaking-part2", ""} {
		if IsVerbatimTask(id) {
			t.Errorf("IsVerbatimTask(%q) = true", id)
		}
	}
}

func TestDeliveryFluency(t *testing.T) {
	steady := Delivery{DurationSeconds: 40, WordsSpoken: 90, PauseCount: 1, LongestPauseSeconds: 0.8, SpeakingRatio: 0.85}
	if got := steady.FluencyAccuracy(); got < 0.95 {
		t.Errorf("steady delivery scored %.2f, want near 1", got)
	}

	halting := Delivery{DurationSeconds: 40, WordsSpoken: 30, PauseCount: 9, LongestPauseSeconds: 4.5, SpeakingRatio: 0.35}
	if got := halting.FluencyAccuracy(); got > 0.5 {
		t.Errorf("halting delivery scored %.2f, want well below steady", got)
	}

	// Without measurements there is nothing to judge, and inventing a pace mark
	// is the thing this whole path exists to avoid.
	var unmeasured Delivery
	if unmeasured.Measured() || unmeasured.FluencyAccuracy() != 0 {
		t.Error("unmeasured delivery reported a fluency score")
	}
}
