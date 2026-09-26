"""Checks for a new content batch: no repeats of anything already in the bank,
and no wording that marks a text as machine-written or non-British.

Everything already in the bank is in scripts/bank_snapshot.json (written by
scripts/export_bank_snapshot.py from a fully migrated database). Use:

    from scripts.bank_check import Batch
    batch = Batch()
    batch.question('IELTS', 'ielts-speaking-part1', title, text)       # each new item
    batch.passage(title, text)                                          # reading/listening texts
    batch.style(text, where='ielts-b4-spk1-001')                        # every learner-facing text
    batch.assert_clean()                                                # fails with every problem found

A repeat is either the same or a near-same title within the same task type, or
text whose distinctive words overlap an existing item's heavily, or any shared
run of eight words. Pass allowed={...existing titles or ids} for a pair a
person has reviewed and judged different.
"""
import json
import re
from pathlib import Path

SNAPSHOT = json.loads(Path(__file__).with_name('bank_snapshot.json').read_text(encoding='utf-8'))

# Words that carry no topic on their own.
STOP = set("""
a about above after again against all also am an and any are as at be because been before being below between
both but by can could did do does doing down during each few for from further had has have having he her here
hers him his how i if in into is it its itself just me more most my no nor not now of off on once only or other
our out over own same she should so some such than that the their them then there these they this those through
to too under until up very was we were what when where which while who whom why will with would you your yours
people person many much often usually some think believe others agree disagree discuss views opinion give reasons
answer include relevant examples knowledge experience write least words suggested time minutes task describe
talk say explain should also well like one two three four five first second last new old make made take way ways
thing things lot get got really quite rather very good bad better best more less most least year years day days
""".split())

AMERICAN = {
    'color': 'colour', 'colors': 'colours', 'favorite': 'favourite', 'favorites': 'favourites', 'center': 'centre',
    'centers': 'centres', 'theater': 'theatre', 'theaters': 'theatres', 'meter': 'metre', 'meters': 'metres',
    'liter': 'litre', 'liters': 'litres', 'organize': 'organise', 'organized': 'organised', 'organization': 'organisation',
    'organizations': 'organisations', 'realize': 'realise', 'realized': 'realised', 'recognize': 'recognise',
    'recognized': 'recognised', 'analyze': 'analyse', 'analyzed': 'analysed', 'behavior': 'behaviour',
    'behaviors': 'behaviours', 'neighbor': 'neighbour', 'neighbors': 'neighbours', 'neighborhood': 'neighbourhood',
    'neighborhoods': 'neighbourhoods', 'labor': 'labour', 'honor': 'honour', 'humor': 'humour', 'flavor': 'flavour',
    'harbor': 'harbour', 'rumor': 'rumour', 'traveling': 'travelling', 'traveled': 'travelled', 'traveler': 'traveller',
    'travelers': 'travellers', 'canceled': 'cancelled', 'program': 'programme', 'programs': 'programmes',
    'defense': 'defence', 'offense': 'offence', 'license': 'licence', 'practicing': 'practising', 'catalog': 'catalogue',
    'dialog': 'dialogue', 'gray': 'grey', 'mom': 'mum', 'vacation': 'holiday', 'apartment': 'flat', 'apartments': 'flats',
    'gasoline': 'petrol', 'sidewalk': 'pavement', 'soccer': 'football', 'fall semester': 'autumn term',
    'jewelry': 'jewellery', 'aluminum': 'aluminium', 'fulfill': 'fulfil', 'enrollment': 'enrolment', 'modeling': 'modelling',
    'specialize': 'specialise', 'specialized': 'specialised', 'emphasize': 'emphasise', 'minimize': 'minimise',
    'maximize': 'maximise', 'prioritize': 'prioritise', 'summarize': 'summarise', 'apologize': 'apologise',
    'criticize': 'criticise', 'utilize': 'use', 'utilization': 'use', 'fiber': 'fibre', 'tire': 'tyre', 'tires': 'tyres',
}
# Words and phrases that mark a text as machine-written. IELTS and PTE texts are
# plain and specific; none of these appear in them.
AI_TELLS = [
    'delve', 'tapestry', 'testament to', 'fast-paced world', "in today's world", 'in the modern era', 'ever-evolving',
    'ever-changing', 'plethora', 'myriad', 'bustling', 'vibrant', 'navigate the complexities', 'navigating the',
    'it is important to note', "it's important to note", 'it is worth noting', 'embark', 'realm', 'unlock', 'unleash',
    'game-changer', 'game changer', 'seamless', 'seamlessly', 'holistic', 'synergy', 'paradigm', 'cutting-edge',
    'pivotal', 'crucial role', 'plays a vital role', 'plays a crucial role', 'foster', 'fostering', 'harness',
    'underscore', 'underscores', 'meticulous', 'intricate', 'nuanced', 'landscape of', 'digital age', 'showcase',
    'elevate', 'empower', 'robust', 'leverage', 'commendable', 'in conclusion, it is clear', 'a rich history',
    'rich cultural heritage', 'stands as', 'nestled', 'boasts', 'hidden gem', 'journey of', 'beacon', 'resonate',
]
# Invented place names of the kind generated text favours ("Oakhaven",
# "Silverdale", "Crestmont"). Real, ordinary places or plain descriptions
# ("a town in the north of England", "a coastal city") read as genuine.
FAKE_PLACE = re.compile(r"\b(?:Oak|Silver|Crest|Stone|Green|King|Blythe|Bell|Eld|Port|Mill|River|West|East|North|South|"
                        r"Brook|Maple|Pine|Willow|Rose|Fair|Clear|Bright|Golden|Iron|Raven|Sun|Moon|Star|Val|Ash|Elm|"
                        r"Cedar|Birch|Hazel|Wren|Thorn|Lake|Spring|Summer|Winter|Autumn)"
                        r"(?:haven|vale|view|crest|mont|wood|brook|ridge|well|dale|field|ford|bridge|shire|stead|"
                        r"lake|view|mere|wick|holm|ton|oria|ia)\b")


