package users

import (
	"context"
	"fmt"
)

// ReconcileRoles updates user roles based on their current subscription status.
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
