-- Which exams a question is set by, recorded on the question itself.
--
-- Why this exists. A reading question's exam has always been its passage's
-- exam: 000008 seeds it with `SELECT ... p.exam`. That is about to stop being
-- true. A passage is text, and the same text is set by IELTS and by PTE with
-- different tasks on it — so eligibility has to live one level down, on the
-- question, where the task type already lives.
--
-- This migration only adds the column and fills it from what is already known.
-- Nothing reads it yet; 000025 moves selection onto it once passages are
-- exam-neutral.

ALTER TABLE questions
    -- Which exams may deal this question. A question belongs to one exam
    -- normally, and to both only where the task is genuinely the same on both
    -- papers, which in reading means the two multiple-choice types and nothing
    -- else. It is not derived from the passage and never should be.
    ADD COLUMN supported_exams TEXT[] NOT NULL DEFAULT '{}',

    -- Which paragraphs of the passage this question is answered from, by the
    -- labels already stored in reading_passages.paragraphs.
    --
    -- Empty means the whole passage, which is the honest default: most tasks
    -- are answered from anywhere in the text, and claiming otherwise would be
    -- inventing an answer key.
    ADD COLUMN source_paragraphs TEXT[] NOT NULL DEFAULT '{}',

    -- How this question is marked, when the default is not what it needs.
    --
    -- Empty means "use points", which is what every row does today, so this
    -- column changes no score until something is written into it. It exists so
    -- that a scoring rule is data rather than another branch in the grader.
    ADD COLUMN scoring_config JSONB NOT NULL DEFAULT '{}';

UPDATE questions SET supported_exams = ARRAY[exam] WHERE cardinality(supported_exams) = 0;

-- A question no exam sets cannot be dealt to anyone, so it is a content bug
-- rather than a state worth supporting.
ALTER TABLE questions
    ADD CONSTRAINT questions_supported_exams_nonempty
        CHECK (cardinality(supported_exams) > 0);

ALTER TABLE questions
    ADD CONSTRAINT questions_supported_exams_valid
        CHECK (supported_exams <@ ARRAY['PTE', 'IELTS']);

-- Selection asks "questions on this group that this exam sets", which is a
-- containment test.
CREATE INDEX idx_questions_supported_exams ON questions USING GIN (supported_exams);

-- The bank-size query in progress.go reads this rather than questions.exam: it
-- counts what a learner could be given, and a shared question is available to
-- both. Asserted here so the backfill above cannot leave it counting nothing.
DO $$
DECLARE
    orphans INT;
BEGIN
    SELECT count(*) INTO orphans FROM questions WHERE cardinality(supported_exams) = 0;
    IF orphans > 0 THEN
        RAISE EXCEPTION 'question eligibility: % questions have no supported exam', orphans;
    END IF;
END $$;
