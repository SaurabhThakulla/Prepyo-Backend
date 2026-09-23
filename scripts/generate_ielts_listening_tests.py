"""Generate original IELTS Listening mock tests (migration 000075).

Each test follows the public IELTS Listening format: four parts of ten
questions; Part 1 an everyday conversation, Part 2 an everyday monologue,
Part 3 an educational discussion, Part 4 an academic lecture. All content is
original Prepyo practice material, not official or recalled test content.

Every question carries `evidence`: the exact words in the script where its
answer is given. The script checks, before writing SQL, that
  - every piece of evidence appears in its part's script,
  - evidence follows the order of the questions ("The questions are in the
    same order as the information in the recording"),
  - every completion answer is heard in the script and fits the word limit,
  - each part has ten questions and each test forty.
Output is deterministic and replay-safe (ON CONFLICT DO NOTHING).
"""
import json
import re
from pathlib import Path
from urllib.parse import quote
from html import escape

ROOT = Path(__file__).resolve().parents[1]
VERSION = 'ielts-2026-01'


def map_svg():
    """Plan of Oakfield Nature Reserve for Part 2. Letters only; no answers."""
    parts = ['<svg xmlns="http://www.w3.org/2000/svg" width="640" height="480" viewBox="0 0 640 480">',
             '<rect width="640" height="480" fill="#ffffff"/>',
             '<g font-family="Arial" fill="#172554">',
             '<text x="20" y="28" font-size="18" font-weight="bold">Oakfield Nature Reserve</text>',
             '<text x="590" y="28" font-size="16">N &#8593;</text>',
             # woodland
             '<rect x="40" y="50" width="240" height="95" rx="14" fill="#dcfce7" stroke="#15803d"/>',
             '<text x="120" y="102" font-size="15" fill="#166534">Woodland</text>',
             # lake
             '<ellipse cx="480" cy="210" rx="110" ry="68" fill="#dbeafe" stroke="#1d4ed8"/>',
             '<text x="458" y="215" font-size="15" fill="#1e3a8a">Lake</text>',
             # stream
             '<path d="M0 335 C 120 320, 200 350, 320 335 S 520 320, 640 338" stroke="#60a5fa" stroke-width="10" fill="none"/>',
             '<text x="30" y="365" font-size="13" fill="#1e3a8a">Stream</text>',
             # main path
             '<path d="M320 455 V 45" stroke="#a8a29e" stroke-width="14" fill="none"/>',
             '<text x="328" y="250" font-size="12" fill="#57534e" transform="rotate(90 328 250)">Main path</text>',
             # lakeside path
             '<path d="M320 300 H 470 C 540 300, 590 290, 600 250" stroke="#a8a29e" stroke-width="8" fill="none"/>',
             # bridge
             '<rect x="304" y="322" width="32" height="26" fill="#78350f"/>',
             '<text x="342" y="358" font-size="12">Bridge</text>',
             # entrance
             '<rect x="286" y="455" width="68" height="20" fill="#172554"/>',
             '<text x="272" y="447" font-size="13">Main entrance</text>']
    boxes = {'A': (190, 405), 'B': (400, 405), 'C': (230, 272), 'D': (70, 385),
             'E': (458, 118), 'F': (150, 158), 'G': (300, 52), 'H': (598, 192)}
    for letter, (x, y) in boxes.items():
        parts.append(f'<rect x="{x}" y="{y}" width="40" height="30" rx="4" fill="#fef3c7" stroke="#92400e"/>')
        parts.append(f'<text x="{x + 13}" y="{y + 21}" font-size="16" font-weight="bold">{letter}</text>')
    parts.append('</g></svg>')
    return 'data:image/svg+xml,' + quote(''.join(parts), safe='')


ABC = lambda *texts: [{'id': chr(65 + i), 'text': t} for i, t in enumerate(texts)]
MAP_OPTIONS = [{'id': letter, 'text': f'Location {letter}'} for letter in 'ABCDEFGH']
DECISIONS = ABC('introduce it immediately', 'try it out first', 'reject it')

