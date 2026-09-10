-- A summary-completion set is printed inside a titled box.
--
-- The real paper heads the box with a title of its own — "Archaeological
-- discoveries", "AI in government, medicine and the law" — which is not the task
-- name and not the instruction line. It is part of the question: the title tells
-- the learner which part of the passage the summary covers, and a set printed
-- without it reads as loose sentences.
ALTER TABLE reading_question_groups
    ADD COLUMN box_title TEXT NOT NULL DEFAULT '';
