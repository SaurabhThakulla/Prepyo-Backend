DROP TABLE IF EXISTS writing_mock_sessions;
-- The mock row stays when attempts reference it.
DELETE FROM mocks m WHERE m.id = 'mock-ielts-writing-gen'
   AND NOT EXISTS (SELECT 1 FROM mock_attempts a WHERE a.mock_id = m.id);
