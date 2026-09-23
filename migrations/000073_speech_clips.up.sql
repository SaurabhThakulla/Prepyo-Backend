-- Server-side speech for browsers that cannot speak a script themselves.
--
-- Listening scripts are read by the browser's own voice where it has one. A
-- browser without speech synthesis gets the same script spoken by the audio
-- provider (Groq Orpheus by default) instead. Orpheus takes 200 characters a
-- request, so a script is spoken as a run of short clips, one voice per
-- speaker. Each clip is made the first time someone plays it and kept, so a
-- script is paid for once, not once per learner.
--
-- id is a hash of model, voice and text: a clip is reused only for exactly the
-- same words in the same voice from the same model.
CREATE TABLE speech_clips (
    id              TEXT PRIMARY KEY,
    model           TEXT NOT NULL,
    voice           TEXT NOT NULL,
    text            TEXT NOT NULL,
    content_type    TEXT NOT NULL DEFAULT 'audio/wav',
    data            BYTEA,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    synthesized_at  TIMESTAMPTZ
);
