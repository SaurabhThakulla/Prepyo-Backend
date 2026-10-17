package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prepyo/backend/internal/scoring"
)

// TranscriptionPromptVersion is stamped on usage for a recording turned into text by Gemini.
const TranscriptionPromptVersion = "audio.transcription.gemini.v1"

// transcription is what the audio model heard. It has no word timings, so
// pauses come only from the browser's own measurement.
type transcription struct {
	Text string
}

// transcriptionPrompt asks for the words as said: the learner's mistakes are what gets marked, so none may be fixed.
const transcriptionPrompt = `Transcribe this recording of an English test answer exactly as spoken.
- Write every word the speaker says, in order, including mistakes, wrong word endings, repetitions, false starts and fillers such as um, uh, er.
- Never correct grammar, pronunciation or word choice, never add a word that was not said, never finish a cut-off sentence.
- Write a word you cannot make out as the closest word you heard; do not guess at what was meant.
- Write numbers as words, as spoken.
- If there is no speech, return an empty transcript.
Reply with JSON only: {"transcript": "<what was said>"}`

// transcribeAudio turns one recording into text with the audio model (Gemini on OpenRouter).
// It is never shown the text the learner was asked to say, so it cannot fill in what was missed.
func (g *Gateway) transcribeAudio(ctx context.Context, req SpeakingRequest) (transcription, Usage, error) {
	if !AudioFormats[req.AudioFormat] {
		return transcription{}, Usage{}, fmt.Errorf("%w: audio format %q is not supported", ErrBadOutput, req.AudioFormat)
	}
	messages := []chatMessage{
		{Role: "system", Content: transcriptionPrompt},
		{Role: "user", Content: []contentPart{
			{Type: "text", Text: "Transcribe this recording."},
			{Type: "input_audio", Audio: &audioInput{Data: req.AudioBase64, Format: req.AudioFormat}},
		}},
	}
	raw, usage, err := g.complete(ctx, g.audio, g.models.Speaking, TranscriptionPromptVersion, messages, true)
	if err != nil {
		return transcription{}, usage, err
	}
	var reply struct {
		Transcript string `json:"transcript"`
	}
	if err := json.Unmarshal([]byte(unfence(raw)), &reply); err != nil {
		return transcription{}, usage, fmt.Errorf("%w: transcription was not valid JSON", ErrBadOutput)
	}
	return transcription{Text: strings.TrimSpace(reply.Transcript)}, usage, nil
}

// TranscribeMarked is the transcript with no pause marks: Gemini returns no word timings.
func (g *Gateway) TranscribeMarked(ctx context.Context, audio []byte, format string) (string, scoring.Delivery, Usage, error) {
	return g.TranscribeTimed(ctx, audio, format)
}

// TranscribeTimed is Transcribe with an unmeasured delivery; the browser's measurement is used instead.
func (g *Gateway) TranscribeTimed(ctx context.Context, audio []byte, format string) (string, scoring.Delivery, Usage, error) {
	if !g.audio.configured() {
		return "", scoring.Delivery{}, Usage{}, fmt.Errorf("%w: no audio provider is configured", ErrUnavailable)
	}
	heard, usage, err := g.transcribeAudio(ctx, SpeakingRequest{AudioBase64: base64.StdEncoding.EncodeToString(audio), AudioFormat: format})
	return heard.Text, scoring.Delivery{}, usage, err
}

// Transcriber is what the mocks transcribe with; ai.Gateway implements it.
type Transcriber interface {
	Transcribe(ctx context.Context, audio []byte, format string) (string, Usage, error)
}

// TranscribeWithPauses transcribes; pauses are left unmeasured, for the browser's measurement to fill.
func TranscribeWithPauses(ctx context.Context, l Transcriber, audio []byte, format string) (string, scoring.Delivery, error) {
	text, _, err := l.Transcribe(ctx, audio, format)
	return text, scoring.Delivery{}, err
}

// PauseNote describes measured pauses in one line for a prompt, or is empty.
func PauseNote(d scoring.Delivery) string {
	if !d.Measured() {
		return ""
	}
	return fmt.Sprintf("%d pauses over %.1fs, longest %.1fs, speaking %.0f%% of the time",
		d.PauseCount, scoring.PauseThresholdSeconds, d.LongestPauseSeconds, d.SpeakingRatio*100)
}
