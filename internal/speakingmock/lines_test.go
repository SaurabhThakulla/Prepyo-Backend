package speakingmock

import (
	"encoding/json"
	"strings"
	"testing"
)

const sampleSet = `{
  "part1": [
    {"topic": "your home", "questions": ["Do you live in a house or a flat?", "What do you like most about your home?"]},
    {"topic": "music", "questions": ["What kind of music do you enjoy?"]}
  ],
  "part2": {"topic": "a skill you learned", "cue": "Describe a skill.", "points": ["what"], "explain": "and explain.", "rounding": "Do you still use this skill today?"},
  "part3": {"topic": "learning skills", "questions": ["Why do some people give up?", "Is it better to learn from a teacher?"]}
}`

func TestExaminerLinesAreEveryLeadAndQuestionTheTestSays(t *testing.T) {
	lines, err := ExaminerLines([]byte(sampleSet))
	if err != nil {
		t.Fatal(err)
	}
	var content setContent
	if err := json.Unmarshal([]byte(sampleSet), &content); err != nil {
		t.Fatal(err)
	}
	steps := StepsFor(content)
	want := map[string]bool{}
	for _, step := range steps {
		for _, line := range []string{step.Lead, step.Question} {
			if line = strings.TrimSpace(line); line != "" {
				want[line] = true
			}
		}
	}
	if len(lines) != len(want) {
		t.Fatalf("got %d lines, want %d distinct", len(lines), len(want))
	}
	for _, line := range lines {
		if !want[line] {
			t.Errorf("unexpected line %q", line)
		}
	}
	// The Part 2 lead and question are said apart, so they are recorded apart.
	if !want["Remember, you have one to two minutes for this. I'll tell you when the time is up. Can you start speaking now, please?"] {
		t.Error("the Part 2 question is missing")
	}
	if lines[0] != "In this first part, I'd like to ask you some questions about yourself. Let's talk about your home." {
		t.Errorf("lines start with %q, want the Part 1 introduction", lines[0])
	}
}

func TestExaminerLinesRejectsBrokenContent(t *testing.T) {
	if _, err := ExaminerLines([]byte(`{"part1": "nope"}`)); err == nil {
		t.Fatal("want an error for content that is not a speaking set")
	}
}
