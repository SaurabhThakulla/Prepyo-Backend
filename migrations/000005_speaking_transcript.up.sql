-- Speaking evaluations store what the learner was heard to say.
--
-- The transcript is not a nicety: every other field on a speaking evaluation is
-- a judgement about it, and feedback that quotes a sentence the learner cannot
-- see quoted back is not checkable. Writing evaluations leave it empty, since
-- the learner's own text is already on the question.

ALTER TABLE ai_evaluations
    ADD COLUMN transcript TEXT NOT NULL DEFAULT '';
