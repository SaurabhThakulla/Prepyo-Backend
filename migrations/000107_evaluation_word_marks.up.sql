-- Read Aloud and Repeat Sentence results mark the given text word by word.

ALTER TABLE ai_evaluations ADD COLUMN word_marks JSONB NOT NULL DEFAULT '[]';
