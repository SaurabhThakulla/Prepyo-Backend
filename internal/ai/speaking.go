package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// SpeakingPromptVersion is stamped on every stored evaluation. Change it
// whenever the prompt below changes, so old feedback stays traceable to the
// wording that produced it.
const SpeakingPromptVersion = "speaking.v1"

// AudioFormats are the encodings the provider accepts inline. Callers convert
// a browser recording to one of these before submitting; the gateway does not
// transcode, because a silent re-encode is a good way to ship a corrupted
// upload that scores as "we could not hear you".
var AudioFormats = map[string]bool{"wav": true, "mp3": true}

type SpeakingRequest struct {
	Exam     models.ExamType
	TaskName string
	Prompt   string
	// ExpectedText is the passage a Read Aloud learner was told to read, or the
	// sentence they were told to repeat. Empty for open-ended tasks such as an
	// IELTS cue card, where there is no single right answer to compare against.
	ExpectedText string
	// AudioBase64 is the recording, already in AudioFormat.
	AudioBase64 string
	AudioFormat string
	// DurationSeconds is what the browser actually recorded. It is passed to the
	// model as context, never used to derive a score here.
	DurationSeconds int
	MinScore        float64
	MaxScore        float64
}

// speakingPayload is the writing shape plus the transcript. The transcript is
// what makes speaking feedback checkable: it is both shown to the learner and
// used to verify that every quoted correction refers to something they said.
type speakingPayload struct {
	evaluationPayload
	Transcript string `json:"transcript"`
}

// EvaluateSpeaking transcribes and scores one recording in a single call.
//
// The estimate is a practice estimate. Neither this package nor its callers
// present it as an official Pearson or IELTS result.
func (g *Gateway) EvaluateSpeaking(ctx context.Context, req SpeakingRequest) (models.Evaluation, Usage, error) {
	if !AudioFormats[req.AudioFormat] {
		return models.Evaluation{}, Usage{}, fmt.Errorf("%w: audio format %q is not supported", ErrBadOutput, req.AudioFormat)
	}

	messages := []chatMessage{
		{Role: "system", Content: speakingSystemPrompt(req)},
		{Role: "user", Content: []contentPart{
			{Type: "text", Text: speakingUserPrompt(req)},
			{Type: "input_audio", Audio: &audioInput{Data: req.AudioBase64, Format: req.AudioFormat}},
		}},
	}

	var usage Usage

	for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
		raw, attemptUsage, err := g.complete(ctx, g.models.Speaking, SpeakingPromptVersion, messages, true)
		usage.add(attemptUsage)
		if err != nil {
			return models.Evaluation{}, Usage{}, err
		}

		evaluation, problem := parseSpeaking(raw, req)
		if problem == nil {
			return evaluation, usage, nil
		}

		g.log.Error("ai output failed validation",
			"model", g.models.Speaking, "attempt", attempt, "error", problem)

		if attempt == maxValidationAttempts || ctx.Err() != nil {
			break
		}

		// The correction round quotes the reply and the complaint back, exactly
		// as writing does. The recording stays in the first user turn, so the
		// model still has it without the audio being uploaded twice.
		messages = append(messages,
			chatMessage{Role: "assistant", Content: raw},
			chatMessage{Role: "user", Content: fmt.Sprintf(
				"That reply was rejected: %s. Send the complete JSON again with that fixed and everything else unchanged.",
				problem)},
		)
	}

	return models.Evaluation{}, usage, ErrBadOutput
}

// parseSpeaking decodes and validates one reply. It returns the underlying
// problem rather than ErrBadOutput so the caller can quote it back to the model.
func parseSpeaking(raw string, req SpeakingRequest) (models.Evaluation, error) {
	var payload speakingPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return models.Evaluation{}, fmt.Errorf("the reply was not valid JSON (%w)", err)
	}
	return validateSpeaking(payload, req)
}

