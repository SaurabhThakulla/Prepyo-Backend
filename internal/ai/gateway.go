// Package ai is the only place in the backend that talks to a model provider.
//
// Product modules ask for a capability ("evaluate this essay"), never for a
// specific model. Model choice, retries, timeouts and usage accounting all live
// here, so swapping providers touches this package and nothing else.
//
// When no provider is configured the gateway returns ErrUnavailable. It does
// not fall back to a canned score: a made-up band that looks real is worse for
// a learner than an honest "not available".
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/prepyo/backend/pkg/config"
)

var (
	// ErrUnavailable means the gateway cannot reach a provider right now.
	ErrUnavailable = errors.New("ai provider unavailable")
	// ErrBadOutput means the model replied with something that failed
	// validation. The caller must not persist anything.
	ErrBadOutput = errors.New("ai response failed validation")
)

const (
	openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	// maxAttempts covers one retry. Evaluations run while a learner waits, so
	// a long retry chain is worse than failing quickly.
	maxAttempts = 2
)

type Gateway struct {
	client  *http.Client
	apiKey  string
	models  config.AIModels
	enabled bool
	log     *slog.Logger
}

func NewGateway(cfg *config.Config, log *slog.Logger) *Gateway {
	return &Gateway{
		client:  &http.Client{Timeout: cfg.AIRequestTimeout},
		apiKey:  cfg.OpenRouterAPIKey,
		models:  cfg.AIModels,
		enabled: cfg.AIEnabled(),
		log:     log,
	}
}

func (g *Gateway) Available() bool { return g.enabled }

// Usage is what one provider call cost. Token counts come from the provider
// response, never from an estimate, so cost reporting reflects reality.
type Usage struct {
	Provider         string
	Model            string
	PromptVersion    string
	PromptTokens     int
	CompletionTokens int
	LatencyMS        int
}

// chatRequest is the OpenRouter payload. Kept unexported: no other package
// should be able to construct a raw model call.
type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat *responseFmt  `json:"response_format,omitempty"`
	Temperature    float64       `json:"temperature"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFmt struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// complete sends one chat request and returns the assistant text plus usage.
func (g *Gateway) complete(ctx context.Context, model, promptVersion string, messages []chatMessage, wantJSON bool) (string, Usage, error) {
	if !g.enabled {
		return "", Usage{}, ErrUnavailable
	}

	payload := chatRequest{
		Model:    model,
		Messages: messages,
		// Low temperature: evaluation should be as repeatable as the provider
		// allows, so two learners with similar work get similar feedback.
		Temperature: 0.2,
		MaxTokens:   2000,
	}
	if wantJSON {
		payload.ResponseFormat = &responseFmt{Type: "json_object"}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", Usage{}, fmt.Errorf("encode ai request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		started := time.Now()
		text, usage, err := g.send(ctx, body, model, promptVersion, started)
		if err == nil {
			return text, usage, nil
		}
		lastErr = err

		// The caller's context being done means the learner is gone or the
		// deadline passed; another attempt would only waste a call.
		if ctx.Err() != nil {
			return "", Usage{}, ErrUnavailable
		}
		if attempt < maxAttempts {
			g.log.Warn("ai request failed, retrying", "model", model, "attempt", attempt, "error", err)
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
	}

	g.log.Error("ai request failed", "model", model, "error", lastErr)
	return "", Usage{}, ErrUnavailable
}

func (g *Gateway) send(ctx context.Context, body []byte, model, promptVersion string, started time.Time) (string, Usage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	res, err := g.client.Do(req)
	if err != nil {
		return "", Usage{}, err
	}
	defer res.Body.Close()

	// Cap the read so a malformed or hostile response cannot exhaust memory.
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", Usage{}, err
	}
	if res.StatusCode != http.StatusOK {
		return "", Usage{}, fmt.Errorf("provider returned %d", res.StatusCode)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", Usage{}, fmt.Errorf("decode provider response: %w", err)
	}
	if parsed.Error != nil {
		return "", Usage{}, fmt.Errorf("provider error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", Usage{}, errors.New("provider returned no choices")
	}
	// A truncated reply cannot be trusted to be complete JSON.
	if parsed.Choices[0].FinishReason == "length" {
		return "", Usage{}, errors.New("provider response was truncated")
	}

	usage := Usage{
		Provider:         "openrouter",
		Model:            model,
		PromptVersion:    promptVersion,
		PromptTokens:     parsed.Usage.PromptTokens,
		CompletionTokens: parsed.Usage.CompletionTokens,
		LatencyMS:        int(time.Since(started).Milliseconds()),
	}
	return parsed.Choices[0].Message.Content, usage, nil
}
