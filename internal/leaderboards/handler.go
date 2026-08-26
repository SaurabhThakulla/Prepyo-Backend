package leaderboards

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// regions are the districts the board can be filtered by.
var regions = []string{"all", "Kathmandu", "Lalitpur", "Bhaktapur", "Pokhara", "Chitwan", "Butwal", "Dharan", "Biratnagar"}

type Handler struct {
	repo *Repository
	log  *slog.Logger
}

func NewHandler(repo *Repository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	query := r.URL.Query()
	page := httpx.ReadPage(r)

	exam := models.ExamType(query.Get("exam"))
	if exam == "" {
		exam = user.TargetExam
	}
	period := parsePeriod(query.Get("period"))

	entries, err := h.repo.List(r.Context(), ListParams{
		Exam:     exam,
		Region:   query.Get("region"),
		Period:   period,
		ViewerID: user.ID,
		Limit:    page.Limit,
	})
	if err != nil {
		httpx.Internal(w, h.log, "leaderboards.list", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"leaderboard": entries,
		"period":      period,
		"regions":     regions,
	})
}
