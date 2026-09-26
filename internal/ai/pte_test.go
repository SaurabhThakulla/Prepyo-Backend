package ai

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func TestPTEPromptsDoNotContainStrictOrPunitiveLanguage(t *testing.T) {
	forbidden := []string{
		"rigorous",
		"strict",
		"scrutini",
		"lenient",
		"native-like",
		"Firstly",
		"wpm",
		"110",
		"175",
		"PTE 55",
		"50-64",
	}

	testCases := []struct {
		name   string
		prompt string
	}{
		{
			name: "writing-essay",
			prompt: writingSystemPrompt(WritingRequest{
				Exam:     models.ExamPTE,
				TaskName: "Write Essay",
				TypeID:   "pte-write-essay",
				MinScore: 10,
				MaxScore: 90,
			}),
		},
		{
			name: "writing-swt",
			prompt: writingSystemPrompt(WritingRequest{
				Exam:     models.ExamPTE,
				TaskName: "Summarize Written Text",
				TypeID:   "pte-summarize-written-text",
				MinScore: 10,
				MaxScore: 90,
			}),
		},
		{
			name: "speaking-describe-image",
			prompt: speakingSystemPrompt(SpeakingRequest{
				Exam:     models.ExamPTE,
				TaskName: "Describe Image",
				MinScore: 10,
				MaxScore: 90,
			}),
		},
		{
			name: "speaking-respond-to-situation",
			prompt: speakingSystemPrompt(SpeakingRequest{
				Exam:     models.ExamPTE,
				TaskName: "Respond to a Situation",
				MinScore: 10,
				MaxScore: 90,
			}),
		},
		{
			name: "speaking-summarise-group-discussion",
			prompt: speakingSystemPrompt(SpeakingRequest{
				Exam:     models.ExamPTE,
				TaskName: "Summarise Group Discussion",
				MinScore: 10,
				MaxScore: 90,
			}),
		},
		{
			name: "transcript-describe-image",
			prompt: transcriptSystemPrompt(SpokenTranscriptRequest{
				Exam:     models.ExamPTE,
				TaskName: "Describe Image",
				MinScore: 10,
				MaxScore: 90,
			}),
		},
		{
			name: "transcript-respond-to-situation",
			prompt: transcriptSystemPrompt(SpokenTranscriptRequest{
				Exam:     models.ExamPTE,
				TaskName: "Respond to a Situation",
				MinScore: 10,
				MaxScore: 90,
			}),
		},
	}

	for _, tc := range testCases {
		for _, word := range forbidden {
			if strings.Contains(strings.ToLower(tc.prompt), strings.ToLower(word)) {
				t.Errorf("%s prompt must not contain %q:\n%s", tc.name, word, tc.prompt)
			}
		}
	}
}

