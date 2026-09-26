"""Passage 2: dating by tree rings."""

from scripts.ielts_b4_reading_a._build import opts, paras

PASSAGE = {
    "id_slug": "treerings",
    "title": "Reading the Rings",
    "subtitle": "How the growth of trees became a calendar for archaeologists and climate scientists",
    "topic": "Archaeology and dating methods",
    "paragraphs": paras(
        "Anyone who has counted the rings on a tree stump knows the basic principle: in climates with a "
        "distinct growing season, a tree adds one layer of wood each year. The wood formed in spring is usually "
        "pale and open, while that formed in late summer is darker and denser, so the boundary between one year "
        "and the next is easy to see. What is less obvious is that the rings differ in width. In a year with "
        "good rainfall and a mild summer, a tree may lay down a broad ring; in a dry or cold year the ring may "
        "be so narrow that it is barely visible. Because all the trees in a region experience the same weather, "
        "trees of one species growing in the same area tend to show the same pattern of wide and narrow rings. "
        "This simple observation is the foundation of dendrochronology, the science of dating by tree rings.",

        "The method was developed in the early twentieth century by Andrew Ellicott Douglass, an American "
        "astronomer who was interested not in trees but in the Sun. Douglass suspected that changes in solar "
        "activity affected the Earth's climate, and he hoped that the rings of old pine trees in Arizona would "
        "provide a record of rainfall stretching back further than written weather records. The link with the "
        "Sun that he was looking for was never firmly established, but he noticed that particular sequences of "
        "narrow rings appeared again and again in different trees. By matching these sequences, a process now "
        "called cross-dating, he could line up wood from living trees with timber from older, dead ones and "
        "extend the record backwards. In 1929 a charred beam from a ruined settlement allowed him to connect two "
        "sequences that had previously floated apart, and in doing so he fixed the dates of dozens of ancient "
        "buildings in the American south-west to the exact year.",

        "In Europe the most useful trees have proved to be oaks, which live for a long time and were widely used "
        "for buildings, ships and furniture. By joining the ring patterns of living trees to those of timbers "
        "from churches, barns and old houses, and then to trunks preserved in bogs and river gravels, "
        "researchers have built continuous chronologies. One oak and pine sequence from southern Germany now "
        "extends back more than 12,000 years, making it longer than any written calendar. Building such a "
        "record is slow work. A single sample may contain only 80 or 100 rings, and it can be dated only if its "
        "pattern overlaps securely with part of the master sequence. Samples with fewer than about 50 rings are "
        "often impossible to place, because a short pattern may match several points in the chronology equally "
        "well.",

        "For archaeologists and historians, the results can be remarkably precise. If the outermost ring under "
        "the bark survives, a dendrochronologist can say not only in which year a tree was felled but often in "
        "which season, since a partly formed ring shows that growth had already begun. Timber was usually used "
        "soon after felling, so the date of a roof or a wooden floor can be fixed closely. Where the outer rings "
        "have been trimmed away, the specialist must estimate how many are missing, using the typical width of "
        "the sapwood, the pale outer band of the trunk. The method has been applied to objects as well as "
        "buildings. The oak panels on which early Flemish artists painted have been dated in this way, and in "
        "some cases the results have shown that a picture could not have been produced as early as had been "
        "believed.",

        "Tree rings are also a source of information about past climates. In places where tree growth is "
        "limited mainly by one factor, such as summer temperature near the northern edge of the forests or "
        "rainfall in semi-arid regions, the width and density of the rings can be converted into estimates of "
        "that factor for each year. Such reconstructions have identified droughts lasting decades in North "
        "America, and cold summers following large volcanic eruptions, when ash and gases in the upper "
        "atmosphere reduced sunlight. Tree rings have one further use that few people outside science are "
        "aware of. They provided the samples of known age with which scientists corrected, or calibrated, the "
        "radiocarbon dating method, which had been found to give results that were wrong by centuries for some "
        "periods.",

        "None of this means that the method is without problems. Some trees occasionally fail to produce a ring "
        "in a very bad year, and others produce two in a year with an unusual break in the growing season, so "
        "single trees cannot be trusted; only by comparing many samples can such errors be identified. More "
        "seriously, in some northern forests the close relationship between temperature and ring width appears "
        "to have weakened since the 1960s, a puzzle which researchers call the 'divergence problem'. If trees "
        "now respond to warmth differently from the way they did in the past, estimates of earlier temperatures "
        "based on them may be less reliable than was assumed. In my view this is a reason for caution rather "
        "than despair. The dates themselves, which depend on matching patterns rather than on interpreting "
        "them, are on much firmer ground than the climate estimates, and remain among the most secure that "
        "archaeology possesses.",
    ),
    "mcq_single": [
        {
            "prompt": "Why do trees of the same species in one area show similar ring patterns?",
            "options": opts(
                "They are usually of the same age.",
                "They were planted at the same time.",
                "They experience the same weather.",
                "They grow at a fixed rate each year.",
            ),
            "key": "C",
            "explanation": "Paragraph A says that because all the trees in a region experience the same weather, they tend to show the same pattern of wide and narrow rings.",
        },
        {
            "prompt": "What was Douglass's main interest when he began studying tree rings?",
            "options": opts(
                "whether solar activity affected the Earth's climate",
                "how long pine trees in Arizona could live",
                "how quickly desert forests were disappearing",
                "the age of ancient settlements in the south-west",
            ),
            "key": "A",
            "explanation": "Paragraph B says Douglass was interested not in trees but in the Sun, and suspected that changes in solar activity affected the climate.",
        },
        {
            "prompt": "Why are samples with fewer than about 50 rings often impossible to date?",
            "options": opts(
                "The wood has usually decayed too far.",
                "They normally come from very young trees.",
                "They do not include any sapwood.",
                "Their pattern may fit several places in the chronology.",
            ),
            "key": "D",
            "explanation": "Paragraph C explains that a short pattern may match several points in the chronology equally well.",
        },
        {
            "prompt": "What does a partly formed outer ring tell a dendrochronologist?",
            "options": opts(
                "that the timber had been used before",
                "the season in which the tree was cut down",
                "that the sapwood has been trimmed away",
                "the age at which the tree stopped growing",
            ),
            "key": "B",
            "explanation": "Paragraph D says a specialist can often tell in which season a tree was felled, since a partly formed ring shows that growth had already begun.",
        },
        {
            "prompt": "What is the 'divergence problem'?",
            "options": opts(
                "the failure of some trees to form a ring in a bad year",
                "the difference between the ring patterns of oak and pine",
                "a weaker link between temperature and ring width in some forests",
                "disagreement between tree-ring dates and radiocarbon dates",
            ),
            "key": "C",
            "explanation": "Paragraph F says that in some northern forests the relationship between temperature and ring width appears to have weakened, a puzzle called the 'divergence problem'.",
        },
    ],
    "mcq_multiple": [
        {
            "prompt": "Which TWO conditions does the writer link with a broad ring?",
            "options": opts(
                "a cold winter",
                "good rainfall",
                "a long period of drought",
                "strong winds",
                "a mild summer",
            ),
            "keys": ["B", "E"],
            "explanation": "Paragraph A says that in a year with good rainfall (B) and a mild summer (E), a tree may lay down a broad ring.",
        },
        {
            "prompt": "Which TWO things did Douglass achieve through cross-dating?",
            "options": opts(
                "He proved that the Sun controls rainfall.",
                "He linked the wood of living trees with older timber.",
                "He found the oldest living tree in America.",
                "He fixed the exact dates of many ancient buildings.",
                "He built the first chronology of European oaks.",
            ),
            "keys": ["B", "D"],
            "explanation": "Paragraph B says cross-dating let him line up wood from living trees with timber from older, dead ones (B), and that he fixed the dates of dozens of ancient buildings to the exact year (D).",
        },
        {
            "prompt": "Which TWO sources of wood have been used to build the European chronologies?",
            "options": opts(
                "timbers from churches and barns",
                "wooden tools from ancient graves",
                "masts of ships kept in museums",
                "trunks preserved in bogs",
                "wood taken from coal mines",
            ),
            "keys": ["A", "D"],
            "explanation": "Paragraph C lists timbers from churches, barns and old houses (A) and trunks preserved in bogs and river gravels (D). Oak was used for ships, but ships are not given as a source of samples.",
        },
        {
            "prompt": "Which TWO kinds of past event have tree-ring studies identified?",
            "options": opts(
                "changes in sea level",
                "the spread of farming",
                "droughts lasting decades",
                "an increase in forest fires",
                "cold summers after volcanic eruptions",
            ),
            "keys": ["C", "E"],
            "explanation": "Paragraph E says reconstructions have identified droughts lasting decades in North America (C) and cold summers following large volcanic eruptions (E).",
        },
        {
            "prompt": "Which TWO difficulties with tree-ring evidence does the writer mention?",
            "options": opts(
                "A tree may fail to form a ring in a very bad year.",
                "Suitable oak timber is becoming hard to find.",
                "Laboratories often disagree about dates.",
                "Sampling has become too expensive.",
                "The link between warmth and growth has changed in some places.",
            ),
            "keys": ["A", "E"],
            "explanation": "Paragraph F says some trees fail to produce a ring in a very bad year (A) and that in some forests the relationship between temperature and ring width has weakened (E).",
        },
    ],
    "tfng": [
        {
            "statement": "Wood formed late in the summer is paler than wood formed in spring.",
            "key": "FALSE",
            "explanation": "Paragraph A says spring wood is pale and open, while late-summer wood is darker and denser.",
        },
        {
            "statement": "Douglass trained as a botanist before turning to astronomy.",
            "key": "NOT_GIVEN",
            "explanation": "Paragraph B describes Douglass as an American astronomer but says nothing about any training in botany.",
        },
        {
            "statement": "The German oak and pine sequence covers a longer period than any written calendar.",
            "key": "TRUE",
            "explanation": "Paragraph C says it extends back more than 12,000 years, making it longer than any written calendar.",
        },
        {
            "statement": "Timber was normally left to dry for many years before it was used in buildings.",
            "key": "FALSE",
            "explanation": "Paragraph D says timber was usually used soon after felling.",
        },
        {
            "statement": "Tree rings have been used to correct errors in radiocarbon dating.",
            "key": "TRUE",
            "explanation": "Paragraph E says tree rings provided the samples of known age with which scientists corrected, or calibrated, the radiocarbon method.",
        },
    ],
    "ynng": [
        {
            "statement": "Building a long tree-ring chronology takes a great deal of time.",
            "key": "YES",
            "explanation": "Paragraph C says that building such a record is slow work.",
        },
        {
            "statement": "Tree-ring dates for buildings are usually too approximate to help historians.",
            "key": "NO",
            "explanation": "Paragraph D says that for archaeologists and historians the results can be remarkably precise.",
        },
        {
            "statement": "All panel paintings in museums should be dated by their tree rings.",
            "key": "NOT_GIVEN",
            "explanation": "Paragraph D reports that some Flemish panels have been dated, but the writer gives no view on whether all paintings should be.",
        },
        {
            "statement": "The divergence problem is a good reason to give up using tree rings to study past climates.",
            "key": "NO",
            "explanation": "Paragraph F says that in the writer's view it is a reason for caution rather than despair.",
        },
        {
            "statement": "Dates taken from tree rings are more reliable than climate estimates based on them.",
            "key": "YES",
            "explanation": "Paragraph F says the dates are on much firmer ground than the climate estimates.",
        },
    ],
    "matching": [
        {
            "prompt": "an explanation of why the boundary between years can be seen",
            "key": "A",
            "evidence": "while that formed in late summer is darker and denser, so the boundary between one year and the next is easy to see",
            "explanation": "Paragraph A contrasts pale spring wood with darker late-summer wood.",
        },
        {
            "prompt": "a reference to a find that joined two separate sequences",
            "key": "B",
            "evidence": "a charred beam from a ruined settlement allowed him to connect two sequences that had previously floated apart",
            "explanation": "Paragraph B describes the beam found in 1929.",
        },
        {
            "prompt": "a description of the part of a trunk used to estimate missing rings",
            "key": "D",
            "evidence": "the sapwood, the pale outer band of the trunk",
            "explanation": "Paragraph D explains how the width of the sapwood is used when outer rings are missing.",
        },
        {
            "prompt": "a mention of places where growth depends chiefly on a single condition",
            "key": "E",
            "evidence": "In places where tree growth is limited mainly by one factor",
            "explanation": "Paragraph E refers to places where growth is limited mainly by summer temperature or by rainfall.",
        },
        {
            "prompt": "an example of how a single tree could give a misleading result",
            "key": "F",
            "evidence": "others produce two in a year with an unusual break in the growing season",
            "explanation": "Paragraph F explains that some trees produce two rings in one year, so single trees cannot be trusted.",
        },
    ],
    "completion": [
        {
            "prompt": "The science of dating by tree rings is known as ________.",
            "answer": "dendrochronology",
            "explanation": "Paragraph A defines dendrochronology as the science of dating by tree rings.",
        },
        {
            "prompt": "Matching sequences of rings between different pieces of wood is called ________.",
            "answer": "cross-dating",
            "explanation": "Paragraph B says matching these sequences is a process now called cross-dating.",
        },
        {
            "prompt": "A sample can be dated only if its pattern overlaps with part of the ________ sequence.",
            "answer": "master",
            "explanation": "Paragraph C says a sample can be dated only if its pattern overlaps securely with part of the master sequence.",
        },
        {
            "prompt": "Missing outer rings are estimated from the typical width of the ________.",
            "answer": "sapwood",
            "explanation": "Paragraph D says the specialist uses the typical width of the sapwood.",
        },
        {
            "prompt": "The radiocarbon dating method was corrected, or ________, using tree-ring samples of known age.",
            "answer": "calibrated",
            "explanation": "Paragraph E says scientists corrected, or calibrated, the radiocarbon dating method.",
        },
    ],
}
