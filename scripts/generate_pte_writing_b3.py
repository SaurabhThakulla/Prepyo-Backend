"""Generate PTE Writing Bank Batch 3 (Migration 000088).

This script generates 150 original PTE Writing practice questions:
  - 50 Summarize Written Text (SWT) (context_passage 150-300 words, model_answer 5-75 words in 1 sentence, time 600s, points 10)
  - 50 Write Essay (WE) (prompt statement + instructions, model_essay 200-300 words, time 1200s, points 15)
  - 50 Repeated Essay Questions (REQ) (prompt statement + instructions, model_essay 200-300 words, time 1200s, points 15)

Self-validation enforces:
  - Exactly 50 questions per task type (150 total).
  - All IDs are unique and prefixed with pte-b3-<task>-<nnn>.
  - Constraints on word counts, sentence counts, and model answers.
  - No digits, no "percent", no "studies" in model essays.
  - Real newlines in essay prompts (no literal \n).
  - No item repeats a live question of the same task or a live reading topic
    (question_clashes in scripts/live_topics.py) unless the pair is reviewed and allowed;
    essay prompts check against BOTH live Write Essay and live Repeated Essay Questions.
  - Generates migrations/000088_pte_writing_bank_3.up.sql (INSERT ... ON CONFLICT (id) DO NOTHING)
    and migrations/000088_pte_writing_bank_3.down.sql (SELECT 1;).
  - Uses Unix line endings (newline='\n').
"""

import hashlib
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))

from scripts.live_topics import question_clashes
from scripts.pte_writing_b3.swt import SWT_ITEMS
from scripts.pte_writing_b3.we import WE_ITEMS
from scripts.pte_writing_b3.repeated_essay import REQ_ITEMS

VERSION = 'pte-2026-01'
UP_MIGRATION = ROOT / 'migrations' / '000088_pte_writing_bank_3.up.sql'
DOWN_MIGRATION = ROOT / 'migrations' / '000088_pte_writing_bank_3.down.sql'

