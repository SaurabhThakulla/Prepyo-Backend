-- How often a plan's AI gradings renew, as data rather than a rule in code.
--
--   'day'   - every day at the learner's local midnight
--   'plan'  - once per paid plan term, from plan_started_at to plan_valid_until
--   'month' - on the 1st of each calendar month (the free plan)
--
-- Taiyari moves from 350 per 33-day plan to 22 a day, resetting daily: up to
-- 726 in a plan. A typical learner (about 5 a day) keeps an 83-89% margin; a
-- learner who uses all 22 every day leaves about a 6% margin on AI costs. That
-- is a deliberate choice for a more generous daily allowance.

ALTER TABLE plans ADD COLUMN ai_gradings_period TEXT NOT NULL DEFAULT 'plan'
    CHECK (ai_gradings_period IN ('day', 'plan', 'month'));

UPDATE plans SET ai_gradings_period = 'month' WHERE id = 'free';
UPDATE plans SET ai_gradings_period = 'day', ai_gradings_per_period = 22 WHERE id = 'pro';

UPDATE plans SET features = ARRAY[
    '33 days access',
    '50 practice sub-tests per day',
    '22 AI gradings per day for speaking & writing',
    '5 full mock exams / month'
] WHERE id = 'pro';
