-- Profile photo and cover, stored as bytes on the user row.
--
-- These lived in the browser's localStorage, which meant an avatar set on a
-- phone was invisible on a laptop and a cleared cache erased it. Keeping them
-- here makes them the account's, not the device's.
--
-- The bytes deliberately do not join selectUserFields: session middleware loads
-- the user row on every authenticated request, and it has no business dragging
-- a megabyte of JPEG along with it. Only the served image endpoint reads them.
-- The timestamps are cheap enough to carry, and are what tells the client an
-- image exists at all.

ALTER TABLE users
    ADD COLUMN avatar_image        BYTEA,
    ADD COLUMN avatar_content_type TEXT,
    ADD COLUMN avatar_updated_at   TIMESTAMPTZ,
    ADD COLUMN cover_image         BYTEA,
    ADD COLUMN cover_content_type  TEXT,
    ADD COLUMN cover_updated_at    TIMESTAMPTZ;

-- Bytes and their type travel together: one without the other cannot be served.
ALTER TABLE users
    ADD CONSTRAINT users_avatar_complete CHECK (
        (avatar_image IS NULL AND avatar_content_type IS NULL AND avatar_updated_at IS NULL)
        OR (avatar_image IS NOT NULL AND avatar_content_type IS NOT NULL AND avatar_updated_at IS NOT NULL)
    ),
    ADD CONSTRAINT users_cover_complete CHECK (
        (cover_image IS NULL AND cover_content_type IS NULL AND cover_updated_at IS NULL)
        OR (cover_image IS NOT NULL AND cover_content_type IS NOT NULL AND cover_updated_at IS NOT NULL)
    );
