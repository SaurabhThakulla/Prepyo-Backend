UPDATE plans SET ai_gradings_per_period = 350 WHERE id = 'pro';
UPDATE plans SET features = ARRAY[
    '33 days access',
    '50 practice sub-tests per day',
    '350 AI gradings for speaking & writing',
    '5 full mock exams / month'
] WHERE id = 'pro';

ALTER TABLE plans DROP COLUMN IF EXISTS ai_gradings_period;
