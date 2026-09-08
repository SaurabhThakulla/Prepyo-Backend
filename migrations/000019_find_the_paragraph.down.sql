-- Restores Find the Writer: the groups, the six questions per passage and the
-- Writer A-E commentaries they are answered from.

UPDATE reading_question_groups
   SET type_id      = 'reading-find-the-writer',
       type_name    = 'Find the Writer',
       instructions = 'Look at the five commentaries, Writer A to Writer E. Which writer makes each of the following statements?'
 WHERE type_id = 'reading-find-the-paragraph';

UPDATE questions q
   SET type_id         = 'reading-find-the-writer',
       type_name       = 'Find the Writer',
       title           = 'Find the Writer ' || q.group_position::text,
       prompt          = v.prompt,
       correct_answers = v.correct_answers::jsonb,
       explanation     = v.explanation,
       tags            = ARRAY['IELTS Reading', 'Find the Writer']
  FROM (VALUES
    ('q-choc-014','Which writer claims that chocolate has significant cardiovascular benefits?','["C"]','Writer C cites flavanols and cardiovascular benefit.'),
    ('q-choc-015','Which writer focuses on the ethical concerns of modern cocoa farming?','["A"]','Writer A calls ethics the central unresolved problem.'),
    ('q-choc-016','Which writer documented the ceremonial use of cacao by the Aztecs?','["D"]','Writer D works from the Florentine Codex and Nahuatl sources.'),
    ('q-choc-017','Which writer attributes the popularisation of chocolate to European industrialisation?','["B"]','Writer B: it became popular because industrialisation made it cheap.'),
    ('q-choc-018','Which writer argues that sugar ruins the natural antioxidant properties of cacao?','["C"]','Writer C says the sugar undoes the antioxidant properties.'),
    ('q-choc-019','Which writer provided the first recipe for a solid chocolate bar?','["E"]','Writer E cites house records for 1847.'),

    ('q-bees-014','Which writer argues that additional hives can harm wild pollinators?','["A"]','Writer A calls each hive thirty thousand competitors.'),
    ('q-bees-015','Which writer attributes the growth of urban beekeeping to media coverage?','["B"]','Writer B ties the boom to the reporting from 2006.'),
    ('q-bees-016','Which writer claims that urban honey is chemically cleaner than rural honey?','["C"]','Writer C cites lower pesticide residues in city honey.'),
    ('q-bees-017','Which writer documented the keeping of bees inside medieval city walls?','["D"]','Writer D works from monastic accounts and fabric rolls.'),
    ('q-bees-018','Which writer published the first hive design intended for an apartment roof?','["E"]','Writer E describes the bulletin design.'),
    ('q-bees-019','Which writer argues that heating honey destroys the properties it is valued for?','["C"]','Writer C warns about heating past enzyme survival.'),

    ('q-paper-014','Which writer questions whether acid-free paper lasts as long as is claimed?','["C"]','Writer C calls permanence claims optimistic.'),
    ('q-paper-015','Which writer focuses on the environmental cost of plantation pulpwood?','["A"]','Writer A argues the accounting excludes the forest replaced.'),
    ('q-paper-016','Which writer worked from the guild records of the Fabriano workshops?','["D"]','Writer D reconstructs the workshop from guild records.'),
    ('q-paper-017','Which writer credits cheap paper with the spread of literacy in Europe?','["B"]','Writer B: literacy follows the paper price.'),
    ('q-paper-018','Which writer argues that de-acidification treatments damage bindings?','["C"]','Writer C calls the remedies worse than the problem.'),
    ('q-paper-019','Which writer gave the first printed description of a continuous paper machine?','["E"]','Writer E describes the 1807 account drawn from the installation.')
  ) AS v(id, prompt, correct_answers, explanation)
 WHERE q.id = v.id;

UPDATE questions SET options = '[
    {"id":"A","text":"Writer A"},
    {"id":"B","text":"Writer B"},
    {"id":"C","text":"Writer C"},
    {"id":"D","text":"Writer D"},
    {"id":"E","text":"Writer E"}]'::jsonb
WHERE type_id = 'reading-find-the-writer' AND passage_id IS NOT NULL;

-- The commentaries the restored task is answered from, exactly as 000008 seeded them.

