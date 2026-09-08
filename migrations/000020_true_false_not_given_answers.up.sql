-- 000018 added Not Given as an option; this makes it an answer.
--
-- An option that is never correct teaches the wrong lesson: a learner works out
-- within a set or two that Not Given can be ignored, which is the opposite of
-- the skill the task exists to test. Two statements per passage are rewritten
-- so each set of six runs 2 True / 2 False / 2 Not Given.
--
-- Rewritten in place rather than added. True/False is six of the thirteen
-- questions in paper slot 1, and 000009 asserts 13/13/14; inserting new rows
-- would break the count.
--
-- Each new statement is one the passage neither supports nor contradicts. That
-- is the distinction being tested: False means the passage says otherwise, Not
-- Given means the passage does not say. A statement that is merely a plausible
-- inference from what is written is Not Given, however tempting the inference.

UPDATE questions q
   SET prompt          = v.prompt,
       correct_answers = v.correct_answers::jsonb,
       explanation     = v.explanation
  FROM (VALUES
    ('q-choc-010','The Maya and the Olmec traded cacao beans with each other.','["NOT_GIVEN"]','Paragraph A has both peoples processing cacao, but says nothing about trade between them.'),
    ('q-choc-012','Fair Trade certified chocolate costs more in shops than uncertified chocolate.','["NOT_GIVEN"]','Paragraph J describes a minimum price guaranteed to farmers, not what a shopper pays.'),

    ('q-bees-010','Beekeepers must complete a training course before they may keep hives in London.','["NOT_GIVEN"]','Paragraph I calls training requirements hard to justify in law and easier to defend in practice, but never says a city requires one.'),
    ('q-bees-011','Hives kept in cities produce more honey than hives kept in the countryside.','["NOT_GIVEN"]','Paragraph B gives cities more varied forage and milder winters, but no comparison of yields is made anywhere in the passage.'),

    ('q-paper-011','The mills at Fabriano employed more workers than the mills of Islamic Spain.','["NOT_GIVEN"]','Paragraph C compares what the two produced and how, never the size of their workforces.'),
    ('q-paper-012','Recycled paper is cheaper to produce than paper made from new fibre.','["NOT_GIVEN"]','Paragraph I sets out the limits of recycling but makes no claim about cost.')
  ) AS v(id, prompt, correct_answers, explanation)
 WHERE q.id = v.id;

-- Every True/False set must now offer all three answers, or the set still fails
-- to teach the distinction.
DO $$
DECLARE
    bad TEXT;
BEGIN
    SELECT string_agg(format('%s has %s', passage_id, answers), '; ')
      INTO bad
      FROM (
          SELECT passage_id, string_agg(DISTINCT correct_answers->>0, '/' ORDER BY correct_answers->>0) AS answers
            FROM questions
           WHERE type_id = 'reading-true-false' AND passage_id IS NOT NULL
           GROUP BY passage_id
      ) sets
     WHERE answers <> 'FALSE/NOT_GIVEN/TRUE';

    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'true/false/not given: every set must use all three answers, but %', bad;
    END IF;
END $$;
