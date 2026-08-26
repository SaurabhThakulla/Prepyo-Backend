-- 000004: Make listening practice answerable.
--
-- The seed in 000002 inserted questions without touching audio_url or
-- audio_transcript — both columns were simply left off the INSERT column list.
-- That left pte-lis-001, a Write from Dictation task, telling the learner "You
-- will hear a sentence" with nothing to hear and no transcript to fall back on.
-- IELTS had no listening question at all, so that skill rendered empty.
--
-- These transcripts are the stand-in the practice UI already knows how to
-- render, labelled as a script rather than dressed up as a player. Real audio
-- files replace them by setting audio_url; the UI prefers audio whenever it is
-- present.

-- The correct answer for a dictation task is the sentence itself, so the
-- transcript has to match correct_answers exactly.
UPDATE questions
SET audio_transcript = 'All submitted assignments must contain appropriate academic citations and a comprehensive bibliography.'
WHERE id = 'pte-lis-001';

INSERT INTO questions (
    id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('ielts-lis-001', 'ielts-2026-01', 'IELTS', 'listening', 'ielts-listening-form', 'Form Completion',
        'Community Library Membership',
        'Listen to the recording and complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.',
        E'Membership type: ____ (1)\nAnnual fee: NPR ____ (2)\nBorrowing limit: ____ (3) books',
        E'Receptionist: Good morning, Lalitpur Community Library. How can I help you?\n'
        'Caller: I would like to join, please.\n'
        'Receptionist: Certainly. We offer two categories. The standard membership is for general readers, and the student membership carries a reduced rate.\n'
        'Caller: I am at university, so student membership please.\n'
        'Receptionist: That is fine. The student rate is eight hundred rupees for the year, compared with twelve hundred for standard.\n'
        'Caller: And how many books can I take out?\n'
        'Receptionist: Students may borrow six books at a time, for three weeks each.',
        0, 300,
        '["student","800","six"]'::jsonb, NULL,
        NULL,
        'Form completion answers are usually said once and not repeated. The word limit is part of the marking: "eight hundred rupees" is written 800, not "800 rupees".',
        'medium', ARRAY['IELTS Listening', 'Form Completion'], 12);

-- The IELTS diagnostic advertises "all four skills" but only ran Reading and
-- Speaking, because Writing and Listening had nowhere to draw from. Both mocks
-- now cover what their descriptions claim.
INSERT INTO mock_sections (id, mock_id, position, name, skill, duration_minutes, question_ids) VALUES
    ('sec-ielts-diag-3', 'mock-ielts-diag-01', 3, 'Writing',   'writing',   10, ARRAY['ielts-wrt-001']),
    ('sec-ielts-diag-4', 'mock-ielts-diag-01', 4, 'Listening', 'listening',  5, ARRAY['ielts-lis-001']),
    ('sec-ielts-full-4', 'mock-ielts-full-01', 4, 'Listening', 'listening', 30, ARRAY['ielts-lis-001']);

-- Keep the advertised total honest: it is what the mock card shows a learner
-- before they commit to sitting it.
UPDATE mocks SET total_duration_minutes = 45  WHERE id = 'mock-ielts-diag-01';
UPDATE mocks SET total_duration_minutes = 165 WHERE id = 'mock-ielts-full-01';
