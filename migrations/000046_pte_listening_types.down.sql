-- Only revert the original seed rows, never newly authored questions.
UPDATE questions q
SET type_id = mapping.old_id, type_name = mapping.old_name
FROM (VALUES
    ('pte-lis-006', 'pte-listening-mcma', 'multiple-choice-multiple', 'Multiple Choice, Multiple Answers'),
    ('pte-lis-007', 'pte-listening-mcma', 'multiple-choice-multiple', 'Multiple Choice, Multiple Answers'),
    ('pte-lis-008', 'pte-listening-fib', 'fill-in-the-blanks', 'Fill in the Blanks'),
    ('pte-lis-009', 'pte-listening-fib', 'fill-in-the-blanks', 'Fill in the Blanks'),
    ('pte-lis-010', 'pte-highlight-correct-summary', 'highlight-correct-summary', 'Highlight Correct Summary'),
    ('pte-lis-011', 'pte-highlight-correct-summary', 'highlight-correct-summary', 'Highlight Correct Summary'),
    ('pte-lis-012', 'pte-select-missing-word', 'select-missing-word', 'Select Missing Word'),
    ('pte-lis-013', 'pte-select-missing-word', 'select-missing-word', 'Select Missing Word')
) AS mapping(question_id, new_id, old_id, old_name)
WHERE q.id = mapping.question_id AND q.type_id = mapping.new_id
  AND q.exam = 'PTE' AND q.skill = 'listening';
