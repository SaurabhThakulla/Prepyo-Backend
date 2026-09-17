-- 000047: Add five PTE listening questions for each of the eight subtasks.
--
-- Following 000039 and 000046: canonical type IDs, transcripts for browser TTS
-- (no audio_url), question IDs pte-lis-020..059. Mock sections are untouched.
--
-- Subtask 1: Write from Dictation. The transcript is the answer, so
-- correct_answers[0] must equal audio_transcript and blanks stay empty.
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-020', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'write-from-dictation', 'Write from Dictation',
        'Examination Period Opening Hours',
        'You will hear a sentence. Type the sentence exactly as you hear it. Write what you hear, not what you expect to hear.',
        NULL,
        'The university library extends its opening hours during the final examination period.',
        0, 60,
        '["The university library extends its opening hours during the final examination period."]'::jsonb,
        NULL, NULL,
        'Type every word in order, including the small words: its, the, during.',
        'easy', ARRAY['PTE Listening', 'Write from Dictation'], 10),

    ('pte-lis-021', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'write-from-dictation', 'Write from Dictation',
        'Warm Winter Attribution',
        'You will hear a sentence. Type the sentence exactly as you hear it. Write what you hear, not what you expect to hear.',
        NULL,
        'Climate researchers attribute this unusually warm winter to shifting ocean currents.',
        0, 60,
        '["Climate researchers attribute this unusually warm winter to shifting ocean currents."]'::jsonb,
        NULL, NULL,
        'Capture the precise wording: attribute, unusually, shifting.',
        'medium', ARRAY['PTE Listening', 'Write from Dictation'], 10),

    ('pte-lis-022', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'write-from-dictation', 'Write from Dictation',
        'Laboratory Deadlines',
        'You will hear a sentence. Type the sentence exactly as you hear it. Write what you hear, not what you expect to hear.',
        NULL,
        'All laboratory reports must be submitted before the end of the semester.',
        0, 60,
        '["All laboratory reports must be submitted before the end of the semester."]'::jsonb,
        NULL, NULL,
        'Every word counts: all, must, before, semester.',
        'easy', ARRAY['PTE Listening', 'Write from Dictation'], 10),

    ('pte-lis-023', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'write-from-dictation', 'Write from Dictation',
        'Building Design',
        'You will hear a sentence. Type the sentence exactly as you hear it. Write what you hear, not what you expect to hear.',
        NULL,
        'Modern architecture often combines sustainable materials with traditional construction methods.',
        0, 60,
        '["Modern architecture often combines sustainable materials with traditional construction methods."]'::jsonb,
        NULL, NULL,
        'Listen for the pairing: sustainable materials with traditional construction methods.',
        'medium', ARRAY['PTE Listening', 'Write from Dictation'], 10),

    ('pte-lis-024', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'write-from-dictation', 'Write from Dictation',
        'Missed Deadlines',
        'You will hear a sentence. Type the sentence exactly as you hear it. Write what you hear, not what you expect to hear.',
        NULL,
        'Students who miss the deadline should contact their tutor as soon as possible.',
        0, 60,
        '["Students who miss the deadline should contact their tutor as soon as possible."]'::jsonb,
        NULL, NULL,
        'The sentence ends with the phrase as soon as possible.',
        'easy', ARRAY['PTE Listening', 'Write from Dictation'], 10);

