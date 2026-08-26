// Package admin exposes operational metrics.
//
// Every number here is counted from the database. Nothing is estimated or
// filled in with a plausible-looking constant: a dashboard that invents its own
// figures is worse than no dashboard.
//
// Routes are mounted behind RequireUser + RequireAdmin.
package admin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewHandler(db *pgxpool.Pool, log *slog.Logger) *Handler {
	return &Handler{db: db, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/metrics", h.metrics)
	return r
}

type metrics struct {
	Questions        int `json:"questions"`
	Mocks            int `json:"mocks"`
	Learners         int `json:"learners"`
	ActiveLast7Days  int `json:"activeLast7Days"`
	PracticeAttempts int `json:"practiceAttempts"`
	MockAttempts     int `json:"mockAttempts"`

	Evaluations      int `json:"evaluations"`
	EvaluationsToday int `json:"evaluationsToday"`
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	// MedianLatencyMS is the 50th percentile of real recorded latencies.
	MedianLatencyMS int `json:"medianLatencyMs"`
}

func (h *Handler) metrics(w http.ResponseWriter, r *http.Request) {
	m, err := h.read(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "admin.metrics", err)
		return
	}

	byModel, err := h.usageByModel(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "admin.usageByModel", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"metrics":      m,
		"usageByModel": byModel,
	})
}

func (h *Handler) read(ctx context.Context) (metrics, error) {
	var m metrics
	err := h.db.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM questions WHERE is_published),
			(SELECT count(*) FROM mocks),
			(SELECT count(*) FROM users WHERE role = 'learner'),
			(SELECT count(DISTINCT user_id) FROM practice_attempts WHERE created_at > now() - interval '7 days'),
			(SELECT count(*) FROM practice_attempts),
			(SELECT count(*) FROM mock_attempts),
			(SELECT count(*) FROM ai_evaluations),
			(SELECT count(*) FROM ai_evaluations WHERE created_at >= date_trunc('day', now())),
			(SELECT COALESCE(sum(prompt_tokens), 0) FROM ai_evaluations),
			(SELECT COALESCE(sum(completion_tokens), 0) FROM ai_evaluations),
			(SELECT COALESCE(percentile_cont(0.5) WITHIN GROUP (ORDER BY latency_ms), 0)::int FROM ai_evaluations)`,
	).Scan(&m.Questions, &m.Mocks, &m.Learners, &m.ActiveLast7Days, &m.PracticeAttempts,
		&m.MockAttempts, &m.Evaluations, &m.EvaluationsToday, &m.PromptTokens,
		&m.CompletionTokens, &m.MedianLatencyMS)
	if err != nil {
		return metrics{}, fmt.Errorf("read admin metrics: %w", err)
	}
	return m, nil
}

type modelUsage struct {
	Model            string `json:"model"`
	Requests         int    `json:"requests"`
	PromptTokens     int    `json:"promptTokens"`
	CompletionTokens int    `json:"completionTokens"`
}

// usageByModel supports cost tracking per model. Prices are not stored here:
// they change independently of the app, so cost is worked out from these token
// counts wherever the current rate card lives.
func (h *Handler) usageByModel(ctx context.Context) ([]modelUsage, error) {
	rows, err := h.db.Query(ctx, `
		SELECT model, count(*), COALESCE(sum(prompt_tokens), 0), COALESCE(sum(completion_tokens), 0)
		FROM ai_evaluations
		GROUP BY model
		ORDER BY count(*) DESC`)
	if err != nil {
		return nil, fmt.Errorf("read usage by model: %w", err)
	}
	defer rows.Close()

	usage := []modelUsage{}
	for rows.Next() {
		var u modelUsage
		if err := rows.Scan(&u.Model, &u.Requests, &u.PromptTokens, &u.CompletionTokens); err != nil {
			return nil, fmt.Errorf("scan usage by model: %w", err)
		}
		usage = append(usage, u)
	}
	return usage, rows.Err()
}
