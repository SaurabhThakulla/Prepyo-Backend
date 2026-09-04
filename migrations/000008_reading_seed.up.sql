-- Three IELTS Academic reading passages, each carrying the full spread of task
-- types, which is what makes them eligible for a generated mock.
--
-- Every passage is built to the same shape:
--
--   paragraphs A-J   the body; the labels are the answers to Matching Information
--   sources A-E      attributed excerpts; the answers to Find the Writer
--   8 groups         6 task types, two of them appearing twice, as a real
--                    IELTS passage does
--   40 questions     7 / 6 / 6 / 4 / 3 / 6 / 4 / 4
--
-- Answers are worth one point each so that a paper of any length scores as the
-- share of it the learner got right.
--
-- Questions are inserted through a join rather than spelled out column by
-- column: exam, skill, type and passage all come from the group the question
-- belongs to, so a question cannot end up claiming a type its group does not
-- have. Shared answer pools are applied per type at the end of the file for the
-- same reason — one statement decides what True/False means everywhere.

-- ---------------------------------------------------------------------------
-- Passage 1: chocolate
-- ---------------------------------------------------------------------------

INSERT INTO reading_passages (id, exam_version_id, exam, title, subtitle, paragraphs, sources, word_count, difficulty, topic, tags) VALUES
('rp-choc-01', 'ielts-2026-01', 'IELTS', 'A History of Chocolate',
 'From a bitter ceremonial drink to a regulated commodity',
 $j$[
  {"label":"A","text":"Long before chocolate became a confection it was a drink, and before it was a drink it was a seed nobody had thought to ferment. The Olmec of the Mexican Gulf coast are the earliest people known to have processed the pods of Theobroma cacao, and residue analysis of their vessels pushes the practice back some three and a half thousand years. The Maya who followed them left a fuller record. Their texts describe cacao as a gift from the gods, granted to humanity at the founding of the world, and their preparation of it would be almost unrecognisable today: the roasted seeds were ground with water, chilli and maize flour, then poured from vessel to vessel until a thick head of foam formed. It was a bitter beverage, and it was meant to be."},
  {"label":"B","text":"Under the Aztecs the seed acquired a second life as money. Cacao beans were counted out for cloth, for canoes and for labour, and an early colonial source prices a rabbit at ten beans and a turkey at a hundred. Because the beans were so valuable they were used as currency across much of Mesoamerica, and because they were currency they were also forged, hollowed out and packed with earth. The drink itself stayed ceremonial and largely aristocratic. Bernal Diaz del Castillo, who watched Montezuma dine, reported that the emperor drank fifty cups of chocolate a day, served in gold vessels that were thrown away after a single use."},
  {"label":"C","text":"Europe met cacao by accident. Christopher Columbus was the first European to encounter cacao beans, in 1502, when his men boarded a Maya trading canoe off the coast of Honduras and noticed that the crew scrambled after the spilled beans as though they were coins. He carried some back to Spain, where nobody knew what to do with them. Cortes and the Spanish who followed understood the drink better, and what they mainly understood was that it needed fixing. The Spanish added sugar and cinnamon to make the drink more palatable, served it hot rather than cold, and kept it to themselves. For most of a century the recipe circulated only among European royalty and the religious houses that prepared it, guarded closely enough that chocolate reached the French court as a curiosity rather than a commodity."},
  {"label":"D","text":"The secret did not hold. By the middle of the seventeenth century chocolate had reached England, and chocolate houses became popular meeting places in seventeenth-century London, where a bowl cost more than a labourer earned in a day and the company ran to merchants, gamblers and political factions who found the rooms convenient for talking. What was drunk in them bore no resemblance to a modern bar. It was dark and, by present standards, barely sweetened, gritty with the fat that would not stay in suspension, and it would be two hundred years before anyone thought to put milk in it."},
  {"label":"E","text":"Industrialisation changed chocolate more in seventy years than the preceding three thousand had. In 1828 the Dutch chemist Coenraad van Houten patented a press that squeezed the fat out of the ground bean: the invention of the cocoa press made it possible to separate cocoa butter from the roasted bean, leaving a powder that mixed cleanly with liquid and a fat that could be added back where it was wanted. Far from slowing production, steam power and the factory system multiplied it. In 1847 the English firm of Fry worked cocoa powder, sugar and melted cocoa butter into a paste that could be cast in a mould, and the first modern solid chocolate bar was created in the year 1847. Milk followed. To create milk chocolate, Daniel Peter used powdered milk invented by Henri Nestle, and the bar he produced in 1875 is why Switzerland is still recognised for perfecting milk chocolate. Rodolphe Lindt added the last step four years later: the conching process improves the texture of chocolate by reducing its acidity and grinding the particles below the threshold the tongue can feel."},
  {"label":"F","text":"The tree itself has never been industrialised. Theobroma cacao will not grow in dry or desert conditions. It needs deep shade, constant humidity and a band of about twenty degrees either side of the equator, and yields swing sharply with rainfall, temperature and the length of the dry season, so a single poor year moves the world price. The majority of the world cocoa crop today is grown in West Africa, principally in Cote d Ivoire and Ghana, on smallholdings of a few hectares. Cacao pods are harvested by hand using a long-handled machete, because the pods grow directly from the trunk and the older wood must not be damaged. After harvesting, the beans must be fermented for several days to develop their flavour; unfermented beans taste of almost nothing, and no amount of later processing recovers what the fermentation heap did not make."},
  {"label":"G","text":"Roasting is where the maker judgement first shows. Temperatures vary considerably with origin and intention: a fine-flavour bean from Ecuador may be taken to only 110 degrees Celsius to keep its floral character, while a bulk West African bean bound for a confectionery filling is often pushed past 150, where the sugars caramelise and the sharper acids are driven off. Time and temperature trade against each other, and two roasters working from the same sack will agree on neither."},
  {"label":"H","text":"Why chocolate is craved rather than merely liked has attracted more research than it has settled. The bean carries theobromine and caffeine, both mild stimulants, along with phenylethylamine and tryptophan. The last of these is the precursor from which the body builds serotonin, and it is the compound most often credited with the lift that follows a bar. The quantities are small, and some researchers argue the effect owes more to fat, sugar and the melting point of cocoa butter, which sits just below body temperature, than to any pharmacology at all."},
  {"label":"I","text":"The twentieth century made chocolate ordinary. During the Second World War chocolate was included in soldiers rations, in a formulation deliberately made to taste, in the specification own words, only a little better than a boiled potato, so that it would not be eaten before it was needed. The mass production that followed traded flavour for shelf life and cost, and the result is a product I find hard to defend: industrial chocolate is a lesser thing than the bars now made by small producers who buy beans by origin, roast in small batches and publish what they paid. The artisanal revival is dismissed as nostalgia, but a side-by-side tasting settles the argument faster than any essay."},
  {"label":"J","text":"Regulation has followed, slowly. European law sets a legal definition for dark chocolate, requiring a minimum of 35 per cent cocoa solids of which at least 18 per cent must be cocoa butter, and the same directive is why white chocolate contains no cocoa solids at all: it is cocoa butter, milk and sugar, and its place in the category is still contested. Certification covers the other end of the chain. Fair Trade certification guarantees farmers a minimum price that does not fall when the commodity market does, and demand for certified organic chocolate has risen steadily over the past decade rather than fallen, though the two together still account for a small fraction of the trade."}
 ]$j$::jsonb,
 $j$[
  {"label":"Writer A","text":"The bar on the shelf hides a supply chain nobody has fixed. Roughly two million children work on West African cocoa farms, most of them on family land, and every certification scheme so far has audited a fraction of it. Ethics is not a marketing category here. It is the central unresolved problem of modern cocoa farming."},
  {"label":"Writer B","text":"Chocolate did not become popular because it improved. It became popular because European industrialisation made it cheap. The press, the steam mill and the railway between them turned a court luxury into a factory product, and the taste of the age followed the price down."},
  {"label":"Writer C","text":"The flavanols in cacao are among the best-evidenced dietary compounds we have for cardiovascular benefit, with measurable effects on blood pressure and on endothelial function. What the same trials show, and what the confectionery industry does not advertise, is that the sugar loaded into a commercial bar undoes the natural antioxidant properties the cacao arrived with."},
  {"label":"Writer D","text":"The Florentine Codex and the Nahuatl sources let us reconstruct what the Aztecs actually did with cacao. It was poured at betrothals, drunk by warriors before battle and offered at funerals, and the ceremonial record they left is far richer than the culinary one."},
  {"label":"Writer E","text":"Our house records for 1847 hold the first recipe for a solid eating bar: cocoa powder, sugar, and cocoa butter returned to the mass in proportion, cast in a mould and left to set. It reads like a note to a colleague, which is what it was."}
 ]$j$::jsonb,
 980, 'medium', 'Food history', ARRAY['IELTS Reading', 'History', 'Food science']);

