package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/notifications"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	gateway *Gateway
	db      *pgxpool.Pool
	log     *slog.Logger
}

func NewHandler(gateway *Gateway, db *pgxpool.Pool, log *slog.Logger) *Handler {
	return &Handler{gateway: gateway, db: db, log: log}
}

// Daily tutor allowance. Every reply is a paid provider call, so a free
// account gets enough to try it and a paid plan enough for real study.
const (
	freeTutorMessagesPerDay = 10
	paidTutorMessagesPerDay = 100
)

// maxTutorMessageChars bounds one message sent to the tutor, and
// maxTutorContextChars the task it is told about. Without them one request
// could carry about a megabyte of text, all of it billed.
const (
	maxTutorMessageChars = 4000
	maxTutorContextChars = 4000
)

// tutorAllowance is the learner's daily limit, or -1 for no limit.
func tutorAllowance(user models.User) int {
	switch {
	case user.IsAdmin():
		return -1
	case user.HasActivePaidPlan() && user.Role != models.RoleSuru:
		return paidTutorMessagesPerDay
	default:
		return freeTutorMessagesPerDay
	}
}

// reserveTutorMessage records a message against today's allowance before the
// provider is called, so parallel requests cannot all slip under the limit. It
// returns the reservation id, or 0 when the allowance is used up.
func (h *Handler) reserveTutorMessage(ctx context.Context, user models.User) (int64, error) {
	limit := tutorAllowance(user)
	if limit < 0 {
		limit = 1 << 30
	}
	var id int64
	err := h.db.QueryRow(ctx, `
		INSERT INTO tutor_messages (user_id)
		SELECT $1
		WHERE (SELECT count(*) FROM tutor_messages
		       WHERE user_id = $1 AND created_at >= $2 AND created_at < $3) < $4
		RETURNING id`,
		user.ID, gamification.LocalDayStart(user), gamification.LocalDayEnd(user), limit).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("reserve tutor message: %w", err)
	}
	return id, nil
}

// releaseTutorMessage gives a reservation back when no reply was delivered,
// so a provider failure does not cost the learner part of their allowance.
func (h *Handler) releaseTutorMessage(ctx context.Context, id int64) {
	if _, err := h.db.Exec(ctx, `DELETE FROM tutor_messages WHERE id = $1`, id); err != nil {
		h.log.Warn("release tutor message", "error", err)
	}
}

// truncateRunes cuts s to at most n characters without splitting one.
func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/capabilities", h.capabilities)
	r.Post("/tutor", h.tutor)
	return r
}

// capabilities tells the app which kinds of feedback are actually available, so
// it can offer a speaking task as recording practice instead of letting a
// learner record an answer nothing is going to score.
func (h *Handler) capabilities(w http.ResponseWriter, r *http.Request) {
	// speakingTranscript is the fallback for a provider with no audio models:
	// the device transcribes, the text model judges the words, and nothing
	// pretends to have heard how they were said.
	httpx.JSON(w, http.StatusOK, map[string]any{
		"writing":            h.gateway.Available(),
		"speaking":           h.gateway.SpeakingAvailable(),
		"speakingTranscript": h.gateway.Available(),
		"tutor":              h.gateway.Available(),
	})
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
	latest := req.Messages[len(req.Messages)-1]
	if utf8.RuneCountInString(strings.TrimSpace(latest.Content)) > maxTutorMessageChars {
		httpx.ValidationError(w, map[string]string{
			"messages": fmt.Sprintf("Please keep your question under %d characters.", maxTutorMessageChars),
		})
		return
	}
	// Earlier turns are only context; long ones are cut rather than refused.
	for i := range req.Messages {
		req.Messages[i].Content = truncateRunes(req.Messages[i].Content, maxTutorMessageChars)
	}

	user := reqctx.MustUser(r.Context())
	exam := user.TargetExam
	if !exam.Valid() {
		exam = models.ExamPTE
	}

	reservation, err := h.reserveTutorMessage(r.Context(), user)
	if err != nil {
		httpx.Internal(w, h.log, "ai.tutor.reserve", err)
		return
	}
	if reservation == 0 {
		msg := fmt.Sprintf("You have used today's %d tutor messages. They reset at midnight.", tutorAllowance(user))
		if tutorAllowance(user) == freeTutorMessagesPerDay {
			msg += " Upgrade your plan for more."
		}
		if err := notifications.NewRepository(h.db).Notify(context.WithoutCancel(r.Context()), notifications.CreateParams{
			UserID: user.ID, Type: notifications.TypeLimit,
			Title: "Daily tutor limit reached", Message: msg,
			ActionURL: "/subscription",
			DedupeKey: "limit:tutor:" + gamification.LocalDay(user),
		}); err != nil {
			h.log.Warn("tutor limit notification failed", "error", err)
		}
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeLimitReached, msg)
		return
	}

	reply, usage, err := h.gateway.Tutor(r.Context(), TutorRequest{
		Exam:        exam,
		Messages:    req.Messages,
		TaskContext: truncateRunes(req.TaskContext, maxTutorContextChars),
	})
	if err != nil {
		// A fresh context: the request's may be the reason this failed.
		h.releaseTutorMessage(context.WithoutCancel(r.Context()), reservation)
		h.log.Error("ai.tutor failed", "err", err)
		if errors.Is(err, ErrUnavailable) || errors.Is(err, ErrBadOutput) {
			// The cause is logged above; the learner only needs to know to retry.
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
				"The tutor is unavailable right now. Please try again in a moment.")
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
