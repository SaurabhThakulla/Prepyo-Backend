package notifications

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/reqctx"
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
	r.Post("/read-all", h.markAllRead)
	r.Post("/{notificationID}/read", h.markRead)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	page := httpx.ReadPage(r)

	list, total, unread, err := h.repo.List(r.Context(), user.ID, page.Limit, page.Offset)
	if err != nil {
		httpx.Internal(w, h.log, "notifications.list", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"notifications": list,
		"unreadCount":   unread,
		"pagination":    page.Meta(total),
	})
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	if err := h.repo.MarkRead(r.Context(), user.ID, chi.URLParam(r, "notificationID")); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That notification does not exist.")
			return
		}
		httpx.Internal(w, h.log, "notifications.markRead", err)
		return
	}
	httpx.JSON(w, http.StatusOK, nil)
}

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	if err := h.repo.MarkAllRead(r.Context(), user.ID); err != nil {
		httpx.Internal(w, h.log, "notifications.markAllRead", err)
		return
	}
	httpx.JSON(w, http.StatusOK, nil)
}
