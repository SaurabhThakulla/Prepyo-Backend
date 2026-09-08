UPDATE questions q
   SET prompt          = v.prompt,
       correct_answers = v.correct_answers::jsonb,
       explanation     = v.explanation
  FROM (VALUES
    ('q-choc-010','Early European royalty kept the recipe for chocolate a secret.','["TRUE"]','Paragraph C: the recipe circulated only among royalty and religious houses.'),
    ('q-choc-012','The industrial revolution slowed down the production of chocolate.','["FALSE"]','Paragraph E: far from slowing production, it multiplied it.'),

    ('q-bees-010','Colonies in arable landscapes have a longer flowering season than city colonies.','["FALSE"]','Paragraph B: arable colonies face one crop for three weeks.'),
    ('q-bees-011','Most urban beekeepers recover the cost of their equipment.','["FALSE"]','Paragraph D: very few recover their costs.'),

    ('q-paper-011','Some countries banned the export of rags.','["TRUE"]','Paragraph D: some countries forbade their export.'),
    ('q-paper-012','Keller pulping process was patented before the Fourdrinier machine was financed.','["FALSE"]','Paragraph E: the machine came in 1806, the patent in 1844.')
  ) AS v(id, prompt, correct_answers, explanation)
 WHERE q.id = v.id;
