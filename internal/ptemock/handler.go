package ptemock

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/mockpapers"
	"github.com/prepyo/backend/internal/models"
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

// Routes are mounted at /pte/mocks.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/catalog", h.catalog)
	r.Get("/sessions", h.history)
	r.Post("/", h.start)
	r.Get("/{sessionID}", h.get)
	r.Put("/{sessionID}/items/{position}/draft", h.saveDraft)
	r.Post("/{sessionID}/items/{position}", h.submit)
	r.Post("/{sessionID}/finish", h.finish)
	r.Post("/{sessionID}/abandon", h.abandon)
	r.Get("/{sessionID}/report", h.report)
	return r
}

func (h *Handler) catalog(w http.ResponseWriter, r *http.Request) {
	entries, err := h.svc.Catalog(r.Context(), reqctx.MustUser(r.Context()))
	if err != nil {
		h.writeError(w, "ptemock.catalog", err)
		return
	}
	tasks := make([]Task, 0, len(Tasks))
	for _, bp := range Blueprints[:1] {
		for _, slot := range bp.Slots {
			tasks = append(tasks, Tasks[slot.Task])
		}
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"mocks": entries,
		"tasks": tasks,
		// Whether spoken answers are transcribed here (Whisper), so the
		// pre-test check knows whether the browser's recognition matters.
		"serverTranscription": h.svc.ServerTranscription(),
	})
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	page := httpx.ReadPage(r)
	list, total, err := h.svc.History(r.Context(), reqctx.MustUser(r.Context()), page.Limit, page.Offset)
	if err != nil {
		h.writeError(w, "ptemock.history", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"sessions": list, "pagination": page.Meta(total)})
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Kind    Kind   `json:"kind"`
		PaperID string `json:"paperId"`
	}
	if !httpx.Decode(w, r, &req, h.log, "ptemock.start") {
		return
	}
	view, err := h.svc.Start(r.Context(), reqctx.MustUser(r.Context()), req.Kind, req.PaperID)
	if err != nil {
		h.writeError(w, "ptemock.start", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"session": view})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.Get(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "ptemock.get", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": view})
}

func (h *Handler) saveDraft(w http.ResponseWriter, r *http.Request) {
	position, ok := h.position(w, r)
	if !ok {
		return
	}
	var req struct {
		Response *models.AnswerSubmission `json:"response"`
	}
	if !httpx.Decode(w, r, &req, h.log, "ptemock.saveDraft") {
		return
	}
	view, err := h.svc.SaveDraft(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), position, req.Response)
	if err != nil {
		h.writeError(w, "ptemock.saveDraft", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"clock": view.Clock, "current": view.Current, "status": view.Status})
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	position, ok := h.position(w, r)
	if !ok {
		return
	}
	var req Answer
	// A spoken answer may carry its recording, for server transcription.
	if !httpx.DecodeLimit(w, r, &req, h.log, "ptemock.submit", httpx.MaxAudioBodyBytes) {
		return
	}
	view, err := h.svc.Submit(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), position, req)
	if err != nil {
		h.writeError(w, "ptemock.submit", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": view})
}

func (h *Handler) finish(w http.ResponseWriter, r *http.Request) {
	view, err := h.svc.Finish(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "ptemock.finish", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": view})
}

func (h *Handler) abandon(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Abandon(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID")); err != nil {
		h.writeError(w, "ptemock.abandon", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.Report(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "ptemock.report", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"report": report})
}

func (h *Handler) position(w http.ResponseWriter, r *http.Request) (int, bool) {
	position, err := strconv.Atoi(chi.URLParam(r, "position"))
	if err != nil || position < 1 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That item does not exist.")
		return 0, false
	}
	return position, true
}

func (h *Handler) writeError(w http.ResponseWriter, op string, err error) {
	if mockpapers.WriteError(w, err) {
		return
	}
	switch {
	case errors.Is(err, ErrNotPTE):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "PTE mock tests are PTE papers. Switch your exam to PTE to take one.")
	case errors.Is(err, ErrUnknownKind):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Choose the full test or a sectional test: speaking, writing, reading or listening.")
	case errors.Is(err, ErrBankTooSmall):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "There are not enough questions to build this test yet. Please try again soon.")
	case errors.Is(err, ErrSessionNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That mock test does not exist.")
	case errors.Is(err, ErrNotInProgress):
		httpx.Error(w, http.StatusConflict, httpx.CodePaperExpired, "This test is no longer in progress.")
	case errors.Is(err, ErrOutOfOrder):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "That item is no longer on screen. The test only moves forward.")
	case errors.Is(err, ErrNoDraft):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "A spoken answer is recorded, not drafted.")
	case errors.Is(err, ErrNotFinished):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "This test has no score report: it is still in progress or was left unfinished.")
	case errors.Is(err, billing.ErrMockLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
			"You have used all the full mock tests in your plan. Upgrade or refer friends to unlock more.")
	case errors.Is(err, billing.ErrGradingLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached, billing.GradingLimitMessage)
	case errors.Is(err, billing.ErrLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
			fmt.Sprintf("A sectional test uses %d of your daily practice sub-tests, and you don't have enough left today. They reset tomorrow, or upgrade your plan for more.", billing.SectionMockSubTests))
	default:
		httpx.Internal(w, h.log, op, err)
	}
}
