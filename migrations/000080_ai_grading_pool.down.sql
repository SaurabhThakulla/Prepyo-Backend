DROP TABLE IF EXISTS ai_grading_usage;

UPDATE plans SET mock_tests_included = 10 WHERE id = 'elite';

UPDATE plans SET features = ARRAY['Core practice tasks', '5 practice sub-tests per day', '1 full mock exam']
 WHERE id = 'free';
UPDATE plans SET features = ARRAY['7 days access', '40 practice sub-tests per day', '10 section mock tests / week', '2 full mock exams']
 WHERE id = 'weekly';
UPDATE plans SET features = ARRAY['33 days access', '50 practice sub-tests per day', '20 section mock tests / month', '5 full mock exams']
 WHERE id = 'pro';
UPDATE plans SET features = ARRAY['33 days full access', 'Unlimited practice sub-tests', 'Unlimited full mock exams', 'Unlimited AI tutor', 'Priority evaluation queue']
 WHERE id = 'elite';

ALTER TABLE plans DROP COLUMN IF EXISTS ai_gradings_per_period;
