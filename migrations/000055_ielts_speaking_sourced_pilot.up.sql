-- Three original, unofficial IELTS Speaking practice tasks, not official/recalled items.
-- Authoring source: original Prepyo practice content; no external factual claims.
-- Format reference ONLY (accessed 2026-09-18): IELTS partners, Speaking test format:
-- https://ielts.org/take-a-test/test-types/ielts-academic-test/ielts-academic-format-speaking
-- This reference applies to all three IDs below, not to their question wording.
-- Part 1/3 timers are practice budgets, not official per-question time limits.
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
    title, prompt, prep_time_seconds, time_limit_seconds, points, difficulty,
    tags, explanation
) VALUES
('is55-1-1', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'speaking',
 'ielts-speaking-part1', 'Introduction', 'Making weekend plans',
 'Do you prefer to plan your weekends in advance or decide what to do on the day? Why? Original, unofficial practice task.',
 0, 60, 10, 'medium', ARRAY['IELTS', 'Original practice', 'Unofficial', 'Sourced pilot 55'],
 'Give your preference, support it with a reason and illustrate it with an everyday example. A qualified answer depending on your responsibilities is acceptable. There is no single correct answer. Feedback considers fluency and coherence, lexical resource, grammatical range and accuracy, and pronunciation; practice feedback is not an official IELTS score.'),
('is55-2-1', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'speaking',
 'ielts-speaking-part2', 'Speaking Part 2 (Cue Card)', 'A plan interrupted by weather',
 E'Describe a time when the weather caused you to change a plan.\n\nYou should say:\n- what you had planned\n- what the weather was like\n- what you did instead\nand explain how you felt about the change.\n\nYou have 1 minute to prepare. Speak for 1-2 minutes. Original, unofficial practice task.',
 60, 120, 15, 'medium', ARRAY['IELTS', 'Original practice', 'Unofficial', 'Sourced pilot 55'],
 'Contrast the original plan with the alternative and explain your response. Organise events clearly and add relevant detail. An ordinary experience is sufficient; extreme weather is not required. There is no single correct answer. Feedback considers fluency and coherence, lexical resource, grammatical range and accuracy, and pronunciation; practice feedback is not an official IELTS score.'),
('is55-3-1', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'speaking',
 'ielts-speaking-part3', 'Speaking Part 3 (Discussion)', 'Planning under uncertainty',
 'Is it better for organisations to prepare several alternative plans or concentrate on making one plan work well? Original, unofficial practice task.',
 0, 120, 15, 'medium', ARRAY['IELTS', 'Original practice', 'Unofficial', 'Sourced pilot 55'],
 'Extend the changed-plan theme from Part 2 to organisations. Compare flexibility, preparation costs and the likelihood of disruption. Support a position with an example and consider circumstances that could change it. There is no single correct answer. Feedback considers fluency and coherence, lexical resource, grammatical range and accuracy, and pronunciation; practice feedback is not an official IELTS score.')
ON CONFLICT (id) DO NOTHING;
