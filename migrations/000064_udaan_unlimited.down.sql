UPDATE plans SET
    duration_months = 3,
    duration_days   = 90,
    bonus_days      = 7,
    features        = ARRAY[
        '97 days full access',
        '60 practice sub-tests per day',
        '10 full mock exams per calendar month',
        'Priority evaluation queue'
    ]
WHERE id = 'elite';
ALTER TABLE plans DROP COLUMN IF EXISTS unlimited;
