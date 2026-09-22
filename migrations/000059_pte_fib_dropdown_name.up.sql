-- Rename PTE "Reading & Writing: Fill in the Blanks" to Pearson's current name.
-- Only the label changes; type_id (fill-in-blanks-rw) is what grading and the
-- practice menu key off, so nothing else moves.
UPDATE reading_question_groups SET type_name = 'Fill in the Blanks (Dropdown)' WHERE type_id = 'fill-in-blanks-rw';
UPDATE questions SET type_name = 'Fill in the Blanks (Dropdown)' WHERE type_id = 'fill-in-blanks-rw';
