-- ---------------------------------------------------------------------------
-- Subscription tiers as roles, and the date a plan period began
-- ---------------------------------------------------------------------------
--
-- role stops being just an access flag and becomes the learner's tier as well:
-- admin, or one of the four plan names (Suru, Abhyas, Taiyari, Udaan). plan_id
-- still decides entitlement — limits are read from the plans table through
-- billing.planIsActive — so role is a denormalised copy kept in step with it.
--
-- Because it is a copy, it goes stale on its own: a paid plan expires by the
-- passing of a date, with no write to the row. cleanExpiredRoles in cmd/api
-- reconciles it hourly and at boot, and every purchase writes it directly.
--
-- plan_started_at is the start of the CURRENT period, not the first ever
-- subscription: buying again resets it, so it pairs with plan_valid_until to
-- describe the window a learner is in right now.

ALTER TABLE users ADD COLUMN IF NOT EXISTS plan_started_at DATE;

-- Backfill from the ledger where a payment exists for the plan the user is on,
-- and fall back to the day the account was created for everyone else.
UPDATE users u
SET plan_started_at = COALESCE(
    (SELECT max(sp.processed_at)::date
       FROM subscription_payments sp
      WHERE sp.user_id = u.id
        AND sp.plan_id = u.plan_id
        AND sp.status = 'success'),
    u.created_at::date)
WHERE u.plan_started_at IS NULL;

ALTER TABLE users ALTER COLUMN plan_started_at SET DEFAULT CURRENT_DATE;
ALTER TABLE users ALTER COLUMN plan_started_at SET NOT NULL;

-- The constraint has to come off before the rewrite: the new values are not
-- legal under the old CHECK.
ALTER TABLE users ALTER COLUMN role DROP DEFAULT;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;

-- An expired paid plan lands on 'suru', matching what billing already gives it.
UPDATE users SET role = CASE
    WHEN role = 'admin' THEN 'admin'
    WHEN plan_valid_until IS NULL OR plan_valid_until <= CURRENT_DATE THEN 'suru'
    WHEN plan_id = 'weekly' THEN 'abhyas'
    WHEN plan_id = 'pro'    THEN 'taiyari'
    WHEN plan_id = 'elite'  THEN 'udaan'
    ELSE 'suru'
END;

ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('admin', 'suru', 'abhyas', 'taiyari', 'udaan'));
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'suru';

-- Reconciling expired roles scans by expiry, and the leaderboard and admin
-- counts now filter on role <> 'admin'.
CREATE INDEX IF NOT EXISTS idx_users_plan_valid_until
    ON users(plan_valid_until)
    WHERE plan_valid_until IS NOT NULL;
