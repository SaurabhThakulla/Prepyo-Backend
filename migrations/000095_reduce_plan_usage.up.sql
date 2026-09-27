-- Prices stay; AI-backed allowances come down so a learner who uses all of
-- them still leaves a margin (about NPR 0.5 per grading, NPR 5 per full mock
-- on OpenRouter prices):
--
--   free    AI gradings 10/month -> 5/month
--   weekly  AI gradings 110 -> 70                      (~52% at full use)
--   pro     AI gradings 22/day -> 8/day, mocks 5 -> 4  (~46%, was -34%)
--   elite   AI gradings 1,100 -> 600, mocks 20 -> 12   (~37%, was -65%)
--
-- The AI tutor's daily caps live in internal/ai/handler.go. Features are
-- shown as tiers (Limited / Standard / Extended / Maximum); exact numbers are
-- in the plan columns, not in the text learners compare.

UPDATE plans SET ai_gradings_per_period = 5 WHERE id = 'free';
UPDATE plans SET ai_gradings_per_period = 70 WHERE id = 'weekly';
UPDATE plans SET ai_gradings_per_period = 8, mock_tests_included = 4 WHERE id = 'pro';
UPDATE plans SET ai_gradings_per_period = 600, mock_tests_included = 12 WHERE id = 'elite';

UPDATE plans SET features = ARRAY[
    'Core practice tasks',
    'Limited AI gradings for speaking & writing',
    'Limited full mock exams',
    'Limited AI tutor'
] WHERE id = 'free';

UPDATE plans SET features = ARRAY[
    '7 days access',
    'Standard daily practice',
    'Standard AI gradings for speaking & writing',
    'Standard full mock exams',
    'Standard AI tutor'
] WHERE id = 'weekly';

UPDATE plans SET features = ARRAY[
    '33 days access',
    'Extended daily practice',
    'Extended AI gradings for speaking & writing',
    'Extended full mock exams',
    'Extended AI tutor'
] WHERE id = 'pro';

UPDATE plans SET features = ARRAY[
    '33 days full access',
    'Unlimited practice sub-tests',
    'Maximum AI gradings for speaking & writing',
    'Maximum full mock exams',
    'Maximum AI tutor',
    'Priority evaluation queue'
] WHERE id = 'elite';
