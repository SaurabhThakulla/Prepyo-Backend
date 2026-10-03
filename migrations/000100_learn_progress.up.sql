-- A learner's place in the Learn Korean course, kept on the server so it
-- follows them across devices. The browser owns the shape of the record (which
-- lessons are done, XP, streak, word review schedule) and merges it with this
-- copy; the server stores one JSON object per learner.
CREATE TABLE IF NOT EXISTS learn_progress (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    data       JSONB NOT NULL CHECK (jsonb_typeof(data) = 'object'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
