-- 000039: Expand listening content and fix IELTS/PTE listening subtests.
--
-- 1. Fix ielts-lis-001: Set type_name to 'Fill in the blanks' (matching IELTS syllabus),
--    and populate blanks JSONB so it renders blank inputs and is answerable.
UPDATE questions
SET type_name = 'Fill in the blanks',
    blanks = '[{"id":"1","correctAnswer":"student"},{"id":"2","correctAnswer":"800"},{"id":"3","correctAnswer":"six"}]'::jsonb
WHERE id = 'ielts-lis-001';

-- ---------------------------------------------------------------------------
-- IELTS Listening Content
-- ---------------------------------------------------------------------------

-- IELTS Subtest 1: Fill in the blanks (additional items)
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('ielts-lis-002', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'listening', 'ielts-listening-form', 'Fill in the blanks',
        'Apartment Rental Inquiry',
        'Listen to the conversation and fill in the missing information below. Write NO MORE THAN TWO WORDS AND/OR A NUMBER.',
        E'Property address: 42 Elmwood Street\nNumber of bedrooms: ____ (1)\nMonthly rent: £ ____ (2)\nAvailability date: 1st of ____ (3)',
        E'Tenant: Hello, I am inquiring about the rental apartment listed on Elmwood Street.\n'
        'Agent: Yes, that is 42 Elmwood Street. It is a lovely two-bedroom property with a newly renovated kitchen.\n'
        'Tenant: Great, and how much is the monthly rent?\n'
        'Agent: It is nine hundred and fifty pounds per month, excluding utility bills.\n'
        'Tenant: When is it available for move in?\n'
        'Agent: The current tenancy ends late this month, so it will be ready on the 1st of October.',
        0, 180,
        '["two","950","October"]'::jsonb,
        '[{"id":"1","correctAnswer":"two"},{"id":"2","correctAnswer":"950"},{"id":"3","correctAnswer":"October"}]'::jsonb,
        NULL,
        'Pay close attention to numbers and dates. Write "two", "950", and "October".',
        'easy', ARRAY['IELTS Listening', 'Fill in the blanks'], 10),

    ('ielts-lis-003', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'listening', 'ielts-listening-notes', 'Fill in the blanks',
        'Campus Fitness Club Registration',
        'Listen to the audio recording and complete the notes below. Write ONE WORD AND/OR A NUMBER for each blank.',
        E'Membership category: ____ (1)\nAccess hours: 6:00 AM to ____ (2) PM\nInduction session scheduled for: ____ (3)',
        E'Instructor: Welcome to the campus sports center! Which membership tier are you looking to sign up for?\n'
        'Student: I would like the premium membership so I can use both the gym and the Olympic pool.\n'
        'Instructor: Excellent choice. Premium members get unrestricted access from 6:00 AM until 10:00 PM every day.\n'
        'Student: Wonderful. Do I need an equipment induction before using the weights room?\n'
        'Instructor: Yes, all new members must attend an induction. We have an open slot this Thursday at 4 PM.',
        0, 180,
        '["premium","10","Thursday"]'::jsonb,
        '[{"id":"1","correctAnswer":"premium"},{"id":"2","correctAnswer":"10"},{"id":"3","correctAnswer":"Thursday"}]'::jsonb,
        NULL,
        'Write exact words spoken: "premium", "10", "Thursday".',
        'medium', ARRAY['IELTS Listening', 'Fill in the blanks'], 10);

