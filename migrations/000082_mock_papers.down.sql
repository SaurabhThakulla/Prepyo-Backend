-- 000082_mock_papers.down.sql

DROP TRIGGER IF EXISTS trg_mock_papers_immutable ON mock_papers;
DROP TRIGGER IF EXISTS trg_mock_papers_no_delete ON mock_papers;
DROP FUNCTION IF EXISTS check_mock_paper_immutability();
DROP FUNCTION IF EXISTS check_mock_paper_no_delete();

DROP INDEX IF EXISTS idx_listening_mock_sessions_user_paper;
ALTER TABLE listening_mock_sessions DROP COLUMN IF EXISTS paper_id;

DROP INDEX IF EXISTS idx_speaking_mock_sessions_user_paper;
ALTER TABLE speaking_mock_sessions DROP COLUMN IF EXISTS paper_id;

DROP INDEX IF EXISTS idx_writing_mock_sessions_user_paper;
ALTER TABLE writing_mock_sessions DROP COLUMN IF EXISTS paper_id;

DROP INDEX IF EXISTS idx_reading_mock_sessions_user_paper;
ALTER TABLE reading_mock_sessions DROP COLUMN IF EXISTS paper_id;

DROP INDEX IF EXISTS idx_pte_mock_sessions_user_paper;
ALTER TABLE pte_mock_sessions DROP COLUMN IF EXISTS paper_id;

DROP TABLE IF EXISTS mock_paper_counters;
DROP TABLE IF EXISTS mock_papers;
