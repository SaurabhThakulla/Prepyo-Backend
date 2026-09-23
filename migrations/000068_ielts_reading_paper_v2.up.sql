-- IELTS Academic Reading paper, version 2.
--
-- Why this exists. The generated IELTS paper (bp-ielts-reading, 000026 and
-- 000051) asked one passage for 7 sentence-completion and 6 True/False/Not
-- Given items in section 1, and for 6 `reading-find-the-paragraph` items in
-- section 2. No IELTS passage in the bank carries that shape: the original
-- passages (000054, 000067) carry 5 items of each task, and none carries
-- find-the-paragraph at all. Composition failed closed for every learner.
--
-- The new shape is still 40 one-mark questions over three passages in 60
-- minutes (13 + 13 + 14), and every original passage can fill any section, so
-- composition always has at least as many candidates as there are passages.
-- It also covers more of the official task families than before: multiple
-- choice returns, alongside TFNG, YNNG, matching information and sentence
-- completion. "Choose TWO" items stay out of the paper because they are two
-- marks on one row, and the paper is counted in one-mark rows.
--
--   Section 1 (Q1-13):  TFNG 5, sentence completion 5, multiple choice 3
--   Section 2 (Q14-26): matching information 5, YNNG 5, multiple choice 3
--   Section 3 (Q27-40): matching information 4, sentence completion 5, YNNG 5
--
-- Papers already dealt are untouched: they rebuild from their stored ids.

DELETE FROM reading_mock_blueprint_slots WHERE blueprint_id = 'bp-ielts-reading';

INSERT INTO reading_mock_blueprint_slots
    (blueprint_id, position, ordinal, type_id, question_count) VALUES
    ('bp-ielts-reading', 1, 1, 'reading-true-false',           5),
    ('bp-ielts-reading', 1, 2, 'reading-sentence-completion',  5),
    ('bp-ielts-reading', 1, 3, 'reading-mcq-single',           3),

    ('bp-ielts-reading', 2, 1, 'reading-matching-information', 5),
    ('bp-ielts-reading', 2, 2, 'reading-yes-no-not-given',     5),
    ('bp-ielts-reading', 2, 3, 'reading-mcq-single',           3),

    ('bp-ielts-reading', 3, 1, 'reading-matching-information', 4),
    ('bp-ielts-reading', 3, 2, 'reading-sentence-completion',  5),
    ('bp-ielts-reading', 3, 3, 'reading-yes-no-not-given',     5);

UPDATE reading_mock_blueprints
   SET passage_count = 3, total_questions = 40, duration_minutes = 60
 WHERE id = 'bp-ielts-reading';

-- The sum of the slots must be the paper the blueprint advertises.
DO $$
DECLARE
    slot_total INT;
BEGIN
    SELECT COALESCE(sum(question_count), 0) INTO slot_total
      FROM reading_mock_blueprint_slots WHERE blueprint_id = 'bp-ielts-reading';
    IF slot_total <> 40 THEN
        RAISE EXCEPTION 'bp-ielts-reading slots total %, want 40', slot_total;
    END IF;
END $$;

-- ---------------------------------------------------------------------------
-- Question order
-- ---------------------------------------------------------------------------
--
-- For multiple choice, sentence completion and short-answer questions the
-- official format states that questions follow the order of the information in
-- the text; True/False/Not Given and Yes/No/Not Given follow it in official and
-- published practice papers too. These original sets had items out of text
-- order. Positions are reordered by the paragraph each explanation cites; a
-- NOT GIVEN item has no location and keeps its place. Only positions change,
-- never ids, so attempts and mistakes still point at the same questions.

UPDATE questions q SET group_position = v.pos
FROM (VALUES
    ('q-ir67-desal-multiple-2', 1),
    ('q-ir67-desal-multiple-1', 2),
    ('q-ir67-desal-tfng-5', 2),
    ('q-ir67-desal-tfng-2', 3),
    ('q-ir67-desal-tfng-3', 5),
    ('q-ir67-lighthouse-multiple-2', 1),
    ('q-ir67-lighthouse-multiple-1', 2),
    ('q-ir67-mangrove-completion-4', 2),
    ('q-ir67-mangrove-completion-2', 3),
    ('q-ir67-mangrove-completion-3', 4),
    ('q-ir67-mangrove-tfng-3', 2),
    ('q-ir67-mangrove-tfng-2', 3),
    ('q-ir67-mangrove-ynng-5', 1),
    ('q-ir67-mangrove-ynng-1', 2),
    ('q-ir67-mangrove-ynng-2', 3),
    ('q-ir67-mangrove-ynng-3', 5),
    ('q-ir67-oceannoise-completion-5', 4),
    ('q-ir67-oceannoise-completion-4', 5),
    ('q-ir67-oceannoise-ynng-3', 1),
    ('q-ir67-oceannoise-ynng-1', 2),
    ('q-ir67-oceannoise-ynng-2', 3),
    ('q-ir67-queues-ynng-5', 1),
    ('q-ir67-queues-ynng-1', 2),
    ('q-ir67-queues-ynng-2', 3),
    ('q-ir67-queues-ynng-3', 5),
    ('q-ir67-sleep-tfng-5', 4),
    ('q-ir67-sleep-tfng-4', 5),
    ('q-ir67-vertical-tfng-2', 1),
    ('q-ir67-vertical-tfng-1', 2)
) AS v(id, pos)
WHERE q.id = v.id;

-- Sets whose order carries meaning must never be dealt shuffled. The column
-- default also becomes FALSE, so a newly authored set keeps its order unless
-- an author chooses otherwise. The service enforces the same rule by type.
UPDATE reading_question_groups
   SET shuffle_questions = FALSE
 WHERE shuffle_questions
   AND type_id IN ('reading-true-false', 'reading-yes-no-not-given',
                   'reading-sentence-completion', 'reading-summary-completion',
                   'reading-short-answer', 'reading-mcq-single', 'reading-mcq-multiple');

ALTER TABLE reading_question_groups ALTER COLUMN shuffle_questions SET DEFAULT FALSE;
