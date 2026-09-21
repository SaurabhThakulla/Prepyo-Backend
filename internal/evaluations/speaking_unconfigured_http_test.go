package evaluations

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/config"
)

// Drives the real handler over HTTP with the deployed configuration: no
// AI_AUDIO_*, so no model can hear a recording. The answer must arrive without
// a provider call, without touching the database, and pointing at the route
// that can actually grade the answer.
func TestSpeakingSubmitWithoutAudioProvider(t *testing.T) {
	cfg := &config.Config{
		AIBaseURL:         "https://example.invalid/v1",
		AIAPIKey:          "text-key",
		AIAudioBaseURL:    "https://example.invalid/v1",
		AIAudioAPIKey:     "text-key",
		SpeakingEvaluated: false,
		AIModels:          config.AIModels{Writing: "text-model", Speaking: "audio-model"},
		AIRequestTimeout:  30 * time.Second,
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// A nil pool is the point: a refusal that reaches the database has already
	// spent longer than it should, and this one must not get that far.
	service := NewService(nil, nil, nil, nil, nil, ai.NewGateway(cfg, log), nil)
	handler := NewHandler(service, nil, log)

	body := `{"questionId":"ielts-audit-011","audio":"` +
		base64.StdEncoding.EncodeToString(make([]byte, 64_000)) +
		`","format":"mp3","durationSeconds":23}`

	request := httptest.NewRequest(http.MethodPost, "/speaking", strings.NewReader(body))
	request = request.WithContext(reqctx.WithUser(request.Context(), models.User{ID: "learner"}))
	response := httptest.NewRecorder()

	started := time.Now()
	handler.Routes().ServeHTTP(response, request)
	elapsed := time.Since(started)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body=%s", response.Code, response.Body.String())
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body %q: %v", response.Body.String(), err)
	}
	if payload.Error.Code != "not_configured" {
		t.Fatalf("code = %q, want not_configured", payload.Error.Code)
	}
	if !strings.Contains(payload.Error.Message, "transcript") {
		t.Fatalf("message = %q, want it to point at the transcript route", payload.Error.Message)
	}

	// The old path uploaded the recording three times before giving up.
	if elapsed > time.Second {
		t.Fatalf("took %s to refuse a submission it could never score", elapsed)
	}
}
