-- 000040: Track practice and test sessions for start-time credit consumption.

CREATE TABLE practice_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    exam        TEXT NOT NULL,
    skill       TEXT NOT NULL,
    item_id     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'stopped')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    stopped_at  TIMESTAMPTZ
);

CREATE INDEX idx_practice_sessions_user_created ON practice_sessions(user_id, created_at DESC);
CREATE INDEX idx_practice_sessions_status ON practice_sessions(user_id, status);