-- ---------------------------------------------------------------------------
-- Passage 2: urban beekeeping
-- ---------------------------------------------------------------------------

INSERT INTO reading_passages (id, exam_version_id, exam, title, subtitle, paragraphs, sources, word_count, difficulty, topic, tags) VALUES
('rp-bees-01', 'ielts-2026-01', 'IELTS', 'The Return of Urban Beekeeping',
 'Why cities suit honeybees, and what that costs everything else',
 $j$[
  {"label":"A","text":"Bees have been kept in cities for as long as there have been cities to keep them in, but for most of the twentieth century the practice was quietly outlawed. New York City classified the honeybee alongside ferrets and venomous snakes on its list of prohibited animals, and the ban stood until it was lifted in 2010. Within two years the number of registered hives in the five boroughs had trebled. Paris, which had never troubled to prohibit anything, already had hives on the roof of the Opera Garnier, where they had been producing honey since 1985."},
  {"label":"B","text":"The surprise of the urban hive is that it works at all, and the reason it does is variety. A colony in an arable landscape may face a single crop in flower for three weeks and nothing whatever for the rest of the season. A colony on a city roof forages across gardens, parks, street trees, railway embankments and window boxes, which between them flower from February to November. Cities are also warmer than the countryside around them, since the urban heat island can lift winter minima by several degrees, and they carry a far lighter load of agricultural pesticide."},
  {"label":"C","text":"What turned a curiosity into a movement was alarm. From 2006 beekeepers in North America and Europe began reporting colonies that emptied without explanation, the workers simply failing to return. The phenomenon was named colony collapse disorder, and although its causes are now understood to be several rather than one, the coverage at the time was singular and loud. Keeping a hive became something an ordinary person could do about it."},
  {"label":"D","text":"The economics are modest and are usually misunderstood. A complete hive with bees costs between four and six hundred pounds to establish, and a healthy urban colony yields perhaps twenty kilograms of surplus honey in a good season and considerably less in a poor one. Very few urban beekeepers recover their costs. Almost none make a living."},
  {"label":"E","text":"Enthusiasm has produced its own problem. Hive density in some London boroughs passed ten colonies per square kilometre, well above the level the available forage can support, and a managed honeybee colony is a competitor before it is anything else. Wild bees, the solitary and bumblebee species that do most of the pollination and carry most of the conservation risk, lose out first when nectar runs short. Adding hives is not, on its own, conservation."},
  {"label":"F","text":"Weather governs the year more than any decision the keeper makes. Colony survival over winter and the size of the summer surplus both track temperature and rainfall closely: a cold wet spring keeps foragers indoors through the weeks when the colony must build up, and a drought shuts down the nectar flow at the height of the season. Yields between neighbouring apiaries can differ by half on nothing more than aspect and shelter."},
  {"label":"G","text":"Disease management is the part nobody advertises. The varroa mite reached Europe in the 1970s and is now effectively universal; untreated colonies rarely last three years. Treatment turns on temperature. Oxalic acid is applied when the colony is broodless in midwinter, thymol works only above about 15 degrees Celsius, and a hive treated at the wrong point in the season is a hive treated for nothing."},
  {"label":"H","text":"What ends up in the jar is decided by the flower and by the bee. Nectar sugars are inverted by enzymes the forager adds on the way home, chiefly invertase, and the water content is driven below twenty per cent by fanning before the cell is capped. Colour and flavour follow the forage: lime gives a pale honey with a green note, chestnut a dark and almost bitter one, and an urban honey is usually a blend of dozens of sources, no two years alike."},
  {"label":"I","text":"Cities have begun to regulate, and they are right to. Registration is what allows disease to be traced, and a compulsory register is the single measure I would put ahead of any campaign to install more hives; voluntary schemes reach the beekeepers who least need reaching and miss the ones who most do. Training requirements are harder to justify in law but easier to defend in practice, because the commonest cause of a dead colony is an owner who has read one book."},
  {"label":"J","text":"Labelling has not kept pace. There is no legal definition of urban honey anywhere in the European Union, and the term carries no requirement about where the bees foraged or how far they flew. The rules that do exist concern the product itself: honey may have nothing added to it and nothing taken away, and heating it past the point where its enzymes are destroyed costs it the name. Demand for local and traceable honey has grown every year for a decade, which is precisely why the absence of a definition matters."}
 ]$j$::jsonb,
 $j$[
  {"label":"Writer A","text":"Every hive placed on a roof is thirty thousand competitors dropped into a forage base that was already thin. The species that lose are the solitary bees and bumblebees that carry the conservation risk, and adding honeybees to a city is not pollinator conservation however it is marketed."},
  {"label":"Writer B","text":"The urban hive boom has a date and a cause. Colony collapse disorder was reported from 2006, the media coverage was relentless, and keeping bees became the one thing an ordinary reader could do about a problem they had just been told was catastrophic."},
  {"label":"Writer C","text":"Assays of city honey consistently show lower pesticide residues than samples from arable districts, which is the opposite of what most people expect, and on that measure urban honey is the chemically cleaner product. It is also the more fragile one: heat it past the point where its enzymes survive and the properties it is valued for are gone."},
  {"label":"Writer D","text":"Monastic accounts from the twelfth century onward record hives kept inside city walls for wax as much as for honey, and the cathedral fabric rolls let us follow the practice through five hundred years of it."},
  {"label":"Writer E","text":"The design published in our bulletin was the first intended for an apartment roof rather than a field: a shallow footprint, a flight path angled away from neighbouring windows, and a stand two people can carry up a stairwell."}
 ]$j$::jsonb,
 900, 'medium', 'Environment', ARRAY['IELTS Reading', 'Environment', 'Urban ecology']);