ALLOWED_TOPIC_OVERLAPS = {
    'Academic Ability Streaming in Secondary Schools': {'Cursive handwriting in schools', 'Financial literacy in schools', 'Creativity in schools', 'Teaching coding in schools', 'Homework in primary school', 'Competitive sport in schools', 'Later Starts: Teenagers and the School Day', 'Single-sex schools'},
    'Acoustic Barriers Along High-Speed Railway Corridors': {'Trade along the Silk Road'},
    'Aesthetic Ornamentation versus Utilitarian Minimalism': {'Exams versus continuous assessment', 'Experience versus qualifications', 'Traditional versus modern medicine', 'Small shops versus supermarkets'},
    'Airborne Multispectral Cameras for Forest Health Monitoring': {"Nepal's community forests", 'Mangrove forests', 'Noise and health', 'Forests of the Tide'},
    'Algorithmic News Selection': {'News and entertainment'},
    'Artificial Intelligence in Creative Literary Writing': {'Octopus intelligence', 'Artificial intelligence in the workplace', 'Can machines be creative?'},
    'Attending University Locally versus Moving Away': {'Exams versus continuous assessment', 'Experience versus qualifications', 'University tuition fees', 'Vocational training or university', 'Small shops versus supermarkets', 'Traditional versus modern medicine'},
    'Audio Recorders for Monitoring Forest Wildlife': {"Nepal's community forests", 'Mangrove forests', 'Forests of the Tide'},
    'Automated Acoustic Monitoring for Livestock Welfare': {'Antibiotics in livestock farming'},
    'Autonomous Cargo Drone Delivery in Rural Regions': {'Rural service for new doctors'},
    'Banning Billboards Along Scenic Corridors': {'Banning cars from city centres'},
    'Banning Disposable Alkaline Batteries': {'Banning cars from city centres'},
    'Banning Drive-Through Restaurants': {'Banning cars from city centres'},
    'Bird Friendly Architectural Glass Mandates': {'How birds find north'},
    'Carbon Border Adjustment Taxes': {'Higher taxes on high earners', 'Fast food taxes', 'Do carbon offsets work?'},
    'Citizen Juries in Urban Planning': {'Urban beekeeping', 'Cooling the urban heat island', 'Urban heat islands'},
    'Compulsory Culinary Education in Secondary Schools': {'Cursive handwriting in schools', 'Compulsory voting', 'Financial literacy in schools', 'Compulsory foreign languages', 'Creativity in schools', 'Teaching coding in schools', 'Homework in primary school', 'Competitive sport in schools', 'Later Starts: Teenagers and the School Day', 'Mother tongue education', 'Compulsory national service', 'Westfield Adult Education: Autumn Short Courses', 'Single-sex schools'},
    'Compulsory Data Literacy for Undergraduates': {'Compulsory foreign languages', 'Compulsory voting', 'Financial literacy in schools', 'Compulsory national service'},
    'Dark Sky Environmental Lighting Codes': {'Lighting the way: a short history of lighthouses'},
    'Deep Lakes with Permanently Separated Water Layers': {"Nepal's rising glacial lakes", 'Glaciers and water supply', 'Turning seawater into drinking water', 'Glacial lakes in the Himalaya', 'Drinking water from the sea', 'Ballast water and species transfer'},
    'Digital Nomad Visas and Housing': {'Handwriting in the digital age', 'Digital archives', 'Museums in the digital age'},
    'Domestic Medical Equipment Manufacturing': {'Animal testing in medical research'},
    'Early Warning Vibration Sensors on Large Dams': {'Early childhood education'},
    'Earth Observation Using Satellite Radar Imagery': {'Research using twins', 'Rammed earth'},
    'Employee Access to Workplace Evaluation Records': {'Artificial intelligence in the workplace'},
    'Employee Choice in Flexible Working Hours': {'Working longer hours'},
    'Extended Producer Responsibility for Packaging': {'Single-use packaging', 'Health: individual or government responsibility', 'Climate responsibility', 'Why we buy extended warranties'},
    'Food Raw Ingredient Origin Labeling': {'Fast food taxes', 'The origins of chess', 'Origins of the marathon', 'Fermented foods', 'Food waste'},
    'Fossil Shells as Records of Ancient Oceans': {'The ocean garbage patch', 'Mapping the ocean floor'},
    'Free Open-Access University Textbooks': {'Free public transport', 'Vocational training or university', 'University tuition fees'},
    'Free-Market Competition and Innovation': {'Free public transport', 'Competition in the classroom'},
    'Funding University Degrees with Low Market Demand': {'Science or humanities funding', 'Vocational training or university', 'Funding the arts', 'University tuition fees'},
    'Graduated Driver Licensing for Adults': {'Westfield Adult Education: Autumn Short Courses'},
    'Harvesting Electrical Energy from Highway Traffic Pressure': {'Tidal energy', 'Harvesting fog', 'Lithium and the energy transition'},
    'Healthcare Allocations for Disease Prevention': {'Public or private healthcare'},
    'Heat Exchangers for Reclaiming Energy from Ceramic Kilns': {'Tidal energy', 'Cooling the urban heat island', 'Heat pumps', 'Urban heat islands', 'Lithium and the energy transition'},
    'High-Interest Credit Cards and Youth Vulnerability': {'Credit unions', 'Supermarket loyalty cards', 'Parents and youth crime'},
    'Hiring Generalists versus Technical Specialists': {'Exams versus continuous assessment', 'Experience versus qualifications', 'Traditional versus modern medicine', 'Small shops versus supermarkets'},
    'Household Appliance Repair Education': {'Mother tongue education', 'Westfield Adult Education: Autumn Short Courses'},
    'How Seed Dormancy Protects Crops Against Drought': {'Crop-growing skyscrapers'},
    'Integrating Agriculture into School Grounds': {'Cursive handwriting in schools', 'Financial literacy in schools', 'Creativity in schools', 'Teaching coding in schools', 'Homework in primary school', 'Competitive sport in schools', 'Later Starts: Teenagers and the School Day', 'Single-sex schools'},
    'Mandated Paid Educational Leave for Workers': {'Retraining automated workers', 'Harlow Logistics Staff Handbook: Annual Leave and Absence'},
    'Mandatory Water Efficiency in Buildings': {'Turning seawater into drinking water', 'Drinking water from the sea', 'Water scarcity', 'Ballast water and species transfer'},
    'Manufacturing Low-Carbon Cement from Industrial Ash': {'Carbon pricing', 'Do carbon offsets work?'},
    'Mapping Subsurface Water Tables with Electrical Resistivity': {'Glaciers and water supply', 'Turning seawater into drinking water', 'Mapping the ocean floor', 'Drinking water from the sea', 'Ballast water and species transfer'},
    'Marine Microorganisms and Oceanic Carbon Storage': {'Carbon pricing', 'Do carbon offsets work?'},
    'Measuring Historical Lake Levels from Ancient Shorelines': {'Measuring what matters', 'Glacial lakes in the Himalaya', "Nepal's rising glacial lakes"},
    'Measuring Ocean Temperature with Sound Waves': {'Listening Beneath the Waves', 'Measuring what matters', 'An Introduction to Film Sound', 'Mapping the ocean floor', 'The ocean garbage patch'},
    'Mechanical Flywheels for Grid Energy Storage': {'Tidal energy', 'Lithium and the energy transition'},
    'Micro-Apartments in Urban Centers': {'Urban beekeeping', 'Cooling the urban heat island', 'Urban heat islands'},
    'Monitoring Bridge Safety with Fiber-Optic Sensors': {'Roundabouts and road safety'},
    'National Highway Maximum Speed Limits': {'Social media age limits', 'Compulsory national service'},
    'National Seed Bank Networks': {'Compulsory national service'},
    'Natural Artistic Genius versus Deliberate Practice': {'Exams versus continuous assessment', 'Experience versus qualifications', 'Traditional versus modern medicine', 'Small shops versus supermarkets'},
    'Nutrient Cycling by Microorganisms in Coastal Mudflats': {'Mangroves as coastal defence'},
    'Ocean Exploration Funding': {'Science or humanities funding', 'Plastic in the oceans', 'Funding the arts', 'Mapping the ocean floor', 'The ocean garbage patch', 'Spending on space exploration'},
    'Open-Source Government Software Mandates': {'Health: individual or government responsibility'},
    'Phasing Out Heavy Bunker Fuel in Commercial Maritime Transport': {'Free public transport'},
    'Phasing Out Paper Receipts': {'From coins to paper: the invention of banknotes', 'The first paper money'},
    'Philosophy in Primary Curricula': {'Homework in primary school'},
    'Preserving Physical Paper Archives': {'Digital archives', 'From coins to paper: the invention of banknotes', 'The first paper money', 'Preserving minority languages'},
    'Preserving Vernacular Architecture': {'Preserving minority languages'},
    'Print Journalism versus Digital News Portals': {'Exams versus continuous assessment', 'News and entertainment', 'Museums in the digital age', 'Handwriting in the digital age', 'Experience versus qualifications', 'Digital archives', 'Small shops versus supermarkets', 'Traditional versus modern medicine'},
    'Protecting Underground Pipelines from Soil Corrosion': {'Microplastics in farm soil'},
    'Public Funding for Nuclear Fusion Energy': {'Free public transport', 'Science or humanities funding', 'Nuclear power', 'Privatising public services', 'Public or private healthcare', 'Funding the arts', 'Public libraries'},
    'Public Funding for Planetary Asteroid Defense and Monitoring': {'Free public transport', 'Science or humanities funding', 'Privatising public services', 'Public or private healthcare', 'Funding the arts', 'Public libraries'},
    'Public Parks versus Indoor Sports Centers': {'Exams versus continuous assessment', 'Free public transport', 'Experience versus qualifications', 'Privatising public services', 'Competitive sport in schools', 'Public or private healthcare', 'Public libraries', "Sports stars' salaries", 'Small shops versus supermarkets', 'Traditional versus modern medicine'},
    'Public Quiet Zones in Cities': {'Free public transport', 'Privatising public services', 'The birth of time zones', 'Public or private healthcare', 'Public libraries'},
    'Public Waterfront Walking Corridors': {'Privatising public services', 'Free public transport', 'Public or private healthcare', 'Public libraries'},
    'Reflective Rooftop Coatings That Cool Without Energy': {'Tidal energy', 'Lithium and the energy transition'},
    'Relocating Government Agencies to Regional Towns': {'Health: individual or government responsibility'},
    'Replenishing Underground Aquifers with Seasonal Runoff': {'Choosing seasonal vaccine strains'},
    'Restricting Mass Tourism in Historical Enclaves': {'Tourism and local culture', 'Tourism in the Himalaya'},
    'Right to Disconnect Legislation': {'Languages without left and right'},
    'Small Independent Colleges versus Massive Universities': {'Exams versus continuous assessment', 'Experience versus qualifications', 'University tuition fees', 'Vocational training or university', 'Small shops versus supermarkets', 'Traditional versus modern medicine'},
    'Soil Conservation with Seasonal Cover Crops': {'Choosing seasonal vaccine strains', 'Microplastics in farm soil', 'Crop-growing skyscrapers'},
    'Standardized Testing in University Admissions': {'Genetic testing', 'Vocational training or university', 'Animal testing in medical research', 'University tuition fees'},
    'Street Orientation and Daylight in Urban Design': {'Urban beekeeping', 'Daylight saving time', 'Cooling the urban heat island', 'Urban heat islands'},
    'Subsidising Electric Agricultural Machinery': {'Electric cars in Nepal'},
    'Subsidising Electric Bicycles': {'Electric cars in Nepal'},
    'Taxpayer Funding for Historical Monument Conservation': {'Science or humanities funding', 'Funding the arts'},
    'Traffic Free Zones Around Primary Schools': {'Free public transport', 'Cursive handwriting in schools', 'Financial literacy in schools', 'Creativity in schools', 'Teaching coding in schools', 'Homework in primary school', 'The birth of time zones', 'Competitive sport in schools', 'Later Starts: Teenagers and the School Day', 'Single-sex schools'},
    'Universal Basic Income': {'Basic research or practical research'},
    'Urban Heating from Geothermal Water Networks': {'Glaciers and water supply', 'Urban beekeeping', 'Turning seawater into drinking water', 'Cooling the urban heat island', 'Urban heat islands', 'Drinking water from the sea', 'Ballast water and species transfer'},
    'Using Beneficial Insects in Greenhouse Crop Protection': {'Research using twins', 'Crop-growing skyscrapers'},
    'Using Living Tree Roots to Stabilize Hillside Slopes': {'Research using twins'},
    'Water Conservation with Precision Drip Irrigation': {'Glaciers and water supply', 'Turning seawater into drinking water', 'Drinking water from the sea', 'Ballast water and species transfer'},
    'Water-Repellent Coatings Inspired by Lotus Leaves': {'Glaciers and water supply', 'Turning seawater into drinking water', 'Drinking water from the sea', 'Rubber leaves the Amazon', 'Ballast water and species transfer'},
    'Weighting Teaching Quality in Universities': {'Teaching children to read', 'Vocational training or university', 'Teaching coding in schools', 'University tuition fees'},
    'Workplace Keystroke Monitoring': {'Artificial intelligence in the workplace'},
    'Workplace Mindfulness Initiatives': {'Artificial intelligence in the workplace'},
}


