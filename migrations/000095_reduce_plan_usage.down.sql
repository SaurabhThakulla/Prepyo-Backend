UPDATE plans SET ai_gradings_per_period = 10 WHERE id = 'free';
UPDATE plans SET ai_gradings_per_period = 110 WHERE id = 'weekly';
UPDATE plans SET ai_gradings_per_period = 22, mock_tests_included = 5 WHERE id = 'pro';
UPDATE plans SET ai_gradings_per_period = 1100, mock_tests_included = 20 WHERE id = 'elite';

UPDATE plans SET features = ARRAY[
    'Core practice tasks',
    '5 practice sub-tests per day',
    '10 AI gradings per month',
    '1 full mock exam'
] WHERE id = 'free';

UPDATE plans SET features = ARRAY[
    '7 days access',
    '40 practice sub-tests per day',
    '110 AI gradings for speaking & writing',
    '2 full mock exams'
] WHERE id = 'weekly';

UPDATE plans SET features = ARRAY[
    '33 days access',
    '50 practice sub-tests per day',
    '22 AI gradings per day for speaking & writing',
    '5 full mock exams / month'
] WHERE id = 'pro';

UPDATE plans SET features = ARRAY[
    '33 days full access',
    'Unlimited practice sub-tests',
    '1,100 AI gradings for speaking & writing',
    '20 full mock exams / month',
    'Unlimited AI tutor',
    'Priority evaluation queue'
] WHERE id = 'elite';
