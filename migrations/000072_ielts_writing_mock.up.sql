-- The IELTS Writing mock: Task 1 and Task 2 in one 60-minute sitting.
--
-- Why this exists. Writing could only be practised one task at a time, and a
-- single task estimate is not a Writing band: IELTS rates both tasks, and
-- "Task 2 contributes twice as much as Task 1 to the Writing score". The mock
-- deals one Task 1 for the learner's module (a visual for Academic, a letter
-- for General Training) and one Task 2, keeps time on the server, saves drafts
-- as the learner writes, and reports a Writing band weighted 1:2.

-- The attempt row every mock result references. It is not offered in the
-- generic mock list, which starts papers through a different flow; the
-- sidebar's Writing mock link starts it.
INSERT INTO mocks (id, exam_version_id, exam, title, description,
                   total_duration_minutes, is_diagnostic, is_generated, is_available) VALUES
    ('mock-ielts-writing-gen', 'ielts-2026-01', 'IELTS', 'IELTS Writing Mock',
     'Task 1 and Task 2 in 60 minutes, rated against the public band descriptors. Task 2 counts twice as much as Task 1.',
     60, FALSE, TRUE, FALSE)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE writing_mock_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exam                TEXT NOT NULL CHECK (exam = 'IELTS'),
    module              TEXT NOT NULL CHECK (module IN ('academic', 'general_training')),
    task1_id            TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    task2_id            TEXT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    -- The latest text saved while time remained. A submission after the
    -- deadline is graded from these.
    task1_draft         TEXT NOT NULL DEFAULT '',
    task2_draft         TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL DEFAULT 'in_progress'
                            CHECK (status IN ('in_progress', 'submitted', 'abandoned')),
    duration_minutes    INT  NOT NULL CHECK (duration_minutes > 0),
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    submitted_at        TIMESTAMPTZ,
    mock_attempt_id     UUID REFERENCES mock_attempts(id) ON DELETE SET NULL,
    task1_evaluation_id UUID REFERENCES ai_evaluations(id) ON DELETE SET NULL,
    task2_evaluation_id UUID REFERENCES ai_evaluations(id) ON DELETE SET NULL
);

-- One open paper per learner: a retried start resumes it instead of dealing
-- and charging for another.
CREATE UNIQUE INDEX idx_writing_mock_live
    ON writing_mock_sessions(user_id) WHERE status = 'in_progress';

CREATE INDEX idx_writing_mock_user ON writing_mock_sessions(user_id, created_at DESC);