-- Subtask 2: Summarize Spoken Text. correct_answers hold key concepts for the
-- approximate grader; model_answer is the review reference. 50-70 words scores best.
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, audio_transcript, prep_time_seconds, time_limit_seconds,
    correct_answers, blanks, model_answer, explanation, difficulty, tags, points
) VALUES
    ('pte-lis-025', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text',
        'Sleep and Memory Consolidation',
        'You will hear a short lecture. Write a summary for a fellow student who was not present. You should write 50-70 words.',
        NULL,
        E'During deep sleep the brain replays the experiences of the day, transferring fragile memories from the hippocampus into long-term storage. Researchers call this consolidation, and it explains why a good night of sleep improves recall before an examination. Sleep deprivation narrows attention, impairs new learning, and strengthens emotional reactions. Students who sacrifice sleep to study often undermine the very memory they are trying to build.',
        0, 600,
        '["sleep","memory","consolidation","hippocampus","recall"]'::jsonb,
        NULL,
        'During deep sleep the brain replays recent experiences and moves fragile memories from the hippocampus into long-term storage through consolidation, which explains why sleeping well improves recall before examinations. Sleep deprivation narrows attention, impairs new learning and strengthens emotional reactions, so students who sacrifice sleep to study undermine the memory they are trying to build.',
        'Cover the consolidation process, the role of the hippocampus, and the effect of sleep deprivation on learning.',
        'medium', ARRAY['PTE Listening', 'Summarize Spoken Text'], 12),

    ('pte-lis-026', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text',
        'Urban Green Space',
        'You will hear a short lecture. Write a summary for a fellow student who was not present. You should write 50-70 words.',
        NULL,
        E'Urban vegetation does far more than decorate a city. Trees cool streets by several degrees during heatwaves, absorb rainwater that would otherwise overwhelm drains, and filter polluted air. People who live near parks walk more and report lower stress. Planners who treat green space as essential infrastructure find these benefits reach every neighbourhood, while cities that reserve parks for wealthy districts widen health inequalities.',
        0, 600,
        '["green","flood","air","stress","neighbourhood"]'::jsonb,
        NULL,
        'Urban vegetation cools streets during heatwaves, absorbs rainwater that would otherwise flood drains, and filters polluted air, while living near parks encourages walking and reduces stress. The speaker argues that green space planned as public infrastructure benefits every neighbourhood, whereas restricting parks to wealthy districts deepens existing health inequalities across the city.',
        'Mention the three environmental benefits and the equity argument about who lives near parks.',
        'medium', ARRAY['PTE Listening', 'Summarize Spoken Text'], 12),

    ('pte-lis-027', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text',
        'The Office After Remote Work',
        'Listen and write a summary of 50-70 words.', NULL,
        'Working from home changed more than the commute. Many employees report deeper concentration on individual tasks, and companies save on office space. Yet new colleagues learn less without people nearby, and video meetings consume the informal hours that once built trust. Organisations that succeed now treat the office as a place for collaboration and mentoring, while focused individual work happens wherever it suits the employee.',
        0, 600, '["remote","concentration","mentoring","trust","collaboration"]'::jsonb, NULL,
        'Remote work increases concentration on individual tasks and reduces office costs, but new employees learn less without colleagues nearby and video meetings erode the informal time that builds trust. Successful organisations therefore use the office mainly for collaboration and mentoring, while allowing focused work to happen wherever employees work best.',
        'Balance concentration and cost benefits against slower learning and weaker trust; explain the new role of the office.',
        'medium', ARRAY['PTE Listening', 'Summarize Spoken Text'], 12),
    ('pte-lis-028', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text',
        'Declining Pollinators',
        'Listen and write a summary of 50-70 words.', NULL,
        'Bees pollinate roughly a third of the crops that feed the world, from apples to almonds. Yet beekeepers report unusually high colony losses each winter. Scientists point to three pressures acting together: pesticides that interfere with bee navigation, parasites carried into hives, and farming that removes the wildflowers bees depend on between harvests. Protecting pollinators therefore means changing how farmland is planted, not only restricting chemicals.',
        0, 600, '["pollinate","crops","pesticides","parasites","wildflowers"]'::jsonb, NULL,
        'Bees pollinate about one third of global food crops, but colony losses are rising. Pesticides disrupt bee navigation, parasites spread through hives, and intensive farming removes the wildflowers bees rely on between harvests. Because these pressures combine, the speaker argues that protecting pollinators requires changing farmland planting as well as limiting chemical use.',
        'Name the three combined pressures and explain why protecting bees also requires changes to farmland planting.',
        'hard', ARRAY['PTE Listening', 'Summarize Spoken Text'], 12),
    ('pte-lis-029', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'summarize-spoken-text', 'Summarize Spoken Text',
        'Renewables and Grid Stability',
        'Listen and write a summary of 50-70 words.', NULL,
        'Solar and wind power are now the cheapest sources of new electricity in most countries, but the sun sets and the wind pauses. Keeping supplies stable therefore requires storage, from household batteries to pumped hydroelectric reservoirs, together with transmission lines that move power between regions. Countries that invested early in both storage and interconnection have kept their grids reliable while cutting emissions sharply.',
        0, 600, '["solar","wind","storage","transmission","emissions"]'::jsonb, NULL,
        'Solar and wind power are now the cheapest new electricity sources in most countries, but their variability requires storage, ranging from household batteries to pumped hydro, and stronger transmission links between regions. Nations that invested early in both storage and regional interconnection have maintained reliable electricity grids while sharply reducing emissions.',
        'Explain the low cost and variability of renewables, and how storage and transmission maintain reliable supplies.',
        'medium', ARRAY['PTE Listening', 'Summarize Spoken Text'], 12);

