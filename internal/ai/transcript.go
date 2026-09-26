package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// SpeakingTranscriptPromptVersion identifies the transcript-based speaking
// prompt. It is deliberately a different version from SpeakingPromptVersion:
// feedback produced this way saw words, not a recording, and stored evaluations
// have to stay traceable to which of the two produced them.
const SpeakingTranscriptPromptVersion = "speaking.transcript.v3"

// SpokenTranscriptRequest is a speaking answer that was transcribed on the
// learner's device, for providers that serve text models only.
//
// This is the honest half of speaking practice without an audio provider: what
// the learner said can be judged, how they said it cannot. Nothing here scores
// pronunciation, and the prompts below forbid guessing at it.
type SpokenTranscriptRequest struct {
	Exam     models.ExamType
	TaskName string
	Prompt   string
	// ExpectedText is the passage a Read Aloud learner was told to read, or the
	// sentence they were told to repeat. Empty for open tasks.
	ExpectedText string
	// SourceText and ReferenceAnswer are as on SpeakingRequest: material to
	// respond to, and a model answer to calibrate content against.
	SourceText      string
	ReferenceAnswer string
	// Transcript is what the device's speech recognition heard. It carries
	// recognition errors as well as the learner's own, which is why the prompt
	// tells the model not to treat every oddity as a language mistake.
	Transcript string
	// DurationSeconds and WordCount give the model the one part of delivery a
	// transcript can support on its own: how much was said in how long.
	DurationSeconds int
	WordCount       int
	// Delivery is what the browser measured from the waveform — pace, pauses,
	// how much of the recording had voice in it. It is handed to the model as
	// evidence so any fluency comment rests on a measurement. Empty when the
	// browser measured nothing, and then there is nothing to say about pace.
	Delivery string
	MinScore float64
	MaxScore float64
}

// EvaluateSpokenTranscript scores a spoken answer from its transcript using the
// text provider.
//
// The estimate is a practice estimate, and a narrower one than a recording
// would give: it is explicitly provisional, and the summary says so.
func (g *Gateway) EvaluateSpokenTranscript(ctx context.Context, req SpokenTranscriptRequest) (models.Evaluation, Usage, error) {
	if strings.TrimSpace(req.Transcript) == "" {
		return models.Evaluation{}, Usage{}, fmt.Errorf("%w: the transcript is empty", ErrBadOutput)
	}
	if !g.text.configured() {
		return models.Evaluation{}, Usage{}, fmt.Errorf("%w: no text provider is configured", ErrUnavailable)
	}

	messages := []chatMessage{
		{Role: "system", Content: transcriptSystemPrompt(req)},
		{Role: "user", Content: transcriptUserPrompt(req)},
	}

	var usage Usage
	for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
		raw, attemptUsage, err := g.complete(ctx, g.text, g.models.Writing, SpeakingTranscriptPromptVersion, messages, true)
		usage.add(attemptUsage)
		if err != nil {
			return models.Evaluation{}, usage, err
		}

		evaluation, problem := parseSpokenTranscript(raw, req)
		if problem == nil {
			return evaluation, usage, nil
		}

		g.log.Error("transcript speaking output failed validation",
			"model", g.models.Writing, "attempt", attempt, "error", problem)

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

func parseSpokenTranscript(raw string, req SpokenTranscriptRequest) (models.Evaluation, error) {
	transcript := strings.TrimSpace(req.Transcript)

	var payload evaluationPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return models.Evaluation{}, fmt.Errorf("the reply was not valid JSON (%w)", err)
	}

	evaluation, err := validateFeedback(payload, feedbackSpec{
		Exam:     req.Exam,
		Skill:    models.SkillSpeaking,
		MinScore: req.MinScore,
		MaxScore: req.MaxScore,
		// Every quoted correction still has to come from something the learner
		// actually said.
		Quotable: transcript,
		TaskName: req.TaskName,
		// Pronunciation is not observable in a transcript, so the usual
		// four-criterion IELTS shape does not apply here.
		WithoutPronunciation: true,
	})
	if err != nil {
		return models.Evaluation{}, err
	}

	evaluation.Transcript = transcript
	return evaluation, nil
}

