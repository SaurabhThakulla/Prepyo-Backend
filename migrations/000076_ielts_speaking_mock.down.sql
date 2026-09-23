DROP TABLE IF EXISTS speaking_mock_sessions;
DROP TABLE IF EXISTS speaking_mock_sets;
DELETE FROM mocks m WHERE m.id = 'mock-ielts-speaking-gen'
   AND NOT EXISTS (SELECT 1 FROM mock_attempts a WHERE a.mock_id = m.id);
