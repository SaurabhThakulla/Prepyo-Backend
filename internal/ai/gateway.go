// Package ai handles communication with AI model providers for evaluation and tutoring.
package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
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

// minCallBudget is the least time worth starting a provider call with. A call
// that cannot finish before the request deadline is worse than no call: it
// consumes the remaining budget and fails anyway.
const minCallBudget = 5 * time.Second

// timeLeftForCall reports whether the deadline leaves room for another attempt.
// A context without a deadline always does.
func timeLeftForCall(ctx context.Context) bool {
	deadline, ok := ctx.Deadline()
	if !ok {
		return ctx.Err() == nil
	}
	return ctx.Err() == nil && time.Until(deadline) > minCallBudget
}

type provider struct {
	name    string
	baseURL string
	apiKey  string
}

func (p provider) configured() bool { return p.apiKey != "" && p.baseURL != "" }

func (p provider) completionsURL() string { return p.baseURL + "/chat/completions" }

func (p provider) transcriptionsURL() string { return p.baseURL + "/audio/transcriptions" }

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
	speaking  bool
	ttsModel  string
	ttsVoices []string
	models    config.AIModels
	maxTokens int
	log       *slog.Logger
}

func NewGateway(cfg *config.Config, log *slog.Logger) *Gateway {
	return &Gateway{
		client:    &http.Client{Timeout: cfg.AIRequestTimeout},
		text:      newProvider(cfg.AIBaseURL, cfg.AIAPIKey),
		audio:     newProvider(cfg.AIAudioBaseURL, cfg.AIAudioAPIKey),
		speaking:  cfg.SpeakingEvaluated,
		ttsModel:  cfg.AITTSModel,
		ttsVoices: cfg.AITTSVoices,
		models:    cfg.AIModels,
		maxTokens: cfg.AIMaxTokens,
		log:       log,
	}
}

// Available reports whether text evaluation and tutoring can run.
func (g *Gateway) Available() bool { return g.text.configured() }

// SpeakingAvailable reports whether a recording can be scored: an audio
// provider that was configured as one, not the text provider standing in for it
// because its key was inherited.
func (g *Gateway) SpeakingAvailable() bool { return g.speaking && g.audio.configured() }

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
		if strings.Contains(primary, "/") {
			pool = []string{primary, "google/gemini-2.5-flash", "google/gemini-2.5-flash-lite"}
		} else {
			pool = []string{primary, "gemini-3.7-flash", "gemini-3.6-flash"}
		}
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
			if strings.Contains(primary, "/") {
				result = []string{"google/gemini-2.5-flash"}
			} else {
				result = []string{"gemini-3.7-flash"}
			}
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
		// Starting a call the deadline cannot fit spends what little time is
		// left and leaves the caller holding a cancellation instead of a clean
		// "unavailable", which is the difference between telling a learner to
		// try again and telling them something broke.
		if !timeLeftForCall(ctx) {
			g.log.Warn("skipping ai candidate, not enough time left",
				"model", candidateModel, "primaryModel", model)
			if lastErr == nil {
				lastErr = context.DeadlineExceeded
			}
			break
		}

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

		unresponsive := false
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
			// A host that never sends headers is down for every model on it,
			// not just this one. Walking the candidate list costs the learner
			// one client timeout per entry and ends in the same failure, so
			// stop at the first one and say so while they are still waiting.
			if requestTimedOut(err) {
				unresponsive = true
				break
			}
			if errors.Is(err, errPermanent) {
				break
			}
			if attempt < maxAttempts {
				if !timeLeftForCall(ctx) {
					break
				}
				g.log.Warn("ai request failed, retrying", "model", candidateModel, "attempt", attempt, "error", err)
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			}
		}

		if unresponsive {
			g.log.Error("ai provider is not responding, skipping remaining candidates",
				"provider", p.name, "failedModel", candidateModel, "error", lastErr)
			break
		}

		g.log.Warn("ai model failed, attempting fallback candidate", "failedModel", candidateModel, "error", lastErr)
	}

	g.log.Error("all ai candidate models failed", "primaryModel", model, "error", lastErr)
	return "", Usage{}, fmt.Errorf("%w: %w", ErrUnavailable, lastErr)
}

// requestTimedOut reports whether the provider never answered, as opposed to
// answering with a refusal. The client's own timeout surfaces as a deadline on
// the request, which is why this is not the same check as ctx.Err().
func requestTimedOut(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err)
}

const defaultAIUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"

func (g *Gateway) send(ctx context.Context, p provider, body []byte, model, promptVersion string, started time.Time) (string, Usage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.completionsURL(), bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("User-Agent", defaultAIUserAgent)
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

// transcribeAudio uploads the raw recording to a dedicated audio transcription endpoint (such as Groq Whisper).
func (g *Gateway) transcribeAudio(ctx context.Context, req SpeakingRequest) (string, Usage, error) {
	started := time.Now()
	audioBytes, err := base64.StdEncoding.DecodeString(req.AudioBase64)
	if err != nil {
		return "", Usage{}, fmt.Errorf("decode audio base64: %w", err)
	}

	model := g.models.Speaking
	if !strings.Contains(strings.ToLower(model), "whisper") {
		model = "whisper-large-v3-turbo"
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	filename := "audio." + req.AudioFormat
	if req.AudioFormat == "" {
		filename = "audio.wav"
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", Usage{}, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(audioBytes); err != nil {
		return "", Usage{}, fmt.Errorf("write audio to form: %w", err)
	}

	if err := writer.WriteField("model", model); err != nil {
		return "", Usage{}, fmt.Errorf("write model field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return "", Usage{}, fmt.Errorf("close multipart writer: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.audio.transcriptionsURL(), &body)
	if err != nil {
		return "", Usage{}, fmt.Errorf("create transcription request: %w", err)
	}
	httpReq.Header.Set("User-Agent", defaultAIUserAgent)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+g.audio.apiKey)

	res, err := g.client.Do(httpReq)
	if err != nil {
		return "", Usage{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", Usage{}, err
	}

	if res.StatusCode != http.StatusOK {
		g.log.Error("transcription provider error", "status", res.StatusCode, "body", string(raw))
		return "", Usage{}, fmt.Errorf("transcription provider returned %d: %s", res.StatusCode, string(raw))
	}

	var resp struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", Usage{}, fmt.Errorf("decode transcription response: %w", err)
	}

	usage := Usage{
		Provider:      g.audio.name,
		Model:         model,
		PromptVersion: "audio.transcription",
		LatencyMS:     int(time.Since(started).Milliseconds()),
	}
	return strings.TrimSpace(resp.Text), usage, nil
}
