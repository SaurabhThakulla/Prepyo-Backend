package gamification

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
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

// Routes exposes reads only.
//
// There is deliberately no endpoint to add XP or complete a mission. Both
// happen as a side effect of real work in the practice, mocks and evaluations
// modules, which is what keeps them unforgeable.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.summary)
	r.Get("/history", h.history)
	return r
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	missions, err := h.service.TodayMissions(r.Context(), h.db, user)
	if err != nil {
		httpx.Internal(w, h.log, "gamification.summary", err)
		return
	}

	level := models.LevelForXP(user.XP)
	httpx.JSON(w, http.StatusOK, map[string]any{
		"xp":            user.XP,
		"level":         level,
		"xpToNextLevel": level*models.XPPerLevel - user.XP,
		"streakDays":    user.StreakDays,
		"dailyMissions": missions,
	})
}

func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	page := httpx.ReadPage(r)

	transactions, err := h.service.History(r.Context(), h.db, user.ID, page.Limit)
	if err != nil {
		httpx.Internal(w, h.log, "gamification.history", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"transactions": transactions})
}
