-- One row per tutor reply a learner received, so the tutor can have a daily
-- allowance. Every message is a paid provider call; before this the only limit
-- was 20 a minute per IP, which a script could keep up all day.
CREATE TABLE tutor_messages (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tutor_messages_user_created ON tutor_messages(user_id, created_at DESC);
