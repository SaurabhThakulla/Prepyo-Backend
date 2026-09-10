DELETE FROM practice_attempts WHERE question_id IN (
    SELECT id FROM questions WHERE passage_id IN ('rp-container-01', 'rp-icecore-01', 'rp-seedvault-01'));
DELETE FROM mistakes WHERE question_id IN (
    SELECT id FROM questions WHERE passage_id IN ('rp-container-01', 'rp-icecore-01', 'rp-seedvault-01'));
DELETE FROM ai_evaluations WHERE question_id IN (
    SELECT id FROM questions WHERE passage_id IN ('rp-container-01', 'rp-icecore-01', 'rp-seedvault-01'));
DELETE FROM reading_passages WHERE id IN ('rp-container-01', 'rp-icecore-01', 'rp-seedvault-01');