def _norm(word):
    for suffix, repl in (('ies', 'y'), ('es', ''), ('s', '')):
        if word.endswith(suffix) and len(word) - len(suffix) >= 4:
            return word[: -len(suffix)] + repl
    return word


def words(text):
    return {_norm(w) for w in re.findall(r"[a-z]+", (text or '').lower()) if len(w) >= 4 and w not in STOP}


def shingles(text, n=8):
    toks = re.findall(r"[a-z0-9]+", (text or '').lower())
    return {' '.join(toks[i:i + n]) for i in range(len(toks) - n + 1)}


# Instructions every item of a type shares ("Give reasons for your answer...")
# would match on every item, so shared runs that appear in more than a few
# existing items are not evidence of a repeat.
_counts = {}
for _q in SNAPSHOT['questions']:
    for _s in shingles(_q['text']):
        _counts[_s] = _counts.get(_s, 0) + 1
BOILERPLATE = {s for s, n in _counts.items() if n > 3}


def _index(items):
    return [(it, words(it.get('title', '')), words(it.get('text', '')), shingles(it.get('text', '')) - BOILERPLATE)
            for it in items]


# Words most items of a task type share are its template ("You should say",
# "Give reasons for your answer"), not its topic, so they are left out when
# items of that type are compared.
_df = {}
_sizes = {}
for _q in SNAPSHOT['questions']:
    _sizes[_q['type_id']] = _sizes.get(_q['type_id'], 0) + 1
    for _w in words(_q['text']):
        _df[(_q['type_id'], _w)] = _df.get((_q['type_id'], _w), 0) + 1
TEMPLATE = {}
for (_t, _w), _n in _df.items():
    if _n > max(3, 0.15 * _sizes[_t]):
        TEMPLATE.setdefault(_t, set()).add(_w)

