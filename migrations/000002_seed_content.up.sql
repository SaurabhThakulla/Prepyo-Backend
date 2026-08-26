-- Content seed: exam versions, plans, a starter question bank, mock blueprints
-- and the daily mission catalogue.
--
-- This file contains no learner data. Users are created through
-- POST /api/v1/auth/register like everyone else.

INSERT INTO exam_versions (id, exam, label, description, min_score, max_score, score_step, is_current) VALUES
    ('pte-2026-01',   'PTE',   'PTE Academic 2026.1',   'Pearson Test of English Academic.', 10, 90, 1,   TRUE),
    ('ielts-2026-01', 'IELTS', 'IELTS Academic 2026.1', 'IELTS Academic, band scale.',        0,  9, 0.5, TRUE);

INSERT INTO plans (id, name, price_npr, duration_months, features, ai_evaluations_per_day, mock_tests_included, is_popular, sort_order) VALUES
    ('free', 'Free Learner', 0, 0,
        ARRAY['Core practice tasks', '3 AI evaluations per day', '1 full mock exam'],
        3, 1, FALSE, 1),
    ('pro', 'Pro Prep', 1499, 1,
        ARRAY['Unlimited practice', '30 AI evaluations per day', '5 full mock exams', 'Sentence-level rewrites'],
        30, 5, TRUE, 2),
    ('elite', 'Elite Master', 2999, 3,
        ARRAY['90 days full access', '100 AI evaluations per day', '15 full mock exams', 'Priority evaluation queue'],
        100, 15, FALSE, 3);

-- ---------------------------------------------------------------------------
-- Questions
-- ---------------------------------------------------------------------------

INSERT INTO questions (
    id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
    context_passage, prep_time_seconds, time_limit_seconds, correct_answers,
    blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-spk-001', 'pte-2026-01', 'PTE', 'speaking', 'read-aloud', 'Read Aloud',
        'Renewable Energy Transition in Developing Economies',
        'Look at the text below. In 35 seconds, read this text aloud as naturally and clearly as possible.',
        'The global transition toward sustainable energy sources represents not only an ecological imperative but also a significant economic opportunity for developing economies. By decentralising power generation through solar microgrids and localised wind harvesting, rural communities can achieve uninterrupted electricity access while reducing greenhouse gas emissions.',
        35, 40, NULL, NULL,
        NULL,
        'Read at a steady pace. Do not stop to correct a small slip; continuous delivery scores better than a restart.',
        'medium', ARRAY['PTE Speaking', 'Read Aloud'], 15),

    ('pte-wrt-001', 'pte-2026-01', 'PTE', 'writing', 'summarize-written-text', 'Summarize Written Text',
        'Urban Heat Islands and Microclimates',
        'Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.',
        'Urban heat islands occur when cities replace natural land cover with dense concentrations of pavement and buildings that retain heat, increasing cooling energy costs and elevating heat-related illnesses. Municipal planners increasingly respond with reflective roofing materials, expanded tree canopies and permeable surfaces.',
        0, 600, NULL, NULL,
        'Because dense urban surfaces retain heat and raise both energy costs and health risks, planners are adopting reflective materials, green roofs and wider tree cover to reduce urban heat island effects.',
        'The response must be a single sentence within the 5-75 word range. A second sentence scores zero for form.',
        'medium', ARRAY['PTE Writing', 'Summarize Written Text'], 10),

    ('pte-rdg-001', 'pte-2026-01', 'PTE', 'reading', 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
        'Cognitive Benefits of Bilingualism',
        'Choose the word that best fits each blank.',
        'For decades, bilingual children were assumed to suffer a lasting academic disadvantage. Contemporary neurological research has [[b1]] this misconception, showing instead that managing two language systems [[b2]] executive function.',
        0, 180, NULL,
        '[{"id":"b1","options":["reinforced","debunked","ignored","predicted"],"correctAnswer":"debunked"},
          {"id":"b2","options":["strengthens","weakens","replaces","delays"],"correctAnswer":"strengthens"}]'::jsonb,
        NULL,
        '"Debunked" means shown to be false, which matches "instead" in the following clause. "Reinforced" would contradict it.',
        'medium', ARRAY['PTE Reading', 'Fill in the Blanks'], 10),

    ('pte-rdg-002', 'pte-2026-01', 'PTE', 'reading', 'reorder-paragraphs', 'Re-order Paragraphs',
        'The Spread of Movable Type',
        'The text boxes below are in the wrong order. Restore the original order.',
        NULL, 0, 300,
        '["p1","p3","p2","p4"]'::jsonb, NULL,
        NULL,
        'Find the sentence that introduces the topic without back-references. Sentences opening with "This", "Such" or "However" always follow something.',
        'hard', ARRAY['PTE Reading', 'Re-order Paragraphs'], 12),

    ('pte-lis-001', 'pte-2026-01', 'PTE', 'listening', 'write-from-dictation', 'Write from Dictation',
        'Academic Submission Guidelines',
        'You will hear a sentence. Type the sentence exactly as you hear it.',
        NULL, 0, 60,
        '["All submitted assignments must contain appropriate academic citations and a comprehensive bibliography."]'::jsonb,
        NULL, NULL,
        'Every word counts, including plurals and articles. Write what you hear, not what you expect to hear.',
        'medium', ARRAY['PTE Listening', 'Write from Dictation'], 10),

    ('ielts-wrt-001', 'ielts-2026-01', 'IELTS', 'writing', 'ielts-writing-task2', 'Writing Task 2',
        'Remote Work versus Office Work',
        'In many countries employees increasingly work from home. Do the advantages outweigh the disadvantages? Give reasons and include relevant examples. Write at least 250 words.',
        NULL, 0, 2400, NULL, NULL,
        NULL,
        'Assessed on task response, coherence and cohesion, lexical resource, and grammatical range and accuracy.',
        'hard', ARRAY['IELTS Writing', 'Task 2'], 25),

    ('ielts-rdg-001', 'ielts-2026-01', 'IELTS', 'reading', 'ielts-reading-tfng', 'True / False / Not Given',
        'The Indus Valley Script',
        'Do the following statements agree with the information in the passage?',
        'Thousands of short inscriptions in the Indus script survive on seals, pottery shards and copper tablets. No bilingual text has ever been found, and the language underlying the signs remains unidentified. Estimates of the number of distinct signs range from 400 to over 600.',
        0, 300,
        '["TRUE","NOT GIVEN","FALSE"]'::jsonb, NULL,
        NULL,
        'NOT GIVEN means the passage neither confirms nor contradicts the claim. Do not fill the gap with outside knowledge.',
        'medium', ARRAY['IELTS Reading', 'True/False/Not Given'], 12),

    ('ielts-spk-001', 'ielts-2026-01', 'IELTS', 'speaking', 'ielts-speaking-part2', 'Speaking Part 2 (Cue Card)',
        'Describe a Skill You Learned',
        'Describe a skill you learned that was useful. Say what it was, how you learned it, how long it took, and explain why it was useful. You have one minute to prepare and should speak for one to two minutes.',
        NULL, 60, 120, NULL, NULL,
        NULL,
        'Cover every bullet on the card. Running short is penalised more heavily than a small grammatical slip.',
        'medium', ARRAY['IELTS Speaking', 'Part 2'], 15);

