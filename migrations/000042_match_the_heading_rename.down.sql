UPDATE reading_question_groups
   SET type_name = 'Arrange the Passage'
 WHERE type_id = 'reading-arrange-passage'
    OR type_name = 'Match the Heading';

UPDATE questions
   SET type_name = 'Arrange the Passage',
       tags = array_replace(tags, 'Match the Heading', 'Arrange the Passage')
 WHERE type_id = 'reading-arrange-passage'
    OR type_name = 'Match the Heading';