TESTS = [dict(
    key='t1',
    title='IELTS Listening Practice Test 1',
    parts=[
        dict(
            setting='You will hear a man phoning a holiday activity centre to book a course for his daughter.',
            script=(
                "Receptionist: Good morning, Riverside Holiday Activity Centre. How can I help you? "
                "Caller: Hello. I'd like to book a place on one of your summer courses for my daughter, please. "
                "Receptionist: Of course. I'll fill in a booking form for you. Could I have your daughter's name? "
                "Caller: Yes, it's Emily Hartley. That's H, A, R, T, L, E, Y. "
                "Receptionist: Thank you. And how old is Emily? "
                "Caller: She's ten at the moment, but she'll be eleven by the time the course starts, so I suppose you should put eleven. "
                "Receptionist: Yes, we go by the age on the first day. Now, which course is she interested in? At the moment we still have places on sailing, climbing and mountain biking. "
                "Caller: She did sailing last year and enjoyed it, but this time she's really keen on climbing. "
                "Receptionist: Climbing, then. There are two climbing courses in August. The first one runs from the seventh to the eleventh, but I'm afraid that one's full. The second runs from the fourteenth to the eighteenth. "
                "Caller: The second one's fine. We're away at the beginning of the month anyway. "
                "Receptionist: Good. Most of our courses start at nine o'clock, but because the climbing group travels to the quarry first, that one starts at half past nine each day. "
                "Caller: Half past nine. That's easier for us, actually. What does she need to bring? "
                "Receptionist: She'll need a packed lunch every day, and she should wear trainers rather than sandals. We provide the helmets and the harnesses, so she doesn't need those. But please make sure she brings a hat, because the group is outdoors for most of the day. "
                "Caller: A hat. Right. Is there anything else you need to know? "
                "Receptionist: Yes. Does Emily have any allergies or medical conditions? "
                "Caller: She's not allergic to anything we know of, except peanuts. She has to be careful with those. "
                "Receptionist: Peanuts. I'll make a note of that, and we'll tell the instructors. Now, about payment. The course costs one hundred and twenty pounds, but if you've booked with us before, you get a discount. "
                "Caller: We came last year, yes. "
                "Receptionist: In that case there's a discount of ten per cent. It used to be fifteen, but that changed in January. "
                "Caller: Ten per cent is still very welcome. "
                "Receptionist: One more question for our records. How did you first hear about the centre? Was it the radio advertisement? "
                "Caller: I think my wife heard something on the radio, but actually we found you through the local newspaper. There was an article about your summer courses. "
                "Receptionist: The newspaper, lovely. And finally, who will be collecting Emily at the end of each day? "
                "Caller: I'll bring her in the mornings, but I work in the afternoons, so her grandmother will collect her. "
                "Receptionist: Her grandmother. She'll need to show some identification on the first day. "
                "Caller: No problem, I'll let her know. "
                "Receptionist: That's everything. We'll email you a confirmation this afternoon. "
                "Caller: Thank you very much. Goodbye."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Riverside Holiday Activity Centre: booking form',
                instructions='Complete the form below. Write ONE WORD AND/OR A NUMBER for each answer.', limit=1,
                questions=[
                    ('Child\'s surname: ________', ['Hartley'], 'H, A, R, T, L, E, Y'),
                    ('Age on the first day: ________', ['11', 'eleven'], 'put eleven'),
                    ('Course chosen: ________', ['climbing'], 'really keen on climbing'),
                    ('Course dates: 14 to ________ August', ['18', '18th', 'eighteenth'], 'fourteenth to the eighteenth'),
                    ('Start time each day: ________ a.m.', ['9.30', '9:30'], 'that one starts at half past nine'),
                    ('Bring: packed lunch, trainers and a ________', ['hat'], 'brings a hat'),
                    ('Allergy: ________', ['peanuts'], 'except peanuts'),
                    ('Returning customer discount: ________ %', ['10', 'ten'], 'discount of ten per cent'),
                    ('Heard about the centre from: the local ________', ['newspaper'], 'through the local newspaper'),
                    ('Collected each afternoon by: her ________', ['grandmother'], 'her grandmother will collect her'),
                ])],
        ),
        dict(
            setting='You will hear a ranger talking to a group of visitors at Oakfield Nature Reserve.',
            script=(
                "Ranger: Good morning, everyone, and welcome to Oakfield Nature Reserve. My name's Sam and I'm one of the rangers here. Before you set off, I'd like to tell you a little about the reserve and show you where things are on the map. "
                "The land here has an interesting history. Many visitors assume it was once farmland, because of the farms that surround it, but in fact, for most of the last century, this was a quarry. Sand and gravel were dug out here until the nineteen eighties, and the lake you'll see is simply the old quarry pit, which filled with water when the digging stopped. "
                "We've made quite a few changes recently. Last year we opened the visitor centre, and a lot of people ask whether the café is new, but it has been here for years. What's really new this year is the boardwalk across the wet meadow, which means you can now get close to the reed beds without getting your feet wet. "
                "If you'd like to join a guided walk, they leave from the visitor centre. In the summer they used to start at ten, but we found that most families arrive a bit later, so now there's one walk a day, at eleven o'clock. There's also a bat walk on Friday evenings, but that has to be booked in advance. "
                "We ask all visitors to follow a few simple rules. Dogs are welcome as long as they're kept on a lead, and feeding the ducks is fine as long as you use the special food we sell in the shop. What we do not allow anywhere on the reserve is the use of drones, because they disturb nesting birds. "
                "You might be wondering how a place like this is paid for. We do receive a small grant from the local council, and we've had lottery money for particular projects such as the boardwalk, but the reserve depends mainly on the membership fees paid by our supporters, so please do think about joining. "
                "Now, if you look at your map, you'll see the main entrance at the bottom. As you come through the main entrance, the café is immediately on your left. Don't confuse it with the ticket office, which is on the right-hand side of the entrance. "
                "If you follow the main path north, you'll come to a bridge over the stream. Straight after the bridge, on your left, you'll find the picnic area, with tables and a small play area for children. "
                "From the bridge, take the path to the right, which runs along the southern shore of the lake. Some people expect the bird hide to be at the north end of the lake, where the old boathouse was, but that building is now used for storage. The new bird hide is where the lakeside path ends, on the far eastern side of the lake. "
                "Back on the main path, carry on north. Where the woodland begins on your left, you'll see the wildflower meadow, which runs along the southern edge of the trees. It's at its best in June. "
                "And at the very top of the reserve, where the main path ends, there are toilets next to the viewing point. Right, has anyone got any questions before we start?"),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='',
                     instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('What was the site used for before it became a nature reserve?', ABC('farming', 'a quarry', 'a boating lake'), 'B', 'this was a quarry'),
                         ('What is new at the reserve this year?', ABC('the visitor centre', 'the café', 'the boardwalk'), 'C', "What's really new this year is the boardwalk"),
                         ('When does the daily guided walk start?', ABC('10 a.m.', '11 a.m.', 'in the evening'), 'B', "at eleven o'clock"),
                         ('What are visitors not allowed to do on the reserve?', ABC('bring dogs', 'feed the ducks', 'fly drones'), 'C', 'the use of drones'),
                         ('How is the reserve mainly paid for?', ABC('a grant from the local council', 'lottery money', 'membership fees'), 'C', 'depends mainly on the membership fees'),
                     ]),
                dict(type_id='ielts-listening-map', heading='Oakfield Nature Reserve',
                     instructions='Label the map below. Choose the correct letter, A-H.', image='map',
                     questions=[
                         ('Café', MAP_OPTIONS, 'A', 'the café is immediately on your left'),
                         ('Picnic area', MAP_OPTIONS, 'C', "you'll find the picnic area"),
                         ('Bird hide', MAP_OPTIONS, 'H', 'The new bird hide is where the lakeside path ends'),
                         ('Wildflower meadow', MAP_OPTIONS, 'F', "you'll see the wildflower meadow"),
                         ('Toilets', MAP_OPTIONS, 'G', 'there are toilets next to the viewing point'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear two students, Maya and Tom, discussing their project on food waste with their tutor.',
            script=(
                "Tutor: So, Maya and Tom, how's the food waste project going? "
                "Maya: Quite well, I think. As you know, we decided to focus on the main canteen. "
                "Tutor: Remind me why you chose the canteen rather than the halls of residence. "
                "Tom: Well, at first we assumed the halls would produce more waste, but we couldn't get access to them. Then the canteen manager contacted the department. She'd read about our course and asked if any students could help her reduce the amount of food being thrown away, so that decided it. "
                "Tutor: That's a useful contact to have. And how did the first survey go? "
                "Maya: Not as well as we'd hoped. We had plenty of responses, over two hundred, and it only took about five minutes to fill in. The trouble was that some of the questions were confusing. For example, we asked how often people wasted food, and people understood that in very different ways. "
                "Tom: So we rewrote those questions for the second round, and gave examples. "
                "Tutor: Good. What about the waste data itself? Was there anything you didn't expect? "
                "Tom: We expected most of the waste to come from lunch, because that's the busiest meal. And it did, in total. But what surprised us was that breakfast produced more waste per person than any other meal. People take cereal and toast and then rush off to lectures. "
                "Tutor: Interesting. Have you thought about how you'll present all this? "
                "Maya: We were going to write it up as a standard report. "
                "Tutor: That's fine, but I think the canteen staff would find it easier to use if you also produced a short summary with charts, as well as the full report. Managers rarely have time to read forty pages. "
                "Maya: That makes sense. We could do a two-page version. "
                "Tutor: And what's your next step? "
                "Tom: We'd thought about interviewing the kitchen staff, but they're very busy at the moment, so we've agreed to spend a week weighing the waste from each meal ourselves, to check the figures the canteen gave us. "
                "Tutor: Sensible. Now, you've come up with a number of ideas for reducing waste. Which ones are you recommending? "
                "Maya: The first is to use smaller plates. The evidence from other universities is really strong, and the manager has already ordered them, so there's no reason to wait. We'll recommend introducing them straight away. "
                "Tom: Then there's the idea of a daily menu vote, where students choose the next day's dishes online. We like it, but we don't know whether enough people would take part, so we'd want to run it for a month in one canteen before deciding. "
                "Maya: Discounted leftovers after seven in the evening was another suggestion. Unfortunately the manager told us that food safety rules don't allow cooked food to be kept and sold later, so that one's out. "
                "Tom: Posters showing how much food is wasted each week were popular in our survey, and they're cheap, so we'll recommend putting those up immediately. "
                "Maya: And finally, a composting scheme. It sounds good, but we're not sure how much of the waste is actually suitable for composting, so we'd like to try it with the kitchen waste only, for a trial period, and see what happens. "
                "Tutor: That all sounds very sensible."),
            groups=[
                dict(type_id='ielts-listening-mcq', heading='',
                     instructions='Choose the correct letter, A, B or C.',
                     questions=[
                         ('Why did the students choose the main canteen for their study?', ABC('It produces more waste than the halls.', 'The canteen manager asked for help.', 'They could use its records.'), 'B', 'asked if any students could help her'),
                         ('What was the main problem with the first survey?', ABC('Too few people responded.', 'It took too long to complete.', 'Some questions were unclear.'), 'C', 'some of the questions were confusing'),
                         ('What surprised the students about the waste data?', ABC('Lunch produced the most waste overall.', 'Breakfast produced the most waste per person.', 'Dinner produced very little waste.'), 'B', 'breakfast produced more waste per person'),
                         ('What does the tutor suggest the students should do?', ABC('shorten their full report', 'add a short summary with charts', 'present their findings to the kitchen staff'), 'B', 'a short summary with charts'),
                         ('What will the students do next?', ABC('interview the kitchen staff', 'weigh the waste themselves', 'ask the canteen for more figures'), 'B', 'weighing the waste from each meal ourselves'),
                     ]),
                dict(type_id='ielts-listening-matching', heading='Ideas for reducing waste',
                     instructions='What do the students decide about each of the following ideas? Choose the correct letter, A, B or C. A: introduce it immediately. B: try it out first. C: reject it.',
                     questions=[
                         ('smaller plates', DECISIONS, 'A', 'introducing them straight away'),
                         ('a daily menu vote', DECISIONS, 'B', 'run it for a month in one canteen'),
                         ('discounted leftovers in the evening', DECISIONS, 'C', "so that one's out"),
                         ('posters showing weekly waste', DECISIONS, 'A', 'putting those up immediately'),
                         ('a composting scheme', DECISIONS, 'B', 'for a trial period'),
                     ]),
            ],
        ),
        dict(
            setting='You will hear part of a lecture about soundscape ecology.',
            script=(
                "Lecturer: Today I'm going to talk about a relatively young field of research called soundscape ecology. Most ecologists study what they can see: they count plants or track animals. Soundscape ecologists study what they can hear. Their aim is to record and understand all of the sounds in a habitat, and to use those sounds as a way of measuring how healthy that habitat is. "
                "Soundscape ecologists usually divide the sounds they record into three groups. The first group, known as biophony, is made up of the sounds produced by living organisms: birds singing, insects calling, frogs croaking, even the clicking of shrimps. The second group, geophony, is the sound of the non-living environment, such as wind in the trees, running water and, on the coast, the waves. The third group is called anthrophony, and it covers the sounds that humans make. Road traffic is the obvious example, but in many quiet places the most common human sound is actually aircraft passing overhead. "
                "So how is this research done? Rather than sending people into the field to listen, researchers use small, weatherproof recorders which can be left in place for several months at a time. That produces an enormous amount of sound, far more than any person could listen to, so the recordings are analysed by software which picks out patterns, for example how busy the soundscape is at different times of day. "
                "Some of the most striking results have come from studies of coral reefs. A healthy reef is a noisy place, full of the crackle of shrimps and the calls of fish. A damaged reef, on the other hand, is much quieter. Researchers found that healthy reefs were consistently louder than damaged ones, and this difference matters, because young fish rely on sound to locate a reef when they are ready to settle. In one experiment, loudspeakers playing the sounds of a healthy reef were placed on damaged sections of reef, and the number of young fish arriving there almost doubled. "
                "Of course, the method has its limits. A recording can tell us that a species is present, and roughly how active it is, but it cannot tell us about the health of individual animals. For that, traditional observation is still needed. Nonetheless, sound gives us a cheap and continuous way of watching over habitats which are difficult or dangerous to visit, and I expect it to become a standard tool."),
            groups=[dict(
                type_id='ielts-listening-completion', heading='Soundscape ecology',
                instructions='Complete the notes below. Write ONE WORD ONLY for each answer.', limit=1,
                questions=[
                    ('Aim: to record and understand all the sounds in a ________', ['habitat'], 'all of the sounds in a habitat'),
                    ('Biophony: sounds produced by living ________', ['organisms'], 'sounds produced by living organisms'),
                    ('Geophony: e.g. wind, running water and, on the coast, ________', ['waves'], 'on the coast, the waves'),
                    ('Anthrophony: in many quiet places the most common human sound is ________', ['aircraft'], 'is actually aircraft'),
                    ('Recorders are left in place for several ________', ['months'], 'for several months at a time'),
                    ('Recordings are analysed by ________', ['software'], 'analysed by software'),
                    ('Healthy reefs are ________ than damaged ones', ['louder'], 'consistently louder than damaged ones'),
                    ('Young fish use sound to ________ a reef', ['locate'], 'to locate a reef'),
                    ('Playing reef sounds: the number of young fish almost ________', ['doubled'], 'almost doubled'),
                    ('Limit: recordings cannot show the ________ of individual animals', ['health'], 'about the health of individual animals'),
                ])],
        ),
    ],
)]