-- ---------------------------------------------------------------------------
-- Passage 3: papermaking
-- ---------------------------------------------------------------------------

INSERT INTO reading_passages (id, exam_version_id, exam, title, subtitle, paragraphs, sources, word_count, difficulty, topic, tags) VALUES
('rp-paper-01', 'ielts-2026-01', 'IELTS', 'The Long Road of Papermaking',
 'Two thousand years from bark and rags to a defined standard',
 $j$[
  {"label":"A","text":"The invention is dated, unusually for something so old, to a single year and a single official. In 105 CE the court eunuch Cai Lun presented to the Han emperor a writing material made from macerated bark, hemp waste, old rags and fishing nets, and the Chinese histories record the presentation as the moment paper began. The date is almost certainly too late, since fragments predating it by two centuries have been excavated, but the contribution was real: he made the process cheap."},
  {"label":"B","text":"Paper travelled west slowly and by war. The conventional account has the technique passing to the Islamic world after the battle of Talas in 751, when Chinese prisoners taken to Samarkand were said to include papermakers. Whether or not the story is true, Samarkand was making paper within a generation, Baghdad had a paper mill by the end of the century, and the administrative culture of the Abbasid caliphate ran on it."},
  {"label":"C","text":"Europe was the last of the old world to take it up. Paper reached Islamic Spain in the twelfth century and Italy in the thirteenth, where the mills at Fabriano added three things the Chinese had not: water-powered stamping hammers, animal gelatine sizing, and the wire watermark. Fabriano paper was better than what it copied, and Italy exported it for three hundred years."},
  {"label":"D","text":"For seven centuries European paper was made from linen and cotton rags, and the supply of rags was the ceiling on the whole industry. Mills advertised for them, parishes collected them, some countries forbade their export and others forbade burying the dead in linen. By the early nineteenth century the shortage was acute enough that chemists were paid to look at straw, nettles, thistles and, following a suggestion drawn from watching wasps build nests, wood."},
  {"label":"E","text":"Two inventions ended the rag age. In 1806 the Fourdrinier brothers financed a machine that formed paper as a continuous web rather than one sheet at a time, and in 1844 the German weaver Friedrich Gottlob Keller patented a process for grinding logs against a wet stone to make pulp. Wood was effectively unlimited, and within fifty years it had displaced rag almost entirely."},
  {"label":"F","text":"Fibre is not interchangeable. Softwoods give long fibres and strong paper, hardwoods give short ones and a smoother surface, and the length of the growing season, the rainfall and the spacing of the trees all change the fibre yield a hectare returns. A plantation on a poor site may take twice as long to reach the yield of one on a good site, which is why pulp economics are geography before they are anything else."},
  {"label":"G","text":"Pulping is chemistry at temperature. The kraft process cooks chips in sodium hydroxide and sodium sulphide at about 170 degrees Celsius, mechanical pulping grinds at close to 100 and keeps far more of the wood, and bleaching sequences run cooler, between 60 and 90 depending on the stage. Every ten degrees costs energy and buys brightness, and that trade is the industry central argument with itself."},
  {"label":"H","text":"The reason nineteenth-century books are crumbling is chemical. Wood pulp carries lignin, which yellows and embrittles in light, and the alum and rosin sizing adopted in the same period left the sheet acidic; the acid slowly hydrolyses the cellulose until the paper snaps rather than folds. A book printed in 1550 on rag is generally in better condition than one printed in 1890."},
  {"label":"I","text":"Recycling is where the romance stops. Fibre shortens each time it is pulped and can be recycled perhaps five to seven times before it is fit only for board, so a recycled sheet always depends on new fibre entering the system somewhere. That is not an argument against recycling, and the people who make it usually have something to sell. It is an argument for being honest about what recycling is."},
  {"label":"J","text":"Standards arrived after the damage. The ISO permanence standard sets out what a paper must satisfy to be expected to last several hundred years, and the term acid-free is defined by pH: a sheet measuring 7.0 or above when tested cold. Recycled content, by contrast, is governed by no single international rule, and demand for certified recycled office paper has climbed steadily since the 1990s."}
 ]$j$::jsonb,
 $j$[
  {"label":"Writer A","text":"A pulp plantation is a monoculture, and the accounting that makes it look benign quietly excludes the forest it replaced. The environmental cost of plantation pulpwood is not the carbon in the standing trees. It is the biodiversity of whatever was there before."},
  {"label":"Writer B","text":"Printing gets the credit, but printing without cheap paper is a workshop curiosity. It was the price of the sheet that put books into the hands of people who had never owned one, and the spread of literacy in Europe follows the paper price more closely than it follows the press."},
  {"label":"Writer C","text":"Acid-free is a claim about pH on the day of testing, not a guarantee of centuries, and accelerated ageing trials keep finding permanence claims optimistic. The remedies are worse: mass de-acidification treatments leave bindings weakened and sheets unevenly treated, and libraries have paid a great deal for the privilege."},
  {"label":"Writer D","text":"The Fabriano guild records survive almost intact, and from them the workshop can be reconstructed hammer by hammer: who owned which vat, what a journeyman was paid, and the year the watermark first appears in the accounts."},
  {"label":"Writer E","text":"The description we published in 1807 was, so far as we can establish, the first account in print of a machine forming paper as a continuous web, drawn from the installation itself rather than from the patent."}
 ]$j$::jsonb,
 940, 'hard', 'Technology history', ARRAY['IELTS Reading', 'History', 'Technology']);

