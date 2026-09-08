-- The proof that a passage is shared content.
--
-- Everything above this migration is mechanism: eligibility on the question, the
-- exam off the passage, composition from a blueprint. None of it is worth
-- anything until one passage actually carries both exams, and until this
-- migration none does — the four passages in the bank are three IELTS ones and
-- one PTE one, exactly as before.
--
-- So: PTE task sets are authored on the three passages that were IELTS-only, and
-- IELTS task sets on the passage that was PTE-only. No new passage text is
-- written. The gap-fill texts are rewrites of paragraphs that are already there,
-- which is what a PTE gap-fill is, and every question is answerable from the
-- passage it hangs on.
--
-- It is also what makes a PTE reading paper possible. bp-pte-reading needs four
-- distinct passages carrying 2 Reading & Writing gap-fills, 2 Reading gap-fills,
-- 4 single-answer and 2 multiple-answer questions between them; one passage with
-- seven questions on it could never fill that.

-- ---------------------------------------------------------------------------
-- A History of Chocolate: PTE tasks
-- ---------------------------------------------------------------------------

INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds) VALUES

('g-choc-9', 'rp-choc-01', 9, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180),

('g-choc-10', 'rp-choc-01', 10, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the text and answer the multiple-choice question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 120);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, blanks, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position) VALUES

