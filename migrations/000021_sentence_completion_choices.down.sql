UPDATE questions SET options = NULL
 WHERE type_id = 'reading-sentence-completion' AND passage_id IS NOT NULL;

UPDATE reading_question_groups
   SET instructions = 'Complete the sentences below. Write ONE WORD ONLY from the passage in each gap.'
 WHERE type_id = 'reading-sentence-completion' AND instructions LIKE 'Complete the sentences%';

UPDATE reading_question_groups
   SET instructions = 'Complete the summary below. Write ONE WORD OR NUMBER from the passage in each gap.'
 WHERE type_id = 'reading-sentence-completion' AND instructions LIKE 'Complete the summary%';
