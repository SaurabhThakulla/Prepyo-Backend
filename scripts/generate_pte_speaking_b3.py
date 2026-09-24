"""Generate PTE Speaking Bank Batch 3 (Migration 000087).

This script generates 350 original PTE speaking practice questions:
  - 50 Read Aloud (RA) (context_passage <= 60 words, prep 35s, time 40s)
  - 50 Repeat Sentence (RS) (audio_transcript 8-15 words, time 15s)
  - 50 Describe Image (DI) (10 bar, 10 line, 10 pie, 10 table, 10 process, inline SVG)
  - 50 Re-tell Lecture (RL) (audio_transcript 150-250 words, prep 10s, time 40s)
  - 50 Answer Short Questions (ASQ) (model_answer 1-3 words, time 10s)
  - 50 Summarise Group Discussion (SGD) (3 labelled speakers, prep 10s, time 120s, Rule 8)
  - 50 Respond to Situation (RTS) (context_passage == audio_transcript, prep 20s, time 40s)

Self-validation enforces:
  - Exactly 50 questions per task type (350 total).
  - All IDs are unique and prefixed with pte-b3-<task>-<nnn>.
  - Constraints on word counts, speaker turns, SVG data URIs, and fields.
  - No item repeats a live question of the same task or a live reading topic
    (question_clashes in scripts/live_topics.py) unless the pair is reviewed;
    no Answer Short Question repeats a live one (short_question_repeats), and
    no ASQ title gives away its answer.
  - type_name matches the live bank ('Answer Short Questions',
    'Summarise Group Discussion') so the practice lists show the items.
  - Generates migrations/000087_pte_speaking_bank_3.up.sql (INSERT ... ON CONFLICT (id) DO NOTHING)
    and migrations/000087_pte_speaking_bank_3.down.sql (SELECT 1;).
  - Uses Unix line endings (newline='\n').
"""

import json
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from scripts.live_topics import question_clashes, short_question_repeats
from scripts.pte_speaking_b3.ra import RA_ITEMS
from scripts.pte_speaking_b3.rs import RS_ITEMS
from scripts.pte_speaking_b3.di import get_di_items
from scripts.pte_speaking_b3.rl import RL_ITEMS
from scripts.pte_speaking_b3.asq import ASQ_ITEMS
from scripts.pte_speaking_b3.sgd import SGD_ITEMS
from scripts.pte_speaking_b3.rts import RTS_ITEMS

VERSION = 'pte-2026-01'
UP_MIGRATION = ROOT / 'migrations' / '000087_pte_speaking_bank_3.up.sql'
DOWN_MIGRATION = ROOT / 'migrations' / '000087_pte_speaking_bank_3.down.sql'

