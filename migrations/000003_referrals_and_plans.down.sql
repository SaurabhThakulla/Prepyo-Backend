-- 000003: Referrals and plans down migration

DROP TABLE IF EXISTS subscription_payments;
DROP TABLE IF EXISTS referrals;

DELETE FROM plans WHERE id = 'weekly';
ALTER TABLE plans DROP COLUMN IF EXISTS bonus_days;
ALTER TABLE plans DROP COLUMN IF EXISTS duration_days;

ALTER TABLE users DROP COLUMN IF EXISTS bonus_pro_days;
ALTER TABLE users DROP COLUMN IF EXISTS bonus_mock_tests;
ALTER TABLE users DROP COLUMN IF EXISTS referral_code;

ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_type_check;
ALTER TABLE notifications ADD CONSTRAINT notifications_type_check
    CHECK (type IN ('streak', 'evaluation', 'mission', 'system'));
