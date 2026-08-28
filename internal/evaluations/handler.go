package evaluations

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	service *Service
	repo    *Repository
	log     *slog.Logger
}

func NewHandler(service *Service, repo *Repository, log *slog.Logger) *Handler {
	return &Handler{service: service, repo: repo, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/writing", h.evaluateWriting)
	r.Post("/speaking", h.evaluateSpeaking)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	page := httpx.ReadPage(r)

	list, total, err := h.repo.List(r.Context(), ListParams{UserID: user.ID, Limit: page.Limit, Offset: page.Offset})
	if err != nil {
		httpx.Internal(w, h.log, "evaluations.list", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"evaluations": list,
		"pagination":  page.Meta(total),
	})
}

func (h *Handler) evaluateWriting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuestionID string `json:"questionId"`
		Text       string `json:"text"`
	}
	if !httpx.Decode(w, r, &req, h.log, "evaluations.evaluateWriting") {
		return
	}

	problems := map[string]string{}
	if strings.TrimSpace(req.QuestionID) == "" {
		problems["questionId"] = "Required."
	}
	if strings.TrimSpace(req.Text) == "" {
		problems["text"] = "Write your response before submitting."
	}
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	user := reqctx.MustUser(r.Context())
	outcome, err := h.service.EvaluateWriting(r.Context(), Request{
		User:       user,
		QuestionID: req.QuestionID,
		Text:       req.Text,
	})
	if err != nil {
		h.writeError(w, err, "evaluations.evaluateWriting", map[string]string{
			"text": "Write at least 20 words so there is something to give feedback on.",
		})
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"evaluation":   outcome.Evaluation,
		"reused":       outcome.Reused,
		"xpAwarded":    outcome.XPAwarded,
		"streak":       outcome.Streak,
		"missions":     outcome.Missions,
		"subscription": outcome.Subscription,
	})
}

// maxRecordingSeconds is longer than any single task in either exam — the
// longest is an IELTS Part 2 turn at two minutes — with headroom for a learner
// who starts a moment early.
const maxRecordingSeconds = 180

func (h *Handler) evaluateSpeaking(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuestionID string `json:"questionId"`
		// Audio is base64, without a data: URL prefix.
		Audio           string `json:"audio"`
		Format          string `json:"format"`
		DurationSeconds int    `json:"durationSeconds"`
	}
	if !httpx.DecodeLimit(w, r, &req, h.log, "evaluations.evaluateSpeaking", httpx.MaxAudioBodyBytes) {
		return
	}

	problems := map[string]string{}
	if strings.TrimSpace(req.QuestionID) == "" {
		problems["questionId"] = "Required."
	}
	if strings.TrimSpace(req.Audio) == "" {
		problems["audio"] = "Record your answer before submitting."
	}
	if !ai.AudioFormats[req.Format] {
		problems["format"] = "That recording format is not supported."
	}
	if req.DurationSeconds <= 0 || req.DurationSeconds > maxRecordingSeconds {
		problems["durationSeconds"] = "That recording is longer than any speaking task allows."
	}
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	// Decoding here rather than in the service means a corrupt upload is a
	// rejected request, not a provider call the learner pays an allowance for.
	audio, err := base64.StdEncoding.DecodeString(req.Audio)
	if err != nil {
		h.log.Warn("rejected recording", "op", "evaluations.evaluateSpeaking", "error", err)
		httpx.ValidationError(w, map[string]string{
			"audio": "That recording could not be read. Please record it again.",
		})
		return
	}

	user := reqctx.MustUser(r.Context())
	outcome, err := h.service.EvaluateSpeaking(r.Context(), SpeakingRequest{
		User:            user,
		QuestionID:      req.QuestionID,
		Audio:           audio,
		AudioFormat:     req.Format,
		DurationSeconds: req.DurationSeconds,
	})
	if err != nil {
		h.writeError(w, err, "evaluations.evaluateSpeaking", map[string]string{
			"audio": "That recording was too short to give feedback on. Speak for a few seconds and try again.",
		})
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"evaluation":   outcome.Evaluation,
		"reused":       outcome.Reused,
		"xpAwarded":    outcome.XPAwarded,
		"streak":       outcome.Streak,
		"missions":     outcome.Missions,
		"subscription": outcome.Subscription,
	})
}

// writeError maps a service error onto the response. `tooShort` is the wording
// for a submission with nothing in it, which differs by skill: one learner
// needs more words, the other needs to actually speak.
func (h *Handler) writeError(w http.ResponseWriter, err error, op string, tooShort map[string]string) {
	switch {
	case errors.Is(err, questions.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That question does not exist.")

	case errors.Is(err, ErrWrongSkill):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"That question is not a speaking task.")

	case errors.Is(err, ErrEmptyResponse):
		httpx.ValidationError(w, tooShort)

	case errors.Is(err, ErrLimitReached):
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeLimitReached,
			"You have used all of today's AI evaluations. They reset at midnight.")

	case errors.Is(err, ai.ErrUnavailable), errors.Is(err, ai.ErrBadOutput):
		// Nothing was stored and no score was invented. The learner's work is
		// still in the browser, so they can retry.
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
			"Evaluation is unavailable right now. Your response was not lost - please try again shortly.")

	default:
		httpx.Internal(w, h.log, op, err)
	}
}
