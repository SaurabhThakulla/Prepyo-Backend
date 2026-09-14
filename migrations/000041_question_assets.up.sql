CREATE TABLE IF NOT EXISTS question_assets (
    id TEXT PRIMARY KEY,
    content_type TEXT NOT NULL,
    byte_size INT NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_question_assets_created_at ON question_assets(created_at);