-- ---------------------------------------------------------------------------
-- Groups
-- ---------------------------------------------------------------------------
--
-- Eight groups per passage over six types. Sentence completion and Yes/No/Not
-- Given each appear twice, which is why type is not unique within a passage.
--
-- shuffle_questions is FALSE only for the ordering task, where the position of
-- a question in the set is the question.

INSERT INTO reading_question_groups (id, passage_id, position, type_id, type_name, instructions, resources, shuffle_questions, time_limit_seconds) VALUES

-- Passage 1
('g-choc-1', 'rp-choc-01', 1, 'reading-sentence-completion', 'Sentence Completion',
 'Complete the sentences below. Write ONE WORD ONLY from the passage in each gap.', '[]'::jsonb, TRUE, 480),
('g-choc-2', 'rp-choc-01', 2, 'reading-true-false', 'True / False',
 'Do the following statements agree with the information in the passage? Answer True or False.', '[]'::jsonb, TRUE, 420),
('g-choc-3', 'rp-choc-01', 3, 'reading-find-the-writer', 'Find the Writer',
 'Look at the five commentaries, Writer A to Writer E. Which writer makes each of the following statements?', '[]'::jsonb, TRUE, 420),
('g-choc-4', 'rp-choc-01', 4, 'reading-arrange-passage', 'Arrange the Passage',
 'The four boxes below summarise stages in the history of chocolate but are printed out of order. The questions are listed in the correct chronological order: choose the box that belongs at each position.',
 $j$[
  {"label":"Paragraph A","text":"Spanish ships carry the beans home, sugar and cinnamon are stirred into the drink, and it circulates among European courts as a guarded recipe."},
  {"label":"Paragraph B","text":"Small producers buy beans by origin, publish what they paid for them, and certification schemes attach a floor price to the farmer end of the chain."},
  {"label":"Paragraph C","text":"On the Mexican Gulf coast the pods of a shade-loving tree are fermented, roasted and ground for the first time, three and a half thousand years ago."},
  {"label":"Paragraph D","text":"A Dutch chemist presses the fat out of the ground bean, leaving a powder that mixes with liquid and a butter that can be added back where it is wanted."}
 ]$j$::jsonb, FALSE, 300),
('g-choc-5', 'rp-choc-01', 5, 'reading-yes-no-not-given', 'Yes / No / Not Given',
 'Do the following statements agree with the views of the writer? Answer Yes, No or Not Given.', '[]'::jsonb, TRUE, 240),
('g-choc-6', 'rp-choc-01', 6, 'reading-sentence-completion', 'Sentence Completion',
 'Complete the summary below. Write ONE WORD OR NUMBER from the passage in each gap.', '[]'::jsonb, TRUE, 420),
('g-choc-7', 'rp-choc-01', 7, 'reading-matching-information', 'Matching Information',
 'The passage has ten paragraphs, A to J. Which paragraph contains each of the following?', '[]'::jsonb, TRUE, 300),
('g-choc-8', 'rp-choc-01', 8, 'reading-yes-no-not-given', 'Yes / No / Not Given',
 'Do the following statements agree with the views of the writer? Answer Yes, No or Not Given.', '[]'::jsonb, TRUE, 300),

-- Passage 2
('g-bees-1', 'rp-bees-01', 1, 'reading-sentence-completion', 'Sentence Completion',
 'Complete the sentences below. Write ONE WORD OR NUMBER from the passage in each gap.', '[]'::jsonb, TRUE, 480),
('g-bees-2', 'rp-bees-01', 2, 'reading-true-false', 'True / False',
 'Do the following statements agree with the information in the passage? Answer True or False.', '[]'::jsonb, TRUE, 420),
('g-bees-3', 'rp-bees-01', 3, 'reading-find-the-writer', 'Find the Writer',
 'Look at the five commentaries, Writer A to Writer E. Which writer makes each of the following statements?', '[]'::jsonb, TRUE, 420),
('g-bees-4', 'rp-bees-01', 4, 'reading-arrange-passage', 'Arrange the Passage',
 'The four boxes below summarise stages in the story of urban beekeeping but are printed out of order. The questions are listed in the correct chronological order: choose the box that belongs at each position.',
 $j$[
  {"label":"Paragraph A","text":"Colonies begin emptying without explanation, the reporting is relentless, and a minority hobby acquires a public cause."},
  {"label":"Paragraph B","text":"Councils start registering hives and asking out loud whether a borough can carry the density it now has."},
  {"label":"Paragraph C","text":"Hives sit on opera house and warehouse roofs for decades, unremarked, long before anyone calls it a movement."},
  {"label":"Paragraph D","text":"Municipal prohibitions are repealed and the number of registered colonies trebles inside two years."}
 ]$j$::jsonb, FALSE, 300),
('g-bees-5', 'rp-bees-01', 5, 'reading-yes-no-not-given', 'Yes / No / Not Given',
 'Do the following statements agree with the views of the writer? Answer Yes, No or Not Given.', '[]'::jsonb, TRUE, 240),
('g-bees-6', 'rp-bees-01', 6, 'reading-sentence-completion', 'Sentence Completion',
 'Complete the summary below. Write ONE WORD OR NUMBER from the passage in each gap.', '[]'::jsonb, TRUE, 420),
('g-bees-7', 'rp-bees-01', 7, 'reading-matching-information', 'Matching Information',
 'The passage has ten paragraphs, A to J. Which paragraph contains each of the following?', '[]'::jsonb, TRUE, 300),
('g-bees-8', 'rp-bees-01', 8, 'reading-yes-no-not-given', 'Yes / No / Not Given',
 'Do the following statements agree with the views of the writer? Answer Yes, No or Not Given.', '[]'::jsonb, TRUE, 300),

-- Passage 3
('g-paper-1', 'rp-paper-01', 1, 'reading-sentence-completion', 'Sentence Completion',
 'Complete the sentences below. Write ONE WORD OR NUMBER from the passage in each gap.', '[]'::jsonb, TRUE, 480),
('g-paper-2', 'rp-paper-01', 2, 'reading-true-false', 'True / False',
 'Do the following statements agree with the information in the passage? Answer True or False.', '[]'::jsonb, TRUE, 420),
('g-paper-3', 'rp-paper-01', 3, 'reading-find-the-writer', 'Find the Writer',
 'Look at the five commentaries, Writer A to Writer E. Which writer makes each of the following statements?', '[]'::jsonb, TRUE, 420),
