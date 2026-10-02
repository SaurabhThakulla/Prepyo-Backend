-- A listening item can carry an English version of its script, written and
-- spoken, so a learner who did not follow the Korean can hear what was said.
-- Both are shown only after the answer is in: the script is the answer key.
--
-- audio_translation follows the script's speaker labels in English ("Man:",
-- "Woman:") so each speaker keeps their voice. translation_audio_url points
-- at a stored recording of it, made offline by prepyo-voicegen.
ALTER TABLE questions ADD COLUMN IF NOT EXISTS audio_translation TEXT;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS translation_audio_url TEXT;

UPDATE questions AS q SET audio_translation = t.english
FROM (VALUES
    ('eps-seed-l-sound-1', 'Factory.'),
    ('eps-seed-l-sound-2', 'Market.'),
    ('eps-seed-l-picture-1', 'I drink milk.'),
    ('eps-seed-l-picture-2', 'Woman: What are you doing now? Man: I''m moving boxes.'),
    ('eps-seed-l-response-1', 'Man: Where are you from?'),
    ('eps-seed-l-response-2', 'Woman: What time do you finish work today?'),
    ('eps-seed-l-next-1', 'Woman: This machine is dangerous, so don''t use it on your own.'),
    ('eps-seed-l-next-2', 'Man: I have to go to the hospital tomorrow, so I''d like to take the morning off.'),
    ('eps-seed-l-number-1', 'Woman: How much are these gloves? Man: They''re three thousand five hundred won.'),
    ('eps-seed-l-number-2', 'Man: When is payday? Woman: The fifteenth of every month.'),
    ('eps-seed-l-pq-1', 'Man: Where is the bag?'),
    ('eps-seed-l-pq-2', 'Woman: What time is it now?'),
    ('eps-seed-l-dialogue-1', 'Man: Sunita, we have a team dinner tomorrow. Can you come? Woman: I can''t go tomorrow because I have a friend''s wedding. Man: Really? Then let''s go together next time.'),
    ('eps-seed-l-dialogue-2', 'Woman: Welcome. How can I help you? Man: I''d like to send money home. Woman: Please fill in this form and show me your alien registration card.')
) AS t(id, english)
WHERE q.id = t.id AND COALESCE(q.audio_translation, '') = '';
