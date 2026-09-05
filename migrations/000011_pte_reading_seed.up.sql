-- One PTE reading passage carrying all four passage-backed task types, plus
-- three standalone Re-order Paragraphs items.
--
-- The point of the single passage is authoring economy: one text yields a
-- Reading & Writing gap-fill, a Reading gap-fill, a single-answer MCQ set and a
-- multiple-answer MCQ set. The two gap-fills carry their own gapped rewrites in
-- context_passage and their groups are marked passage_display = 'hidden', so a
-- learner working on them never sees the clean original.
--
-- The re-order items are deliberately about something else. They are standalone
-- by design — see 000010 — and giving them their own subjects means a learner
-- who has read this passage has not been handed any of their answers.

-- ---------------------------------------------------------------------------
-- Passage
-- ---------------------------------------------------------------------------

INSERT INTO reading_passages (
    id, exam_version_id, exam, title, subtitle, paragraphs, sources,
    word_count, difficulty, topic, tags) VALUES

('rp-time-01', 'pte-2026-01', 'PTE',
 'The Coming of Standard Time',
 'How the railways turned a local fact into a public utility',
 $j$[
  {"label":"A","text":"For most of recorded history a town kept its own time, and the time it kept was astronomical. Noon was the moment the sun stood highest over that particular place, and since the sun does not stand highest over two places at once, no two towns agreed. A clock in Bristol ran about ten minutes behind a clock in London, and one in Plymouth about sixteen. Nobody minded, because nobody could travel fast enough for the difference to be noticed. A traveller who lost a quarter of an hour crossing England lost it gradually, over several days, in a coach that was never going to keep to the minute anyway. Coach guards carried watches deliberately set to gain along the westward route and lose on the way back, which was less a correction than a running joke among them. What existed was not a system. It was several thousand unrelated local agreements, each of them perfectly accurate for the place that kept it and useless twenty miles away."},
  {"label":"B","text":"The railway broke this arrangement almost immediately. A train that left London at ten and reached Exeter at two was consulting two different clocks, and a timetable printed in local time was a document that could not be read consistently anywhere along the line. Worse, single-track working depended on knowing where a train was at a given moment, and two stations that disagreed about the moment could route two trains onto the same rails. The Great Western Railway settled the matter for itself in 1840 by adopting London time across its whole network, and the other companies followed within a decade. For some years many station clocks carried two minute hands, one for railway time and one for the town, a compromise that satisfied nobody and made the point better than any argument could. Parliament did not legislate until 1880, by which time the railways had already made Greenwich the effective standard and the law was only recording what had happened."},
  {"label":"C","text":"Britain was small enough to run on a single clock. Continental countries were mostly not, and North America emphatically was not. An American traveller crossing the country in the 1870s encountered dozens of local times, and the major railroads maintained their own competing standards on top of them, so that a large station might display five or six clocks with no obvious way to reconcile them. The Canadian engineer Sandford Fleming, having missed a train in Ireland because a timetable printed the hour ambiguously, spent much of the following decade arguing for a scheme of worldwide zones, each an hour wide and each an exact number of hours from one agreed meridian. The American railroads did not wait for governments. On a Sunday in November 1883 they simply switched, dividing the continent into five zones on their own authority, and most towns adopted railroad time because the alternative was to be out of step with everything that arrived."},
  {"label":"D","text":"Governments caught up the following year. The International Meridian Conference met in Washington in October 1884 with delegates from twenty-five countries and settled on Greenwich as the prime meridian. The choice was practical rather than political: roughly two-thirds of the world shipping tonnage already navigated on charts drawn from Greenwich, and adopting anything else would have meant redrawing them. France abstained, proposed instead a neutral meridian passing through no country at all, and went on reckoning from Paris until 1911, when it adopted Greenwich time under a description that carefully avoided naming it. The conference agreed the meridian and the universal day. It did not agree the zones, which were left to each country to adopt as it saw fit, and that is why the map of world time is not a set of neat vertical stripes but a jagged thing that follows borders, and in places bends several hundred miles to keep one country whole."},
  {"label":"E","text":"What the nineteenth century did was quietly change what time meant. Before the railways it was a fact about a place, read off the sky and belonging to whoever lived there. After 1884 it was a public utility, agreed between strangers, maintained centrally and distributed outward, first by telegraph, then by radio, now by satellite. The gain is so obvious that it has become invisible: every timetable, every market opening and every international call assumes it. The loss is harder to name but it is real. Clock noon and solar noon can now differ by an hour or more near the edge of a wide zone, so that the sun stands highest over western Spain well after the clocks there say midday. Daylight saving, introduced during the First World War to save coal, widened that gap deliberately. The arrangement is a convention repeated so often that it feels like a discovery, which is roughly what happens to any convention that works."}
 ]$j$::jsonb,
 '[]'::jsonb,
 804, 'medium', 'History of science and technology',
 ARRAY['PTE Reading', 'History']);