-- Subtask 3: MCQ Multiple Answer.
INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
    title, prompt, audio_transcript, time_limit_seconds, options, correct_answers, explanation, tags)
SELECT id, 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'pte-listening-mcma', 'MCQ Multiple Answer',
    title, prompt, script, 120, options::jsonb, answers::jsonb, explanation, ARRAY['PTE Listening', 'MCQ Multiple Answer']
FROM (VALUES
    ('pte-lis-030', 'Museum Evening Trial', 'Which TWO outcomes followed the trial of evening opening hours?',
     'The museum opened until nine on Thursdays for a six week trial. The total number of visitors barely changed, but more visitors arrived after work. Staff also recorded longer visits to the permanent collection. Sales in the shop did not rise, and the museum did not recruit additional guides. The director will review staffing costs before making the arrangement permanent.',
     '[{"id":"A","text":"More people visited after work."},{"id":"B","text":"Shop sales increased."},{"id":"C","text":"Visits to the permanent collection became longer."},{"id":"D","text":"Additional guides were hired."}]', '["A","C"]',
     'The speaker reports more after-work visits and longer stays in the permanent collection, not increased sales or recruitment.'),
    ('pte-lis-031', 'Field Course Preparations', 'Which TWO things must students do before the field course?',
     'Before our geology field course, everyone must upload an emergency contact form and attend the equipment briefing. You do not need to buy a hammer because the department lends them out. The reading list is optional preparation, although it will help you recognise the rock formations. Travel tickets will be issued at the briefing rather than booked individually.',
     '[{"id":"A","text":"Buy a geological hammer."},{"id":"B","text":"Submit emergency contact details."},{"id":"C","text":"Book individual travel tickets."},{"id":"D","text":"Attend the equipment briefing."}]', '["B","D"]',
     'The form and briefing are compulsory. Equipment and tickets are supplied; reading is optional.'),
    ('pte-lis-032', 'Repair Cafe Benefits', 'Which TWO benefits does the speaker attribute to the repair cafe?',
     'Our repair cafe pairs volunteers with residents who bring broken household items. Its main achievement is keeping usable goods out of landfill. Residents also learn practical repair skills by working alongside volunteers. Repairs are not guaranteed, and replacement parts sometimes cost money. Although a few local shops donate tools, the project has not created any paid jobs.',
     '[{"id":"A","text":"It guarantees every repair."},{"id":"B","text":"It reduces waste sent to landfill."},{"id":"C","text":"It creates paid jobs."},{"id":"D","text":"It teaches practical skills."}]', '["B","D"]',
     'Waste reduction and learning repair skills are explicit benefits. Repairs are not guaranteed and there are no paid jobs.'),
    ('pte-lis-033', 'Choosing Survey Participants', 'Which TWO changes does the lecturer recommend?',
     'Your pilot survey recruited only students leaving the sports centre at lunchtime. That sample cannot represent the entire campus. For the next round, recruit at several locations and include evening sessions. Keep the wording of the questions unchanged so we can compare responses. Increasing the number of questions would make participation harder, and paying only athletes would reinforce the original bias.',
     '[{"id":"A","text":"Recruit at multiple campus locations."},{"id":"B","text":"Add more survey questions."},{"id":"C","text":"Collect responses in the evening too."},{"id":"D","text":"Pay only athletes to participate."}]', '["A","C"]',
     'Broader locations and times reduce the sampling bias without changing the survey wording.'),
    ('pte-lis-034', 'Coastal Dune Restoration', 'Which TWO measures are included in the restoration plan?',
     'The restoration plan will protect dunes by planting native grasses and directing visitors onto raised walkways. Grass roots bind loose sand, while the walkways prevent trampling. The beach will remain open throughout the project. Engineers rejected a concrete wall because it could worsen erosion nearby. Imported sand may be considered later, but it is not part of the current plan.',
     '[{"id":"A","text":"Close the beach permanently."},{"id":"B","text":"Plant native grasses."},{"id":"C","text":"Build a concrete sea wall."},{"id":"D","text":"Install raised visitor walkways."}]', '["B","D"]',
     'Native planting and raised walkways are approved. Closure and a concrete wall are rejected; imported sand is only a future possibility.')
) AS seed(id, title, prompt, script, options, answers, explanation);

