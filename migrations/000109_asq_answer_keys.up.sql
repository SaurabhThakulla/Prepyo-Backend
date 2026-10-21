-- 000109_asq_answer_keys.up.sql
-- Answer Short Question bank 3: accepted spoken answers, so replies are marked by matching (model answer first, shown in feedback).
-- The matcher already ignores case, articles, possessives, simple plurals and digits vs words; these list the other variants.

UPDATE questions q
SET correct_answers = v.answers::jsonb
FROM (VALUES
    ('pte-b3-asq-001', '["Scissors"]'),
    ('pte-b3-asq-002', '["A stamp"]'),
    ('pte-b3-asq-003', '["A dishwasher", "dish washer", "dishwashing machine"]'),
    ('pte-b3-asq-004', '["Socks"]'),
    ('pte-b3-asq-005', '["A towel"]'),
    ('pte-b3-asq-006', '["An alarm clock", "alarm"]'),
    ('pte-b3-asq-007', '["The platform"]'),
    ('pte-b3-asq-008', '["A marker", "whiteboard pen", "felt-tip pen"]'),
    ('pte-b3-asq-009', '["A mirror", "looking glass"]'),
    ('pte-b3-asq-010', '["A forest", "woods", "woodland", "rainforest", "jungle"]'),
    ('pte-b3-asq-011', '["The source", "spring", "headwaters"]'),
    ('pte-b3-asq-012', '["The coast", "coastline", "seashore", "shore", "seaside", "beach"]'),
    ('pte-b3-asq-013', '["Seven"]'),
    ('pte-b3-asq-014', '["A pharmacy", "chemist", "chemist''s", "drugstore", "drug store", "medical store", "medical shop"]'),
    ('pte-b3-asq-015', '["Your teeth", "tooth", "molars"]'),
    ('pte-b3-asq-016', '["A prescription"]'),
    ('pte-b3-asq-017', '["An ambulance"]'),
    ('pte-b3-asq-018', '["A vet", "veterinarian", "veterinary surgeon", "veterinary doctor", "animal doctor"]'),
    ('pte-b3-asq-019', '["Leave", "depart", "go away"]'),
    ('pte-b3-asq-020', '["Mice"]'),
    ('pte-b3-asq-021', '["Went"]'),
    ('pte-b3-asq-022', '["Wide", "broad"]'),
    ('pte-b3-asq-023', '["An abbreviation", "contraction"]'),
    ('pte-b3-asq-024', '["A puppy", "pup"]'),
    ('pte-b3-asq-025', '["Hail", "hailstones"]'),
    ('pte-b3-asq-026', '["A giraffe"]'),
    ('pte-b3-asq-027', '["Nectar"]'),
    ('pte-b3-asq-028', '["A tadpole", "polliwog", "pollywog"]'),
    ('pte-b3-asq-029', '["A librarian"]'),
    ('pte-b3-asq-030', '["A postman", "postwoman", "postmen", "postal worker", "mailman", "mail carrier", "letter carrier", "postie"]'),
    ('pte-b3-asq-031', '["A baker"]'),
    ('pte-b3-asq-032', '["Firefighters", "fire fighters", "fireman", "firemen", "firewoman", "fire brigade", "fire service", "fire department"]'),
    ('pte-b3-asq-033', '["A plumber"]'),
    ('pte-b3-asq-034', '["Mars"]'),
    ('pte-b3-asq-035', '["Steam", "water vapour", "vapour", "vapor"]'),
    ('pte-b3-asq-036', '["The leaf", "leaves"]'),
    ('pte-b3-asq-037', '["A zoologist"]'),
    ('pte-b3-asq-038', '["Iron"]'),
    ('pte-b3-asq-039', '["Seven"]'),
    ('pte-b3-asq-040', '["One hundred", "a hundred"]'),
    ('pte-b3-asq-041', '["Four"]'),
    ('pte-b3-asq-042', '["December"]'),
    ('pte-b3-asq-043', '["One hundred", "a hundred"]'),
    ('pte-b3-asq-044', '["The deadline", "due date"]'),
    ('pte-b3-asq-045', '["The syllabus", "curriculum", "course outline"]'),
    ('pte-b3-asq-046', '["A tutorial", "supervision"]'),
    ('pte-b3-asq-047', '["A student card", "student ID", "ID card", "identity card", "library card", "university card", "campus card"]'),
    ('pte-b3-asq-048', '["The Bagmati", "Bagmathi"]'),
    ('pte-b3-asq-049', '["Phewa Lake", "Phewa", "Fewa", "Phewa Tal", "Fewa Tal"]'),
    ('pte-b3-asq-050', '["Kathmandu", "Katmandu"]')
) AS v(id, answers)
WHERE q.id = v.id
  AND q.type_id IN ('answer-short-question', 'pte-answer-short-question')
  AND (q.correct_answers IS NULL OR q.correct_answers = '[]'::jsonb);
