-- Reading paper slots: which section of a generated paper a task set belongs to.
--
-- Why this exists. A generated IELTS reading paper was dealing three passages
-- and taking every group from each: 3 x 40 = 120 questions inside a 60 minute
-- blueprint. A real IELTS Academic Reading paper is 40 questions in three
-- sections, and the sections are not interchangeable — each one carries its own
-- mix of task types:
--
--   Slot 1  Sentence Completion 7, True/False 6                        = 13
--   Slot 2  Find the Writer 6, Arrange the Passage 4, Yes/No/Not Given 3 = 13
--   Slot 3  Sentence Completion 6, Matching Information 4, Y/N/NG 4    = 14
--
-- The slot lives on the group rather than on the passage on purpose. A passage
-- is still authored with all eight groups, so practice can reach any task type
-- on any passage and the bank stays as deep as it is today. What changes is
-- composition: a paper deals three passages, assigns them slots 1, 2 and 3, and
-- takes only that slot's groups from each.

-- ---------------------------------------------------------------------------
-- The slot
-- ---------------------------------------------------------------------------

-- Zero means "not part of any paper section". It is the default, so every
-- existing row keeps working, and it is what task types that do not belong to
-- an IELTS paper carry.
ALTER TABLE reading_question_groups
    ADD COLUMN paper_slot SMALLINT NOT NULL DEFAULT 0
        CHECK (paper_slot BETWEEN 0 AND 3);

-- The seeded groups run in slot order already: positions 1-2 are the first
-- section, 3-5 the second, 6-8 the third. Backfilling by position rather than by
-- id keeps this correct for any passage authored to the same shape.
UPDATE reading_question_groups
SET paper_slot = CASE
        WHEN position BETWEEN 1 AND 2 THEN 1
        WHEN position BETWEEN 3 AND 5 THEN 2
        WHEN position BETWEEN 6 AND 8 THEN 3
        ELSE 0
    END
WHERE passage_id IN (SELECT id FROM reading_passages WHERE exam = 'IELTS');

-- Composing a paper asks for one passage's groups in one slot, which is the
-- only query this column is read by.
CREATE INDEX idx_reading_groups_slot
    ON reading_question_groups(passage_id, paper_slot) WHERE paper_slot > 0;

-- ---------------------------------------------------------------------------
-- Passage length
-- ---------------------------------------------------------------------------

-- The three seeded word counts were authored by hand and none of them is right:
-- the stored numbers are 980, 900 and 940 against real bodies of 1215, 799 and
-- 697 words. That matters because PassageReader prints this number and derives
-- its reading-time estimate from it, so a learner is currently told a passage is
-- shorter or longer than it is.
--
-- Counts are the paragraph body only. `sources` — the five Writer A-E
-- commentaries — is material belonging to the Find the Writer task, not part of
-- the text being read.

-- rp-paper-01 is 697 words, three short of the floor. One sentence is added to
-- paragraph C, which carries no answer: nothing in the sentence completion sets
-- is lifted from it, and none of the Matching Information answers (F, G, H, J)
-- point at it. Appending rather than rewriting means no existing answer word can
-- move or disappear.
UPDATE reading_passages
SET paragraphs = (
        SELECT jsonb_agg(
                   CASE WHEN elem->>'label' = 'C'
                        THEN jsonb_set(elem, '{text}',
                                 to_jsonb(elem->>'text' || ' The craft moved north from there along the trade routes, reaching France in the fourteenth century and the German and Dutch towns in the fifteenth, always settling where clean running water and a supply of rags happened to meet.'))
                        ELSE elem
                   END
                   ORDER BY ord)
        FROM jsonb_array_elements(paragraphs) WITH ORDINALITY AS t(elem, ord)
    )
WHERE id = 'rp-paper-01';

UPDATE reading_passages SET word_count = 1215 WHERE id = 'rp-choc-01';
UPDATE reading_passages SET word_count =  799 WHERE id = 'rp-bees-01';
UPDATE reading_passages SET word_count =  736 WHERE id = 'rp-paper-01';

-- 700 words is the floor, not a range. A passage may run longer — rp-choc-01 is
-- 1215 and stays that way — but one shorter than this does not give a learner
-- enough text to carry three task types. Zero is allowed so a passage can be
-- inserted before its body is counted.
ALTER TABLE reading_passages
    ADD CONSTRAINT reading_passages_word_count_floor
        CHECK (word_count = 0 OR word_count >= 700);

-- ---------------------------------------------------------------------------
-- Papers already in flight
-- ---------------------------------------------------------------------------

-- A live paper stores the question ids it was dealt, and those were dealt under
-- the old composition: every group on three passages, 120 questions. After this
-- migration the same session hydrates against one slot per passage, so two
-- thirds of what the learner was working on would simply stop appearing.
--
-- They are abandoned rather than migrated. There is no honest way to convert a
-- 120 question paper into a 40 question one halfway through — the learner would
-- be graded on a paper they never saw — and abandoning frees them to start a
-- correct one immediately.
UPDATE reading_mock_sessions
   SET status = 'abandoned', submitted_at = now()
 WHERE status = 'in_progress';

-- ---------------------------------------------------------------------------
-- Guard
-- ---------------------------------------------------------------------------

-- A paper is composed from whatever this migration leaves behind, so the shape
-- is asserted here rather than discovered when a learner sits one. Failing the
-- migration is the cheap version of this problem.
DO $$
DECLARE
    bad TEXT;
BEGIN
    SELECT string_agg(format('%s slot %s has %s questions', passage_id, paper_slot, n), '; ')
      INTO bad
      FROM (
          SELECT g.passage_id, g.paper_slot, count(q.id) AS n
            FROM reading_question_groups g
            JOIN reading_passages p ON p.id = g.passage_id AND p.exam = 'IELTS' AND p.is_published
            LEFT JOIN questions q ON q.group_id = g.id AND q.is_published
           WHERE g.paper_slot > 0
           GROUP BY g.passage_id, g.paper_slot
      ) counts
     WHERE (paper_slot = 1 AND n <> 13)
        OR (paper_slot = 2 AND n <> 13)
        OR (paper_slot = 3 AND n <> 14);

    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'reading slots: expected 13/13/14 questions per slot, got %', bad;
    END IF;

    SELECT string_agg(id, ', ')
      INTO bad
      FROM reading_passages p
     WHERE p.exam = 'IELTS' AND p.is_published
       AND (SELECT count(DISTINCT paper_slot)
              FROM reading_question_groups g
             WHERE g.passage_id = p.id AND g.paper_slot > 0) <> 3;

    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'reading slots: these passages do not carry all three slots: %', bad;
    END IF;
END $$;
