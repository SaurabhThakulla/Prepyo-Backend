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

// ActivateQueuedPlans starts plans that have been waiting for the current one
// to end.
//
// A plan bought while another was running is parked in queued_plans. Nothing
// writes to the user row when the running plan simply expires, so without this
// the learner would sit on the free tier with a paid plan waiting behind it.
//
// Runs immediately before ReconcileRoles: activating first means the role sweep
// sees the new plan, and a learner never drops to the free tier for an hour
// between one plan ending and the next beginning.
//
// Oldest first, one per user per pass. Two plans that both come due are started
// in the order they were bought, on successive ticks.
func (r *Repository) ActivateQueuedPlans(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		WITH due AS (
			SELECT DISTINCT ON (q.user_id) q.id, q.user_id, q.plan_id, q.days
			FROM queued_plans q
			JOIN users u ON u.id = q.user_id
			WHERE q.status = 'queued'
			  AND (u.plan_valid_until IS NULL OR u.plan_valid_until <= CURRENT_DATE)
			ORDER BY q.user_id, q.created_at
		),
		started AS (
			UPDATE users u
			SET plan_id = due.plan_id,
			    plan_started_at = CURRENT_DATE,
			    plan_valid_until = CURRENT_DATE + make_interval(days => due.days),
			    updated_at = now()
			FROM due
			WHERE u.id = due.user_id
			RETURNING due.id
		)
		UPDATE queued_plans q
		SET status = 'activated', activated_at = now()
		FROM started
		WHERE q.id = started.id`)
	if err != nil {
		return 0, fmt.Errorf("activate queued plans: %w", err)
	}
	return tag.RowsAffected(), nil
}
