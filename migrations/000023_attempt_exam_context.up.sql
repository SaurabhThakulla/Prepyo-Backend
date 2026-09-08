-- The exam a learner was working under belongs to the attempt, not the question.
--
-- Why this exists. Until now "which exam was this?" was answered by joining to
-- questions.exam, which works only while every question belongs to exactly one
-- exam. Passages are about to be shared between IELTS and PTE, and a question on
-- a shared passage may be answerable under either — so the join stops being able
-- to answer the question at all.
--
-- Five queries read questions.exam today:
--   progress.go:70   score estimate      -> a.exam
--   progress.go:142  skill breakdown     -> a.exam
--   progress.go:235  bank size           -> stays on the question (see below)
--   mistakes.go:67   filter              -> m.exam
--   mistakes.go:81   projection          -> m.exam
--
-- The bank-size query is deliberately not migrated here. It counts how much
-- content exists for an exam, which is a property of the question and not of any
-- attempt; it moves to supported_exams in 000024.

-- ---------------------------------------------------------------------------
-- Attempts
-- ---------------------------------------------------------------------------

ALTER TABLE practice_attempts ADD COLUMN exam TEXT;

-- question_id is NOT NULL and references questions(id), so every existing row
-- gets a value here. Backfilling from the question is correct for history: every
-- attempt recorded before this migration was made against a question that
-- belonged to exactly one exam.
UPDATE practice_attempts a
   SET exam = q.exam
  FROM questions q
 WHERE q.id = a.question_id;

ALTER TABLE practice_attempts
    ALTER COLUMN exam SET NOT NULL,
    ADD CONSTRAINT practice_attempts_exam_valid CHECK (exam IN ('PTE', 'IELTS'));

-- Progress reads "this learner, this exam, since this date".
CREATE INDEX idx_practice_attempts_user_exam
    ON practice_attempts(user_id, exam, created_at DESC);

-- ---------------------------------------------------------------------------
-- Mistakes
-- ---------------------------------------------------------------------------

ALTER TABLE mistakes ADD COLUMN exam TEXT;

UPDATE mistakes m
   SET exam = q.exam
  FROM questions q
 WHERE q.id = m.question_id;

ALTER TABLE mistakes
    ALTER COLUMN exam SET NOT NULL,
    ADD CONSTRAINT mistakes_exam_valid CHECK (exam IN ('PTE', 'IELTS'));

-- The unique key stays (user_id, question_id): one row per learner per question.
--
-- Widening it to include exam was the alternative, and it is rejected on
-- purpose. A learner who gets the same question wrong under both exams has one
-- weakness, not two, and splitting the row would show it to them twice and let
-- them resolve it once. `exam` therefore records the context of the most recent
-- failure, which is what the mistake bank filters on.

CREATE INDEX idx_mistakes_user_exam ON mistakes(user_id, exam);
