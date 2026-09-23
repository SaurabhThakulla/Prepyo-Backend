-- Notifications become the product's inbox: payments, plan changes, daily
-- limits, report replies and announcements all land here, for learners and for
-- admins.

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_type_check
    CHECK (type IN ('streak', 'evaluation', 'mission', 'system', 'referral',
                    'payment', 'plan', 'limit', 'report', 'announcement'));

-- A notice that must go out at most once ("limit reached today", "plan ends
-- in 3 days") carries a key; the same key for the same user is dropped.
ALTER TABLE notifications ADD COLUMN dedupe_key TEXT;
CREATE UNIQUE INDEX idx_notifications_dedupe
    ON notifications(user_id, dedupe_key) WHERE dedupe_key IS NOT NULL;

-- The bell asks for the unread count every minute.
CREATE INDEX idx_notifications_unread ON notifications(user_id) WHERE NOT read;

-- A report is now a conversation: the learner's original message stays on
-- issue_reports, and every reply after it, from either side, is a row here.
CREATE TABLE issue_report_replies (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id  UUID NOT NULL REFERENCES issue_reports(id) ON DELETE CASCADE,
    author_id  UUID REFERENCES users(id) ON DELETE SET NULL,
    from_staff BOOLEAN NOT NULL,
    message    TEXT NOT NULL CHECK (length(btrim(message)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_issue_report_replies_report ON issue_report_replies(report_id, created_at);
