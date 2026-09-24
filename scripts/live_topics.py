"""Topics already live, and a check that new content does not repeat them.

LIVE_READING_TITLES is every IELTS and PTE reading passage title on the live
site on 2026-09-25 (GET /api/v1/reading/passages for both exams, paged).
Refresh it from the live API before each new content batch.

topic_clashes() flags a new title or heading that shares a distinctive word
with a live title (for example "Understanding Urban Heat Islands" and "Cooling
the urban heat island"). A flag is not proof of a repeat ("Ancient Glassmaking"
and "The Pen and the Keyboard" never clash, but "Living Light" and "Light
pollution" do and are different topics), so each generator lists the pairs a
person has reviewed in an ALLOWED set and fails on any other.
"""
import re

LIVE_READING_TITLES = [
    "A Library of Things", "A costly mix-up of units", "A gentle push: the rise of nudge policy",
    "A pinch of history: the story of salt", "Algorithms that hire", "An Introduction to Film Sound",
    "Around Millbrook: Community Noticeboard", "Ballast water and species transfer", "Beavers as engineers",
    "Bike-sharing schemes", "Bringing back the red kite", "Building the Panama Canal", "Building with Grass",
    "Building with bamboo", "Buildings that float on earthquakes", "Can coral reefs be rebuilt?",
    "Can machines be creative?", "Can you see the Great Wall from space?", "Chimpanzee culture",
    "Community radio in Nepal", "Contrails and climate", "Cooling the urban heat island", "Credit unions",
    "Crop-growing skyscrapers", "Crowdfunding", "Cursive handwriting in schools", "Daylight saving time",
    "Do carbon offsets work?", "Drinking the Sea", "Drinking water from the sea", "Electric cars in Nepal",
    "Eradicating smallpox", "Esperanto", "Fairtrade", "Farming Upwards", "Farming ants", "Fermented foods",
    "Forests of the Tide", "Foxes in the city", "From coins to paper: the invention of banknotes",
    "Great Migrations", "Growing up with two languages",
    "Harlow Logistics Staff Handbook: Annual Leave and Absence", "Harvesting fog", "Heat pumps",
    "Himalayan gold", "How bats find their way", "How birds find north", "How lightning works",
    "How tall is Everest?", "How the Mona Lisa became famous", "Hydropower in Nepal",
    "Is hosting the Olympics worth it?", "Keepers of the Light", "Keeping the sea out of Venice",
    "Land from the sea", "Languages without left and right", "Later Starts: Teenagers and the School Day",
    "Life at the top: how Himalayan peoples adapted to altitude", "Lifeboats on the Titanic",
    "Light pollution", "Lighting the way: a short history of lighthouses", "Listening Beneath the Waves",
    "Lumbini", "Mangrove forests", "Mapping the ocean floor", "Measuring what matters", "Momos",
    "Money sent home", "Museums in the digital age", "Music and memory", "Navigating without instruments",
    "Nepal's community forests", "Nepal's flag", "Nepal's tigers", "Nepal's rising glacial lakes",
    "Neuroaesthetics", "Newar craftsmanship", "Noise on hospital wards", "Octopus intelligence",
    "Origins of the marathon", "Painting steel bridges", "Paragliding over Pokhara",
    "Pashmina from the Himalaya", "Pidgins and creoles", "Plant-based meat", "Power from the tides",
    "Preface to 'How the other half thinks: Adventures in mathematical reasoning'", "Raising the Mary Rose",
    "Rammed earth", "Reducing the Effects of Climate Change", "Remembering the ends of a list",
    "Research using twins", "Rhinos of Chitwan", "Rice and fish together",
    "Riverside Hotel Group: A Guide for New Front-of-House Staff", "Roundabouts and road safety",
    "Rubber leaves the Amazon", "Salt marshes", "Saving the Eiffel Tower", "Sea otters and kelp",
    "Sign languages", "Sleep and the making of memories", "Sleeping with half a brain",
    "Snow leopards in Nepal", "Spacing out study", "Supermarket loyalty cards", "Tardigrades",
    "Teaching children to read", "The Antikythera mechanism", "The Dutch tulip craze", "The Falkirk Wheel",
    "The Fresnel lens", "The Great Stink and the sewers of London", "The Pen and the Keyboard",
    "The QWERTY keyboard", "The Rosetta Stone", "The Science of Waiting", "The birth of time zones",
    "The bystander effect", "The caves of Mustang", "The dawn chorus", "The disappearing sparrow",
    "The discovery of penicillin", "The end of the dodo", "The first barcodes",
    "The first domesticated animal", "The first paper money", "The first stethoscope",
    "The ghost of the mountains", "The guthi system", "The language of honeybees", "The launch of Sputnik",
    "The long journey of chocolate", "The myth of multitasking", "The ocean garbage patch",
    "The origins of chess", "The placebo effect", "The potato's journey", "The power of the placebo",
    "The printing press", "The return of night trains", "The rise of the bicycle", "The sand shortage",
    "The second-hand clothing trade", "The shipping container", "The shrinking Dead Sea",
    "The spread of coffee", "The spread of our numerals", "The spread of tea", "The statues of Easter Island",
    "The story of braille", "The story of silk", "The tea gardens of eastern Nepal", "The uses of boredom",
    "The value of salt", "The wisdom and folly of crowds", "The year without a summer",
    "Tihar and the day of the dog", "Timber skyscrapers", "Trialling a four-day week",
    "Turning seawater into drinking water", "Two Wheels Forward", "Urban beekeeping", "Urban heat islands",
    "Votes for women in New Zealand", "Washing raw chicken", "Westfield Adult Education: Autumn Short Courses",
    "What destroyed the civilisation of Easter Island?", "When a language dies", "When languages die",
    "Why airlines overbook", "Why cats purr", "Why maps distort", "Why onions make us cry",
    "Why we buy extended warranties", "Why we have leap years", "Why we yawn", "Wolves in Yellowstone",
    "Working at night", "'This Marvellous Invention'",
]