('g-paper-4', 'rp-paper-01', 4, 'reading-arrange-passage', 'Arrange the Passage',
 'The four boxes below summarise stages in the history of paper but are printed out of order. The questions are listed in the correct chronological order: choose the box that belongs at each position.',
 $j$[
  {"label":"Paragraph A","text":"Mills advertise for linen, parishes collect it, and governments legislate over it, because the rag heap is the only thing paper can be made from."},
  {"label":"Paragraph B","text":"After a century of books that crumble on the shelf, committees write down what permanence means and give acid-free a number."},
  {"label":"Paragraph C","text":"A court official lays a sheet of macerated bark and rag before an emperor and, more importantly, makes it cheap to produce."},
  {"label":"Paragraph D","text":"A machine forms paper as an endless web and a weaver grinds logs against a wet stone, and the shortage simply ends."}
 ]$j$::jsonb, FALSE, 300),
('g-paper-5', 'rp-paper-01', 5, 'reading-yes-no-not-given', 'Yes / No / Not Given',
 'Do the following statements agree with the views of the writer? Answer Yes, No or Not Given.', '[]'::jsonb, TRUE, 240),
('g-paper-6', 'rp-paper-01', 6, 'reading-sentence-completion', 'Sentence Completion',
 'Complete the summary below. Write ONE WORD OR NUMBER from the passage in each gap.', '[]'::jsonb, TRUE, 420),
('g-paper-7', 'rp-paper-01', 7, 'reading-matching-information', 'Matching Information',
 'The passage has ten paragraphs, A to J. Which paragraph contains each of the following?', '[]'::jsonb, TRUE, 300),
('g-paper-8', 'rp-paper-01', 8, 'reading-yes-no-not-given', 'Yes / No / Not Given',
 'Do the following statements agree with the views of the writer? Answer Yes, No or Not Given.', '[]'::jsonb, TRUE, 300);

-- ---------------------------------------------------------------------------
-- Questions
-- ---------------------------------------------------------------------------
--
-- Exam, skill, type and passage are taken from the group, so a question cannot
-- claim a type its group does not have. correct_answers holds every accepted
-- spelling for a typed gap and the option id for a choice.

INSERT INTO questions (
    id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
    correct_answers, explanation, time_limit_seconds, points, difficulty, tags,
    passage_id, group_id, group_position)
SELECT v.id, p.exam_version_id, p.exam, 'reading', g.type_id, g.type_name,
       g.type_name || ' ' || v.position::text, v.prompt,
       v.correct_answers::jsonb, v.explanation, 60, 1, p.difficulty,
       ARRAY['IELTS Reading', g.type_name],
       p.id, g.id, v.position
