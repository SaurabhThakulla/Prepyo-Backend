// Package notifications stores in-app messages for a learner.
package notifications

import (
	"context"
	"errors"
	"fmt"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("notification not found")

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
}

func (r *Repository) Create(ctx context.Context, db database.DB, p CreateParams) error {
	var actionURL *string
	if p.ActionURL != "" {
		actionURL = &p.ActionURL
	}
	_, err := db.Exec(ctx, `
		INSERT INTO notifications (user_id, title, message, type, action_url)
		VALUES ($1, $2, $3, $4, $5)`,
		p.UserID, p.Title, p.Message, p.Type, actionURL)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context, userID string, limit, offset int) ([]models.Notification, int, int, error) {
	var total, unread int
	err := r.db.QueryRow(ctx, `
		SELECT count(*), count(*) FILTER (WHERE NOT read)
		FROM notifications WHERE user_id = $1`, userID).Scan(&total, &unread)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("count notifications: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, title, message, type, read, COALESCE(action_url, ''), created_at
		FROM notifications
		WHERE user_id = $1
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
