-- Remove only the 40 questions introduced by 000047. Existing seeds and mocks
-- are not changed. Referenced attempts may prevent rollback through their FKs.
DELETE FROM questions
WHERE exam = 'PTE' AND skill = 'listening' AND id IN (
    'pte-lis-020', 'pte-lis-021', 'pte-lis-022', 'pte-lis-023', 'pte-lis-024',
    'pte-lis-025', 'pte-lis-026', 'pte-lis-027', 'pte-lis-028', 'pte-lis-029',
    'pte-lis-030', 'pte-lis-031', 'pte-lis-032', 'pte-lis-033', 'pte-lis-034',
    'pte-lis-035', 'pte-lis-036', 'pte-lis-037', 'pte-lis-038', 'pte-lis-039',
    'pte-lis-040', 'pte-lis-041', 'pte-lis-042', 'pte-lis-043', 'pte-lis-044',
    'pte-lis-045', 'pte-lis-046', 'pte-lis-047', 'pte-lis-048', 'pte-lis-049',
    'pte-lis-050', 'pte-lis-051', 'pte-lis-052', 'pte-lis-053', 'pte-lis-054',
    'pte-lis-055', 'pte-lis-056', 'pte-lis-057', 'pte-lis-058', 'pte-lis-059'
);