-- IELTS Subtest 2: Choose the correct answer (MCQ)
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    options, correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('ielts-lis-004', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'listening', 'ielts-listening-mcq', 'Choose the correct answer',
        'Academic Advisory Orientation',
        'Listen to the academic adviser explaining course module selection and answer the question below.',
        NULL,
        E'Adviser: When selecting your elective modules for the second term, keep in mind that module registration closes strictly at midnight on Friday. While online submissions are preferred, students experiencing portal login difficulties should submit a physical signed form directly to the student services desk in the North Atrium before 4:00 PM.',
        0, 120,
        '[{"id":"opt-a","text":"Online submissions are accepted until Monday morning"},
          {"id":"opt-b","text":"Physical forms must be delivered to student services in the North Atrium by 4:00 PM"},
          {"id":"opt-c","text":"Late module registration carries a mandatory penalty fee"},
          {"id":"opt-d","text":"Electives cannot be changed once term starts"}]'::jsonb,
        '["opt-b"]'::jsonb,
        NULL,
        NULL,
        'The adviser states that paper forms must be delivered to student services in the North Atrium before 4:00 PM.',
        'medium', ARRAY['IELTS Listening', 'Choose the correct answer'], 10),

    ('ielts-lis-005', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'listening', 'ielts-listening-mcq', 'Choose the correct answer',
        'Marine Ecology Field Research Briefing',
        'Listen to the lead researcher discussing coastal erosion and select the correct reason for recent changes.',
        NULL,
        E'Dr. Reynolds: Over the past three decades, coastal dune retreat along the southern coastline has accelerated dramatically. While seasonal storm surges have played a minor part, the primary catalyst has been the removal of mangrove root networks during beachfront development, leaving sediment completely unprotected.',
        0, 120,
        '[{"id":"opt-a","text":"Abnormally severe seasonal storm surges"},
          {"id":"opt-b","text":"Destruction of protective mangrove root structures during coastal development"},
          {"id":"opt-c","text":"Commercial overfishing near reef borders"},
          {"id":"opt-d","text":"Industrial chemical runoff from inland agriculture"}]'::jsonb,
        '["opt-b"]'::jsonb,
        NULL,
        NULL,
        'Dr. Reynolds explicitly identifies the removal of mangrove root networks as the primary catalyst.',
        'medium', ARRAY['IELTS Listening', 'Choose the correct answer'], 10);

-- IELTS Subtest 3: Navigating according to audio
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    options, correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('ielts-lis-006', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'listening', 'ielts-listening-navigation', 'Navigating according to audio',
        'University Science Campus Navigation',
        'Listen to the guide giving directions across the university campus. Where is the New Materials Laboratory situated?',
        NULL,
        E'Guide: We are standing at the main archway facing north toward the central lawn. If you look straight ahead, you will see the Great Hall. To reach the New Materials Laboratory, turn right at the fountain, follow the paved walkway past the botanical greenhouse, and the laboratory is the modern glass building immediately behind the engineering workshops.',
        0, 120,
        '[{"id":"loc-a","text":"Opposite the Great Hall facing the central lawn"},
          {"id":"loc-b","text":"Inside the old chemistry tower adjacent to the library"},
          {"id":"loc-c","text":"Immediately behind the engineering workshops after turning right at the fountain"},
          {"id":"loc-d","text":"Adjacent to the west athletic stadium"}]'::jsonb,
        '["loc-c"]'::jsonb,
        NULL,
        NULL,
        'The guide directs visitors to turn right at the fountain and locates the laboratory behind the engineering workshops.',
        'medium', ARRAY['IELTS Listening', 'Navigating according to audio'], 10),

    ('ielts-lis-007', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'listening', 'ielts-listening-navigation', 'Navigating according to audio',
        'Metropolitan Museum Gallery Floorplan',
        'Listen to the gallery curator describing the layout of the exhibition rooms. Which gallery houses the Ancient Bronze Collection?',
        NULL,
        E'Curator: Welcome to the third floor gallery corridor. Immediately as you exit the central elevators, the Renaissance oil paintings are on your left. Directly across the hall on your right is the contemporary photography gallery. To explore our Ancient Bronze Collection, continue straight down the central corridor to the very end room overlooking the east courtyard.',
        0, 120,
        '[{"id":"room-1","text":"The first gallery on the left as you exit the central elevators"},
          {"id":"room-2","text":"Across the hall on the right in the photography gallery"},
          {"id":"room-3","text":"The room at the very end of the central corridor overlooking the east courtyard"},
          {"id":"room-4","text":"The basement annex below the main staircase"}]'::jsonb,
        '["room-3"]'::jsonb,
        NULL,
        NULL,
        'The curator notes that the Ancient Bronze Collection is located in the end room overlooking the east courtyard.',
        'hard', ARRAY['IELTS Listening', 'Navigating according to audio'], 10);

-- ---------------------------------------------------------------------------
-- PTE Listening Content
-- ---------------------------------------------------------------------------

-- PTE Subtest 1: Write from Dictation (additional items)
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-002', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'write-from-dictation', 'Write from Dictation',
        'Library Loan Policies',
        'You will hear a sentence. Type the sentence exactly as you hear it. Write what you hear, not what you expect to hear.',
        NULL,
        'The university library offers extended opening hours during final examination weeks.',
        0, 60,
        '["The university library offers extended opening hours during final examination weeks."]'::jsonb,
        NULL, NULL,
        'Spell every word accurately, including plural nouns and punctuation.',
        'easy', ARRAY['PTE Listening', 'Write from Dictation'], 10),

    ('pte-lis-003', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'write-from-dictation', 'Write from Dictation',
        'Engineering Research Standards',
        'You will hear a sentence. Type the sentence exactly as you hear it. Write what you hear, not what you expect to hear.',
        NULL,
        'Scientific research requires rigorous data validation before any publication.',
        0, 60,
        '["Scientific research requires rigorous data validation before any publication."]'::jsonb,
        NULL, NULL,
        'Capture words precisely: "Scientific", "rigorous", "validation", "publication".',
        'medium', ARRAY['PTE Listening', 'Write from Dictation'], 10);

