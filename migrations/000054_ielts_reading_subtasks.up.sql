-- Original IELTS reading practice: five questions per current sidebar subtask.
-- Multiple Choice combines three single-answer and two multiple-answer items.
-- No legacy sequencing/heading tasks, existing content or mock blueprints change.
-- The fictional case study is explicitly labelled; answers require only this text.
-- Replay preserves editorial changes. All new IDs use the ir54 namespace.

INSERT INTO reading_passages
    (id, exam_version_id, title, subtitle, paragraphs, word_count, difficulty, topic, tags, passage_slot)
SELECT 'rp-ir54-library', 'ielts-2026-01', 'A Library of Things',
       'An original fictional case study of shared equipment', body,
       (SELECT count(*) FROM jsonb_array_elements(body) p,
        regexp_split_to_table(trim(p->>'text'), '\s+') word),
       'medium', 'Community services', ARRAY['IELTS Reading', 'Original practice'], 'custom'
FROM (VALUES ($j$[
 {"label":"A","text":"When the fictional town of Bellford opened a lending room for household equipment, its organisers resisted calling it a shop. Residents would borrow objects rather than buy them, returning each one for somebody else to use. The project began in a room beside the public library in 2018. Its initial stock came entirely from donations, although the team rejected items without instructions or with missing parts. Many donors imagined that their rarely used possessions would immediately become popular. The organisers were less certain: an object that was useful to its owner might be too specialised for a shared collection. They therefore recorded requests before deciding which additional equipment to acquire. The room was modest, with open shelves for small objects and lockable cupboards for powered equipment. Visitors could examine a catalogue at the entrance without becoming members. This mattered because many people initially assumed that borrowing required a complicated application. Staff wanted the first conversation to concern the job a resident hoped to complete, rather than the name of a particular machine. Sometimes a simpler object already on the shelves was sufficient."},
 {"label":"B","text":"The first year challenged the assumption that larger objects would attract the most interest. Carpet cleaners accounted for more loans than any other category, while a collection of heavy garden machines was borrowed infrequently. Interviews suggested that transporting those machines was a greater obstacle than operating them. The team also discovered that a long waiting list could be misleading. Several residents had registered for the same weekend and would not accept alternative dates. Rather than simply purchasing duplicates of everything on the list, staff began recording when requests were made and whether a later loan would still be useful. This distinguished a persistent shortage from a brief seasonal peak."},
 {"label":"C","text":"The lending procedure changed in 2020. Initially, every borrower had paid the same refundable deposit. Organisers replaced this with a deposit waiver for residents referred by local support organisations, while keeping the ordinary deposit for other borrowers. Both groups remained responsible for reporting damage. Staff also introduced a short demonstration before anyone could borrow a powered tool for the first time. This was not a formal qualification: borrowers watched a safe-use demonstration and practised the basic controls under supervision. A printed checklist travelled with each tool. The annual membership fee, by contrast, remained unchanged throughout the first three years, even though the collection grew."},
 {"label":"D","text":"Behind the counter, maintenance required more work than the founders had expected. Volunteers inspected every returned object and entered faults in a shared log. A red label meant that an item could not be lent until it had been repaired and checked again. Replacement parts were purchased from the maintenance budget, not from a separate charge imposed on the next borrower. The log eventually revealed that three models of the same appliance required different filters. Standardising future purchases would make repairs simpler, but immediately discarding the existing models would waste working equipment. Staff chose to standardise gradually, replacing machines only when repair was no longer practical."},
 {"label":"E","text":"Evaluating the environmental benefits proved difficult. Counting loans was straightforward; establishing what would have happened without the service was not. Some borrowers said they would otherwise have bought a new object, but others would have borrowed from a neighbour or abandoned the job. The evaluation report therefore refused to treat every loan as a purchase prevented. It also recorded journeys to collect equipment, because an extra car trip could offset part of the environmental saving. The writer considers this caution a strength rather than a failure. Claims about reduced consumption are more credible when they acknowledge uncertainty, and a busy lending desk alone is not proof of an environmental success."},
 {"label":"F","text":"Bellford's experience suggests that shared ownership is as much a service-design problem as a storage problem. In the writer's view, expanding the collection should come after making the existing stock easier to borrow. Convenient opening hours and reliable maintenance deserve priority over an impressive catalogue. The writer also argues that paid coordination is necessary for a dependable long-term service; goodwill is valuable, but should not be treated as an unlimited supply of labour. A proposed delivery service might help residents who cannot carry equipment, yet its costs and environmental effects still need investigation. The sensible next step is a small trial, not an immediate promise to deliver every object to every address."}
]$j$::jsonb)) AS content(body)
ON CONFLICT (id) DO NOTHING;

INSERT INTO reading_question_groups
    (id, passage_id, position, type_id, type_name, instructions, resources, passage_display, shuffle_questions, time_limit_seconds)
