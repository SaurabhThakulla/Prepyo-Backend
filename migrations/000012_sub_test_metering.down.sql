-- Restores the plan rows exactly as 000003 left them, including the drifted
-- feature strings — a down migration puts the database back, it does not
-- improve on history.

-- The rename comes first: everything below writes the old column name.
ALTER TABLE plans RENAME COLUMN sub_tests_per_day TO ai_evaluations_per_day;

UPDATE plans SET
    name                   = 'Free Learner',
    ai_evaluations_per_day = 3,
    features               = ARRAY[
        'Core practice tasks',
        '3 AI evaluations per day',
        '1 full mock exam'
    ]
WHERE id = 'free';

UPDATE plans SET
    name                   = 'Weekly Sprint',
    ai_evaluations_per_day = 10,
    mock_tests_included    = 1,
    features               = ARRAY[
        '7 days intensive sprint',
        '10 AI evaluations per day',
        '1 full mock exam',
        'All practice question types'
    ]
WHERE id = 'weekly';

UPDATE plans SET
    name                   = 'Pro Prep',
    ai_evaluations_per_day = 15,
    mock_tests_included    = 3,
    features               = ARRAY[
        'Unlimited practice',
        '30 AI evaluations per day',
        '5 full mock exams',
        'Sentence-level rewrites'
    ]
WHERE id = 'pro';

UPDATE plans SET
    name                   = 'Elite Master',
    ai_evaluations_per_day = 25,
    features               = ARRAY[
        '90 days full access',
        '100 AI evaluations per day',
        '15 full mock exams',
        'Priority evaluation queue'
    ]
WHERE id = 'elite';