UPDATE reading_passages SET sources = $j$[
 {
  "label": "Writer A",
  "text": "The bar on the shelf hides a supply chain nobody has fixed. Roughly two million children work on West African cocoa farms, most of them on family land, and every certification scheme so far has audited a fraction of it. Ethics is not a marketing category here. It is the central unresolved problem of modern cocoa farming."
 },
 {
  "label": "Writer B",
  "text": "Chocolate did not become popular because it improved. It became popular because European industrialisation made it cheap. The press, the steam mill and the railway between them turned a court luxury into a factory product, and the taste of the age followed the price down."
 },
 {
  "label": "Writer C",
  "text": "The flavanols in cacao are among the best-evidenced dietary compounds we have for cardiovascular benefit, with measurable effects on blood pressure and on endothelial function. What the same trials show, and what the confectionery industry does not advertise, is that the sugar loaded into a commercial bar undoes the natural antioxidant properties the cacao arrived with."
 },
 {
  "label": "Writer D",
  "text": "The Florentine Codex and the Nahuatl sources let us reconstruct what the Aztecs actually did with cacao. It was poured at betrothals, drunk by warriors before battle and offered at funerals, and the ceremonial record they left is far richer than the culinary one."
 },
 {
  "label": "Writer E",
  "text": "Our house records for 1847 hold the first recipe for a solid eating bar: cocoa powder, sugar, and cocoa butter returned to the mass in proportion, cast in a mould and left to set. It reads like a note to a colleague, which is what it was."
 }
]$j$::jsonb WHERE id = 'rp-choc-01';

UPDATE reading_passages SET sources = $j$[
 {
  "label": "Writer A",
  "text": "Every hive placed on a roof is thirty thousand competitors dropped into a forage base that was already thin. The species that lose are the solitary bees and bumblebees that carry the conservation risk, and adding honeybees to a city is not pollinator conservation however it is marketed."
 },
 {
  "label": "Writer B",
  "text": "The urban hive boom has a date and a cause. Colony collapse disorder was reported from 2006, the media coverage was relentless, and keeping bees became the one thing an ordinary reader could do about a problem they had just been told was catastrophic."
 },
 {
  "label": "Writer C",
  "text": "Assays of city honey consistently show lower pesticide residues than samples from arable districts, which is the opposite of what most people expect, and on that measure urban honey is the chemically cleaner product. It is also the more fragile one: heat it past the point where its enzymes survive and the properties it is valued for are gone."
 },
 {
  "label": "Writer D",
  "text": "Monastic accounts from the twelfth century onward record hives kept inside city walls for wax as much as for honey, and the cathedral fabric rolls let us follow the practice through five hundred years of it."
 },
 {
  "label": "Writer E",
  "text": "The design published in our bulletin was the first intended for an apartment roof rather than a field: a shallow footprint, a flight path angled away from neighbouring windows, and a stand two people can carry up a stairwell."
 }
]$j$::jsonb WHERE id = 'rp-bees-01';

UPDATE reading_passages SET sources = $j$[
 {
  "label": "Writer A",
  "text": "A pulp plantation is a monoculture, and the accounting that makes it look benign quietly excludes the forest it replaced. The environmental cost of plantation pulpwood is not the carbon in the standing trees. It is the biodiversity of whatever was there before."
 },
 {
  "label": "Writer B",
  "text": "Printing gets the credit, but printing without cheap paper is a workshop curiosity. It was the price of the sheet that put books into the hands of people who had never owned one, and the spread of literacy in Europe follows the paper price more closely than it follows the press."
 },
 {
  "label": "Writer C",
  "text": "Acid-free is a claim about pH on the day of testing, not a guarantee of centuries, and accelerated ageing trials keep finding permanence claims optimistic. The remedies are worse: mass de-acidification treatments leave bindings weakened and sheets unevenly treated, and libraries have paid a great deal for the privilege."
 },
 {
  "label": "Writer D",
  "text": "The Fabriano guild records survive almost intact, and from them the workshop can be reconstructed hammer by hammer: who owned which vat, what a journeyman was paid, and the year the watermark first appears in the accounts."
 },
 {
  "label": "Writer E",
  "text": "The description we published in 1807 was, so far as we can establish, the first account in print of a machine forming paper as a continuous web, drawn from the installation itself rather than from the patent."
 }
]$j$::jsonb WHERE id = 'rp-paper-01';
