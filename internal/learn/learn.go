// Package learn keeps each learner's progress through the Learn Korean course,
// so it follows them from one device to another.
//
// The course itself lives in the web app, and so does the shape of a learner's
// record: which lessons are done, XP, the streak and the word review schedule.
// The browser merges its copy with the one stored here; the server keeps one
// JSON object per learner and checks only that it is an object of sane size.
package learn

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// MaxProgressBytes caps one learner's record. A finished course with every
// word under review is a few tens of kilobytes.
const MaxProgressBytes = 256 << 10

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

// Load returns the learner's stored progress, or nil if there is none yet.
func (r *Repository) Load(ctx context.Context, userID string) (json.RawMessage, error) {
	var data []byte
	err := r.db.QueryRow(ctx, `SELECT data FROM learn_progress WHERE user_id = $1`, userID).Scan(&data)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return data, err
}

// Save replaces the learner's stored progress.
func (r *Repository) Save(ctx context.Context, userID string, data json.RawMessage) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO learn_progress (user_id, data) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET data = EXCLUDED.data, updated_at = now()`, userID, data)
	return err
}

type Handler struct {
	repo *Repository
	log  *slog.Logger
}

func NewHandler(repo *Repository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/progress", h.load)
	r.Put("/progress", h.save)
	return r
}

func (h *Handler) load(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	data, err := h.repo.Load(r.Context(), user.ID)
	if err != nil {
		httpx.Internal(w, h.log, "learn.load", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpx.JSON(w, http.StatusOK, map[string]any{"progress": data})
}

func (h *Handler) save(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	var req struct {
		Progress json.RawMessage `json:"progress"`
	}
	if !httpx.DecodeLimit(w, r, &req, h.log, "learn.save", MaxProgressBytes) {
		return
	}
	if !isObject(req.Progress) {
		httpx.ValidationError(w, map[string]string{"progress": "Send the progress as a JSON object."})
		return
	}
	if err := h.repo.Save(r.Context(), user.ID, req.Progress); err != nil {
		httpx.Internal(w, h.log, "learn.save", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"saved": true})
}

// isObject reports whether raw is a JSON object, the only shape stored.
func isObject(raw json.RawMessage) bool {
	var value map[string]json.RawMessage
	return len(raw) > 0 && json.Unmarshal(raw, &value) == nil && value != nil
}
