-- Reading passages: the content model behind passage-driven reading practice
-- and generated reading mocks.
--
-- Why this exists. Until now a reading task was one row in `questions`, and a
-- set of six True/False statements lived as one row with six strings in
-- correct_answers. That is fine for a single task but it cannot answer either
-- of the two questions this feature is built on: "give me a random passage that
-- has Matching Information questions", and "give this learner three passages
-- they have never seen in a mock". Both need the passage to be a row, and each
-- question to be a row underneath it.
--
-- Shape:
--   reading_passages         one text, split into labelled paragraphs
--   reading_question_groups  one task set on that passage (a type + instructions)
--   questions                one question, now optionally pointing at a group
--
-- Questions stay in `questions` rather than moving to a table of their own.
-- That is deliberate: practice_attempts, mistakes, scoring and mock grading all
-- key off questions.id, and a parallel question table would need a parallel
-- copy of every one of them.

-- ---------------------------------------------------------------------------
-- Passages
-- ---------------------------------------------------------------------------

CREATE TABLE reading_passages (
    id               TEXT PRIMARY KEY,
    exam_version_id  TEXT NOT NULL REFERENCES exam_versions(id),
    exam             TEXT NOT NULL CHECK (exam IN ('PTE', 'IELTS')),
    title            TEXT NOT NULL,
    subtitle         TEXT NOT NULL DEFAULT '',

    -- The body, as an ordered array of {"label": "A", "text": "..."}.
    --
    -- Stored split rather than as one blob because the labels are answers:
    -- Matching Information asks which paragraph contains something, and the
    -- grader compares against "F". A blob would mean re-deriving those labels
    -- on every read and hoping the split matched the one the author intended.
    paragraphs       JSONB NOT NULL,

    -- Attributed excerpts, {"label": "Writer A", "text": "..."}, for passages
    -- that carry several voices. Find the Writer sets match a claim to one of
    -- these; passages without them leave this empty.
    sources          JSONB NOT NULL DEFAULT '[]',

    word_count       INT NOT NULL DEFAULT 0,
    difficulty       TEXT NOT NULL DEFAULT 'medium'
                         CHECK (difficulty IN ('easy', 'medium', 'hard')),
    topic            TEXT NOT NULL DEFAULT '',
    tags             TEXT[] NOT NULL DEFAULT '{}',
    is_published     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT reading_passages_paragraphs_array CHECK (jsonb_typeof(paragraphs) = 'array'),
    CONSTRAINT reading_passages_sources_array    CHECK (jsonb_typeof(sources) = 'array')
);

CREATE INDEX idx_reading_passages_exam ON reading_passages(exam) WHERE is_published;

-- ---------------------------------------------------------------------------
-- Question groups
-- ---------------------------------------------------------------------------

-- A group is one task set on one passage: a type, the instruction line that
-- goes above it, and the questions in it.
--
-- The same type may appear twice on a passage — a real IELTS passage often runs
-- two separate Yes/No/Not Given sets — so (passage_id, type_id) is not unique.
-- Position is what orders them.
CREATE TABLE reading_question_groups (
    id                 TEXT PRIMARY KEY,
    passage_id         TEXT NOT NULL REFERENCES reading_passages(id) ON DELETE CASCADE,
    position           INT NOT NULL,
    type_id            TEXT NOT NULL,
    type_name          TEXT NOT NULL,
    instructions       TEXT NOT NULL DEFAULT '',

    -- Material belonging to the task rather than to the passage: the four
    -- boxes of an ordering task, a summary with gaps, a list to match against.
    -- Same {"label", "text"} shape as paragraphs.
    resources          JSONB NOT NULL DEFAULT '[]',

    -- Whether the questions in this group may be dealt in a random order.
    -- FALSE for sets whose order is part of the task: gap-fills that track the
    -- passage top to bottom read as nonsense shuffled.
    shuffle_questions  BOOLEAN NOT NULL DEFAULT TRUE,

    time_limit_seconds INT NOT NULL DEFAULT 0,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (passage_id, position),
    CONSTRAINT reading_question_groups_resources_array CHECK (jsonb_typeof(resources) = 'array')
);

CREATE INDEX idx_reading_groups_passage ON reading_question_groups(passage_id, position);

-- Picking a random passage for a chosen task type is the hot path of practice.
CREATE INDEX idx_reading_groups_type ON reading_question_groups(type_id, passage_id);

