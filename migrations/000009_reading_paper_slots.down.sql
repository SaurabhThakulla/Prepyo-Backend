ALTER TABLE reading_passages
    DROP CONSTRAINT IF EXISTS reading_passages_word_count_floor;

-- The stored counts go back to what 000008 seeded, wrong as they were, so this
-- migration is a clean round trip.
UPDATE reading_passages SET word_count = 980 WHERE id = 'rp-choc-01';
UPDATE reading_passages SET word_count = 900 WHERE id = 'rp-bees-01';
UPDATE reading_passages SET word_count = 940 WHERE id = 'rp-paper-01';

-- Remove the sentence appended to rp-paper-01 paragraph C.
UPDATE reading_passages
SET paragraphs = (
        SELECT jsonb_agg(
                   CASE WHEN elem->>'label' = 'C'
                        THEN jsonb_set(elem, '{text}',
                                 to_jsonb(replace(elem->>'text',
                                     ' The craft moved north from there along the trade routes, reaching France in the fourteenth century and the German and Dutch towns in the fifteenth, always settling where clean running water and a supply of rags happened to meet.',
                                     '')))
                        ELSE elem
                   END
                   ORDER BY ord)
        FROM jsonb_array_elements(paragraphs) WITH ORDINALITY AS t(elem, ord)
    )
WHERE id = 'rp-paper-01';

DROP INDEX IF EXISTS idx_reading_groups_slot;

ALTER TABLE reading_question_groups
    DROP COLUMN IF EXISTS paper_slot;
