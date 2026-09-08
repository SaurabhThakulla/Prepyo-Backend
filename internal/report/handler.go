// Package report records the issues learners raise from inside the app.
//
// Reports used to be emailed through Gmail. That made every report depend on
// SMTP credentials being present and correct, and left no record of anything:
// a failed send, or an inbox nobody opened, and the report simply did not
// exist. They are rows now, and the admin dashboard works through them.
package report

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

const (
	// maxMessage is generous: a learner describing a scoring problem writes
	// more than a sentence, and truncating loses the detail that makes the
	// report worth having.
	maxMessage = 4000
	maxPath    = 300
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
	r.Post("/", h.submit)
	return r
}

type submitRequest struct {
	Message string `json:"message"`
	Path    string `json:"path"`
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	var req submitRequest
	if !httpx.Decode(w, r, &req, h.log, "report.submit") {
		return
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"Please write a little about what went wrong.")
		return
	}
	if len(message) > maxMessage {
		message = message[:maxMessage]
	}

	path := strings.TrimSpace(req.Path)
	if path == "" {
		path = "unknown"
	}
	if len(path) > maxPath {
		path = path[:maxPath]
	}

	if err := h.store(r.Context(), user.ID, path, message); err != nil {
		httpx.Internal(w, h.log, "report.submit", err)
		return
	}

	h.log.Info("issue reported", "user", user.Email, "path", path)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"message": "Thanks — your report is with the team.",
	})
}

func (h *Handler) store(ctx context.Context, userID, path, message string) error {
	_, err := h.db.Exec(ctx, `
		INSERT INTO issue_reports (user_id, path, message)
		VALUES ($1, $2, $3)`, userID, path, message)
	if err != nil {
		return fmt.Errorf("insert issue report: %w", err)
	}
	return nil
}
