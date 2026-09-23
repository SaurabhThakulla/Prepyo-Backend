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
const SpeakingPromptVersion = "speaking.v4"

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
	// SourceText is material the learner responds to rather than repeats: the
	// lecture to retell, the discussion to summarise, the situation to answer,
	// the question to reply to. Never set together with ExpectedText.
	SourceText string
	// ReferenceAnswer is the item's model answer, a calibration point for
	// content rather than wording the learner has to match.
	ReferenceAnswer string
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

	if !g.SpeakingAvailable() {
		return models.Evaluation{}, Usage{}, fmt.Errorf("%w: no audio provider is configured", ErrUnavailable)
	}

	// If the configured audio provider is a transcription service (such as Groq Whisper),
	// transcribe the audio recording first and then evaluate the transcript via the text provider.
	if strings.Contains(strings.ToLower(g.audio.baseURL), "groq") || strings.Contains(strings.ToLower(g.models.Speaking), "whisper") {
		transcript, transcribeUsage, err := g.transcribeAudio(ctx, req)
		if err != nil {
			return models.Evaluation{}, transcribeUsage, err
		}
		if transcript == "" {
			return models.Evaluation{}, transcribeUsage, fmt.Errorf("%w: could not transcribe speech from audio", ErrBadOutput)
		}

		eval, evalUsage, err := g.EvaluateSpokenTranscript(ctx, SpokenTranscriptRequest{
			Exam:            req.Exam,
			TaskName:        req.TaskName,
			Prompt:          req.Prompt,
			ExpectedText:    req.ExpectedText,
			SourceText:      req.SourceText,
			ReferenceAnswer: req.ReferenceAnswer,
			Transcript:      transcript,
			DurationSeconds: req.DurationSeconds,
			WordCount:       len(strings.Fields(transcript)),
			MinScore:        req.MinScore,
			MaxScore:        req.MaxScore,
		})
		evalUsage.add(transcribeUsage)
		return eval, evalUsage, err
	}

	return g.evaluateSpeakingMultimodal(ctx, req)
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
	if isIELTS(req.Exam) {
		return ieltsSpeakingSystemPrompt(req)
	}

	var b strings.Builder
	b.WriteString("You are a strict, certified senior ")
	b.WriteString(string(req.Exam))
	b.WriteString(" speaking examiner. Your duty is to provide authentic, rigorous exam scoring that mirrors real British Council/IDP and Pearson automated testing center standards.\n\n")
	b.WriteString("You are given one audio recording of the learner's spoken response.\n\n")
	b.WriteString("Evaluation Rigor & Strictness Guidelines:\n")
	b.WriteString("- CRITICAL: Grade strictly and objectively against official published rubric band descriptors. DO NOT inflate scores, flatter the candidate, or be lenient.\n")
	b.WriteString("- Real Exam Benchmark: Average test-takers score between Band 5.5 and 6.5 (PTE 50-64). Reserve Band 7.5+ (PTE 75+) strictly for effortless, native-like delivery with flawless rhythm, natural intonation, and sophisticated lexical/grammatical control.\n")
	b.WriteString("- Scrutinize Fluency & Delivery: Penalize unnatural pauses, mid-sentence hesitations, repetitions, self-corrections, and filler words ('um', 'uh', 'like'). Pauses longer than 2 seconds or robotic/staccato pacing must cap fluency strictly below Band 6.0 (PTE 55).\n")
	b.WriteString("- Scrutinize Pronunciation & Intonation: Penalize non-native phonemic distortion, misplaced word stress, missing consonant clusters or word endings (-ed, -s), flat monotone delivery, and mother-tongue phonetic interference that impairs intelligibility.\n")
	b.WriteString("- Scrutinize Lexical Resource & Grammar: Penalize basic vocabulary, repetitive sentence structures, missing articles, tense errors, and subject-verb disagreements.\n")
	b.WriteString("- First transcribe what you actually hear into `transcript`, verbatim. Include the learner's own errors, repetitions and false starts. Do not tidy them up.\n")
	b.WriteString("- If the recording is silent, unintelligible, or contains no speech, set `transcript` to \"\" and estimatedScore.value to null, and say so plainly in the summary.\n")
	b.WriteString("- Judge only this recording. Never quote or invent words the learner did not say.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from `transcript` into `original` and provide an elevated, high-scoring correction.\n")
	b.WriteString("- Judge pronunciation, fluency and content from the audio itself: hesitation, pace, stress, intonation and intelligibility. Do not score pronunciation from the transcript alone.\n")

	if strings.TrimSpace(req.ExpectedText) != "" {
		b.WriteString("- The learner was given a fixed text to say. Compare what you heard against it and treat omissions, substitutions and additions as content errors.\n")
	}
	b.WriteString(responseMaterialRules(req.Exam, req.SourceText, req.ReferenceAnswer))

	if req.Exam == models.ExamPTE {
		b.WriteString("- Use the published PTE speaking criteria for this task type: content, oral fluency and pronunciation.\n")
		b.WriteString("- CRITICAL FOR PTE: estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 70.0, 79, 85). DO NOT output 0-9 IELTS band numbers.\n")
	} else {
		b.WriteString("- For a scored IELTS response return exactly four criteria: Fluency and Coherence, Lexical Resource, Grammatical Range and Accuracy, and Pronunciation. Each has maxScore 9 and nonempty evidence-based feedback citing specific weaknesses. The estimate must be the equally weighted criterion mean rounded to the nearest half band. A single recorded task is only a practice estimate, not a full speaking test band.\n")
		b.WriteString("- CRITICAL FOR IELTS: estimatedScore.value MUST be on the 0.0-9.0 IELTS band scale in 0.5 steps (e.g. 6.5, 7.0, 7.5).\n")
	}

	b.WriteString("- Set estimatedScore.confidence to low, medium or high based on how much the recording gives you. A very short recording is low confidence.\n")
	b.WriteString("- criteria[].maxScore is the maximum the published rubric gives that one criterion. It is not the exam's overall scale.\n")
	b.WriteString("- The example below already uses this exam's scale. Copy its shape, never its numbers.\n\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "transcript": "exactly what the learner said",
  "summary": "two or three sentences of concise, rigorous assessment",
  "estimatedScore": {"value": %.1f, "confidence": "medium"},
  "criteria": [{"name": "...", "score": <this criterion's score>, "maxScore": <this criterion's maximum>, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "pronunciation", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))
	return b.String()
}