QUESTIONS = _index(SNAPSHOT['questions'])
TEXTS = _index(SNAPSHOT['reading_passages'] + SNAPSHOT['reorder_items'] + SNAPSHOT['listening_parts']
               + SNAPSHOT['speaking_sets'])


def _similar(a, b):
    """Share of the two texts' distinctive words that they have in common."""
    return len(a & b) / len(a | b) if a and b else 0.0


def _overlap(a, b):
    return len(a & b) / max(1, min(len(a), len(b))) if a and b else 0.0


class Batch:
    def __init__(self):
        self.problems = []
        self.new_questions = []
        self.new_texts = []

    def _compare(self, where, title, text, pool, same_type):
        tw, xw, sh = words(title), words(text), shingles(text) - BOILERPLATE
        for it, itw, ixw, ish in pool:
            if not same_type(it):
                continue
            template = TEMPLATE.get(it.get('type_id'), set())
            xw, ixw = xw - template, ixw - template
            label = f"{it.get('id')} '{it.get('title')}'"
            if it.get('id') in self._allowed or it.get('title') in self._allowed:
                continue
            if title and (title.strip().lower() == (it.get('title') or '').strip().lower()):
                self.problems.append(f'{where}: same title as {label}')
            elif len(tw) >= 1 and _overlap(tw, itw) >= 1.0 and len(tw & itw) >= 2:
                self.problems.append(f'{where}: title repeats {label}')
            if len(xw) >= 6 and _similar(xw, ixw) >= 0.5:
                self.problems.append(f'{where}: text overlaps {label} ({_similar(xw, ixw):.0%} of distinctive words)')
            elif sh & ish:
                self.problems.append(f'{where}: shares the words "{next(iter(sh & ish))}" with {label}')

    def question(self, exam, type_id, title, text, where=None, allowed=()):
        """A new question of a task type; compared with every question of the
        same exam and task type, old and new."""
        where = where or title
        self._allowed = set(allowed)
        same = lambda it: it.get('exam') == exam and it.get('type_id') == type_id
        self._compare(where, title, text, QUESTIONS, same)
        self._compare(where, title, text, self.new_questions, same)
        self.new_questions.append(({'id': where, 'exam': exam, 'type_id': type_id, 'title': title},
                                   words(title), words(text), shingles(text) - BOILERPLATE))
        self.style(f'{title} {text}', where)

    def passage(self, title, text, where=None, allowed=()):
        """A reading passage, reorder text, listening script or speaking set;
        compared with every such text, old and new."""
        where = where or title
        self._allowed = set(allowed)
        self._compare(where, title, text, TEXTS, lambda it: True)
        self._compare(where, title, text, self.new_texts, lambda it: True)
        self.new_texts.append(({'id': where, 'title': title}, words(title), words(text), shingles(text) - BOILERPLATE))
        self.style(f'{title} {text}', where)

    def style(self, text, where, allowed=()):
        low = (text or '').lower()
        for us, uk in AMERICAN.items():
            if us not in allowed and re.search(r'\b' + re.escape(us) + r'\b', low):
                self.problems.append(f"{where}: American '{us}' (use '{uk}')")
        for tell in AI_TELLS:
            if tell not in allowed and re.search(r'\b' + re.escape(tell) + r'\b', low):
                self.problems.append(f"{where}: machine-sounding '{tell}'")
        if '—' in (text or ''):
            self.problems.append(f'{where}: em dash (use a comma, colon or full stop)')
        for m in FAKE_PLACE.finditer(text or ''):
            if m.group(0) not in allowed:
                self.problems.append(f"{where}: invented-sounding place name '{m.group(0)}'")

    def assert_clean(self):
        if self.problems:
            raise SystemExit('Bank check failed:\n  ' + '\n  '.join(sorted(set(self.problems))))
