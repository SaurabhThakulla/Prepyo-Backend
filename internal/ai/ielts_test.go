package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func ieltsReply(score, criterionScore float64, speaking bool) string {
	names := []string{"Task Response", "Coherence and Cohesion", "Lexical Resource", "Grammatical Range and Accuracy"}
	if speaking {
		names[0], names[1] = "Pronunciation", "Fluency and Coherence"
	}
	criteria := make([]map[string]any, 0, 4)
	for _, name := range names {
		criteria = append(criteria, map[string]any{"name": name, "score": criterionScore, "maxScore": 9, "feedback": "Relevant evidence."})
	}
	body, _ := json.Marshal(map[string]any{"summary": "Practice feedback.", "transcript": "Cities retain heat.", "estimatedScore": map[string]any{"value": score, "confidence": "low"}, "criteria": criteria})
	return string(body)
}

func TestIELTSRejectsRawOffScaleBeforeRounding(t *testing.T) {
	for _, score := range []float64{-0.1, 9.1, 70} {
		for _, speaking := range []bool{false, true} {
			t.Run(fmt.Sprintf("score=%g/speaking=%t", score, speaking), func(t *testing.T) {
				boundary := 9.0
				if score < 0 {
					boundary = 0
				}
				raw := ieltsReply(score, boundary, speaking)
				var err error
				if speaking {
					_, err = parseSpeaking(raw, SpeakingRequest{Exam: models.ExamIELTS, MinScore: 0, MaxScore: 9})
				} else {
					_, err = parseWriting(raw, WritingRequest{Exam: models.ExamIELTS, MinScore: 0, MaxScore: 9})
				}
				if err == nil || !strings.Contains(err.Error(), "outside") {
					t.Fatalf("raw score %g must be rejected before clamping: %v", score, err)
				}
			})
		}
	}
}

func TestIELTSCriteriaValidation(t *testing.T) {
	req := WritingRequest{Exam: models.ExamIELTS, MinScore: 0, MaxScore: 9}
	valid := ieltsReply(7, 7, false)
	for _, tc := range []struct {
		name, raw string
		valid     bool
	}{
		{"complete", valid, true},
		{"missing", `{"summary":"Brief","estimatedScore":{"value":7},"criteria":[]}`, false},
		{"duplicate", strings.Replace(valid, "Task Response", "Lexical Resource", 1), false},
		{"wrong scale", strings.ReplaceAll(valid, `"maxScore":9`, `"maxScore":90`), false},
		{"wrong mean", ieltsReply(8, 7, false), false},
		{"null", `{"summary":"Too short to score","estimatedScore":{"value":null},"criteria":[]}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseWriting(tc.raw, req)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%t: %v", tc.valid, err)
			}
		})
	}
	req.TaskName = "Describe the figure"
	if _, err := parseWriting(strings.Replace(valid, "Task Response", "Task Achievement", 1), req); err != nil {
		t.Fatal(err)
	}
	if _, err := parseSpeaking(ieltsReply(7, 7, true), SpeakingRequest{Exam: models.ExamIELTS, MinScore: 0, MaxScore: 9}); err != nil {
		t.Fatal(err)
	}
}