def check(test):
    total = 0
    for part_no, part in enumerate(test['parts'], 1):
        script = part['script']
        lowered = script.lower()
        last = -1
        count = 0
        for group in part['groups']:
            for q in group['questions']:
                count += 1
                evidence = q[-1]
                where = script.find(evidence)
                assert where >= 0, (test['key'], part_no, 'evidence not in script', evidence)
                assert where > last, (test['key'], part_no, 'question out of recording order', evidence)
                last = where
                if group['type_id'] == 'ielts-listening-completion':
                    prompt, answers, _ = q
                    assert '________' in prompt
                    for a in answers:
                        assert len(a.split()) <= group['limit'] or a.replace('.', '').replace(':', '').isdigit(), (a, group['limit'])
                    spoken = [a for a in answers if re.search(r'\b' + re.escape(a.lower()) + r'\b', lowered)]
                    numeric = any(re.fullmatch(r'[\d.:]+(st|nd|rd|th)?', a) for a in answers)
                    assert spoken or numeric, (test['key'], part_no, 'answer never heard', answers)
                else:
                    _, options, key, _ = q
                    assert key in {o['id'] for o in options}
        assert count == 10, (test['key'], part_no, count)
        total += count
    assert total == 40


for test in TESTS:
    check(test)


def lit(value):
    if value is None:
        return 'NULL'
    if isinstance(value, bool):
        return 'TRUE' if value else 'FALSE'
    if isinstance(value, int):
        return str(value)
    if isinstance(value, (list, dict)):
        return "'" + json.dumps(value, ensure_ascii=False).replace("'", "''") + "'::jsonb"
    return "'" + str(value).replace("'", "''") + "'"