('q-choc-101', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'Cacao before Europe',
 'Select the appropriate answer choice for each blank.',
 'Cacao was a drink long before it became a confection, and the Maya preparation of it would be almost [[b1]] today: roasted seeds ground with water, chilli and maize flour, then poured from vessel to vessel until a thick head of foam formed. Under the Aztecs the same seed served as money, and because it was money it was forged, the beans hollowed out and [[b2]] with earth. When the Spanish met the drink, what they mainly understood was that it needed fixing, and the sugar and cinnamon they stirred into it were [[b3]] to make it palatable to a European palate. For most of a century the recipe was [[b4]] closely enough that chocolate reached the French court as a curiosity rather than a commodity.',
 $j$[
  {"id":"b1","options":["unrecognisable","unremarkable","unreliable","unreasonable"],"correctAnswer":"unrecognisable"},
  {"id":"b2","options":["packed","pushed","pressed","plotted"],"correctAnswer":"packed"},
  {"id":"b3","options":["intended","invented","inverted","indented"],"correctAnswer":"intended"},
  {"id":"b4","options":["guarded","granted","gathered","glanced"],"correctAnswer":"guarded"}
 ]$j$::jsonb,
 'Each gap is settled by the clause around it: a preparation nobody would know today is unrecognisable, and a recipe kept among royalty and religious houses was guarded.',
 180, 4, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-choc-01', 'g-choc-9', 1),

('q-choc-102', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'The bean and the factory',
 'Select the appropriate answer choice for each blank.',
 'Industrialisation changed chocolate more in seventy years than the preceding three thousand had. Van Houten''s press [[b1]] the fat from the ground bean, leaving a powder that mixed cleanly with liquid. The tree itself was never industrialised in the same way: cacao will not grow in dry conditions, and its yields swing so sharply with rainfall that a single poor year [[b2]] the world price. Pods are cut by hand because they grow directly from the trunk and the older wood must not be [[b3]]. After harvesting, the beans are fermented for several days to [[b4]] their flavour, and no amount of later processing recovers what the fermentation heap did not make.',
 $j$[
  {"id":"b1","options":["separated","suspended","sweetened","surrounded"],"correctAnswer":"separated"},
  {"id":"b2","options":["moves","mends","marks","mixes"],"correctAnswer":"moves"},
  {"id":"b3","options":["damaged","delayed","darkened","divided"],"correctAnswer":"damaged"},
  {"id":"b4","options":["develop","deliver","detect","dissolve"],"correctAnswer":"develop"}
 ]$j$::jsonb,
 'The press took the cocoa butter out, so it separated it; a poor year moving the world price is what the passage says about yields.',
 180, 4, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-choc-01', 'g-choc-9', 2);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, source_paragraphs,
    time_limit_seconds, points, difficulty, tags, passage_id, group_id, group_position) VALUES

('q-choc-103', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 1',
 'According to the passage, why were cacao beans counterfeited under the Aztecs?',
 $j$[
  {"id":"A","text":"Because they were used as currency across much of Mesoamerica."},
  {"id":"B","text":"Because the Maya regarded them as a gift from the gods."},
  {"id":"C","text":"Because the Spanish had restricted their supply."},
  {"id":"D","text":"Because the pods were difficult to harvest by hand."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'The passage states the beans were forged because they were currency, which is the reason it gives rather than any of the other facts it also reports.',
 ARRAY['B'], 120, 1, 'easy', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-choc-01', 'g-choc-10', 1),

('q-choc-104', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 2',
 'What did van Houten''s 1828 invention make possible?',
 $j$[
  {"id":"A","text":"Separating cocoa butter from the roasted bean."},
  {"id":"B","text":"Casting a paste into a mould to form a solid bar."},
  {"id":"C","text":"Adding powdered milk to chocolate."},
  {"id":"D","text":"Grinding the particles below the threshold the tongue can feel."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'The other three are the achievements of Fry, Peter and Lindt, each named separately in the same paragraph.',
 ARRAY['E'], 120, 1, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-choc-01', 'g-choc-10', 2),

('q-choc-105', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 3',
 'What does the passage say about beans that have not been fermented?',
 $j$[
  {"id":"A","text":"They taste of almost nothing, and later processing cannot recover the flavour."},
  {"id":"B","text":"They are sweeter but less aromatic than fermented beans."},
  {"id":"C","text":"They are preferred for bulk confectionery fillings."},
  {"id":"D","text":"They must be roasted above 150 degrees Celsius to be usable."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'The passage is explicit that no amount of later processing recovers what the fermentation heap did not make.',
 ARRAY['F'], 120, 1, 'easy', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-choc-01', 'g-choc-10', 3),

('q-choc-106', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 4',
 'Which statement best describes the writer''s attitude to mass-produced chocolate?',
 $j$[
  {"id":"A","text":"They consider it a lesser thing than small-batch chocolate."},
  {"id":"B","text":"They regard the artisanal revival as mostly nostalgia."},
  {"id":"C","text":"They accept that shelf life is worth the loss of flavour."},
  {"id":"D","text":"They believe regulation has already settled the question."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'The writer calls industrial chocolate hard to defend and a lesser thing, and dismisses the charge of nostalgia rather than making it.',
 ARRAY['I'], 120, 1, 'hard', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-choc-01', 'g-choc-10', 4);

-- ---------------------------------------------------------------------------
-- The Return of Urban Beekeeping: PTE tasks
-- ---------------------------------------------------------------------------

-- The word bank is on the group, not on the question: one pool shared across
-- every gap in the task, with more words in it than there are gaps.
INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds) VALUES

('g-bees-9', 'rp-bees-01', 9, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"outlawed"},
  {"label":"w2","text":"trebled"},
  {"label":"w3","text":"variety"},
  {"label":"w4","text":"warmer"},
  {"label":"w5","text":"support"},
  {"label":"w6","text":"competitor"},
  {"label":"w7","text":"short"},
  {"label":"w8","text":"universal"},
  {"label":"w9","text":"encouraged"},
  {"label":"w10","text":"halved"},
  {"label":"w11","text":"shelter"},
  {"label":"w12","text":"optional"}
 ]$j$::jsonb,
 'hidden', FALSE, 180),

('g-bees-10', 'rp-bees-01', 10, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the text and answer the question by selecting all the correct responses. More than one response is correct.',
 '[]'::jsonb, 'full', TRUE, 150);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, blanks, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position) VALUES

('q-bees-101', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'Hives in the city',
 'Drag a word from the box into each gap.',
 'For most of the twentieth century keeping bees in cities was quietly [[b1]]. New York listed the honeybee among its prohibited animals until 2010, and within two years the number of registered hives in the five boroughs had [[b2]]. The urban hive works because of [[b3]]: a colony on a city roof forages across gardens, parks, street trees, railway embankments and window boxes, which between them flower from February to November. Cities are also [[b4]] than the countryside around them, and they carry a far lighter load of agricultural pesticide.',
 $j$[
  {"id":"b1","correctAnswer":"outlawed"},
  {"id":"b2","correctAnswer":"trebled"},
  {"id":"b3","correctAnswer":"variety"},
  {"id":"b4","correctAnswer":"warmer"}
 ]$j$::jsonb,
 'Each word is lifted from the passage: the practice was outlawed, hives trebled, the reason is variety, and the urban heat island makes cities warmer.',
 180, 4, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-bees-01', 'g-bees-9', 1),

('q-bees-102', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'Too many hives',
 'Drag a word from the box into each gap.',
 'Enthusiasm has produced its own problem. Hive density in some London boroughs passed ten colonies per square kilometre, well above the level the available forage can [[b1]]. A managed honeybee colony is a [[b2]] before it is anything else, and the wild bees that do most of the pollination lose out first when nectar runs [[b3]]. Disease is the other constraint: the varroa mite reached Europe in the 1970s and is now effectively [[b4]], so untreated colonies rarely last three years.',
 $j$[
  {"id":"b1","correctAnswer":"support"},
  {"id":"b2","correctAnswer":"competitor"},
  {"id":"b3","correctAnswer":"short"},
  {"id":"b4","correctAnswer":"universal"}
 ]$j$::jsonb,
 'The passage says forage cannot support the density, that a colony is a competitor first, that wild bees lose out when nectar runs short, and that varroa is now effectively universal.',
 180, 4, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-bees-01', 'g-bees-9', 2);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, source_paragraphs,
    time_limit_seconds, points, difficulty, tags, passage_id, group_id, group_position) VALUES

('q-bees-103', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Multiple Choice, Multiple Answers 1',
 'Which TWO reasons does the passage give for city colonies doing well?',
 $j$[
  {"id":"A","text":"Forage is available across a much longer season than on arable land."},
  {"id":"B","text":"Cities are warmer, and the heat island lifts winter minima."},
  {"id":"C","text":"Urban honey sells for more than rural honey."},
  {"id":"D","text":"Registration is compulsory in most boroughs."},
  {"id":"E","text":"Most urban beekeepers recover their costs within a season."}
 ]$j$::jsonb,
 '["A", "B"]'::jsonb,
 'Variety of forage and the warmth of the city are the two reasons the passage gives. It says the opposite of E, and says nothing about C.',
 ARRAY['B'], 150, 2, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-bees-01', 'g-bees-10', 1),

('q-bees-104', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Multiple Choice, Multiple Answers 2',
 'Which TWO statements about varroa treatment are supported by the passage?',
 $j$[
  {"id":"A","text":"Oxalic acid is applied when the colony is broodless in midwinter."},
  {"id":"B","text":"Thymol works only above about 15 degrees Celsius."},
  {"id":"C","text":"Urban colonies are not affected by the mite."},
  {"id":"D","text":"Untreated colonies typically survive for five years."},
  {"id":"E","text":"The timing of treatment makes little difference to the outcome."}
 ]$j$::jsonb,
 '["A", "B"]'::jsonb,
 'Both are stated directly. The passage says untreated colonies rarely last three years, and that a hive treated at the wrong point is treated for nothing.',
 ARRAY['G'], 150, 2, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-bees-01', 'g-bees-10', 2);

-- ---------------------------------------------------------------------------
-- The Long Road of Papermaking: PTE tasks
-- ---------------------------------------------------------------------------

INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds) VALUES

('g-paper-9', 'rp-paper-01', 9, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180),

('g-paper-10', 'rp-paper-01', 10, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the text and answer the multiple-choice question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 120),

('g-paper-11', 'rp-paper-01', 11, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the text and answer the question by selecting all the correct responses. More than one response is correct.',
 '[]'::jsonb, 'full', TRUE, 150);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, blanks, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position) VALUES

('q-paper-101', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'A dated invention',
 'Select the appropriate answer choice for each blank.',
 'Paper is dated, unusually for something so old, to a single year and a single official: in 105 CE Cai Lun presented to the Han emperor a material made from macerated bark, hemp waste, old rags and fishing nets. The date is almost certainly too [[b1]], since fragments predating it by two centuries have been excavated, but his contribution was real — he made the process [[b2]]. The craft then travelled west slowly and by war, and Europe was the [[b3]] of the old world to take it up. At Fabriano the mills added three things the Chinese had not, and the paper they made was better than what it [[b4]].',
 $j$[
  {"id":"b1","options":["late","early","loose","light"],"correctAnswer":"late"},
  {"id":"b2","options":["cheap","complex","central","curious"],"correctAnswer":"cheap"},
  {"id":"b3","options":["last","least","latest","largest"],"correctAnswer":"last"},
  {"id":"b4","options":["copied","created","carried","corrected"],"correctAnswer":"copied"}
 ]$j$::jsonb,
 'Older fragments make the traditional date too late; the passage says Cai Lun made the process cheap, that Europe was last, and that Fabriano paper was better than what it copied.',
 180, 4, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-paper-01', 'g-paper-9', 1),

('q-paper-102', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks', 'The end of the rag age',
 'Select the appropriate answer choice for each blank.',
 'For seven centuries European paper was made from linen and cotton rags, and the supply of rags was the [[b1]] on the whole industry. By the early nineteenth century the shortage was acute enough that chemists were paid to look at straw, nettles, thistles and, following a suggestion drawn from watching wasps build nests, wood. Two inventions ended the rag age, and within fifty years wood had [[b2]] rag almost entirely. The change carried a cost that took a century to show: wood pulp holds lignin, which yellows and [[b3]] in light, and the acidic sizing adopted in the same period slowly [[b4]] the cellulose until the paper snaps rather than folds.',
 $j$[
  {"id":"b1","options":["ceiling","corner","capital","crossing"],"correctAnswer":"ceiling"},
  {"id":"b2","options":["displaced","discounted","disguised","dispatched"],"correctAnswer":"displaced"},
  {"id":"b3","options":["embrittles","enriches","enlarges","embraces"],"correctAnswer":"embrittles"},
  {"id":"b4","options":["hydrolyses","harmonises","hesitates","highlights"],"correctAnswer":"hydrolyses"}
 ]$j$::jsonb,
 'Every gap is a word the passage itself uses: rags were the ceiling, wood displaced rag, lignin embrittles in light, and the acid hydrolyses the cellulose.',
 180, 4, 'hard', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-paper-01', 'g-paper-9', 2);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, source_paragraphs,
    time_limit_seconds, points, difficulty, tags, passage_id, group_id, group_position) VALUES

('q-paper-103', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 1',
 'What does the passage say about the traditional date given for the invention of paper?',
 $j$[
  {"id":"A","text":"It is probably too late, because older fragments have been excavated."},
  {"id":"B","text":"It is probably too early, because the histories were written later."},
  {"id":"C","text":"It is accurate, and confirmed by excavation."},
  {"id":"D","text":"It cannot be assessed, because no physical evidence survives."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'Fragments predating 105 CE by two centuries put the traditional date too late, though the passage still credits Cai Lun with making the process cheap.',
 ARRAY['A'], 120, 1, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-paper-01', 'g-paper-10', 1),

('q-paper-104', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 2',
 'According to the passage, what limited the European paper industry for seven centuries?',
 $j$[
  {"id":"A","text":"The supply of rags."},
  {"id":"B","text":"The availability of clean running water."},
  {"id":"C","text":"The cost of animal gelatine sizing."},
  {"id":"D","text":"Restrictions on exporting the technique."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'The passage calls the supply of rags the ceiling on the whole industry, which is why parishes collected them and some countries forbade their export.',
 ARRAY['D'], 120, 1, 'easy', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-paper-01', 'g-paper-10', 2),

('q-paper-105', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 3',
 'Why does the passage say that pulp economics are "geography before they are anything else"?',
 $j$[
  {"id":"A","text":"Because site quality changes how long a plantation takes to reach a given yield."},
  {"id":"B","text":"Because softwoods and hardwoods are grown on different continents."},
  {"id":"C","text":"Because mills must be built beside running water."},
  {"id":"D","text":"Because bleaching sequences run cooler in warm climates."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'The passage says a plantation on a poor site may take twice as long to reach the yield of one on a good site.',
 ARRAY['F'], 120, 1, 'hard', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-paper-01', 'g-paper-10', 3),

('q-paper-106', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 4',
 'What does the passage say about recycled fibre?',
 $j$[
  {"id":"A","text":"It shortens each time it is pulped, so new fibre must keep entering the system."},
  {"id":"B","text":"It can be recycled indefinitely if the sheet is kept acid-free."},
  {"id":"C","text":"It is governed by a single international standard."},
  {"id":"D","text":"It is unsuitable for board once it has been pulped twice."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'Five to seven pulpings, after which the fibre is fit only for board, is why the passage says a recycled sheet always depends on new fibre somewhere.',
 ARRAY['I'], 120, 1, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-paper-01', 'g-paper-10', 4),

('q-paper-107', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Multiple Choice, Multiple Answers 1',
 'Which TWO of the following did the mills at Fabriano add to the process?',
 $j$[
  {"id":"A","text":"Water-powered stamping hammers."},
  {"id":"B","text":"The wire watermark."},
  {"id":"C","text":"A machine that formed paper as a continuous web."},
  {"id":"D","text":"Grinding logs against a wet stone to make pulp."},
  {"id":"E","text":"Cooking chips in sodium hydroxide and sodium sulphide."}
 ]$j$::jsonb,
 '["A", "B"]'::jsonb,
 'Fabriano added hammers, gelatine sizing and the watermark. The other three belong to the Fourdriniers, Keller and the kraft process, centuries later.',
 ARRAY['C'], 150, 2, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-paper-01', 'g-paper-11', 1),

('q-paper-108', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Multiple Choice, Multiple Answers 2',
 'Which TWO statements about pulping are supported by the passage?',
 $j$[
  {"id":"A","text":"The kraft process cooks chips at about 170 degrees Celsius."},
  {"id":"B","text":"Mechanical pulping keeps far more of the wood than kraft does."},
  {"id":"C","text":"Bleaching runs hotter than kraft cooking."},
  {"id":"D","text":"Kraft pulping uses sodium hydroxide alone."},
  {"id":"E","text":"Every ten degrees of extra heat saves energy."}
 ]$j$::jsonb,
 '["A", "B"]'::jsonb,
 'Bleaching runs cooler, kraft uses sodium sulphide as well, and the passage says every ten degrees costs energy rather than saving it.',
 ARRAY['G'], 150, 2, 'hard', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-paper-01', 'g-paper-11', 2);

-- ---------------------------------------------------------------------------
-- The Coming of Standard Time: IELTS tasks, and two more PTE sets
-- ---------------------------------------------------------------------------

-- The proof in the other direction. This passage was authored for PTE and has
-- carried nothing else; the True/False and Matching Information sets below are
-- IELTS tasks on the same text, and neither exam's questions are visible to the
-- other.
--
-- The two extra PTE sets are so that this passage can fill a Reading gap-fill or
-- a single-answer section on its own, which a paper needs when the section
-- before it has already taken the passage that would otherwise have done.

INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds) VALUES

('g-time-5', 'rp-time-01', 5, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"authority"},
  {"label":"w2","text":"adopted"},
  {"label":"w3","text":"abstained"},
  {"label":"w4","text":"borders"},
  {"label":"w5","text":"agreed"},
  {"label":"w6","text":"declined"},
  {"label":"w7","text":"longitude"},
  {"label":"w8","text":"refused"}
 ]$j$::jsonb,
 'hidden', FALSE, 180),

('g-time-6', 'rp-time-01', 6, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the text and answer the multiple-choice question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 120),

('g-time-7', 'rp-time-01', 7, 'reading-true-false', 'True / False / Not Given',
 'Do the following statements agree with the information in the passage? Answer True, False or Not Given.',
 '[]'::jsonb, 'full', TRUE, 420),

('g-time-8', 'rp-time-01', 8, 'reading-matching-information', 'Matching Information',
 'The passage has five paragraphs, A to E. Which paragraph contains each of the following?',
 '[]'::jsonb, 'full', TRUE, 300);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    context_passage, blanks, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position) VALUES

('q-time-101', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks', 'Zones and meridians',
 'Drag a word from the box into each gap.',
 'The American railroads did not wait for governments. On a Sunday in November 1883 they divided the continent into five zones on their own [[b1]], and most towns [[b2]] railroad time because the alternative was to be out of step with everything that arrived. Governments caught up the following year at Washington, where France [[b3]] and went on reckoning from Paris until 1911. The conference settled the meridian but left the zones to each country, which is why the map of world time follows [[b4]] rather than neat vertical stripes.',
 $j$[
  {"id":"b1","correctAnswer":"authority"},
  {"id":"b2","correctAnswer":"adopted"},
  {"id":"b3","correctAnswer":"abstained"},
  {"id":"b4","correctAnswer":"borders"}
 ]$j$::jsonb,
 'Each word appears in the passage: the railroads acted on their own authority, towns adopted railroad time, France abstained, and the map follows borders.',
 180, 4, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-time-01', 'g-time-5', 1);

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, source_paragraphs,
    time_limit_seconds, points, difficulty, tags, passage_id, group_id, group_position) VALUES

('q-time-102', 'pte-2026-01', 'PTE', ARRAY['PTE'], 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 4',
 'Why does the passage say Greenwich was chosen as the prime meridian?',
 $j$[
  {"id":"A","text":"Most of the world shipping already navigated on charts drawn from it."},
  {"id":"B","text":"It was the only meridian passing through no inhabited country."},
  {"id":"C","text":"Britain had legislated for it in 1880."},
  {"id":"D","text":"The American railroads had already adopted it in 1883."}
 ]$j$::jsonb,
 '["A"]'::jsonb,
 'The passage calls the choice practical rather than political, and gives the shipping charts as the reason: any other choice would have meant redrawing them.',
 ARRAY['D'], 120, 1, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-time-01', 'g-time-6', 1);

-- The IELTS sets on the same passage.

INSERT INTO questions (
    id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, source_paragraphs,
    time_limit_seconds, points, difficulty, tags, passage_id, group_id, group_position) VALUES

('q-time-201', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-true-false',
 'True / False / Not Given', 'True / False / Not Given 1',
 'Before the railways, clocks in Bristol and London showed the same time.',
 $j$[{"id":"TRUE","text":"True"},{"id":"FALSE","text":"False"},{"id":"NOT_GIVEN","text":"Not Given"}]$j$::jsonb,
 '["FALSE"]'::jsonb,
 'The passage gives Bristol as about ten minutes behind London.',
 ARRAY['A'], 60, 1, 'easy', ARRAY['IELTS Reading', 'True / False / Not Given'],
 'rp-time-01', 'g-time-7', 1),

('q-time-202', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-true-false',
 'True / False / Not Given', 'True / False / Not Given 2',
 'Coach guards set their watches to gain time on the westward journey.',
 $j$[{"id":"TRUE","text":"True"},{"id":"FALSE","text":"False"},{"id":"NOT_GIVEN","text":"Not Given"}]$j$::jsonb,
 '["TRUE"]'::jsonb,
 'Stated directly: watches were deliberately set to gain along the westward route and lose on the way back.',
 ARRAY['A'], 60, 1, 'medium', ARRAY['IELTS Reading', 'True / False / Not Given'],
 'rp-time-01', 'g-time-7', 2),

('q-time-203', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-true-false',
 'True / False / Not Given', 'True / False / Not Given 3',
 'The Great Western Railway adopted London time before Parliament legislated on the question.',
 $j$[{"id":"TRUE","text":"True"},{"id":"FALSE","text":"False"},{"id":"NOT_GIVEN","text":"Not Given"}]$j$::jsonb,
 '["TRUE"]'::jsonb,
 'The GWR adopted London time in 1840; Parliament did not legislate until 1880.',
 ARRAY['B'], 60, 1, 'easy', ARRAY['IELTS Reading', 'True / False / Not Given'],
 'rp-time-01', 'g-time-7', 3),

('q-time-204', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-true-false',
 'True / False / Not Given', 'True / False / Not Given 4',
 'Station clocks with two minute hands were removed within a year of being installed.',
 $j$[{"id":"TRUE","text":"True"},{"id":"FALSE","text":"False"},{"id":"NOT_GIVEN","text":"Not Given"}]$j$::jsonb,
 '["NOT_GIVEN"]'::jsonb,
 'The passage says the two-handed clocks existed for some years and satisfied nobody, but never says when or how they were removed.',
 ARRAY['B'], 60, 1, 'hard', ARRAY['IELTS Reading', 'True / False / Not Given'],
 'rp-time-01', 'g-time-7', 4),

('q-time-205', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-true-false',
 'True / False / Not Given', 'True / False / Not Given 5',
 'Sandford Fleming began campaigning for time zones after missing a train.',
 $j$[{"id":"TRUE","text":"True"},{"id":"FALSE","text":"False"},{"id":"NOT_GIVEN","text":"Not Given"}]$j$::jsonb,
 '["TRUE"]'::jsonb,
 'He missed a train in Ireland because a timetable printed the hour ambiguously, and spent much of the following decade arguing for worldwide zones.',
 ARRAY['C'], 60, 1, 'medium', ARRAY['IELTS Reading', 'True / False / Not Given'],
 'rp-time-01', 'g-time-7', 5),

('q-time-206', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-true-false',
 'True / False / Not Given', 'True / False / Not Given 6',
 'France adopted Greenwich time immediately after the 1884 conference.',
 $j$[{"id":"TRUE","text":"True"},{"id":"FALSE","text":"False"},{"id":"NOT_GIVEN","text":"Not Given"}]$j$::jsonb,
 '["FALSE"]'::jsonb,
 'France abstained and went on reckoning from Paris until 1911, twenty-seven years later.',
 ARRAY['D'], 60, 1, 'easy', ARRAY['IELTS Reading', 'True / False / Not Given'],
 'rp-time-01', 'g-time-7', 6),

('q-time-207', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-matching-information',
 'Matching Information', 'Matching Information 1',
 'An explanation of how disagreement about the time could put two trains on the same rails.',
 $j$[{"id":"A","text":"Paragraph A"},{"id":"B","text":"Paragraph B"},{"id":"C","text":"Paragraph C"},{"id":"D","text":"Paragraph D"},{"id":"E","text":"Paragraph E"}]$j$::jsonb,
 '["B"]'::jsonb,
 'Paragraph B is where single-track working and the risk of routing two trains onto the same rails appear.',
 ARRAY['B'], 60, 1, 'medium', ARRAY['IELTS Reading', 'Matching Information'],
 'rp-time-01', 'g-time-8', 1),

('q-time-208', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-matching-information',
 'Matching Information', 'Matching Information 2',
 'The date on which railway companies divided a continent into five zones without legal authority.',
 $j$[{"id":"A","text":"Paragraph A"},{"id":"B","text":"Paragraph B"},{"id":"C","text":"Paragraph C"},{"id":"D","text":"Paragraph D"},{"id":"E","text":"Paragraph E"}]$j$::jsonb,
 '["C"]'::jsonb,
 'Paragraph C gives the Sunday in November 1883 when the American railroads switched on their own authority.',
 ARRAY['C'], 60, 1, 'easy', ARRAY['IELTS Reading', 'Matching Information'],
 'rp-time-01', 'g-time-8', 2),

('q-time-209', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-matching-information',
 'Matching Information', 'Matching Information 3',
 'A commercial reason for a decision that might have been expected to be political.',
 $j$[{"id":"A","text":"Paragraph A"},{"id":"B","text":"Paragraph B"},{"id":"C","text":"Paragraph C"},{"id":"D","text":"Paragraph D"},{"id":"E","text":"Paragraph E"}]$j$::jsonb,
 '["D"]'::jsonb,
 'Paragraph D explains that Greenwich was chosen because two-thirds of world shipping tonnage already navigated on charts drawn from it.',
 ARRAY['D'], 60, 1, 'medium', ARRAY['IELTS Reading', 'Matching Information'],
 'rp-time-01', 'g-time-8', 3),

('q-time-210', 'ielts-2026-01', 'IELTS', ARRAY['IELTS'], 'reading', 'reading-matching-information',
 'Matching Information', 'Matching Information 4',
 'A contrast between what time meant to a community before the railways and what it means now.',
 $j$[{"id":"A","text":"Paragraph A"},{"id":"B","text":"Paragraph B"},{"id":"C","text":"Paragraph C"},{"id":"D","text":"Paragraph D"},{"id":"E","text":"Paragraph E"}]$j$::jsonb,
 '["E"]'::jsonb,
 'Paragraph E is the one that sets a fact about a place against a centrally maintained public utility.',
 ARRAY['E'], 60, 1, 'medium', ARRAY['IELTS Reading', 'Matching Information'],
 'rp-time-01', 'g-time-8', 4);

-- ---------------------------------------------------------------------------
-- Guards
-- ---------------------------------------------------------------------------

-- A blueprint that cannot be filled is a mock that fails when a learner starts
-- it, and finding that out here is far cheaper than finding it out then.
--
-- The check is per section: enough distinct passages must be able to fill the
-- sections that need one, and each section's tasks must each be satisfiable.
DO $$
DECLARE
    bad TEXT;
BEGIN
    SELECT string_agg(format('%s section %s', blueprint_id, position), '; ')
      INTO bad
      FROM (
          SELECT s.blueprint_id, s.position
            FROM reading_mock_blueprint_slots s
            JOIN reading_mock_blueprints b ON b.id = s.blueprint_id
           WHERE s.source = 'passage'
           GROUP BY s.blueprint_id, s.position, b.exam
          HAVING NOT EXISTS (
              SELECT 1
                FROM reading_passages p
               WHERE p.is_published
                 AND NOT EXISTS (
                     SELECT 1
                       FROM reading_mock_blueprint_slots want
                      WHERE want.blueprint_id = s.blueprint_id
                        AND want.position = s.position
                        AND (
                            SELECT count(*)
                              FROM questions q
                              JOIN reading_question_groups g ON g.id = q.group_id
                             WHERE g.passage_id = p.id
                               AND g.type_id = want.type_id
                               AND q.is_published
                               AND b.exam = ANY(q.supported_exams)
                        ) < want.question_count
                 )
          )
      ) unfillable;

    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'blueprint: no passage can fill %', bad;
    END IF;
END $$;

-- One passage now carries both exams, which is the whole point of the refactor.
DO $$
DECLARE
    shared INT;
BEGIN
    SELECT count(*) INTO shared
      FROM reading_passages p
     WHERE EXISTS (SELECT 1 FROM questions q WHERE q.passage_id = p.id AND 'IELTS' = ANY(q.supported_exams))
       AND EXISTS (SELECT 1 FROM questions q WHERE q.passage_id = p.id AND 'PTE'   = ANY(q.supported_exams));

    IF shared < 4 THEN
        RAISE EXCEPTION 'shared passages: % carry both exams, expected all 4', shared;
    END IF;
END $$;
