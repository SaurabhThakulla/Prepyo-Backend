-- Extend the existing relationship ledger, not a parallel referral system.
CREATE UNIQUE INDEX users_referral_code_normalized_unique ON users (upper(btrim(referral_code)));
ALTER TABLE users ALTER COLUMN referral_code SET DEFAULT ('PREP-' || upper(substr(replace(gen_random_uuid()::text, '-', ''), 1, 10)));
ALTER TABLE referrals
    ADD COLUMN reward_payment_id UUID UNIQUE REFERENCES subscription_payments(id),
    ADD COLUMN reward_days INT NOT NULL DEFAULT 0 CHECK (reward_days IN (0, 3)),
    ADD CONSTRAINT referral_purchase_reward CHECK ((reward_payment_id IS NULL AND reward_days = 0) OR (reward_payment_id IS NOT NULL AND reward_days = 3));
ALTER TABLE subscription_payments DROP CONSTRAINT subscription_payments_status_check;
ALTER TABLE subscription_payments ADD CONSTRAINT subscription_payments_status_check
    CHECK (status IN ('pending', 'success', 'failed', 'cancelled', 'refunded'));
-- Repair still-live legacy milestone grants that never activated entitlements.
-- Do not replay expired awards or alter anyone's paid tier/expiry.
UPDATE users SET plan_id = 'pro', plan_started_at = CURRENT_DATE,
    role = CASE WHEN role = 'admin' THEN role ELSE 'taiyari' END, updated_at = now()
WHERE plan_id = 'free' AND bonus_pro_days > 0 AND plan_valid_until > CURRENT_DATE;
