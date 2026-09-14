package ai

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/pkg/config"
)

func TestLiveTutorEvaluation(t *testing.T) {
	if os.Getenv("AI_LIVE_TEST") == "" {
		t.Skip("set AI_LIVE_TEST=1 to run against the real provider")
	}

	cfg := &config.Config{
		AIBaseURL: "https://codecraftapi.com/v1",
		AIAPIKey:  os.Getenv("AI_API_KEY"),
		AIModels: config.AIModels{
			Tutoring: "gemini-3.8-flash-high", // Intentional bad model to test fallback!
		},
		AIRequestTimeout: 30 * time.Second,
		AIMaxTokens:      2000,
	}

	g := NewGateway(cfg, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reply, usage, err := g.Tutor(ctx, TutorRequest{
		Exam: models.ExamPTE,
		Messages: []TutorMessage{
			{Role: "user", Content: "Hello! Give me one PTE tip in 10 words."},
		},
	})
	if err != nil {
		t.Fatalf("Tutor call failed: %v", err)
	}

	t.Logf("SUCCESS! Model used: %s, Latency: %dms, Reply: %s", usage.Model, usage.LatencyMS, reply)
}
