"""IELTS Listening practice, batch 4: plan labelling drills.

Original Prepyo plans in the public IELTS Listening map-labelling format, drawn
like the batch 2 plans: the main entrance at the bottom (south), a main
corridor running north from it, and a second corridor crossing it half way.
Rooms stand in four rows on each side; two are named on the plan (landmarks)
and six carry the letters A-F.

The scripts are written by hand, so each item lists the direction that gives
the answer and the directions that point at other rooms (the distractors).
`check` reads every one of those directions against the plan and proves that
it picks out exactly one room: the answer for the answer, and a different
lettered room for each distractor. Not official, recalled or copied content.
"""
import re

SIDES = ('W', 'E')      # west = left and east = right, walking in from the south
ROWS = (1, 2, 3, 4)     # 1 is nearest the entrance; the cross corridor runs between 2 and 3
ALL_SLOTS = [(s, r) for s in SIDES for r in ROWS]
OTHER_SIDE = {'W': 'E', 'E': 'W'}
PAIR = {1: 2, 2: 1, 3: 4, 4: 3}    # the room next door on the same side of the cross corridor

# Words that fix the side or the row of a room, as a candidate set of slots.
SIDE_WORDS = [(r'\b(?:left|west)\b', 'W'), (r'\b(?:right|east)\b', 'E')]
ROW_WORDS = [
    (r'\bfirst room\b|\bas you come in\b|\bnearest the entrance\b|\bjust inside the entrance\b', 1),
    (r'\bsecond room\b|\bbefore (?:you (?:get to|reach) )?the cross corridor\b|\bjust south of the cross corridor\b', 2),
    (r'\bthird room\b|\b(?:after|past|beyond) the cross corridor\b|\bjust north of the cross corridor\b|\bonce you(?:\'ve| have) crossed\b', 3),
    (r'\blast room\b|\bfourth room\b|\bfar end\b|\bat the (?:very )?top\b', 4),
]
CORNERS = {'south-west': ('W', 1), 'south-east': ('E', 1), 'north-west': ('W', 4), 'north-east': ('E', 4)}


def V(venue, speaker, landmarks, rooms, items):
    return dict(venue=venue, speaker=speaker, landmarks=landmarks, rooms=rooms, items=items)


def names(plan):
    """Every named place on the plan, lower case without 'the', to its slot."""
    out = {name.lower(): slot for slot, name in plan['landmarks'].items()}
    out.update({re.sub(r'^the ', '', name.lower()): slot for slot, name in plan['rooms'].values()})
    return out


def resolve(phrase, plan):
    """The rooms a direction can mean. A good direction means exactly one."""
    low = phrase.lower().replace('’', "'")
    found = set(ALL_SLOTS)
    used = False
    for pattern, side in SIDE_WORDS:
        if re.search(pattern, re.sub(r'(?:opposite|next to|beside) (?:the )?[\w\' -]+', '', low)):
            found &= {s for s in ALL_SLOTS if s[0] == side}
            used = True
    for pattern, row in ROW_WORDS:
        if re.search(pattern, low):
            found &= {s for s in ALL_SLOTS if s[1] == row}
            used = True
    for corner, slot in CORNERS.items():
        if corner + ' corner' in low:
            found &= {slot}
            used = True
    for name, (side, row) in names(plan).items():
        if re.search(r'\bopposite (?:the )?' + re.escape(name) + r'\b', low):
            found &= {(OTHER_SIDE[side], row)}
            used = True
        if re.search(r'\b(?:next to|beside|next door to) (?:the )?' + re.escape(name) + r'\b', low):
            found &= {(side, PAIR[row])}
            used = True
    assert used, ('direction has no position in it', phrase)
    return found


def check(plans):
    """Proves every answer and distractor direction names exactly one room."""
    for plan in plans:
        venue = plan['venue']
        slots = list(plan['landmarks']) + [slot for slot, _ in plan['rooms'].values()]
        assert len(plan['landmarks']) == 2 and sorted(plan['rooms']) == list('ABCDEF'), venue
        assert len(slots) == 8 and set(slots) == set(ALL_SLOTS), venue
        for name in names(plan):
            assert not re.search(r'\b(?:left|right|east|west|north|south)\b', name), (venue, name)
        letter_at = {slot: letter for letter, (slot, _) in plan['rooms'].items()}
        for key, script, answer, decoys in plan['items']:
            facility = plan['rooms'][key][1]
            assert script.startswith(plan['speaker'] + ':'), (venue, key)
            assert facility.lower() in script.lower(), (venue, key, 'facility never named')
            assert answer in script, (venue, key, 'answer direction not in script', answer)
            assert resolve(answer, plan) == {plan['rooms'][key][0]}, (venue, key, answer, resolve(answer, plan))
            assert decoys, (venue, key, 'no distractor')
            for phrase, letter in decoys:
                assert phrase in script, (venue, key, 'distractor not in script', phrase)
                assert letter != key, (venue, key, phrase)
                assert resolve(phrase, plan) == {plan['rooms'][letter][0]}, (venue, key, phrase, resolve(phrase, plan))
                assert letter_at[plan['rooms'][letter][0]] == letter
            # The answer is the last direction given, as the speaker settles on it.
            assert script.rfind(answer) > max(script.find(p) for p, _ in decoys), (venue, key, 'answer before distractor')


