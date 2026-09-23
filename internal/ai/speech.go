package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrSpeechUnavailable means server-side speech cannot run: no audio provider,
// or the provider refused the model (for example, its terms have not been
// accepted on the account). Callers fall back to the browser's own voice.
var ErrSpeechUnavailable = errors.New("speech synthesis unavailable")

// MaxSpeechChars is the most text one speech request may carry. Groq's Orpheus
// model accepts 200 characters per request.
const MaxSpeechChars = 200

// SpeechAvailable reports whether the server can speak a script itself.
func (g *Gateway) SpeechAvailable() bool {
	return g.audio.configured() && g.ttsModel != "" && len(g.ttsVoices) > 0
}

// SpeechModel is the model speech clips are made with; it is part of a clip's
// cache key, so changing it never serves audio from the old model.
func (g *Gateway) SpeechModel() string { return g.ttsModel }

// SpeechVoices are the voices handed out to speakers, in order.
func (g *Gateway) SpeechVoices() []string { return g.ttsVoices }

// Synthesize speaks one short piece of text (at most MaxSpeechChars) in a
// voice and returns WAV audio.
func (g *Gateway) Synthesize(ctx context.Context, voice, text string) ([]byte, error) {
	if !g.SpeechAvailable() {
		return nil, fmt.Errorf("%w: no audio provider is configured", ErrSpeechUnavailable)
	}
	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > MaxSpeechChars {
		return nil, fmt.Errorf("speech input must be 1-%d characters, got %d", MaxSpeechChars, len([]rune(text)))
	}

	body, err := json.Marshal(map[string]string{
		"model":           g.ttsModel,
		"voice":           voice,
		"input":           text,
		"response_format": "wav",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.audio.baseURL+"/audio/speech", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create speech request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.audio.apiKey)

	started := time.Now()
	res, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		g.log.Error("speech provider error", "status", res.StatusCode, "body", string(raw[:min(len(raw), 500)]))
		// A refused model or a bad request will not succeed on retry; say so
		// plainly so the browser falls back to its own voice.
		if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusForbidden || res.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: provider returned %d", ErrSpeechUnavailable, res.StatusCode)
		}
		return nil, fmt.Errorf("%w: speech provider returned %d", ErrUnavailable, res.StatusCode)
	}
	g.log.Info("speech synthesised", "model", g.ttsModel, "voice", voice, "chars", len(text), "ms", time.Since(started).Milliseconds())
	return raw, nil
}

// Transcribe turns one recording into text on the audio provider (Whisper on
// Groq). format is the file extension of the audio, such as "wav" or "mp3".
func (g *Gateway) Transcribe(ctx context.Context, audio []byte, format string) (string, Usage, error) {
	if !g.audio.configured() {
		return "", Usage{}, fmt.Errorf("%w: no audio provider is configured", ErrUnavailable)
	}
	return g.transcribeAudio(ctx, SpeakingRequest{
		AudioBase64: base64.StdEncoding.EncodeToString(audio),
		AudioFormat: format,
	})
}
