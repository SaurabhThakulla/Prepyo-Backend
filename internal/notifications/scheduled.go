package notifications

import (
	"context"
	"fmt"
)

// RunScheduled sends the notices that depend on the date rather than on
// something a learner did, and prunes old ones. It runs hourly; every notice
// here carries a dedupe key, so running it again sends nothing twice.
func (r *Repository) RunScheduled(ctx context.Context) (int64, error) {
	var sent int64
	for _, job := range []struct {
		name  string
		query string
	}{
		{"plan ending", planEndingSQL},
		{"plan ended", planEndedSQL},
		{"streak reminder", streakReminderSQL},
		{"cleanup", cleanupSQL},
	} {
		tag, err := r.db.Exec(ctx, job.query)
		if err != nil {
			return sent, fmt.Errorf("%s: %w", job.name, err)
		}
		if job.name != "cleanup" {
			sent += tag.RowsAffected()
		}
	}
	return sent, nil
}

// planEndingSQL warns once, three days out, unless another plan is already
// queued to follow on.
const planEndingSQL = `
	INSERT INTO notifications (user_id, title, message, type, action_url, dedupe_key)
	SELECT u.id,
	       'Your ' || p.name || ' plan ends soon',
	       'It ends on ' || to_char(u.plan_valid_until, 'FMDD Mon') ||
	           '. Renew now to keep your full practice allowance.',
	       'plan', '/subscription',
	       'plan-ending:' || u.plan_valid_until
	FROM users u
	JOIN plans p ON p.id = u.plan_id
	WHERE u.plan_id <> 'free'
	  AND u.plan_valid_until BETWEEN CURRENT_DATE + 1 AND CURRENT_DATE + 3
	  AND NOT EXISTS (SELECT 1 FROM queued_plans q WHERE q.user_id = u.id AND q.status = 'queued')
	ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING`

// planEndedSQL runs after queued plans have been started, so a learner whose
// next plan just began is not told they have dropped to free. The week's
// window stops a first deploy from notifying everyone whose plan lapsed long ago.
const planEndedSQL = `
	INSERT INTO notifications (user_id, title, message, type, action_url, dedupe_key)
	SELECT u.id,
	       'Your ' || p.name || ' plan has ended',
	       'You are on the free plan now. Your progress is all still here; upgrade any time to pick up where you left off.',
	       'plan', '/subscription',
	       'plan-ended:' || u.plan_valid_until
	FROM users u
	JOIN plans p ON p.id = u.plan_id
	WHERE u.plan_id <> 'free'
	  AND u.plan_valid_until <= CURRENT_DATE
	  AND u.plan_valid_until > CURRENT_DATE - 7
	ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING`

// streakReminderSQL nudges, from 6pm in the learner's own zone, anyone whose
// streak is still alive from yesterday but who has not practised today. Zones
// Postgres does not know are skipped rather than failing the whole query.
const streakReminderSQL = `
	INSERT INTO notifications (user_id, title, message, type, action_url, dedupe_key)
	SELECT u.id,
	       'Keep your ' || u.streak_days || '-day streak going',
	       'You have not practised yet today. One short task keeps the streak alive.',
	       'streak', '/practice',
	       'streak:' || (now() AT TIME ZONE u.timezone)::date
	FROM users u
	WHERE u.role <> 'admin'
	  AND u.streak_days > 0
	  AND u.timezone IN (SELECT name FROM pg_timezone_names)
	  AND u.streak_last_active_date = (now() AT TIME ZONE u.timezone)::date - 1
	  AND (now() AT TIME ZONE u.timezone)::time >= '18:00'
	ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING`

// cleanupSQL keeps the inbox fast: read notices older than 90 days go.
const cleanupSQL = `
	DELETE FROM notifications
	WHERE read AND created_at < now() - interval '90 days'`