-- ---------------------------------------------------------------------------
-- Groups
-- ---------------------------------------------------------------------------
--
-- paper_slot stays 0 throughout: these are practice tasks, not sections of a
-- generated IELTS paper.

INSERT INTO reading_question_groups (
    id, passage_id, position, type_id, type_name, instructions, resources,
    passage_display, shuffle_questions, time_limit_seconds) VALUES

('g-time-1', 'rp-time-01', 1, 'fill-in-blanks-rw', 'Reading & Writing: Fill in the Blanks',
 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',
 '[]'::jsonb, 'hidden', FALSE, 180),

-- The word bank lives on the group rather than on the question because it is
-- shared across every gap in the task: one pool, drawn from once per blank, with
-- more words in it than there are gaps.
('g-time-2', 'rp-time-01', 2, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',
 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. There are more words than gaps.',
 $j$[
  {"label":"w1","text":"necessary"},
  {"label":"w2","text":"consistently"},
  {"label":"w3","text":"moment"},
  {"label":"w4","text":"recorded"},
  {"label":"w5","text":"optional"},
  {"label":"w6","text":"gradually"}
 ]$j$::jsonb, 'hidden', FALSE, 120),

('g-time-3', 'rp-time-01', 3, 'reading-mcq-single', 'Multiple Choice, Single Answer',
 'Read the passage and answer the question by selecting the correct response. Only one response is correct.',
 '[]'::jsonb, 'full', TRUE, 270),

('g-time-4', 'rp-time-01', 4, 'reading-mcq-multiple', 'Multiple Choice, Multiple Answers',
 'Read the passage and answer the question by selecting all the correct responses. More than one response is correct. Incorrect responses cancel out correct ones.',
 '[]'::jsonb, 'full', TRUE, 240);

-- ---------------------------------------------------------------------------
-- Gap-fill questions
-- ---------------------------------------------------------------------------
--
-- One question per gapped text, not one per gap. The whole text is the item, and
-- scoring.gradeBlanks already awards a share of the marks per blank filled, so
-- splitting it into five rows would duplicate the passage five times to no end.
--
-- context_passage carries the gapped text with [[bN]] markers; blanks carries
-- the answer key. PublicQuestion() strips Blank.CorrectAnswer before any of this
-- reaches a learner.

INSERT INTO questions (
    id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
    context_passage, blanks, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position) VALUES

