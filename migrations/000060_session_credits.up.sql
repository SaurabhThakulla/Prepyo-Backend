-- A practice session can now cost more than one sub-test. Ordinary tasks stay
-- at 1; a section mock (a full reading paper, for instance) costs 5, because it
-- is several task sets taken in one sitting. Daily usage becomes SUM(credits)
-- rather than count(*), so every existing row keeps counting exactly as before.
ALTER TABLE practice_sessions
    ADD COLUMN credits INT NOT NULL DEFAULT 1 CHECK (credits > 0);
