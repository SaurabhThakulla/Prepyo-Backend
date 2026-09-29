-- Removes the starter EPS-TOPIK questions and their pictures, with anything a
-- learner did on them.

DELETE FROM mistakes WHERE question_id LIKE 'eps-seed-%';
DELETE FROM practice_attempts WHERE question_id LIKE 'eps-seed-%';
DELETE FROM questions WHERE id LIKE 'eps-seed-%';
DELETE FROM question_assets WHERE id LIKE 'eps_seed_%';
