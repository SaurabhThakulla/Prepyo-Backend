package progress

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	db      *pgxpool.Pool
	service *Service
	log     *slog.Logger
}

func NewHandler(db *pgxpool.Pool, service *Service, log *slog.Logger) *Handler {
	return &Handler{db: db, service: service, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.summary)
	r.Get("/activity", h.activity)
	return r
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	estimate, err := h.service.Estimate(r.Context(), h.db, user)
	if err != nil {
		httpx.Internal(w, h.log, "progress.estimate", err)
		return
	}

	skills, err := h.service.Skills(r.Context(), h.db, user)
	if err != nil {
		httpx.Internal(w, h.log, "progress.skills", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"estimate": estimate,
		"skills":   skills,
	})
}

// activity backs the practice heatmap on the profile page.
func (h *Handler) activity(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	days := 365
	if v, err := strconv.Atoi(r.URL.Query().Get("days")); err == nil && v > 0 {
		days = v
	}

	summary, err := h.service.Activity(r.Context(), h.db, user, days)
	if err != nil {
		httpx.Internal(w, h.log, "progress.activity", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"activity": summary})
}
