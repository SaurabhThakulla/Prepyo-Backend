package media

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// RetentionDays is how long a speaking recording is kept. The privacy page
// states the same number.
const RetentionDays = 30

var audioTypes = map[string]string{"wav": "audio/wav", "mp3": "audio/mpeg"}

// Recordings keeps a practice answer's recording so its owner can replay it
// from the score card, and deletes it after RetentionDays.
type Recordings struct {
	db    *pgxpool.Pool
	store Store
	log   *slog.Logger
}

// NewRecordings returns nil when there is no store: recordings are then
// dropped after scoring, as before.
func NewRecordings(db *pgxpool.Pool, store Store, log *slog.Logger) *Recordings {
	if store == nil {
		return nil
	}
	return &Recordings{db: db, store: store, log: log}
}

// Save uploads one evaluation's recording and notes its key. It runs after the
// learner already has their result, so a failure is logged, not returned.
func (rc *Recordings) Save(userID, evaluationID string, audio []byte, format string) {
	contentType, ok := audioTypes[format]
	if !ok || len(audio) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	key := fmt.Sprintf("recordings/%s/%s.%s", userID, evaluationID, format)
	if err := rc.store.Put(ctx, key, contentType, audio); err != nil {
		rc.log.Warn("recording not stored", "op", "media.saveRecording", "evaluationId", evaluationID, "error", err)
		return
	}
	tag, err := rc.db.Exec(ctx, `UPDATE ai_evaluations SET recording_key = $1 WHERE id = $2 AND user_id = $3`,
		key, evaluationID, userID)
	if err != nil || tag.RowsAffected() == 0 {
		rc.log.Warn("recording stored but not linked", "op", "media.saveRecording", "evaluationId", evaluationID, "error", err)
		_ = rc.store.Delete(ctx, key)
	}
}

// Serve streams a recording to the learner who made it, within RetentionDays.
func (rc *Recordings) Serve(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	var key string
	err := rc.db.QueryRow(r.Context(), `
		SELECT recording_key FROM ai_evaluations
		WHERE id::text = $1 AND user_id = $2 AND recording_key IS NOT NULL
		  AND created_at > now() - make_interval(days => $3)`,
		chi.URLParam(r, "id"), user.ID, RetentionDays).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That recording is no longer available.")
		return
	}
	if err != nil {
		httpx.Internal(w, rc.log, "media.serveRecording", err)
		return
	}

	body, contentType, err := rc.store.Get(r.Context(), key)
	if errors.Is(err, ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That recording is no longer available.")
		return
	}
	if err != nil {
		httpx.Internal(w, rc.log, "media.serveRecording", err)
		return
	}
	defer body.Close()
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	_, _ = io.Copy(w, body)
}

// Sweep deletes recordings past RetentionDays, so the limit holds even if the
// bucket has no lifecycle rule.
func (rc *Recordings) Sweep(ctx context.Context) {
	rows, err := rc.db.Query(ctx, `
		SELECT id::text, recording_key FROM ai_evaluations
		WHERE recording_key IS NOT NULL AND created_at <= now() - make_interval(days => $1)
		LIMIT 500`, RetentionDays)
	if err != nil {
		rc.log.Warn("recording sweep failed", "op", "media.sweep", "error", err)
		return
	}
	type expired struct{ id, key string }
	var due []expired
	for rows.Next() {
		var e expired
		if err := rows.Scan(&e.id, &e.key); err == nil {
			due = append(due, e)
		}
	}
	rows.Close()

	for _, e := range due {
		if err := rc.store.Delete(ctx, e.key); err != nil {
			rc.log.Warn("expired recording not deleted", "op", "media.sweep", "error", err)
			continue
		}
		_, _ = rc.db.Exec(ctx, `UPDATE ai_evaluations SET recording_key = NULL WHERE id::text = $1`, e.id)
	}
}

// RunSweeper sweeps once now and then hourly until ctx ends.
func (rc *Recordings) RunSweeper(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		rc.Sweep(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
