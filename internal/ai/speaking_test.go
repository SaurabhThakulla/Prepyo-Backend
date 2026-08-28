package ai

import (
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/models"
)

func pteSpeaking() SpeakingRequest {
	return SpeakingRequest{
		Exam: models.ExamPTE, TaskName: "Read Aloud",
		AudioBase64: "AAAA", AudioFormat: "wav",
		MinScore: 10, MaxScore: 90,
	}
}

// A recording the model could not make out is a real outcome, and the only
// honest response to it is no score. A number with no transcript behind it is
// unfalsifiable: the learner cannot see what was judged, and the sentence
// quoting rule has nothing to check against.
func TestParseSpeakingRejectsScoreWithoutTranscript(t *testing.T) {
	raw := `{"transcript":"","summary":"Sounded confident.","estimatedScore":{"value":72,"confidence":"high"},
	         "criteria":[],"strengths":[],"weaknesses":[],"sentenceFeedback":[]}`

	if _, err := parseSpeaking(raw, pteSpeaking()); err == nil {
		t.Fatal("expected a score with an empty transcript to be rejected")
	} else if !strings.Contains(err.Error(), "transcript") {
		t.Errorf("error should name the transcript problem, got: %v", err)
	}
}

// Silence itself must still get through, so the learner is told their
// microphone caught nothing rather than seeing a 503.
func TestParseSpeakingAcceptsSilenceWithNullScore(t *testing.T) {
	raw := `{"transcript":"","summary":"No speech could be heard in this recording.",
	         "estimatedScore":{"value":null,"confidence":"low"},
	         "criteria":[],"strengths":[],"weaknesses":[],"sentenceFeedback":[]}`

	evaluation, err := parseSpeaking(raw, pteSpeaking())
	if err != nil {
		t.Fatalf("an honest silent result was rejected: %v", err)
	}
	if evaluation.EstimatedScore != nil {
		t.Errorf("silence should carry no score, got %v", *evaluation.EstimatedScore)
	}
	if evaluation.Skill != models.SkillSpeaking {
		t.Errorf("skill = %q, want speaking", evaluation.Skill)
	}
}

func TestParseSpeakingCarriesTranscript(t *testing.T) {
	raw := `{"transcript":"The global transition toward sustainable energy.",
	         "summary":"Clear and steady.","estimatedScore":{"value":79,"confidence":"medium"},
	         "criteria":[{"name":"Oral Fluency","score":4,"maxScore":5,"feedback":"Even pace."}],
	         "strengths":["Steady pace"],"weaknesses":["Flat intonation"],
	         "sentenceFeedback":[{"original":"The global transition toward sustainable energy.","correction":"x","issueType":"pronunciation","explanation":"y"}]}`

	evaluation, err := parseSpeaking(raw, pteSpeaking())
	if err != nil {
		t.Fatalf("valid reply was rejected: %v", err)
	}
	if evaluation.Transcript != "The global transition toward sustainable energy." {
		t.Errorf("transcript not carried through: %q", evaluation.Transcript)
	}
	if len(evaluation.SentenceFeedback) != 1 {
		t.Errorf("feedback quoting the transcript should survive, got %d entries", len(evaluation.SentenceFeedback))
	}
}

// Speaking feedback is quoted against the transcript, not against text the
// learner typed, so an invented sentence must be dropped the same way.
func TestParseSpeakingDropsSentencesNotInTranscript(t *testing.T) {
	raw := `{"transcript":"Cities retain heat.","summary":"Brief but clear.",
	         "estimatedScore":{"value":65,"confidence":"low"},
	         "criteria":[],"strengths":[],"weaknesses":[],
	         "sentenceFeedback":[{"original":"A sentence never spoken.","correction":"x","issueType":"grammar","explanation":"y"}]}`

	evaluation, err := parseSpeaking(raw, pteSpeaking())
	if err != nil {
		t.Fatalf("valid reply was rejected: %v", err)
	}
	if len(evaluation.SentenceFeedback) != 0 {
		t.Errorf("invented sentence should have been dropped, got %d entries", len(evaluation.SentenceFeedback))
	}
}

// The off-scale guard is shared with writing; this pins it for speaking too,
// because the speaking prompt carries its own worked example.
func TestParseSpeakingRejectsOffScaleScore(t *testing.T) {
	raw := `{"transcript":"Cities retain heat.","summary":"Brief.",
	         "estimatedScore":{"value":8.5,"confidence":"high"},
	         "criteria":[],"strengths":[],"weaknesses":[],"sentenceFeedback":[]}`

	if _, err := parseSpeaking(raw, pteSpeaking()); err == nil {
		t.Fatal("expected an 8.5 band to be rejected on the 10-90 PTE scale")
	}
}

func TestSpeakingSystemPromptExampleMatchesExam(t *testing.T) {
	pte := speakingSystemPrompt(SpeakingRequest{Exam: models.ExamPTE, MinScore: 10, MaxScore: 90})
	if !strings.Contains(pte, `"value": 70.0`) {
		t.Errorf("PTE prompt should show a PTE-scale example score:\n%s", pte)
	}

	ielts := speakingSystemPrompt(SpeakingRequest{Exam: models.ExamIELTS, MinScore: 0, MaxScore: 9})
	if !strings.Contains(ielts, `"value": 7.0`) {
		t.Errorf("IELTS prompt should show an IELTS-scale example score:\n%s", ielts)
	}
}

// An unsupported container reaching the provider costs a call and comes back as
// an opaque failure, so it is caught before the request is built.
func TestEvaluateSpeakingRejectsUnsupportedFormat(t *testing.T) {
	for _, format := range []string{"webm", "ogg", "m4a", ""} {
		if AudioFormats[format] {
			t.Errorf("format %q should not be accepted by the provider", format)
		}
	}
	for _, format := range []string{"wav", "mp3"} {
		if !AudioFormats[format] {
			t.Errorf("format %q should be accepted", format)
		}
	}
}
