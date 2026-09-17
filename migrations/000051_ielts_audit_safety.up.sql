-- Preserve attempts and authored content. Incomplete mocks cannot be sold or started.
ALTER TABLE mocks ADD COLUMN is_available BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE mock_attempts ADD COLUMN answers JSONB NOT NULL DEFAULT '[]'::jsonb;
UPDATE mocks SET is_available=FALSE
WHERE NOT is_generated AND (
 NOT EXISTS (SELECT 1 FROM mock_sections s WHERE s.mock_id=mocks.id)
 OR EXISTS (SELECT 1 FROM mock_sections s WHERE s.mock_id=mocks.id
   AND (cardinality(s.question_ids)=0 OR s.skill IN ('writing','speaking')))
);
-- Paragraph information is not heading matching; historical IDs stay stable.
UPDATE reading_question_groups SET type_name='Matching Information' WHERE type_id='reading-find-the-paragraph';
UPDATE questions SET type_name='Matching Information' WHERE type_id='reading-find-the-paragraph';
UPDATE reading_question_groups SET type_name='Sequence the Summary (legacy practice)' WHERE type_id='reading-arrange-passage';
UPDATE questions SET type_name='Sequence the Summary (legacy practice)' WHERE type_id='reading-arrange-passage';
-- Replace the non-IELTS sequencing slot with four sentence-completion items.
-- Composition still requires three distinct passages and exactly forty items;
-- an insufficient bank fails closed rather than returning a partial paper.
UPDATE reading_mock_blueprint_slots s SET type_id='reading-sentence-completion'
FROM reading_mock_blueprints b
WHERE s.blueprint_id=b.id AND b.exam='IELTS' AND s.type_id='reading-arrange-passage';
