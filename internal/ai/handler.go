package ai

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	gateway *Gateway
	log     *slog.Logger
}

func NewHandler(gateway *Gateway, log *slog.Logger) *Handler {
	return &Handler{gateway: gateway, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/tutor", h.tutor)
	return r
}

func (h *Handler) tutor(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Messages    []TutorMessage `json:"messages"`
		TaskContext string         `json:"taskContext"`
	}
	if !httpx.Decode(w, r, &req, h.log, "ai.tutor") {
		return
	}
	if len(req.Messages) == 0 {
		httpx.ValidationError(w, map[string]string{"messages": "Send at least one message."})
		return
	}

	user := reqctx.MustUser(r.Context())
	exam := user.TargetExam
	if !exam.Valid() {
		exam = models.ExamPTE
	}

	reply, usage, err := h.gateway.Tutor(r.Context(), TutorRequest{
		Exam:        exam,
		Messages:    req.Messages,
		TaskContext: req.TaskContext,
	})
	if err != nil {
		if errors.Is(err, ErrUnavailable) || errors.Is(err, ErrBadOutput) {
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
				"The tutor is unavailable right now. Please try again shortly.")
			return
		}
		httpx.Internal(w, h.log, "ai.tutor", err)
		return
	}

	h.log.Info("tutor reply",
		"userId", user.ID,
		"model", usage.Model,
		"promptTokens", usage.PromptTokens,
		"completionTokens", usage.CompletionTokens,
		"latencyMs", usage.LatencyMS)

	httpx.JSON(w, http.StatusOK, map[string]any{"reply": reply})
}
