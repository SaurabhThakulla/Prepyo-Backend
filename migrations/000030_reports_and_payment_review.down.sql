DROP INDEX IF EXISTS idx_subscription_payments_pending;

ALTER TABLE subscription_payments
    DROP COLUMN IF EXISTS proof_image,
    DROP COLUMN IF EXISTS proof_image_type,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS review_note;

DROP INDEX IF EXISTS idx_issue_reports_user;
DROP INDEX IF EXISTS idx_issue_reports_open;
DROP TABLE IF EXISTS issue_reports;
