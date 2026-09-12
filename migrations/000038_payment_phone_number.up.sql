ALTER TABLE subscription_payments
    ADD COLUMN IF NOT EXISTS phone_number TEXT NOT NULL DEFAULT '';
