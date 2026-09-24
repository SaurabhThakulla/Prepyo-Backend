DROP TABLE IF EXISTS pte_mock_items;
DROP TABLE IF EXISTS pte_mock_sessions;

DELETE FROM mock_attempts
 WHERE mock_id IN ('mock-pte-full', 'mock-pte-speaking', 'mock-pte-writing', 'mock-pte-reading', 'mock-pte-listening');
DELETE FROM mocks
 WHERE id IN ('mock-pte-full', 'mock-pte-speaking', 'mock-pte-writing', 'mock-pte-reading', 'mock-pte-listening');
