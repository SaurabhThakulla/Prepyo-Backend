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
const SpeakingTranscriptPromptVersion = "speaking.transcript.v1"

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
	var b strings.Builder
	b.WriteString("You are an experienced ")
	b.WriteString(string(req.Exam))
	b.WriteString(" speaking examiner giving practice feedback to a learner in Nepal.\n\n")
	b.WriteString("You are given a transcript of the learner's spoken answer, produced by speech recognition on their device. You cannot hear the recording.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Judge only what the transcript shows: content, coherence, vocabulary and grammar.\n")
	b.WriteString("- NEVER score or comment on pronunciation, accent, intonation or stress. You did not hear them. Do not include a pronunciation criterion.\n")
	b.WriteString("- Speech recognition makes its own mistakes. A single odd word is more likely a recognition error than a learner error: ignore it unless the pattern repeats.\n")
	b.WriteString("- Base any comment on delivery only on the speaking rate and length given below, never on guesswork.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from the transcript into `original`.\n")
	b.WriteString("- Say in the summary that this is a provisional estimate from a transcript, and that pronunciation was not assessed.\n")

	if strings.TrimSpace(req.ExpectedText) != "" {
		b.WriteString("- The learner was given a fixed text to say. Compare the transcript against it and treat omissions and substitutions as content errors, allowing for recognition error.\n")
	}

	if req.Exam == models.ExamPTE {
		b.WriteString("- Return exactly two criteria: Content and Oral Fluency. Each has maxScore 90 and evidence-based feedback. Do not return a Pronunciation criterion.\n")
		b.WriteString("- CRITICAL FOR PTE: estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 79). DO NOT output 0-9 IELTS band numbers.\n")
	} else {
		b.WriteString("- Return exactly three criteria: Fluency and Coherence, Lexical Resource, and Grammatical Range and Accuracy. Each has maxScore 9 and evidence-based feedback. Do not return a Pronunciation criterion.\n")
		b.WriteString("- The estimate must be the equally weighted mean of those three criteria, rounded to the nearest half band.\n")
		b.WriteString("- CRITICAL FOR IELTS: estimatedScore.value MUST be on the 0.0-9.0 IELTS band scale in 0.5 steps (e.g. 6.0, 6.5, 7.0).\n")
	}

	b.WriteString("- Set estimatedScore.confidence to low or medium. A transcript-only judgement is never high confidence.\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "summary": "two or three sentences, ending with the provisional-estimate caveat",
  "estimatedScore": {"value": %.1f, "confidence": "low"},
  "criteria": [{"name": "...", "score": <this criterion's score>, "maxScore": <this criterion's maximum>, "feedback": "..."}],
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