('q-time-001', 'pte-2026-01', 'PTE', 'reading', 'fill-in-blanks-rw',
 'Reading & Writing: Fill in the Blanks',
 'Standard Time', 'Select the appropriate answer choice for each blank.',
 'Before the railways, time was a local matter. Every town set its clocks by the sun, so that noon in one place arrived several minutes [[b1]] noon in another, and nobody was inconvenienced because nobody could travel quickly enough to notice. The railway made that arrangement [[b2]]. A timetable printed in local time could not be read consistently along a line that crossed several of them, and signalling on a single track depended on two distant stations agreeing about the present moment. British railway companies solved the problem privately, adopting London time across their networks decades before Parliament thought to [[b3]] the practice. Elsewhere the difficulty was larger. A traveller crossing North America in the 1870s met dozens of local times layered under the competing standards of the railroads themselves. The Canadian engineer Sandford Fleming argued for a global scheme of hour-wide zones measured from a single meridian, but it was the American railroads, acting without any legal authority whatever, who imposed five zones on the continent in 1883. The international conference that followed in Washington chose Greenwich as the prime meridian for a reason that was commercial rather than [[b4]]: most of the world shipping already navigated on charts drawn from it, and any other choice would have required them all to be redrawn. France abstained and kept Paris time for another twenty-seven years. The conference fixed the meridian but left the zones to individual governments, which is why the modern map of world time follows political borders rather than lines of longitude. The deeper change was one of meaning. Time stopped being an astronomical fact belonging to a place and became a public utility, agreed between strangers and [[b5]] outward from a central source.',
 $j$[
  {"id":"b1","options":["behind","beneath","beyond","besides"],"correctAnswer":"behind"},
  {"id":"b2","options":["unworkable","unremarkable","unavoidable","uneventful"],"correctAnswer":"unworkable"},
  {"id":"b3","options":["regulate","reverse","relieve","resemble"],"correctAnswer":"regulate"},
  {"id":"b4","options":["political","practical","personal","provincial"],"correctAnswer":"political"},
  {"id":"b5","options":["distributed","distracted","distinguished","disturbed"],"correctAnswer":"distributed"}
 ]$j$::jsonb,
 'Each gap is settled by the clause around it: "several minutes ___ noon in another" needs a comparison of lag; "commercial rather than ___" needs the opposite of commercial, which the passage gives as political.',
 180, 5, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-time-01', 'g-time-1', 1),

('q-time-002', 'pte-2026-01', 'PTE', 'reading', 'fill-in-blanks-r',
 'Reading: Fill in the Blanks',
 'Railway Time', 'Drag a word from the box into each gap.',
 'The railway made a single standard [[b1]]. A timetable printed in local time could not be read [[b2]] anywhere along a route that crossed several of them, and signalling on a single track depended on two distant stations agreeing about the present [[b3]]. The British companies adopted London time across their whole networks during the 1840s, some forty years before Parliament finally [[b4]] the practice in law.',
 $j$[
  {"id":"b1","correctAnswer":"necessary"},
  {"id":"b2","correctAnswer":"consistently"},
  {"id":"b3","correctAnswer":"moment"},
  {"id":"b4","correctAnswer":"recorded"}
 ]$j$::jsonb,
 'Two words in the box are never used. "Optional" reverses the sense of the first gap, and "gradually" describes how coach travellers lost time, not what Parliament did.',
 120, 4, 'medium', ARRAY['PTE Reading', 'Fill in the Blanks'],
 'rp-time-01', 'g-time-2', 1);

-- ---------------------------------------------------------------------------
-- Multiple choice
-- ---------------------------------------------------------------------------
--
-- correct_answers holds option ids, which is what the client submits and what
-- scoring.gradeChoice compares. Points equal the number of correct options, so
-- one correctly chosen option is worth one mark in both sets.

INSERT INTO questions (
    id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, time_limit_seconds, points,
    difficulty, tags, passage_id, group_id, group_position) VALUES

('q-time-003', 'pte-2026-01', 'PTE', 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 1',
 'According to the passage, why did differences between local times cause so little difficulty before the railways?',
 $j$[
  {"id":"A","text":"Clocks of the period were too inaccurate to show them."},
  {"id":"B","text":"Travel was too slow for the difference to be noticed."},
  {"id":"C","text":"Very few people outside the towns owned a clock at all."},
  {"id":"D","text":"Coach guards corrected every watch along the route."}
 ]$j$::jsonb, '["B"]'::jsonb,
 'Paragraph A: nobody minded because nobody could travel fast enough for the difference to be noticed, and a quarter of an hour was lost gradually over several days.',
 90, 1, 'easy', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-time-01', 'g-time-3', 1),

('q-time-004', 'pte-2026-01', 'PTE', 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 2',
 'What does the writer suggest about the station clocks that carried two minute hands?',
 $j$[
  {"id":"A","text":"They were an elegant solution that was widely copied."},
  {"id":"B","text":"They were introduced to comply with the 1880 legislation."},
  {"id":"C","text":"They illustrated the problem more effectively than argument did."},
  {"id":"D","text":"They showed railway time alongside true solar time."}
 ]$j$::jsonb, '["C"]'::jsonb,
 'Paragraph B: the compromise satisfied nobody and made the point better than any argument could. The second hand showed the town time, not solar time, and the clocks predate the 1880 Act.',
 90, 1, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-time-01', 'g-time-3', 2),

('q-time-005', 'pte-2026-01', 'PTE', 'reading', 'reading-mcq-single',
 'Multiple Choice, Single Answer', 'Multiple Choice, Single Answer 3',
 'Why was Greenwich adopted as the prime meridian in 1884?',
 $j$[
  {"id":"A","text":"It was the only meridian formally proposed to the conference."},
  {"id":"B","text":"Most of the world shipping already navigated on charts drawn from it."},
  {"id":"C","text":"It was politically neutral in a way that Paris was not."},
  {"id":"D","text":"France withdrew its own proposal before the vote was taken."}
 ]$j$::jsonb, '["B"]'::jsonb,
 'Paragraph D: roughly two-thirds of world shipping tonnage already used Greenwich charts and anything else would have meant redrawing them. France proposed a neutral meridian and did not withdraw it; it abstained.',
 90, 1, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-time-01', 'g-time-3', 3),

('q-time-006', 'pte-2026-01', 'PTE', 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Multiple Choice, Multiple Answers 1',
 'Which TWO problems does the passage attribute to running railways on local time?',
 $j$[
  {"id":"A","text":"A timetable could not be read consistently along the whole line."},
  {"id":"B","text":"Stations disagreeing about the moment could route two trains onto the same rails."},
  {"id":"C","text":"Parliament was forced to legislate within a year."},
  {"id":"D","text":"Coach guards stopped adjusting their watches along the route."},
  {"id":"E","text":"Solar noon and clock noon drifted apart across a wide zone."}
 ]$j$::jsonb, '["A","B"]'::jsonb,
 'Both are in paragraph B. Parliament waited until 1880, four decades on; the drift between solar and clock noon is a consequence of wide zones in paragraph E, not of local time.',
 120, 2, 'medium', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-time-01', 'g-time-4', 1),

('q-time-007', 'pte-2026-01', 'PTE', 'reading', 'reading-mcq-multiple',
 'Multiple Choice, Multiple Answers', 'Multiple Choice, Multiple Answers 2',
 'Which THREE statements about the 1884 International Meridian Conference are supported by the passage?',
 $j$[
  {"id":"A","text":"It settled on Greenwich as the prime meridian."},
  {"id":"B","text":"It agreed on the universal day."},
  {"id":"C","text":"It left the adoption of time zones to individual countries."},
  {"id":"D","text":"France voted in favour of the Greenwich proposal."},
  {"id":"E","text":"It fixed the boundaries of the world time zones."}
 ]$j$::jsonb, '["A","B","C"]'::jsonb,
 'Paragraph D states all three. France abstained and proposed a neutral meridian instead, and the conference explicitly did not agree the zones, which is why the modern map follows borders.',
 120, 3, 'hard', ARRAY['PTE Reading', 'Multiple Choice'],
 'rp-time-01', 'g-time-4', 2);

-- ---------------------------------------------------------------------------
-- Re-order Paragraphs
-- ---------------------------------------------------------------------------
--
-- paragraphs is stored in the correct order — the array order is the answer —
-- but the labels are deliberately not alphabetical. The labels are the option
-- ids the client submits, so running them A, B, C, D down the correct sequence
-- would let anyone read the answer straight out of the response.

INSERT INTO reading_reorder_items (
    id, exam_version_id, exam, title, paragraphs, topic, word_count,
    difficulty, tags) VALUES

('ri-lock-01', 'pte-2026-01', 'PTE', 'How a Lock Raises a Boat',
 $j$[
  {"label":"p3","text":"A boat travelling upstream enters the lower chamber, and the gates behind it are closed."},
  {"label":"p1","text":"Water is then admitted from the upper level, and the boat rises with it until the two levels match."},
  {"label":"p4","text":"The upper gates can now be opened, because there is no longer any difference in pressure across them."},
  {"label":"p2","text":"The boat leaves at the higher level, and the chamber is emptied again for whatever is coming down."}
 ]$j$::jsonb,
 'Engineering', 66, 'easy', ARRAY['PTE Reading', 'Re-order Paragraphs']),

('ri-penicillin-01', 'pte-2026-01', 'PTE', 'The Discovery of Penicillin',
 $j$[
  {"label":"p4","text":"Alexander Fleming returned from holiday in 1928 to find a culture plate he had left out contaminated by a mould."},
  {"label":"p2","text":"Around the mould the staphylococcus colonies had dissolved, which suggested it was producing something that killed them."},
  {"label":"p5","text":"He published the observation the following year, but could not extract the substance in any useful quantity, and the work stalled."},
  {"label":"p1","text":"A decade later a team at Oxford took up the problem and purified enough of the compound to test it on infected mice."},
  {"label":"p3","text":"Mass production followed in the United States during the war, and a drug that had sat unused for ten years reached patients in millions."}
 ]$j$::jsonb,
 'History of medicine', 108, 'medium', ARRAY['PTE Reading', 'Re-order Paragraphs']),

('ri-expansion-01', 'pte-2026-01', 'PTE', 'Why Bridges Are Given Expansion Joints',
 $j$[
  {"label":"p2","text":"Steel and concrete both grow slightly longer as they warm, and a long span can gain several centimetres between a winter night and a summer afternoon."},
  {"label":"p4","text":"If the ends were held rigidly in place that growth would have nowhere to go, and would instead build up as stress inside the structure."},
  {"label":"p1","text":"Engineers therefore leave a deliberate gap at one or both ends, bridged by interlocking metal teeth that slide across each other."},
  {"label":"p3","text":"The rhythmic knock a car makes crossing a long bridge is the sound of those joints doing exactly what they were designed to do."}
 ]$j$::jsonb,
 'Engineering', 95, 'medium', ARRAY['PTE Reading', 'Re-order Paragraphs']);

-- One question per item. options are the boxes and correct_answers is the
-- sequence, which is what scoring.gradeReorder reads: it scores adjacent pairs,
-- so a learner who gets three of four boxes consecutive still earns marks.
--
-- Points are len(boxes) - 1, the number of adjacent pairs, so one correctly
-- placed neighbour is worth one mark.
INSERT INTO questions (
    id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
    options, correct_answers, explanation, time_limit_seconds, points,
    difficulty, tags, reorder_item_id)
SELECT
    'q-' || i.id,
    i.exam_version_id, i.exam, 'reading', 'reorder-paragraphs', 'Re-order Paragraphs',
    i.title,
    'The text boxes below have been placed in a random order. Restore the original order.',
    (SELECT jsonb_agg(jsonb_build_object('id', e->>'label', 'text', e->>'text'))
       FROM jsonb_array_elements(i.paragraphs) e),
    (SELECT jsonb_agg(e->>'label')
       FROM jsonb_array_elements(i.paragraphs) e),
    'Find the box that opens the sequence without referring back to anything. Boxes beginning with "This", "Then", "Therefore" or a pronoun always follow something else.',
    300,
    jsonb_array_length(i.paragraphs) - 1,
    i.difficulty,
    i.tags,
    i.id
FROM reading_reorder_items i
WHERE i.exam_version_id = 'pte-2026-01';

-- ---------------------------------------------------------------------------
-- Guard
-- ---------------------------------------------------------------------------

DO $$
DECLARE
    n INT;
    bad TEXT;
BEGIN
    SELECT count(DISTINCT g.type_id) INTO n
      FROM reading_question_groups g
      JOIN reading_passages p ON p.id = g.passage_id AND p.exam = 'PTE' AND p.is_published
     WHERE EXISTS (SELECT 1 FROM questions q WHERE q.group_id = g.id AND q.is_published);
    IF n <> 4 THEN
        RAISE EXCEPTION 'pte reading seed: % passage task types with questions, need 4', n;
    END IF;

    -- A gap-fill whose group renders the passage would show the learner the
    -- ungapped original, which is the whole reason passage_display exists.
    SELECT string_agg(id, ', ') INTO bad
      FROM reading_question_groups
     WHERE type_id IN ('fill-in-blanks-rw', 'fill-in-blanks-r')
       AND passage_display <> 'hidden';
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'pte reading seed: gap-fill groups must hide the passage: %', bad;
    END IF;

    -- Every marker in a gapped text needs an answer, and every answer a marker.
    SELECT string_agg(q.id, ', ') INTO bad
      FROM questions q
     WHERE q.type_id IN ('fill-in-blanks-rw', 'fill-in-blanks-r')
       AND q.context_passage IS NOT NULL
       AND (SELECT count(*) FROM regexp_matches(q.context_passage, '\[\[b[0-9]+\]\]', 'g'))
           <> jsonb_array_length(q.blanks);
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'pte reading seed: gap markers and blanks disagree in: %', bad;
    END IF;

    -- The answer sequence has to name the boxes the learner is given, or the
    -- task is unanswerable.
    SELECT string_agg(q.id, ', ') INTO bad
      FROM questions q
     WHERE q.reorder_item_id IS NOT NULL
       AND (SELECT count(*) FROM jsonb_array_elements_text(q.correct_answers) a
             WHERE NOT EXISTS (SELECT 1 FROM jsonb_array_elements(q.options) o
                                WHERE o->>'id' = a)) > 0;
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'pte reading seed: re-order answers name boxes that do not exist in: %', bad;
    END IF;

    SELECT count(*) INTO n FROM reading_reorder_items WHERE exam = 'PTE' AND is_published;
    IF n < 3 THEN
        RAISE EXCEPTION 'pte reading seed: % re-order items, need at least 3', n;
    END IF;
END $$;