SELECT 'g-ir54-' || key, 'rp-ir54-library', position, type_id, type_name,
       instructions, '[]'::jsonb, 'full', FALSE, 420
FROM (VALUES
 ('single', 1, 'reading-mcq-single', 'Multiple Choice, Single Answer', 'Choose ONE letter, A, B, C or D.'),
 ('multiple', 2, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers', 'Choose TWO letters for each question.'),
 ('tfng', 3, 'reading-true-false', 'True / False / Not Given', 'Write TRUE if the statement agrees with the information, FALSE if it contradicts it, or NOT GIVEN if there is no information on this.'),
 ('ynng', 4, 'reading-yes-no-not-given', 'Yes / No / Not Given', 'Write YES if the statement agrees with the views of the writer, NO if it contradicts them, or NOT GIVEN if the view is not stated.'),
 ('matching', 5, 'reading-matching-information', 'Matching Information', 'Which paragraph, A-F, contains the following information? You may use any letter more than once.'),
 ('completion', 6, 'reading-sentence-completion', 'Sentence Completion', 'Complete each sentence. Write ONE WORD ONLY from the passage in each gap.')
) AS groups(key, position, type_id, type_name, instructions)
ON CONFLICT (id) DO NOTHING;

-- Derive type and passage from the group so their identities cannot drift.
INSERT INTO questions
    (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
     title, prompt, options, correct_answers, explanation, points,
     difficulty, tags, passage_id, group_id, group_position, is_published, time_limit_seconds)
SELECT 'q-ir54-' || v.key || '-' || v.position, 'ielts-2026-01', 'IELTS', ARRAY['IELTS'],
       'reading', g.type_id, g.type_name, 'A Library of Things', v.prompt,
       v.options::jsonb, v.answers::jsonb, v.explanation, 1, 'medium',
       ARRAY['IELTS Reading', 'Original practice'], g.passage_id, g.id, v.position, TRUE, 0
FROM (VALUES
 ('single', 1, 'Why did organisers record requests before acquiring more equipment?',
  '[{"id":"A","text":"To calculate how much to pay donors"},{"id":"B","text":"To identify objects residents actually needed"},{"id":"C","text":"To replace the public library with a shop"},{"id":"D","text":"To accept objects with missing parts"}]',
  '["B"]', 'Paragraph A: usefulness to an individual owner did not establish demand for a shared collection, so requests informed acquisitions.'),
 ('single', 2, 'What mainly discouraged residents from borrowing heavy garden machines?',
  '[{"id":"A","text":"The difficulty of operating them"},{"id":"B","text":"The absence of written instructions"},{"id":"C","text":"The difficulty of transporting them"},{"id":"D","text":"Their high purchase price"}]',
  '["C"]', 'Paragraph B explicitly identifies transport as a greater obstacle than operation.'),
 ('single', 3, 'Why did staff decide to standardise appliances gradually?',
  '[{"id":"A","text":"Working equipment would otherwise be wasted"},{"id":"B","text":"Different filters were easier to store"},{"id":"C","text":"Volunteers refused to keep a fault log"},{"id":"D","text":"All existing machines were already beyond repair"}]',
  '["A"]', 'Paragraph D: replacing all existing models immediately would discard functioning equipment.'),
 ('multiple', 1, 'Which TWO changes were introduced in 2020?',
  '[{"id":"A","text":"The membership fee was increased"},{"id":"B","text":"All borrowers stopped paying deposits"},{"id":"C","text":"Borrowers no longer had to report damage"},{"id":"D","text":"Some referred residents could borrow without a deposit"},{"id":"E","text":"First-time powered-tool borrowers received a demonstration"}]',
  '["D","E"]', 'Paragraph C describes a deposit waiver for referred residents and a first-use demonstration; the fee stayed unchanged and damage still had to be reported.'),
 ('multiple', 2, 'Which TWO factors made it difficult to establish environmental savings?',
  '[{"id":"A","text":"Some borrowers would not otherwise have bought equipment"},{"id":"B","text":"The service kept no record of loan numbers"},{"id":"C","text":"Collection journeys could reduce the benefit"},{"id":"D","text":"Every loan required a new purchase"},{"id":"E","text":"The evaluation excluded all transport information"}]',
  '["A","C"]', 'Paragraph E distinguishes avoided purchases from other alternatives and includes the impact of collection journeys.')
) AS v(key, position, prompt, options, answers, explanation)
JOIN reading_question_groups g ON g.id = 'g-ir54-' || v.key
ON CONFLICT (id) DO NOTHING;

INSERT INTO questions
    (id, exam_version_id, exam, supported_exams, skill, type_id, type_name,
     title, prompt, options, correct_answers, explanation, points,
     difficulty, tags, passage_id, group_id, group_position, is_published, time_limit_seconds)
SELECT 'q-ir54-' || v.key || '-' || v.position, 'ielts-2026-01', 'IELTS', ARRAY['IELTS'],
       'reading', g.type_id, g.type_name, 'A Library of Things', v.prompt,
       CASE v.key
         WHEN 'tfng' THEN '[{"id":"TRUE","text":"True"},{"id":"FALSE","text":"False"},{"id":"NOT_GIVEN","text":"Not Given"}]'::jsonb
         WHEN 'ynng' THEN '[{"id":"YES","text":"Yes"},{"id":"NO","text":"No"},{"id":"NOT_GIVEN","text":"Not Given"}]'::jsonb
         WHEN 'matching' THEN '[{"id":"A","text":"Paragraph A"},{"id":"B","text":"Paragraph B"},{"id":"C","text":"Paragraph C"},{"id":"D","text":"Paragraph D"},{"id":"E","text":"Paragraph E"},{"id":"F","text":"Paragraph F"}]'::jsonb
         ELSE '[]'::jsonb
       END,
       jsonb_build_array(v.answer), v.explanation, 1, 'medium',
       ARRAY['IELTS Reading', 'Original practice'], g.passage_id, g.id, v.position, TRUE, 0
FROM (VALUES
 ('tfng', 1, 'All equipment in the original collection was donated.', 'TRUE', 'Paragraph A says the initial stock came entirely from donations.'),
 ('tfng', 2, 'Heavy garden machines were the most frequently borrowed category in the first year.', 'FALSE', 'Paragraph B identifies carpet cleaners as the most borrowed category and says garden machines were borrowed infrequently.'),
 ('tfng', 3, 'More than half of the first-year borrowers lived within walking distance of the lending room.', 'NOT_GIVEN', 'The passage gives no distribution of borrower addresses or distances from the lending room.'),
 ('tfng', 4, 'The annual membership fee increased during the first three years.', 'FALSE', 'Paragraph C explicitly says the annual membership fee remained unchanged throughout the first three years.'),
 ('tfng', 5, 'Each returned object was inspected before being lent again.', 'TRUE', 'Paragraph D says volunteers inspected every returned object and prevented faulty items from being lent until repaired and checked.'),
 ('ynng', 1, 'Acknowledging uncertainty makes environmental claims more convincing.', 'YES', 'Paragraph E says claims are more credible when they acknowledge uncertainty; the writer regards this caution as a strength.'),
 ('ynng', 2, 'High borrowing numbers alone demonstrate environmental success.', 'NO', 'Paragraph E explicitly rejects a busy lending desk as sufficient proof of environmental success.'),
 ('ynng', 3, 'Every town should legally be required to establish an equipment-lending service.', 'NOT_GIVEN', 'The writer discusses priorities for a service, not a legal requirement for every town to create one.'),
 ('ynng', 4, 'Making the existing collection easier to borrow should precede expansion.', 'YES', 'Paragraph F explicitly puts improving access to the current stock before expanding the collection.'),
 ('ynng', 5, 'A reliable long-term lending service should depend entirely on unpaid labour.', 'NO', 'Paragraph F argues that paid coordination is necessary and goodwill should not be treated as unlimited labour.'),
 ('matching', 1, 'An explanation of why a waiting list can exaggerate lasting demand', 'B', 'Paragraph B distinguishes several requests for the same weekend from persistent demand.'),
 ('matching', 2, 'A visual signal that prevents faulty equipment from being borrowed', 'D', 'Paragraph D describes the red label used to block lending until repairs and rechecking are complete.'),
 ('matching', 3, 'Reasons for rejecting some offered equipment at the outset', 'A', 'Paragraph A says donations with missing parts or without instructions were rejected.'),
 ('matching', 4, 'A proposal that should be tested on a limited scale before wider introduction', 'F', 'Paragraph F recommends a small trial of delivery rather than a promise to deliver everything everywhere.'),
 ('matching', 5, 'Examples of what borrowers might have done if the service had not existed', 'E', 'Paragraph E lists buying an object, borrowing from a neighbour and abandoning the job as alternatives.'),
 ('completion', 1, 'The initial equipment collection consisted entirely of ________.', 'donations', 'Paragraph A: the initial stock came entirely from donations.'),
 ('completion', 2, 'Some residents on the waiting list would not accept alternative ________.', 'dates', 'Paragraph B says residents who wanted the same weekend would not accept alternative dates.'),
 ('completion', 3, 'A printed ________ accompanied each powered tool.', 'checklist', 'Paragraph C states that a printed checklist travelled with each tool.'),
 ('completion', 4, 'Volunteers recorded equipment faults in a shared ________.', 'log', 'Paragraph D says volunteers entered faults in a shared log.'),
 ('completion', 5, 'Three models of one appliance needed different ________.', 'filters', 'Paragraph D identifies different filters as the reason standardisation could simplify repairs.')
) AS v(key, position, prompt, answer, explanation)
JOIN reading_question_groups g ON g.id = 'g-ir54-' || v.key
ON CONFLICT (id) DO NOTHING;
