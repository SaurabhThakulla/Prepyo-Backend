package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSessionNotFound = errors.New("session not found or expired")

type sessionRepository struct {
	db *pgxpool.Pool
}

func (r *sessionRepository) create(ctx context.Context, tokenHash []byte, userID string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO sessions (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)`,
		tokenHash, userID, expiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// userIDByToken returns the owner of a live session. Expired rows are treated
// as missing so a stale cookie cannot be used even before cleanup runs.
func (r *sessionRepository) userIDByToken(ctx context.Context, tokenHash []byte) (string, error) {
	var userID string
	err := r.db.QueryRow(ctx, `
		SELECT user_id FROM sessions
		WHERE token_hash = $1 AND expires_at > now()`,
		tokenHash).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrSessionNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lookup session: %w", err)
	}
	return userID, nil
}

func (r *sessionRepository) delete(ctx context.Context, tokenHash []byte) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, tokenHash); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// deleteAllForUser revokes every session a user has. Used on password change
// and account deletion.
func (r *sessionRepository) deleteAllForUser(ctx context.Context, userID string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("revoke sessions: %w", err)
	}
	return nil
}

// deleteExpired clears rows that can no longer authenticate anything.
func (r *sessionRepository) deleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE expires_at < now()`)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}
