DROP TABLE IF EXISTS listening_mock_sessions;
ALTER TABLE questions DROP COLUMN IF EXISTS listening_group_id;
DROP TABLE IF EXISTS listening_question_groups;
DROP TABLE IF EXISTS listening_parts;
DROP TABLE IF EXISTS listening_tests;
DELETE FROM mocks m WHERE m.id = 'mock-ielts-listening-gen'
   AND NOT EXISTS (SELECT 1 FROM mock_attempts a WHERE a.mock_id = m.id);
