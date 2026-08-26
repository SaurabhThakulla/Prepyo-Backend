// Package leaderboards ranks learners by XP.
//
// Ranking is computed in the database from the XP ledger, so a learner cannot
// place themselves. XP itself is only awarded for real work (see
// internal/gamification), which is what keeps the board from rewarding clicks.
package leaderboards

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
)

// Period selects the window XP is counted over.
type Period string

const (
	PeriodWeekly  Period = "weekly"
	PeriodMonthly Period = "monthly"
	PeriodAllTime Period = "all-time"
)

func (p Period) since() *time.Time {
	now := time.Now()
	switch p {
	case PeriodWeekly:
		t := now.AddDate(0, 0, -7)
		return &t
	case PeriodMonthly:
		t := now.AddDate(0, -1, 0)
		return &t
	default:
		return nil
	}
}

func parsePeriod(raw string) Period {
	switch Period(raw) {
	case PeriodWeekly:
		return PeriodWeekly
	case PeriodMonthly:
		return PeriodMonthly
	default:
		return PeriodAllTime
	}
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type ListParams struct {
	Exam     models.ExamType
	Region   string
	Period   Period
	ViewerID string
	Limit    int
}

// List returns the ranked board.
//
// For a bounded period the XP total is summed from the ledger rather than read
// off the user row, since the row holds a lifetime total.
func (r *Repository) List(ctx context.Context, p ListParams) ([]models.LeaderboardEntry, error) {
	rows, err := r.db.Query(ctx, `
		WITH period_xp AS (
			SELECT u.id,
			       CASE WHEN $3::timestamptz IS NULL THEN u.xp
			            ELSE COALESCE((
			                SELECT sum(t.amount) FROM xp_transactions t
			                WHERE t.user_id = u.id AND t.created_at >= $3
			            ), 0)
			       END AS score
			FROM users u
			WHERE u.role = 'learner'
			  AND ($1 = '' OR u.target_exam = $1)
			  AND ($2 = '' OR $2 = 'all' OR u.nepal_region = $2)
		)
		SELECT u.id, u.name, u.nepal_region, u.target_exam, p.score, u.streak_days
		FROM period_xp p
		JOIN users u ON u.id = p.id
		WHERE p.score > 0
		ORDER BY p.score DESC, u.created_at
		LIMIT $4`,
		p.Exam, p.Region, p.Period.since(), p.Limit)
	if err != nil {
		return nil, fmt.Errorf("list leaderboard: %w", err)
	}
	defer rows.Close()

	entries := []models.LeaderboardEntry{}
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.UserID, &e.Name, &e.NepalRegion, &e.Exam, &e.XP, &e.StreakDays); err != nil {
			return nil, fmt.Errorf("scan leaderboard row: %w", err)
		}
		e.Rank = len(entries) + 1
		e.Level = models.LevelForXP(e.XP)
		e.IsYou = e.UserID == p.ViewerID
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