-- PTE Subtest 2: Summarize Spoken Text
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-004', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text',
        'Renewable Energy Integration and Grid Reliability',
        'You will hear a short lecture. Write a summary for a fellow student who was not present. You should write 50-70 words.',
        NULL,
        E'Lecturer: As national power systems shift away from fossil fuels, the intermittent nature of solar and wind generation presents novel challenges for grid stability. Traditional thermal power stations provided base-load inertia, maintaining constant frequency. Modern electrical engineers are addressing this volatility through utility-scale battery storage, decentralised microgrids, and intelligent predictive algorithms that modulate energy consumption in real time.',
        0, 600,
        '["renewable","grid","storage","intermittent","battery","stability","energy"]'::jsonb,
        NULL,
        'The transition toward renewable energy introduces intermittency and grid instability previously mitigated by thermal power. To address these volatility issues, electrical engineers are deploying utility-scale battery storage, decentralised microgrids, and predictive algorithms to regulate demand and preserve grid stability.',
        'Target 50-70 words. Mention the core challenge (intermittency of renewables) and modern engineering solutions (battery storage, microgrids, predictive algorithms).',
        'medium', ARRAY['PTE Listening', 'Summarize Spoken Text'], 12),

    ('pte-lis-005', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text',
        'Neuroplasticity and Adult Learning',
        'You will hear a short lecture. Write a summary for a fellow student who was not present. You should write 50-70 words.',
        NULL,
        E'Prof. Vance: Historically, neuroscientists maintained that the adult human brain was structurally immutable after adolescence. Groundbreaking synaptic research has completely revised this paradigm. We now recognise that neuroplasticity persists throughout life, allowing neural circuits to reorganise and forge novel synaptic connections in response to continuous cognitive stimulation, deliberate practice, and new language acquisition.',
        0, 600,
        '["neuroplasticity","brain","synaptic","learning","cognitive","connections","adult"]'::jsonb,
        NULL,
        'Although neuroscientists once believed the adult brain could not change structurally after adolescence, modern research confirms that neuroplasticity continues across lifespan. The brain constantly reorganises neural circuits and creates new synaptic connections when challenged with sustained cognitive activities, deliberate practice, and new language learning.',
        'Maintain strict 50-70 word count. Summarize the shift from fixed brain structure to lifelong neuroplasticity stimulated by cognitive practice.',
        'hard', ARRAY['PTE Listening', 'Summarize Spoken Text'], 12);

-- PTE Subtest 3: Multiple Choice, Multiple Answers
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    options, correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-006', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'multiple-choice-multiple', 'Multiple Choice, Multiple Answers',
        'Global Biodiversity Threats',
        'Listen to the recording and answer the question by selecting all the correct responses. More than one response is correct.',
        NULL,
        E'Speaker: Contemporary conservation biology identifies habitat fragmentation as the preeminent cause of vertebrate species decline. When contiguous forest corridors are severed by transport infrastructure or commercial agriculture, isolated animal populations experience severe genetic bottlenecks. Furthermore, the introduction of invasive predatory species exacerbates native population collapse by outcompeting endemic species for vital resources.',
        0, 180,
        '[{"id":"mc-a","text":"Habitat fragmentation creates genetic bottlenecks in isolated populations"},
          {"id":"mc-b","text":"Invasive predatory species outcompete endemic animals for resources"},
          {"id":"mc-c","text":"Captive breeding initiatives have replaced natural wilderness reserves"},
          {"id":"mc-d","text":"Endemic species have completely adapted to highway noise corridors"}]'::jsonb,
        '["mc-a","mc-b"]'::jsonb,
        NULL,
        NULL,
        'The speaker explicitly highlights both genetic bottlenecks from fragmented habitats (A) and invasive species outcompeting native animals (B).',
        'medium', ARRAY['PTE Listening', 'Multiple Choice, Multiple Answers'], 10),

    ('pte-lis-007', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'multiple-choice-multiple', 'Multiple Choice, Multiple Answers',
        'Urban Microclimates and Architecture',
        'Listen to the recording and select all correct statements regarding mitigation of urban heat islands.',
        NULL,
        E'Speaker: Urban heat islands significantly intensify residential cooling demands during summer heatwaves. Architects and municipal engineers utilize two proven strategies: installing vegetative green roofs that absorb solar radiation through evapotranspiration, and replacing asphalt with high-albedo permeable pavements that reflect sunlight.',
        0, 180,
        '[{"id":"mc-1","text":"Installing vegetative green roofs enhances cooling via evapotranspiration"},
          {"id":"mc-2","text":"Permeable pavements with high albedo reflect incident solar radiation"},
          {"id":"mc-3","text":"Dark asphalt surfaces are mandatory for subterranean insulation"},
          {"id":"mc-4","text":"Urban heat islands eliminate winter heating requirements entirely"}]'::jsonb,
        '["mc-1","mc-2"]'::jsonb,
        NULL,
        NULL,
        'The speaker states that green roofs (1) and high-albedo permeable pavements (2) mitigate urban heat.',
        'medium', ARRAY['PTE Listening', 'Multiple Choice, Multiple Answers'], 10);

