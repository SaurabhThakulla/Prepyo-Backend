-- Sentence completion becomes a choice task, the way PTE Reading & Writing
-- Fill in the Blanks works: the gap is filled by picking one of four words
-- rather than by typing.
--
-- Nothing about the grader changes. gradeShortAnswer already accepts a single
-- selected option in place of typed text, "the same answer arriving by a
-- different route", so correct_answers stays exactly as it was, alternative
-- spellings included. That is why each option id is the word itself rather
-- than a letter: the id submitted is a string the grader already accepts.
--
-- The number of question rows is untouched. Sentence completion is 7 questions
-- in slot 1 and 6 in slot 3 on every passage, and 000009 asserts 13/13/14.
--
-- Distractors are drawn from the same passage wherever possible, so choosing
-- still requires reading it. One pair is deliberately reciprocal: lignin and
-- rosin appear in each other option lists, because a learner who has skimmed
-- paragraph H will have seen both words without registering which does what.

UPDATE questions q SET options = v.options::jsonb
  FROM (VALUES
    ('q-choc-001','[{"id":"gods","text":"gods"},{"id":"kings","text":"kings"},{"id":"priests","text":"priests"},{"id":"ancestors","text":"ancestors"}]'),
    ('q-choc-002','[{"id":"bitter","text":"bitter"},{"id":"sweet","text":"sweet"},{"id":"creamy","text":"creamy"},{"id":"spiced","text":"spiced"}]'),
    ('q-choc-003','[{"id":"currency","text":"currency"},{"id":"tribute","text":"tribute"},{"id":"medicine","text":"medicine"},{"id":"ornament","text":"ornament"}]'),
    ('q-choc-004','[{"id":"cinnamon","text":"cinnamon"},{"id":"vanilla","text":"vanilla"},{"id":"chilli","text":"chilli"},{"id":"honey","text":"honey"}]'),
    ('q-choc-005','[{"id":"London","text":"London"},{"id":"Madrid","text":"Madrid"},{"id":"Paris","text":"Paris"},{"id":"Amsterdam","text":"Amsterdam"}]'),
    ('q-choc-006','[{"id":"butter","text":"butter"},{"id":"powder","text":"powder"},{"id":"solids","text":"solids"},{"id":"paste","text":"paste"}]'),
    ('q-choc-007','[{"id":"1847","text":"1847"},{"id":"1828","text":"1828"},{"id":"1875","text":"1875"},{"id":"1502","text":"1502"}]'),
    ('q-choc-027','[{"id":"Nestle","text":"Nestle"},{"id":"Peter","text":"Peter"},{"id":"Lindt","text":"Lindt"},{"id":"van Houten","text":"van Houten"}]'),
    ('q-choc-028','[{"id":"acidity","text":"acidity"},{"id":"bitterness","text":"bitterness"},{"id":"moisture","text":"moisture"},{"id":"sweetness","text":"sweetness"}]'),
    ('q-choc-029','[{"id":"rations","text":"rations"},{"id":"uniforms","text":"uniforms"},{"id":"parcels","text":"parcels"},{"id":"kits","text":"kits"}]'),
    ('q-choc-030','[{"id":"Africa","text":"Africa"},{"id":"Indies","text":"Indies"},{"id":"Asia","text":"Asia"},{"id":"America","text":"America"}]'),
    ('q-choc-031','[{"id":"machete","text":"machete"},{"id":"ladder","text":"ladder"},{"id":"hook","text":"hook"},{"id":"blade","text":"blade"}]'),
    ('q-choc-032','[{"id":"flavour","text":"flavour"},{"id":"colour","text":"colour"},{"id":"aroma","text":"aroma"},{"id":"texture","text":"texture"}]'),

    ('q-bees-001','[{"id":"2010","text":"2010"},{"id":"1985","text":"1985"},{"id":"2006","text":"2006"},{"id":"2012","text":"2012"}]'),
    ('q-bees-002','[{"id":"Garnier","text":"Garnier"},{"id":"Bastille","text":"Bastille"},{"id":"Nationale","text":"Nationale"},{"id":"Comique","text":"Comique"}]'),
    ('q-bees-003','[{"id":"embankments","text":"embankments"},{"id":"sidings","text":"sidings"},{"id":"tunnels","text":"tunnels"},{"id":"bridges","text":"bridges"}]'),
    ('q-bees-004','[{"id":"heat","text":"heat"},{"id":"warmth","text":"warmth"},{"id":"thermal","text":"thermal"},{"id":"climate","text":"climate"}]'),
    ('q-bees-005','[{"id":"collapse","text":"collapse"},{"id":"decline","text":"decline"},{"id":"failure","text":"failure"},{"id":"desertion","text":"desertion"}]'),
    ('q-bees-006','[{"id":"kilograms","text":"kilograms"},{"id":"litres","text":"litres"},{"id":"pounds","text":"pounds"},{"id":"tonnes","text":"tonnes"}]'),
    ('q-bees-007','[{"id":"kilometre","text":"kilometre"},{"id":"mile","text":"mile"},{"id":"hectare","text":"hectare"},{"id":"borough","text":"borough"}]'),
    ('q-bees-027','[{"id":"broodless","text":"broodless"},{"id":"dormant","text":"dormant"},{"id":"queenless","text":"queenless"},{"id":"clustered","text":"clustered"}]'),
    ('q-bees-028','[{"id":"15","text":"15"},{"id":"10","text":"10"},{"id":"20","text":"20"},{"id":"25","text":"25"}]'),
    ('q-bees-029','[{"id":"invertase","text":"invertase"},{"id":"glucose","text":"glucose"},{"id":"amylase","text":"amylase"},{"id":"sucrose","text":"sucrose"}]'),
    ('q-bees-030','[{"id":"fanning","text":"fanning"},{"id":"capping","text":"capping"},{"id":"heating","text":"heating"},{"id":"drying","text":"drying"}]'),
    ('q-bees-031','[{"id":"bitter","text":"bitter"},{"id":"pale","text":"pale"},{"id":"floral","text":"floral"},{"id":"sweet","text":"sweet"}]'),
    ('q-bees-032','[{"id":"book","text":"book"},{"id":"hive","text":"hive"},{"id":"manual","text":"manual"},{"id":"season","text":"season"}]'),

    ('q-paper-001','[{"id":"105","text":"105"},{"id":"751","text":"751"},{"id":"1502","text":"1502"},{"id":"1806","text":"1806"}]'),
    ('q-paper-002','[{"id":"cheap","text":"cheap"},{"id":"quicker","text":"quicker"},{"id":"stronger","text":"stronger"},{"id":"secret","text":"secret"}]'),
    ('q-paper-003','[{"id":"Talas","text":"Talas"},{"id":"Samarkand","text":"Samarkand"},{"id":"Baghdad","text":"Baghdad"},{"id":"Fabriano","text":"Fabriano"}]'),
    ('q-paper-004','[{"id":"Baghdad","text":"Baghdad"},{"id":"Samarkand","text":"Samarkand"},{"id":"Cairo","text":"Cairo"},{"id":"Damascus","text":"Damascus"}]'),
    ('q-paper-005','[{"id":"watermark","text":"watermark"},{"id":"hammer","text":"hammer"},{"id":"sizing","text":"sizing"},{"id":"mould","text":"mould"}]'),
    ('q-paper-006','[{"id":"rags","text":"rags"},{"id":"bark","text":"bark"},{"id":"hemp","text":"hemp"},{"id":"straw","text":"straw"}]'),
    ('q-paper-007','[{"id":"wasps","text":"wasps"},{"id":"bees","text":"bees"},{"id":"ants","text":"ants"},{"id":"birds","text":"birds"}]'),
    ('q-paper-027','[{"id":"strong","text":"strong"},{"id":"smooth","text":"smooth"},{"id":"thin","text":"thin"},{"id":"bright","text":"bright"}]'),
    ('q-paper-028','[{"id":"170","text":"170"},{"id":"100","text":"100"},{"id":"90","text":"90"},{"id":"60","text":"60"}]'),
    ('q-paper-029','[{"id":"lignin","text":"lignin"},{"id":"cellulose","text":"cellulose"},{"id":"alum","text":"alum"},{"id":"rosin","text":"rosin"}]'),
    ('q-paper-030','[{"id":"rosin","text":"rosin"},{"id":"gelatine","text":"gelatine"},{"id":"lignin","text":"lignin"},{"id":"sulphide","text":"sulphide"}]'),
    ('q-paper-031','[{"id":"board","text":"board"},{"id":"pulp","text":"pulp"},{"id":"card","text":"card"},{"id":"waste","text":"waste"}]'),
    ('q-paper-032','[{"id":"7.0","text":"7.0"},{"id":"6.5","text":"6.5"},{"id":"7.5","text":"7.5"},{"id":"8.0","text":"8.0"}]')
  ) AS v(id, options)
 WHERE q.id = v.id;

