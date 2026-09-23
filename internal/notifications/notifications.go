package notifications

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("notification not found")

// uuidPattern screens ids from the URL, so a malformed one is "not found"
// rather than a database error.
var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Notification types. The database check constraint lists the same set.
const (
	TypeStreak       = "streak"
	TypeEvaluation   = "evaluation"
	TypeMission      = "mission"
	TypeSystem       = "system"
	TypeReferral     = "referral"
	TypePayment      = "payment"
	TypePlan         = "plan"
	TypeLimit        = "limit"
	TypeReport       = "report"
	TypeAnnouncement = "announcement"
)

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

type CreateParams struct {
	UserID    string
	Title     string
	Message   string
	Type      string
	ActionURL string
	// DedupeKey, when set, makes the notice once-only: a second one with the
	// same key for the same user is silently dropped.
	DedupeKey string
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Send writes one notification through db, so a caller inside a transaction
// gets it committed or rolled back with the rest of its work. It returns the
// new id, or "" when the dedupe key had already been used.
func Send(ctx context.Context, db database.DB, p CreateParams) (string, error) {
	var id string
	err := db.QueryRow(ctx, `
		INSERT INTO notifications (user_id, title, message, type, action_url, dedupe_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING
		RETURNING id`,
		p.UserID, p.Title, p.Message, p.Type, nullable(p.ActionURL), nullable(p.DedupeKey)).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("create notification: %w", err)
	}
	return id, nil
}

// SendToAdmins gives every admin account the same notice. UserID in p is
// ignored; a DedupeKey applies per admin.
func SendToAdmins(ctx context.Context, db database.DB, p CreateParams) error {
	_, err := db.Exec(ctx, `
		INSERT INTO notifications (user_id, title, message, type, action_url, dedupe_key)
		SELECT id, $1, $2, $3, $4, $5 FROM users WHERE role = 'admin'
		ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING`,
		p.Title, p.Message, p.Type, nullable(p.ActionURL), nullable(p.DedupeKey))
	if err != nil {
		return fmt.Errorf("notify admins: %w", err)
	}
	return nil
}

func (r *Repository) Create(ctx context.Context, db database.DB, p CreateParams) error {
	_, err := Send(ctx, db, p)
	return err
}

// Notify sends outside any caller transaction, on the repository's own pool.
// For notices that must survive the request failing, such as "you have hit
// today's limit", which is sent exactly when the request is refused.
func (r *Repository) Notify(ctx context.Context, p CreateParams) error {
	_, err := Send(ctx, r.db, p)
	return err
}

// Announce sends one notice to every learner and returns how many got it.
func (r *Repository) Announce(ctx context.Context, title, message, actionURL string) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		INSERT INTO notifications (user_id, title, message, type, action_url)
		SELECT id, $1, $2, 'announcement', $3 FROM users WHERE role <> 'admin'`,
		title, message, nullable(actionURL))
	if err != nil {
		return 0, fmt.Errorf("announce: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *Repository) List(ctx context.Context, userID string, limit, offset int) ([]models.Notification, int, int, error) {
	var total, unread int
	err := r.db.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE NOT read)
		FROM notifications WHERE user_id = $1 AND NOT dismissed`, userID).Scan(&total, &unread)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("count notifications: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, title, message, type, read, COALESCE(action_url, ''), created_at
		FROM notifications
		WHERE user_id = $1 AND NOT dismissed
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	list := []models.Notification{}
	for rows.Next() {
		var n models.Notification
		if err := rows.Scan(&n.ID, &n.Title, &n.Message, &n.Type, &n.Read, &n.ActionURL, &n.CreatedAt); err != nil {
			return nil, 0, 0, fmt.Errorf("scan notification: %w", err)
		}
		list = append(list, n)
	}
	return list, total, unread, rows.Err()
}

// MarkRead flags one notification. The user_id condition is the ownership
// check: another learner's id matches no rows.
func (r *Repository) MarkRead(ctx context.Context, userID, notificationID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE notifications SET read = TRUE
		WHERE id = $1 AND user_id = $2`, notificationID, userID)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID string) error {
	if _, err := r.db.Exec(ctx, `UPDATE notifications SET read = TRUE WHERE user_id = $1 AND NOT read`, userID); err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

// Delete removes one notification. As with MarkRead, the user_id condition is
// the ownership check.
//
// A once-only notice (one with a dedupe key) is hidden rather than deleted, so
// its key still stops it being sent again: clearing "daily limit reached" must
// not make it come straight back. Hidden rows are also marked read, so the
// scheduled cleanup prunes them like any other old read notice.
func (r *Repository) Delete(ctx context.Context, userID, notificationID string) error {
	if !uuidPattern.MatchString(notificationID) {
		return ErrNotFound
	}
	tag, err := r.db.Exec(ctx, `
		WITH target AS (
			SELECT id, dedupe_key IS NOT NULL AS keep_key
			FROM notifications WHERE id = $1 AND user_id = $2 AND NOT dismissed
		), gone AS (
			DELETE FROM notifications n USING target t
			WHERE n.id = t.id AND NOT t.keep_key
			RETURNING n.id
		), hidden AS (
			UPDATE notifications n SET dismissed = TRUE, read = TRUE
			FROM target t
			WHERE n.id = t.id AND t.keep_key
			RETURNING n.id
		)
		SELECT id FROM gone UNION ALL SELECT id FROM hidden`, notificationID, userID)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteAll clears the whole inbox, on the same terms as Delete.
func (r *Repository) DeleteAll(ctx context.Context, userID string) (int64, error) {
	var removed int64
	err := r.db.QueryRow(ctx, `
		WITH gone AS (
			DELETE FROM notifications
			WHERE user_id = $1 AND dedupe_key IS NULL
			RETURNING 1
		), hidden AS (
			UPDATE notifications SET dismissed = TRUE, read = TRUE
			WHERE user_id = $1 AND dedupe_key IS NOT NULL AND NOT dismissed
			RETURNING 1
		)
		SELECT (SELECT count(*) FROM gone) + (SELECT count(*) FROM hidden)`, userID).Scan(&removed)
	if err != nil {
		return 0, fmt.Errorf("clear notifications: %w", err)
	}
	return removed, nil
}

