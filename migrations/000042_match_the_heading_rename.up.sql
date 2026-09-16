-- Rename "Arrange the Passage" to "Match the Heading"
--
-- type_id is preserved ('reading-arrange-passage') to maintain backwards compatibility
-- with test sessions, graders, and stored answer records.

UPDATE reading_question_groups
   SET type_name = 'Match the Heading'
 WHERE type_id = 'reading-arrange-passage'
    OR type_name = 'Arrange the Passage';

UPDATE questions
   SET type_name = 'Match the Heading',
       tags = array_replace(tags, 'Arrange the Passage', 'Match the Heading')
 WHERE type_id = 'reading-arrange-passage'
    OR type_name = 'Arrange the Passage';