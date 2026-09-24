"""IELTS Listening practice, batch 2: plan/map labelling drills.

Original Prepyo plans in the public IELTS Listening map-labelling format.
Every plan has the same shape, as a building plan in the test often does: the
entrance at the bottom (south), a main corridor running north from it, and a
second corridor crossing it half way. Rooms stand in four rows on each side;
two are named on the plan and six carry the letters A-F.

The directions are generated from the plan, not written by hand, so each
answer is correct by construction: every phrase used ("the first room on your
left", "opposite the café") picks out exactly one room, and `check` below
proves it before any SQL is written. Not official, recalled or copied content.
"""
from html import escape
from urllib.parse import quote

SIDES = ('W', 'E')      # west = left and east = right, walking in from the south
ROWS = (1, 2, 3, 4)     # 1 is nearest the entrance; the cross corridor runs between 2 and 3

# name, speaker label, landmarks {slot: name}, letters in slot order, targets [(letter, facility)]
VENUES = [
    ('Hillside Community Centre', 'Manager', {('W', 2): 'Café', ('E', 4): 'Toilets'},
     [('W', 1), ('E', 1), ('E', 2), ('W', 3), ('E', 3), ('W', 4)],
     [('A', 'the art room'), ('B', 'the reception desk'), ('C', 'the meeting room'), ('D', 'the children’s playroom'), ('F', 'the computer room')]),
    ('Brightwater College, Block C', 'Guide', {('E', 1): 'Lifts', ('W', 3): 'Library'},
     [('W', 1), ('W', 2), ('E', 2), ('E', 3), ('W', 4), ('E', 4)],
     [('A', 'the student advice centre'), ('B', 'the printing room'), ('C', 'the language lab'), ('E', 'the staff room'), ('F', 'the lecture theatre')]),
    ('Riverside Sports Centre', 'Attendant', {('W', 1): 'Reception', ('E', 3): 'Pool'},
     [('E', 1), ('W', 2), ('E', 2), ('W', 3), ('W', 4), ('E', 4)],
     [('A', 'the changing rooms'), ('B', 'the gym'), ('C', 'the first-aid room'), ('D', 'the squash courts'), ('E', 'the dance studio')]),
    ('Greenfield Hospital Outpatients', 'Volunteer', {('E', 2): 'Pharmacy', ('W', 4): 'X-ray'},
     [('W', 1), ('E', 1), ('W', 2), ('W', 3), ('E', 3), ('E', 4)],
     [('A', 'the waiting area'), ('B', 'the eye clinic'), ('C', 'the blood test room'), ('E', 'the physiotherapy unit'), ('F', 'the children’s clinic')]),
    ('Harbour Maritime Museum', 'Guide', {('W', 1): 'Shop', ('E', 4): 'Café'},
     [('E', 1), ('W', 2), ('E', 2), ('W', 3), ('E', 3), ('W', 4)],
     [('A', 'the cloakroom'), ('B', 'the shipwreck gallery'), ('C', 'the map room'), ('D', 'the lighthouse exhibition'), ('F', 'the education room')]),
    ('Castleford Central Library', 'Librarian', {('E', 1): 'Returns desk', ('W', 3): 'Stairs'},
     [('W', 1), ('W', 2), ('E', 2), ('E', 3), ('W', 4), ('E', 4)],
     [('A', 'the children’s section'), ('B', 'the newspaper room'), ('C', 'the study room'), ('D', 'the local history archive'), ('E', 'the computer suite')]),
    ('Northgate Business Centre', 'Receptionist', {('W', 2): 'Kitchen', ('E', 3): 'Lifts'},
     [('W', 1), ('E', 1), ('E', 2), ('W', 3), ('W', 4), ('E', 4)],
     [('A', 'the security office'), ('B', 'the post room'), ('C', 'the training room'), ('E', 'the boardroom'), ('F', 'the quiet room')]),
    ('Millbank Arts Centre', 'Coordinator', {('E', 2): 'Box office', ('W', 4): 'Theatre'},
     [('W', 1), ('E', 1), ('W', 2), ('W', 3), ('E', 3), ('E', 4)],
     [('A', 'the pottery studio'), ('B', 'the cloakroom'), ('C', 'the music room'), ('D', 'the photography darkroom'), ('F', 'the gallery')]),
    ('Lakeside Conference Centre', 'Organiser', {('W', 1): 'Registration', ('E', 3): 'Restaurant'},
     [('E', 1), ('W', 2), ('E', 2), ('W', 3), ('W', 4), ('E', 4)],
     [('A', 'the information desk'), ('B', 'the prayer room'), ('C', 'the main hall'), ('E', 'the exhibitors’ area'), ('F', 'the press office')]),
    ('Station Road Shopping Arcade', 'Security Guard', {('E', 1): 'Bank', ('W', 3): 'Food court'},
     [('W', 1), ('W', 2), ('E', 2), ('E', 3), ('W', 4), ('E', 4)],
     [('A', 'the key-cutting shop'), ('B', 'the pharmacy'), ('C', 'the bookshop'), ('D', 'the phone repair shop'), ('F', 'the customer service desk')]),
]

