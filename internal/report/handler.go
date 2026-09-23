// Package report records the issues learners raise from inside the app, and
// the conversation with the Prepyo team that follows.
//
// Reports used to be emailed through Gmail. That made every report depend on
// SMTP credentials being present and correct, and left no record of anything:
// a failed send, or an inbox nobody opened, and the report simply did not
// exist. They are rows now, and the admin dashboard works through them.
package report

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/notifications"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

const maxPath = 300

type Handler struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewHandler(db *pgxpool.Pool, log *slog.Logger) *Handler {
	return &Handler{db: db, log: log}
}

// Routes mounts the learner's side. limitWrites throttles the two routes that
// create something; reading your own reports is not rate limited beyond the
// signed-in default.
func (h *Handler) Routes(limitWrites func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.With(limitWrites).Post("/", h.submit)
	r.Get("/", h.mine)
	r.Get("/{id}", h.one)
	r.With(limitWrites).Post("/{id}/replies", h.reply)
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

	message := Clip(strings.TrimSpace(req.Message))
	if message == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
			"Please write a little about what went wrong.")
		return
	}

	path := strings.TrimSpace(req.Path)
	if path == "" {
		path = "unknown"
	}
	if len(path) > maxPath {
		path = path[:maxPath]
	}

	id, err := h.store(r.Context(), user.ID, user.Name, path, message)
	if err != nil {
		httpx.Internal(w, h.log, "report.submit", err)
		return
	}

	h.log.Info("issue reported", "user", user.Email, "path", path)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"id":      id,
		"message": "Thanks — your report is with the team. You will get a notification when they reply.",
	})
}

// store saves the report and tells every admin, in one transaction so there is
// never a report nobody was told about.
func (h *Handler) store(ctx context.Context, userID, userName, path, message string) (string, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin issue report: %w", err)
	}
	defer tx.Rollback(ctx)

	var id string
	if err := tx.QueryRow(ctx, `
		INSERT INTO issue_reports (user_id, path, message)
		VALUES ($1, $2, $3) RETURNING id`, userID, path, message).Scan(&id); err != nil {
		return "", fmt.Errorf("insert issue report: %w", err)
	}
	if err := notifications.SendToAdmins(ctx, tx, notifications.CreateParams{
		Type:      notifications.TypeReport,
		Title:     "New report from " + userName,
		Message:   Snippet(message),
		ActionURL: "/admin/reports?report=" + id,
	}); err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

// mine lists the learner's own reports, newest first, without the replies.
func (h *Handler) mine(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	rows, err := h.db.Query(r.Context(), `
		SELECT r.id, r.path, r.message, r.status, r.created_at,
		       (SELECT count(*) FROM issue_report_replies p WHERE p.report_id = r.id)
		FROM issue_reports r
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
		LIMIT 50`, user.ID)
	if err != nil {
		httpx.Internal(w, h.log, "report.mine", err)
		return
	}
	defer rows.Close()

	list := []Report{}
	for rows.Next() {
		rep := Report{UserID: user.ID, UserName: user.Name}
		var created time.Time
		if err := rows.Scan(&rep.ID, &rep.Path, &rep.Message, &rep.Status, &created, &rep.ReplyCount); err != nil {
			httpx.Internal(w, h.log, "report.mine.scan", err)
			return
		}
		rep.CreatedAt = created.Format(time.RFC3339)
		list = append(list, rep)
	}
	if err := rows.Err(); err != nil {
		httpx.Internal(w, h.log, "report.mine.rows", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"reports": list})
}

func (h *Handler) one(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	rep, err := Load(r.Context(), h.db, chi.URLParam(r, "id"), user.ID)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That report could not be found.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "report.one", err)
		return
	}
	rep.UserEmail = ""
	httpx.JSON(w, http.StatusOK, map[string]any{"report": rep})
}

type replyRequest struct {
	Message string `json:"message"`
}

// reply adds the learner's answer to their own open report and tells the team.
func (h *Handler) reply(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	var req replyRequest
	if !httpx.Decode(w, r, &req, h.log, "report.reply") {
		return
	}
	message := Clip(strings.TrimSpace(req.Message))
	if message == "" {
		httpx.ValidationError(w, map[string]string{"message": "Write a reply first."})
		return
	}

	ctx := r.Context()
	rep, err := Load(ctx, h.db, chi.URLParam(r, "id"), user.ID)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That report could not be found.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "report.reply.load", err)
		return
	}
	if rep.Status != "open" {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"This report is resolved. If something is still wrong, please send a new report.")
		return
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		httpx.Internal(w, h.log, "report.reply.begin", err)
		return
	}
	defer tx.Rollback(ctx)

	if err := AddReply(ctx, tx, rep.ID, user.ID, false, message); err != nil {
		httpx.Internal(w, h.log, "report.reply", err)
		return
	}
	if err := notifications.SendToAdmins(ctx, tx, notifications.CreateParams{
		Type:      notifications.TypeReport,
		Title:     user.Name + " replied to their report",
		Message:   Snippet(message),
		ActionURL: "/admin/reports?report=" + rep.ID,
	}); err != nil {
		httpx.Internal(w, h.log, "report.reply.notify", err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		httpx.Internal(w, h.log, "report.reply.commit", err)
		return
	}

	updated, err := Load(ctx, h.db, rep.ID, user.ID)
	if err != nil {
		httpx.Internal(w, h.log, "report.reply.reload", err)
		return
	}
	updated.UserEmail = ""
	httpx.JSON(w, http.StatusOK, map[string]any{"report": updated})
}