TYPE_NAMES = {
    'ielts-listening-completion': 'Listening completion',
    'ielts-listening-mcq': 'Listening multiple choice',
    'ielts-listening-map': 'Listening map labelling',
    'ielts-listening-matching': 'Listening matching',
}

out = ['-- Generated by scripts/generate_ielts_listening_tests.py. Original practice content.',
       '-- Full IELTS Listening tests: four parts, forty one-mark questions each.',
       '-- Not official or recalled IELTS material. Replay preserves editorial changes.', '']
for test in TESTS:
    tid = f"lt-{test['key']}"
    out.append(f"INSERT INTO listening_tests (id, exam_version_id, title) VALUES ({lit(tid)}, {lit(VERSION)}, {lit(test['title'])}) ON CONFLICT (id) DO NOTHING;")
    number = 0
    for part_no, part in enumerate(test['parts'], 1):
        pid = f'{tid}-p{part_no}'
        out.append(f"INSERT INTO listening_parts (id, test_id, part_no, setting, script, reading_seconds) VALUES "
                   f"({lit(pid)}, {lit(tid)}, {part_no}, {lit(part['setting'])}, {lit(part['script'])}, 30) ON CONFLICT (id) DO NOTHING;")
        for position, group in enumerate(part['groups'], 1):
            gid = f'{pid}-g{position}'
            image = map_svg() if group.get('image') == 'map' else None
            out.append(f"INSERT INTO listening_question_groups (id, part_id, position, type_id, instructions, heading, image_url) VALUES "
                       f"({lit(gid)}, {lit(pid)}, {position}, {lit(group['type_id'])}, {lit(group['instructions'])}, {lit(group['heading'])}, {lit(image)}) ON CONFLICT (id) DO NOTHING;")
            rows = []
            for index, q in enumerate(group['questions'], 1):
                number += 1
                qid = f'{gid}-q{index}'
                if group['type_id'] == 'ielts-listening-completion':
                    prompt, answers, evidence = q
                    options, correct = [], answers
                    explanation = f'The recording says: "{evidence}".'
                else:
                    prompt, options, key, evidence = q
                    correct = [key]
                    explanation = f'The recording says: "{evidence}".'
                rows.append('(' + ', '.join([
                    lit(qid), lit(VERSION), lit('IELTS'), "ARRAY['IELTS']", lit('listening'), lit(group['type_id']),
                    lit(TYPE_NAMES[group['type_id']]), lit(f"{test['title']}, Question {number}"), lit(prompt),
                    lit(options) if options else "'[]'::jsonb", lit(correct), lit(explanation), '1', lit('medium'),
                    "ARRAY['IELTS Listening', 'Mock test', 'Original practice']", lit(gid), lit(index), 'TRUE', '0']) + ')')
            out.append('INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id, type_name, title, prompt, options, correct_answers, explanation, points, difficulty, tags, listening_group_id, group_position, is_published, time_limit_seconds) VALUES')
            out.append(',\n'.join(rows))
            out.append('ON CONFLICT (id) DO NOTHING;')
    assert number == 40

(ROOT / 'migrations/000075_ielts_listening_tests_content.up.sql').write_text('\n'.join(out) + '\n', encoding='utf-8', newline='\n')
(ROOT / 'migrations/000075_ielts_listening_tests_content.down.sql').write_text(
    '-- Preserve authored content and learner references on rollback.\nSELECT 1;\n', encoding='utf-8', newline='\n')
for test in TESTS:
    words = sum(len(p['script'].split()) for p in test['parts'])
    print(test['key'], 'parts:', len(test['parts']), 'questions: 40', 'script words:', words)