func TestPTEPromptsContainEvidenceBasedGuidance(t *testing.T) {
	writingReq := WritingRequest{
		Exam:     models.ExamPTE,
		TaskName: "Write Essay",
		TypeID:   "pte-write-essay",
		MinScore: 10,
		MaxScore: 90,
	}
	writing := writingSystemPrompt(writingReq)
	if !strings.Contains(writing, "experienced PTE Academic assessor") {
		t.Error("writing prompt should frame assessor neutrally")
	}
	if !strings.Contains(writing, "10-90 PTE points scale") {
		t.Error("writing prompt should mention 10-90 scale")
	}
	if !strings.Contains(writing, "Every criterion is scored on the 10-90 scale with maxScore 90") {
		t.Error("writing prompt should state criteria maxScore is 90")
	}
	if !strings.Contains(writing, "Development, Structure and Coherence") || !strings.Contains(writing, "General Linguistic Range") {
		t.Error("writing essay prompt should name Write Essay Pearson traits")
	}

	speakingReq := SpeakingRequest{
		Exam:     models.ExamPTE,
		TaskName: "Describe Image",
		MinScore: 10,
		MaxScore: 90,
	}
	speaking := speakingSystemPrompt(speakingReq)
	if !strings.Contains(speaking, "accent is not a fault") {
		t.Error("speaking prompt must include the accent-is-not-a-fault guidance")
	}
	if !strings.Contains(speaking, "natural pause are not failures and do not cap the score") {
		t.Error("speaking prompt must include the natural pause guidance")
	}
	if !strings.Contains(speaking, "Oral Fluency") || !strings.Contains(speaking, "Pronunciation") {
		t.Error("speaking prompt should name Oral Fluency and Pronunciation traits")
	}
	if !strings.Contains(speaking, "10-90 PTE points scale") {
		t.Error("speaking prompt should mention 10-90 scale")
	}
	if !strings.Contains(speaking, "Every criterion is scored on the 10-90 scale with maxScore 90") {
		t.Error("speaking prompt should state criteria maxScore is 90")
	}

	transcriptReq := SpokenTranscriptRequest{
		Exam:     models.ExamPTE,
		TaskName: "Describe Image",
		MinScore: 10,
		MaxScore: 90,
	}
	transcript := transcriptSystemPrompt(transcriptReq)
	if !strings.Contains(transcript, "Do not include a pronunciation criterion") {
		t.Error("transcript prompt must instruct not to include pronunciation")
	}
	if !strings.Contains(transcript, "10-90 PTE points scale") {
		t.Error("transcript prompt should mention 10-90 scale")
	}
	if !strings.Contains(transcript, "Every criterion is scored on the 10-90 scale with maxScore 90") {
		t.Error("transcript prompt should state criteria maxScore is 90")
	}
}

