DELETE FROM mocks WHERE is_generated;

ALTER TABLE mocks DROP COLUMN IF EXISTS is_generated;

DROP TABLE IF EXISTS reading_mock_sessions;
DROP TABLE IF EXISTS user_passage_exposures;

ALTER TABLE questions
    DROP CONSTRAINT IF EXISTS questions_group_implies_passage;

-- The two indexes go with the columns they cover.
DROP INDEX IF EXISTS idx_questions_group;
DROP INDEX IF EXISTS idx_questions_passage;

-- Reading questions are rows in `questions` that only make sense next to their
-- passage, so they go with it. Standalone questions have no passage_id and are
-- untouched.
DELETE FROM questions WHERE passage_id IS NOT NULL;

ALTER TABLE questions
    DROP COLUMN IF EXISTS passage_id,
    DROP COLUMN IF EXISTS group_id,
    DROP COLUMN IF EXISTS group_position;

DROP TABLE IF EXISTS reading_question_groups;
DROP TABLE IF EXISTS reading_passages;
