-- 000109_asq_answer_keys.down.sql
-- Clears the Answer Short Question bank 3 keys again, back to empty as before.

UPDATE questions
SET correct_answers = NULL
WHERE id LIKE 'pte-b3-asq-0%'
  AND type_id IN ('answer-short-question', 'pte-answer-short-question');
