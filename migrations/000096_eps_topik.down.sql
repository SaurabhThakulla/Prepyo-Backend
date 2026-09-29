-- Rows under EPS-TOPIK would break the narrower checks, so they go first:
-- learner history, then content, then the version itself. A learner who chose
-- EPS-TOPIK is moved to PTE, the default target, rather than deleted.

UPDATE users SET target_exam = 'PTE' WHERE target_exam = 'EPS_TOPIK';

DELETE FROM mistakes WHERE exam = 'EPS_TOPIK';
DELETE FROM ai_evaluations WHERE question_id IN (SELECT id FROM questions WHERE exam = 'EPS_TOPIK');
DELETE FROM practice_attempts WHERE exam = 'EPS_TOPIK';
DELETE FROM daily_missions WHERE exam = 'EPS_TOPIK';
DELETE FROM user_passage_exposures WHERE exam = 'EPS_TOPIK';
DELETE FROM mock_attempts WHERE exam_version_id = 'eps-topik-2026-01';

DELETE FROM questions WHERE exam = 'EPS_TOPIK';
UPDATE questions SET supported_exams = array_remove(supported_exams, 'EPS_TOPIK')
WHERE 'EPS_TOPIK' = ANY (supported_exams);

DELETE FROM reading_mock_sessions WHERE exam = 'EPS_TOPIK';
DELETE FROM reading_mock_blueprints WHERE exam = 'EPS_TOPIK';
DELETE FROM reading_reorder_items WHERE exam = 'EPS_TOPIK';
DELETE FROM reading_tests WHERE exam = 'EPS_TOPIK';
DELETE FROM mocks WHERE exam = 'EPS_TOPIK';
DELETE FROM exam_versions WHERE exam = 'EPS_TOPIK';

DO $$
DECLARE
    c record;
BEGIN
    FOR c IN
        SELECT conrelid::regclass AS tbl, conname, pg_get_constraintdef(oid) AS def
        FROM pg_constraint
        WHERE contype = 'c'
          AND connamespace = 'public'::regnamespace
          AND pg_get_constraintdef(oid) LIKE '%ARRAY[''PTE''::text, ''IELTS''::text, ''EPS_TOPIK''::text]%'
    LOOP
        EXECUTE format('ALTER TABLE %s DROP CONSTRAINT %I', c.tbl, c.conname);
        EXECUTE format('ALTER TABLE %s ADD CONSTRAINT %I %s', c.tbl, c.conname,
            replace(c.def,
                'ARRAY[''PTE''::text, ''IELTS''::text, ''EPS_TOPIK''::text]',
                'ARRAY[''PTE''::text, ''IELTS''::text]'));
    END LOOP;
END $$;
