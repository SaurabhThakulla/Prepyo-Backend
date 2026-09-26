"""Generate IELTS Writing Task 1 bank 4 (migration 000093).

  50 Academic Task 1 figures   ielts-b4-fig-001..050     ielts-writing-task1-figure
  50 General Training letters  ielts-b4-letter-001..050  ielts-writing-task1-letter

Original Prepyo practice content; none of it is official, recalled or copied
test material. The figures use real countries and cities (or plain
descriptions such as "a village in the south-west of England") with invented
but plausible data. The content lives in scripts/ielts_b4_writing_task1; this
script checks it and writes SQL. Rows are modelled on the Task 1 figures of
scripts/generate_ielts_bank2.py (migration 000083) and the letters of
scripts/generate_ielts_general_training.py (migration 000071); the chart
renderers are copied from the bank 2 generator (importing it would run it and
rewrite its migrations), with small layout fixes: long labels wrap, up to three
pies fit side by side, axes pick a tighter scale, and two charts can be shown
together. Nothing is written if a check fails:
  - every item passes scripts/bank_check.py (no repeats of the bank or of each
    other, no machine-sounding or American wording, no invented place names);
    the repeat check compares each item's content (its figure_data, or a
    letter's situation and bullet points), because the fixed exam instructions
    are the same on every item; the whole prompt goes through the style check;
  - pie charts total 100, chart series match their axes, tables are
    rectangular, map areas stay on the grid, every label fits its box;
  - every image is well-formed SVG;
  - exactly fifty figures and fifty letters, with the planned mix of kinds
    and tones.

Output is deterministic and replay-safe (ON CONFLICT DO NOTHING).
Run from any directory: python scripts/generate_ielts_writing_task1_b4.py
"""
import json
import math
import sys
import xml.etree.ElementTree as ET
from collections import Counter
from html import escape
from pathlib import Path
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
sys.path.insert(0, str(Path(__file__).resolve().parent))

from scripts.bank_check import Batch                                      # noqa: E402
from ielts_b4_writing_task1.figures import FIGURES                        # noqa: E402
from ielts_b4_writing_task1.letters import LETTERS                        # noqa: E402

VERSION = 'ielts-2026-01'
MIGRATION = '000093_ielts_writing_task1_bank_4'
PALETTE = ['#1d4ed8', '#f59e0b', '#059669', '#dc2626', '#7c3aed', '#0891b2']


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


def text_array(values):
    return 'ARRAY[' + ', '.join(lit(v) for v in values) + ']'


def data_uri(svg):
    return 'data:image/svg+xml,' + quote(svg, safe='')


# ------------------------------------------------------------------ charts

def nice_scale(value):
    """Smallest round top of the axis at or above value, with 4-6 intervals."""
    best = None
    for mag in (0.1, 1, 10, 100, 1000, 10000):
        for base in (1, 2, 2.5, 5):
            step = base * mag
            for n in (4, 5, 6):
                if step * n >= value and (best is None or step * n < best[0]):
                    best = (step * n, step, n)
    return best


def fmt(v):
    return f'{v:,.0f}' if float(v).is_integer() else f'{v:,.1f}'


def wrap(text, width):
    words, lines, line = text.split(), [], ''
    for word in words:
        if len(line) + len(word) + 1 > width and line:
            lines.append(line)
            line = word
        else:
            line = (line + ' ' + word).strip()
    lines.append(line)
    return lines


def text_width(text, size):
    """Rough rendered width of Arial text: average glyph about 0.56 em."""
    return len(text) * size * 0.56


def fitted(text, box_w, size, max_lines, where):
    """Wrap text to a box; fail if it would be cut off or spill out."""
    lines = wrap(text, max(4, int(box_w / (size * 0.56))))
    assert len(lines) <= max_lines, (where, 'label needs too many lines', text, lines)
    for line in lines:
        assert text_width(line, size) <= box_w + 2, (where, 'label too wide', text, line)
    return lines


def svg_frame(title, height=470, width=760, size=18):
    return [f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">',
            f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
            '<g font-family="Arial" fill="#172554">',
            f'<text x="{width / 2}" y="28" font-size="{size}" font-weight="bold" text-anchor="middle">{escape(title)}</text>']


def legend(parts, names, x, y, where):
    for i, name in enumerate(names):
        parts.append(f'<rect x="{x}" y="{y}" width="14" height="14" fill="{PALETTE[i % len(PALETTE)]}"/>')
        for k, line in enumerate(fitted(name, 760 - x - 26, 13, 3, where)):
            parts.append(f'<text x="{x + 20}" y="{y + 12 + k * 15}" font-size="13">{escape(line)}</text>')
            last = k
        y += 20 + last * 15
    return y


