package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// SpeakingTestPromptVersion identifies the whole-test speaking prompt.
const SpeakingTestPromptVersion = "speaking.test.v1"

// SpeakingTestTurn is one question of a Speaking test and the answer given.
type SpeakingTestTurn struct {
	Part     int
	Question string
	Answer   string
	Seconds  int
}

// SpeakingTestRequest is a complete Speaking test, rated as one performance.
type SpeakingTestRequest struct {
	Turns    []SpeakingTestTurn
	MinScore float64
	MaxScore float64
}

// EvaluateSpeakingTest rates a whole IELTS Speaking test from transcripts of
// each answer. IELTS rates "average performance across all parts of the test",
// so the three parts are judged together rather than answer by answer.
//
// The answers were transcribed (Whisper on the audio provider, or the
// browser's recogniser), so pronunciation cannot be judged: the estimate uses
// Fluency and Coherence, Lexical Resource and Grammatical Range and Accuracy,
// and says so.
func (g *Gateway) EvaluateSpeakingTest(ctx context.Context, req SpeakingTestRequest) (models.Evaluation, Usage, error) {
	if !g.text.configured() {
		return models.Evaluation{}, Usage{}, fmt.Errorf("%w: no text provider is configured", ErrUnavailable)
	}

	var transcript, candidate strings.Builder
	part := 0
	for _, turn := range req.Turns {
		if turn.Part != part {
			part = turn.Part
			fmt.Fprintf(&transcript, "\nPart %d\n", part)
		}
		answer := strings.TrimSpace(turn.Answer)
		if answer == "" {
			answer = "(no answer was heard)"
		} else {
			candidate.WriteString(answer + "\n")
		}
		fmt.Fprintf(&transcript, "Examiner: %s\nCandidate (%ds): %s\n", turn.Question, turn.Seconds, answer)
	}

	var system strings.Builder
	system.WriteString("You are an experienced IELTS Speaking assessor giving practice feedback on a complete Speaking test. You are not an official IELTS examiner, and your estimate is not an official score.\n\n")
	system.WriteString("You are given transcripts of the candidate's answers to all three parts, produced by speech recognition. You cannot hear the recordings.\n")
	system.WriteString("- Speech recognition removes many hesitations and makes its own mistakes. A single odd word is more likely a recognition error; be cautious about fluency judgements.\n")
	system.WriteString("- NEVER score or comment on pronunciation, accent, intonation or stress. Do not include a pronunciation criterion.\n")
	system.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from one of the candidate's answers into `original` and give an improved version with a short explanation.\n\n")
	system.WriteString(ieltsSpeakingGuidance("full test", false))
	fmt.Fprintf(&system, "\nReturn exactly three criteria: %s. Each has maxScore 9, a whole-band score, and feedback that cites evidence from across the test. Do not return a Pronunciation criterion.\n", ieltsSpeakingCriteria(false))
	system.WriteString("estimatedScore.value is the mean of the three criteria rounded to the nearest half band (.25 rounds up to the next half band, .75 up to the next whole band). Because pronunciation is not assessed it is a three-criterion practice estimate, not an IELTS Speaking band; say so in the summary.\n")
	system.WriteString("If the candidate said too little to rate, set estimatedScore.value to null and say so.\n")
	system.WriteString("Set estimatedScore.confidence to low or medium.\n")
	system.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "summary": "three or four sentences on the whole test, ending with the note that pronunciation was not assessed",
  "estimatedScore": {"value": %.1f, "confidence": "low"},
  "criteria": [{"name": "...", "score": <whole-band score>, "maxScore": 9, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "grammar", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))

	messages := []chatMessage{
		{Role: "system", Content: system.String()},
		{Role: "user", Content: "Transcript of the Speaking test:\n" + transcript.String() + "\nRate the whole test under the rules above."},
	}

	var usage Usage
	for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
		raw, attemptUsage, err := g.complete(ctx, g.text, g.models.Writing, SpeakingTestPromptVersion, messages, true)
		usage.add(attemptUsage)
		if err != nil {
			return models.Evaluation{}, usage, err
		}
		var payload evaluationPayload
		problem := json.Unmarshal([]byte(raw), &payload)
		var evaluation models.Evaluation
		if problem == nil {
			evaluation, problem = validateFeedback(payload, feedbackSpec{
				Exam:                 models.ExamIELTS,
				Skill:                models.SkillSpeaking,
				MinScore:             req.MinScore,
				MaxScore:             req.MaxScore,
				Quotable:             candidate.String(),
				TaskName:             "full test",
				WithoutPronunciation: true,
			})
		}
		if problem == nil {
			evaluation.Transcript = strings.TrimSpace(transcript.String())
			return evaluation, usage, nil
		}
		g.log.Error("speaking test output failed validation", "attempt", attempt, "error", problem)
		if attempt == maxValidationAttempts || ctx.Err() != nil {
			break
		}
		messages = append(messages,
			chatMessage{Role: "assistant", Content: raw},
			chatMessage{Role: "user", Content: fmt.Sprintf("That reply was rejected: %s. Send the complete JSON again with that fixed and everything else unchanged.", problem)},
		)
	}
	return models.Evaluation{}, usage, ErrBadOutput
}
