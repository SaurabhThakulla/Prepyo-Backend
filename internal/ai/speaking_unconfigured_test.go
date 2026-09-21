package ai

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/pkg/config"
)

// Without an audio provider a recording cannot be scored, and the answer must
// be "unavailable" rather than a band invented by a text model that never
// heard it. Deployed without AI_AUDIO_*, the old fallback did exactly that.
func TestSpeakingWithoutAudioProviderIsUnavailable(t *testing.T) {
	cfg := &config.Config{
		AIBaseURL:      "https://example.invalid/v1",
		AIAPIKey:       "text-key",
		AIAudioBaseURL: "https://example.invalid/v1",
		AIAudioAPIKey:  "text-key",
		// What config.Load() computes when no AI_AUDIO_* var is set.
		SpeakingEvaluated: false,
		AIModels:          config.AIModels{Writing: "text-model", Speaking: "audio-model"},
		AIRequestTimeout:  5 * time.Second,
	}
	g := NewGateway(cfg, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})))

	if g.SpeakingAvailable() {
		t.Fatal("SpeakingAvailable() = true with no audio provider configured")
	}

	_, _, err := g.EvaluateSpeaking(context.Background(), SpeakingRequest{
		Exam:            models.ExamIELTS,
		TaskName:        "IELTS Speaking Part 1",
		AudioBase64:     base64.StdEncoding.EncodeToString([]byte("not really audio")),
		AudioFormat:     "wav",
		DurationSeconds: 20,
		MaxScore:        9,
	})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("EvaluateSpeaking error = %v, want ErrUnavailable", err)
	}
}