-- Subtask 4: Listening Fill in the Blanks. Replacing each marker with its
-- answer reconstructs the spoken script exactly.
INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
    title, prompt, context_passage, audio_transcript, time_limit_seconds, blanks, explanation, tags)
SELECT id, 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'pte-listening-fib', 'Listening Fill in the Blanks',
    title, 'Listen and type ONE WORD in each blank.', passage, script, 120, blanks::jsonb,
    explanation, ARRAY['PTE Listening', 'Listening Fill in the Blanks']
FROM (VALUES
    ('pte-lis-035', 'Preserving Historical Maps',
     'Old maps reveal how communities understood their [[b1]]. Before a map can be displayed, conservators examine the paper for signs of [[b2]]. They then use carefully controlled [[b3]] to prevent further damage. Digital copies allow researchers to study fragile originals without repeated handling.',
     'Old maps reveal how communities understood their surroundings. Before a map can be displayed, conservators examine the paper for signs of decay. They then use carefully controlled humidity to prevent further damage. Digital copies allow researchers to study fragile originals without repeated handling.',
     '[{"id":"b1","correctAnswer":"surroundings"},{"id":"b2","correctAnswer":"decay"},{"id":"b3","correctAnswer":"humidity"}]',
     'The missing words are surroundings, decay and humidity. Keep the spelling and plural ending of surroundings.'),
    ('pte-lis-036', 'Community Astronomy',
     'The observatory invites local residents to an evening of [[b1]]. Volunteers explain how to adjust each telescope before visitors look at distant [[b2]]. If clouds cover the sky, the session moves indoors for a discussion of planetary [[b3]]. Advance booking is essential because places are limited.',
     'The observatory invites local residents to an evening of astronomy. Volunteers explain how to adjust each telescope before visitors look at distant galaxies. If clouds cover the sky, the session moves indoors for a discussion of planetary exploration. Advance booking is essential because places are limited.',
     '[{"id":"b1","correctAnswer":"astronomy"},{"id":"b2","correctAnswer":"galaxies"},{"id":"b3","correctAnswer":"exploration"}]',
     'Listen for astronomy, galaxies and exploration. Galaxies is plural.'),
    ('pte-lis-037', 'Public Transport Data',
     'Transport planners measure passenger [[b1]] throughout the day rather than relying on ticket sales alone. This helps them identify routes with insufficient [[b2]]. They can then adjust service [[b3]] to reduce crowding without sending empty buses through quiet neighbourhoods.',
     'Transport planners measure passenger demand throughout the day rather than relying on ticket sales alone. This helps them identify routes with insufficient capacity. They can then adjust service frequency to reduce crowding without sending empty buses through quiet neighbourhoods.',
     '[{"id":"b1","correctAnswer":"demand"},{"id":"b2","correctAnswer":"capacity"},{"id":"b3","correctAnswer":"frequency"}]',
     'Demand describes passenger need, capacity describes available space, and frequency describes how often services run.'),
    ('pte-lis-038', 'Learning Through Retrieval',
     'Reading a chapter repeatedly can create an illusion of [[b1]]. A more effective approach is to close the book and practise [[b2]] of the main ideas. Checking the answers afterwards provides useful [[b3]] and reveals which topics need further study.',
     'Reading a chapter repeatedly can create an illusion of understanding. A more effective approach is to close the book and practise retrieval of the main ideas. Checking the answers afterwards provides useful feedback and reveals which topics need further study.',
     '[{"id":"b1","correctAnswer":"understanding"},{"id":"b2","correctAnswer":"retrieval"},{"id":"b3","correctAnswer":"feedback"}]',
     'The speaker contrasts an illusion of understanding with retrieval practice supported by feedback.'),
    ('pte-lis-039', 'Wetland Monitoring',
     'A healthy wetland supports remarkable [[b1]]. Researchers regularly measure water quality and record changes in seasonal [[b2]]. These observations help distinguish natural variation from damage caused by nearby [[b3]]. Long records are especially valuable because a single visit may give a misleading impression.',
     'A healthy wetland supports remarkable biodiversity. Researchers regularly measure water quality and record changes in seasonal flooding. These observations help distinguish natural variation from damage caused by nearby construction. Long records are especially valuable because a single visit may give a misleading impression.',
     '[{"id":"b1","correctAnswer":"biodiversity"},{"id":"b2","correctAnswer":"flooding"},{"id":"b3","correctAnswer":"construction"}]',
     'Write biodiversity, flooding and construction. Each blank takes one word from the recording.')
) AS seed(id, title, passage, script, blanks, explanation);