# Live titles that share a word with a new item, each reviewed and judged a
# different topic. Checked against live questions of the same task and live
# reading titles (question_clashes in scripts/live_topics.py). The closest
# pairs, kept on purpose: Bellview (household use by season) and "Electricity
# demand over a typical day" (hourly); Eldridge (home water use by activity)
# and "Daily water use per person in four cities"; the water-quality table and
# "How drinking water is treated"; metro ridership over five years and a
# late-night transit survey; the swimming-pool table and "Four hotels".
ALLOWED_TOPIC_OVERLAPS = {
    # Read Aloud
    'Cognitive Reserve and Lifelong Learning': {'Learning in the home language'},
    'Metamorphic Rock Foliation': {'Life in rock pools'},
    # Describe Image
    'Oakhaven Refuse Sorting Rates across Municipalities (%)': {'Exam pass rates in three subjects'},
    'Blythewood Solar Energy Generation (Megawatt Hours)': {'Average weekly hours of paid work by year of study', 'The share of electricity from solar power in five countries'},
    'Kingswell Clinic Consultation Delay Times (Minutes)': {'Causes of delays on a railway line'},
    'Greenvale Residential Thermal Power Sources (%)': {"The sources of a university's income"},
    'Port Haven Container Crane Productivity (Moves/Hour)': {'Average weekly hours of paid work by year of study', 'The shipping container'},
    'Valoria Metro Transit Ridership (Millions)': {'Survey on late-night transit ridership'},
    'Stonebridge Wind Farm Capacity Factor (%)': {'New hydropower capacity added each year'},
    'Bellview District Domestic Electricity Usage (kWh/Day)': {'Electricity demand over a typical day', 'The area planted with each crop in a district', 'The number of new small businesses registered in four districts', 'The share of electricity from solar power in five countries'},
    'Silverlake Municipal Recycling Rate Trajectory (%)': {'Exam pass rates in three subjects'},
    'Crestmont High School Graduation Percentage (%)': {'Cursive handwriting in schools', 'Later Starts: Teenagers and the School Day'},
    'Eldridge Municipal Domestic Water Consumption Share (%)': {'Ballast water and species transfer', 'Daily water use per person in four cities', 'Drinking water from the sea', 'How drinking water is treated', 'The share of electricity from solar power in five countries', 'The share of retail sales made online', 'Turning seawater into drinking water'},
    'Oakhaven Municipal Land Allocation Breakdown (%)': {'Land from the sea'},
    'Silverdale Mobile Computing Operating Platform Share (%)': {'Four mobile phones', 'The share of electricity from solar power in five countries', 'The share of retail sales made online'},
    'Kingswell District Generation Portfolio Breakdown (%)': {'The area planted with each crop in a district', 'The number of new small businesses registered in four districts'},
    'How time is spent at a youth summer camp': {'The year without a summer'},
    'Portview Container Export Destination Profile (%)': {'The shipping container'},
    'Forestry Timber Mechanical Density and Strength': {'Timber skyscrapers'},
    'Municipal Meteorological Sensor Annual Summary': {'Harlow Logistics Staff Handbook: Annual Leave and Absence'},
    'Hospital Urgent Care Admission and Treatment Metrics': {'Noise on hospital wards'},
    'Secondary School Extracurricular Participation Statistics': {'Cursive handwriting in schools', 'Later Starts: Teenagers and the School Day'},
    'Coastal Ferry Route Performance and Patronage': {'Changes in altitude along a trekking route'},
    'Municipal Potable Water Quality Laboratory Readings': {'Ballast water and species transfer', 'Daily water use per person in four cities', 'Drinking water from the sea', 'How drinking water is treated', 'Turning seawater into drinking water'},
    'Urban Rainwater Biofiltration Basin Sequence': {'Cooling the urban heat island', 'Urban beekeeping', 'Urban heat islands'},
    'Solar Photovoltaic Silicon Ingot Fabrication': {'The share of electricity from solar power in five countries'},
    # Re-tell Lecture
    'Why desert nights are cold': {'The return of night trains', 'Working at night'},
    'The discovery of X-rays': {'The discovery of penicillin'},
    "How birds' skeletons are built for flight": {'How birds find north'},
    'The invention of the steam engine': {"'This Marvellous Invention'", 'From coins to paper: the invention of banknotes'},
    'Robots that work in swarms': {'Locust swarms'},
    'Carbon nanotubes in aircraft materials': {'Capturing carbon', 'Do carbon offsets work?'},
    'The paints used in Ice Age cave art': {'The caves of Mustang'},
    'The invention of refrigeration': {"'This Marvellous Invention'", 'From coins to paper: the invention of banknotes'},
    'Factories that share their waste': {"Waste on the world's highest mountain"},
    'A sudden return to cold at the end of the Ice Age': {'The return of night trains'},
    'How chameleons change their skin': {'Reducing the Effects of Climate Change', 'Why some rivers change course'},
    'Why some people develop food allergies': {'Fermented foods', 'The food miles debate'},
    'Where domestic horses came from': {'Where money came from'},
    'Earthquakes caused by human activity': {'Buildings that float on earthquakes'},
    'Water wheels in the Roman world': {'Ballast water and species transfer', 'Drinking water from the sea', 'The Falkirk Wheel', 'Turning seawater into drinking water', 'Two Wheels Forward'},
    "The ocean's global conveyor belt": {'Mapping the ocean floor', 'The ocean garbage patch'},
    'The invention of spectacles': {"'This Marvellous Invention'", 'From coins to paper: the invention of banknotes'},
    'How the air bends sound': {'An Introduction to Film Sound'},
    'Einstein and the photoelectric effect': {'Reducing the Effects of Climate Change', 'The bystander effect', 'The placebo effect', 'The spacing effect'},
    # Summarise Group Discussion
    'Campus Cycle Service Pod': {'Plastic bottle ban on campus'},
    'A second-hand furniture exchange': {'The second-hand clothing trade'},
    'Textbook Exchange Peer System': {'The guthi system'},
    'Campus Evening Minivan Service': {'Plastic bottle ban on campus'},
    'Student Garden Plot Lottery': {'The tea gardens of eastern Nepal'},
    'Orientation Cohort Walking Groups': {'Group project roles', 'Riverside Hotel Group: A Guide for New Front-of-House Staff'},
    'Starting a campus choir': {'Plastic bottle ban on campus'},
    'Student Media Podcast Production': {'Social media for news'},
    'Moot Court Tournament Hosting': {'Is hosting the Olympics worth it?'},
    'Language Partner Intake Matching': {'The language of honeybees', 'When a language dies'},
    'A free printing allowance': {'The printing press'},
    'Open Source Indoor Navigation Mapping': {'Mapping the ocean floor'},
    'Campus Flora Inventory BioBlitz': {'Plastic bottle ban on campus'},
    'Student Gazette Digital Format': {'Museums in the digital age'},
    'Maker Space Print Job Scheduling': {'Can you see the Great Wall from space?'},
    'Observatory Meadow Telescope Night': {'The return of night trains', 'Working at night'},
    # Respond to a Situation
    'A broken radiator in your room': {'Complaining about a hotel room'},
    'Someone else is in your booked room': {'Complaining about a hotel room'},
    'Asking for a day off for a wedding': {'Asking a professor for a reference'},
    'Turning down a party invitation': {'Turning seawater into drinking water'},
    'A work shift clashes with a class': {'Changing a class schedule'},
    'Asking a neighbour to take in a parcel': {'Asking a professor for a reference', 'Noisy neighbour'},
    'A bike locked to a handrail': {'Bike-sharing schemes'},
    'Asking to swap seats on a train': {'Asking a professor for a reference', 'The return of night trains'},
    'Late for a lab because of a train': {'The return of night trains'},
    'Cold air in a lecture hall': {'Borrowing lecture notes'},
    'Changing your open day hours': {'Changing a class schedule'},
    'The wrong size graduation gown': {'Wrong order at a café'},
    'Changing a restaurant booking': {'Changing a class schedule'},
    'A wallet left on the shuttle bus': {'Languages without left and right'},
    'A wrong charge on your meal plan': {'Wrong order at a café'},
    'Inviting a classmate to a study group': {'Group member not contributing', 'Presenting a complaint for your group', 'Riverside Hotel Group: A Guide for New Front-of-House Staff'},
    'A vending machine takes your money': {'Money sent home', 'The first paper money'},
    "A roommate's guest who stays for days": {"Roommate's messy kitchen"},
    'Asking which textbooks to buy': {'Asking a professor for a reference'},
    'Asking for more club funding': {'Asking a professor for a reference'},
    'Cycling past a group on a shared path': {'Group member not contributing', 'Presenting a complaint for your group', 'Riverside Hotel Group: A Guide for New Front-of-House Staff'},
    'Asking a professor about research work': {'Asking a professor for a reference'},
    'Asking a guide to slow down': {'Asking a professor for a reference'},
    'No internet in a basement room': {'Complaining about a hotel room'},
    'Noise in the silent study area': {'Noise on hospital wards'},
    'Asking to skip a required course': {'Asking a professor for a reference'},
    'The wrong textbook delivered': {'Wrong order at a café'},
}