// validateSpeaking adds the two rules that only apply to audio, then defers to
// the shared validator.
func validateSpeaking(p speakingPayload, req SpeakingRequest) (models.Evaluation, error) {
	transcript := strings.TrimSpace(p.Transcript)

	// A score with no transcript behind it cannot be checked by anyone: not by
	// the learner reading their feedback, and not by the sentence-quoting rule
	// below. Silence is a real outcome — a learner whose microphone was muted
	// must be told that, not handed a number the model guessed from nothing.
	if transcript == "" && p.EstimatedScore.Value != nil {
		return models.Evaluation{}, fmt.Errorf("scored %.2f but returned an empty transcript; if you could not make out any speech, set estimatedScore.value to null", *p.EstimatedScore.Value)
	}

	evaluation, err := validateFeedback(p.evaluationPayload, feedbackSpec{
		Exam:     req.Exam,
		Skill:    models.SkillSpeaking,
		MinScore: req.MinScore,
		MaxScore: req.MaxScore,
		Quotable: transcript,
	})
	if err != nil {
		return models.Evaluation{}, err
	}

	evaluation.Transcript = transcript
	return evaluation, nil
}

func speakingSystemPrompt(req SpeakingRequest) string {
	var b strings.Builder
	b.WriteString("You are an experienced ")
	b.WriteString(string(req.Exam))
	b.WriteString(" speaking examiner giving practice feedback to a learner in Nepal.\n\n")
	b.WriteString("You are given one audio recording of the learner's spoken response.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- First transcribe what you actually hear into `transcript`, verbatim. Include the learner's own errors, repetitions and false starts. Do not tidy them up.\n")
	b.WriteString("- If the recording is silent, unintelligible, or contains no speech, set `transcript` to \"\" and estimatedScore.value to null, and say so plainly in the summary.\n")
	b.WriteString("- Judge only this recording. Never quote or invent words the learner did not say.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from `transcript` into `original`.\n")
	b.WriteString("- Judge pronunciation, fluency and content from the audio itself: hesitation, pace, stress, intonation and intelligibility. Do not score pronunciation from the transcript alone.\n")

	if strings.TrimSpace(req.ExpectedText) != "" {
		b.WriteString("- The learner was given a fixed text to say. Compare what you heard against it and treat omissions, substitutions and additions as content errors.\n")
	}

	if req.Exam == models.ExamPTE {
		b.WriteString("- Use the published PTE speaking criteria for this task type: content, oral fluency and pronunciation.\n")
		b.WriteString("- CRITICAL FOR PTE: estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 79, 85). DO NOT output 0-9 IELTS band numbers.\n")
	} else {
		b.WriteString("- Use the published IELTS speaking criteria: fluency and coherence, lexical resource, grammatical range and accuracy, and pronunciation.\n")
		b.WriteString("- CRITICAL FOR IELTS: estimatedScore.value MUST be on the 0.0-9.0 IELTS band scale in 0.5 steps (e.g. 6.5, 7.0, 7.5).\n")
	}

	b.WriteString("- Set estimatedScore.confidence to low, medium or high based on how much the recording gives you. A very short recording is low confidence.\n")
	b.WriteString("- criteria[].maxScore is the maximum the published rubric gives that one criterion. It is not the exam's overall scale.\n")
	b.WriteString("- The example below already uses this exam's scale. Copy its shape, never its numbers.\n\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "transcript": "exactly what the learner said",
  "summary": "two or three sentences",
  "estimatedScore": {"value": %.1f, "confidence": "medium"},
  "criteria": [{"name": "...", "score": <this criterion's score>, "maxScore": <this criterion's maximum>, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "pronunciation", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))
	return b.String()
}

func speakingUserPrompt(req SpeakingRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n\nInstructions given to the learner:\n%s\n", req.TaskName, req.Prompt)

	if text := strings.TrimSpace(req.ExpectedText); text != "" {
		fmt.Fprintf(&b, "\nThe text the learner was asked to say:\n%s\n", text)
	}
	if req.DurationSeconds > 0 {
		fmt.Fprintf(&b, "\nRecording length: %d seconds.\n", req.DurationSeconds)
	}

	b.WriteString("\nThe learner's recording is attached. Transcribe it, then evaluate it.")
	return b.String()
}
