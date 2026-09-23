package report

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
)

var ErrNotFound = errors.New("report not found")

// Report is one issue a learner raised, with the conversation that followed.
type Report struct {
	ID         string  `json:"id"`
	UserID     string  `json:"userId"`
	UserName   string  `json:"userName"`
	UserEmail  string  `json:"userEmail,omitempty"`
	Path       string  `json:"path"`
	Message    string  `json:"message"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"createdAt"`
	ResolvedAt string  `json:"resolvedAt,omitempty"`
	ReplyCount int     `json:"replyCount"`
	Replies    []Reply `json:"replies,omitempty"`
}

// Reply is one message after the first, from the learner or the Prepyo team.
type Reply struct {
	ID         string `json:"id"`
	FromStaff  bool   `json:"fromStaff"`
	AuthorName string `json:"authorName"`
	Message    string `json:"message"`
	CreatedAt  string `json:"createdAt"`
}

// MaxMessageRunes bounds one message, counted in characters so a Nepali
// message is not cut through the middle of a letter.
const MaxMessageRunes = 4000

// Clip trims a message to MaxMessageRunes characters.
func Clip(s string) string {
	if utf8.RuneCountInString(s) <= MaxMessageRunes {
		return s
	}
	return string([]rune(s)[:MaxMessageRunes])
}

// Load reads one report and its replies. When ownerID is set, a report that
// belongs to somebody else is reported as not found.
func Load(ctx context.Context, db database.DB, id, ownerID string) (Report, error) {
	var rep Report
	var created time.Time
	var resolved *time.Time
	err := db.QueryRow(ctx, `
		SELECT r.id, r.user_id, u.name, COALESCE(u.email, ''), r.path, r.message, r.status,
		       r.created_at, r.resolved_at
		FROM issue_reports r
		JOIN users u ON u.id = r.user_id
		WHERE r.id = $1 AND ($2 = '' OR r.user_id::text = $2)`, id, ownerID).
		Scan(&rep.ID, &rep.UserID, &rep.UserName, &rep.UserEmail, &rep.Path, &rep.Message,
			&rep.Status, &created, &resolved)
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, ErrNotFound
	}
	if err != nil {
		return Report{}, fmt.Errorf("load report: %w", err)
	}
	rep.CreatedAt = created.Format(time.RFC3339)
	if resolved != nil {
		rep.ResolvedAt = resolved.Format(time.RFC3339)
	}

	// Staff replies are signed "Prepyo team" rather than with an admin's own
	// name: the learner is talking to the product, not to one person.
	rows, err := db.Query(ctx, `
		SELECT p.id, p.from_staff,
		       CASE WHEN p.from_staff THEN 'Prepyo team' ELSE COALESCE(u.name, 'You') END,
		       p.message, p.created_at
		FROM issue_report_replies p
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.report_id = $1
		ORDER BY p.created_at`, id)
	if err != nil {
		return Report{}, fmt.Errorf("load report replies: %w", err)
	}
	defer rows.Close()

	rep.Replies = []Reply{}
	for rows.Next() {
		var reply Reply
		var at time.Time
		if err := rows.Scan(&reply.ID, &reply.FromStaff, &reply.AuthorName, &reply.Message, &at); err != nil {
			return Report{}, fmt.Errorf("scan report reply: %w", err)
		}
		reply.CreatedAt = at.Format(time.RFC3339)
		rep.Replies = append(rep.Replies, reply)
	}
	rep.ReplyCount = len(rep.Replies)
	return rep, rows.Err()
}

// AddReply appends one message to a report's conversation.
func AddReply(ctx context.Context, db database.DB, reportID, authorID string, fromStaff bool, message string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO issue_report_replies (report_id, author_id, from_staff, message)
		VALUES ($1, $2, $3, $4)`, reportID, authorID, fromStaff, message)
	if err != nil {
		return fmt.Errorf("add report reply: %w", err)
	}
	return nil
}

// Snippet shortens a message for a notification line.
func Snippet(s string) string {
	const max = 80
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max-1]) + "…"
}
