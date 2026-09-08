-- What a reading paper is made of, as data rather than as a Go constant.
--
-- Why this exists. Composition lived in two places, and both of them said
-- IELTS. reading.PaperSlots hard-coded three sections of 13/13/14 questions over
-- six task types, and reading_question_groups.paper_slot stamped each authored
-- group with the section it belonged to. Between them they made three
-- assumptions that a shared bank cannot keep:
--
--   1. Every paper has three passages. PTE's does not.
--   2. Every passage carries all three sections, which is what
--      PickMockPassages asserts with COUNT(DISTINCT paper_slot) = 3. It means a
--      passage cannot join the bank until it has all eight task sets on it.
--   3. There is one paper shape, so PTE gets no reading mock at all.
--
-- Sections become rows, a passage qualifies for a section by carrying enough
-- questions of the tasks that section asks for, and a second exam is a second
-- blueprint rather than a second code path.

CREATE TABLE reading_mock_blueprints (
    id               TEXT PRIMARY KEY,

    -- The mocks row this paper reports as. mock_attempts.mock_id references
    -- mocks(id), so a generated paper still needs one to be gradeable; the
    -- blueprint is what says how it is composed.
    mock_id          TEXT NOT NULL REFERENCES mocks(id),

    exam             TEXT NOT NULL CHECK (exam IN ('PTE', 'IELTS')),
    passage_count    INT  NOT NULL CHECK (passage_count > 0),
    total_questions  INT  NOT NULL CHECK (total_questions > 0),
    duration_minutes INT  NOT NULL CHECK (duration_minutes > 0),
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One live blueprint per exam. A second would make "start a reading mock"
-- ambiguous, and an inactive one is how a shape is retired without invalidating
-- the papers composed under it.
CREATE UNIQUE INDEX idx_reading_blueprint_active
    ON reading_mock_blueprints(exam) WHERE is_active;

-- One row per task within a section, not one row per section.
--
-- A section is not "13 questions of these types" — a real paper says Questions
-- 1-7 are sentence completion and 8-13 are True/False/Not Given. Storing the
-- count per task is what lets a section be reproduced exactly instead of filled
-- greedily from whichever task set the passage happens to list first.
CREATE TABLE reading_mock_blueprint_slots (
    blueprint_id   TEXT NOT NULL REFERENCES reading_mock_blueprints(id) ON DELETE CASCADE,

    -- The section. Every row sharing a position is one section, and one section
    -- is filled from one passage.
    position       INT  NOT NULL CHECK (position > 0),

    -- The task's place within its section, which is the order a learner meets
    -- them in.
    ordinal        INT  NOT NULL CHECK (ordinal > 0),

    type_id        TEXT NOT NULL,
    question_count INT  NOT NULL CHECK (question_count > 0),

    -- Where the section's content comes from. 'passage' takes a passage and the
    -- matching task sets on it. 'reorder' takes standalone Re-order Paragraphs
    -- items, which have no passage to sit on — PTE sets them and IELTS does
    -- not, which is exactly the kind of difference a shared bank has to be able
    -- to express.
    source         TEXT NOT NULL DEFAULT 'passage'
                       CHECK (source IN ('passage', 'reorder')),

    PRIMARY KEY (blueprint_id, position, ordinal)
);

-- ---------------------------------------------------------------------------
-- IELTS: the paper that is being composed today
-- ---------------------------------------------------------------------------

-- Reproduced rather than redesigned. The three sections carry the same task
-- types in the same counts that PaperSlots has been producing — 7+6, 6+4+3,
-- 6+4+4 — so an IELTS learner sits the same 40-question paper after this
-- migration as before it.

INSERT INTO reading_mock_blueprints
    (id, mock_id, exam, passage_count, total_questions, duration_minutes) VALUES
    ('bp-ielts-reading', 'mock-ielts-reading-gen', 'IELTS', 3, 40, 60);

INSERT INTO reading_mock_blueprint_slots
    (blueprint_id, position, ordinal, type_id, question_count) VALUES
    ('bp-ielts-reading', 1, 1, 'reading-sentence-completion',  7),
    ('bp-ielts-reading', 1, 2, 'reading-true-false',           6),

    ('bp-ielts-reading', 2, 1, 'reading-find-the-paragraph',   6),
    ('bp-ielts-reading', 2, 2, 'reading-arrange-passage',      4),
    ('bp-ielts-reading', 2, 3, 'reading-yes-no-not-given',     3),

    ('bp-ielts-reading', 3, 1, 'reading-sentence-completion',  6),
    ('bp-ielts-reading', 3, 2, 'reading-matching-information', 4),
    ('bp-ielts-reading', 3, 3, 'reading-yes-no-not-given',     4);

-- ---------------------------------------------------------------------------
-- PTE: its own shape, over the same bank
-- ---------------------------------------------------------------------------

-- The current PTE Academic reading section is 13-18 items in about half an hour,
-- across five task types. This is thirteen: the low end, which is what the bank
-- can honestly fill. The counts are data and can be raised without a code change
-- once there is more content.
--
-- Each section draws its own passage, which is closer to the real paper than the
-- IELTS arrangement is: PTE reading items are independent texts rather than
-- several tasks hung off one long passage.
--
-- Re-order Paragraphs is its own section because its items are not passage
-- tasks at all.

INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated) VALUES
    ('mock-pte-reading-gen', 'pte-2026-01', 'PTE', 'PTE Academic Reading Mock',
     'A full reading section, composed fresh each time from texts you have not sat.',
     30, FALSE, TRUE);

INSERT INTO reading_mock_blueprints
    (id, mock_id, exam, passage_count, total_questions, duration_minutes) VALUES
    ('bp-pte-reading', 'mock-pte-reading-gen', 'PTE', 4, 13, 30);

INSERT INTO reading_mock_blueprint_slots
    (blueprint_id, position, ordinal, type_id, question_count, source) VALUES
    ('bp-pte-reading', 1, 1, 'fill-in-blanks-rw',    2, 'passage'),
    ('bp-pte-reading', 2, 1, 'fill-in-blanks-r',     2, 'passage'),
    ('bp-pte-reading', 3, 1, 'reading-mcq-single',   4, 'passage'),
    ('bp-pte-reading', 4, 1, 'reading-mcq-multiple', 2, 'passage'),
    ('bp-pte-reading', 5, 1, 'reorder-paragraphs',   3, 'reorder');

-- ---------------------------------------------------------------------------
-- Papers already in flight
-- ---------------------------------------------------------------------------

-- Left alone, deliberately.
--
-- 000009 had to abandon every in-flight paper when it changed composition,
-- because a stored paper was rebuilt by re-running composition and filtering the
-- result: when the sections moved, two thirds of what a learner was working on
-- stopped appearing. This migration removes that dependency instead. Hydration
-- now works from the stored question ids alone, so a paper dealt under the old
-- composition rebuilds exactly as it was dealt.