LEFT_RIGHT = {'W': 'left', 'E': 'right'}
COMPASS_SIDE = {'W': 'west', 'E': 'east'}


def position(slot, compass):
    """A noun phrase naming exactly one room by where it is."""
    side, row = slot
    if compass:
        corner_row = {1: 'south', 4: 'north'}.get(row)
        if corner_row:
            return f'the room in the {corner_row}-{COMPASS_SIDE[side]} corner'
        return f'the room on the {COMPASS_SIDE[side]} side just {"south" if row == 2 else "north"} of the cross corridor'
    lr = LEFT_RIGHT[side]
    return {1: f'the first room on your {lr} as you come in',
            2: f'the room on your {lr} just before the cross corridor',
            3: f'the room on your {lr} just after the cross corridor',
            4: f'the last room on your {lr}, at the far end'}[row]


def neighbours(slot, landmarks):
    """A landmark opposite or beside the slot, if there is one, as a phrase."""
    side, row = slot
    other = ('E' if side == 'W' else 'W', row)
    if other in landmarks:
        return f'It’s directly opposite the {landmarks[other].lower()}.'
    pair = {1: 2, 2: 1, 3: 4, 4: 3}[row]
    if (side, pair) in landmarks:
        closer = 'further from' if pair < row else 'nearer'
        return f'It’s next to the {landmarks[(side, pair)].lower()}, on the side {closer} the entrance.'
    return ''


def items():
    """Fifty drills: five targets on each of the ten plans."""
    out = []
    n = 0
    for venue, speaker, landmarks, letter_slots, targets in VENUES:
        slot_of = dict(zip('ABCDEF', letter_slots))
        for t, (letter, facility) in enumerate(targets):
            compass = (n % 3 == 2)
            slot = slot_of[letter]
            # A distractor: another lettered room, named as something else.
            other_letter, other_facility = targets[(t + 2) % len(targets)]
            decoy = slot_of[other_letter]
            script = (f"{speaker}: Welcome to {venue}. On your plan, the main entrance is at the bottom, "
                      f"and the main corridor runs straight up the middle, with a second corridor crossing it half way along. "
                      f"Some people get lost looking for {facility}. "
                      f"It isn't {position(decoy, compass)}; that's {other_facility}. "
                      f"{facility[0].upper() + facility[1:]} is {position(slot, compass)}. {neighbours(slot, landmarks)}").strip()
            out.append(dict(venue=venue, title=f'{venue}: {facility}', script=script,
                            facility=facility, key=letter, landmarks=landmarks, letter_slots=letter_slots,
                            compass=compass))
            n += 1
    return out


def check():
    """Proves that each position phrase names exactly one room."""
    all_slots = [(s, r) for s in SIDES for r in ROWS]
    for compass in (False, True):
        phrases = [position(s, compass) for s in all_slots]
        assert len(set(phrases)) == len(phrases), 'two rooms share a description'
    for venue, _, landmarks, letter_slots, targets in VENUES:
        slots = list(landmarks) + list(letter_slots)
        assert len(slots) == 8 and len(set(slots)) == 8, venue
        assert len(targets) == 5 and len({l for l, _ in targets}) == 5, venue


def svg(venue, landmarks, letter_slots):
    """The plan: two columns of rooms either side of the main corridor."""
    col_x = {'W': 60, 'E': 400}
    row_y = {4: 60, 3: 150, 2: 290, 1: 380}
    parts = ['<svg xmlns="http://www.w3.org/2000/svg" width="640" height="520" viewBox="0 0 640 520">',
             '<rect width="640" height="520" fill="#ffffff"/>',
             '<g font-family="Arial" fill="#172554">',
             f'<text x="20" y="30" font-size="18" font-weight="bold">{escape(venue)}</text>',
             '<text x="590" y="30" font-size="16">N &#8593;</text>',
             '<rect x="40" y="45" width="560" height="425" fill="none" stroke="#172554" stroke-width="3"/>',
             '<rect x="280" y="45" width="80" height="425" fill="#e2e8f0"/>',
             '<rect x="40" y="240" width="560" height="40" fill="#e2e8f0"/>',
             '<text x="300" y="265" font-size="12" fill="#475569">corridor</text>',
             '<rect x="285" y="470" width="70" height="14" fill="#172554"/>',
             '<text x="268" y="505" font-size="14">Main entrance</text>']
    letters = dict(zip(letter_slots, 'ABCDEF'))
    for (side, row) in [(s, r) for s in SIDES for r in ROWS]:
        x, y = col_x[side], row_y[row]
        parts.append(f'<rect x="{x}" y="{y}" width="180" height="75" fill="#dbeafe" stroke="#1e3a8a"/>')
        if (side, row) in landmarks:
            parts.append(f'<text x="{x + 90}" y="{y + 43}" font-size="15" text-anchor="middle">{escape(landmarks[(side, row)])}</text>')
        else:
            parts.append(f'<text x="{x + 90}" y="{y + 46}" font-size="24" font-weight="bold" text-anchor="middle">{letters[(side, row)]}</text>')
    parts.append('</g></svg>')
    return 'data:image/svg+xml,' + quote(''.join(parts), safe='')
