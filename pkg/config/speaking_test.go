package config

import (
	"testing"
)

func TestSpeakingEvaluatedFollowsAudioConfiguration(t *testing.T) {
	t.Setenv("AI_API_KEY", "text-key")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/prepyo?sslmode=disable")

	t.Run("off when no audio provider is set", func(t *testing.T) {
		// The deployed shape of the bug: docker-compose passes AI_AUDIO_* through
		// as empty, the key silently inherits the text provider's, and speaking
		// looks available while nothing can score a recording.
		t.Setenv("AI_AUDIO_BASE_URL", "")
		t.Setenv("AI_AUDIO_API_KEY", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if cfg.SpeakingEvaluated {
			t.Fatal("SpeakingEvaluated = true with no audio provider configured")
		}
	})

	t.Run("on when an audio provider is set", func(t *testing.T) {
		t.Setenv("AI_AUDIO_BASE_URL", "https://openrouter.ai/api/v1")
		t.Setenv("AI_AUDIO_API_KEY", "audio-key")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		if !cfg.SpeakingEvaluated {
			t.Fatal("SpeakingEvaluated = false with an audio provider configured")
		}
	})
}
