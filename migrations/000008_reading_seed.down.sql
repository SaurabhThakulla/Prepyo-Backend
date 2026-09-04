-- Questions cascade from their group, groups from their passage, so dropping
-- the three passages takes the whole seed with it.
DELETE FROM reading_passages WHERE id IN ('rp-choc-01', 'rp-bees-01', 'rp-paper-01');