def escape_sql(val):
    if val is None:
        return 'NULL'
    return "'" + str(val).replace("'", "''") + "'"


def validate_all():
    print(f"Validating Batch C items: SWT={len(SWT_ITEMS)}, WE={len(WE_ITEMS)}, REQ={len(REQ_ITEMS)}")
    assert len(SWT_ITEMS) == 50, f"Expected 50 SWT, got {len(SWT_ITEMS)}"
    assert len(WE_ITEMS) == 50, f"Expected 50 WE, got {len(WE_ITEMS)}"
    assert len(REQ_ITEMS) == 50, f"Expected 50 REQ, got {len(REQ_ITEMS)}"

    seen_titles = set()
    seen_texts = set()

    # 1. Summarize Written Text
    for i, item in enumerate(SWT_ITEMS):
        t = item['title']
        assert t not in seen_titles, f"Duplicate title: {t}"
        seen_titles.add(t)

        pw = len(item['passage'].split())
        assert 150 <= pw <= 300, f"SWT #{i+1} passage word count {pw} not in [150, 300]"

        mw = len(item['model_answer'].split())
        assert 5 <= mw <= 75, f"SWT #{i+1} model_answer word count {mw} not in [5, 75]"
        assert item['model_answer'].endswith('.'), f"SWT #{i+1} model answer does not end with period"
        assert item['model_answer'].count('.') == 1, f"SWT #{i+1} model answer contains multiple periods"

        c = question_clashes('PTE', t, allowed=ALLOWED_TOPIC_OVERLAPS.get(t, ()), type_name='Summarize Written Text')
        assert not c, f"SWT #{i+1} '{t}' has unreviewed clashes: {c}"

        assert item['passage'] not in seen_texts, f"Duplicate passage in SWT #{i+1}"
        seen_texts.add(item['passage'])

    # 2. Write Essay
    for i, item in enumerate(WE_ITEMS):
        t = item['title']
        assert t not in seen_titles, f"Duplicate title: {t}"
        seen_titles.add(t)

        p = item['prompt']
        assert "\n\n" in p, f"WE #{i+1} prompt missing real newlines"
        assert "\\n" not in p, f"WE #{i+1} prompt contains literal backslash-n"

        e = item['model_essay']
        ew = len(e.split())
        assert 200 <= ew <= 300, f"WE #{i+1} model essay word count {ew} not in [200, 300]"
        assert not re.search(r"\d", e), f"WE #{i+1} has digits in model_essay: {re.findall(r'\d+', e)}"
        assert "percent" not in e.lower(), f"WE #{i+1} has 'percent' in model_essay"
        assert not re.search(r"\bstudies\b", e, re.I), f"WE #{i+1} has 'studies' in model_essay"

        c_we = question_clashes('PTE', t, allowed=ALLOWED_TOPIC_OVERLAPS.get(t, ()), type_name='Write Essay')
        c_req = question_clashes('PTE', t, allowed=ALLOWED_TOPIC_OVERLAPS.get(t, ()), type_name='Repeated Essay Questions')
        c = list(set(c_we + c_req))
        assert not c, f"WE #{i+1} '{t}' has unreviewed clashes: {c}"

        assert item['statement'] not in seen_texts, f"Duplicate statement in WE #{i+1}"
        seen_texts.add(item['statement'])

    # 3. Repeated Essay Questions
    for i, item in enumerate(REQ_ITEMS):
        t = item['title']
        assert t not in seen_titles, f"Duplicate title: {t}"
        seen_titles.add(t)

        p = item['prompt']
        assert "\n\n" in p, f"REQ #{i+1} prompt missing real newlines"
        assert "\\n" not in p, f"REQ #{i+1} prompt contains literal backslash-n"

        e = item['model_essay']
        ew = len(e.split())
        assert 200 <= ew <= 300, f"REQ #{i+1} model essay word count {ew} not in [200, 300]"
        assert not re.search(r"\d", e), f"REQ #{i+1} has digits in model_essay: {re.findall(r'\d+', e)}"
        assert "percent" not in e.lower(), f"REQ #{i+1} has 'percent' in model_essay"
        assert not re.search(r"\bstudies\b", e, re.I), f"REQ #{i+1} has 'studies' in model_essay"

        c_we = question_clashes('PTE', t, allowed=ALLOWED_TOPIC_OVERLAPS.get(t, ()), type_name='Write Essay')
        c_req = question_clashes('PTE', t, allowed=ALLOWED_TOPIC_OVERLAPS.get(t, ()), type_name='Repeated Essay Questions')
        c = list(set(c_we + c_req))
        assert not c, f"REQ #{i+1} '{t}' has unreviewed clashes: {c}"

        assert item['statement'] not in seen_texts, f"Duplicate statement in REQ #{i+1}"
        seen_texts.add(item['statement'])

    # 4. Assert that no two essay statements in WE + REQ share most of their key words
    STOP_WORDS = set("""
        a about above after again against all also am an and any are aren't as at be because been before being below between both but by can cannot couldn't did didn't do does doesn't doing don't down during each few for from further had hadn't has hasn't have haven't having he he'd he'll he's her here here's hers herself him himself his how how's i i'd i'll i'm i've if in into is isn't it it's its itself let's me more most mustn't my myself no nor not of off on once only or other ought our ours ourselves out over own same shan't she she'd she'll she's should shouldn't so some such than that that's the their theirs them themselves then there there's these they they'd they'll they're they've this those through to too under until up very was wasn't we we'd we'll we're we've were weren't what what's when when's where where's which while who who's whom why why's with won't would wouldn't you you'd you'll you're you've government governments should public people citizens society students school schools university universities
    """.split())
    all_essays = WE_ITEMS + REQ_ITEMS
    for i in range(len(all_essays)):
        s1 = set(re.findall(r'[a-z]+', all_essays[i]['statement'].lower())) - STOP_WORDS
        for j in range(i + 1, len(all_essays)):
            s2 = set(re.findall(r'[a-z]+', all_essays[j]['statement'].lower())) - STOP_WORDS
            shared = s1 & s2
            min_len = min(len(s1), len(s2))
            ratio = len(shared) / min_len if min_len > 0 else 0
            assert ratio < 0.50, (
                f"Essay statements share most key words ({ratio:.2%}):\n"
                f"  '{all_essays[i]['title']}' vs '{all_essays[j]['title']}' (shared: {shared})"
            )

    print("All 150 items strictly passed self-validation!")