func TestPTEWritingTraitsByTask(t *testing.T) {
	cases := []struct {
		typeID   string
		taskName string
		want     []string
	}{
		{
			typeID:   "pte-write-essay",
			taskName: "Write Essay",
			want:     []string{"Content", "Form", "Development, Structure and Coherence", "Grammar", "General Linguistic Range", "Vocabulary Range", "Spelling"},
		},
		{
			typeID:   "pte-repeated-essay-questions",
			taskName: "Repeated Essay Questions",
			want:     []string{"Content", "Form", "Development, Structure and Coherence", "Grammar", "General Linguistic Range", "Vocabulary Range", "Spelling"},
		},
		{
			typeID:   "summarize-written-text",
			taskName: "Summarize Written Text",
			want:     []string{"Content", "Form", "Grammar", "Vocabulary"},
		},
		{
			typeID:   "summarize-spoken-text",
			taskName: "Summarize Spoken Text",
			want:     []string{"Content", "Form", "Grammar", "Vocabulary", "Spelling"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.taskName, func(t *testing.T) {
			guidance := pteWritingGuidance(WritingRequest{Exam: models.ExamPTE, TypeID: tc.typeID, TaskName: tc.taskName})
			for _, trait := range tc.want {
				if !strings.Contains(guidance, trait) {
					t.Errorf("%s guidance missing trait %q:\n%s", tc.taskName, trait, guidance)
				}
			}
		})
	}
}

func TestPTESpeakingTraitsByTask(t *testing.T) {
	openSpeakingTasks := []string{
		"Read Aloud",
		"Repeat Sentence",
		"Describe Image",
		"Re-tell Lecture",
		"Summarize Group Discussion",
		"Summarise Group Discussion",
	}

	for _, taskName := range openSpeakingTasks {
		t.Run(taskName, func(t *testing.T) {
			guidance := pteSpeakingGuidance(SpeakingRequest{Exam: models.ExamPTE, TaskName: taskName})
			for _, trait := range []string{"Content", "Oral Fluency", "Pronunciation"} {
				if !strings.Contains(guidance, trait) {
					t.Errorf("%s guidance missing trait %q", taskName, trait)
				}
			}
		})
	}

	asqGuidance := pteSpeakingGuidance(SpeakingRequest{Exam: models.ExamPTE, TaskName: "Answer Short Question"})
	if !strings.Contains(asqGuidance, "Content") {
		t.Errorf("Answer Short Question missing Content trait")
	}
}

func TestPTERespondToSituationSpeakingRubric(t *testing.T) {
	for _, taskName := range []string{"Respond to a Situation", "Respond to Situation"} {
		t.Run(taskName, func(t *testing.T) {
			guidance := pteSpeakingGuidance(SpeakingRequest{Exam: models.ExamPTE, TaskName: taskName})

			for _, trait := range []string{"Appropriacy", "Oral Fluency", "Pronunciation"} {
				if !strings.Contains(guidance, trait) {
					t.Errorf("%s speaking guidance missing trait %q:\n%s", taskName, trait, guidance)
				}
			}
			if strings.Contains(guidance, "Return the published PTE speaking criteria: Content") || strings.Contains(guidance, "Content:\n- Judge how accurately") {
				t.Errorf("%s speaking guidance must NOT ask for Content:\n%s", taskName, guidance)
			}
			if !strings.Contains(guidance, "suits the situation, the person being spoken to and the purpose") {
				t.Errorf("%s speaking guidance missing Appropriacy description:\n%s", taskName, guidance)
			}
		})
	}
}

func TestPTERespondToSituationTranscriptRubric(t *testing.T) {
	for _, taskName := range []string{"Respond to a Situation", "Respond to Situation"} {
		t.Run(taskName, func(t *testing.T) {
			guidance := pteTranscriptGuidance(SpokenTranscriptRequest{Exam: models.ExamPTE, TaskName: taskName})

			if !strings.Contains(guidance, "Appropriacy") || !strings.Contains(guidance, "Oral Fluency") {
				t.Errorf("%s transcript guidance missing Appropriacy or Oral Fluency:\n%s", taskName, guidance)
			}
			if !strings.Contains(guidance, "Do not return a Pronunciation criterion") {
				t.Errorf("%s transcript guidance must instruct not to include Pronunciation:\n%s", taskName, guidance)
			}
			if !strings.Contains(guidance, "Return exactly two criteria: Appropriacy and Oral Fluency") {
				t.Errorf("%s transcript guidance must return Appropriacy and Oral Fluency:\n%s", taskName, guidance)
			}
			if strings.Contains(guidance, "Return exactly two criteria: Content and Oral Fluency") {
				t.Errorf("%s transcript guidance must not return Content:\n%s", taskName, guidance)
			}
			if !strings.Contains(guidance, "suits the situation, the person being spoken to and the purpose") {
				t.Errorf("%s transcript guidance missing Appropriacy description:\n%s", taskName, guidance)
			}
		})
	}
}

func TestPTETaskNameMatching(t *testing.T) {
	tests := []struct {
		typeID   string
		taskName string
		want     string
	}{
		{"", "Respond to a Situation", pteTaskRespondToSituation},
		{"", "Respond to Situation", pteTaskRespondToSituation},
		{"pte-respond-to-situation", "Respond to a Situation", pteTaskRespondToSituation},
		{"", "Summarise Group Discussion", pteTaskSummarizeGroupDiscussion},
		{"", "Summarize Group Discussion", pteTaskSummarizeGroupDiscussion},
		{"pte-summarise-group-discussion", "", pteTaskSummarizeGroupDiscussion},
		{"pte-summarize-group-discussion", "", pteTaskSummarizeGroupDiscussion},
		{"", "Read Aloud", pteTaskReadAloud},
		{"pte-read-aloud", "", pteTaskReadAloud},
		{"", "Repeat Sentence", pteTaskRepeatSentence},
		{"pte-repeat-sentence", "", pteTaskRepeatSentence},
		{"", "Describe Image", pteTaskDescribeImage},
		{"pte-describe-image", "", pteTaskDescribeImage},
		{"", "Re-tell Lecture", pteTaskRetellLecture},
		{"", "Retell Lecture", pteTaskRetellLecture},
		{"pte-retell-lecture", "", pteTaskRetellLecture},
		{"", "Answer Short Question", pteTaskAnswerShortQuestion},
		{"pte-answer-short-question", "", pteTaskAnswerShortQuestion},
	}

	for _, tc := range tests {
		got := pteSpeakingTask(tc.typeID, tc.taskName)
		if got != tc.want {
			t.Errorf("pteSpeakingTask(%q, %q) = %q, want %q", tc.typeID, tc.taskName, got, tc.want)
		}
	}
}
