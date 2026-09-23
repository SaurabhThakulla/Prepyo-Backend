-- Restores the previous blueprint shape. Question positions and shuffle flags
-- are left as they are: text order is a correction, not a preference, and
-- undoing it would put the sets back out of order.
DELETE FROM reading_mock_blueprint_slots WHERE blueprint_id = 'bp-ielts-reading';

INSERT INTO reading_mock_blueprint_slots
    (blueprint_id, position, ordinal, type_id, question_count) VALUES
    ('bp-ielts-reading', 1, 1, 'reading-sentence-completion',  7),
    ('bp-ielts-reading', 1, 2, 'reading-true-false',           6),
    ('bp-ielts-reading', 2, 1, 'reading-find-the-paragraph',   6),
    ('bp-ielts-reading', 2, 2, 'reading-sentence-completion',  4),
    ('bp-ielts-reading', 2, 3, 'reading-yes-no-not-given',     3),
    ('bp-ielts-reading', 3, 1, 'reading-sentence-completion',  6),
    ('bp-ielts-reading', 3, 2, 'reading-matching-information', 4),
    ('bp-ielts-reading', 3, 3, 'reading-yes-no-not-given',     4);

ALTER TABLE reading_question_groups ALTER COLUMN shuffle_questions SET DEFAULT TRUE;
