-- Historical entitlements are deliberately not revoked on rollback.
ALTER TABLE referrals DROP CONSTRAINT referral_purchase_reward;
ALTER TABLE referrals DROP COLUMN reward_payment_id, DROP COLUMN reward_days;
ALTER TABLE users ALTER COLUMN referral_code DROP DEFAULT;
DROP INDEX users_referral_code_normalized_unique;
-- Keep the refunded payment status: converting financial history is unsafe.
