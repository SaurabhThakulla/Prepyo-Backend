-- Fails while any asset is only in the bucket: run `server assets -restore -yes` first.
ALTER TABLE question_assets DROP CONSTRAINT IF EXISTS question_assets_has_bytes;
ALTER TABLE question_assets ALTER COLUMN data SET NOT NULL;
ALTER TABLE question_assets DROP COLUMN IF EXISTS storage_key;