def axes(parts, scale, unit, x0=90, y0=60, w=520, h=310):
    top, step, n = scale
    for i in range(n + 1):
        y = y0 + h - h * i / n
        parts.append(f'<line x1="{x0}" y1="{y:.1f}" x2="{x0 + w}" y2="{y:.1f}" stroke="#e2e8f0"/>')
        parts.append(f'<text x="{x0 - 8}" y="{y + 4:.1f}" font-size="12" text-anchor="end">{fmt(step * i)}</text>')
    parts.append(f'<line x1="{x0}" y1="{y0}" x2="{x0}" y2="{y0 + h}" stroke="#172554"/>')
    parts.append(f'<line x1="{x0}" y1="{y0 + h}" x2="{x0 + w}" y2="{y0 + h}" stroke="#172554"/>')
    parts.append(f'<text x="22" y="{y0 + h / 2}" font-size="12" text-anchor="middle" transform="rotate(-90 22 {y0 + h / 2})">{escape(unit)}</text>')


def bar_svg(f, size=18):
    parts = svg_frame(f['title'], size=size)
    series = list(f['series'].items())
    scale = nice_scale(max(max(v) for _, v in series))
    top = scale[0]
    x0, y0, w, h = 90, 60, 520, 310
    axes(parts, scale, f['unit'])
    cats = f['categories']
    group_w = w / len(cats)
    bar_w = min(38, (group_w - 16) / len(series))
    for c, cat in enumerate(cats):
        gx = x0 + group_w * c + (group_w - bar_w * len(series)) / 2
        for s, (_, values) in enumerate(series):
            bh = h * values[c] / top
            parts.append(f'<rect x="{gx + s * bar_w:.1f}" y="{y0 + h - bh:.1f}" width="{bar_w - 2:.1f}" height="{bh:.1f}" fill="{PALETTE[s]}"/>')
        for k, line in enumerate(fitted(cat, group_w - 4, 12, 2, f['title'])):
            parts.append(f'<text x="{x0 + group_w * c + group_w / 2:.1f}" y="{y0 + h + 20 + k * 14}" font-size="12" text-anchor="middle">{escape(line)}</text>')
    legend(parts, [n for n, _ in series], 625, 70, f['title'])
    parts.append('</g></svg>')
    return ''.join(parts)


def line_svg(f, size=18):
    parts = svg_frame(f['title'], size=size)
    series = list(f['series'].items())
    scale = nice_scale(max(max(v) for _, v in series))
    top = scale[0]
    x0, y0, w, h = 90, 60, 520, 310
    axes(parts, scale, f['unit'])
    xs = f['x']
    step = w / (len(xs) - 1)
    for i, x in enumerate(xs):
        assert text_width(str(x), 12) <= step, (f['title'], 'x label too wide', x)
        parts.append(f'<text x="{x0 + step * i:.1f}" y="{y0 + h + 20}" font-size="12" text-anchor="middle">{escape(str(x))}</text>')
    for s, (_, values) in enumerate(series):
        pts = ' '.join(f'{x0 + step * i:.1f},{y0 + h - h * v / top:.1f}' for i, v in enumerate(values))
        parts.append(f'<polyline points="{pts}" fill="none" stroke="{PALETTE[s]}" stroke-width="3"/>')
        for i, v in enumerate(values):
            parts.append(f'<circle cx="{x0 + step * i:.1f}" cy="{y0 + h - h * v / top:.1f}" r="4" fill="{PALETTE[s]}"/>')
    legend(parts, [n for n, _ in series], 625, 70, f['title'])
    parts.append('</g></svg>')
    return ''.join(parts)


