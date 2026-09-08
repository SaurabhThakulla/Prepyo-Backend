UPDATE reading_question_groups
   SET type_name = 'Sentence Completion'
 WHERE type_id = 'reading-sentence-completion';

UPDATE questions
   SET type_name = 'Sentence Completion',
       title     = 'Sentence Completion ' || group_position::text,
       tags      = ARRAY['IELTS Reading', 'Sentence Completion']
 WHERE type_id = 'reading-sentence-completion' AND passage_id IS NOT NULL;
