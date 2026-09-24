package writingmock

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.start)
	r.Get("/{sessionID}", h.get)
	r.Put("/{sessionID}/drafts", h.saveDrafts)
	r.Post("/{sessionID}/submit", h.submit)
	return r
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Start(r.Context(), reqctx.MustUser(r.Context()))
	if err != nil {
		h.writeError(w, "writingmock.start", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"session": session})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Resume(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "writingmock.get", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": session})
}

type draftsRequest struct {
	Task1 string `json:"task1"`
	Task2 string `json:"task2"`
}

func (h *Handler) saveDrafts(w http.ResponseWriter, r *http.Request) {
	var req draftsRequest
	if !httpx.Decode(w, r, &req, h.log, "writingmock.saveDrafts") {
		return
	}
	session, err := h.svc.SaveDrafts(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), req.Task1, req.Task2)
	if err != nil {
		h.writeError(w, "writingmock.saveDrafts", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"secondsRemaining": session.SecondsRemaining})
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var req draftsRequest
	if !httpx.Decode(w, r, &req, h.log, "writingmock.submit") {
		return
	}
	result, err := h.svc.Submit(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), req.Task1, req.Task2)
	if err != nil {
		h.writeError(w, "writingmock.submit", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"attempt":         result.Attempt,
		"writingBand":     result.WritingBand,
		"task1Evaluation": result.Task1Evaluation,
		"task2Evaluation": result.Task2Evaluation,
		"xpAwarded":       result.XPAwarded,
		"streak":          result.Streak,
	})
}

func (h *Handler) writeError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, ErrNotIELTS):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "The writing mock is an IELTS paper. Switch your exam to IELTS to take it.")
	case errors.Is(err, ErrNoTasks):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "There are not enough writing tasks to build a mock yet. Please try again soon.")
	case errors.Is(err, ErrSessionNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That writing mock does not exist.")
	case errors.Is(err, ErrAlreadySubmitted):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "This writing mock has already been submitted.")
	case errors.Is(err, ErrPaperClosed):
		httpx.Error(w, http.StatusConflict, httpx.CodePaperExpired, "This writing mock is closed. Its time has run out or it has been submitted.")
	case errors.Is(err, billing.ErrGradingLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached, billing.GradingLimitMessage)
	case errors.Is(err, billing.ErrLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
			fmt.Sprintf("A section mock test uses %d of your daily practice sub-tests, and you don't have enough left today. They reset tomorrow, or upgrade your plan for more.", billing.SectionMockSubTests))
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
			"Rating took longer than we could wait. Your writing is saved - please submit again.")
	case errors.Is(err, ai.ErrUnavailable), errors.Is(err, ai.ErrBadOutput):
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
			"Evaluation is unavailable right now. Your writing is saved - please submit again shortly.")
	default:
		httpx.Internal(w, h.log, op, err)
	}
}