// ieltsSpeakingSystemPrompt scores one recorded IELTS answer from its audio.
func ieltsSpeakingSystemPrompt(req SpeakingRequest) string {
	var b strings.Builder
	b.WriteString("You are an experienced IELTS Speaking assessor giving practice feedback on one recorded answer. You are not an official IELTS examiner, and your estimate is not an official score.\n\n")
	b.WriteString("You are given one audio recording of the learner's spoken response.\n")
	b.WriteString("- First transcribe what you actually hear into `transcript`, verbatim. Include the learner's own errors, repetitions and false starts. Do not tidy them up.\n")
	b.WriteString("- If the recording is silent, unintelligible, or contains no speech, set `transcript` to \"\" and estimatedScore.value to null, and say so plainly in the summary.\n")
	b.WriteString("- Judge pronunciation and fluency from the audio itself, not from the transcript alone.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from `transcript` into `original` and give an improved version with a short explanation. Never quote or invent words the learner did not say.\n\n")
	b.WriteString(ieltsSpeakingGuidance(req.TaskName, true))
	b.WriteString(responseMaterialRules(req.Exam, req.SourceText, req.ReferenceAnswer))
	fmt.Fprintf(&b, "\nReturn exactly four criteria: %s. Each has maxScore 9, a whole-band score, and feedback that cites evidence from the recording. ", ieltsSpeakingCriteria(true))
	b.WriteString("estimatedScore.value is the mean of the four criteria rounded to the nearest half band (.25 rounds up to the next half band, .75 up to the next whole band). It is a practice estimate for this answer, not a Speaking band.\n")
	b.WriteString("Set estimatedScore.confidence to low, medium or high according to how much the recording gives you. A very short recording is low confidence.\n")
	b.WriteString("The example below already uses this exam's scale. Copy its shape, never its numbers.\n\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "transcript": "exactly what the learner said",
  "summary": "two or three sentences of evidence-based assessment",
  "estimatedScore": {"value": %.1f, "confidence": "medium"},
  "criteria": [{"name": "...", "score": <whole-band score>, "maxScore": 9, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "grammar", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))
	return b.String()
}

// responseMaterialRules tells the model how to use the material behind a task
// that is answered in the learner's own words. Without it, a lecture or a
// question handed over as "text" reads as something to repeat, and a correct
// retelling or a one-word answer is marked down for every word it leaves out.
func responseMaterialRules(exam models.ExamType, source, reference string) string {
	var b strings.Builder
	if strings.TrimSpace(source) != "" {
		b.WriteString("- The learner was not asked to repeat the material supplied. They were asked to respond to it: retell, summarise, answer or reply in their own words. Judge content by how accurately and fully the response does what the task asks. Never penalise paraphrase, and never count words of the material the learner did not repeat as omissions.\n")
	}
	// IELTS Speaking has no content-accuracy criterion: a sample answer is only
	// an illustration, and an answer with different ideas is not wrong.
	if isIELTS(exam) {
		if strings.TrimSpace(reference) != "" {
			b.WriteString("- A sample answer is supplied only to illustrate one possible response. Different ideas or opinions are equally valid; never score an answer by how closely its content matches the sample.\n")
		}
		return b.String()
	}
	if strings.TrimSpace(reference) != "" {
		b.WriteString("- A reference answer is supplied for calibration. It shows the content a strong response covers; any wording that conveys the same content earns full content credit. For a short-answer question, any answer with the same meaning as the reference is correct, and an answer with a different meaning is wrong however fluent it is.\n")
	}
	return b.String()
}

// responseMaterialContext is the user-message half of responseMaterialRules.
func responseMaterialContext(source, reference string) string {
	var b strings.Builder
	if text := strings.TrimSpace(source); text != "" {
		fmt.Fprintf(&b, "\nMaterial the learner responded to (not a text to repeat):\n%s\n", text)
	}
	if text := strings.TrimSpace(reference); text != "" {
		fmt.Fprintf(&b, "\nReference answer (for calibration, not required wording):\n%s\n", text)
	}
	return b.String()
}

func speakingUserPrompt(req SpeakingRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n\nInstructions given to the learner:\n%s\n", req.TaskName, req.Prompt)

	if text := strings.TrimSpace(req.ExpectedText); text != "" {
		fmt.Fprintf(&b, "\nThe text the learner was asked to say:\n%s\n", text)
	}
	b.WriteString(responseMaterialContext(req.SourceText, req.ReferenceAnswer))
	if req.DurationSeconds > 0 {
		fmt.Fprintf(&b, "\nRecording length: %d seconds.\n", req.DurationSeconds)
	}

	b.WriteString("\nThe learner's recording is attached. Transcribe it, then evaluate it.")
	return b.String()
}