def sql_lit(val):
    if val is None:
        return 'NULL'
    if isinstance(val, bool):
        return 'TRUE' if val else 'FALSE'
    if isinstance(val, (int, float)):
        return str(val)
    if isinstance(val, (dict, list)):
        return "'" + json.dumps(val, ensure_ascii=False).replace("'", "''") + "'::jsonb"
    return "'" + str(val).replace("'", "''") + "'"


def sql_text_array(items):
    return 'ARRAY[' + ', '.join(sql_lit(x) for x in items) + ']'


def validate_and_compile():
    print("Validating PTE Speaking Batch 3 items...")
    di_items = get_di_items()

    tasks = [
        ('RA', RA_ITEMS),
        ('RS', RS_ITEMS),
        ('DI', di_items),
        ('RL', RL_ITEMS),
        ('ASQ', ASQ_ITEMS),
        ('SGD', SGD_ITEMS),
        ('RTS', RTS_ITEMS),
    ]

    for code, items in tasks:
        print(f"  {code}: {len(items)} items")
        assert len(items) == 50, f"Expected 50 items for {code}, got {len(items)}"

    # Specific constraint checks
    for i, item in enumerate(RA_ITEMS):
        words = len(item['text'].split())
        assert words <= 60, f"RA {i+1} exceeds 60 words: {words}"

    for i, item in enumerate(RS_ITEMS):
        words = len(item['sentence'].split())
        assert 8 <= words <= 15, f"RS {i+1} word count out of range: {words}"

    for i, item in enumerate(di_items):
        assert item['image_url'].startswith('data:image/svg+xml,'), f"DI {i+1} invalid image_url"
        assert item['figure_data'] and item['model_answer'] and item['explanation']

    for i, item in enumerate(RL_ITEMS):
        words = len(item['audio_transcript'].split())
        assert 150 <= words <= 250, f"RL {i+1} word count out of range: {words}"
        assert item['model_answer'] and item['explanation']

    for i, item in enumerate(ASQ_ITEMS):
        words = len(item['model_answer'].split())
        assert 1 <= words <= 3, f"ASQ {i+1} model_answer out of range: {words}"
        assert item['audio_transcript'] and item['explanation']

    colon_pat = re.compile(r'\b[A-Za-z]+:')
    for i, item in enumerate(SGD_ITEMS):
        colons = colon_pat.findall(item['audio_transcript'])
        assert colons == ['Alex:', 'Blair:', 'Casey:', 'Blair:', 'Alex:'], f"SGD {i+1} unexpected colons: {colons}"
        assert item['model_answer'] and item['explanation']

    for i, item in enumerate(RTS_ITEMS):
        assert item['situation'] and item['model_answer'] and item['explanation']

    # Topic checks: no item may repeat a live question of the same task or a
    # live reading topic, unless the pair is reviewed in ALLOWED_TOPIC_OVERLAPS.
    print("Checking topic clashes against the live bank...")
    type_names = {'RA': 'Read Aloud', 'RS': 'Repeat Sentence', 'DI': 'Describe Image', 'RL': 'Re-tell Lecture',
                  'SGD': 'Summarise Group Discussion', 'RTS': 'Respond to a Situation'}
    titles = [item['title'] for code, items in tasks if code != 'ASQ' for item in items]
    assert len(titles) == len(set(titles)), "duplicate titles within the batch"
    total_allowed = 0
    for code, items in tasks:
        if code == 'ASQ':
            continue
        for i, item in enumerate(items):
            title = item['title']
            allowed = ALLOWED_TOPIC_OVERLAPS.get(title, set())
            total_allowed += bool(allowed)
            clashes = question_clashes('PTE', title, allowed=allowed, type_name=type_names[code])
            assert not clashes, f"Unreviewed topic clash in {code} #{i+1} '{title}': {clashes}"
    stale = set(ALLOWED_TOPIC_OVERLAPS) - set(titles)
    assert not stale, f"ALLOWED_TOPIC_OVERLAPS names titles not in the batch: {stale}"
    print(f"Topic clash checks passed! ({total_allowed} reviewed overlaps)")

    # Answer Short Questions: no repeat of a live question, no repeat within the
    # batch, and the title (shown in the practice list) never gives the answer.
    questions = [item['audio_transcript'] for item in ASQ_ITEMS]
    assert len(questions) == len(set(questions)), "duplicate ASQ questions"
    for i, item in enumerate(ASQ_ITEMS):
        repeats = short_question_repeats(item['audio_transcript'])
        assert not repeats, f"ASQ {i+1} repeats a live question: {repeats}"
        answer = re.sub(r'^(a|an|the|your) ', '', item['model_answer'].lower())
        assert answer not in item['title'].lower(), f"ASQ {i+1} title gives the answer"

    print("All task constraints passed! Compiling rows...")

    rows = []
    seen_ids = set()

    # 1. RA
    for i, item in enumerate(RA_ITEMS):
        qid = f"pte-b3-ra-{i+1:03d}"
        assert qid not in seen_ids
        seen_ids.add(qid)
        rows.append({
            'id': qid,
            'exam_version_id': VERSION,
            'exam': 'PTE',
            'skill': 'speaking',
            'type_id': 'read-aloud',
            'type_name': 'Read Aloud',
            'title': item['title'],
            'prompt': 'Look at the text below. In 35 seconds, you must read this text aloud as naturally and clearly as possible. You have 40 seconds to read aloud.',
            'context_passage': item['text'],
            'prep_time_seconds': 35,
            'time_limit_seconds': 40,
            'points': 15,
            'difficulty': 'medium',
            'explanation': item['explanation'],
            'tags': ['PTE', 'Original practice', 'Batch 3', 'Read Aloud'],
            'supported_exams': ['PTE'],
            'is_published': True,
        })

    # 2. RS
    for i, item in enumerate(RS_ITEMS):
        qid = f"pte-b3-rs-{i+1:03d}"
        assert qid not in seen_ids
        seen_ids.add(qid)
        rows.append({
            'id': qid,
            'exam_version_id': VERSION,
            'exam': 'PTE',
            'skill': 'speaking',
            'type_id': 'pte-repeat-sentence',
            'type_name': 'Repeat Sentence',
            'title': item['title'],
            'prompt': 'You will hear a sentence. Please repeat the sentence exactly as you hear it. You will hear the sentence only once.',
            'audio_transcript': item['sentence'],
            'model_answer': item['sentence'],
            'prep_time_seconds': 0,
            'time_limit_seconds': 15,
            'points': 10,
            'difficulty': 'medium',
            'explanation': item['explanation'],
            'tags': ['PTE', 'Original practice', 'Batch 3', 'Repeat Sentence'],
            'supported_exams': ['PTE'],
            'is_published': True,
        })

    # 3. DI
    for i, item in enumerate(di_items):
        qid = f"pte-b3-di-{i+1:03d}"
        assert qid not in seen_ids
        seen_ids.add(qid)
        rows.append({
            'id': qid,
            'exam_version_id': VERSION,
            'exam': 'PTE',
            'skill': 'speaking',
            'type_id': 'pte-describe-image',
            'type_name': 'Describe Image',
            'title': item['title'],
            'prompt': 'Look at the image below. In 25 seconds, please speak into the microphone and describe in detail what the image is showing. You will have 40 seconds to give your response.',
            'image_url': item['image_url'],
            'figure_data': item['figure_data'],
            'model_answer': item['model_answer'],
            'prep_time_seconds': 25,
            'time_limit_seconds': 40,
            'points': 15,
            'difficulty': 'medium',
            'explanation': item['explanation'],
            'tags': ['PTE', 'Original practice', 'Batch 3', 'Describe Image'],
            'supported_exams': ['PTE'],
            'is_published': True,
        })

    # 4. RL
    for i, item in enumerate(RL_ITEMS):
        qid = f"pte-b3-rl-{i+1:03d}"
        assert qid not in seen_ids
        seen_ids.add(qid)
        rows.append({
            'id': qid,
            'exam_version_id': VERSION,
            'exam': 'PTE',
            'skill': 'speaking',
            'type_id': 'pte-retell-lecture',
            'type_name': 'Re-tell Lecture',
            'title': item['title'],
            'prompt': 'Listen to the lecture. After 10 seconds of preparation, retell the main points in your own words in 40 seconds.',
            'audio_transcript': item['audio_transcript'],
            'model_answer': item['model_answer'],
            'prep_time_seconds': 10,
            'time_limit_seconds': 40,
            'points': 15,
            'difficulty': 'medium',
            'explanation': item['explanation'],
            'tags': ['PTE', 'Original practice', 'Batch 3', 'Re-tell Lecture'],
            'supported_exams': ['PTE'],
            'is_published': True,
        })

    # 5. ASQ
    for i, item in enumerate(ASQ_ITEMS):
        qid = f"pte-b3-asq-{i+1:03d}"
        assert qid not in seen_ids
        seen_ids.add(qid)
        rows.append({
            'id': qid,
            'exam_version_id': VERSION,
            'exam': 'PTE',
            'skill': 'speaking',
            'type_id': 'pte-answer-short-question',
            'type_name': 'Answer Short Questions',
            'title': item['title'],
            'prompt': 'You will hear a question. Please give a simple and short answer. Often just one or a few words is enough.',
            'audio_transcript': item['audio_transcript'],
            'model_answer': item['model_answer'],
            'prep_time_seconds': 0,
            'time_limit_seconds': 10,
            'points': 5,
            'difficulty': 'medium',
            'explanation': item['explanation'],
            'tags': ['PTE', 'Original practice', 'Batch 3', 'Answer Short Questions'],
            'supported_exams': ['PTE'],
            'is_published': True,
        })

    # 6. SGD
    for i, item in enumerate(SGD_ITEMS):
        qid = f"pte-b3-sgd-{i+1:03d}"
        assert qid not in seen_ids
        seen_ids.add(qid)
        rows.append({
            'id': qid,
            'exam_version_id': VERSION,
            'exam': 'PTE',
            'skill': 'speaking',
            'type_id': 'pte-summarise-group-discussion',
            'type_name': 'Summarise Group Discussion',
            'title': item['title'],
            'prompt': 'You will hear three people having a discussion. When you hear the beep, summarise the whole discussion. You will have 10 seconds to prepare and 2 minutes to give your response. (Practice note: this discussion is shorter than in the real test.)',
            'audio_transcript': item['audio_transcript'],
            'model_answer': item['model_answer'],
            'prep_time_seconds': 10,
            'time_limit_seconds': 120,
            'points': 15,
            'difficulty': 'medium',
            'explanation': item['explanation'],
            'tags': ['PTE', 'Original practice', 'Batch 3', 'Summarise Group Discussion'],
            'supported_exams': ['PTE'],
            'is_published': True,
        })

    # 7. RTS
    for i, item in enumerate(RTS_ITEMS):
        qid = f"pte-b3-rts-{i+1:03d}"
        assert qid not in seen_ids
        seen_ids.add(qid)
        rows.append({
            'id': qid,
            'exam_version_id': VERSION,
            'exam': 'PTE',
            'skill': 'speaking',
            'type_id': 'pte-respond-to-situation',
            'type_name': 'Respond to a Situation',
            'title': item['title'],
            'prompt': item['situation'] + ' In 20 seconds, please prepare your response. You will have 40 seconds to give your response.',
            'context_passage': item['situation'],
            'audio_transcript': item['situation'],
            'model_answer': item['model_answer'],
            'prep_time_seconds': 20,
            'time_limit_seconds': 40,
            'points': 15,
            'difficulty': 'medium',
            'explanation': item['explanation'],
            'tags': ['PTE', 'Original practice', 'Batch 3', 'Respond to a Situation'],
            'supported_exams': ['PTE'],
            'is_published': True,
        })

    assert len(rows) == 350, f"Expected 350 rows, got {len(rows)}"
    assert len(seen_ids) == 350, f"Expected 350 unique IDs, got {len(seen_ids)}"

    print(f"Generated {len(rows)} question records successfully.")

    # All columns present across rows
    columns = [
        'id', 'exam_version_id', 'exam', 'skill', 'type_id', 'type_name',
        'title', 'prompt', 'context_passage', 'audio_transcript', 'image_url',
        'figure_data', 'model_answer', 'prep_time_seconds', 'time_limit_seconds',
        'points', 'difficulty', 'explanation', 'tags', 'supported_exams', 'is_published'
    ]

    def format_row(r):
        vals = []
        for col in columns:
            if col in ('tags', 'supported_exams'):
                vals.append(sql_text_array(r.get(col, [])))
            else:
                vals.append(sql_lit(r.get(col)))
        return f"    ({', '.join(vals)})"

    up_sql = [
        "-- 000087_pte_speaking_bank_3.up.sql",
        "-- Batch B: 350 original PTE Speaking practice questions (50 each for RA, RS, DI, RL, ASQ, SGD, RTS).",
        "-- Synthetic playback; inline SVG charts; deterministic IDs pte-b3-<task>-<nnn>.",
        "",
        f"INSERT INTO questions ({', '.join(columns)}) VALUES",
        ",\n".join(format_row(r) for r in rows),
        "ON CONFLICT (id) DO NOTHING;",
        ""
    ]

    UP_MIGRATION.write_text("\n".join(up_sql), encoding='utf-8', newline='\n')
    print(f"Wrote {UP_MIGRATION} ({UP_MIGRATION.stat().st_size:,} bytes)")

    down_sql = (
        "-- 000087_pte_speaking_bank_3.down.sql\n"
        "-- Preserve learner references and editorial changes on rollback.\n"
        "SELECT 1;\n"
    )
    DOWN_MIGRATION.write_text(down_sql, encoding='utf-8', newline='\n')
    print(f"Wrote {DOWN_MIGRATION}")


if __name__ == '__main__':
    validate_and_compile()
