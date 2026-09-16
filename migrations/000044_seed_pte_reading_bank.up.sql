-- 000044_seed_pte_reading_bank.up.sql
-- Authentic PTE Academic Reading question bank: 5 questions for each of the 5 task types (25 questions total).
--
-- 1. fill-in-blanks-rw (Reading & Writing: Fill in the Blanks) - 5 questions
-- 2. reading-mcq-multiple (Multiple Choice, Multiple Answers) - 5 questions
-- 3. reorder-paragraphs (Re-order Paragraphs) - 5 questions
-- 4. fill-in-blanks-r (Reading: Fill in the Blanks) - 5 questions
-- 5. reading-mcq-single (Multiple Choice, Single Answer) - 5 questions

-- ============================================================================
-- 1. PASSAGES FOR TASK 1: fill-in-blanks-rw (5 passages, passage_display = 'hidden')
-- ============================================================================

INSERT INTO reading_passages (
    id, exam_version_id, title, subtitle, paragraphs, sources,
    word_count, difficulty, topic, tags, is_published, passage_slot
) VALUES
('rp-pte-fbrw-01', 'pte-2026-01', 'Urban Heat Islands and Microclimates', 'Thermal dynamics in urban environments',
 $j$[{"label":"A","text":"The phenomenon known as the urban heat island effect occurs when metropolitan areas experience significantly higher temperatures than surrounding rural regions. This thermal disparity is largely attributable to the replacement of natural vegetation with dense concentrations of concrete, asphalt, and masonry, which absorb and retain solar radiation throughout the day. At night, these artificial surfaces gradually release stored thermal energy, impeding the atmospheric cooling that typically occurs after dusk. In addition, tall skyscrapers create geometric canyons that obstruct airflow and prevent convective heat dissipation. Urban planners are increasingly turning to architectural interventions, such as green roofs and reflective pavements, to mitigate these localized temperature increases. Strategic canopy expansion can lower surface temperatures by up to eight degrees Celsius, reducing cooling demand and enhancing urban habitability."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Environmental Science', ARRAY['PTE', 'Reading', 'Fill in the Blanks RW'], TRUE, 'custom'),

('rp-pte-fbrw-02', 'pte-2026-01', 'Bioluminescence in the Abyssal Zone', 'Biochemical adaptation in the deep ocean',
 $j$[{"label":"A","text":"In the perpetual darkness of the deep ocean, bioluminescence serves as a vital tool for survival and communication. Unlike terrestrial creatures that primarily rely on reflected sunlight, bathypelagic organisms have evolved biochemical pathways to produce light through enzymatic reactions. The primary molecule involved in this process is luciferin, which emits visible radiation when oxidized by the enzyme luciferase in the presence of oxygen. Deep-sea predators use luminous lures to attract unsuspecting prey, while smaller organisms employ counter-illumination to disguise their silhouettes against the faint downwelling light from above. Furthermore, some species release sudden clouds of luminescent fluid when threatened, an evasion tactic that temporarily disorients pursuers and affords the prey sufficient time to escape. The wide distribution of this trait suggests substantial evolutionary advantages in aphotic environments."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Marine Biology', ARRAY['PTE', 'Reading', 'Fill in the Blanks RW'], TRUE, 'custom'),

('rp-pte-fbrw-03', 'pte-2026-01', 'Cognitive Archaeology and Symbolism', 'Reconstructing prehistoric thought from material culture',
 $j$[{"label":"A","text":"Cognitive archaeology seeks to reconstruct the conceptual frameworks of ancient human populations from material culture. Rather than merely documenting the physical tools fabricated by early hominins, researchers analyze the cognitive capacities required to conceptualize, plan, and manufacture complex composite implements. The emergence of symbolic thought is commonly linked to the appearance of personal ornaments, engraved ochre pieces, and parietal art dating to the Middle Paleolithic. Such decorative artifacts demonstrate that our ancestors were capable of externalizing information and maintaining shared social meanings beyond immediate sensory perceptions. However, researchers must avoid the anachronistic tendency to impose modern concepts onto prehistoric mindsets. Combining archaeological excavations with neurological insights establishes a rigorous foundation for understanding symbolic cognition."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Archaeology', ARRAY['PTE', 'Reading', 'Fill in the Blanks RW'], TRUE, 'custom'),

('rp-pte-fbrw-04', 'pte-2026-01', 'The Genetic Architecture of Crop Domestication', 'Selective breeding and physiological mutations in early farming',
 $j$[{"label":"A","text":"The transition from wild gathering to systematic agriculture represents one of the most profound turning points in human prehistory. Over thousands of years, early farmers exerted selective pressure on wild grasses, favoring mutations that altered key physiological traits. In wild cereals, seeds shatter and disperse spontaneously upon reaching maturity, an adaptation that ensures reproductive dispersal but severely hampers manual harvesting. Early domesticators deliberately selected non-shattering mutants, whose grains remained firmly attached to the central stalk until threshed. This fundamental alteration made cultivation viable, although it simultaneously rendered the plants wholly dependent on human intervention for propagation. Genomics reveals that major phenotypic changes can occur with minimal genomic alteration."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Genetics & Agriculture', ARRAY['PTE', 'Reading', 'Fill in the Blanks RW'], TRUE, 'custom'),

('rp-pte-fbrw-05', 'pte-2026-01', 'Behavioral Economics and Choice Architecture', 'How default options and nudges guide decision-making',
 $j$[{"label":"A","text":"Traditional neoclassical economics rests on the assumption that individuals behave as rational actors who consistently evaluate all available information to maximize personal utility. In contrast, contemporary behavioral economists argue that human judgment is frequently constrained by cognitive heuristics and systematic biases. When faced with complex financial decisions, consumers often experience choice overload and opt for inaction. To address this inertia, policymakers restructure choice architecture without restricting liberty. The most influential manifestation is the use of default options. Automatic enrollment for retirement savings yields higher participation rates, as individuals accept the pre-selected path unless they make a conscious effort to change it. Such interventions yield societal dividends while maintaining autonomy."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Economics & Psychology', ARRAY['PTE', 'Reading', 'Fill in the Blanks RW'], TRUE, 'custom');

-- GROUPS FOR TASK 1
INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds
) VALUES
('g-pte-fbrw-01', 'rp-pte-fbrw-01', 1, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180),
('g-pte-fbrw-02', 'rp-pte-fbrw-02', 1, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180),
('g-pte-fbrw-03', 'rp-pte-fbrw-03', 1, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180),
('g-pte-fbrw-04', 'rp-pte-fbrw-04', 1, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180),
('g-pte-fbrw-05', 'rp-pte-fbrw-05', 1, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180);

-- QUESTIONS FOR TASK 1
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, blanks, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position
) VALUES
('q-pte-fbrw-01', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'Urban Heat Islands and Microclimates',
 'Select the appropriate answer choice for each blank.',
 'The phenomenon known as the urban heat island effect occurs when metropolitan areas experience significantly higher temperatures than surrounding rural regions. This thermal disparity is largely [[b1]] to the replacement of natural vegetation with dense concentrations of concrete, asphalt, and masonry, which absorb and retain solar radiation throughout the day. At night, these artificial surfaces gradually release stored thermal energy, [[b2]] the atmospheric cooling that typically occurs after dusk. In addition, tall skyscrapers create geometric canyons that obstruct airflow and prevent convective heat dissipation. Urban planners are increasingly turning to architectural interventions, such as green roofs and reflective pavements, to [[b3]] these localized temperature increases. Strategic canopy expansion can lower surface temperatures by up to eight degrees Celsius, reducing cooling demand and enhancing urban [[b4]].',
 $j$[
   {"id":"b1","options":["attributable","susceptible","vulnerable","incompatible"],"correctAnswer":"attributable"},
   {"id":"b2","options":["impeding","expediting","facilitating","overlooking"],"correctAnswer":"impeding"},
   {"id":"b3","options":["mitigate","escalate","aggravate","prolong"],"correctAnswer":"mitigate"},
   {"id":"b4","options":["habitability","hostility","fragility","obscurity"],"correctAnswer":"habitability"}
 ]$j$::jsonb,
 '"attributable to" is the required collocation; "impeding" reflects blocking cooling; "mitigate" means alleviate; "habitability" denotes liveability.',
 180, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbrw-01', 'g-pte-fbrw-01', 1),

('q-pte-fbrw-02', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'Bioluminescence in the Abyssal Zone',
 'Select the appropriate answer choice for each blank.',
 'In the perpetual darkness of the deep ocean, bioluminescence serves as a vital tool for survival and communication. Unlike terrestrial creatures that primarily rely on reflected sunlight, bathypelagic organisms have evolved biochemical pathways to produce light through enzymatic reactions. The primary molecule involved in this process is luciferin, which emits visible radiation when [[b1]] by the enzyme luciferase in the presence of oxygen. Deep-sea predators use luminous lures to attract unsuspecting prey, while smaller organisms employ counter-illumination to [[b2]] their silhouettes against the faint downwelling light from above. Furthermore, some species release sudden clouds of luminescent fluid when threatened, an evasion tactic that temporarily [[b3]] pursuers and affords the prey sufficient time to escape. The wide distribution of this trait suggests substantial evolutionary [[b4]] in aphotic environments.',
 $j$[
   {"id":"b1","options":["oxidized","diluted","neutralized","coagulated"],"correctAnswer":"oxidized"},
   {"id":"b2","options":["disguise","magnify","accentuate","dismantle"],"correctAnswer":"disguise"},
   {"id":"b3","options":["disorients","illuminates","reassures","stimulates"],"correctAnswer":"disorients"},
   {"id":"b4","options":["advantages","impediments","hazards","penalties"],"correctAnswer":"advantages"}
 ]$j$::jsonb,
 'Luciferin is "oxidized" in the reaction; counter-illumination serves to "disguise" silhouettes; luminous fluid "disorients" predators; the trait confers "advantages".',
 180, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbrw-02', 'g-pte-fbrw-02', 1),

('q-pte-fbrw-03', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'Cognitive Archaeology and Symbolism',
 'Select the appropriate answer choice for each blank.',
 'Cognitive archaeology seeks to reconstruct the conceptual frameworks of ancient human populations from material culture. Rather than merely documenting the physical tools fabricated by early hominins, researchers analyze the cognitive [[b1]] required to conceptualize, plan, and manufacture complex composite implements. The emergence of symbolic thought is commonly linked to the appearance of personal ornaments, engraved ochre pieces, and parietal art dating to the Middle Paleolithic. Such decorative artifacts demonstrate that our ancestors were capable of externalizing information and maintaining shared social meanings beyond immediate sensory [[b2]]. However, researchers must avoid the [[b3]] tendency to impose modern concepts onto prehistoric mindsets. Combining archaeological excavations with neurological insights establishes a [[b4]] foundation for understanding symbolic cognition.',
 $j$[
   {"id":"b1","options":["capacities","limitations","concessions","interferences"],"correctAnswer":"capacities"},
   {"id":"b2","options":["perceptions","rejections","distractions","inhibitions"],"correctAnswer":"perceptions"},
   {"id":"b3","options":["anachronistic","altruistic","enigmatic","aesthetic"],"correctAnswer":"anachronistic"},
   {"id":"b4","options":["rigorous","fictitious","negligent","tentative"],"correctAnswer":"rigorous"}
 ]$j$::jsonb,
 '"cognitive capacities" is standard; "sensory perceptions" represents immediate physical input; "anachronistic" means placing modern ideas into antiquity; "rigorous foundation" is the academic collocation.',
 180, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbrw-03', 'g-pte-fbrw-03', 1),

('q-pte-fbrw-04', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'The Genetic Architecture of Crop Domestication',
 'Select the appropriate answer choice for each blank.',
 'The transition from wild gathering to systematic agriculture represents one of the most profound turning points in human prehistory. Over thousands of years, early farmers exerted selective pressure on wild grasses, favoring mutations that altered key physiological [[b1]]. In wild cereals, seeds shatter and disperse spontaneously upon reaching maturity, an adaptation that ensures reproductive dispersal but severely [[b2]] manual harvesting. Early domesticators deliberately selected non-shattering mutants, whose grains remained firmly attached to the central stalk until threshed. This fundamental alteration made cultivation viable, although it simultaneously rendered the plants wholly [[b3]] on human intervention for propagation. Genomics reveals that major phenotypic changes can occur with minimal genomic [[b4]].',
 $j$[
   {"id":"b1","options":["traits","flaws","margins","hazards"],"correctAnswer":"traits"},
   {"id":"b2","options":["hampers","enhances","justifies","underestimates"],"correctAnswer":"hampers"},
   {"id":"b3","options":["dependent","indifferent","hostile","negligent"],"correctAnswer":"dependent"},
   {"id":"b4","options":["alteration","retention","consumption","validation"],"correctAnswer":"alteration"}
 ]$j$::jsonb,
 'Physiological "traits"; shattering "hampers" harvesting; domestic crops became "dependent" on humans; "genomic alteration" describes genetic mutation.',
 180, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbrw-04', 'g-pte-fbrw-04', 1),

('q-pte-fbrw-05', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'Behavioral Economics and Choice Architecture',
 'Select the appropriate answer choice for each blank.',
 'Traditional neoclassical economics rests on the assumption that individuals behave as rational actors who consistently evaluate all available information to maximize personal utility. In contrast, contemporary behavioral economists argue that human judgment is frequently [[b1]] by cognitive heuristics and systematic biases. When faced with complex financial decisions, consumers often experience choice overload and opt for inaction. To address this inertia, policymakers restructure choice architecture without restricting liberty. The most influential manifestation is the use of default options. Automatic enrollment for retirement savings yields higher participation rates, as individuals accept the pre-selected path unless they make a [[b2]] effort to change it. Such interventions yield societal [[b3]] while maintaining [[b4]].',
 $j$[
   {"id":"b1","options":["constrained","clarified","liberated","vindicated"],"correctAnswer":"constrained"},
   {"id":"b2","options":["conscious","negligible","sporadic","coercive"],"correctAnswer":"conscious"},
   {"id":"b3","options":["dividends","penalties","liabilities","setbacks"],"correctAnswer":"dividends"},
   {"id":"b4","options":["autonomy","compliance","conformity","stagnation"],"correctAnswer":"autonomy"}
 ]$j$::jsonb,
 'Judgment is "constrained" by biases; requires a "conscious" effort; yields societal "dividends" (benefits); preserves personal "autonomy".',
 180, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbrw-05', 'g-pte-fbrw-05', 1);

-- ============================================================================
-- 2. PASSAGES FOR TASK 2: reading-mcq-multiple (5 passages, passage_display = 'full')
-- ============================================================================

INSERT INTO reading_passages (
    id, exam_version_id, title, subtitle, paragraphs, sources,
    word_count, difficulty, topic, tags, is_published, passage_slot
) VALUES
('rp-pte-mcqm-01', 'pte-2026-01', 'Ocean Warming and Scleractinian Corals', 'Thermal thresholds and bleaching dynamics',
 $j$[{"label":"A","text":"Coral reefs harbor approximately one quarter of all marine species despite covering less than one percent of the seabed. Scleractinian corals maintain an obligate endosymbiosis with photosynthetic microalgae called zooxanthellae. Zooxanthellae supply up to ninety percent of the coral host's energetic needs by translocating photosynthate. However, when sea temperatures exceed regional summer thresholds by even one degree Celsius, algal photosynthetic machinery produces damaging reactive oxygen species. In response, the coral host expels the endosymbionts, resulting in coral bleaching. Bleached corals do not die immediately, but experience severe nutritional deprivation. If elevated temperatures persist for multiple weeks, mortality ensues and opportunistic macroalgae overgrow the bare limestone skeleton, preventing larval settlement."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Marine Ecology', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqm-02', 'pte-2026-01', 'The Antikythera Mechanism and Hellenistic Astronomy', 'Mechanical computation in ancient Greece',
 $j$[{"label":"A","text":"Retrieved in 1901 from a Roman shipwreck off Antikythera, the Antikythera Mechanism is recognized as the oldest surviving analog computer, dating to between 150 and 100 BCE. The corroded bronze artifact puzzled early historians who doubted ancient mechanics could construct differential gearing. Modern micro-focus X-ray computed tomography revealed over thirty bronze gears operated by a hand crank. The device tracked the solar and lunar positions, naked-eye planets, and eclipse cycles using the 223-month Saros cycle. Significantly, it incorporated a pin-and-slot mechanism to simulate the non-uniform speed of the Moon across its elliptical orbit. Calendar rings also tracked the quadrennial Panhellenic Games, integrating scientific computation with civic life."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'History of Science', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqm-03', 'pte-2026-01', 'The Preservation of Classical Texts in Monasteries', 'The role of Benedictine scriptoriums in preserving literature',
 $j$[{"label":"A","text":"After the collapse of the Western Roman Empire, secular schooling and book production waned across Europe. Benedictine monastic communities stepped into this cultural vacuum as primary conservators of literacy. St. Benedict's Rule prescribed manual labor, which abbeys interpreted to include transcribing manuscripts in scriptoriums. Scribes worked exclusively under daylight to avoid fire hazards from lamps. While their paramount objective was preserving scripture and patristic theology, monks also systematically transcribed pagan Roman authors including Virgil, Cicero, and Horace. Monasteries copied these secular classical writings not for pagan philosophy, but because they provided indispensable grammatical models and rhetorical standards essential for high Latin literacy."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Medieval History', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqm-04', 'pte-2026-01', 'Permafrost Degradation and Climate Feedback', 'Subterranean carbon release from thawing Arctic tundra',
 $j$[{"label":"A","text":"High-latitude permafrost soils contain an estimated 1,500 billion metric tons of organic carbon, more than double the carbon currently in the atmosphere. Sub-zero temperatures have historically prevented microbial breakdown of this organic matter. As Arctic air warms at four times the global rate, widespread thaw is occurring. In well-drained upland soils, aerobic microbes decompose organic material into carbon dioxide. In waterlogged thermokarst collapses, anaerobic conditions cause methanogens to discharge methane, which has a warming impact up to thirty times greater than carbon dioxide. The release of both gases intensifies atmospheric warming, causing further permafrost thaw in a self-reinforcing climate feedback loop."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Climate Science', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqm-05', 'pte-2026-01', 'The Transition to Maritime Silk Routes', 'Technological catalysts for oceanic commerce in East Asia',
 $j$[{"label":"A","text":"For centuries, the overland Silk Road linked China with Mediterranean markets, but caravans suffered severe logistical limits. Camels carried restricted payloads, and merchant caravans were vulnerable to banditry and regional taxation. During the Tang and Song dynasties, Chinese trade pivoted toward maritime routes. Innovations like the magnetic compass, watertight bulkhead hulls, and stern rudders transformed seafaring. Merchant junks transported hundreds of tons per voyage at a fraction of overland costs. This allowed trade to transition from light luxury goods like raw silk to heavy bulk items including ceramic porcelain, grain, and timber, turning coastal ports into thriving commercial emporiums."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Economic History', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom');

-- GROUPS FOR TASK 2
INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds
) VALUES
('g-pte-mcqm-01', 'rp-pte-mcqm-01', 1, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the passage and answer the question by selecting all the correct responses. More than one response is correct. Incorrect responses cancel out correct ones.',
 '[]'::jsonb, 'full', TRUE, 150),
('g-pte-mcqm-02', 'rp-pte-mcqm-02', 1, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the passage and answer the question by selecting all the correct responses. More than one response is correct. Incorrect responses cancel out correct ones.',
 '[]'::jsonb, 'full', TRUE, 150),
('g-pte-mcqm-03', 'rp-pte-mcqm-03', 1, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the passage and answer the question by selecting all the correct responses. More than one response is correct. Incorrect responses cancel out correct ones.',
 '[]'::jsonb, 'full', TRUE, 150),
('g-pte-mcqm-04', 'rp-pte-mcqm-04', 1, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the passage and answer the question by selecting all the correct responses. More than one response is correct. Incorrect responses cancel out correct ones.',
 '[]'::jsonb, 'full', TRUE, 150),
('g-pte-mcqm-05', 'rp-pte-mcqm-05', 1, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the passage and answer the question by selecting all the correct responses. More than one response is correct. Incorrect responses cancel out correct ones.',
 '[]'::jsonb, 'full', TRUE, 150);

-- QUESTIONS FOR TASK 2
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position
) VALUES
('q-pte-mcqm-01', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Coral Bleaching Dynamics',
 'According to the passage, which TWO of the following statements about coral bleaching are correct?',
 $j$[
   {"id":"A","text":"Zooxanthellae provide the majority of the coral host's nutritional energy under normal conditions."},
   {"id":"B","text":"Corals immediately perish at the moment their algal endosymbionts are expelled."},
   {"id":"C","text":"Prolonged thermal stress causes microalgae to generate damaging reactive chemical species."},
   {"id":"D","text":"Scleractinian corals cover approximately twenty-five percent of the global marine floor."},
   {"id":"E","text":"Macroalgae assist in coral recovery by stabilizing bleached limestone skeletons."}
 ]$j$::jsonb,
 '["A","C"]'::jsonb,
 'Paragraph A: zooxanthellae supply up to 90% of energy (A); photosynthetic machinery produces damaging reactive oxygen species (C). B is refuted (do not die immediately); D is incorrect (cover <1%); E is refuted (macroalgae overgrow and prevent settlement).',
 150, 2, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqm-01', 'g-pte-mcqm-01', 1),

('q-pte-mcqm-02', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'The Antikythera Mechanism',
 'Which TWO statements regarding the Antikythera Mechanism are supported by the text?',
 $j$[
   {"id":"A","text":"Modern non-destructive radiographic imaging was essential in deciphering the internal mechanism."},
   {"id":"B","text":"It was initially constructed during the Roman Imperial period after 100 CE."},
   {"id":"C","text":"The device accounted for variations in the speed of the Moon's apparent orbit."},
   {"id":"D","text":"Early twentieth-century scholars immediately recognized its computational purpose."},
   {"id":"E","text":"The mechanism was driven entirely by an automated water clock system."}
 ]$j$::jsonb,
 '["A","C"]'::jsonb,
 'Micro-focus X-ray CT was necessary to inspect workings without destroying it (A); pin-and-slot mechanism simulated the non-uniform speed of the Moon (C).',
 150, 2, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqm-02', 'g-pte-mcqm-02', 1),

('q-pte-mcqm-03', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Medieval Classical Texts',
 'Which TWO of the following factors contributed to the preservation and copying of classical secular texts in medieval monasteries?',
 $j$[
   {"id":"A","text":"Artificial illumination allowed scribes to transcribe manuscripts continuously through the night."},
   {"id":"B","text":"Classical authors provided essential stylistic and grammatical models for Latin literacy."},
   {"id":"C","text":"Benedictine monastic regulations recognized manuscript production as a form of prescribed manual work."},
   {"id":"D","text":"Secular schools funded scriptoriums to maintain legal and commercial archives."},
   {"id":"E","text":"Monks actively embraced pagan philosophical doctrines as superior to Christian theology."}
 ]$j$::jsonb,
 '["B","C"]'::jsonb,
 'Classical writings provided indispensable grammatical models for Latin (B); Rule of St Benedict prescribed manual labor including transcription (C).',
 150, 2, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqm-03', 'g-pte-mcqm-03', 1),

('q-pte-mcqm-04', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Permafrost Degradation Feedback',
 'Which TWO of the following outcomes are identified in the passage as consequences of permafrost thawing?',
 $j$[
   {"id":"A","text":"Complete suppression of methanogenic microbial activity in Arctic wetlands."},
   {"id":"B","text":"Structural ground collapse forming thermokarst water bodies."},
   {"id":"C","text":"Substantial emission of greenhouse gases that further accelerate warming trends."},
   {"id":"D","text":"A reduction in the overall depth of the active surface soil layer."},
   {"id":"E","text":"Absorption of atmospheric carbon dioxide into expanding sub-zero soils."}
 ]$j$::jsonb,
 '["B","C"]'::jsonb,
 'Thermokarst collapses form waterlogged depressions (B); discharge of CO2 and methane intensifies warming in a feedback loop (C).',
 150, 2, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqm-04', 'g-pte-mcqm-04', 1),

('q-pte-mcqm-05', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Silk Road Maritime Transition',
 'According to the passage, which TWO advantages led merchants to favor maritime routes over the traditional overland Silk Road?',
 $j$[
   {"id":"A","text":"Vessels were immune to adverse weather conditions and seasonal monsoons."},
   {"id":"B","text":"Ships could transport vastly greater volumes of cargo at reduced operational costs."},
   {"id":"C","text":"Overland transport was restricted by the physical carrying capacity of pack animals."},
   {"id":"D","text":"Land routes completely ceased operations due to the loss of silk manufacturing techniques."},
   {"id":"E","text":"Maritime voyages were managed entirely without navigational instruments."}
 ]$j$::jsonb,
 '["B","C"]'::jsonb,
 'Merchant junks transported hundreds of tons at a fraction of overland costs (B); camels carried restricted payloads overland (C).',
 150, 2, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqm-05', 'g-pte-mcqm-05', 1);

-- ============================================================================
-- 3. RE-ORDER PARAGRAPHS (5 items in reading_reorder_items + questions)
-- ============================================================================

INSERT INTO reading_reorder_items (
    id, exam_version_id, exam, title, paragraphs, topic, word_count,
    difficulty, tags, is_published
) VALUES
('ri-pte-ro-01', 'pte-2026-01', 'PTE', 'Peatlands and Long-Term Carbon Storage',
 $j$[
  {"label":"p3","text":"Peatlands cover only about three percent of the global land surface, yet they store more carbon than all the world's forests combined."},
  {"label":"p1","text":"This astonishing storage capacity is due to waterlogged conditions, which create an oxygen-depleted environment where dead vegetation cannot fully decay."},
  {"label":"p4","text":"Over thousands of years, these partially decomposed plant layers accumulate into thick deposits of peat, effectively locking carbon out of the atmosphere."},
  {"label":"p2","text":"However, when peatlands are drained for agriculture or commercial forestry, oxygen penetrates the peat, triggering rapid microbial decomposition and releasing massive quantities of carbon dioxide."}
 ]$j$::jsonb,
 'Environmental Science', 88, 'medium', ARRAY['PTE', 'Reading', 'Re-order Paragraphs'], TRUE),

('ri-pte-ro-02', 'pte-2026-01', 'PTE', 'The Advent of Movable Type Printing',
 $j$[
  {"label":"p4","text":"In mid-fifteenth-century Mainz, Johannes Gutenberg combined several existing technologies to create the first practical movable metal type printing system in Europe."},
  {"label":"p2","text":"His crucial innovation was not simply the screw press itself, but the invention of a hand mold that allowed metal type pieces to be cast with unprecedented uniformity and speed."},
  {"label":"p5","text":"Once these interchangeable letters could be arranged into pages, inked, and impressed onto paper, books could be produced in weeks rather than the years required by hand copying."},
  {"label":"p1","text":"The resulting plunge in book production costs democratized access to written knowledge, igniting the rapid dissemination of scientific and religious ideas across the continent."}
 ]$j$::jsonb,
 'History of Technology', 98, 'medium', ARRAY['PTE', 'Reading', 'Re-order Paragraphs'], TRUE),

('ri-pte-ro-03', 'pte-2026-01', 'PTE', 'Swarm Intelligence in Computational Optimization',
 $j$[
  {"label":"p2","text":"In nature, foraging ants find the shortest path between their nest and a food source by laying down volatile chemical trails called pheromones."},
  {"label":"p4","text":"Because ants traveling along shorter routes complete round trips more quickly, pheromone concentrations on those paths accumulate at a faster rate."},
  {"label":"p1","text":"Subsequent ants are biologically predisposed to follow trails with stronger chemical markings, reinforcing the optimal route while longer paths gradually evaporate."},
  {"label":"p3","text":"Computer scientists recognized the elegance of this decentralized behavior and adapted it into Ant Colony Optimization algorithms, which are now widely used to solve complex routing and logistical problems."}
 ]$j$::jsonb,
 'Computer Science', 95, 'medium', ARRAY['PTE', 'Reading', 'Re-order Paragraphs'], TRUE),

('ri-pte-ro-04', 'pte-2026-01', 'PTE', 'The Terrestrial Ancestry of Whales',
 $j$[
  {"label":"p3","text":"For centuries, naturalists were perplexed by the aquatic adaptations of whales and dolphins, wondering how mammals could have originated in open oceans."},
  {"label":"p1","text":"Fossil discoveries in Pakistan and India during the late twentieth century revealed that cetaceans actually evolved from four-legged terrestrial artiodactyls that lived fifty million years ago."},
  {"label":"p4","text":"Early transitional forms like Pakicetus still possessed functional limbs for walking on land, though their ear structures already showed adaptations for hearing underwater."},
  {"label":"p2","text":"Over successive epochs, descendants developed streamlined bodies, webbed paddle-like limbs, and nostrils migrated to the top of the skull, completing the transition to obligate marine life."}
 ]$j$::jsonb,
 'Evolutionary Biology', 92, 'medium', ARRAY['PTE', 'Reading', 'Re-order Paragraphs'], TRUE),

('ri-pte-ro-05', 'pte-2026-01', 'PTE', 'Extreme Ultraviolet Lithography in Semiconductor Manufacturing',
 $j$[
  {"label":"p4","text":"As semiconductor manufacturers pushed to pack billions of transistors onto a single silicon chip, conventional optical lithography reached the physical diffraction limit of visible light."},
  {"label":"p1","text":"To etch features smaller than ten nanometers, engineers had to shift to extreme ultraviolet (EUV) light, which has an extremely short wavelength of just 13.5 nanometers."},
  {"label":"p3","text":"Generating this extreme light requires firing high-powered lasers at microscopic droplets of molten tin fifty thousand times per second inside a vacuum chamber."},
  {"label":"p2","text":"Because EUV radiation is absorbed by almost all matter, including ordinary glass lenses, the entire optical system must rely on specialized multilayer mirrors to focus the beam onto the silicon wafer."}
 ]$j$::jsonb,
 'Applied Physics', 104, 'medium', ARRAY['PTE', 'Reading', 'Re-order Paragraphs'], TRUE);

-- QUESTIONS FOR TASK 3 (reorder-paragraphs)
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, time_limit_seconds, points,
    difficulty, tags, reorder_item_id
)
SELECT
    'q-' || i.id,
    i.exam_version_id, i.exam, ARRAY['PTE'], 'reading', 'reorder-paragraphs', 'Re-order Paragraphs',
    i.title,
    'The text boxes below have been placed in a random order. Restore the original order.',
    (SELECT jsonb_agg(jsonb_build_object('id', e->>'label', 'text', e->>'text'))
       FROM jsonb_array_elements(i.paragraphs) e),
    (SELECT jsonb_agg(e->>'label')
       FROM jsonb_array_elements(i.paragraphs) e),
    'Find the standalone topic sentence that introduces the subject without pronouns or backward references. Track chronological or cause-effect connectors to place subsequent boxes.',
    240,
    jsonb_array_length(i.paragraphs) - 1,
    i.difficulty,
    i.tags,
    i.id
FROM reading_reorder_items i
WHERE i.id IN ('ri-pte-ro-01', 'ri-pte-ro-02', 'ri-pte-ro-03', 'ri-pte-ro-04', 'ri-pte-ro-05');

-- ============================================================================
-- 4. PASSAGES FOR TASK 4: fill-in-blanks-r (5 passages, passage_display = 'hidden')
-- ============================================================================

INSERT INTO reading_passages (
    id, exam_version_id, title, subtitle, paragraphs, sources,
    word_count, difficulty, topic, tags, is_published, passage_slot
) VALUES
('rp-pte-fbr-01', 'pte-2026-01', 'Synaptic Plasticity and Learning', 'Dynamic restructuring of neural circuits',
 $j$[{"label":"A","text":"For many decades, neuroscientists believed that the adult mammalian brain was structurally fixed and incapable of generating new neural pathways. However, groundbreaking research has demonstrated that the brain retains remarkable plasticity throughout life. Through a mechanism known as synaptic plasticity, the strength of connections between neurons can be dynamically altered in response to learning, novel experiences, or recovery from injury. When specific neural circuits are repeatedly stimulated, the synapses strengthen, facilitating faster and more efficient signal transmission. This continuous reorganization proves that cognitive training and physical rehabilitation can induce tangible anatomical changes in neural circuitry."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Neuroscience', ARRAY['PTE', 'Reading', 'Fill in the Blanks R'], TRUE, 'custom'),

('rp-pte-fbr-02', 'pte-2026-01', 'Ice Sheet Dynamics and Calving', 'Mechanical fracture and sea level implications',
 $j$[{"label":"A","text":"Glaciers and polar ice sheets are not static masses of ice, but rather dynamic systems that flow under their own weight toward the sea. As an outlet glacier reaches the ocean coastline, the terminus often becomes buoyant, forming an expansive floating ice shelf. Stresses induced by tidal fluctuations and internal fractures cause colossal blocks of ice to detach and crash into the sea in an event known as calving. While the detachment of already floating ice does not directly elevate sea level, the loss of buttressing shelves accelerates the discharge of grounded ice from inland ice sheets. Monitoring these calving events via satellite radar provides scientists with vital data to forecast future rates of global coastal inundation."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Glaciology', ARRAY['PTE', 'Reading', 'Fill in the Blanks R'], TRUE, 'custom'),

('rp-pte-fbr-03', 'pte-2026-01', 'Declining Pollinators and Agricultural Stability', 'Ecological and nutritional risks of pollinator loss',
 $j$[{"label":"A","text":"Insect pollinators, particularly wild and domesticated bees, play an indispensable role in maintaining terrestrial biodiversity and agricultural productivity. Approximately three-quarters of leading global food crop species depend, at least in part, on animal pollination for optimal fruit and seed yield. In recent years, however, entomologists have documented alarming declines in pollinator populations worldwide. This widespread loss is driven by a combination of habitat fragmentation, pesticide exposure, and emerging pathogens. If these downward trends continue unabated, the reduced availability of pollination services could severely compromise crop yields and undermine the stability of global nutritional security, particularly for micronutrient-dense crops such as fruits, vegetables, and legumes."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Ecology & Agriculture', ARRAY['PTE', 'Reading', 'Fill in the Blanks R'], TRUE, 'custom'),

('rp-pte-fbr-04', 'pte-2026-01', 'Decentralized Energy and Electrical Microgrids', 'Enhancing resilience with localized generation and storage',
 $j$[{"label":"A","text":"Traditional electrical grids rely on centralized power generation facilities that transmit high-voltage electricity across extensive distances to consumer centers. While this architecture has powered industrial development for over a century, it suffers from significant transmission losses and vulnerability to localized disruptions. To enhance energy resilience, many communities are deploying decentralized microgrids that integrate local renewable generation, such as rooftop solar panels and small wind turbines, with advanced battery storage systems. A microgrid can operate autonomously from the main utility grid during power outages, ensuring that critical facilities maintain an uninterrupted supply of electricity. Furthermore, smart digital controllers can optimize power flows in real time, stabilizing overall grid reliability."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Energy Engineering', ARRAY['PTE', 'Reading', 'Fill in the Blanks R'], TRUE, 'custom'),

('rp-pte-fbr-05', 'pte-2026-01', 'The Exceptional Properties of Carbon Nanotubes', 'Mechanical strength and electrical conductivity at the nanoscale',
 $j$[{"label":"A","text":"Discovered in the early 1990s, carbon nanotubes are cylindrical molecules composed entirely of rolled graphene sheets with diameters measured in nanometers. Due to the strength of the sp2 covalent bonds between adjacent carbon atoms, these nanostructures exhibit extraordinary tensile strength, estimated to be over one hundred times greater than that of high-strength structural steel at a fraction of the weight. In addition to their mechanical prowess, nanotubes possess remarkable electrical conductivity, allowing electrons to travel along the tube axis with minimal resistance. Consequently, material scientists are actively investigating their incorporation into lightweight aerospace composites, flexible electronic displays, and advanced battery electrodes, aiming to pioneer a new generation of high-performance materials."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Materials Science', ARRAY['PTE', 'Reading', 'Fill in the Blanks R'], TRUE, 'custom');

-- GROUPS FOR TASK 4 (with word bank in resources)
INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds
) VALUES
('g-pte-fbr-01', 'rp-pte-fbr-01', 1, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"plasticity"},
  {"label":"w2","text":"altered"},
  {"label":"w3","text":"stimulated"},
  {"label":"w4","text":"circuitry"},
  {"label":"w5","text":"stagnant"},
  {"label":"w6","text":"suppressed"}
 ]$j$::jsonb, 'hidden', FALSE, 120),

('g-pte-fbr-02', 'rp-pte-fbr-02', 1, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"buoyant"},
  {"label":"w2","text":"elevate"},
  {"label":"w3","text":"forecast"},
  {"label":"w4","text":"inundation"},
  {"label":"w5","text":"rigid"},
  {"label":"w6","text":"dismantle"}
 ]$j$::jsonb, 'hidden', FALSE, 120),

('g-pte-fbr-03', 'rp-pte-fbr-03', 1, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"yield"},
  {"label":"w2","text":"compromise"},
  {"label":"w3","text":"stability"},
  {"label":"w4","text":"legumes"},
  {"label":"w5","text":"prosper"},
  {"label":"w6","text":"artificial"}
 ]$j$::jsonb, 'hidden', FALSE, 120),

('g-pte-fbr-04', 'rp-pte-fbr-04', 1, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"disruptions"},
  {"label":"w2","text":"supply"},
  {"label":"w3","text":"stabilizing"},
  {"label":"w4","text":"reliability"},
  {"label":"w5","text":"abandoning"},
  {"label":"w6","text":"redundant"}
 ]$j$::jsonb, 'hidden', FALSE, 120),

('g-pte-fbr-05', 'rp-pte-fbr-05', 1, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"strength"},
  {"label":"w2","text":"conductivity"},
  {"label":"w3","text":"pioneer"},
  {"label":"w4","text":"materials"},
  {"label":"w5","text":"fragility"},
  {"label":"w6","text":"diminish"}
 ]$j$::jsonb, 'hidden', FALSE, 120);

-- QUESTIONS FOR TASK 4
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, blanks, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position
) VALUES
('q-pte-fbr-01', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'Synaptic Plasticity and Learning',
 'Drag a word from the box into each gap.',
 'For many decades, neuroscientists believed that the adult mammalian brain was structurally fixed and incapable of generating new neural pathways. However, groundbreaking research has demonstrated that the brain retains remarkable [[b1]] throughout life. Through a mechanism known as synaptic plasticity, the strength of connections between neurons can be dynamically [[b2]] in response to learning, novel experiences, or recovery from injury. When specific neural circuits are repeatedly [[b3]], the synapses strengthen, facilitating faster and more efficient signal transmission. This continuous reorganization proves that cognitive training and physical rehabilitation can induce tangible anatomical changes in neural [[b4]].',
 $j$[
   {"id":"b1","correctAnswer":"plasticity"},
   {"id":"b2","correctAnswer":"altered"},
   {"id":"b3","correctAnswer":"stimulated"},
   {"id":"b4","correctAnswer":"circuitry"}
 ]$j$::jsonb,
 '"plasticity" refers to the brain adaptability; connections are dynamically "altered"; neural circuits are repeatedly "stimulated"; changes occur in neural "circuitry". "stagnant" and "suppressed" are unused distractors.',
 120, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbr-01', 'g-pte-fbr-01', 1),

('q-pte-fbr-02', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'Ice Sheet Dynamics and Calving',
 'Drag a word from the box into each gap.',
 'Glaciers and polar ice sheets are not static masses of ice, but rather dynamic systems that flow under their own weight toward the sea. As an outlet glacier reaches the ocean coastline, the terminus often becomes [[b1]], forming an expansive floating ice shelf. Stresses induced by tidal fluctuations and internal fractures cause colossal blocks of ice to detach and crash into the sea in an event known as calving. While the detachment of already floating ice does not directly [[b2]] sea level, the loss of buttressing shelves accelerates the discharge of grounded ice from inland ice sheets. Monitoring these calving events via satellite radar provides scientists with vital data to [[b3]] future rates of global coastal [[b4]].',
 $j$[
   {"id":"b1","correctAnswer":"buoyant"},
   {"id":"b2","correctAnswer":"elevate"},
   {"id":"b3","correctAnswer":"forecast"},
   {"id":"b4","correctAnswer":"inundation"}
 ]$j$::jsonb,
 'Terminus floats so it is "buoyant"; does not directly "elevate" sea level; data helps "forecast" rates of coastal "inundation" (flooding). "rigid" and "dismantle" are distractors.',
 120, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbr-02', 'g-pte-fbr-02', 1),

('q-pte-fbr-03', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'Declining Pollinators and Agricultural Stability',
 'Drag a word from the box into each gap.',
 'Insect pollinators, particularly wild and domesticated bees, play an indispensable role in maintaining terrestrial biodiversity and agricultural productivity. Approximately three-quarters of leading global food crop species depend, at least in part, on animal pollination for optimal fruit and seed [[b1]]. In recent years, however, entomologists have documented alarming declines in pollinator populations worldwide. This widespread loss is driven by a combination of habitat fragmentation, pesticide exposure, and emerging pathogens. If these downward trends continue unabated, the reduced availability of pollination services could severely [[b2]] crop yields and undermine the [[b3]] of global nutritional security, particularly for micronutrient-dense crops such as fruits, vegetables, and [[b4]].',
 $j$[
   {"id":"b1","correctAnswer":"yield"},
   {"id":"b2","correctAnswer":"compromise"},
   {"id":"b3","correctAnswer":"stability"},
   {"id":"b4","correctAnswer":"legumes"}
 ]$j$::jsonb,
 'Optimal fruit and seed "yield"; could severely "compromise" yields; undermine the "stability" of food security; crops such as "legumes". "prosper" and "artificial" are distractors.',
 120, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbr-03', 'g-pte-fbr-03', 1),

('q-pte-fbr-04', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'Decentralized Energy and Electrical Microgrids',
 'Drag a word from the box into each gap.',
 'Traditional electrical grids rely on centralized power generation facilities that transmit high-voltage electricity across extensive distances to consumer centers. While this architecture has powered industrial development for over a century, it suffers from significant transmission losses and vulnerability to localized [[b1]]. To enhance energy resilience, many communities are deploying decentralized microgrids that integrate local renewable generation, such as rooftop solar panels and small wind turbines, with advanced battery storage systems. A microgrid can operate autonomously from the main utility grid during power outages, ensuring that critical facilities maintain an uninterrupted [[b2]] of electricity. Furthermore, smart digital controllers can optimize power flows in real time, [[b3]] overall grid [[b4]].',
 $j$[
   {"id":"b1","correctAnswer":"disruptions"},
   {"id":"b2","correctAnswer":"supply"},
   {"id":"b3","correctAnswer":"stabilizing"},
   {"id":"b4","correctAnswer":"reliability"}
 ]$j$::jsonb,
 'Vulnerability to localized "disruptions"; uninterrupted "supply"; "stabilizing" overall grid "reliability". "abandoning" and "redundant" are distractors.',
 120, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbr-04', 'g-pte-fbr-04', 1),

('q-pte-fbr-05', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'The Exceptional Properties of Carbon Nanotubes',
 'Drag a word from the box into each gap.',
 'Discovered in the early 1990s, carbon nanotubes are cylindrical molecules composed entirely of rolled graphene sheets with diameters measured in nanometers. Due to the strength of the sp2 covalent bonds between adjacent carbon atoms, these nanostructures exhibit extraordinary tensile [[b1]], estimated to be over one hundred times greater than that of high-strength structural steel at a fraction of the weight. In addition to their mechanical prowess, nanotubes possess remarkable electrical [[b2]], allowing electrons to travel along the tube axis with minimal resistance. Consequently, material scientists are actively investigating their incorporation into lightweight aerospace composites, flexible electronic displays, and advanced battery electrodes, aiming to [[b3]] a new generation of high-performance [[b4]].',
 $j$[
   {"id":"b1","correctAnswer":"strength"},
   {"id":"b2","correctAnswer":"conductivity"},
   {"id":"b3","correctAnswer":"pioneer"},
   {"id":"b4","correctAnswer":"materials"}
 ]$j$::jsonb,
 'Tensile "strength"; electrical "conductivity"; aim to "pioneer" high-performance "materials". "fragility" and "diminish" are distractors.',
 120, 4, 'medium', ARRAY['PTE', 'Reading', 'Fill in the Blanks'],
 'rp-pte-fbr-05', 'g-pte-fbr-05', 1);

-- ============================================================================
-- 5. PASSAGES FOR TASK 5: reading-mcq-single (5 passages, passage_display = 'full')
-- ============================================================================

INSERT INTO reading_passages (
    id, exam_version_id, title, subtitle, paragraphs, sources,
    word_count, difficulty, topic, tags, is_published, passage_slot
) VALUES
('rp-pte-mcqs-01', 'pte-2026-01', 'Chemosynthetic Life at Hydrothermal Vents', 'Biotic ecosystems decoupled from solar radiation',
 $j$[{"label":"A","text":"Prior to the discovery of deep-sea hydrothermal vents in 1977 along the Galápagos Rift, biologists assumed that all terrestrial life depended fundamentally on solar radiation driving photosynthesis. Vent ecosystems, however, thrive in absolute darkness under crushing hydrostatic pressures miles beneath the ocean surface. Instead of sunlight, the biological foundation of these communities is formed by chemolithoautotrophic bacteria. These extremophiles metabolize hydrogen sulfide spewing from volcanic fissures to synthesize organic matter. Giant tube worms, vent shrimp, and blind crabs congregate around these mineral chimneys, forming dense biotic oases decoupled from the sun. The existence of chemosynthetic communities fundamentally expanded our understanding of life's ecological boundaries and revolutionized astrobiological hypotheses regarding subsurface oceans on icy moons like Europa."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Oceanography', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqs-02', 'pte-2026-01', 'Principles of Quantum Key Distribution', 'Physics-based cryptographic security protocols',
 $j$[{"label":"A","text":"Modern digital communications rely on public-key cryptography algorithms, such as RSA, whose security depends on the mathematical difficulty of factoring large integers. However, fault-tolerant quantum computers threaten to render these barriers obsolete via Shor's algorithm, which solves factorization in polynomial time. To counter this, cryptographers developed Quantum Key Distribution (QKD), notably the BB84 protocol. Unlike classical encryption, QKD derives its security from quantum mechanics rather than mathematical complexity. Cryptographic keys are transmitted via single photons in superposed quantum states. According to the Heisenberg uncertainty principle and the no-cloning theorem, any eavesdropper measuring the photon stream disturbs their quantum states, creating detectable errors. The legitimate parties can immediately identify the intrusion and discard the compromised key before sensitive data is sent."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Quantum Physics', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqs-03', 'pte-2026-01', 'Linguistic Framing and Spatial Orientation', 'Absolute directional reference in indigenous languages',
 $j$[{"label":"A","text":"In most Western languages, speakers employ an egocentric spatial reference system, describing locations relative to their own bodily perspective (e.g., 'to my left' or 'behind me'). In contrast, certain indigenous languages, such as Guugu Yimithirr in Queensland, rely exclusively on absolute geocentric coordinates corresponding to cardinal directions (north, south, east, west). A Guugu Yimithirr speaker never describes an object as being on their 'left', but rather instructs someone to move 'to the northwest' or warns of an insect on a listener's 'south leg'. Psycholinguistic experiments show that native speakers maintain an internal compass at all times, instinctively knowing their geographical orientation even inside windowless rooms. This phenomenon provides compelling support for linguistic relativity, showing that language structures can habituate the mind to attend to specific dimensions of the physical environment."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Linguistics & Cognition', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqs-04', 'pte-2026-01', 'The Durability of Roman Pozzolanic Concrete', 'Self-healing mineral chemistry in ancient marine engineering',
 $j$[{"label":"A","text":"While modern Portland cement structures frequently deteriorate within decades in marine environments, ancient Roman breakwaters remain intact after two millennia of tidal pounding. Recent material analysis reveals that the secret of Roman opus caementicium lies in its mineralogical composition. Roman builders mixed slaked lime with volcanic ash from Pozzuoli, creating a pozzolanic mortar containing millimeter-sized clasts of unslaked lime known as 'lime clasts'. When microcracks formed within the concrete over time, percolating seawater dissolved these calcium-rich clasts. The resulting solution reacted with volcanic minerals to precipitate tobermorite and phillipsite crystals directly across the fissures. Rather than degrading the material, chemical interaction with seawater triggered an active self-healing mechanism that structurally reinforced the concrete with age."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Materials History', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom'),

('rp-pte-mcqs-05', 'pte-2026-01', 'Governing Common-Pool Resources', 'Elinor Ostrom and self-governing community institutions',
 $j$[{"label":"A","text":"In his 1968 essay, Garrett Hardin formulated 'the tragedy of the commons', positing that individuals acting in their own self-interest will inevitably overexploit and deplete shared finite resources like pastures, fisheries, and aquifers. Hardin asserted that the only remedies were either state regulation or complete privatization. However, political economist Elinor Ostrom challenged this dichotomy through extensive field studies of communities managing common-pool resources worldwide. Ostrom documented that local users frequently craft enduring community governance institutions that prevent environmental degradation without relying on central state control or private ownership. By establishing clear boundaries, participatory rule-making, graduated sanctions for violators, and accessible conflict resolution, local collectives have successfully managed shared irrigation networks and forests sustainably across centuries."}]$j$::jsonb,
 '[]'::jsonb, 0, 'medium', 'Political Economy', ARRAY['PTE', 'Reading', 'Multiple Choice'], TRUE, 'custom');

-- GROUPS FOR TASK 5
INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds
) VALUES
('g-pte-mcqs-01', 'rp-pte-mcqs-01', 1, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the passage and answer the question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 90),
('g-pte-mcqs-02', 'rp-pte-mcqs-02', 1, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the passage and answer the question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 90),
('g-pte-mcqs-03', 'rp-pte-mcqs-03', 1, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the passage and answer the question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 90),
('g-pte-mcqs-04', 'rp-pte-mcqs-04', 1, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the passage and answer the question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 90),
('g-pte-mcqs-05', 'rp-pte-mcqs-05', 1, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the passage and answer the question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 90);

-- QUESTIONS FOR TASK 5
INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position
) VALUES
('q-pte-mcqs-01', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Hydrothermal Vent Chemosynthesis',
 'What is the primary significance of hydrothermal vent ecosystems according to the text?',
 $j$[
   {"id":"A","text":"They proved that volcanic activity is responsible for heating ocean currents."},
   {"id":"B","text":"They demonstrated that complex biological ecosystems can exist independently of solar energy."},
   {"id":"C","text":"They confirmed that photosynthesis occurs at much deeper oceanic levels than previously thought."},
   {"id":"D","text":"They revealed that giant tube worms represent the evolutionary ancestors of modern marine life."}
 ]$j$::jsonb,
 '["B"]'::jsonb,
 'The text explains that before vents were found, biologists assumed all life depended on sunlight; vents showed communities thrive in darkness using chemosynthesis.',
 90, 1, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqs-01', 'g-pte-mcqs-01', 1),

('q-pte-mcqs-02', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Quantum Key Distribution',
 'According to the author, why is Quantum Key Distribution impervious to undetected interception?',
 $j$[
   {"id":"A","text":"It relies on mathematical algorithms that are too computationally complex for quantum machines."},
   {"id":"B","text":"Any physical attempt to measure the quantum carrier alters its state, alerting the legitimate parties."},
   {"id":"C","text":"It transmits cryptographic keys through shielded fiber-optic lines that physically block interception."},
   {"id":"D","text":"The photons travel faster than the speed of light, preventing recording devices from capturing them."}
 ]$j$::jsonb,
 '["B"]'::jsonb,
 'The text states that measuring the photon stream disturbs their delicate quantum states, creating detectable errors that reveal the eavesdropper.',
 90, 1, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqs-02', 'g-pte-mcqs-02', 1),

('q-pte-mcqs-03', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Spatial Orientation in Language',
 'The author mentions the linguistic practices of Guugu Yimithirr speakers primarily to illustrate:',
 $j$[
   {"id":"A","text":"How geographical terminology can lead to frequent confusion in everyday communication."},
   {"id":"B","text":"Why Western spatial reference systems are superior in dense urban environments."},
   {"id":"C","text":"How the structure of a language can influence spatial cognition and environmental awareness."},
   {"id":"D","text":"That indigenous languages lack the grammatical capacity to describe bodily movements."}
 ]$j$::jsonb,
 '["C"]'::jsonb,
 'The example demonstrates linguistic relativity: how language structures habituate speakers to maintain continuous awareness of cardinal directions.',
 90, 1, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqs-03', 'g-pte-mcqs-03', 1),

('q-pte-mcqs-04', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Roman Pozzolanic Concrete',
 'According to the passage, why do ancient Roman marine concrete structures survive so well in seawater?',
 $j$[
   {"id":"A","text":"Roman engineers added synthetic waterproofing seals to keep seawater away from the core."},
   {"id":"B","text":"The interaction between seawater and lime inclusions generates minerals that repair internal cracks."},
   {"id":"C","text":"The builders used solid granite blocks instead of cement wherever water contact occurred."},
   {"id":"D","text":"Portland cement was added in heavy proportions to prevent the formation of lime clasts."}
 ]$j$::jsonb,
 '["B"]'::jsonb,
 'Seawater dissolves lime clasts and reacts with volcanic ash to precipitate tobermorite crystals across fissures, triggering self-healing.',
 90, 1, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqs-04', 'g-pte-mcqs-04', 1),

('q-pte-mcqs-05', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Governing Common-Pool Resources',
 'What was Elinor Ostrom''s primary objection to Garrett Hardin''s thesis?',
 $j$[
   {"id":"A","text":"Hardin underestimated the total volume of natural resources available on Earth."},
   {"id":"B","text":"Privatization is always more environmentally destructive than state bureaucratic management."},
   {"id":"C","text":"Communities can develop effective self-governing arrangements to manage shared resources sustainably."},
   {"id":"D","text":"Self-interest plays no role in human decisions regarding common-pool resources."}
 ]$j$::jsonb,
 '["C"]'::jsonb,
 'Ostrom documented that local resource users craft enduring community institutions to govern shared resources without state control or privatization.',
 90, 1, 'medium', ARRAY['PTE', 'Reading', 'Multiple Choice'],
 'rp-pte-mcqs-05', 'g-pte-mcqs-05', 1);

-- ============================================================================
-- 6. UPDATE PTE READING MOCK BLUEPRINT TO REFLECT 1 ITEM PER INDEPENDENT TASK
-- ============================================================================

DELETE FROM reading_mock_blueprint_slots WHERE blueprint_id = 'bp-pte-reading';

UPDATE reading_mock_blueprints
   SET passage_count = 4,
       total_questions = 5,
       duration_minutes = 30
 WHERE id = 'bp-pte-reading';

INSERT INTO reading_mock_blueprint_slots
    (blueprint_id, position, ordinal, type_id, question_count, source) VALUES
    ('bp-pte-reading', 1, 1, 'fill-in-blanks-rw',    1, 'passage'),
    ('bp-pte-reading', 2, 1, 'reading-mcq-multiple', 1, 'passage'),
    ('bp-pte-reading', 3, 1, 'reorder-paragraphs',   1, 'reorder'),
    ('bp-pte-reading', 4, 1, 'fill-in-blanks-r',     1, 'passage'),
    ('bp-pte-reading', 5, 1, 'reading-mcq-single',   1, 'passage');

-- ============================================================================
-- 7. INTEGRITY GUARDS
-- ============================================================================

DO $$
DECLARE
    cnt INT;
    bad TEXT;
BEGIN
    -- Exactly 5 questions for each of the 5 PTE reading task types
    FOR bad IN SELECT unnest(ARRAY['fill-in-blanks-rw', 'reading-mcq-multiple', 'reorder-paragraphs', 'fill-in-blanks-r', 'reading-mcq-single'])
    LOOP
        SELECT count(*) INTO cnt
          FROM questions q
         WHERE q.skill = 'reading'
           AND q.type_id = bad
           AND q.is_published
           AND 'PTE' = ANY(q.supported_exams);
        IF cnt < 5 THEN
            RAISE EXCEPTION 'pte reading seed: % has % questions, need at least 5', bad, cnt;
        END IF;
    END LOOP;

    -- Gap-fill groups must hide the passage
    SELECT string_agg(id, ', ') INTO bad
      FROM reading_question_groups
     WHERE type_id IN ('fill-in-blanks-rw', 'fill-in-blanks-r')
       AND passage_display <> 'hidden';
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'pte reading seed: gap-fill groups must hide the passage: %', bad;
    END IF;

    -- Gap markers [[bN]] and blanks array length must match exactly
    SELECT string_agg(q.id, ', ') INTO bad
      FROM questions q
     WHERE q.type_id IN ('fill-in-blanks-rw', 'fill-in-blanks-r')
       AND q.context_passage IS NOT NULL
       AND (SELECT count(*) FROM regexp_matches(q.context_passage, '\[\[b[0-9]+\]\]', 'g'))
           <> jsonb_array_length(q.blanks);
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'pte reading seed: gap markers and blanks disagree in: %', bad;
    END IF;

    -- Re-order answer options must match box options
    SELECT string_agg(q.id, ', ') INTO bad
      FROM questions q
     WHERE q.reorder_item_id IS NOT NULL
       AND (SELECT count(*) FROM jsonb_array_elements_text(q.correct_answers) a
             WHERE NOT EXISTS (SELECT 1 FROM jsonb_array_elements(q.options) o
                                WHERE o->>'id' = a)) > 0;
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'pte reading seed: re-order answers name boxes that do not exist in: %', bad;
    END IF;
END $$;
