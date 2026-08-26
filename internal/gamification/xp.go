// Package gamification owns XP, levels, streaks and daily missions.
//
// Two rules shape everything here:
//   - the client never sends an XP amount, it only reports what it did;
//   - every award names the thing it is paying for, so replaying a request
//     cannot pay twice.
package gamification

import (
	"context"
	"fmt"
	"time"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

// Award amounts. Kept together so the economy can be read in one place rather
// than reverse-engineered from scattered call sites.
const (
	XPPracticeCorrect   = 40
	XPPracticeAttempted = 10
	XPMistakeResolved   = 30
	XPMockCompleted     = 300
	XPEvaluation        = 25
)

type Service struct{}

func NewService() *Service { return &Service{} }

// AwardParams describes one payment into the XP ledger.
type AwardParams struct {
	UserID string
	Amount int
	Reason string

	// SourceKey identifies what is being paid for, such as
	// "practice_attempt:<uuid>". A repeat award with the same key is ignored.
	SourceKey string
}

// Award records XP and adds it to the user's balance, both or neither.
//
// Pass a transaction as db when the award belongs with another write. The
// returned amount is 0 when this source has already been paid, which callers
// can report as "no XP this time" rather than an error.
func (s *Service) Award(ctx context.Context, db database.DB, p AwardParams) (int, error) {
	if p.Amount <= 0 {
		return 0, nil
	}

	tag, err := db.Exec(ctx, `
		INSERT INTO xp_transactions (user_id, amount, reason, source_key)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, source_key) DO NOTHING`,
		p.UserID, p.Amount, p.Reason, p.SourceKey)
	if err != nil {
		return 0, fmt.Errorf("record xp: %w", err)
	}

	// Nothing inserted means this source was already paid. Stopping here is
	// what makes a replayed request harmless.
	if tag.RowsAffected() == 0 {
		return 0, nil
	}

	if _, err := db.Exec(ctx, `UPDATE users SET xp = xp + $2, updated_at = now() WHERE id = $1`, p.UserID, p.Amount); err != nil {
		return 0, fmt.Errorf("add xp to balance: %w", err)
	}
	return p.Amount, nil
}

// History returns recent XP awards, newest first.
func (s *Service) History(ctx context.Context, db database.DB, userID string, limit int) ([]models.XPTransaction, error) {
	rows, err := db.Query(ctx, `
		SELECT id, amount, reason, created_at
		FROM xp_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("read xp history: %w", err)
	}
	defer rows.Close()

	transactions := make([]models.XPTransaction, 0, limit)
	for rows.Next() {
		var t models.XPTransaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.Reason, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan xp history: %w", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

// TouchStreak extends the streak when a learner does qualifying work.
//
// Signing in is not qualifying work; only callers that just recorded practice,
// a mock or an evaluation call this. Day boundaries use the learner's own
// timezone, so someone practising at 11pm in Kathmandu gets credit for that
// day rather than the next UTC one.
func (s *Service) TouchStreak(ctx context.Context, db database.DB, user models.User) (int, error) {
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		// A bad timezone in the database should not block a learner's
		// progress; fall back to the product's home timezone.
		location = nepalTime()
	}
	today := truncateToDay(time.Now().In(location))

	switch {
	case user.StreakLastActiveDate == nil:
		// First qualifying activity ever.
	case sameDay(*user.StreakLastActiveDate, today):
		// Already counted today; nothing to do.
		return user.StreakDays, nil
	case sameDay(user.StreakLastActiveDate.AddDate(0, 0, 1), today):
		// Yesterday, so the streak continues.
	default:
		// A gap; the streak restarts at 1.
		user.StreakDays = 0
	}

	newStreak := user.StreakDays + 1
	if _, err := db.Exec(ctx, `
		UPDATE users SET streak_days = $2, streak_last_active_date = $3, updated_at = now()
		WHERE id = $1`,
		user.ID, newStreak, today); err != nil {
		return user.StreakDays, fmt.Errorf("update streak: %w", err)
	}
	return newStreak, nil
}

func nepalTime() *time.Location {
	loc, err := time.LoadLocation("Asia/Kathmandu")
	if err != nil {
		// The tzdata package is missing from the image; UTC keeps the app
		// running with slightly off day boundaries.
		return time.UTC
	}
	return loc
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
