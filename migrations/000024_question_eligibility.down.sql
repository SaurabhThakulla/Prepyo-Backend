DROP INDEX IF EXISTS idx_questions_supported_exams;

ALTER TABLE questions
    DROP CONSTRAINT IF EXISTS questions_supported_exams_valid,
    DROP CONSTRAINT IF EXISTS questions_supported_exams_nonempty,
    DROP COLUMN IF EXISTS scoring_config,
    DROP COLUMN IF EXISTS source_paragraphs,
    DROP COLUMN IF EXISTS supported_exams;
