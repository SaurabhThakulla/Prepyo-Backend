package listeningmock

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/internal/speech"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	svc    *Service
	speech *speech.Service
	log    *slog.Logger
}

func NewHandler(svc *Service, speechService *speech.Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, speech: speechService, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.start)
	r.Get("/{sessionID}", h.get)
	r.Get("/{sessionID}/parts/{partNo}/script", h.script)
	r.Get("/{sessionID}/parts/{partNo}/audio", h.audio)
	r.Put("/{sessionID}/answers", h.saveDrafts)
	r.Post("/{sessionID}/submit", h.submit)
	return r
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Start(r.Context(), reqctx.MustUser(r.Context()), StartOptions{Charge: true})
	if err != nil {
		h.writeError(w, "listeningmock.start", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"session": session})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Resume(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "listeningmock.get", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": session})
}

func (h *Handler) partScript(w http.ResponseWriter, r *http.Request) (string, bool) {
	partNo, err := strconv.Atoi(chi.URLParam(r, "partNo"))
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That part is not on this paper.")
		return "", false
	}
	script, err := h.svc.Script(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), partNo)
	if err != nil {
		h.writeError(w, "listeningmock.script", err)
		return "", false
	}
	return script, true
}

func (h *Handler) script(w http.ResponseWriter, r *http.Request) {
	script, ok := h.partScript(w, r)
	if !ok {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpx.JSON(w, http.StatusOK, map[string]any{"script": script})
}

func (h *Handler) audio(w http.ResponseWriter, r *http.Request) {
	script, ok := h.partScript(w, r)
	if !ok {
		return
	}
	speech.WriteSegments(w, r, h.speech, h.log, script, nil)
}

type answersRequest struct {
	Answers []models.AnswerSubmission `json:"answers"`
	// PartsPlayed is how many recordings have started playing.
	PartsPlayed int `json:"partsPlayed"`
}

func (h *Handler) saveDrafts(w http.ResponseWriter, r *http.Request) {
	var req answersRequest
	if !httpx.Decode(w, r, &req, h.log, "listeningmock.saveDrafts") {
		return
	}
	remaining, err := h.svc.SaveDrafts(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), req.Answers, req.PartsPlayed)
	if err != nil {
		h.writeError(w, "listeningmock.saveDrafts", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"secondsRemaining": remaining})
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	var req answersRequest
	if !httpx.Decode(w, r, &req, h.log, "listeningmock.submit") {
		return
	}
	result, err := h.svc.Submit(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), req.Answers)
	if err != nil {
		h.writeError(w, "listeningmock.submit", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"attempt":   result.Attempt,
		"xpAwarded": result.XPAwarded,
		"review":    result.Review,
		"scripts":   result.Scripts,
	})
}

func (h *Handler) writeError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, ErrNotIELTS):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "The listening mock is an IELTS paper. Switch your exam to IELTS to take it.")
	case errors.Is(err, ErrNoTest):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "There is no listening test available yet. Please try again soon.")
	case errors.Is(err, ErrSessionNotFound), errors.Is(err, ErrNoPart):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That listening mock does not exist.")
	case errors.Is(err, ErrAlreadySubmitted):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "This listening mock has already been submitted.")
	case errors.Is(err, ErrPaperClosed):
		httpx.Error(w, http.StatusConflict, httpx.CodePaperExpired, "This listening mock is closed. Its time has run out or it has been submitted.")
	case errors.Is(err, ErrNotAttempted):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeNotAttempted, "Answer at least one question before you submit. A blank paper does not get a band.")
	case errors.Is(err, billing.ErrLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
			fmt.Sprintf("A section mock test uses %d of your daily practice sub-tests, and you don't have enough left today. They reset tomorrow, or upgrade your plan for more.", billing.SectionMockSubTests))
	default:
		httpx.Internal(w, h.log, op, err)
	}
}
