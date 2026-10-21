-- Migration: 000112_alfa_pte_practice_questions.down.sql
DELETE FROM questions WHERE id LIKE 'pte-alfa-%';
DELETE FROM reading_question_groups WHERE id LIKE 'g-pte-alfa-%';
DELETE FROM reading_passages WHERE id LIKE 'rp-pte-alfa-%';
DELETE FROM reading_reorder_items WHERE id LIKE 'ri-pte-alfa-%';
