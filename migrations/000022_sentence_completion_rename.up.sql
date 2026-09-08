-- The task is now answered by choosing a word, so it is named for what it is.
--
-- type_id is left alone. It is the contract with the client, the grader and
-- correct_answers; renaming the label costs nothing, renaming the id would
-- orphan every seeded answer for no gain.

UPDATE reading_question_groups
   SET type_name = 'Fill in the Blanks'
 WHERE type_id = 'reading-sentence-completion';

UPDATE questions
   SET type_name = 'Fill in the Blanks',
       title     = 'Fill in the Blanks ' || group_position::text,
       tags      = ARRAY['IELTS Reading', 'Fill in the Blanks']
 WHERE type_id = 'reading-sentence-completion' AND passage_id IS NOT NULL;
