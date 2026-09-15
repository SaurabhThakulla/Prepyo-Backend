// Package ai handles communication with AI model providers for evaluation and tutoring.
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
	"net/url"
	"strings"
	"time"

	"github.com/prepyo/backend/pkg/config"
)

var (
	// ErrUnavailable means the gateway cannot reach a provider right now.
	ErrUnavailable = errors.New("ai provider unavailable")
	// ErrBadOutput means the model replied with something that failed
	// validation. The caller must not persist anything.
	ErrBadOutput = errors.New("ai response failed validation")

	// errPermanent marks a non-retryable provider error.
	errPermanent = errors.New("provider rejected the request")
)

// retryable reports whether an HTTP status code indicates a retryable error.
func retryable(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests:
		return true
	default:
		return status >= 500
	}
}

const maxAttempts = 2

type provider struct {
	name    string
	baseURL string
	apiKey  string
}

func (p provider) configured() bool { return p.apiKey != "" && p.baseURL != "" }

func (p provider) completionsURL() string { return p.baseURL + "/chat/completions" }

func newProvider(baseURL, apiKey string) provider {
	name := baseURL
	if u, err := url.Parse(baseURL); err == nil && u.Host != "" {
		name = u.Host
	}
	return provider{name: name, baseURL: baseURL, apiKey: apiKey}
}

type Gateway struct {
	client    *http.Client
	text      provider
	audio     provider
	models    config.AIModels
	maxTokens int
	log       *slog.Logger
}

func NewGateway(cfg *config.Config, log *slog.Logger) *Gateway {
	return &Gateway{
		client:    &http.Client{Timeout: cfg.AIRequestTimeout},
		text:      newProvider(cfg.AIBaseURL, cfg.AIAPIKey),
		audio:     newProvider(cfg.AIAudioBaseURL, cfg.AIAudioAPIKey),
		models:    cfg.AIModels,
		maxTokens: cfg.AIMaxTokens,
		log:       log,
	}
}

// Available reports whether text evaluation and tutoring can run.
func (g *Gateway) Available() bool { return g.text.configured() }

// SpeakingAvailable reports whether the audio provider is configured.
func (g *Gateway) SpeakingAvailable() bool { return g.audio.configured() }

// Usage records token counts and latency for a provider call.
type Usage struct {
	Provider         string
	Model            string
	PromptVersion    string
	PromptTokens     int
	CompletionTokens int
	LatencyMS        int
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []chatMessage `json:"messages"`
	ResponseFormat *responseFmt  `json:"response_format,omitempty"`
	Temperature    float64       `json:"temperature"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// contentPart is one element of a multimodal turn.
type contentPart struct {
	Type  string      `json:"type"`
	Text  string      `json:"text,omitempty"`
	Audio *audioInput `json:"input_audio,omitempty"`
}

type audioInput struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type responseFmt struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func fallbackCandidates(primary string, isAudio bool) []string {
	primary = strings.TrimSpace(primary)
	var pool []string
	if isAudio {
		pool = []string{primary, "gemini-3.7-flash", "gemini-3.6-flash"}
	} else {
		pool = []string{primary, "gpt-5.6-luna", "claude-sonnet-5", "gpt-5.5", "gemini-3.7-flash"}
	}

	seen := make(map[string]bool)
	var result []string
	for _, m := range pool {
		m = strings.TrimSpace(m)
		if m == "" || strings.Contains(strings.ToLower(m), "gemini-3.8") || seen[m] {
			continue
		}
		seen[m] = true
		result = append(result, m)
	}
	if len(result) == 0 {
		if isAudio {
			result = []string{"gemini-3.7-flash"}
		} else {
			result = []string{"gpt-5.6-luna"}
		}
	}
	return result
}

// complete sends one chat request and returns the assistant text plus usage.
// If the primary model fails or is out of capacity, it automatically falls back
// to other available models in the pool.
func (g *Gateway) complete(ctx context.Context, p provider, model, promptVersion string, messages []chatMessage, wantJSON bool) (string, Usage, error) {
	if !p.configured() {
		return "", Usage{}, ErrUnavailable
	}

	isAudio := false
	for _, m := range messages {
		if parts, ok := m.Content.([]contentPart); ok {
			for _, part := range parts {
				if part.Audio != nil {
					isAudio = true
					break
				}
			}
		}
	}

	candidates := fallbackCandidates(model, isAudio)

	var lastErr error
	for _, candidateModel := range candidates {
		payload := chatRequest{
			Model:       candidateModel,
			Messages:    messages,
			Temperature: 0.2,
			MaxTokens:   g.maxTokens,
		}
		if wantJSON {
			payload.ResponseFormat = &responseFmt{Type: "json_object"}
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return "", Usage{}, fmt.Errorf("encode ai request: %w", err)
		}

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			started := time.Now()
			text, usage, err := g.send(ctx, p, body, candidateModel, promptVersion, started)
			if err == nil {
				return text, usage, nil
			}
			lastErr = err

			if ctx.Err() != nil {
				return "", Usage{}, ErrUnavailable
			}
			if errors.Is(err, errPermanent) {
				break
			}
			if attempt < maxAttempts {
				g.log.Warn("ai request failed, retrying", "model", candidateModel, "attempt", attempt, "error", err)
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			}
		}

		g.log.Warn("ai model failed, attempting fallback candidate", "failedModel", candidateModel, "error", lastErr)
	}

	g.log.Error("all ai candidate models failed", "primaryModel", model, "error", lastErr)
	return "", Usage{}, fmt.Errorf("%w: %w", ErrUnavailable, lastErr)
}

func (g *Gateway) send(ctx context.Context, p provider, body []byte, model, promptVersion string, started time.Time) (string, Usage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.completionsURL(), bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	res, err := g.client.Do(req)
	if err != nil {
		return "", Usage{}, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", Usage{}, err
	}
	var parsed chatResponse
	decodeErr := json.Unmarshal(raw, &parsed)

	if res.StatusCode != http.StatusOK {
		detail := "no detail"
		if decodeErr == nil && parsed.Error != nil && parsed.Error.Message != "" {
			detail = parsed.Error.Message
		}

		err := fmt.Errorf("provider returned %d: %s", res.StatusCode, detail)
		if !retryable(res.StatusCode) {
			return "", Usage{}, fmt.Errorf("%w: %w", errPermanent, err)
		}
		return "", Usage{}, err
	}

	if decodeErr != nil {
		return "", Usage{}, fmt.Errorf("decode provider response: %w", decodeErr)
	}
	if parsed.Error != nil {
		return "", Usage{}, fmt.Errorf("provider error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return "", Usage{}, errors.New("provider returned no choices")
	}
	if parsed.Choices[0].FinishReason == "length" {
		return "", Usage{}, errors.New("provider response was truncated")
	}

	usage := Usage{
		Provider:         p.name,
		Model:            model,
		PromptVersion:    promptVersion,
		PromptTokens:     parsed.Usage.PromptTokens,
		CompletionTokens: parsed.Usage.CompletionTokens,
		LatencyMS:        int(time.Since(started).Milliseconds()),
	}
	return parsed.Choices[0].Message.Content, usage, nil
}
