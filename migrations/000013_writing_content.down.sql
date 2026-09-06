DELETE FROM questions
WHERE id LIKE 'ielts-wrt-op-%'
   OR id LIKE 'pte-wrt-es-%'
   OR id LIKE 'ielts-wrt-fg-%';

ALTER TABLE questions DROP COLUMN IF EXISTS figure_data;