-- The instructions told the learner to write a word. They now choose one.
UPDATE reading_question_groups
   SET instructions = 'Complete the sentences below. Choose the correct word for each gap.'
 WHERE type_id = 'reading-sentence-completion' AND instructions LIKE 'Complete the sentences%';

UPDATE reading_question_groups
   SET instructions = 'Complete the summary below. Choose the correct word for each gap.'
 WHERE type_id = 'reading-sentence-completion' AND instructions LIKE 'Complete the summary%';

DO $$
DECLARE
    bad TEXT;
BEGIN
    -- Every gap must offer four words, or the set is inconsistent to sit.
    SELECT string_agg(id, ', ' ORDER BY id) INTO bad
      FROM questions
     WHERE type_id = 'reading-sentence-completion' AND passage_id IS NOT NULL
       AND COALESCE(jsonb_array_length(options), 0) <> 4;
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'sentence completion: these questions do not offer exactly four options: %', bad;
    END IF;

    -- An option list that does not contain the answer is an unanswerable gap.
    SELECT string_agg(q.id, ', ' ORDER BY q.id) INTO bad
      FROM questions q
     WHERE q.type_id = 'reading-sentence-completion' AND q.passage_id IS NOT NULL
       AND NOT EXISTS (
           SELECT 1 FROM jsonb_array_elements(q.options) o
            WHERE lower(o->>'id') IN (SELECT lower(a) FROM jsonb_array_elements_text(q.correct_answers) a));
    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'sentence completion: no offered option matches the answer for: %', bad;
    END IF;
END $$;