-- ---------------------------------------------------------------------------
-- Questions belong to a group
-- ---------------------------------------------------------------------------

-- Nullable throughout: every question written before this migration is a
-- standalone task with no passage behind it, and stays that way.
ALTER TABLE questions
    ADD COLUMN passage_id     TEXT REFERENCES reading_passages(id) ON DELETE CASCADE,
    ADD COLUMN group_id       TEXT REFERENCES reading_question_groups(id) ON DELETE CASCADE,
    ADD COLUMN group_position INT NOT NULL DEFAULT 0;

-- A question in a group is a question on that group's passage. Storing both
-- keeps "questions on passage X" a single-table read, and this check is what
-- stops the two from drifting apart.
ALTER TABLE questions
    ADD CONSTRAINT questions_group_implies_passage
        CHECK (group_id IS NULL OR passage_id IS NOT NULL);

CREATE INDEX idx_questions_group ON questions(group_id, group_position)
    WHERE group_id IS NOT NULL;

CREATE INDEX idx_questions_passage ON questions(passage_id)
    WHERE passage_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- What each learner has already read
-- ---------------------------------------------------------------------------

-- One row per learner per passage per context.
--
-- Context is separate rather than a single "seen" flag because the two rules
-- differ: a mock must never re-deal a passage the learner has already sat, but
-- practice is allowed to come back to one — it just prefers not to.
CREATE TABLE user_passage_exposures (
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    passage_id    TEXT NOT NULL REFERENCES reading_passages(id) ON DELETE CASCADE,
    context       TEXT NOT NULL CHECK (context IN ('practice', 'mock')),
    seen_count    INT NOT NULL DEFAULT 1 CHECK (seen_count > 0),
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (user_id, passage_id, context)
);

-- Both selection queries read "this learner, this context, oldest first".
CREATE INDEX idx_passage_exposures_lookup
    ON user_passage_exposures(user_id, context, last_seen_at);

-- ---------------------------------------------------------------------------
-- Generated reading mocks
-- ---------------------------------------------------------------------------

-- A generated mock is composed per learner at the moment they start it, so
-- unlike the fixed blueprints in `mocks` there is no stored list of question
-- ids to grade against. This table is that list.
--
-- It is written when the mock is dealt, not when it is submitted. That is the
-- point: the passages are spent the moment the learner can read them, and a
-- learner who walks away must not be handed the same three next time.
CREATE TABLE reading_mock_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mock_id          TEXT NOT NULL REFERENCES mocks(id),
    exam             TEXT NOT NULL CHECK (exam IN ('PTE', 'IELTS')),
    exam_version_id  TEXT NOT NULL REFERENCES exam_versions(id),

    passage_ids      TEXT[] NOT NULL,
    -- Flattened, in the order they were dealt, including the shuffle. Grading
    -- reads this, so a client cannot add questions to its own paper.
    question_ids     TEXT[] NOT NULL,

    -- TRUE when the bank had fewer unseen passages than the mock needed and one
    -- had to be repeated. Reported to the learner rather than hidden.
    reused_passages  BOOLEAN NOT NULL DEFAULT FALSE,

    duration_minutes INT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'in_progress'
                         CHECK (status IN ('in_progress', 'submitted', 'abandoned')),
    mock_attempt_id  UUID REFERENCES mock_attempts(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at     TIMESTAMPTZ
);

CREATE INDEX idx_reading_mock_sessions_user
    ON reading_mock_sessions(user_id, created_at DESC);

-- One open session per learner per exam. Without this, a client that retries a
-- start request spends three more fresh passages on a paper nobody sat; with
-- it, the second call can only resume the first.
CREATE UNIQUE INDEX idx_reading_mock_sessions_live
    ON reading_mock_sessions(user_id, exam) WHERE status = 'in_progress';

-- ---------------------------------------------------------------------------
-- Generated mock blueprints
-- ---------------------------------------------------------------------------

-- Generated mocks still produce an ordinary mock_attempts row, and that row
-- references mocks(id). So a generated mock needs a blueprint — one that names
-- the paper and carries no fixed question list.
ALTER TABLE mocks
    ADD COLUMN is_generated BOOLEAN NOT NULL DEFAULT FALSE;

INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated) VALUES
    ('mock-ielts-reading-gen', 'ielts-2026-01', 'IELTS', 'IELTS Academic Reading Mock',
     'Three full reading passages, drawn fresh each time. You will not be given a passage you have already sat.',
     60, FALSE, TRUE);
