DELETE FROM reading_mock_blueprint_slots WHERE blueprint_id = 'bp-ielts-gt-reading';
DELETE FROM reading_mock_blueprints WHERE id = 'bp-ielts-gt-reading';
-- The mock row is kept when attempts reference it.
DELETE FROM mocks m WHERE m.id = 'mock-ielts-gt-reading-gen'
   AND NOT EXISTS (SELECT 1 FROM mock_attempts a WHERE a.mock_id = m.id)
   AND NOT EXISTS (SELECT 1 FROM reading_mock_sessions s WHERE s.mock_id = m.id);

ALTER TABLE reading_mock_blueprint_slots DROP COLUMN IF EXISTS passage_tag;
DROP INDEX IF EXISTS idx_reading_blueprint_active;
ALTER TABLE reading_mock_blueprints DROP COLUMN IF EXISTS module;
CREATE UNIQUE INDEX idx_reading_blueprint_active
    ON reading_mock_blueprints(exam) WHERE is_active;
ALTER TABLE reading_passages DROP CONSTRAINT IF EXISTS reading_passages_modules_valid;
ALTER TABLE reading_passages DROP COLUMN IF EXISTS modules;
ALTER TABLE users DROP COLUMN IF EXISTS target_module;
