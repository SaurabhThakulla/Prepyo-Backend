DROP TABLE IF EXISTS issue_report_replies;
DROP INDEX IF EXISTS idx_notifications_unread;
DROP INDEX IF EXISTS idx_notifications_dedupe;
ALTER TABLE notifications DROP COLUMN IF EXISTS dedupe_key;
DELETE FROM notifications
WHERE type NOT IN ('streak', 'evaluation', 'mission', 'system', 'referral');
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_type_check
    CHECK (type IN ('streak', 'evaluation', 'mission', 'system', 'referral'));
