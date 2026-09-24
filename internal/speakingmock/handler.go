package speakingmock

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/internal/speech"
	"github.com/prepyo/backend/pkg/httpx"
)

// examinerVoice is the one voice the examiner speaks in, whichever the browser.
var examinerVoice = map[string]string{"": "daniel"}

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
	r.Get("/{sessionID}/steps/{key}/audio", h.stepAudio)
	r.Post("/{sessionID}/answers", h.saveAnswer)
	r.Post("/{sessionID}/submit", h.submit)
	return r
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Start(r.Context(), reqctx.MustUser(r.Context()), true)
	if err != nil {
		h.writeError(w, "speakingmock.start", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{"session": session})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	session, err := h.svc.Resume(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "speakingmock.get", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": session})
}

// stepAudio is the examiner's words for one step, spoken on the server, for
// browsers that cannot speak them. ?say=lead or ?say=question asks for one
// half, as the Part 2 long turn needs: its introduction comes before the
// minute of preparation and its question after.
func (h *Handler) stepAudio(w http.ResponseWriter, r *http.Request) {
	step, err := h.svc.Step(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), chi.URLParam(r, "key"))
	if err != nil {
		h.writeError(w, "speakingmock.stepAudio", err)
		return
	}
	words := step.Lead + " " + step.Question
	switch r.URL.Query().Get("say") {
	case "lead":
		words = step.Lead
	case "question":
		words = step.Question
	}
	speech.WriteSegments(w, r, h.speech, h.log, strings.TrimSpace(words), examinerVoice)
}

type answerRequest struct {
	Key               string `json:"key"`
	Audio             string `json:"audio"`
	Format            string `json:"format"`
	DurationSeconds   int    `json:"durationSeconds"`
	BrowserTranscript string `json:"browserTranscript"`
}

func (h *Handler) saveAnswer(w http.ResponseWriter, r *http.Request) {
	var req answerRequest
	if !httpx.Decode(w, r, &req, h.log, "speakingmock.saveAnswer") {
		return
	}
	answer, err := h.svc.SaveAnswer(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"), AnswerParams{
		Key: req.Key, AudioBase64: req.Audio, Format: req.Format,
		DurationSeconds: req.DurationSeconds, BrowserTranscript: req.BrowserTranscript,
	})
	if err != nil {
		h.writeError(w, "speakingmock.saveAnswer", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"answer": answer})
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.Submit(r.Context(), reqctx.MustUser(r.Context()), chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeError(w, "speakingmock.submit", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"attempt":    result.Attempt,
		"evaluation": result.Evaluation,
		"xpAwarded":  result.XPAwarded,
	})
}

func (h *Handler) writeError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, ErrNotIELTS):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "The speaking mock is an IELTS test. Switch your exam to IELTS to take it.")
	case errors.Is(err, ErrNoSet):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "There is no speaking test available yet. Please try again soon.")
	case errors.Is(err, ErrSessionNotFound), errors.Is(err, ErrUnknownStep):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That speaking mock does not exist.")
	case errors.Is(err, ErrAlreadySubmitted):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "This speaking mock has already been submitted.")
	case errors.Is(err, ErrPaperClosed):
		httpx.Error(w, http.StatusConflict, httpx.CodePaperExpired, "This speaking mock is closed. Its time has run out or it has been submitted.")
	case errors.Is(err, ErrExpiredTooLittle):
		httpx.Error(w, http.StatusConflict, httpx.CodePaperExpired, "Time ran out before there was enough speech to rate, so this test is closed.")
	case errors.Is(err, ErrTooLittle):
		httpx.Error(w, http.StatusUnprocessableEntity, httpx.CodeNotAttempted, "Too little speech was heard to rate the test: at least three answers need words we could make out. Check that your microphone is working.")
	case errors.Is(err, billing.ErrGradingLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached, billing.GradingLimitMessage)
	case errors.Is(err, billing.ErrLimitReached):
		httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
			fmt.Sprintf("A section mock test uses %d of your daily practice sub-tests, and you don't have enough left today. They reset tomorrow, or upgrade your plan for more.", billing.SectionMockSubTests))
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable, "Rating took longer than we could wait. Your answers are saved - please submit again.")
	case errors.Is(err, ai.ErrUnavailable), errors.Is(err, ai.ErrBadOutput):
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable, "Rating is unavailable right now. Your answers are saved - please submit again shortly.")
	default:
		httpx.Internal(w, h.log, op, err)
	}
}
