-- Questions cascade from their group, groups from the passage, and the re-order
-- questions from their items, so removing the two parents takes the seed with it.
DELETE FROM reading_passages WHERE id = 'rp-time-01';

DELETE FROM reading_reorder_items
 WHERE id IN ('ri-lock-01', 'ri-penicillin-01', 'ri-expansion-01');
