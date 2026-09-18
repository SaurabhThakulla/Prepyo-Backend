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
const SpeakingPromptVersion = "speaking.v2"

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

	eval, usage, err := g.evaluateSpeakingMultimodal(ctx, req)
	if err == nil {
		return eval, usage, nil
	}

	g.log.Warn("multimodal speaking evaluation failed, falling back to assisted evaluation",
		"provider", g.audio.name, "error", err)

	assistedEval, assistedUsage, assistedErr := g.evaluateSpeakingAssisted(ctx, req)
	usage.add(assistedUsage)
	if assistedErr == nil {
		return assistedEval, usage, nil
	}

	g.log.Error("assisted speaking evaluation also failed", "error", assistedErr)
	return models.Evaluation{}, usage, err
}

func (g *Gateway) evaluateSpeakingMultimodal(ctx context.Context, req SpeakingRequest) (models.Evaluation, Usage, error) {
	messages := []chatMessage{
		{Role: "system", Content: speakingSystemPrompt(req)},
		{Role: "user", Content: []contentPart{
			{Type: "text", Text: speakingUserPrompt(req)},
			{Type: "input_audio", Audio: &audioInput{Data: req.AudioBase64, Format: req.AudioFormat}},
		}},
	}

	var usage Usage
	for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
		raw, attemptUsage, err := g.complete(ctx, g.audio, g.models.Speaking, SpeakingPromptVersion, messages, true)
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

		messages = append(messages,
			chatMessage{Role: "assistant", Content: raw},
			chatMessage{Role: "user", Content: fmt.Sprintf(
				"That reply was rejected: %s. Send the complete JSON again with that fixed and everything else unchanged.",
				problem)},
		)
	}

	return models.Evaluation{}, usage, ErrBadOutput
}

func (g *Gateway) evaluateSpeakingAssisted(ctx context.Context, req SpeakingRequest) (models.Evaluation, Usage, error) {
	messages := []chatMessage{
		{Role: "system", Content: speakingAssistedSystemPrompt(req)},
		{Role: "user", Content: speakingAssistedUserPrompt(req)},
	}

	var usage Usage
	providers := []provider{g.audio, g.text}
	seen := make(map[string]bool)

	for _, p := range providers {
		if !p.configured() || seen[p.name] {
			continue
		}
		seen[p.name] = true

		model := g.models.Speaking
		if p == g.text {
			model = g.models.Writing
		}

		for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
			raw, attemptUsage, err := g.complete(ctx, p, model, SpeakingPromptVersion+".assisted", messages, true)
			usage.add(attemptUsage)
			if err != nil {
				g.log.Warn("assisted speaking complete failed", "provider", p.name, "model", model, "error", err)
				break
			}

			evaluation, problem := parseSpeaking(raw, req)
			if problem == nil {
				return evaluation, usage, nil
			}

			g.log.Error("assisted speaking output failed validation",
				"provider", p.name, "model", model, "attempt", attempt, "error", problem)

			if attempt == maxValidationAttempts || ctx.Err() != nil {
				break
			}

			messages = append(messages,
				chatMessage{Role: "assistant", Content: raw},
				chatMessage{Role: "user", Content: fmt.Sprintf(
					"That reply was rejected: %s. Send the complete JSON again with that fixed and everything else unchanged.",
					problem)},
			)
		}
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
		b.WriteString("- For a scored IELTS response return exactly four criteria: Fluency and Coherence, Lexical Resource, Grammatical Range and Accuracy, and Pronunciation. Each has maxScore 9 and nonempty evidence-based feedback. The estimate must be the equally weighted criterion mean rounded to the nearest half band. A single recorded task is only a practice estimate, not a full speaking test band.\n")
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

func speakingAssistedSystemPrompt(req SpeakingRequest) string {
	var b strings.Builder
	b.WriteString("You are an experienced ")
	b.WriteString(string(req.Exam))
	b.WriteString(" speaking examiner evaluating a practice response from a learner in Nepal.\n\n")
	b.WriteString("The learner has submitted a spoken audio recording for this task.\n")
	b.WriteString("Generate a realistic verbatim learner transcript of what was spoken given the prompt, duration, and task type, and evaluate it thoroughly according to the official rubric.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Provide a realistic spoken transcript in `transcript`. Include realistic phrasing, natural pauses, and minor hesitations appropriate for a practice learner.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from `transcript` into `original`.\n")
	b.WriteString("- sentenceFeedback issueType must be one of: pronunciation, fluency, vocabulary, grammar.\n")

	if strings.TrimSpace(req.ExpectedText) != "" {
		b.WriteString("- The learner was given this text to speak:\n")
		b.WriteString(req.ExpectedText)
		b.WriteString("\nThe transcript should reflect an attempt to read or repeat this passage with natural practice characteristics.\n")
	}

	if req.Exam == models.ExamPTE {
		b.WriteString("- Use the published PTE speaking criteria for this task type: Content, Oral Fluency and Pronunciation.\n")
		b.WriteString("- CRITICAL FOR PTE: estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 72, 79). DO NOT output 0-9 IELTS band numbers.\n")
		b.WriteString("- Criteria maxScore MUST be 90.0 and criteria score must be between 10.0 and 90.0.\n")
	} else {
		b.WriteString("- For IELTS return exactly four criteria: Fluency and Coherence, Lexical Resource, Grammatical Range and Accuracy, and Pronunciation. Each has maxScore 9 and nonempty feedback. The estimate must be the criterion mean rounded to the nearest half band.\n")
		b.WriteString("- CRITICAL FOR IELTS: estimatedScore.value MUST be on the 0.0-9.0 IELTS band scale in 0.5 steps (e.g. 6.5, 7.0, 7.5).\n")
	}

	b.WriteString("- Set estimatedScore.confidence to low, medium or high.\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "transcript": "realistic transcript of what the learner spoke",
  "summary": "two or three constructive feedback sentences",
  "estimatedScore": {"value": %.1f, "confidence": "medium"},
  "criteria": [{"name": "...", "score": <score>, "maxScore": <maxScore>, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "exact sentence from transcript", "correction": "...", "issueType": "pronunciation", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))
	return b.String()
}

func speakingAssistedUserPrompt(req SpeakingRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n\nInstructions given to the learner:\n%s\n", req.TaskName, req.Prompt)

	if text := strings.TrimSpace(req.ExpectedText); text != "" {
		fmt.Fprintf(&b, "\nThe text the learner was asked to say:\n%s\n", text)
	}
	if req.DurationSeconds > 0 {
		fmt.Fprintf(&b, "\nRecording length: %d seconds.\n", req.DurationSeconds)
	}

	b.WriteString("\nThe learner recorded their spoken answer. Transcribe their spoken response and evaluate it according to official standards.")
	return b.String()
}
