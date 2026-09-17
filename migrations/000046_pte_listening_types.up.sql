-- Align the legacy listening seed with the authoring catalogue. Keep question
-- IDs, option IDs, answers, timings and tags intact. API filters accept both IDs.
UPDATE questions q
SET type_id = mapping.new_id, type_name = mapping.new_name
FROM (VALUES
    ('pte-lis-006', 'multiple-choice-multiple', 'pte-listening-mcma', 'MCQ Multiple Answer'),
    ('pte-lis-007', 'multiple-choice-multiple', 'pte-listening-mcma', 'MCQ Multiple Answer'),
    ('pte-lis-008', 'fill-in-the-blanks', 'pte-listening-fib', 'Listening Fill in the Blanks'),
    ('pte-lis-009', 'fill-in-the-blanks', 'pte-listening-fib', 'Listening Fill in the Blanks'),
    ('pte-lis-010', 'highlight-correct-summary', 'pte-highlight-correct-summary', 'Highlight Correct Summary'),
    ('pte-lis-011', 'highlight-correct-summary', 'pte-highlight-correct-summary', 'Highlight Correct Summary'),
    ('pte-lis-012', 'select-missing-word', 'pte-select-missing-word', 'Select Missing Word'),
    ('pte-lis-013', 'select-missing-word', 'pte-select-missing-word', 'Select Missing Word')
) AS mapping(question_id, old_id, new_id, new_name)
WHERE q.id = mapping.question_id AND q.type_id = mapping.old_id
  AND q.exam = 'PTE' AND q.skill = 'listening';
