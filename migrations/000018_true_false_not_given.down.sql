UPDATE questions SET options = '[
    {"id":"TRUE","text":"True"},
    {"id":"FALSE","text":"False"}]'::jsonb
WHERE type_id = 'reading-true-false' AND passage_id IS NOT NULL;

UPDATE reading_question_groups
   SET type_name    = 'True / False',
       instructions = 'Do the following statements agree with the information in the passage? Answer True or False.'
 WHERE type_id = 'reading-true-false';
