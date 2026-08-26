package mistakes

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	db   *pgxpool.Pool
	repo *Repository
	xp   *gamification.Service
	log  *slog.Logger
}

func NewHandler(db *pgxpool.Pool, repo *Repository, xp *gamification.Service, log *slog.Logger) *Handler {
	return &Handler{db: db, repo: repo, xp: xp, log: log}
}

// Routes are all private; the caller mounts them behind RequireUser.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Post("/{mistakeID}/resolve", h.resolve)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	query := r.URL.Query()
	page := httpx.ReadPage(r)

	list, total, err := h.repo.List(r.Context(), ListParams{
		UserID:         user.ID,
		Exam:           models.ExamType(query.Get("exam")),
		Skill:          models.SkillType(query.Get("skill")),
		UnresolvedOnly: query.Get("unresolved") == "true",
		Limit:          page.Limit,
		Offset:         page.Offset,
	})
	if err != nil {
		httpx.Internal(w, h.log, "mistakes.list", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"mistakes":   list,
		"pagination": page.Meta(total),
	})
}

func (h *Handler) resolve(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	mistakeID := chi.URLParam(r, "mistakeID")

	tx, err := h.db.Begin(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "mistakes.resolve.begin", err)
		return
	}
	defer tx.Rollback(r.Context())

	if err := h.repo.Resolve(r.Context(), tx, user.ID, mistakeID); err != nil {
		if errors.Is(err, ErrNotFound) {
			// Covers both "no such mistake" and "not yours": the response is
			// the same either way, so this cannot be used to probe other
			// learners' data.
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That mistake is not in your bank, or is already resolved.")
			return
		}
		httpx.Internal(w, h.log, "mistakes.resolve", err)
		return
	}

	awarded, err := h.xp.Award(r.Context(), tx, gamification.AwardParams{
		UserID:    user.ID,
		Amount:    gamification.XPMistakeResolved,
		Reason:    "Resolved a mistake",
		SourceKey: "mistake_resolved:" + mistakeID,
	})
	if err != nil {
		httpx.Internal(w, h.log, "mistakes.resolve.xp", err)
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		httpx.Internal(w, h.log, "mistakes.resolve.commit", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"xpAwarded": awarded})
}