-- PTE Subtest 4: Fill in the Blanks
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-008', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'fill-in-the-blanks', 'Fill in the Blanks',
        'Atmospheric Carbon Measurement',
        'You will hear a recording. Type the missing words in each blank according to what you hear.',
        E'Accurate measurement of greenhouse gases requires sophisticated atmospheric monitoring [[b1]]. Scientists deploy spectrometers aboard orbital satellites to analyze solar [[b2]] reflected through the atmosphere. These instruments detect trace concentrations with unprecedented [[b3]].',
        E'Accurate measurement of greenhouse gases requires sophisticated atmospheric monitoring stations. Scientists deploy spectrometers aboard orbital satellites to analyze solar radiation reflected through the atmosphere. These instruments detect trace concentrations with unprecedented precision.',
        0, 120,
        '["stations","radiation","precision"]'::jsonb,
        '[{"id":"b1","correctAnswer":"stations"},
          {"id":"b2","correctAnswer":"radiation"},
          {"id":"b3","correctAnswer":"precision"}]'::jsonb,
        NULL,
        'Fill in the exact spoken words: "stations", "radiation", and "precision".',
        'medium', ARRAY['PTE Listening', 'Fill in the Blanks'], 10),

    ('pte-lis-009', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'fill-in-the-blanks', 'Fill in the Blanks',
        'Deep-Sea Hydrothermal Vent Ecosystems',
        'You will hear a recording. Type the missing words in each blank according to what you hear.',
        E'Deep-sea hydrothermal vents foster extraordinary ecosystems independent of [[b1]]. Chemoautotrophic bacteria oxidize sulfur compounds to generate [[b2]], supporting specialized tubeworms and crustaceans in complete [[b3]].',
        E'Deep-sea hydrothermal vents foster extraordinary ecosystems independent of sunlight. Chemoautotrophic bacteria oxidize sulfur compounds to generate nutrients, supporting specialized tubeworms and crustaceans in complete darkness.',
        0, 120,
        '["sunlight","nutrients","darkness"]'::jsonb,
        '[{"id":"b1","correctAnswer":"sunlight"},
          {"id":"b2","correctAnswer":"nutrients"},
          {"id":"b3","correctAnswer":"darkness"}]'::jsonb,
        NULL,
        'Listen for exact words: "sunlight", "nutrients", "darkness".',
        'medium', ARRAY['PTE Listening', 'Fill in the Blanks'], 10);

