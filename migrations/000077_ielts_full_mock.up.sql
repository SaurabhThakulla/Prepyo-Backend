-- The IELTS full mock: Listening, Reading, Writing and Speaking as one test,
-- with an overall band.
--
-- Why this exists. Each skill could be sat as a section mock, but nothing put
-- the four together or reported the overall band, which IELTS gives as "the
-- mean of the four component scores" rounded to the nearest half band. A full
-- mock runs the four sections in the order of the test (Listening, Reading and
-- Writing, then Speaking, which may be taken on another day) and reports the
-- four bands with the overall.
--
-- A full mock is paid for from the plan's full-mock allowance, and spends it
-- when it starts, as booking a test does: counting it only when finished would
-- let a learner take section after section free and never finish. Its four
-- sections are not charged sub-tests. Each section still records its own
-- result, so its review is in the learner's history.

INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated, is_available) VALUES
    ('mock-ielts-full', 'ielts-2026-01', 'IELTS', 'IELTS Full Mock Test',
     'Listening, Reading, Writing and Speaking in the order of the test, with an overall band.',
     165, FALSE, FALSE, FALSE)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE full_mock_sessions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    module               TEXT NOT NULL CHECK (module IN ('academic', 'general_training')),
    status               TEXT NOT NULL DEFAULT 'in_progress'
                             CHECK (status IN ('in_progress', 'completed', 'abandoned')),
    -- Each section's paper, set when the learner reaches it.
    listening_session_id UUID REFERENCES listening_mock_sessions(id) ON DELETE SET NULL,
    reading_session_id   UUID REFERENCES reading_mock_sessions(id) ON DELETE SET NULL,
    writing_session_id   UUID REFERENCES writing_mock_sessions(id) ON DELETE SET NULL,
    speaking_session_id  UUID REFERENCES speaking_mock_sessions(id) ON DELETE SET NULL,
    mock_attempt_id      UUID REFERENCES mock_attempts(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at         TIMESTAMPTZ
);

-- One open full mock per learner: a retried start resumes it rather than
-- spending the allowance again.
CREATE UNIQUE INDEX idx_full_mock_live
    ON full_mock_sessions(user_id) WHERE status = 'in_progress';

CREATE INDEX idx_full_mock_user ON full_mock_sessions(user_id, created_at DESC);

-- A full mock's sections are papers of their own, dealt fresh when the learner
-- reaches each section. They are marked so that they sit beside a section
-- mock the learner has open on its own rather than replacing it: a full mock
-- never picks up an old paper, perhaps long out of time, and an open section
-- mock is not swallowed by a full mock.
ALTER TABLE listening_mock_sessions ADD COLUMN in_full_mock BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE reading_mock_sessions   ADD COLUMN in_full_mock BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE writing_mock_sessions   ADD COLUMN in_full_mock BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE speaking_mock_sessions  ADD COLUMN in_full_mock BOOLEAN NOT NULL DEFAULT FALSE;

DROP INDEX idx_listening_mock_live;
CREATE UNIQUE INDEX idx_listening_mock_live
    ON listening_mock_sessions(user_id) WHERE status = 'in_progress' AND NOT in_full_mock;
DROP INDEX idx_reading_mock_sessions_live;
CREATE UNIQUE INDEX idx_reading_mock_sessions_live
    ON reading_mock_sessions(user_id, exam) WHERE status = 'in_progress' AND NOT in_full_mock;
DROP INDEX idx_writing_mock_live;
CREATE UNIQUE INDEX idx_writing_mock_live
    ON writing_mock_sessions(user_id) WHERE status = 'in_progress' AND NOT in_full_mock;
DROP INDEX idx_speaking_mock_live;
CREATE UNIQUE INDEX idx_speaking_mock_live
    ON speaking_mock_sessions(user_id) WHERE status = 'in_progress' AND NOT in_full_mock;
