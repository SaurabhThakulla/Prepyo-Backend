-- Where a speaking answer's recording is kept (Backblaze B2 or a local folder), for 30 days.
ALTER TABLE ai_evaluations ADD COLUMN IF NOT EXISTS recording_key TEXT;