def pie_height(f):
    categories = list(dict.fromkeys(c for p in f['pies'].values() for c in p))
    return 350 + 22 * ((len(categories) + 2) // 3) + 10


def pie_svg(f, size=18):
    pies = list(f['pies'].items())
    assert 1 <= len(pies) <= 3, f['title']
    categories = list(dict.fromkeys(c for _, p in pies for c in p))
    colour = {c: PALETTE[i] if i < len(PALETTE) else '#94a3b8' for i, c in enumerate(categories)}
    width = 760
    parts = svg_frame(f['title'], height=pie_height(f), width=width, size=size)
    r = {1: 105, 2: 105, 3: 92}[len(pies)]
    centres = {1: [380], 2: [200, 560], 3: [130, 380, 630]}[len(pies)]
    for (label, pie), cx in zip(pies, centres):
        cy = 205
        parts.append(f'<text x="{cx}" y="76" font-size="15" font-weight="bold" text-anchor="middle">{escape(label)}</text>')
        angle = -math.pi / 2
        for cat, pct in pie.items():
            sweep = 2 * math.pi * pct / 100
            x1, y1 = cx + r * math.cos(angle), cy + r * math.sin(angle)
            x2, y2 = cx + r * math.cos(angle + sweep), cy + r * math.sin(angle + sweep)
            large = 1 if sweep > math.pi else 0
            parts.append(f'<path d="M{cx},{cy} L{x1:.1f},{y1:.1f} A{r},{r} 0 {large} 1 {x2:.1f},{y2:.1f} Z" fill="{colour[cat]}" stroke="#ffffff" stroke-width="2"/>')
            mid = angle + sweep / 2
            if pct >= 6:     # small slices are labelled just outside the pie
                parts.append(f'<text x="{cx + r * 0.65 * math.cos(mid):.1f}" y="{cy + r * 0.65 * math.sin(mid) + 4:.1f}" font-size="12" fill="#ffffff" font-weight="bold" text-anchor="middle">{pct}%</text>')
            else:
                parts.append(f'<text x="{cx + (r + 16) * math.cos(mid):.1f}" y="{cy + (r + 16) * math.sin(mid) + 4:.1f}" font-size="12" font-weight="bold" text-anchor="middle">{pct}%</text>')
            angle += sweep
    ly = 340
    for i, cat in enumerate(categories):
        x = 40 + (i % 3) * 240
        y = ly + (i // 3) * 22
        fitted(cat, 215, 13, 1, f['title'])
        parts.append(f'<rect x="{x}" y="{y}" width="14" height="14" fill="{colour[cat]}"/>')
        parts.append(f'<text x="{x + 20}" y="{y + 12}" font-size="13">{escape(cat)}</text>')
    parts.append('</g></svg>')
    return ''.join(parts)


def table_height(f):
    return 70 + 40 * (len(f['rows']) + 1) + 20


def table_svg(f, size=18):
    cols = len(f['header'])
    rows = [f['header']] + f['rows']
    col_w = 700 / cols
    parts = svg_frame(f['title'], height=table_height(f), size=size)
    for r, row in enumerate(rows):
        y = 50 + 40 * r
        fill = '#dbeafe' if r == 0 else ('#f8fafc' if r % 2 else '#ffffff')
        for c, cell in enumerate(row):
            x = 30 + col_w * c
            parts.append(f'<rect x="{x:.1f}" y="{y}" width="{col_w:.1f}" height="40" fill="{fill}" stroke="#94a3b8"/>')
            weight = 'bold' if r == 0 or c == 0 else 'normal'
            font = 12 if r == 0 else 14
            lines = fitted(cell, col_w - 8, font, 2, f['title'])
            if len(lines) > 1 and r:
                font = 12
            for k, line in enumerate(lines):
                dy = 25 + (k - (len(lines) - 1) / 2) * 13
                parts.append(f'<text x="{x + col_w / 2:.1f}" y="{y + dy:.1f}" font-size="{font}" font-weight="{weight}" text-anchor="middle">{escape(line)}</text>')
    parts.append('</g></svg>')
    return ''.join(parts)


def process_svg(f, size=18):
    steps = f['steps']
    per_row = 4
    rows = (len(steps) + per_row - 1) // per_row
    height = 80 + rows * 150
    parts = svg_frame(f['title'], height=height, size=size)
    parts.append('<defs><marker id="a" markerWidth="10" markerHeight="10" refX="8" refY="5" orient="auto"><path d="M0,0 L10,5 L0,10 z" fill="#1e3a8a"/></marker></defs>')
    box_w, box_h, gap = 150, 90, 38
    for i, step in enumerate(steps):
        row, col = divmod(i, per_row)
        if row % 2:
            col = per_row - 1 - col          # snake so each arrow is short
        x = 30 + col * (box_w + gap)
        y = 60 + row * 150
        parts.append(f'<rect x="{x}" y="{y}" width="{box_w}" height="{box_h}" rx="10" fill="#eff6ff" stroke="#1e3a8a"/>')
        parts.append(f'<circle cx="{x + 16}" cy="{y + 16}" r="11" fill="#1e3a8a"/><text x="{x + 16}" y="{y + 20}" font-size="12" fill="#ffffff" text-anchor="middle">{i + 1}</text>')
        lines = wrap(step, 20)
        assert len(lines) <= 4, (f['title'], 'stage text too long', step)
        for k, line in enumerate(lines):
            parts.append(f'<text x="{x + box_w / 2}" y="{y + 36 + k * 15}" font-size="12" text-anchor="middle">{escape(line)}</text>')
        if i + 1 < len(steps):
            nrow, ncol = divmod(i + 1, per_row)
            if nrow % 2:
                ncol = per_row - 1 - ncol
            if nrow == row:
                if ncol > col:
                    parts.append(f'<line x1="{x + box_w}" y1="{y + box_h / 2}" x2="{x + box_w + gap - 4}" y2="{y + box_h / 2}" stroke="#1e3a8a" stroke-width="2" marker-end="url(#a)"/>')
                else:
                    parts.append(f'<line x1="{x}" y1="{y + box_h / 2}" x2="{x - gap + 4}" y2="{y + box_h / 2}" stroke="#1e3a8a" stroke-width="2" marker-end="url(#a)"/>')
            else:
                parts.append(f'<line x1="{x + box_w / 2}" y1="{y + box_h}" x2="{x + box_w / 2}" y2="{y + 146}" stroke="#1e3a8a" stroke-width="2" marker-end="url(#a)"/>')
    parts.append('</g></svg>')
    return ''.join(parts)


MAP_FILL = {'water': '#bfdbfe', 'sand': '#fde68a', 'building': '#cbd5e1', 'road': '#e5e7eb', 'green': '#bbf7d0'}


def map_svg(f, size=18):
    panels = list(f['panels'].items())
    assert len(panels) == 2, f['title']
    cell = 32
    parts = svg_frame(f['title'], height=380, size=size)
    for p, (label, areas) in enumerate(panels):
        ox, oy = 30 + p * 370, 70
        parts.append(f'<text x="{ox + 160}" y="{oy - 12}" font-size="15" font-weight="bold" text-anchor="middle">{escape(label)}</text>')
        parts.append(f'<rect x="{ox}" y="{oy}" width="{10 * cell}" height="{8 * cell}" fill="#f8fafc" stroke="#172554"/>')
        for name, x, y, w, h, kind in areas:
            assert 0 <= x and x + w <= 10 and 0 <= y and y + h <= 8, (f['title'], label, name, 'off the grid')
            parts.append(f'<rect x="{ox + x * cell}" y="{oy + y * cell}" width="{w * cell}" height="{h * cell}" fill="{MAP_FILL[kind]}" stroke="#475569"/>')
            lines = fitted(name, w * cell - 4, 11, max(2, int(h * cell / 13)), f['title'])
            for k, line in enumerate(lines):
                parts.append(f'<text x="{ox + (x + w / 2) * cell}" y="{oy + (y + h / 2) * cell + 4 + (k - (len(lines) - 1) / 2) * 13:.1f}" font-size="11" text-anchor="middle">{escape(line)}</text>')
        parts.append(f'<text x="{ox + 10 * cell - 14}" y="{oy + 16}" font-size="13">N&#8593;</text>')
    parts.append('</g></svg>')
    return ''.join(parts)


def part_height(part):
    return {'bar': lambda: 470, 'line': lambda: 470, 'pie': lambda: pie_height(part),
            'table': lambda: table_height(part)}[part['kind']]()


def mixed_svg(f, size=18):
    """Two charts one above the other under the figure's title."""
    height = 50 + sum(part_height(p) for p in f['parts'])
    parts = svg_frame(f['title'], height=height, size=size)
    y = 44
    for part in f['parts']:
        inner = RENDER[part['kind']](part, size=15)
        parts.append(inner.replace('<svg ', f'<svg x="0" y="{y}" ', 1))
        y += part_height(part)
    parts.append('</g></svg>')
    return ''.join(parts)


RENDER = {'bar': bar_svg, 'line': line_svg, 'pie': pie_svg, 'table': table_svg, 'process': process_svg,
          'map': map_svg, 'mixed': mixed_svg}
LEAD = {'bar': 'The bar chart below shows', 'line': 'The graph below shows', 'table': 'The table below gives',
        'process': 'The diagram below shows', 'map': 'The maps below show'}
KIND_NAME = {'bar': 'bar chart', 'line': 'line graph', 'pie': 'pie chart', 'table': 'table'}


def data_lines(f):
    lines, kind = [], f['kind']
    if kind in ('bar', 'line'):
        axis = f['categories'] if kind == 'bar' else f['x']
        lines.append(f"Unit: {f['unit']}")
        lines.append(' | '.join(['Series'] + [str(a) for a in axis]))
        for name, values in f['series'].items():
            lines.append(' | '.join([name] + [fmt(v) for v in values]))
    elif kind == 'pie':
        for label, pie in f['pies'].items():
            lines.append(f'{label}: ' + ', '.join(f'{c} {p}%' for c, p in pie.items()))
    elif kind == 'table':
        for row in [f['header']] + f['rows']:
            lines.append(' | '.join(row))
    elif kind == 'process':
        lines += [f'Stage {i + 1}: {s}' for i, s in enumerate(f['steps'])]
    elif kind == 'map':
        for label, areas in f['panels'].items():
            lines.append(f'{label}: ' + '; '.join(a[0] for a in areas))
    elif kind == 'mixed':
        for n, part in enumerate(f['parts'], 1):
            lines.append(f"Chart {n} ({KIND_NAME[part['kind']]}): {part['title']}")
            lines += data_lines(part)
    return lines


def figure_data(f):
    lines = [f['title'], f['desc'][0].upper() + f['desc'][1:] + '.']
    lines += data_lines(f)
    lines.append('Main features: ' + f['features'])
    return '\n'.join(lines)


def check_chart(f, where):
    kind = f['kind']
    if kind == 'pie':
        for label, pie in f['pies'].items():
            assert sum(pie.values()) == 100, (where, label, sum(pie.values()))
    if kind in ('bar', 'line'):
        n = len(f['categories'] if kind == 'bar' else f['x'])
        assert n >= 2 and all(len(v) == n for v in f['series'].values()), where
        assert all(v >= 0 for vs in f['series'].values() for v in vs), (where, 'negative value')
        assert len(f['series']) <= len(PALETTE), where
    if kind == 'table':
        assert all(len(r) == len(f['header']) for r in f['rows']), where
    if kind == 'mixed':
        assert len(f['parts']) == 2 and f.get('lead'), where
        for part in f['parts']:
            assert part['kind'] in KIND_NAME, where
            check_chart(part, where)
    if kind == 'map':
        before, after = f['panels'].values()
        assert [a[0] for a in before] != [a[0] for a in after], (where, 'maps are identical')
    if kind == 'process':
        assert 6 <= len(f['steps']) <= 12, where


# ------------------------------------------------------------ rows

batch = Batch()
figures, letters = [], []

assert len(FIGURES) == 50, len(FIGURES)
for i, f in enumerate(FIGURES, 1):
    fid = f'ielts-b4-fig-{i:03}'
    check_chart(f, fid)
    lead = f.get('lead') or LEAD.get(f['kind'], '')
    if f['kind'] == 'pie' and not f.get('lead'):
        lead = 'The pie charts below show' if len(f['pies']) > 1 else 'The pie chart below shows'
    if f['kind'] == 'table' and not f.get('lead'):
        lead = 'The table below gives' if f['desc'].startswith('information') else 'The table below shows'
    svg = RENDER[f['kind']](f)
    root = ET.fromstring(svg)                      # well-formed XML or this raises
    assert root.tag == '{http://www.w3.org/2000/svg}svg', fid
    prompt = (f"{lead} {f['desc']}.\n\nSummarise the information by selecting and reporting the main features, "
              f"and make comparisons where relevant.\n\nWrite at least 150 words. You should spend about 20 minutes on this task.")
    data = figure_data(f)
    # The fixed exam wording ("The bar chart below shows...", "Summarise the
    # information...") is the same on every figure, so only the content is
    # compared: figure_data holds the description, the data and the features.
    batch.question('IELTS', 'ielts-writing-task1-figure', f['title'], data, where=fid)
    batch.style(prompt, where=fid)
    batch.style(f['features'], where=fid)
    figures.append(dict(
        id=fid, exam_version_id=VERSION, exam='IELTS', supported_exams=['IELTS'], skill='writing',
        type_id='ielts-writing-task1-figure', type_name='Describe the figure', title=f['title'], prompt=prompt,
        time_limit_seconds=1200, prep_time_seconds=0, points=20,
        difficulty='hard' if f['kind'] in ('process', 'map') else 'medium',
        tags=['Original practice', 'IELTS', 'Writing'], image_url=data_uri(svg), figure_data=data,
        explanation=f['features']))

LETTER_TAIL = ('\n\nWrite at least 150 words.\n\nYou do NOT need to write any addresses.'
               '\n\nBegin your letter as follows:\n\n{opening}')
TONE_GUIDANCE = {
    'formal': 'Tone: formal, to someone you do not know. Open with the salutation given, state your purpose in the first paragraph, and close with "Yours faithfully" and your full name.',
    'semi-formal': 'Tone: semi-formal, polite but personal, to someone you know in a formal relationship. Use their title and surname, state your purpose early, and close with "Yours sincerely" and your full name.',
    'informal': 'Tone: informal and friendly, to someone you know well. Use their first name, write naturally, and close with an informal phrase such as "Best wishes" and your first name.',
}
OPENING = {'formal': 'Dear Sir or Madam,', 'semi-formal': 'Dear ..........,', 'informal': 'Dear ..........,'}

assert len(LETTERS) == 50, len(LETTERS)
for i, (tone, title, situation, recipient, bullets, opening) in enumerate(LETTERS, 1):
    lid = f'ielts-b4-letter-{i:03}'
    assert len(bullets) == 3 and opening == OPENING[tone], lid
    prompt = (f'You should spend about 20 minutes on this task.\n\n{situation}\n\n'
              f'Write a letter to {recipient}. In your letter\n\n'
              + '\n'.join(f'- {b}' for b in bullets)
              + LETTER_TAIL.format(opening=opening))
    explanation = (TONE_GUIDANCE[tone] + ' Cover all three bullet points and develop each one with relevant detail; '
                   'a missing bullet point limits Task Achievement. Make the purpose of the letter clear, '
                   'organise it into paragraphs, and keep the tone consistent throughout. There is no single correct letter.')
    # As for figures, the fixed instructions ("You should spend about 20
    # minutes...", "Write a letter to...", "Write at least 150 words...") are
    # left out of the repeat check; the situation and bullet points are compared.
    batch.question('IELTS', 'ielts-writing-task1-letter', title, situation + '\n' + '\n'.join(bullets), where=lid)
    batch.style(prompt, where=lid)
    batch.style(explanation, where=lid)
    letters.append(dict(
        id=lid, exam_version_id=VERSION, exam='IELTS', supported_exams=['IELTS'], skill='writing',
        type_id='ielts-writing-task1-letter', type_name='Write a letter', title=title, prompt=prompt,
        explanation=explanation, prep_time_seconds=0, time_limit_seconds=1200, points=20, difficulty='medium',
        tags=['Original practice', 'IELTS', 'Writing', 'Task 1', 'General Training', tone.capitalize()]))

batch.assert_clean()

kinds = Counter(f['kind'] for f in FIGURES)
tones = Counter(t[0] for t in LETTERS)
assert kinds == Counter(bar=10, line=10, pie=7, table=7, mixed=4, map=6, process=6), kinds
assert tones == Counter({'formal': 17, 'semi-formal': 17, 'informal': 16}), tones
rows = figures + letters
assert len(rows) == 100 and len({r['id'] for r in rows}) == 100
assert len({r['title'].lower() for r in rows}) == 100

# ------------------------------------------------------------ SQL


def insert(rows):
    columns = list(rows[0])
    out = ['INSERT INTO questions (' + ', '.join(columns) + ') VALUES']
    out.append(',\n'.join('(' + ', '.join(text_array(r[c]) if c in ('supported_exams', 'tags') else lit(r[c]) for c in columns) + ')'
                          for r in rows))
    out.append('ON CONFLICT (id) DO NOTHING;')
    return out


body = ['-- Generated by scripts/generate_ielts_writing_task1_b4.py. Original practice content.',
        '-- IELTS Writing Task 1 bank 4: fifty Academic figures and fifty General Training letters.',
        '-- Not official or recalled IELTS material; figure data is invented for practice.',
        '']
body += insert(figures)
body.append('')
body += insert(letters)

KEEP = '-- Preserve authored content and learner references on rollback.\nSELECT 1;\n'
(ROOT / f'migrations/{MIGRATION}.up.sql').write_text('\n'.join(body) + '\n', encoding='utf-8', newline='\n')
(ROOT / f'migrations/{MIGRATION}.down.sql').write_text(KEEP, encoding='utf-8', newline='\n')

print('figures:', len(figures), dict(kinds))
print('letters:', len(letters), dict(tones))
