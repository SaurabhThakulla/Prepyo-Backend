-- Udaan becomes a one-month plan (30 days + 3 bonus) with no daily practice
-- limit and no mock limit. "Unlimited" is a flag rather than a very large
-- number, so the app can say "Unlimited" instead of "3 / 9999".
--
-- Learners already on Udaan keep the end date they paid for, and payments
-- already waiting for review keep the days recorded when they were made
-- (subscription_payments.effective_days), so nobody loses time they bought.
ALTER TABLE plans ADD COLUMN unlimited BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE plans SET
    duration_months = 1,
    duration_days   = 30,
    bonus_days      = 3,
    unlimited       = TRUE,
    features        = ARRAY[
        '33 days full access',
        'Unlimited practice sub-tests',
        'Unlimited full mock exams',
        'Unlimited AI tutor',
        'Priority evaluation queue'
    ]
WHERE id = 'elite';
