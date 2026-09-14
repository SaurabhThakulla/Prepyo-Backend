-- 000039: Revert listening content expansion

DELETE FROM questions WHERE id IN (
    'ielts-lis-002', 'ielts-lis-003', 'ielts-lis-004', 'ielts-lis-005', 'ielts-lis-006', 'ielts-lis-007',
    'pte-lis-002', 'pte-lis-003', 'pte-lis-004', 'pte-lis-005', 'pte-lis-006', 'pte-lis-007',
    'pte-lis-008', 'pte-lis-009', 'pte-lis-010', 'pte-lis-011', 'pte-lis-012', 'pte-lis-013'
);

UPDATE questions
SET type_name = 'Form Completion',
    blanks = NULL
WHERE id = 'ielts-lis-001';

UPDATE mock_sections
SET question_ids = ARRAY['ielts-lis-001']
WHERE id IN ('sec-ielts-full-4', 'sec-ielts-diag-4');

UPDATE mock_sections
SET question_ids = ARRAY['pte-lis-001']
WHERE id IN ('sec-pte-full-3', 'sec-pte-diag-4');