-- Subtask 5: Highlight Correct Summary.
INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
    title, prompt, audio_transcript, time_limit_seconds, options, correct_answers, explanation, tags)
SELECT id, 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'pte-highlight-correct-summary', 'Highlight Correct Summary',
    title, 'Choose the paragraph that best summarises the recording.', script, 120,
    options::jsonb, answers::jsonb, explanation, ARRAY['PTE Listening', 'Highlight Correct Summary']
FROM (VALUES
    ('pte-lis-040', 'Libraries Beyond Books',
     'Some people predict that public libraries will disappear as reading moves online. Yet lending books is only part of their work. Libraries offer internet access, quiet study spaces and help with digital services. These functions are particularly important for residents without reliable technology at home. The challenge is therefore to adapt library funding to this broader role, not simply to count fewer book loans as evidence of declining value.',
     '[{"id":"A","text":"Falling book loans show that libraries no longer offer useful services and should close."},{"id":"B","text":"Libraries provide access to technology and study space as well as books, so funding should reflect their wider public role."},{"id":"C","text":"Libraries should stop lending books and provide only commercial internet services."},{"id":"D","text":"Digital services have made libraries unnecessary for residents without home internet."}]', '["B"]',
     'The central argument is that book loans alone do not capture the wider value of libraries.'),
    ('pte-lis-041', 'Measuring Food Waste',
     'A restaurant began weighing food discarded after every service. Staff expected most waste to come from customer plates, but preparation accounted for the largest share. By adjusting purchasing quantities and using vegetable trimmings in stock, the kitchen reduced waste without shrinking portions. The exercise shows why measuring where waste occurs can be more useful than immediately asking customers to eat less.',
     '[{"id":"A","text":"Customers caused most waste, so the restaurant reduced portion sizes."},{"id":"B","text":"Weighing waste increased purchasing costs and forced the restaurant to close."},{"id":"C","text":"The restaurant eliminated all waste by asking customers to finish their meals."},{"id":"D","text":"Measurement revealed that preparation produced most waste; changes to purchasing and ingredient use reduced it without smaller portions."}]', '["D"]',
     'Preparation, not customer plates, was the main source. The solution changed kitchen practices rather than portions.'),
    ('pte-lis-042', 'Citizen Weather Records',
     'Weather observations collected by volunteers can fill gaps between official stations. However, a thermometer beside a sunny wall may record a different temperature from one in open shade. Scientists therefore provide instructions about equipment placement and check unusual readings before using the data. Volunteer records are valuable, but their usefulness depends on consistent methods rather than the number of entries alone.',
     '[{"id":"A","text":"Volunteer observations can expand weather coverage, provided consistent collection methods and checks make the data reliable."},{"id":"B","text":"Any large collection of volunteer readings is reliable regardless of equipment placement."},{"id":"C","text":"Official stations should be replaced because volunteer equipment is always more accurate."},{"id":"D","text":"Volunteer weather records cannot contribute to research because temperatures vary."}]', '["A"]',
     'The speaker supports volunteer data with methodological safeguards, not unconditional acceptance or rejection.'),
    ('pte-lis-043', 'Restoring an Old Theatre',
     'Restoring a historic theatre involves more than returning its decoration to an earlier style. Modern audiences need accessible entrances, safe wiring and comfortable seating. Architects must introduce those features without erasing the building features that give it historical importance. A successful restoration balances preservation with practical use, because a beautiful theatre that cannot welcome audiences is unlikely to survive financially.',
     '[{"id":"A","text":"Historic theatres should remain unchanged even when audiences cannot enter safely."},{"id":"B","text":"Financial survival requires replacing every historical feature with modern decoration."},{"id":"C","text":"Restoration should preserve important historical features while adding the accessibility and safety needed for continued use."},{"id":"D","text":"Restoration is mainly about reproducing decoration and has little connection to practical use."}]', '["C"]',
     'The argument balances heritage with access, safety and financial viability.'),
    ('pte-lis-044', 'Flexible University Assessment',
     'Offering a choice between an essay and a presentation can help students demonstrate learning in different ways. But flexibility does not mean lowering standards. Both formats must assess the same learning outcomes through clear criteria. Teachers also need to explain those criteria in advance so that students choose a format for its suitability, rather than assuming one option will be easier.',
     '[{"id":"A","text":"Presentations should replace essays because they are always easier to grade."},{"id":"B","text":"Assessment choice can support different learners if all formats use clear criteria to assess the same outcomes."},{"id":"C","text":"Flexible assessment requires teachers to abandon shared academic standards."},{"id":"D","text":"Students should choose formats before teachers decide what learning to assess."}]', '["B"]',
     'Choice is supported only alongside equivalent outcomes and transparent criteria.')
) AS seed(id, title, script, options, answers, explanation);

