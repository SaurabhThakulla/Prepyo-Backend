-- 000004: Listening content down migration

UPDATE mocks SET total_duration_minutes = 30  WHERE id = 'mock-ielts-diag-01';
UPDATE mocks SET total_duration_minutes = 160 WHERE id = 'mock-ielts-full-01';

DELETE FROM mock_sections
WHERE id IN ('sec-ielts-diag-3', 'sec-ielts-diag-4', 'sec-ielts-full-4');

DELETE FROM questions WHERE id = 'ielts-lis-001';

UPDATE questions SET audio_transcript = NULL WHERE id = 'pte-lis-001';