// speechRate is words per minute, the one delivery measure a transcript plus a
// duration can support honestly. Normal conversational English sits near 130.
func speechRate(words, seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return words * 60 / seconds
}

func transcriptSystemPrompt(req SpokenTranscriptRequest) string {
	if isIELTS(req.Exam) {
		return ieltsTranscriptSystemPrompt(req)
	}

	var b strings.Builder
	b.WriteString(pteTranscriptGuidance(req))
	b.WriteString("\nThe example below already uses this exam's scale. Copy its shape, never its numbers.\n\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "summary": "two or three sentences of evidence-based assessment, ending with the provisional-estimate caveat",
  "estimatedScore": {"value": %.1f, "confidence": "low"},
  "criteria": [{"name": "...", "score": <score on 10-90 scale>, "maxScore": 90, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "grammar", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))
	return b.String()
}

// ieltsTranscriptSystemPrompt scores an IELTS answer from a transcript. It can
// judge three criteria; pronunciation needs the audio and is left out.
func ieltsTranscriptSystemPrompt(req SpokenTranscriptRequest) string {
	var b strings.Builder
	b.WriteString("You are an experienced IELTS Speaking assessor giving practice feedback. You are not an official IELTS examiner, and your estimate is not an official score.\n\n")
	b.WriteString("You are given a transcript of the learner's spoken answer, produced by speech recognition. You cannot hear the recording.\n")
	b.WriteString("- Speech recognition makes its own mistakes and often removes hesitations, fillers and false starts. A single odd word is more likely a recognition error than a learner error: ignore it unless the pattern repeats. Be cautious about fluency judgements: a smooth transcript does not prove smooth speech.\n")
	b.WriteString("- Measured timing, if given below, is evidence about pace and pauses. Use it to describe delivery; there is no fixed words-per-minute threshold in the descriptors, so do not apply one.\n")
	b.WriteString("- NEVER score or comment on pronunciation, accent, intonation or stress. You did not hear them. Do not include a pronunciation criterion.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from the transcript into `original` and give an improved version with a short explanation.\n\n")
	b.WriteString(ieltsSpeakingGuidance(req.TaskName, false))
	b.WriteString(responseMaterialRules(req.Exam, req.SourceText, req.ReferenceAnswer))
	fmt.Fprintf(&b, "\nReturn exactly three criteria: %s. Each has maxScore 9, a whole-band score, and feedback that cites evidence from the transcript. Do not return a Pronunciation criterion.\n", ieltsSpeakingCriteria(false))
	b.WriteString("estimatedScore.value is the mean of those three criteria rounded to the nearest half band (.25 rounds up to the next half band, .75 up to the next whole band). Because pronunciation is missing it is a three-criterion practice estimate, not an IELTS Speaking band; say in the summary that pronunciation was not assessed.\n")
	b.WriteString("Set estimatedScore.confidence to low or medium. A transcript-only judgement is never high confidence.\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "summary": "two or three sentences of evidence-based assessment, ending with the note that pronunciation was not assessed",
  "estimatedScore": {"value": %.1f, "confidence": "low"},
  "criteria": [{"name": "...", "score": <whole-band score>, "maxScore": 9, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "grammar", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))
	return b.String()
}

func transcriptUserPrompt(req SpokenTranscriptRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n\nInstructions given to the learner:\n%s\n", req.TaskName, req.Prompt)

	if text := strings.TrimSpace(req.ExpectedText); text != "" {
		fmt.Fprintf(&b, "\nThe text the learner was asked to say:\n%s\n", text)
	}
	b.WriteString(responseMaterialContext(req.SourceText, req.ReferenceAnswer))

	if measured := strings.TrimSpace(req.Delivery); measured != "" {
		fmt.Fprintf(&b, "\n%s\n", measured)
	} else if req.DurationSeconds > 0 {
		fmt.Fprintf(&b, "\nDelivery: %d words in %d seconds (about %d words per minute).\n",
			req.WordCount, req.DurationSeconds, speechRate(req.WordCount, req.DurationSeconds))
	}

	fmt.Fprintf(&b, "\nTranscript of what the learner said:\n%s\n", strings.TrimSpace(req.Transcript))
	b.WriteString("\nEvaluate it under the rules above.")
	return b.String()
}