-- PTE Subtest 5: Highlight Correct Summary
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    options, correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-010', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'highlight-correct-summary', 'Highlight Correct Summary',
        'Economic Impact of Automated Manufacturing',
        'You will hear a recording. Click on the paragraph that best relates to the recording.',
        NULL,
        E'Speaker: While widespread adoption of robotic manufacturing has displaced manual assembly line labour, economic data shows an aggregate expansion in skilled technical employment. Factories require automation engineers, software maintenance personnel, and quality control analysts, leading to higher median industrial salaries despite reduced headcounts on traditional production floors.',
        0, 120,
        '[{"id":"sum-1","text":"Robotic manufacturing has completely eliminated industrial employment opportunities across all demographic groups." },
          {"id":"sum-2","text":"Although automation reduces manual assembly roles, it stimulates an overall rise in higher-paid technical and engineering positions."},
          {"id":"sum-3","text":"Automation has failed to enhance factory productivity because of excessive maintenance expenses."},
          {"id":"sum-4","text":"Manufacturing plants are actively returning to manual labour due to software unreliability."}]'::jsonb,
        '["sum-2"]'::jsonb,
        NULL,
        NULL,
        'Summary 2 accurately captures the transition from manual assembly jobs to higher-paid technical roles.',
        'medium', ARRAY['PTE Listening', 'Highlight Correct Summary'], 10),

    ('pte-lis-011', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'highlight-correct-summary', 'Highlight Correct Summary',
        'Behavioral Economics and Default Options',
        'You will hear a recording. Click on the paragraph that best relates to the recording.',
        NULL,
        E'Speaker: In behavioral economics, the concept of choice architecture demonstrates that people overwhelmingly adhere to preset default choices. When employers automatically enroll employees in retirement savings plans with opt-out provisions, participation rates exceed ninety percent, whereas voluntary opt-in structures rarely surpass forty percent.',
        0, 120,
        '[{"id":"sum-a","text":"Retirement savings plans are generally unpopular regardless of how enrollment is organized."},
          {"id":"sum-b","text":"Default enrollment settings dramatically increase retirement savings participation compared to voluntary opt-in systems."},
          {"id":"sum-c","text":"Employees prefer manual paperwork over digital automatic savings arrangements."},
          {"id":"sum-d","text":"Opt-out provisions discourage individuals from investing in long-term pensions."}]'::jsonb,
        '["sum-b"]'::jsonb,
        NULL,
        NULL,
        'Summary B accurately reflects the huge disparity in participation when default automatic enrollment is employed.',
        'medium', ARRAY['PTE Listening', 'Highlight Correct Summary'], 10);

-- PTE Subtest 6: Select Missing Word
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    options, correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-012', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'select-missing-word', 'Select Missing Word',
        'Principles of Cognitive Psychology',
        'You will hear a recording about human memory. At the end of the recording the last word or group of words has been replaced by a beep. Select the correct option to complete the recording.',
        NULL,
        E'Speaker: Working memory has an inherently limited capacity, typically holding between four and seven chunks of discrete information simultaneously. When learners are overwhelmed with unnecessary visual clutter or disjointed instructions, they experience cognitive overload, which severely impedes their ability to [beep].',
        0, 90,
        '[{"id":"mw-1","text":"retain new knowledge"},
          {"id":"mw-2","text":"breathe normally"},
          {"id":"mw-3","text":"fall asleep immediately"},
          {"id":"mw-4","text":"purchase expensive software"}]'::jsonb,
        '["mw-1"]'::jsonb,
        NULL,
        NULL,
        'In cognitive psychology, overload prevents students from retaining new knowledge or learning.',
        'medium', ARRAY['PTE Listening', 'Select Missing Word'], 10),

    ('pte-lis-013', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'select-missing-word', 'Select Missing Word',
        'Principles of Aeronautical Engineering',
        'You will hear a recording about aerodynamics. At the end of the recording the last word has been replaced by a beep. Select the option that best completes the sentence.',
        NULL,
        E'Speaker: An aircraft wing generates aerodynamic lift when air moves faster over its curved upper surface than beneath its flatter lower surface. According to Bernoulli principle, this difference in velocity creates a pressure differential that pulls the aircraft upward into the [beep].',
        0, 90,
        '[{"id":"mw-a","text":"sky"},
          {"id":"mw-b","text":"ground"},
          {"id":"mw-c","text":"ocean"},
          {"id":"mw-d","text":"hangar"}]'::jsonb,
        '["mw-a"]'::jsonb,
        NULL,
        NULL,
        'Aerodynamic lift propels the aircraft upward into the sky.',
        'easy', ARRAY['PTE Listening', 'Select Missing Word'], 10);

-- ---------------------------------------------------------------------------
-- Update Mock Exam Sections to include new listening questions
-- ---------------------------------------------------------------------------

UPDATE mock_sections
SET question_ids = ARRAY[
    'ielts-lis-001', 'ielts-lis-002', 'ielts-lis-003',
    'ielts-lis-004', 'ielts-lis-005', 'ielts-lis-006', 'ielts-lis-007'
]
WHERE id = 'sec-ielts-full-4';

UPDATE mock_sections
SET question_ids = ARRAY['ielts-lis-001', 'ielts-lis-002', 'ielts-lis-004']
WHERE id = 'sec-ielts-diag-4';

UPDATE mock_sections
SET question_ids = ARRAY[
    'pte-lis-001', 'pte-lis-002', 'pte-lis-003', 'pte-lis-004',
    'pte-lis-006', 'pte-lis-008', 'pte-lis-010', 'pte-lis-012'
]
WHERE id = 'sec-pte-full-3';

UPDATE mock_sections
SET question_ids = ARRAY['pte-lis-001', 'pte-lis-002', 'pte-lis-008']
WHERE id = 'sec-pte-diag-4';
