ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_avatar_complete,
    DROP CONSTRAINT IF EXISTS users_cover_complete;

ALTER TABLE users
    DROP COLUMN IF EXISTS avatar_image,
    DROP COLUMN IF EXISTS avatar_content_type,
    DROP COLUMN IF EXISTS avatar_updated_at,
    DROP COLUMN IF EXISTS cover_image,
    DROP COLUMN IF EXISTS cover_content_type,
    DROP COLUMN IF EXISTS cover_updated_at;
