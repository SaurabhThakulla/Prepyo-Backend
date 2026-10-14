package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/scoring"
)

// whisperWord is one word of a verbose_json transcription, timed in seconds.
type whisperWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// deliveryFromWords measures pauses from Whisper's word timings: the gap between
// one word's end and the next one's start. Silence before the first word and
// after the last is preparation and stopping, not hesitation, so it is skipped.
func deliveryFromWords(words []whisperWord, durationSeconds float64) scoring.Delivery {
	if len(words) == 0 || durationSeconds <= 0 {
		return scoring.Delivery{}
	}
	var pauses int
	var longest, silent float64
	for i := 1; i < len(words); i++ {
		gap := words[i].Start - words[i-1].End
		if gap > scoring.PauseThresholdSeconds {
			pauses++
			silent += gap
		}
		longest = math.Max(longest, gap)
	}
	spoken := words[len(words)-1].End - words[0].Start - silent
	return scoring.Delivery{
		DurationSeconds:     max(1, int(math.Round(durationSeconds))),
		WordsSpoken:         len(words),
		PauseCount:          pauses,
		LongestPauseSeconds: math.Round(longest*10) / 10,
		SpeakingRatio:       math.Min(1, math.Max(0, spoken/durationSeconds)),
	}
}

// transcription is what Whisper heard, with pauses measured from its word timings.
type transcription struct {
	Text     string
	Delivery scoring.Delivery
	// Marked is Text rebuilt from the timed words with scoring.PauseMark at each pause.
	Marked string
}

// markPauses joins the timed words, putting scoring.PauseMark where a gap is a pause.
func markPauses(words []whisperWord) string {
	var b strings.Builder
	for i, w := range words {
		switch {
		case i == 0:
		case w.Start-words[i-1].End > scoring.PauseThresholdSeconds:
			b.WriteString(scoring.PauseMark)
		default:
			b.WriteString(" ")
		}
		b.WriteString(strings.TrimSpace(w.Word))
	}
	return b.String()
}

// TranscribeMarked is TranscribeTimed with pauses marked in the text; without word timings it is the plain text.
func (g *Gateway) TranscribeMarked(ctx context.Context, audio []byte, format string) (string, scoring.Delivery, Usage, error) {
	if !g.audio.configured() {
		return "", scoring.Delivery{}, Usage{}, fmt.Errorf("%w: no audio provider is configured", ErrUnavailable)
	}
	heard, usage, err := g.transcribeAudio(ctx, SpeakingRequest{AudioBase64: base64.StdEncoding.EncodeToString(audio), AudioFormat: format})
	if heard.Marked == "" {
		heard.Marked = heard.Text
	}
	return heard.Marked, heard.Delivery, usage, err
}

// TranscribeTimed is Transcribe plus the pauses Whisper's word timings show.
func (g *Gateway) TranscribeTimed(ctx context.Context, audio []byte, format string) (string, scoring.Delivery, Usage, error) {
	if !g.audio.configured() {
		return "", scoring.Delivery{}, Usage{}, fmt.Errorf("%w: no audio provider is configured", ErrUnavailable)
	}
	heard, usage, err := g.transcribeAudio(ctx, SpeakingRequest{AudioBase64: base64.StdEncoding.EncodeToString(audio), AudioFormat: format})
	return heard.Text, heard.Delivery, usage, err
}

// Transcriber is what the mocks transcribe with; ai.Gateway implements it.
type Transcriber interface {
	Transcribe(ctx context.Context, audio []byte, format string) (string, Usage, error)
}

// TranscribeWithPauses also returns measured pauses when the listener can
// time words (the gateway can); otherwise the delivery is left unmeasured.
func TranscribeWithPauses(ctx context.Context, l Transcriber, audio []byte, format string) (string, scoring.Delivery, error) {
	if timed, ok := l.(interface {
		TranscribeTimed(context.Context, []byte, string) (string, scoring.Delivery, Usage, error)
	}); ok {
		text, delivery, _, err := timed.TranscribeTimed(ctx, audio, format)
		return text, delivery, err
	}
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
