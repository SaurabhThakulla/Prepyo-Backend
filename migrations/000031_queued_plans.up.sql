-- ---------------------------------------------------------------------------
-- Plans bought while another one is still running
-- ---------------------------------------------------------------------------
--
-- users holds exactly one plan: plan_id, plan_started_at, plan_valid_until.
-- Approving a purchase wrote straight into those columns, so a learner sitting
-- on Udaan who bought Abhyas was moved DOWN to Abhyas and given a few extra
-- days. They paid and came out worse off.
--
-- A purchase approved while a plan is still live is parked here instead, and
-- starts when the current one ends. The learner can bring it forward if they
-- would rather have the new plan now, which forfeits whatever was left of the
-- old one — that is their call to make, not something to do to them quietly.

CREATE TABLE IF NOT EXISTS queued_plans (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id      TEXT NOT NULL REFERENCES plans(id),

    -- Days are frozen when the purchase is approved. A plan's length can be
    -- repriced later, and a learner is owed what they bought.
    days         INT NOT NULL CHECK (days > 0),

    payment_id   UUID REFERENCES subscription_payments(id) ON DELETE SET NULL,

    status       TEXT NOT NULL DEFAULT 'queued'
                     CHECK (status IN ('queued', 'activated', 'cancelled')),

    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    activated_at TIMESTAMPTZ,

    -- An activated row must record when, so "what is this learner entitled to"
    -- can always be answered from the table alone.
    CONSTRAINT queued_plans_activation_complete CHECK (
        status <> 'activated' OR activated_at IS NOT NULL
    )
);

-- The queue is read per learner, oldest first, and swept for everyone whose
-- current plan has lapsed.
CREATE INDEX IF NOT EXISTS idx_queued_plans_waiting
    ON queued_plans(user_id, created_at)
    WHERE status = 'queued';

-- One payment can only ever put one plan in the queue.
CREATE UNIQUE INDEX IF NOT EXISTS idx_queued_plans_payment
    ON queued_plans(payment_id)
    WHERE payment_id IS NOT NULL;
