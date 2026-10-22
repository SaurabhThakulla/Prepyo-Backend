-- Question audio and images can live in the media bucket (B2) instead of the
-- database: storage_key names the object, and data is then left empty.
ALTER TABLE question_assets ADD COLUMN IF NOT EXISTS storage_key TEXT;
ALTER TABLE question_assets ALTER COLUMN data DROP NOT NULL;
ALTER TABLE question_assets
    ADD CONSTRAINT question_assets_has_bytes CHECK (data IS NOT NULL OR storage_key IS NOT NULL);
