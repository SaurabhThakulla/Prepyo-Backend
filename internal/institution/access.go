package institution

import (
	"context"
	"fmt"

	"github.com/prepyo/backend/internal/database"
)

// AIBlocked reports whether the user is a student of an institution on the free plan.
// Such students get reading and listening only: no AI scoring, tutor or full mocks.
// A student who has bought a plan of their own is not blocked.
func AIBlocked(ctx context.Context, db database.DB, userID string) (bool, error) {
	var blocked bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM institution_students s
			JOIN institutions i ON i.id = s.institution_id
			JOIN users u ON u.id = s.user_id
			WHERE s.user_id = $1
			  AND (i.valid_until IS NULL OR i.valid_until <= CURRENT_DATE)
			  AND NOT (u.plan_id <> 'free' AND u.plan_valid_until > CURRENT_DATE)
		)`, userID).Scan(&blocked)
	if err != nil {
		return false, fmt.Errorf("check institution plan: %w", err)
	}
	return blocked, nil
}