-- ---------------------------------------------------------------------------
-- Mock blueprints
-- ---------------------------------------------------------------------------

INSERT INTO mocks (id, exam_version_id, exam, title, description, total_duration_minutes, is_diagnostic) VALUES
    ('mock-pte-diag-01', 'pte-2026-01', 'PTE', 'PTE Diagnostic',
        'Short diagnostic across all four skills. Sets your first score estimate.', 30, TRUE),
    ('mock-pte-full-01', 'pte-2026-01', 'PTE', 'PTE Academic Full Mock #1',
        'Full-length simulation covering Speaking & Writing, Reading and Listening.', 120, FALSE),
    ('mock-ielts-diag-01', 'ielts-2026-01', 'IELTS', 'IELTS Diagnostic',
        'Short diagnostic across all four skills. Sets your first band estimate.', 30, TRUE),
    ('mock-ielts-full-01', 'ielts-2026-01', 'IELTS', 'IELTS Academic Full Mock #1',
        'Full-length IELTS Academic simulation.', 160, FALSE);

INSERT INTO mock_sections (id, mock_id, position, name, skill, duration_minutes, question_ids) VALUES
    ('sec-pte-diag-1', 'mock-pte-diag-01', 1, 'Speaking', 'speaking', 5, ARRAY['pte-spk-001']),
    ('sec-pte-diag-2', 'mock-pte-diag-01', 2, 'Writing', 'writing', 10, ARRAY['pte-wrt-001']),
    ('sec-pte-diag-3', 'mock-pte-diag-01', 3, 'Reading', 'reading', 10, ARRAY['pte-rdg-001', 'pte-rdg-002']),
    ('sec-pte-diag-4', 'mock-pte-diag-01', 4, 'Listening', 'listening', 5, ARRAY['pte-lis-001']),

    ('sec-pte-full-1', 'mock-pte-full-01', 1, 'Speaking & Writing', 'speaking', 54, ARRAY['pte-spk-001', 'pte-wrt-001']),
    ('sec-pte-full-2', 'mock-pte-full-01', 2, 'Reading', 'reading', 32, ARRAY['pte-rdg-001', 'pte-rdg-002']),
    ('sec-pte-full-3', 'mock-pte-full-01', 3, 'Listening', 'listening', 34, ARRAY['pte-lis-001']),

    ('sec-ielts-diag-1', 'mock-ielts-diag-01', 1, 'Reading', 'reading', 15, ARRAY['ielts-rdg-001']),
    ('sec-ielts-diag-2', 'mock-ielts-diag-01', 2, 'Speaking', 'speaking', 15, ARRAY['ielts-spk-001']),

    ('sec-ielts-full-1', 'mock-ielts-full-01', 1, 'Reading', 'reading', 60, ARRAY['ielts-rdg-001']),
    ('sec-ielts-full-2', 'mock-ielts-full-01', 2, 'Writing', 'writing', 60, ARRAY['ielts-wrt-001']),
    ('sec-ielts-full-3', 'mock-ielts-full-01', 3, 'Speaking', 'speaking', 15, ARRAY['ielts-spk-001']);

-- ---------------------------------------------------------------------------
-- Daily missions
-- ---------------------------------------------------------------------------

INSERT INTO daily_missions (id, title, description, exam, skill, task_type, target_count, xp_reward) VALUES
    ('mis-speaking-2', 'Two speaking tasks', 'Record two speaking responses.',        NULL, 'speaking',  '', 2, 100),
    ('mis-writing-1',  'One written response', 'Submit one writing task for evaluation.', NULL, 'writing', '', 1, 150),
    ('mis-reading-3',  'Three reading drills', 'Complete three reading questions.',   NULL, 'reading',   '', 3, 120),
    ('mis-listening-2','Two listening drills', 'Complete two listening questions.',   NULL, 'listening', '', 2, 110);