-- Subtask 6: MCQ Single Answer.
INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
    title, prompt, audio_transcript, time_limit_seconds, options, correct_answers, explanation, tags)
SELECT id, 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'pte-listening-mcsa', 'MCQ Single Answer',
    title, prompt, script, 90, options::jsonb, answers::jsonb, explanation, ARRAY['PTE Listening', 'MCQ Single Answer']
FROM (VALUES
    ('pte-lis-045', 'Seminar Room Change', 'Why has the seminar moved to another room? Select ONE answer.',
     'Tomorrow we will meet in room twelve rather than the lecture theatre. Several students asked whether the change was caused by building repairs, but those finished last week. The seminar includes a discussion in small groups, and room twelve has movable tables. The projector there is older, so please bring printed copies of your notes.',
     '[{"id":"A","text":"Repairs are still taking place."},{"id":"B","text":"Its tables suit small group discussions."},{"id":"C","text":"It has a newer projector."},{"id":"D","text":"The class has doubled in size."}]', '["B"]',
     'Movable tables support the planned group discussion. Repairs are finished and the projector is older.'),
    ('pte-lis-046', 'Interpreting a Trial Result', 'What is the researcher mainly warning against? Select ONE answer.',
     'The new tutoring programme was followed by higher examination scores at one school. That is encouraging, but it does not establish that the programme caused the improvement. The school also introduced longer study periods during the same term. We need a comparison group before attributing the change to tutoring alone.',
     '[{"id":"A","text":"Collecting examination scores."},{"id":"B","text":"Extending any study periods."},{"id":"C","text":"Using comparison groups."},{"id":"D","text":"Assuming an improvement proves causation."}]', '["D"]',
     'Another change occurred at the same time, so the improvement cannot yet be attributed to tutoring.'),
    ('pte-lis-047', 'A Delayed Exhibition', 'What caused the exhibition opening to be postponed? Select ONE answer.',
     'The exhibition labels were printed on schedule, and the lighting installation passed its inspection yesterday. However, the loan agreement requires the paintings to travel with a specialist courier. The courier is unavailable until next Monday, so the opening will move to Wednesday. All tickets already purchased will remain valid.',
     '[{"id":"A","text":"The required specialist courier is unavailable."},{"id":"B","text":"The labels contain printing errors."},{"id":"C","text":"The lighting failed inspection."},{"id":"D","text":"Too few tickets were sold."}]', '["A"]',
     'The transport requirement and courier availability delay the paintings; labels and lighting are ready.'),
    ('pte-lis-048', 'A Historian on Diaries', 'What does the historian say about diaries? Select ONE answer.',
     'Diaries give us details that official records often omit, such as daily routines and private anxieties. But a diary is not a transparent account of everything that happened. Writers select what to record and may imagine a future reader. I use diaries alongside letters and public documents, rather than treating them as complete accounts.',
     '[{"id":"A","text":"They contain only imaginary events."},{"id":"B","text":"They are more complete than all public documents."},{"id":"C","text":"They provide useful but selective evidence."},{"id":"D","text":"They should be excluded from historical research."}]', '["C"]',
     'The historian values the details in diaries but checks their selective accounts against other sources.'),
    ('pte-lis-049', 'The Campus Bicycle Scheme', 'What change will the bicycle scheme introduce first? Select ONE answer.',
     'Students have suggested buying electric bicycles and reducing membership fees. Both ideas remain under review. Our immediate priority is a simpler repair reporting system: from next month, riders can scan a code on a bicycle to report a fault. New parking stands are also planned, but installation depends on a later budget decision.',
     '[{"id":"A","text":"Lower membership fees."},{"id":"B","text":"Code-based fault reporting."},{"id":"C","text":"A fleet of electric bicycles."},{"id":"D","text":"New parking stands at every building."}]', '["B"]',
     'Fault reporting starts next month. Fees, electric bicycles and parking stands are not yet approved.')
) AS seed(id, title, prompt, script, options, answers, explanation);

