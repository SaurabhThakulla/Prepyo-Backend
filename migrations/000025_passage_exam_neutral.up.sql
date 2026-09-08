-- A passage stops belonging to an exam.
--
-- Why this exists. reading_passages.exam has been NOT NULL since 000007, and
-- every selection query filters on it: TaskTypes, PickPracticeGroup and
-- PickMockPassages all ask "passages for this exam". That is the one thing
-- standing between the current bank and a shared one — a text about the coming
-- of standard time is not a PTE text, it is a text, and IELTS could set
-- True/False on it tomorrow.
--
-- Eligibility moved to the question in 000024. This migration takes it off the
-- passage and moves the queries onto the questions underneath.
--
-- The column is made nullable rather than dropped. Dropping it is 000028, after
-- a release has confirmed nothing reads it; until then a NULL means "neutral"
-- and the seeded values stay readable for anyone debugging the transition.

ALTER TABLE reading_passages ALTER COLUMN exam DROP NOT NULL;

-- Topic, source, reading level, whatever an author needs to record about a text
-- that is not one of the columns. Passages are the reusable half of this
-- architecture and will outlive the fields guessed at today.
ALTER TABLE reading_passages ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}';

-- ---------------------------------------------------------------------------
-- Exposure becomes exam-scoped
-- ---------------------------------------------------------------------------

-- One row per learner per passage per exam per context.
--
-- Without the exam in the key, sharing a passage would spend it twice over: a
-- learner who sat "The Coming of Standard Time" in an IELTS mock would find it
-- already used up when they came to sit a PTE one, having never seen a single
-- PTE question on it. The two exams draw from the same bank but they do not
-- share a learner's history of it.

ALTER TABLE user_passage_exposures ADD COLUMN exam TEXT;

-- Every existing row was written under the passage's own exam, which is still
-- readable because the column above is nullable rather than gone.
UPDATE user_passage_exposures e
   SET exam = p.exam
  FROM reading_passages p
 WHERE p.id = e.passage_id;

-- Nothing should be left, but an exposure with no exam cannot be keyed and
-- would be silently dropped by the primary key below.
DO $$
DECLARE
    orphans INT;
BEGIN
    SELECT count(*) INTO orphans FROM user_passage_exposures WHERE exam IS NULL;
    IF orphans > 0 THEN
        RAISE EXCEPTION 'exposure backfill: % rows have no exam', orphans;
    END IF;
END $$;

ALTER TABLE user_passage_exposures
    ALTER COLUMN exam SET NOT NULL,
    ADD CONSTRAINT user_passage_exposures_exam_valid CHECK (exam IN ('PTE', 'IELTS'));

ALTER TABLE user_passage_exposures DROP CONSTRAINT user_passage_exposures_pkey;
ALTER TABLE user_passage_exposures
    ADD CONSTRAINT user_passage_exposures_pkey PRIMARY KEY (user_id, passage_id, exam, context);

DROP INDEX IF EXISTS idx_passage_exposures_lookup;
CREATE INDEX idx_passage_exposures_lookup
    ON user_passage_exposures(user_id, exam, context, last_seen_at);

-- Re-order items keep their own exam. They are not shared content: an item is a
-- set of boxes belonging to one task on one paper, not a text two exams can
-- both set questions on, so there is nothing here to decouple.
