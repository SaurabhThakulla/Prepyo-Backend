DROP INDEX IF EXISTS idx_passage_exposures_lookup;

ALTER TABLE user_passage_exposures DROP CONSTRAINT IF EXISTS user_passage_exposures_pkey;
ALTER TABLE user_passage_exposures
    ADD CONSTRAINT user_passage_exposures_pkey PRIMARY KEY (user_id, passage_id, context);

CREATE INDEX idx_passage_exposures_lookup
    ON user_passage_exposures(user_id, context, last_seen_at);

ALTER TABLE user_passage_exposures
    DROP CONSTRAINT IF EXISTS user_passage_exposures_exam_valid,
    DROP COLUMN IF EXISTS exam;

ALTER TABLE reading_passages DROP COLUMN IF EXISTS metadata;

-- A passage authored while exam was optional has no value to restore, so the
-- column cannot simply be made NOT NULL again. Anything neutral goes back to
-- IELTS, which is what the three original passages were.
UPDATE reading_passages SET exam = 'IELTS' WHERE exam IS NULL;
ALTER TABLE reading_passages ALTER COLUMN exam SET NOT NULL;
