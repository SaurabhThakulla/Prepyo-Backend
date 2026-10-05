-- Stored recordings for the IELTS mocks, made offline by prepyo-voicegen, so
-- the mocks play the same audio on every device instead of a live voice.

-- A listening part's recording. Empty until one is stored; the mock then
-- falls back to reading the script aloud as before.
ALTER TABLE listening_parts ADD COLUMN IF NOT EXISTS audio_url TEXT;

-- The speaking examiner's recorded lines, found by their exact text. A line
-- whose wording changes has no match and goes back to the live voice until it
-- is recorded again.
CREATE TABLE IF NOT EXISTS voice_clips (
    text       TEXT PRIMARY KEY,
    asset_id   TEXT NOT NULL REFERENCES question_assets(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
