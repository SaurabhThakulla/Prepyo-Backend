DROP TABLE IF EXISTS reading_mock_blueprint_slots;
DROP TABLE IF EXISTS reading_mock_blueprints;

-- The PTE paper's own sessions go with its blueprint: without one there is
-- nothing to compose them from, and mock_attempts would reference a mock nobody
-- can sit.
DELETE FROM reading_mock_sessions WHERE mock_id = 'mock-pte-reading-gen';
DELETE FROM mocks WHERE id = 'mock-pte-reading-gen';
