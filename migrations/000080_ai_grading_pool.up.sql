-- AI gradings: a per-plan pool for the answers an AI model marks.
--
-- Why this exists. A practice sub-test paid for starting a question, and any
-- number of answers to it that day were then graded by the AI - every new
-- recording or edited essay was another paid model call. Reading and listening
-- cost nothing to mark, speaking and writing do, and one daily allowance
-- covered both. So the allowance could not keep plans profitable without
-- being cut for everyone.
--
-- Now there are two allowances:
--   - practice sub-tests (sub_tests_per_day), unchanged: starting any task;
--   - AI gradings (ai_gradings_per_period): each speaking or writing answer the
--     AI marks. A plan period is the paid plan's term; the free plan's is the
--     calendar month.
--
-- Usage is kept in fifths of a grading. An AI-marked answer is 5 units. Read
-- Aloud and Repeat Sentence are marked by word-matching against their text,
-- which costs only the transcription, so they are 1 unit. Sectional mocks that
-- are marked by the AI take their real cost from the pool when they start.
--
-- The limits are set so a typical learner keeps an 80%+ margin and no learner,
-- at the most the plan allows, takes it below 50%.

ALTER TABLE plans ADD COLUMN ai_gradings_per_period INT NOT NULL DEFAULT 0
    CHECK (ai_gradings_per_period >= 0);

UPDATE plans SET ai_gradings_per_period = 10 WHERE id = 'free';
UPDATE plans SET ai_gradings_per_period = 110 WHERE id = 'weekly';
UPDATE plans SET ai_gradings_per_period = 350 WHERE id = 'pro';

-- Udaan keeps unlimited practice sub-tests; full mocks are capped at 20 per
-- calendar month and its AI gradings at 1,100 per plan period.
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
    '350 AI gradings for speaking & writing',
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

-- The ledger. One row per charge: an AI-marked practice answer, or the start
-- of a sectional mock that the AI marks.
CREATE TABLE ai_grading_usage (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- Fifths of a grading: 5 for an AI-marked answer, 1 for a word-matched one.
    units      INT  NOT NULL CHECK (units > 0),
    -- What was charged: 'practice-writing', 'practice-speaking',
    -- 'practice-read-repeat', 'mock-pte-speaking', ...
    source     TEXT NOT NULL,
    -- The evaluation or mock session it paid for.
    ref        TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ai_grading_usage_user ON ai_grading_usage(user_id, created_at);
