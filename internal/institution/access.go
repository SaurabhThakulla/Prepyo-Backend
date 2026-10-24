package institution

import (
	"context"
	"fmt"

	"github.com/prepyo/backend/internal/database"
)

// studentOf matches the user's row in institution_students (s), with its institution (i) and user (u).
const studentOf = `
	FROM institution_students s
	JOIN institutions i ON i.id = s.institution_id
	JOIN users u ON u.id = s.user_id
	WHERE s.user_id = $1
	  AND NOT (u.plan_id <> 'free' AND u.plan_valid_until > CURRENT_DATE)`

// planRunning is true while the institution has a plan whose end date has not passed.
const planRunning = `(i.plan IS NOT NULL AND i.valid_until > CURRENT_DATE)`

// AIBlocked reports whether the user is a student of an institution with no running plan.
// Such students get reading and listening only: no AI scoring, tutor or full mocks.
// A student who has bought a plan of their own is not blocked.
func AIBlocked(ctx context.Context, db database.DB, userID string) (bool, error) {
	return exists(ctx, db, `SELECT EXISTS (SELECT 1`+studentOf+` AND NOT `+planRunning+`)`, userID)
}

// PlanCovers reports whether the user gets the institution allowance: a student of an
// institution with a running plan, without a paid plan of their own.
func PlanCovers(ctx context.Context, db database.DB, userID string) (bool, error) {
	return exists(ctx, db, `SELECT EXISTS (SELECT 1`+studentOf+` AND `+planRunning+`)`, userID)
}

func exists(ctx context.Context, db database.DB, query, userID string) (bool, error) {
	var found bool
	if err := db.QueryRow(ctx, query, userID).Scan(&found); err != nil {
		return false, fmt.Errorf("check institution plan: %w", err)
	}
	return found, nil
}