def items(plans):
    """Two drills on each plan."""
    out = []
    for plan in plans:
        letter_slots = [plan['rooms'][letter][0] for letter in 'ABCDEF']
        for key, script, answer, _ in plan['items']:
            facility = plan['rooms'][key][1]
            out.append(dict(venue=plan['venue'], title=f"{plan['venue']}: {facility}", script=script,
                            facility=facility, key=key, landmarks=plan['landmarks'],
                            letter_slots=letter_slots, evidence=answer))
    return out


W, E = 'W', 'E'

PLANS = [
    V('Castle Street Leisure Centre', 'Receptionist',
      {(W, 1): 'Reception', (E, 3): 'Pool'},
      {'A': ((E, 1), 'the changing rooms'), 'B': ((W, 2), 'the gym'), 'C': ((E, 2), 'the café'),
       'D': ((W, 3), 'the sports hall'), 'E': ((W, 4), 'the squash courts'), 'F': ((E, 4), 'the sauna')},
      [('D', "Receptionist: Right, have you got the plan? We're here at reception, just inside the main entrance. "
             "Now, a lot of people looking for the sports hall go into the gym by mistake, because it's got the same kind of doors. "
             "The gym is the room on your left before the cross corridor, so don't go in there. Keep walking, cross over the corridor, "
             "and the sports hall is the room on your left just after the cross corridor, opposite the pool. The five-a-side starts at seven.",
        "the room on your left just after the cross corridor, opposite the pool",
        [("The gym is the room on your left before the cross corridor", 'B')]),
       ('F', "Receptionist: The sauna? It's right up at the top of the building. You go past the café, that's the second room on your right, "
             "and then past the pool. Some people think the sauna's inside the changing rooms, but those are the first room on your right as you come in. "
             "The sauna is the last room on your right, at the far end, next to the pool. It's closed between two and three for cleaning, by the way.",
        "The sauna is the last room on your right, at the far end, next to the pool",
        [("the café, that's the second room on your right", 'C'), ("the first room on your right as you come in", 'A')])]),

    V('Tarrant Road Surgery', 'Receptionist',
      {(W, 2): 'Waiting room', (E, 4): 'Stairs'},
      {'A': ((W, 1), 'the baby clinic'), 'B': ((E, 1), 'the pharmacy'), 'C': ((E, 2), "the nurses' room"),
       'D': ((W, 3), 'the treatment room'), 'E': ((E, 3), 'the counselling room'), 'F': ((W, 4), 'the minor injuries unit')},
      [('C', "Receptionist: You're seeing the nurse, so you don't need to sit in the main waiting room. That's for the doctors' patients. "
             "Go past the pharmacy, which is the first room on your right as you come in, and the nurses' room is the next one along. "
             "So it's the room on your right just before the cross corridor, opposite the waiting room. Just knock and take a seat outside.",
        "the room on your right just before the cross corridor, opposite the waiting room",
        [("the pharmacy, which is the first room on your right as you come in", 'B')]),
       ('F', "Receptionist: If it's a cut that needs stitches, you want the minor injuries unit. It used to be the room next to the waiting room, "
             "on the entrance side, but that's the baby clinic now. The unit moved last year. It's the last room on your left, at the far end, "
             "in the north-west corner of the building. If you get to the stairs, you've gone too far across.",
        "It's the last room on your left, at the far end, in the north-west corner",
        [("the room next to the waiting room, on the entrance side", 'A')])]),

    V("St Luke's Hospital, Ground Floor", 'Volunteer',
      {(E, 1): 'Main reception', (W, 3): 'Lifts'},
      {'A': ((W, 1), 'the coffee shop'), 'B': ((W, 2), 'the blood test clinic'), 'C': ((E, 2), 'the X-ray department'),
       'D': ((E, 3), 'the fracture clinic'), 'E': ((W, 4), 'the quiet room'), 'F': ((E, 4), 'the patient advice office')},
      [('D', "Volunteer: The fracture clinic? Yes, lots of people ask. You'll need to go past X-ray first, that's the room on your right before the cross corridor. "
             "Then over the corridor, and the fracture clinic is the first door you come to, on your right. So that's the room opposite the lifts. "
             "Give your name at the desk inside, and they'll call you.",
        "the room opposite the lifts",
        [("X-ray first, that's the room on your right before the cross corridor", 'C')]),
       ('B', "Volunteer: For a blood test, don't queue at main reception, just go straight to the clinic. Now, the first room on your left as you come in is the coffee shop, "
             "and people sometimes queue there by mistake, believe it or not. The blood test clinic is the room on your left just before the cross corridor. "
             "Take a ticket from the machine by the door.",
        "The blood test clinic is the room on your left just before the cross corridor",
        [("the first room on your left as you come in is the coffee shop", 'A')])]),

