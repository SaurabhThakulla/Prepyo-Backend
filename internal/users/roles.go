package users

import (
	"context"
	"fmt"
)

// ReconcileRoles brings the role column back in line with what each user's plan
// currently entitles them to.
//
// role is a denormalised copy of plan state, and a subscription lapses by a
// date passing rather than by anything writing to the row. Without this, a
// learner whose plan expired last night would still read as 'taiyari'. It is
// the price of storing the tier instead of deriving it, so it runs hourly and
// at boot.
//
// Two directions are covered: expired paid plans fall back to the free tier,
// and a live paid plan whose role drifted (a restored backup, a manual edit) is
// promoted back. Admin is never touched — it is an access level, not a tier.
func (r *Repository) ReconcileRoles(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE users SET role = CASE
		    WHEN plan_valid_until IS NULL OR plan_valid_until <= CURRENT_DATE THEN 'suru'
		    WHEN plan_id = 'weekly' THEN 'abhyas'
		    WHEN plan_id = 'pro'    THEN 'taiyari'
		    WHEN plan_id = 'elite'  THEN 'udaan'
		    ELSE 'suru'
		END,
		updated_at = now()
		WHERE role <> 'admin'
		  AND role IS DISTINCT FROM CASE
		    WHEN plan_valid_until IS NULL OR plan_valid_until <= CURRENT_DATE THEN 'suru'
		    WHEN plan_id = 'weekly' THEN 'abhyas'
		    WHEN plan_id = 'pro'    THEN 'taiyari'
		    WHEN plan_id = 'elite'  THEN 'udaan'
		    ELSE 'suru'
		END`)
	if err != nil {
		return 0, fmt.Errorf("reconcile roles: %w", err)
	}
	return tag.RowsAffected(), nil
}