FROM (VALUES

-- Passage 1, group 1: sentence completion
('q-choc-001','g-choc-1',1,'The ancient Mayans believed that cacao beans were a gift from the ________.','["gods"]','Paragraph A: cacao is described as a gift from the gods.'),
('q-choc-002','g-choc-1',2,'Initially, chocolate was primarily consumed as a ________ beverage.','["bitter"]','Paragraph A: it was a bitter beverage, and it was meant to be.'),
('q-choc-003','g-choc-1',3,'In Aztec society, cacao beans were so valuable they were used as ________.','["currency"]','Paragraph B: used as currency across much of Mesoamerica.'),
('q-choc-004','g-choc-1',4,'The Spanish added sugar and ________ to make the drink more palatable.','["cinnamon"]','Paragraph C names sugar and cinnamon.'),
('q-choc-005','g-choc-1',5,'Chocolate houses became popular meeting places in seventeenth-century ________.','["London"]','Paragraph D: chocolate houses in seventeenth-century London.'),
('q-choc-006','g-choc-1',6,'The invention of the cocoa press made it possible to separate cocoa ________ from the roasted bean.','["butter"]','Paragraph E: the press separates cocoa butter from the bean.'),
('q-choc-007','g-choc-1',7,'The first modern solid chocolate bar was created in the year ________.','["1847"]','Paragraph E: Fry cast the first solid bar in 1847.'),

-- Passage 1, group 2: true / false
('q-choc-008','g-choc-2',1,'Cacao trees can only grow in dry, desert climates.','["FALSE"]','Paragraph F: cacao will not grow in dry or desert conditions.'),
('q-choc-009','g-choc-2',2,'Christopher Columbus was the first European to encounter cacao beans.','["TRUE"]','Paragraph C states this directly, in 1502.'),
('q-choc-010','g-choc-2',3,'Early European royalty kept the recipe for chocolate a secret.','["TRUE"]','Paragraph C: the recipe circulated only among royalty and religious houses.'),
('q-choc-011','g-choc-2',4,'Milk chocolate was invented before dark chocolate.','["FALSE"]','Paragraphs D and E: dark came first, milk chocolate followed in 1875.'),
('q-choc-012','g-choc-2',5,'The industrial revolution slowed down the production of chocolate.','["FALSE"]','Paragraph E: far from slowing production, it multiplied it.'),
('q-choc-013','g-choc-2',6,'Switzerland is historically recognised for perfecting milk chocolate.','["TRUE"]','Paragraph E credits Switzerland with perfecting milk chocolate.'),

-- Passage 1, group 3: find the writer
('q-choc-014','g-choc-3',1,'Which writer claims that chocolate has significant cardiovascular benefits?','["C"]','Writer C cites flavanols and cardiovascular benefit.'),
('q-choc-015','g-choc-3',2,'Which writer focuses on the ethical concerns of modern cocoa farming?','["A"]','Writer A calls ethics the central unresolved problem.'),
('q-choc-016','g-choc-3',3,'Which writer documented the ceremonial use of cacao by the Aztecs?','["D"]','Writer D works from the Florentine Codex and Nahuatl sources.'),
('q-choc-017','g-choc-3',4,'Which writer attributes the popularisation of chocolate to European industrialisation?','["B"]','Writer B: it became popular because industrialisation made it cheap.'),
('q-choc-018','g-choc-3',5,'Which writer argues that sugar ruins the natural antioxidant properties of cacao?','["C"]','Writer C says the sugar undoes the antioxidant properties.'),
('q-choc-019','g-choc-3',6,'Which writer provided the first recipe for a solid chocolate bar?','["E"]','Writer E cites house records for 1847.'),

-- Passage 1, group 4: arrange the passage
('q-choc-020','g-choc-4',1,'Position 1: the first use of cacao by the Olmecs.','["C"]','Box C describes the Gulf coast origin, the earliest stage.'),
('q-choc-021','g-choc-4',2,'Position 2: the Spanish conquest and the introduction of chocolate to Europe.','["A"]','Box A covers the Spanish ships and the guarded court recipe.'),
('q-choc-022','g-choc-4',3,'Position 3: the invention of the cocoa press during the Industrial Revolution.','["D"]','Box D describes van Houten pressing the fat from the bean.'),
('q-choc-023','g-choc-4',4,'Position 4: modern artisanal chocolate making and fair trade.','["B"]','Box B covers small producers and certification, the latest stage.'),

-- Passage 1, group 5: yes / no / not given
('q-choc-024','g-choc-5',1,'The author believes that modern mass-produced chocolate is inferior to artisanal chocolate.','["YES"]','Paragraph I: industrial chocolate is a lesser thing.'),
('q-choc-025','g-choc-5',2,'Chocolate was primarily consumed by children in the eighteenth century.','["NOT_GIVEN"]','The passage says nothing about who drank it in the eighteenth century.'),
('q-choc-026','g-choc-5',3,'The Aztec emperor Montezuma drank fifty cups of chocolate a day.','["YES"]','Paragraph B reports the fifty cups.'),

-- Passage 1, group 6: sentence completion
('q-choc-027','g-choc-6',1,'To create milk chocolate, Daniel Peter used powdered milk invented by Henri ________.','["Nestle","Nestlé"]','Paragraph E names Henri Nestle.'),
('q-choc-028','g-choc-6',2,'The conching process improves the texture of chocolate by reducing its ________.','["acidity"]','Paragraph E: conching reduces acidity.'),
('q-choc-029','g-choc-6',3,'During the Second World War, chocolate was included in soldiers ________.','["rations"]','Paragraph I: included in soldiers rations.'),
('q-choc-030','g-choc-6',4,'The majority of the world cocoa crop today is grown in West ________.','["Africa"]','Paragraph F: principally West Africa.'),
('q-choc-031','g-choc-6',5,'Cacao pods are harvested by hand using a long-handled ________.','["machete"]','Paragraph F names the machete.'),
('q-choc-032','g-choc-6',6,'After harvesting, the beans must be fermented for several days to develop their ________.','["flavour","flavor"]','Paragraph F: fermented to develop flavour.'),

-- Passage 1, group 7: matching information
('q-choc-033','g-choc-7',1,'A description of how the climate affects cacao tree yields.','["F"]','Paragraph F ties yields to rainfall, temperature and dry season.'),
('q-choc-034','g-choc-7',2,'The chemical components in chocolate that trigger serotonin release.','["H"]','Paragraph H covers tryptophan and serotonin.'),
('q-choc-035','g-choc-7',3,'An explanation of the roasting temperature variations.','["G"]','Paragraph G contrasts 110 and 150 degrees.'),
('q-choc-036','g-choc-7',4,'Details regarding the legal definition of dark chocolate in Europe.','["J"]','Paragraph J gives the 35 per cent minimum.'),

-- Passage 1, group 8: yes / no / not given
('q-choc-037','g-choc-8',1,'White chocolate contains no cocoa solids.','["YES"]','Paragraph J states this directly.'),
('q-choc-038','g-choc-8',2,'The demand for organic chocolate has decreased in the last decade.','["NO"]','Paragraph J says demand has risen steadily.'),
('q-choc-039','g-choc-8',3,'Eating chocolate before bed improves sleep quality.','["NOT_GIVEN"]','Sleep is not discussed anywhere in the passage.'),
('q-choc-040','g-choc-8',4,'Fair Trade certification ensures that farmers receive a guaranteed minimum price.','["YES"]','Paragraph J: a minimum price that does not fall with the market.'),

-- Passage 2, group 1: sentence completion
('q-bees-001','g-bees-1',1,'New York City lifted its ban on beekeeping in the year ________.','["2010"]','Paragraph A dates the repeal to 2010.'),
('q-bees-002','g-bees-1',2,'Hives have produced honey on the roof of the Paris Opera ________ since 1985.','["Garnier"]','Paragraph A names the Opera Garnier.'),
('q-bees-003','g-bees-1',3,'A city colony forages across gardens, parks, street trees and railway ________.','["embankments"]','Paragraph B lists railway embankments.'),
('q-bees-004','g-bees-1',4,'Cities are warmer than the surrounding countryside because of the urban ________ island.','["heat"]','Paragraph B: the urban heat island.'),
('q-bees-005','g-bees-1',5,'Colonies that emptied without explanation gave rise to the term colony ________ disorder.','["collapse"]','Paragraph C names colony collapse disorder.'),
('q-bees-006','g-bees-1',6,'A healthy urban colony yields about twenty ________ of surplus honey in a good season.','["kilograms","kilogrammes"]','Paragraph D gives twenty kilograms.'),
('q-bees-007','g-bees-1',7,'Hive density in some London boroughs passed ten colonies per square ________.','["kilometre","kilometer"]','Paragraph E gives ten colonies per square kilometre.'),

-- Passage 2, group 2: true / false
('q-bees-008','g-bees-2',1,'Beekeeping was legal in New York City throughout the twentieth century.','["FALSE"]','Paragraph A: the honeybee was on the prohibited list until 2010.'),
('q-bees-009','g-bees-2',2,'Paris had rooftop hives before New York lifted its ban.','["TRUE"]','Paragraph A: the Opera Garnier hives date from 1985.'),
('q-bees-010','g-bees-2',3,'Colonies in arable landscapes have a longer flowering season than city colonies.','["FALSE"]','Paragraph B: arable colonies face one crop for three weeks.'),
('q-bees-011','g-bees-2',4,'Most urban beekeepers recover the cost of their equipment.','["FALSE"]','Paragraph D: very few recover their costs.'),
('q-bees-012','g-bees-2',5,'The varroa mite is now present almost everywhere in Europe.','["TRUE"]','Paragraph G: effectively universal since the 1970s.'),
('q-bees-013','g-bees-2',6,'Colony collapse disorder is now attributed to a single cause.','["FALSE"]','Paragraph C: the causes are several rather than one.'),

-- Passage 2, group 3: find the writer
('q-bees-014','g-bees-3',1,'Which writer argues that additional hives can harm wild pollinators?','["A"]','Writer A calls each hive thirty thousand competitors.'),
('q-bees-015','g-bees-3',2,'Which writer attributes the growth of urban beekeeping to media coverage?','["B"]','Writer B ties the boom to the reporting from 2006.'),
('q-bees-016','g-bees-3',3,'Which writer claims that urban honey is chemically cleaner than rural honey?','["C"]','Writer C cites lower pesticide residues in city honey.'),
('q-bees-017','g-bees-3',4,'Which writer documented the keeping of bees inside medieval city walls?','["D"]','Writer D works from monastic accounts and fabric rolls.'),
('q-bees-018','g-bees-3',5,'Which writer published the first hive design intended for an apartment roof?','["E"]','Writer E describes the bulletin design.'),
('q-bees-019','g-bees-3',6,'Which writer argues that heating honey destroys the properties it is valued for?','["C"]','Writer C warns about heating past enzyme survival.'),

-- Passage 2, group 4: arrange the passage
('q-bees-020','g-bees-4',1,'Position 1: the keeping of bees on city rooftops before the modern revival.','["C"]','Box C describes decades of unremarked rooftop hives.'),
('q-bees-021','g-bees-4',2,'Position 2: the reporting of colony collapse disorder.','["A"]','Box A covers the colonies emptying and the coverage.'),
('q-bees-022','g-bees-4',3,'Position 3: the lifting of city prohibitions and the rise in hive numbers.','["D"]','Box D describes repeal and the trebling of colonies.'),
('q-bees-023','g-bees-4',4,'Position 4: municipal registration and concern over hive density.','["B"]','Box B covers councils registering and questioning density.'),

-- Passage 2, group 5: yes / no / not given
('q-bees-024','g-bees-5',1,'The author believes that hive registration should be compulsory.','["YES"]','Paragraph I: a compulsory register is the measure the author would put first.'),
('q-bees-025','g-bees-5',2,'Urban beekeepers in Paris earn more than those in London.','["NOT_GIVEN"]','No comparison of earnings between cities is made.'),
('q-bees-026','g-bees-5',3,'Installing more hives is by itself an act of conservation.','["NO"]','Paragraph E: adding hives is not, on its own, conservation.'),

-- Passage 2, group 6: sentence completion
('q-bees-027','g-bees-6',1,'Oxalic acid is applied in midwinter when the colony is ________.','["broodless"]','Paragraph G: applied when the colony is broodless.'),
('q-bees-028','g-bees-6',2,'Thymol is effective only above about ________ degrees Celsius.','["15","fifteen"]','Paragraph G gives about 15 degrees.'),
('q-bees-029','g-bees-6',3,'Nectar sugars are inverted by enzymes the forager adds, chiefly ________.','["invertase"]','Paragraph H names invertase.'),
('q-bees-030','g-bees-6',4,'Bees drive the water content below twenty per cent by ________ before the cell is capped.','["fanning"]','Paragraph H: driven down by fanning.'),
('q-bees-031','g-bees-6',5,'Chestnut honey is dark and almost ________.','["bitter"]','Paragraph H describes chestnut honey as almost bitter.'),
('q-bees-032','g-bees-6',6,'The commonest cause of a dead colony is an owner who has read one ________.','["book"]','Paragraph I ends on exactly this.'),

-- Passage 2, group 7: matching information
('q-bees-033','g-bees-7',1,'A description of how weather governs colony survival and yields.','["F"]','Paragraph F ties survival and surplus to temperature and rainfall.'),
('q-bees-034','g-bees-7',2,'An explanation of the enzymes that act on nectar.','["H"]','Paragraph H covers invertase and inversion.'),
('q-bees-035','g-bees-7',3,'Details of the temperatures at which mite treatments work.','["G"]','Paragraph G contrasts oxalic acid and thymol.'),
('q-bees-036','g-bees-7',4,'A statement about the absence of a legal definition for a labelling term.','["J"]','Paragraph J: no legal definition of urban honey.'),

-- Passage 2, group 8: yes / no / not given
('q-bees-037','g-bees-8',1,'Honey may legally have ingredients added to it.','["NO"]','Paragraph J: nothing added and nothing taken away.'),
('q-bees-038','g-bees-8',2,'Demand for local and traceable honey has grown for a decade.','["YES"]','Paragraph J states this directly.'),
('q-bees-039','g-bees-8',3,'Urban honey sells for a higher price than rural honey.','["NOT_GIVEN"]','Price is never compared in the passage.'),
('q-bees-040','g-bees-8',4,'Voluntary registration schemes fail to reach the beekeepers who most need them.','["YES"]','Paragraph I: they reach the ones who least need reaching.'),

-- Passage 3, group 1: sentence completion
('q-paper-001','g-paper-1',1,'Cai Lun presented his writing material to the Han emperor in the year ________ CE.','["105"]','Paragraph A dates the presentation to 105 CE.'),
('q-paper-002','g-paper-1',2,'Cai Lun real contribution was that he made the process ________.','["cheap"]','Paragraph A: he made the process cheap.'),
('q-paper-003','g-paper-1',3,'Papermaking is said to have reached the Islamic world after the battle of ________.','["Talas"]','Paragraph B names the battle of Talas in 751.'),
('q-paper-004','g-paper-1',4,'By the end of the eighth century there was a paper mill in ________.','["Baghdad"]','Paragraph B: Baghdad had a mill by the end of the century.'),
('q-paper-005','g-paper-1',5,'The Italian mills at Fabriano introduced the wire ________.','["watermark"]','Paragraph C lists the wire watermark.'),
('q-paper-006','g-paper-1',6,'For seven centuries European paper was made from linen and cotton ________.','["rags"]','Paragraph D: made from linen and cotton rags.'),
('q-paper-007','g-paper-1',7,'The suggestion of using wood came from watching ________ build nests.','["wasps"]','Paragraph D credits watching wasps.'),

-- Passage 3, group 2: true / false
('q-paper-008','g-paper-2',1,'Paper was certainly first made in the year Cai Lun presented it.','["FALSE"]','Paragraph A: fragments predate it by two centuries.'),
('q-paper-009','g-paper-2',2,'Samarkand was producing paper within a generation of 751.','["TRUE"]','Paragraph B states this.'),
('q-paper-010','g-paper-2',3,'Fabriano paper was inferior to the Chinese paper it copied.','["FALSE"]','Paragraph C: it was better than what it copied.'),
('q-paper-011','g-paper-2',4,'Some countries banned the export of rags.','["TRUE"]','Paragraph D: some countries forbade their export.'),
('q-paper-012','g-paper-2',5,'Keller pulping process was patented before the Fourdrinier machine was financed.','["FALSE"]','Paragraph E: the machine came in 1806, the patent in 1844.'),
('q-paper-013','g-paper-2',6,'Wood had largely replaced rag within fifty years of Keller patent.','["TRUE"]','Paragraph E states this.'),

-- Passage 3, group 3: find the writer
('q-paper-014','g-paper-3',1,'Which writer questions whether acid-free paper lasts as long as is claimed?','["C"]','Writer C calls permanence claims optimistic.'),
('q-paper-015','g-paper-3',2,'Which writer focuses on the environmental cost of plantation pulpwood?','["A"]','Writer A argues the accounting excludes the forest replaced.'),
('q-paper-016','g-paper-3',3,'Which writer worked from the guild records of the Fabriano workshops?','["D"]','Writer D reconstructs the workshop from guild records.'),
('q-paper-017','g-paper-3',4,'Which writer credits cheap paper with the spread of literacy in Europe?','["B"]','Writer B: literacy follows the paper price.'),
('q-paper-018','g-paper-3',5,'Which writer argues that de-acidification treatments damage bindings?','["C"]','Writer C calls the remedies worse than the problem.'),
('q-paper-019','g-paper-3',6,'Which writer published the first description of a continuous papermaking machine?','["E"]','Writer E cites the 1807 account.'),

-- Passage 3, group 4: arrange the passage
('q-paper-020','g-paper-4',1,'Position 1: a cheap writing material is presented to a Chinese emperor.','["C"]','Box C describes the presentation and the cost.'),
('q-paper-021','g-paper-4',2,'Position 2: rag supply becomes the limit on European paper production.','["A"]','Box A covers mills, parishes and rag legislation.'),
('q-paper-022','g-paper-4',3,'Position 3: machinery and wood pulp remove that limit.','["D"]','Box D covers the continuous web and the grinding stone.'),
('q-paper-023','g-paper-4',4,'Position 4: standards are written to define permanence.','["B"]','Box B covers the standards that followed the damage.'),

-- Passage 3, group 5: yes / no / not given
('q-paper-024','g-paper-5',1,'The author believes claims made for recycling are often overstated.','["YES"]','Paragraph I: an argument for being honest about what recycling is.'),
('q-paper-025','g-paper-5',2,'Chinese paper was more expensive than parchment.','["NOT_GIVEN"]','Parchment is never mentioned.'),
('q-paper-026','g-paper-5',3,'A recycled sheet can be produced without any new fibre entering the system.','["NO"]','Paragraph I: it always depends on new fibre somewhere.'),

-- Passage 3, group 6: sentence completion
('q-paper-027','g-paper-6',1,'Softwoods give long fibres and paper that is ________.','["strong"]','Paragraph F: long fibres and strong paper.'),
('q-paper-028','g-paper-6',2,'The kraft process cooks chips at about ________ degrees Celsius.','["170"]','Paragraph G gives about 170 degrees.'),
('q-paper-029','g-paper-6',3,'Wood pulp carries ________, which yellows and embrittles in light.','["lignin"]','Paragraph H names lignin.'),
('q-paper-030','g-paper-6',4,'Nineteenth-century sizing used alum and ________.','["rosin"]','Paragraph H: alum and rosin sizing.'),
('q-paper-031','g-paper-6',5,'Fibre can be recycled about five to seven times before it is fit only for ________.','["board"]','Paragraph I states this.'),
('q-paper-032','g-paper-6',6,'Acid-free paper must measure ________ or above when its pH is tested cold.','["7.0","7"]','Paragraph J defines acid-free as pH 7.0 or above.'),

-- Passage 3, group 7: matching information
('q-paper-033','g-paper-7',1,'A description of how growing conditions change fibre yield.','["F"]','Paragraph F ties yield to season, rainfall and spacing.'),
('q-paper-034','g-paper-7',2,'An explanation of why nineteenth-century books become brittle.','["H"]','Paragraph H covers lignin and acid hydrolysis.'),
('q-paper-035','g-paper-7',3,'Details of the temperatures used in pulping and bleaching.','["G"]','Paragraph G gives 170, 100 and 60 to 90 degrees.'),
('q-paper-036','g-paper-7',4,'A definition expressed as a measured pH value.','["J"]','Paragraph J defines acid-free by pH.'),

-- Passage 3, group 8: yes / no / not given
('q-paper-037','g-paper-8',1,'Recycled content is governed by a single international rule.','["NO"]','Paragraph J: governed by no single international rule.'),
('q-paper-038','g-paper-8',2,'Demand for certified recycled office paper has risen since the 1990s.','["YES"]','Paragraph J states this.'),
('q-paper-039','g-paper-8',3,'Digital publishing has reduced total paper consumption.','["NOT_GIVEN"]','Digital publishing is never mentioned.'),
('q-paper-040','g-paper-8',4,'A rag-paper book from 1550 is generally better preserved than one from 1890.','["YES"]','Paragraph H makes exactly this comparison.')

) AS v(id, group_id, position, prompt, correct_answers, explanation)
JOIN reading_question_groups g ON g.id = v.group_id
JOIN reading_passages p ON p.id = g.passage_id;

