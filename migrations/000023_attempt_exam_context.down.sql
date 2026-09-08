DROP INDEX IF EXISTS idx_mistakes_user_exam;

ALTER TABLE mistakes
    DROP CONSTRAINT IF EXISTS mistakes_exam_valid,
    DROP COLUMN IF EXISTS exam;

DROP INDEX IF EXISTS idx_practice_attempts_user_exam;

ALTER TABLE practice_attempts
    DROP CONSTRAINT IF EXISTS practice_attempts_exam_valid,
    DROP COLUMN IF EXISTS exam;
