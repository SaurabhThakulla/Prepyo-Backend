DROP TABLE IF EXISTS full_mock_sessions;

-- Full-mock papers that are still open would break the one-open-paper rule.
UPDATE listening_mock_sessions SET status = 'abandoned' WHERE in_full_mock AND status = 'in_progress';
UPDATE reading_mock_sessions   SET status = 'abandoned' WHERE in_full_mock AND status = 'in_progress';
UPDATE writing_mock_sessions   SET status = 'abandoned' WHERE in_full_mock AND status = 'in_progress';
UPDATE speaking_mock_sessions  SET status = 'abandoned' WHERE in_full_mock AND status = 'in_progress';

DROP INDEX IF EXISTS idx_listening_mock_live;
DROP INDEX IF EXISTS idx_reading_mock_sessions_live;
DROP INDEX IF EXISTS idx_writing_mock_live;
DROP INDEX IF EXISTS idx_speaking_mock_live;
ALTER TABLE listening_mock_sessions DROP COLUMN IF EXISTS in_full_mock;
ALTER TABLE reading_mock_sessions   DROP COLUMN IF EXISTS in_full_mock;
ALTER TABLE writing_mock_sessions   DROP COLUMN IF EXISTS in_full_mock;
ALTER TABLE speaking_mock_sessions  DROP COLUMN IF EXISTS in_full_mock;
CREATE UNIQUE INDEX idx_listening_mock_live ON listening_mock_sessions(user_id) WHERE status = 'in_progress';
CREATE UNIQUE INDEX idx_reading_mock_sessions_live ON reading_mock_sessions(user_id, exam) WHERE status = 'in_progress';
CREATE UNIQUE INDEX idx_writing_mock_live ON writing_mock_sessions(user_id) WHERE status = 'in_progress';
CREATE UNIQUE INDEX idx_speaking_mock_live ON speaking_mock_sessions(user_id) WHERE status = 'in_progress';

DELETE FROM mock_attempts WHERE mock_id = 'mock-ielts-full';
DELETE FROM mocks WHERE id = 'mock-ielts-full';
