-- True/False becomes True/False/Not Given, which is the task IELTS actually
-- sets. The type_id does not change: correct_answers already hold TRUE and
-- FALSE option ids, and renaming the type would orphan every seeded answer.
-- Adding a third option leaves those answers valid and only widens the choice.

UPDATE questions SET options = '[
    {"id":"TRUE","text":"True"},
    {"id":"FALSE","text":"False"},
    {"id":"NOT_GIVEN","text":"Not Given"}]'::jsonb
WHERE type_id = 'reading-true-false' AND passage_id IS NOT NULL;

UPDATE reading_question_groups
   SET type_name    = 'True / False / Not Given',
       instructions = 'Do the following statements agree with the information in the passage? Answer True, False or Not Given.'
 WHERE type_id = 'reading-true-false';