-- Subtask 7: Select Missing Word. The missing ending is never spoken.
INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
    title, prompt, audio_transcript, time_limit_seconds, options, correct_answers, explanation, tags)
SELECT id, 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'pte-select-missing-word', 'Select Missing Word',
    title, 'The final word or group of words is replaced by a beep. Select ONE option that completes the recording.',
    script, 90, options::jsonb, answers::jsonb, explanation, ARRAY['PTE Listening', 'Select Missing Word']
FROM (VALUES
    ('pte-lis-050', 'Testing a Measuring Instrument',
     'Before comparing temperatures from different sites, researchers place all their thermometers in the same controlled environment. If one instrument consistently reads higher than the others, they adjust it or record a correction. Otherwise, a difference caused by the equipment could be mistaken for a difference in the [beep]',
     '[{"id":"A","text":"weather conditions"},{"id":"B","text":"research budget"},{"id":"C","text":"publication date"},{"id":"D","text":"team membership"}]', '["A"]',
     'Calibration prevents instrument error being mistaken for a real difference in weather conditions.'),
    ('pte-lis-051', 'Designing Accessible Signs',
     'A sign can look attractive on a computer screen yet be difficult to read in a busy station. Designers must consider viewing distance, lighting and the contrast between letters and background. Testing signs with actual passengers helps ensure that visual style does not come at the expense of [beep]',
     '[{"id":"A","text":"decoration"},{"id":"B","text":"advertising revenue"},{"id":"C","text":"readability"},{"id":"D","text":"printing speed"}]', '["C"]',
     'Distance, lighting and contrast all affect readability, the practical goal contrasted with visual style.'),
    ('pte-lis-052', 'Planning for Rainfall',
     'For years the town enlarged its drains whenever heavy rain caused flooding. Engineers now recommend parks with shallow basins and surfaces that allow water to soak into the ground. Instead of moving every drop away as quickly as possible, this approach gives rainfall somewhere to collect and gradually [beep]',
     '[{"id":"A","text":"increase traffic"},{"id":"B","text":"enter the soil"},{"id":"C","text":"damage buildings"},{"id":"D","text":"block the drains"}]', '["B"]',
     'The alternative design stores rainfall temporarily and allows it to enter the soil rather than immediately draining away.'),
    ('pte-lis-053', 'Evaluating Online Sources',
     'A professional looking website is not necessarily a reliable source. Before citing a claim, students should identify its author, examine the supporting evidence and check when it was published. Comparing the claim with independent sources is another way to assess its [beep]',
     '[{"id":"A","text":"popularity"},{"id":"B","text":"colour scheme"},{"id":"C","text":"subscription price"},{"id":"D","text":"credibility"}]', '["D"]',
     'Authorship, evidence, currency and independent checks help evaluate credibility rather than appearance or popularity.'),
    ('pte-lis-054', 'Keeping a Language Alive',
     'Recording traditional stories preserves valuable examples of an endangered language. Yet an archive alone cannot guarantee that the language will survive. Children need opportunities to speak it with family members and friends. Preservation therefore depends not only on documentation but also on continued [beep]',
     '[{"id":"A","text":"everyday use"},{"id":"B","text":"silent storage"},{"id":"C","text":"translation into other languages"},{"id":"D","text":"restriction of access"}]', '["A"]',
     'The speaker contrasts archival documentation with active, everyday communication needed for survival.')
) AS seed(id, title, script, options, answers, explanation);

