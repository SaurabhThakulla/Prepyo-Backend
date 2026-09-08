-- A passage authored after the column was dropped has no exam to restore, and
-- there is no honest way to invent one: it may carry questions for both. Every
-- passage comes back as IELTS, which is what the three original ones were, and
-- the column goes back to being nullable rather than NOT NULL so the restore
-- cannot fail on content written since.
ALTER TABLE reading_passages ADD COLUMN exam TEXT CHECK (exam IN ('PTE', 'IELTS'));
UPDATE reading_passages SET exam = 'IELTS';

ALTER TABLE reading_question_groups
    ADD COLUMN paper_slot SMALLINT NOT NULL DEFAULT 0 CHECK (paper_slot BETWEEN 0 AND 3);

-- The seeded groups ran in slot order: positions 1-2 were the first section,
-- 3-5 the second, 6-8 the third. Groups authored since carry 0, which is what
-- "not part of any paper section" has always meant.
UPDATE reading_question_groups
   SET paper_slot = CASE
        WHEN position BETWEEN 1 AND 2 THEN 1
        WHEN position BETWEEN 3 AND 5 THEN 2
        WHEN position BETWEEN 6 AND 8 THEN 3
        ELSE 0
    END;

CREATE INDEX idx_reading_groups_slot
    ON reading_question_groups(passage_id, paper_slot) WHERE paper_slot > 0;
