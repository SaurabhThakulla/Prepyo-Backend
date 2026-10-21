-- 000110_pte_exam_memories.down.sql
BEGIN;

DELETE FROM questions WHERE id LIKE 'pte-em-%';
DELETE FROM reading_reorder_items WHERE id LIKE 'ri-pte-em-%';
DELETE FROM reading_question_groups WHERE id LIKE 'g-pte-em-%';
DELETE FROM reading_passages WHERE id LIKE 'rp-pte-em-%';

COMMIT;
