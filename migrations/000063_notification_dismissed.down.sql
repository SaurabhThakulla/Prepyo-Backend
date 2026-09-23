DELETE FROM notifications WHERE dismissed;
ALTER TABLE notifications DROP COLUMN IF EXISTS dismissed;
