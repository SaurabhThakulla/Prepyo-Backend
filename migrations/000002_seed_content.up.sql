-- Core reference data every database needs: the exam versions, the base plans
-- and the daily mission catalogue.
--
-- The starter question bank that used to live here was removed, and this file
-- went with it — which left a fresh database unable to migrate: 000007 inserts
-- a mock that references 'ielts-2026-01', and the startup integrity check
-- requires these plans. Only the reference rows are back. The name is the
-- original one, so a database that already applied it skips this file.
--
-- This file contains no learner data and no questions.

INSERT INTO exam_versions (id, exam, label, description, min_score, max_score, score_step, is_current) VALUES
    ('pte-2026-01',   'PTE',   'PTE Academic 2026.1',   'Pearson Test of English Academic.', 10, 90, 1,   TRUE),
    ('ielts-2026-01', 'IELTS', 'IELTS Academic 2026.1', 'IELTS Academic, band scale.',        0,  9, 0.5, TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO plans (id, name, price_npr, duration_months, features, ai_evaluations_per_day, mock_tests_included, is_popular, sort_order) VALUES
    ('free', 'Free Learner', 0, 0,
        ARRAY['Core practice tasks', '3 AI evaluations per day', '1 full mock exam'],
        3, 1, FALSE, 1),
    ('pro', 'Pro Prep', 1499, 1,
        ARRAY['Unlimited practice', '30 AI evaluations per day', '5 full mock exams', 'Sentence-level rewrites'],
        30, 5, TRUE, 2),
    ('elite', 'Elite Master', 2999, 3,
        ARRAY['90 days full access', '100 AI evaluations per day', '15 full mock exams', 'Priority evaluation queue'],
        100, 15, FALSE, 3)
ON CONFLICT (id) DO NOTHING;

INSERT INTO daily_missions (id, title, description, exam, skill, task_type, target_count, xp_reward) VALUES
    ('mis-speaking-2', 'Two speaking tasks', 'Record two speaking responses.',        NULL, 'speaking',  '', 2, 100),
    ('mis-writing-1',  'One written response', 'Submit one writing task for evaluation.', NULL, 'writing', '', 1, 150),
    ('mis-reading-3',  'Three reading drills', 'Complete three reading questions.',   NULL, 'reading',   '', 3, 120),
    ('mis-listening-2','Two listening drills', 'Complete two listening questions.',   NULL, 'listening', '', 2, 110)
ON CONFLICT (id) DO NOTHING;