-- ---------------------------------------------------------------------------
-- Shared answer pools
-- ---------------------------------------------------------------------------
--
-- Applied per type rather than repeated on every row, so what True means is
-- decided in one place. Sentence completion is absent on purpose: those gaps
-- are typed, and offering options would turn the task into a different one.
--
-- Option ids, not labels, are what correct_answers holds and what the client
-- submits; the text is only what the learner reads.

UPDATE questions SET options = '[
    {"id":"TRUE","text":"True"},
    {"id":"FALSE","text":"False"}]'::jsonb
WHERE type_id = 'reading-true-false' AND passage_id IS NOT NULL;

UPDATE questions SET options = '[
    {"id":"YES","text":"Yes"},
    {"id":"NO","text":"No"},
    {"id":"NOT_GIVEN","text":"Not Given"}]'::jsonb
WHERE type_id = 'reading-yes-no-not-given' AND passage_id IS NOT NULL;

UPDATE questions SET options = '[
    {"id":"A","text":"Writer A"},
    {"id":"B","text":"Writer B"},
    {"id":"C","text":"Writer C"},
    {"id":"D","text":"Writer D"},
    {"id":"E","text":"Writer E"}]'::jsonb
WHERE type_id = 'reading-find-the-writer' AND passage_id IS NOT NULL;

