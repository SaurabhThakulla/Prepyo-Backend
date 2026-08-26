-- Prepyo schema v1
--
-- Two kinds of table live here:
--   * content  - exams, questions, mocks, plans. Shared by every learner.
--   * learner  - users, attempts, mistakes, xp. Always scoped by user_id.
--
-- Content rows use readable text ids ('pte-spk-001') because they are authored
-- by hand. Learner rows use uuids because the server generates them.

-- ---------------------------------------------------------------------------
-- Users and sessions
-- ---------------------------------------------------------------------------

CREATE TABLE users (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                    TEXT NOT NULL UNIQUE,
    password_hash            TEXT NOT NULL,
    name                     TEXT NOT NULL,
    role                     TEXT NOT NULL DEFAULT 'learner'
                                 CHECK (role IN ('learner', 'admin')),

    target_exam              TEXT NOT NULL DEFAULT 'PTE'
                                 CHECK (target_exam IN ('PTE', 'IELTS')),
    target_score             NUMERIC(4, 1),
    exam_date                DATE,
    nepal_region             TEXT NOT NULL DEFAULT 'Kathmandu',

    xp                       INT NOT NULL DEFAULT 0 CHECK (xp >= 0),
    streak_days              INT NOT NULL DEFAULT 0 CHECK (streak_days >= 0),
    streak_last_active_date  DATE,
    timezone                 TEXT NOT NULL DEFAULT 'Asia/Kathmandu',

    plan_id                  TEXT NOT NULL DEFAULT 'free',
    plan_valid_until         DATE,

    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Level and estimated score are derived, not stored: level comes from xp, and
-- the score estimate comes from recent attempts. Storing them lets the two
-- drift apart.

CREATE TABLE sessions (
    -- We store a SHA-256 of the token, never the token itself, so a database
    -- leak does not hand out live sessions.
    token_hash  BYTEA PRIMARY KEY,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- ---------------------------------------------------------------------------
-- Exam content
-- ---------------------------------------------------------------------------

-- An exam version freezes the section list, task types and scoring rules that
-- applied at a point in time. Attempts reference the version they were taken
-- under so old results never change when the exam definition is updated.
CREATE TABLE exam_versions (
    id           TEXT PRIMARY KEY,
    exam         TEXT NOT NULL CHECK (exam IN ('PTE', 'IELTS')),
    label        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    min_score    NUMERIC(4, 1) NOT NULL,
    max_score    NUMERIC(4, 1) NOT NULL,
    score_step   NUMERIC(3, 2) NOT NULL,
    is_current   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Only one version per exam may be the current one.
CREATE UNIQUE INDEX idx_exam_versions_current
    ON exam_versions(exam) WHERE is_current;

CREATE TABLE questions (
    id                 TEXT PRIMARY KEY,
    exam_version_id    TEXT NOT NULL REFERENCES exam_versions(id),
    exam               TEXT NOT NULL CHECK (exam IN ('PTE', 'IELTS')),
    skill              TEXT NOT NULL
                           CHECK (skill IN ('speaking', 'writing', 'reading', 'listening')),
    type_id            TEXT NOT NULL,
    type_name          TEXT NOT NULL,
    title              TEXT NOT NULL,
    prompt             TEXT NOT NULL,
    context_passage    TEXT,
    audio_url          TEXT,
    audio_transcript   TEXT,
    image_url          TEXT,
    prep_time_seconds  INT NOT NULL DEFAULT 0,
    time_limit_seconds INT NOT NULL,
    options            JSONB,
    correct_answers    JSONB,
    blanks             JSONB,
    model_answer       TEXT,
    explanation        TEXT,
    difficulty         TEXT NOT NULL DEFAULT 'medium'
                           CHECK (difficulty IN ('easy', 'medium', 'hard')),
    tags               TEXT[] NOT NULL DEFAULT '{}',
    points             INT NOT NULL DEFAULT 10,
    is_published       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Practice picks questions by exam + skill, which is the hot path.
CREATE INDEX idx_questions_exam_skill ON questions(exam, skill) WHERE is_published;

CREATE TABLE mocks (
    id                     TEXT PRIMARY KEY,
    exam_version_id        TEXT NOT NULL REFERENCES exam_versions(id),
    exam                   TEXT NOT NULL CHECK (exam IN ('PTE', 'IELTS')),
    title                  TEXT NOT NULL,
    description            TEXT NOT NULL DEFAULT '',
    total_duration_minutes INT NOT NULL,
    is_diagnostic          BOOLEAN NOT NULL DEFAULT FALSE,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE mock_sections (
    id               TEXT PRIMARY KEY,
    mock_id          TEXT NOT NULL REFERENCES mocks(id) ON DELETE CASCADE,
    position         INT NOT NULL,
    name             TEXT NOT NULL,
    skill            TEXT NOT NULL
                         CHECK (skill IN ('speaking', 'writing', 'reading', 'listening')),
    duration_minutes INT NOT NULL,
    question_ids     TEXT[] NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_mock_sections_mock ON mock_sections(mock_id, position);

-- ---------------------------------------------------------------------------
-- Learner activity
-- ---------------------------------------------------------------------------

CREATE TABLE practice_attempts (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id         TEXT NOT NULL REFERENCES questions(id),
    exam_version_id     TEXT NOT NULL REFERENCES exam_versions(id),
    is_correct          BOOLEAN NOT NULL,
    score               NUMERIC(6, 2) NOT NULL,
    max_score           NUMERIC(6, 2) NOT NULL,
    accuracy_percentage INT NOT NULL,
    user_response       TEXT,
    feedback            TEXT,
    time_spent_seconds  INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_practice_attempts_user ON practice_attempts(user_id, created_at DESC);

CREATE TABLE mock_attempts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mock_id          TEXT NOT NULL REFERENCES mocks(id),
    exam_version_id  TEXT NOT NULL REFERENCES exam_versions(id),
    exam             TEXT NOT NULL,
    user_score       NUMERIC(5, 2) NOT NULL,
    skill_scores     JSONB NOT NULL,
    total_correct    INT NOT NULL,
    total_questions  INT NOT NULL,
    duration_seconds INT NOT NULL,
    completed_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_mock_attempts_user ON mock_attempts(user_id, completed_at DESC);

CREATE TABLE mistakes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id       TEXT NOT NULL REFERENCES questions(id),
    error_tag         TEXT NOT NULL,
    user_response     TEXT NOT NULL DEFAULT '',
    correct_response  TEXT NOT NULL DEFAULT '',
    explanation       TEXT NOT NULL DEFAULT '',
    failed_count      INT NOT NULL DEFAULT 1,
    resolved          BOOLEAN NOT NULL DEFAULT FALSE,
    last_attempted_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- One row per learner per question; repeat failures bump failed_count.
    UNIQUE (user_id, question_id)
);

CREATE INDEX idx_mistakes_user_unresolved
    ON mistakes(user_id, last_attempted_at DESC) WHERE NOT resolved;

CREATE TABLE ai_evaluations (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id          TEXT REFERENCES questions(id),
    exam                 TEXT NOT NULL,
    skill                TEXT NOT NULL,
    evaluation_version   TEXT NOT NULL,

    -- Idempotency key = hash of (user, question, submitted text). Re-submitting
    -- the same answer returns the stored evaluation instead of paying for a
    -- second provider call.
    request_fingerprint  BYTEA NOT NULL,

    estimated_score      NUMERIC(5, 2),
    score_confidence     TEXT NOT NULL CHECK (score_confidence IN ('low', 'medium', 'high')),
    summary              TEXT NOT NULL DEFAULT '',
    criteria             JSONB NOT NULL DEFAULT '[]',
    strengths            JSONB NOT NULL DEFAULT '[]',
    weaknesses           JSONB NOT NULL DEFAULT '[]',
    sentence_feedback    JSONB NOT NULL DEFAULT '[]',
    model_rewrite        TEXT,

    provider             TEXT NOT NULL,
    model                TEXT NOT NULL,
    prompt_version       TEXT NOT NULL,
    prompt_tokens        INT NOT NULL DEFAULT 0,
    completion_tokens    INT NOT NULL DEFAULT 0,
    latency_ms           INT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, request_fingerprint)
);

CREATE INDEX idx_ai_evaluations_user ON ai_evaluations(user_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- XP, missions, notifications
-- ---------------------------------------------------------------------------

CREATE TABLE xp_transactions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount      INT NOT NULL CHECK (amount > 0),
    reason      TEXT NOT NULL,

    -- Every award names the thing it is paying for, e.g.
    -- 'practice_attempt:<uuid>'. The unique index below makes a repeated
    -- award a no-op, which is what stops XP farming by replaying a request.
    source_key  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, source_key)
);

CREATE INDEX idx_xp_transactions_user ON xp_transactions(user_id, created_at DESC);

CREATE TABLE daily_missions (
    id           TEXT PRIMARY KEY,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL,
    exam         TEXT CHECK (exam IN ('PTE', 'IELTS')),
    skill        TEXT NOT NULL
                     CHECK (skill IN ('speaking', 'writing', 'reading', 'listening')),
    task_type    TEXT NOT NULL DEFAULT '',
    target_count INT NOT NULL CHECK (target_count > 0),
    xp_reward    INT NOT NULL CHECK (xp_reward > 0),
    is_active    BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE user_mission_progress (
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mission_id      TEXT NOT NULL REFERENCES daily_missions(id) ON DELETE CASCADE,
    mission_date    DATE NOT NULL,
    completed_count INT NOT NULL DEFAULT 0,
    completed_at    TIMESTAMPTZ,

    PRIMARY KEY (user_id, mission_id, mission_date)
);

CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title      TEXT NOT NULL,
    message    TEXT NOT NULL,
    type       TEXT NOT NULL
                   CHECK (type IN ('streak', 'evaluation', 'mission', 'system')),
    action_url TEXT,
    read       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, created_at DESC);

-- ---------------------------------------------------------------------------
-- Billing
-- ---------------------------------------------------------------------------

-- Plans are data so pricing can change without a code deploy.
CREATE TABLE plans (
    id                     TEXT PRIMARY KEY,
    name                   TEXT NOT NULL,
    price_npr              INT NOT NULL,
    duration_months        INT NOT NULL,
    features               TEXT[] NOT NULL DEFAULT '{}',
    ai_evaluations_per_day INT NOT NULL,
    mock_tests_included    INT NOT NULL,
    is_popular             BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order             INT NOT NULL DEFAULT 0
);

ALTER TABLE users
    ADD CONSTRAINT users_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES plans(id);
