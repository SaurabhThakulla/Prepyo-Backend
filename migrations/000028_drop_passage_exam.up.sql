-- The columns the refactor replaced.
--
-- This is the destructive step, and it runs last on purpose: everything it drops
-- has already been superseded by something that is tested and in use, so this
-- removes the old path rather than switching to a new one.
--
-- What goes:
--
--   reading_passages.exam            eligibility is on the question (000024)
--   reading_question_groups.paper_slot  composition is a blueprint (000026)
--
-- What stays, and why:
--
--   questions.exam                  still the exam a question was authored for,
--                                   and still read by internal/evaluations to
--                                   pick a rubric for writing and speaking.
--                                   supported_exams answers a different
--                                   question — who may be dealt this — and does
--                                   not replace it.
--
--   reading_reorder_items.exam      a re-order item is not shared content. It is
--                                   a set of boxes belonging to one task on one
--                                   paper, not a text two exams can both set
--                                   questions on, so there is nothing here to
--                                   decouple.

DROP INDEX IF EXISTS idx_reading_groups_slot;

ALTER TABLE reading_question_groups DROP COLUMN paper_slot;

-- Last chance to notice that something still depends on the passage's exam. By
-- this point nothing should: practice, the passage index and both mock papers
-- all read supported_exams.
DO $$
DECLARE
    stranded INT;
BEGIN
    SELECT count(*) INTO stranded
      FROM reading_passages p
     WHERE NOT EXISTS (SELECT 1 FROM questions q WHERE q.passage_id = p.id AND q.is_published);

    IF stranded > 0 THEN
        RAISE WARNING 'dropping passage exam: % published passages carry no questions and are now unreachable', stranded;
    END IF;
END $$;

ALTER TABLE reading_passages DROP COLUMN exam;
