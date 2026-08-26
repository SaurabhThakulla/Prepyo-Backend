package exams

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/pkg/httpx"
)

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
	versions, err := h.repo.List(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "exams.list", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"examVersions": versions})
}
