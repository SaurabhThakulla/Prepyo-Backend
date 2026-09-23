-- The IELTS Listening mock: four parts, forty questions, each recording heard
-- once.
--
-- Why this exists. Listening could only be practised as one- and two-question
-- drills. The official format is four parts of ten questions, about 30
-- minutes, recordings heard once, with questions in the same order as the
-- information in the recording. A listening test is a set of four parts; a
-- part is one recording with its question groups; each question is one
-- numbered answer worth one mark.

CREATE TABLE listening_tests (
    id              TEXT PRIMARY KEY,
    exam_version_id TEXT NOT NULL REFERENCES exam_versions(id),
    title           TEXT NOT NULL,
    is_published    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE listening_parts (
    id              TEXT PRIMARY KEY,
    test_id         TEXT NOT NULL REFERENCES listening_tests(id) ON DELETE CASCADE,
    part_no         INT  NOT NULL CHECK (part_no BETWEEN 1 AND 4),
    -- What the learner is told before the recording, as the test announces it.
    setting         TEXT NOT NULL,
    -- The recording's script, with speaker labels. It is the answer key, so it
    -- is served only at play time and never in the paper itself.
    script          TEXT NOT NULL,
    -- Time to read the questions before the recording starts.
    reading_seconds INT  NOT NULL DEFAULT 30 CHECK (reading_seconds >= 0),
    UNIQUE (test_id, part_no)
);

CREATE TABLE listening_question_groups (
    id              TEXT PRIMARY KEY,
    part_id         TEXT NOT NULL REFERENCES listening_parts(id) ON DELETE CASCADE,
    position        INT  NOT NULL,
    type_id         TEXT NOT NULL,
    instructions    TEXT NOT NULL,
    -- A title for the form, notes or map the questions sit in.
    heading         TEXT NOT NULL DEFAULT '',
    image_url       TEXT,
    UNIQUE (part_id, position)
);

-- A listening mock question belongs to a group; practice lists leave these out.
ALTER TABLE questions
    ADD COLUMN listening_group_id TEXT REFERENCES listening_question_groups(id) ON DELETE CASCADE;

CREATE INDEX idx_questions_listening_group ON questions(listening_group_id, group_position)
    WHERE listening_group_id IS NOT NULL;

CREATE TABLE listening_mock_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    test_id          TEXT NOT NULL REFERENCES listening_tests(id),
    question_ids     TEXT[] NOT NULL,
    status           TEXT NOT NULL DEFAULT 'in_progress'
                         CHECK (status IN ('in_progress', 'submitted', 'abandoned')),
    duration_minutes INT  NOT NULL CHECK (duration_minutes > 0),
    expires_at       TIMESTAMPTZ NOT NULL,
    draft_answers    JSONB NOT NULL DEFAULT '[]'::jsonb,
    reused_test      BOOLEAN NOT NULL DEFAULT FALSE,
    -- How many recordings have started playing. Each is heard once: a paper
    -- reopened mid-test resumes at the next part, never replays one.
    parts_played     INT  NOT NULL DEFAULT 0 CHECK (parts_played BETWEEN 0 AND 4),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at     TIMESTAMPTZ,
    mock_attempt_id  UUID REFERENCES mock_attempts(id) ON DELETE SET NULL,
    CONSTRAINT listening_mock_sessions_drafts_array CHECK (jsonb_typeof(draft_answers) = 'array')
);

CREATE UNIQUE INDEX idx_listening_mock_live
    ON listening_mock_sessions(user_id) WHERE status = 'in_progress';

INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated, is_available) VALUES
    ('mock-ielts-listening-gen', 'ielts-2026-01', 'IELTS', 'IELTS Listening Mock',
     'Four parts and forty questions, each recording heard once, as in the test.',
     40, FALSE, TRUE, FALSE)
ON CONFLICT (id) DO NOTHING;
