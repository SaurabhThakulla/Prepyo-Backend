-- Who transcribed a spoken answer in a PTE mock: the server (Whisper on the
-- audio provider), as the IELTS speaking mock does, or the learner's browser
-- where the server could not. The recording itself is never stored; it is
-- transcribed and dropped.
ALTER TABLE pte_mock_items
    ADD COLUMN transcript_source TEXT CHECK (transcript_source IN ('browser', 'server'));
