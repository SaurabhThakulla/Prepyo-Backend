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

// TestLiveWritingEvaluation runs one real evaluation against the configured
// provider. It costs a paid call, so it only runs when AI_LIVE_TEST is set:
//
//	AI_API_KEY=... AI_LIVE_TEST=1 go test ./internal/ai -run Live -v
//
// It exists because a provider swap is exactly the kind of change every unit
// test passes and production still fails: a wrong model id, a base URL missing
// its /v1, or a response shape that does not decode.
func TestLiveWritingEvaluation(t *testing.T) {
	if os.Getenv("AI_LIVE_TEST") == "" {
		t.Skip("set AI_LIVE_TEST=1 to run against the real provider")
	}

	cfg := &config.Config{
		AIBaseURL:        firstNonEmpty(os.Getenv("AI_BASE_URL"), "https://codecraftapi.com/v1"),
		AIAPIKey:         firstNonEmpty(os.Getenv("AI_API_KEY"), os.Getenv("CODE_CRAFT")),
		AIModels:         config.AIModels{Writing: firstNonEmpty(os.Getenv("AI_MODEL_WRITING"), "gpt-5.6-luna")},
		AIRequestTimeout: 90 * time.Second,
		AIMaxTokens:      8000,
	}
	if cfg.AIAPIKey == "" {
		t.Fatal("AI_API_KEY (or CODE_CRAFT) must be set for the live test")
	}

	g := NewGateway(cfg, slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})))
	if !g.Available() {
		t.Fatal("gateway reports unavailable with a key set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	evaluation, usage, err := g.EvaluateWriting(ctx, WritingRequest{
		Exam:     models.ExamPTE,
		TaskName: "Summarise written text",
		Prompt:   "Summarise the passage in one sentence of 5 to 75 words.",
		LearnerText: "The passage explain that remote work have grown quickly, " +
			"and companies is now rethinking office space, but many workers still " +
			"prefers some days in the office for collaboration.",
		MinScore: 10,
		MaxScore: 90,
	})
	if err != nil {
		t.Fatalf("EvaluateWriting against %s: %v", cfg.AIBaseURL, err)
	}

	if usage.Provider == "" || usage.Model != cfg.AIModels.Writing {
		t.Errorf("usage = %+v, want provider set and model %q", usage, cfg.AIModels.Writing)
	}
	if usage.PromptTokens == 0 || usage.CompletionTokens == 0 {
		t.Errorf("usage reported no tokens: %+v", usage)
	}
	if evaluation.Summary == "" {
		t.Error("evaluation has an empty summary")
	}

	t.Logf("provider=%s model=%s latency=%dms tokens=%d/%d",
		usage.Provider, usage.Model, usage.LatencyMS, usage.PromptTokens, usage.CompletionTokens)
	if evaluation.EstimatedScore != nil {
		t.Logf("score=%.1f", *evaluation.EstimatedScore)
	}
	t.Logf("summary=%s", evaluation.Summary)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