UPDATE questions SET options = '[
    {"id":"A","text":"Paragraph A"},
    {"id":"B","text":"Paragraph B"},
    {"id":"C","text":"Paragraph C"},
    {"id":"D","text":"Paragraph D"}]'::jsonb
WHERE type_id = 'reading-arrange-passage' AND passage_id IS NOT NULL;

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
WHERE type_id = 'reading-matching-information' AND passage_id IS NOT NULL;

-- ---------------------------------------------------------------------------
-- Guard
-- ---------------------------------------------------------------------------
--
-- A generated mock needs three passages carrying all six types. If a later edit
-- to this file breaks that, the migration should fail here rather than leave a
-- database where starting a reading mock returns "not enough passages".

DO $guard$
DECLARE
    eligible INT;
BEGIN
    SELECT count(*) INTO eligible
    FROM reading_passages p
    WHERE p.is_published
      AND (SELECT count(DISTINCT g.type_id)
           FROM reading_question_groups g
           WHERE g.passage_id = p.id
             AND EXISTS (SELECT 1 FROM questions q WHERE q.group_id = g.id AND q.is_published)) = 6;

    IF eligible < 3 THEN
        RAISE EXCEPTION 'reading seed: % passages carry all six task types, need at least 3', eligible;
    END IF;
END
$guard$;
