-- Writing content schema: opinion essays for both exams and IELTS Task 1 figures.
--
-- figure_data carries the numbers behind a chart in words. It is never sent to
-- the browser -- the learner gets the image -- but the evaluator needs it to
-- judge whether a description reports the data accurately.

ALTER TABLE questions ADD COLUMN IF NOT EXISTS figure_data TEXT;