# Words too general to mark a topic on their own.
GENERIC = set("""
about after against among ancient around based become before behind being below beneath between beyond
build building buildings city cities making made first great history story rise science sciences life lives
world power became become from with without into over under their there these those this that what when where which while
whose will would could should than then them they have your year years time times long short high back hidden
living understanding introduction guide modern future past small large little good better best more most less
many much some other every each work works working ways part parts place places people person days
practice research study studies engineering evolution network nature natural global applications uses
""".split())


def _words(text):
    words = set()
    for w in re.findall(r"[a-z]+", text.lower()):
        if len(w) < 4 or w in GENERIC:
            continue
        for suffix, repl in (('ies', 'y'), ('es', ''), ('s', '')):
            if w.endswith(suffix) and len(w) - len(suffix) >= 4:
                w = w[: -len(suffix)] + repl
                break
        words.add(w)
    return words


def topic_clashes(title, allowed=()):
    """Live titles sharing a distinctive word with `title`, minus reviewed pairs."""
    mine = _words(title)
    return [live for live in LIVE_READING_TITLES if mine & _words(live) and live not in allowed]


# ---------------------------------------------------------------------------
# Live questions (scripts/live_questions.json): every PTE and IELTS speaking,
# writing and listening question title on the live site on 2026-09-25, keyed
# "EXAM|Type name". Answer Short Questions are stored as their question text,
# because their titles are only category labels. Refresh with the live API
# (GET /api/v1/questions?exam=&skill=&limit=200&offset=, paged) before a batch.
#
# Learners see only their own exam's questions, so a new item must not repeat
# a live question of the same exam; the reading passages above count for both.
# ---------------------------------------------------------------------------
import json
from pathlib import Path

LIVE_QUESTIONS = json.loads((Path(__file__).with_name('live_questions.json')).read_text(encoding='utf-8'))


def question_clashes(exam, title, allowed=(), type_name=None):
    """Live question titles of the same exam (only the task `type_name`, when
    given), and live reading titles, that share a distinctive word with
    `title`, minus reviewed pairs. Checking a whole exam flags many unrelated
    titles; pass type_name to check for repeats within one task."""
    mine = _words(title)
    prefix = exam.upper() + '|' + (type_name or '')
    found = [t for key, titles in LIVE_QUESTIONS.items()
             if (key == prefix if type_name else key.startswith(prefix))
             for t in titles if mine & _words(t) and t not in allowed]
    return found + topic_clashes(title, allowed)


_QUESTION_WORDS = set('what which when where does call called word describes kind type name used usually often'.split())


def short_question_repeats(question, exam='PTE'):
    """Live Answer Short Questions asking much the same thing: most of the
    shorter question's distinctive words appear in the other."""
    mine = _words(question) - _QUESTION_WORDS
    repeats = []
    for live in LIVE_QUESTIONS.get(f'{exam.upper()}|Answer Short Questions', []):
        theirs = _words(live) - _QUESTION_WORDS
        if mine and theirs and len(mine & theirs) >= max(2, 0.75 * min(len(mine), len(theirs))):
            repeats.append(live)
    return repeats