-- Subtask 8: Highlight Incorrect Word. Whitespace-delimited positions are
-- stable occurrence IDs, including repeated words. Each passage has exactly
-- three substitutions and otherwise matches its spoken transcript.
WITH seed(id, title, passage, script, explanation) AS (VALUES
    ('pte-lis-055', 'Excavation Records',
     'Archaeologists record the location of every object before it is cleaned. Even a large fragment can reveal how a settlement developed. Careful photography preserves evidence that might otherwise be found when soil is removed. These records allow future visitors to reconsider the original interpretation.',
     'Archaeologists record the location of every object before it is cleaned. Even a small fragment can reveal how a settlement developed. Careful photography preserves evidence that might otherwise be lost when soil is removed. These records allow future researchers to reconsider the original interpretation.',
     'Select large, found and visitors. The speaker says small, lost and researchers.'),
    ('pte-lis-056', 'Testing Building Materials',
     'Engineers test new building materials under controlled conditions before recommending them for widespread use. Samples experience repeated cycles of heating and freezing to simulate seasonal change. The results reveal whether a material will remain expensive over time. Testing also helps manufacturers reduce quality during production.',
     'Engineers test new building materials under controlled conditions before recommending them for widespread use. Samples experience repeated cycles of heating and cooling to simulate seasonal change. The results reveal whether a material will remain durable over time. Testing also helps manufacturers reduce waste during production.',
     'Select freezing, expensive and quality. The speaker says cooling, durable and waste.'),
    ('pte-lis-057', 'Library Collection Review',
     'The library reviews its collection every year to identify books that need replacing. The library also asks students which subjects require fewer resources. Damaged volumes are repaired whenever possible, while duplicate copies may be sold to local schools. This process keeps the collection both useful and outdated.',
     'The library reviews its collection every year to identify books that need replacing. The library also asks students which subjects require additional resources. Damaged volumes are repaired whenever possible, while duplicate copies may be donated to local schools. This process keeps the collection both useful and current.',
     'Select fewer, sold and outdated. The speaker says additional, donated and current; repeated occurrences of library are unchanged.'),
    ('pte-lis-058', 'Mountain Weather Stations',
     'Automatic weather stations collect measurements in remote mountain areas throughout the year. Solar panels provide power, while a battery stores energy for cloudy periods. Researchers receive the readings through a satellite connection and check them for errors. Regular neglect is still essential because dust can repair exposed equipment.',
     'Automatic weather stations collect measurements in remote mountain areas throughout the year. Solar panels provide power, while a battery stores energy for cloudy periods. Researchers receive the readings through a satellite connection and check them for errors. Regular maintenance is still essential because ice can damage exposed equipment.',
     'Select neglect, dust and repair. The speaker says maintenance, ice and damage.'),
    ('pte-lis-059', 'Peer Feedback Workshops',
     'During the workshop, students exchange drafts and provide feedback using a shared checklist. They focus first on the clarity of the argument rather than minor spelling errors. Each writer then decides which suggestions to ignore. This structured process develops both critical reading and the inability to revise work independently.',
     'During the workshop, students exchange drafts and provide feedback using a shared checklist. They focus first on the clarity of the argument rather than minor spelling errors. Each writer then decides which suggestions to adopt. This collaborative process develops both critical reading and the ability to revise work independently.',
     'Select ignore, structured and inability. The speaker says adopt, collaborative and ability.')
)
INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
    title, prompt, context_passage, audio_transcript, time_limit_seconds, options, correct_answers, explanation, tags)
SELECT id, 'pte-2026-01', 'PTE', ARRAY['PTE'], 'listening', 'pte-highlight-incorrect-word', 'Highlight Incorrect Word',
    title, 'Select the words in the displayed transcript that differ from what you hear.',
    passage, script, 120, word_keys.options, word_keys.answers, explanation, ARRAY['PTE Listening', 'Highlight Incorrect Word']
FROM seed
CROSS JOIN LATERAL (
    SELECT jsonb_agg(jsonb_build_object('id', 'w' || position, 'text', word) ORDER BY position) AS options,
           jsonb_agg('w' || position ORDER BY position)
               FILTER (WHERE word <> (regexp_split_to_array(seed.script, '\s+'))[position]) AS answers
    FROM regexp_split_to_table(seed.passage, '\s+') WITH ORDINALITY AS words(word, position)
) AS word_keys;
