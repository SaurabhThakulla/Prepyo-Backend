-- PTE reading: the two structures the IELTS passage model cannot express.
--
-- Most PTE reading tasks fit what is already here — a passage, a group per task
-- type, questions underneath — and four of the five need no new storage at all.
-- Two things do not fit, and they fail for the same reason: showing the passage
-- would show the answer.
--
--   1. Reading & Writing: Fill in the Blanks, and Reading: Fill in the Blanks.
--      The text a learner works on is a gapped rewrite of the passage. Printing
--      the clean original beside it hands over every gap. These need no table —
--      the gapped text already has a home in questions.context_passage, which is
--      how the standalone pte-rdg-001 row has always stored it — only a flag
--      saying the shared passage must not be rendered.
--
--   2. Re-order Paragraphs. Here the content *is* the item: four or five boxes
--      whose correct sequence is the answer. There is no passage to point at,
--      and the exposure rule is different in kind — a learner who has read a
--      text intact must never afterwards be asked to re-order it, which a
--      passage_id could not express. That gets its own table.

-- ---------------------------------------------------------------------------
-- Whether a group renders the passage it hangs from
-- ---------------------------------------------------------------------------

-- 'full' renders the shared passage. 'hidden' renders only the question's own
-- context_passage, which is what the two gap-fills need.
--
-- The default keeps every existing IELTS group reading its passage, and it means
-- a gap-fill group can still hang off the passage it was derived from — useful
-- provenance, without the passage reaching the screen.
ALTER TABLE reading_question_groups
    ADD COLUMN passage_display TEXT NOT NULL DEFAULT 'full'
        CHECK (passage_display IN ('full', 'hidden'));

-- ---------------------------------------------------------------------------
-- Re-order Paragraphs
-- ---------------------------------------------------------------------------

CREATE TABLE reading_reorder_items (
    id               TEXT PRIMARY KEY,
    exam_version_id  TEXT NOT NULL REFERENCES exam_versions(id),
    exam             TEXT NOT NULL CHECK (exam IN ('PTE', 'IELTS')),
    title            TEXT NOT NULL,

    -- The boxes in their correct order, {"label": "A", "text": "..."}, the same
    -- shape as reading_passages.paragraphs.
    --
    -- Stored correct rather than shuffled. The order is the answer key, so it
    -- has to be written down somewhere, and keeping it here means the shuffle
    -- happens per deal — two learners, or the same learner twice, do not get the
    -- boxes in the same arrangement.
    paragraphs       JSONB NOT NULL,

    -- Where this item came from, when it came from anywhere.
    --
    -- Nullable and ON DELETE SET NULL because an item does not need a passage to
    -- exist. What it buys is the rule that makes this table worth having: a
    -- learner who has already read that passage must not be dealt this item,
    -- because they have seen the answer. It also leaves room to derive items in
    -- bulk by shuffling a passage that is already written.
    source_passage_id TEXT REFERENCES reading_passages(id) ON DELETE SET NULL,

    topic            TEXT NOT NULL DEFAULT '',
    word_count       INT NOT NULL DEFAULT 0,
    difficulty       TEXT NOT NULL DEFAULT 'medium'
                         CHECK (difficulty IN ('easy', 'medium', 'hard')),
    tags             TEXT[] NOT NULL DEFAULT '{}',
    is_published     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT reading_reorder_items_paragraphs_array
        CHECK (jsonb_typeof(paragraphs) = 'array'),
    -- Two boxes is not a task and the scorer divides by len-1, so one box would
    -- divide by zero.
    CONSTRAINT reading_reorder_items_enough_boxes
        CHECK (jsonb_array_length(paragraphs) >= 3)
);

CREATE INDEX idx_reading_reorder_exam
    ON reading_reorder_items(exam) WHERE is_published;

-- ---------------------------------------------------------------------------
-- The item still needs a question
-- ---------------------------------------------------------------------------

-- practice_attempts, mistakes and mock grading all key off questions.id, so an
-- item that is not a question cannot be answered, scored or remembered. One row
-- per item: it holds the boxes in `options` and the correct label order in
-- `correct_answers`, which is exactly what scoring.gradeReorder already reads.
ALTER TABLE questions
    ADD COLUMN reorder_item_id TEXT REFERENCES reading_reorder_items(id) ON DELETE CASCADE;

-- An item is not a passage task. Allowing both would give a question two parents
-- with different rules about what may be shown alongside it.
ALTER TABLE questions
    ADD CONSTRAINT questions_reorder_excludes_passage
        CHECK (reorder_item_id IS NULL OR passage_id IS NULL);

CREATE UNIQUE INDEX idx_questions_reorder_item
    ON questions(reorder_item_id) WHERE reorder_item_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- What each learner has already been dealt
-- ---------------------------------------------------------------------------

-- The mirror of user_passage_exposures, for items rather than passages. Same
-- reasoning throughout: context is separate because a mock must never re-deal
-- one, while practice may come back to it once the bank has been through.
CREATE TABLE user_reorder_exposures (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    item_id       TEXT NOT NULL REFERENCES reading_reorder_items(id) ON DELETE CASCADE,
    context       TEXT NOT NULL CHECK (context IN ('practice', 'mock')),
    seen_count    INT NOT NULL DEFAULT 1 CHECK (seen_count > 0),
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, item_id, context)
);

CREATE INDEX idx_reorder_exposures_lookup
    ON user_reorder_exposures(user_id, context, last_seen_at);