def generate_migration_sql():
    rows = []

    # 1. Summarize Written Text (50 items)
    for i, item in enumerate(SWT_ITEMS, 1):
        qid = f"pte-b3-swt-{i:03d}"
        prompt = "Read the passage below and summarise it in ONE single sentence of between 5 and 75 words."
        row = (
            qid, VERSION, 'PTE', 'writing', 'summarize-written-text', 'Summarize Written Text',
            item['title'], prompt, item['passage'], None, None, None,
            item['model_answer'], 0, 600, 10, 'medium', item['explanation'],
            "ARRAY['PTE', 'Original practice', 'Batch 3', 'Summarize Written Text']",
            "ARRAY['PTE']", "TRUE"
        )
        rows.append(row)

    # 2. Write Essay (50 items)
    for i, item in enumerate(WE_ITEMS, 1):
        qid = f"pte-b3-we-{i:03d}"
        row = (
            qid, VERSION, 'PTE', 'writing', 'pte-write-essay', 'Write Essay',
            item['title'], item['prompt'], None, None, None, None,
            item['model_essay'], 0, 1200, 15, 'medium', item['explanation'],
            "ARRAY['PTE', 'Original practice', 'Batch 3', 'Write Essay']",
            "ARRAY['PTE']", "TRUE"
        )
        rows.append(row)

    # 3. Repeated Essay Questions (50 items)
    for i, item in enumerate(REQ_ITEMS, 1):
        qid = f"pte-b3-req-{i:03d}"
        row = (
            qid, VERSION, 'PTE', 'writing', 'pte-repeated-essay-questions', 'Repeated Essay Questions',
            item['title'], item['prompt'], None, None, None, None,
            item['model_essay'], 0, 1200, 15, 'medium', item['explanation'],
            "ARRAY['PTE', 'Original practice', 'Batch 3', 'Repeated Essay Questions']",
            "ARRAY['PTE']", "TRUE"
        )
        rows.append(row)

    lines = [
        "-- 000088_pte_writing_bank_3.up.sql",
        "-- Batch C: 150 original PTE Writing practice questions (50 SWT, 50 WE, 50 REQ).",
        "-- Deterministic IDs pte-b3-<task>-<nnn>. Replay-safe INSERT ... ON CONFLICT DO NOTHING.",
        "",
        "INSERT INTO questions (id, exam_version_id, exam, skill, type_id, type_name, title, prompt, context_passage, audio_transcript, image_url, figure_data, model_answer, prep_time_seconds, time_limit_seconds, points, difficulty, explanation, tags, supported_exams, is_published) VALUES",
    ]

    value_clauses = []
    for r in rows:
        v = (
            f"    ({escape_sql(r[0])}, {escape_sql(r[1])}, {escape_sql(r[2])}, {escape_sql(r[3])}, "
            f"{escape_sql(r[4])}, {escape_sql(r[5])}, {escape_sql(r[6])}, {escape_sql(r[7])}, "
            f"{escape_sql(r[8])}, {escape_sql(r[9])}, {escape_sql(r[10])}, {escape_sql(r[11])}, "
            f"{escape_sql(r[12])}, {r[13]}, {r[14]}, {r[15]}, {escape_sql(r[16])}, {escape_sql(r[17])}, "
            f"{r[18]}, {r[19]}, {r[20]})"
        )
        value_clauses.append(v)

    lines.append(",\n".join(value_clauses) + "\nON CONFLICT (id) DO NOTHING;\n")
    return "\n".join(lines)


def main():
    validate_all()
    sql = generate_migration_sql()
    UP_MIGRATION.write_text(sql, encoding='utf-8', newline='\n')
    print(f"Wrote {UP_MIGRATION} ({len(sql)} bytes)")

    down_sql = (
        "-- 000088_pte_writing_bank_3.down.sql\n"
        "-- Preserve learner references and editorial changes on rollback.\n"
        "SELECT 1;\n"
    )
    DOWN_MIGRATION.write_text(down_sql, encoding='utf-8', newline='\n')
    print(f"Wrote {DOWN_MIGRATION} ({len(down_sql)} bytes)")


if __name__ == '__main__':
    main()
