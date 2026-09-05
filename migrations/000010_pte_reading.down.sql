DROP TABLE IF EXISTS user_reorder_exposures;

DROP INDEX IF EXISTS idx_questions_reorder_item;

ALTER TABLE questions
    DROP CONSTRAINT IF EXISTS questions_reorder_excludes_passage;

-- A re-order question only makes sense next to its item, so it goes with it.
-- Standalone questions have no reorder_item_id and are untouched.
DELETE FROM questions WHERE reorder_item_id IS NOT NULL;

ALTER TABLE questions
    DROP COLUMN IF EXISTS reorder_item_id;

DROP TABLE IF EXISTS reading_reorder_items;

ALTER TABLE reading_question_groups
    DROP COLUMN IF EXISTS passage_display;
