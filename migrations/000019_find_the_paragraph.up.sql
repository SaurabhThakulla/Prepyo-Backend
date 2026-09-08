-- Find the Writer becomes Find the Paragraph.
--
-- The task is replaced in place rather than deleted. Find the Writer carried
-- six of the thirteen questions in paper slot 2, and 000009 asserts each slot
-- holds exactly 13/13/14; deleting it would leave every generated paper six
-- questions short. Rewriting the same six rows keeps the slot counts, the
-- group positions and the question ids exactly as they were.
--
-- The five Writer A-E commentaries existed only as the answer key for Find the
-- Writer, so they go with it. The new task is answered from the passage body,
-- which every other task on these passages already uses.

-- Guard: this migration rewrites eighteen questions by id. Any Find the Writer
-- question authored after the seed would be relabelled Find the Paragraph and
-- given paragraph options while keeping its "Which writer..." prompt and its
-- Writer A-E answer -- and because A-E are also valid paragraph labels, that
-- corruption would look plausible instead of failing. Stopping here is cheaper.
DO $$
DECLARE
    unexpected TEXT;
BEGIN
    SELECT string_agg(id, ', ' ORDER BY id) INTO unexpected
      FROM questions
     WHERE type_id = 'reading-find-the-writer'
       AND id NOT IN (
           'q-choc-014','q-choc-015','q-choc-016','q-choc-017','q-choc-018','q-choc-019',
           'q-bees-014','q-bees-015','q-bees-016','q-bees-017','q-bees-018','q-bees-019',
           'q-paper-014','q-paper-015','q-paper-016','q-paper-017','q-paper-018','q-paper-019');

    IF unexpected IS NOT NULL THEN
        RAISE EXCEPTION 'find the paragraph: these Find the Writer questions are not covered by this migration and would be corrupted: %', unexpected;
    END IF;
END $$;

UPDATE reading_question_groups
   SET type_id      = 'reading-find-the-paragraph',
       type_name    = 'Find the Paragraph',
       instructions = 'The passage has ten paragraphs, A to J. Which paragraph contains each of the following?'
 WHERE type_id = 'reading-find-the-writer';

-- Answers avoid paragraphs F, G, H and J: those are the answers to the
-- Matching Information set on the same passage, and a paper that asked for the
-- same paragraph twice would be giving one of them away.

UPDATE questions q
   SET type_id         = 'reading-find-the-paragraph',
       type_name       = 'Find the Paragraph',
       title           = 'Find the Paragraph ' || q.group_position::text,
       prompt          = v.prompt,
       correct_answers = v.correct_answers::jsonb,
       explanation     = v.explanation,
       tags            = ARRAY['IELTS Reading', 'Find the Paragraph']
  FROM (VALUES
    ('q-choc-014','A reference to the earliest people known to have processed cacao pods.','["A"]','Paragraph A names the Olmec and dates the residue analysis.'),
    ('q-choc-015','An account of beans being hollowed out and packed with earth.','["B"]','Paragraph B describes forged currency beans.'),
    ('q-choc-016','A description of how one European first came across cacao at sea.','["C"]','Paragraph C: Columbus men boarded a Maya trading canoe in 1502.'),
    ('q-choc-017','The cost of a drink set against what a labourer earned.','["D"]','Paragraph D: a bowl cost more than a labourer earned in a day.'),
    ('q-choc-018','A sequence of nineteenth-century inventions that reshaped how chocolate was made.','["E"]','Paragraph E runs from the 1828 press to conching in 1879.'),
    ('q-choc-019','A wartime product deliberately formulated to taste barely acceptable.','["I"]','Paragraph I: rations made to taste only a little better than a boiled potato.'),

    ('q-bees-014','A ban on keeping an animal, and the year it was lifted.','["A"]','Paragraph A: New York listed the honeybee as prohibited until 2010.'),
    ('q-bees-015','An explanation of why a city offers a longer flowering season than farmland.','["B"]','Paragraph B contrasts a single crop with forage flowering February to November.'),
    ('q-bees-016','The naming of a phenomenon in which workers failed to return to the hive.','["C"]','Paragraph C names colony collapse disorder, reported from 2006.'),
    ('q-bees-017','Figures for what a hive costs to establish and what it is likely to yield.','["D"]','Paragraph D gives four to six hundred pounds and twenty kilograms.'),
    ('q-bees-018','A warning that hive numbers have passed what local forage can support.','["E"]','Paragraph E cites ten colonies per square kilometre in London boroughs.'),
    ('q-bees-019','An argument that registration matters more than installing further hives.','["I"]','Paragraph I: a compulsory register is put ahead of any campaign for more hives.'),

    ('q-paper-014','A claim that the traditional date given for an invention is too late.','["A"]','Paragraph A: fragments predate 105 CE by two centuries.'),
    ('q-paper-015','An account of a technique spreading as a consequence of a battle.','["B"]','Paragraph B: papermakers taken at Talas in 751.'),
    ('q-paper-016','Three improvements added by Italian mills to a process they had copied.','["C"]','Paragraph C: stamping hammers, gelatine sizing and the watermark.'),
    ('q-paper-017','Laws passed to protect the supply of a raw material.','["D"]','Paragraph D: export bans and rules on burial in linen.'),
    ('q-paper-018','Two inventions that between them ended a long-running shortage.','["E"]','Paragraph E: the Fourdrinier machine of 1806 and Keller grinding process of 1844.'),
    ('q-paper-019','A limit on the number of times a material can be reused.','["I"]','Paragraph I: fibre survives perhaps five to seven pulpings.')
  ) AS v(id, prompt, correct_answers, explanation)
 WHERE q.id = v.id;

UPDATE questions SET options = '[
    {"id":"A","text":"Paragraph A"},
    {"id":"B","text":"Paragraph B"},
    {"id":"C","text":"Paragraph C"},
    {"id":"D","text":"Paragraph D"},
    {"id":"E","text":"Paragraph E"},
    {"id":"F","text":"Paragraph F"},
    {"id":"G","text":"Paragraph G"},
    {"id":"H","text":"Paragraph H"},
    {"id":"I","text":"Paragraph I"},
    {"id":"J","text":"Paragraph J"}]'::jsonb
WHERE type_id = 'reading-find-the-paragraph' AND passage_id IS NOT NULL;

UPDATE reading_passages SET sources = '[]'::jsonb WHERE sources <> '[]'::jsonb;

-- Same guard 000009 applies, re-run because this migration rewrote a slot.
DO $$
DECLARE
    bad TEXT;
BEGIN
    SELECT string_agg(format('%s slot %s has %s questions', passage_id, paper_slot, n), '; ')
      INTO bad
      FROM (
          SELECT g.passage_id, g.paper_slot, count(q.id) AS n
            FROM reading_question_groups g
            JOIN reading_passages p ON p.id = g.passage_id AND p.exam = 'IELTS' AND p.is_published
            LEFT JOIN questions q ON q.group_id = g.id AND q.is_published
           WHERE g.paper_slot > 0
           GROUP BY g.passage_id, g.paper_slot
      ) counts
     WHERE (paper_slot = 1 AND n <> 13)
        OR (paper_slot = 2 AND n <> 13)
        OR (paper_slot = 3 AND n <> 14);

    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'reading slots: expected 13/13/14 questions per slot, got %', bad;
    END IF;
END $$;
