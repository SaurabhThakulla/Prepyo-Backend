-- ---------------------------------------------------------------------------
-- Issue reports, and payments that wait for a human
-- ---------------------------------------------------------------------------
--
-- Two things move out of "fire and forget" and into a queue an admin works
-- through.
--
-- Reports used to be emailed through Gmail and then forgotten: if the mail
-- failed, or nobody read the inbox, the report was gone. There was no record a
-- learner had ever reported anything. They are rows now.
--
-- Payments used to be granted the moment a learner typed a transaction id.
-- Nothing checked that the id was real or that any money arrived, so the plan
-- was free to anyone who guessed the endpoint. A payment now waits as 'pending'
-- with its proof attached until an admin approves it.

CREATE TABLE IF NOT EXISTS issue_reports (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- The page the learner was on. Kept because most reports describe what is
    -- on screen rather than where it is.
    path         TEXT NOT NULL DEFAULT '',
    message      TEXT NOT NULL,

    status       TEXT NOT NULL DEFAULT 'open'
                     CHECK (status IN ('open', 'resolved')),
    resolved_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    resolved_at  TIMESTAMPTZ,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT issue_reports_message_not_blank CHECK (length(btrim(message)) > 0),
    -- A resolved row must say who closed it and when, so the queue cannot hold
    -- a report that is done with no record of who dealt with it.
    CONSTRAINT issue_reports_resolution_complete CHECK (
        status = 'open' OR (resolved_at IS NOT NULL)
    )
);

-- The admin queue reads open reports newest first; that is the only listing.
CREATE INDEX IF NOT EXISTS idx_issue_reports_open
    ON issue_reports(created_at DESC)
    WHERE status = 'open';

CREATE INDEX IF NOT EXISTS idx_issue_reports_user
    ON issue_reports(user_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- Payment proof and review
-- ---------------------------------------------------------------------------

ALTER TABLE subscription_payments
    -- The screenshot of the transfer, stored inline the way profile pictures
    -- already are. Small, read rarely, and one less thing to host.
    ADD COLUMN IF NOT EXISTS proof_image      BYTEA,
    ADD COLUMN IF NOT EXISTS proof_image_type TEXT,
    ADD COLUMN IF NOT EXISTS reviewed_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reviewed_at      TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS review_note      TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_subscription_payments_pending
    ON subscription_payments(created_at DESC)
    WHERE status = 'pending';
