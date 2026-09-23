package fullmock

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/listeningmock"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reading"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/internal/speakingmock"
	"github.com/prepyo/backend/internal/writingmock"
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
	r.Get("/current", h.current)
	r.Get("/{sessionID}", h.get)
	r.Post("/{sessionID}/sections/{skill}", h.startSection)
	r.Post("/{sessionID}/finish", h.finish)
	r.Post("/{sessionID}/abandon", h.abandon)
	return r
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Start(r.Context(), reqctx.MustUser(r.Context()))
	if err != nil {
		h.writeError(w, "fullmock.start", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"session": session})
}

// current is the learner's open full mock; 404 when there is none.
func (h *Handler) current(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Current(r.Context(), reqctx.MustUser(r.Context()))
	if err != nil {
		h.writeError(w, "fullmock.current", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": session})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Get(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "fullmock.get", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": session})
}

// startSection opens the next section. The paper comes back as the section
// mock's own start returns it, for that section's runner.
func (h *Handler) startSection(w http.ResponseWriter, r *http.Request) {
	session, paper, err := h.svc.StartSection(r.Context(), reqctx.MustUser(r.Context()),
		chi.URLParam(r, "sessionID"), models.SkillType(chi.URLParam(r, "skill")))
	if err != nil {
		h.writeError(w, "fullmock.startSection", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"fullMock": session, "session": paper})
}

func (h *Handler) finish(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Finish(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "fullmock.finish", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": session})
}

func (h *Handler) abandon(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Abandon(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID")); err != nil {
		h.writeError(w, "fullmock.abandon", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writeError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, ErrNotIELTS):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "The full mock is an IELTS test. Switch your exam to IELTS to take it.")
	case errors.Is(err, ErrSessionNotFound), errors.Is(err, ErrUnknownSection):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That full mock does not exist.")
	case errors.Is(err, ErrNotInProgress):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "This full mock is already finished.")
	case errors.Is(err, ErrOutOfOrder):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "The sections of a full mock are taken in order: Listening, Reading, Writing, then Speaking.")
	case errors.Is(err, ErrSectionsUnfinished):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "Finish all four sections first.")
	case errors.Is(err, billing.ErrMockLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
			"You have used all the full mock tests in your plan. Upgrade or refer friends to unlock more.")
	case errors.Is(err, listeningmock.ErrNoTest), errors.Is(err, writingmock.ErrNoTasks),
		errors.Is(err, reading.ErrBankTooSmall), errors.Is(err, reading.ErrNoBlueprint),
		errors.Is(err, speakingmock.ErrNoSet):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "This section's paper is not available yet. Please try again soon.")
	default:
		httpx.Internal(w, h.log, op, err)
	}
}
