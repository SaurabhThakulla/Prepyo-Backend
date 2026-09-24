package evaluations

import (
	"context"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/internal/scoring"
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
	r.Post("/speaking/transcript", h.evaluateSpeakingTranscript)
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
			"text": "Write at least 5 words for Summarize Written Text, or at least 20 words for other writing tasks, so there is something to give feedback on.",
		})
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"evaluation":     outcome.Evaluation,
		"reused":         outcome.Reused,
		"xpAwarded":      outcome.XPAwarded,
		"streak":         outcome.Streak,
		"missions":       outcome.Missions,
		"subscription":   outcome.Subscription,
		"notificationId": outcome.NotificationID,
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

	// Checked before the recording is decoded, let alone uploaded: with no
	// audio provider this request can only end one way, and spending a minute
	// discovering that is worse for the learner than being told now.
	if !h.service.SpeakingAvailable() {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeNotConfigured,
			"Recordings are not being scored here. Send the transcript of your answer instead.")
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
		"evaluation":     outcome.Evaluation,
		"reused":         outcome.Reused,
		"xpAwarded":      outcome.XPAwarded,
		"streak":         outcome.Streak,
		"missions":       outcome.Missions,
		"subscription":   outcome.Subscription,
		"notificationId": outcome.NotificationID,
	})
}

// maxTranscriptChars is generous for two minutes of speech and small enough
// that a pasted essay is not smuggled in as a spoken answer.
const maxTranscriptChars = 6000

// evaluateSpeakingTranscript scores a spoken answer from the transcript the
// learner's device produced, which is the only honest way to grade speaking on
// a provider that serves text models only. It judges words, never delivery,
// and the feedback it stores says so.
func (h *Handler) evaluateSpeakingTranscript(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QuestionID      string `json:"questionId"`
		Transcript      string `json:"transcript"`
		DurationSeconds int    `json:"durationSeconds"`
		// Delivery is what the browser measured from the waveform. Absent on a
		// browser that could not measure it, and then pace goes unjudged
		// rather than guessed.
		Delivery *struct {
			DurationSeconds     int     `json:"durationSeconds"`
			PauseCount          int     `json:"pauseCount"`
			LongestPauseSeconds float64 `json:"longestPauseSeconds"`
			SpeakingRatio       float64 `json:"speakingRatio"`
		} `json:"delivery"`
	}
	if !httpx.Decode(w, r, &req, h.log, "evaluations.evaluateSpeakingTranscript") {
		return
	}

	// When recordings can be scored properly, they should be: this path is the
	// fallback, not a cheaper alternative to it.
	if !h.service.SpeakingTranscriptAvailable() {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"Submit the recording itself for this task.")
		return
	}

	transcript := strings.TrimSpace(req.Transcript)
	problems := map[string]string{}
	if strings.TrimSpace(req.QuestionID) == "" {
		problems["questionId"] = "Required."
	}
	// An empty transcript is still an answer: it is scored at the bottom of the
	// scale rather than refused (see tooFewWordsEvaluation).
	if len(transcript) > maxTranscriptChars {
		problems["transcript"] = "That is longer than any speaking task allows."
	}
	if req.DurationSeconds <= 0 || req.DurationSeconds > maxRecordingSeconds {
		problems["durationSeconds"] = "That recording is longer than any speaking task allows."
	}
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	// Measurements arrive from the learner's own browser, so they are checked
	// like any other client input: a "pause" longer than the recording, or a
	// speaking ratio above 1, means the numbers are not describing a recording
	// and are dropped rather than scored.
	var delivery scoring.Delivery
	if m := req.Delivery; m != nil {
		sane := m.DurationSeconds > 0 && m.DurationSeconds <= maxRecordingSeconds &&
			m.PauseCount >= 0 && m.PauseCount <= m.DurationSeconds &&
			m.LongestPauseSeconds >= 0 && m.LongestPauseSeconds <= float64(m.DurationSeconds) &&
			m.SpeakingRatio >= 0 && m.SpeakingRatio <= 1
		if sane {
			delivery = scoring.Delivery{
				DurationSeconds:     m.DurationSeconds,
				WordsSpoken:         len(strings.Fields(transcript)),
				PauseCount:          m.PauseCount,
				LongestPauseSeconds: m.LongestPauseSeconds,
				SpeakingRatio:       m.SpeakingRatio,
			}
		} else {
			h.log.Warn("ignored implausible delivery measurements",
				"op", "evaluations.evaluateSpeakingTranscript", "delivery", m)
		}
	}

	user := reqctx.MustUser(r.Context())
	outcome, err := h.service.EvaluateSpeakingTranscript(r.Context(), TranscriptRequest{
		User:            user,
		QuestionID:      req.QuestionID,
		Transcript:      transcript,
		DurationSeconds: req.DurationSeconds,
		Delivery:        delivery,
	})
	if err != nil {
		h.writeError(w, err, "evaluations.evaluateSpeakingTranscript", map[string]string{
			"transcript": "We caught too few words to give feedback on. Speak for a few seconds and try again.",
		})
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"evaluation":            outcome.Evaluation,
		"reused":                outcome.Reused,
		"xpAwarded":             outcome.XPAwarded,
		"streak":                outcome.Streak,
		"missions":              outcome.Missions,
		"subscription":          outcome.Subscription,
		"notificationId":        outcome.NotificationID,
		"pronunciationAssessed": false,
	})
}

// writeError maps a service error onto the response. `tooShort` is the wording
// for a submission with nothing in it, which differs by skill: one learner
// needs more words, the other needs to actually speak.
func (h *Handler) writeError(w http.ResponseWriter, err error, op string, tooShort map[string]string) {
	switch {
	case errors.Is(err, questions.ErrNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That question does not exist.")

	case errors.Is(err, ErrWrongWritingSkill):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"That question is not a writing task.")

	case errors.Is(err, ErrWrongSkill):
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"That question is not a speaking task.")

	case errors.Is(err, billing.ErrSessionRequired):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"Start this task before submitting answers.")

	case errors.Is(err, ErrEmptyResponse):
		httpx.ValidationError(w, tooShort)

	case errors.Is(err, ErrLimitReached):
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeLimitReached,
			"You have used all of today's practice sub-tests. They reset at midnight.")

	case errors.Is(err, billing.ErrGradingLimitReached):
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeLimitReached, billing.GradingLimitMessage)

	// A deadline that ran out while the provider was answering is the same
	// thing as an unavailable provider from the learner's side: nothing was
	// stored, and trying again is the right advice. Left to the default it
	// became a 500, which tells them we broke rather than that we were slow.
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
			"That took longer than we could wait. Your response was not lost - please try again.")

	case errors.Is(err, ai.ErrUnavailable), errors.Is(err, ai.ErrBadOutput):
		// Nothing was stored and no score was invented. The learner's work is
		// still in the browser, so they can retry.
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
			"Evaluation is unavailable right now. Your response was not lost - please try again shortly.")

	default:
		httpx.Internal(w, h.log, op, err)
	}
}
