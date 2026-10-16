"""Redraw the broken practice figures (migration 000105).

The image audit of 2026-10-16 found these figures broken in the bank:
  ielts-audit-001..005   Task 1 tables drawn as "a | b | c" text lines, with
                         "fictional" in the prompt; redrawn as ruled tables
  ielts-b2-wt1-013       legend cut off at the right edge
  ielts-b2-wt1-050       island map labels hidden under other areas
  ielts-b4-fig-035..038  two charts stacked 924px tall, shrunk to half size in
                         the 760x440 figure frame; now drawn side by side
  pte-b3-di-041..050     stage names spill out of their process boxes

The data comes from the original content modules (or, for the five tables,
from scripts/generate_ielts_practice.py, copied here because running that
script rewrites its migration). Each UPDATE runs only while the row still
holds the original value (md5 match), so anything an admin has replaced since
is left alone. Run from any directory: python scripts/fix_figure_images.py
"""
import math
import sys
import xml.etree.ElementTree as ET
from html import escape
from pathlib import Path
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
sys.path.insert(0, str(Path(__file__).resolve().parent))

from ielts_bank2.writing_task1 import FIGURES as B2                     # noqa: E402
from ielts_b4_writing_task1.figures import FIGURES as B4                # noqa: E402
from scripts.pte_speaking_b3.di import PROCESSES                        # noqa: E402

MIGRATION = '000105_fix_figure_images'
PALETTE = ['#1d4ed8', '#f59e0b', '#059669', '#dc2626', '#7c3aed', '#0891b2']
TASK1_TAIL = ('\n\nSummarise the information by selecting and reporting the main features, and make comparisons '
              'where relevant.\n\nWrite at least 150 words. You should spend about 20 minutes on this task.')


def data_uri(svg):
    ET.fromstring(svg)                                  # fail on malformed SVG
    return 'data:image/svg+xml,' + quote(svg, safe='')


def text_width(text, size, bold=False):
    """Rough rendered width of Arial text."""
    return len(text) * size * (0.62 if bold else 0.56)


def wrap(text, box_w, size, bold=False, max_lines=3, where=''):
    """Wrap text to a box width; fail if it cannot fit."""
    lines, line = [], ''
    for word in text.split():
        trial = (line + ' ' + word).strip()
        if line and text_width(trial, size, bold) > box_w:
            lines.append(line)
            line = word
        else:
            line = trial
    lines.append(line)
    assert len(lines) <= max_lines and all(text_width(l, size, bold) <= box_w for l in lines), (where, text, lines)
    return lines


def fmt(v):
    return f'{v:,.0f}' if float(v).is_integer() else f'{v:,.1f}'


def frame(title, width=760, height=440, size=18):
    return [f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">',
            f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
            '<g font-family="Arial" fill="#172554">',
            f'<text x="{width / 2}" y="30" font-size="{size}" font-weight="bold" text-anchor="middle">{escape(title)}</text>']


# ------------------------------------------------------------ Task 1 tables

TABLES = [
    ('ielts-audit-001', 'Visits to a city library by age group (thousands)',
     'the number of visits to a city library by people in three age groups in 2010 and 2020',
     ['Age group', '2010', '2020'], [['Under 18', '45', '62'], ['18–64', '110', '84'], ['65 and over', '28', '53']],
     ['Total', '183', '199']),
    ('ielts-audit-002', 'Household energy use by purpose (%)',
     'how households in one region used energy in 2005 and 2025',
     ['Purpose', '2005', '2025'], [['Heating', '55', '42'], ['Hot water', '20', '23'], ['Appliances', '15', '25'], ['Lighting', '10', '10']],
     ['Total', '100', '100']),
    ('ielts-audit-003', 'How commuters travel to work in Eastford (%)',
     'how commuters in the town of Eastford travelled to work in 2015 and 2025',
     ['Mode of transport', '2015', '2025'], [['Car', '60', '42'], ['Bus', '20', '25'], ['Bicycle', '8', '18'], ['Walking', '12', '15']],
     ['Total', '100', '100']),
    ('ielts-audit-004', 'Admissions to four museums (thousands)',
     'the number of admissions to four museums in one city in 2018 and 2023',
     ['Museum', '2018', '2023'], [['History', '120', '145'], ['Science', '95', '160'], ['Art', '140', '125'], ['Transport', '55', '80']],
     ['Total', '410', '510']),
    ('ielts-audit-005', 'Water consumption by sector (million cubic metres)',
     'the amount of water used by three sectors in one district in 2000 and 2020',
     ['Sector', '2000', '2020'], [['Agriculture', '80', '65'], ['Industry', '45', '30'], ['Households', '25', '40']],
     ['Total', '150', '135']),
]
# The examiner summaries are unchanged from migration 000052.
TABLE_FEATURES = {
    'ielts-audit-001': 'Total visits rose from 183 to 199 thousand. Working-age visits fell while both other groups increased.',
    'ielts-audit-002': 'Heating remained largest but fell 13 percentage points. Appliances rose 10 points and lighting was unchanged. Each column totals 100%.',
    'ielts-audit-003': 'Cars remained largest but lost share. All other modes gained; cycling more than doubled. Columns total 100%.',
    'ielts-audit-004': 'Science overtook art to lead. Art alone declined. Total admissions increased from 410 to 510 thousand.',
    'ielts-audit-005': 'Overall use fell from 150 to 135 million cubic metres. Agriculture remained largest. Household use rose as the other sectors fell.',
}


def table_svg(title, header, rows, total):
    cols = len(header)
    col_w = [220] + [150] * (cols - 1)
    table_w = sum(col_w)
    x0, y0, row_h = (760 - table_w) / 2, 60, 40
    body = [header] + rows + [total]
    height = y0 + row_h * len(body) + 30
    parts = frame(title, height=height)
    for r, row in enumerate(body):
        y = y0 + r * row_h
        head, last = r == 0, r == len(body) - 1
        fill = '#dbeafe' if head else ('#f1f5f9' if last else '#ffffff')
        x = x0
        for c, cell in enumerate(row):
            parts.append(f'<rect x="{x}" y="{y}" width="{col_w[c]}" height="{row_h}" fill="{fill}" stroke="#94a3b8"/>')
            bold = head or last or c == 0
            wrap(cell, col_w[c] - 16, 15, bold, 1, title)
            tx, anchor = (x + 14, 'start') if c == 0 else (x + col_w[c] / 2, 'middle')
            parts.append(f'<text x="{tx}" y="{y + 26}" font-size="15" font-weight="{"bold" if bold else "normal"}" text-anchor="{anchor}">{escape(cell)}</text>')
            x += col_w[c]
    parts.append(f'<line x1="{x0}" y1="{y0 + row_h * (len(body) - 1)}" x2="{x0 + table_w}" y2="{y0 + row_h * (len(body) - 1)}" stroke="#172554" stroke-width="2"/>')
    parts.append('</g></svg>')
    return ''.join(parts)


def table_data(title, desc, header, rows, total, features):
    lines = [title, desc[0].upper() + desc[1:] + '.'] + [' | '.join(r) for r in [header] + rows + [total]]
    return '\n'.join(lines + ['Main features: ' + features])


# ------------------------------------------------------- charts and panels

def nice_scale(value):
    """Round axis top with three to six steps."""
    for step in (0.5, 1, 2, 2.5, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000):
        n = math.ceil(value / step - 1e-9)
        if n <= 6:
            return step * max(n, 3), step, max(n, 3)
    raise ValueError(value)


def axes(parts, scale, unit, x0, y0, w, h, unit_x):
    top, step, n = scale
    for i in range(n + 1):
        y = y0 + h - h * i / n
        parts.append(f'<line x1="{x0}" y1="{y:.1f}" x2="{x0 + w}" y2="{y:.1f}" stroke="#e2e8f0"/>')
        parts.append(f'<text x="{x0 - 8}" y="{y + 4:.1f}" font-size="12" text-anchor="end">{fmt(step * i)}</text>')
    parts.append(f'<line x1="{x0}" y1="{y0}" x2="{x0}" y2="{y0 + h}" stroke="#172554"/>')
    parts.append(f'<line x1="{x0}" y1="{y0 + h}" x2="{x0 + w}" y2="{y0 + h}" stroke="#172554"/>')
    parts.append(f'<text x="{unit_x}" y="{y0 + h / 2}" font-size="12" text-anchor="middle" transform="rotate(-90 {unit_x} {y0 + h / 2})">{escape(unit)}</text>')


def side_legend(parts, names, x, y, box_w, where):
    """One entry per series, wrapped to fit the space right of the plot."""
    for i, name in enumerate(names):
        parts.append(f'<rect x="{x}" y="{y}" width="14" height="14" fill="{PALETTE[i]}"/>')
        lines = wrap(name, box_w - 20, 13, False, 3, where)
        for k, line in enumerate(lines):
            parts.append(f'<text x="{x + 20}" y="{y + 12 + k * 15}" font-size="13">{escape(line)}</text>')
        y += 22 + (len(lines) - 1) * 15


def row_legend(parts, names, x, y, box_w, where):
    """Entries side by side under a panel, starting a new row when full."""
    cx = x
    for i, name in enumerate(names):
        w = 20 + text_width(name, 12) + 14
        assert w <= box_w, (where, name)
        if cx + w > x + box_w:
            cx, y = x, y + 18
        parts.append(f'<rect x="{cx}" y="{y}" width="12" height="12" fill="{PALETTE[i]}"/>')
        parts.append(f'<text x="{cx + 17}" y="{y + 10}" font-size="12">{escape(name)}</text>')
        cx += w


def line_marks(parts, series, xs, scale, x0, y0, w, h):
    step = w / (len(xs) - 1)
    for s, values in enumerate(series):
        pts = ' '.join(f'{x0 + step * i:.1f},{y0 + h - h * v / scale[0]:.1f}' for i, v in enumerate(values))
        parts.append(f'<polyline points="{pts}" fill="none" stroke="{PALETTE[s]}" stroke-width="3"/>')
        for i, v in enumerate(values):
            parts.append(f'<circle cx="{x0 + step * i:.1f}" cy="{y0 + h - h * v / scale[0]:.1f}" r="4" fill="{PALETTE[s]}"/>')
    for i, x in enumerate(xs):
        parts.append(f'<text x="{x0 + step * i:.1f}" y="{y0 + h + 18}" font-size="12" text-anchor="middle">{escape(str(x))}</text>')


def line_svg(f):
    """Whole-canvas line graph, as in migration 000083, with a legend that wraps."""
    parts = frame(f['title'], height=470)
    series = list(f['series'].items())
    scale = nice_scale(max(max(v) for _, v in series))
    x0, y0, w, h = 90, 60, 500, 330
    axes(parts, scale, f['unit'], x0, y0, w, h, 22)
    line_marks(parts, [v for _, v in series], f['x'], scale, x0, y0, w, h)
    side_legend(parts, [n for n, _ in series], 610, 70, 140, f['title'])
    parts.append('</g></svg>')
    return ''.join(parts)


PANEL_Y, PANEL_H = 92, 220          # plot area of a half-width panel


def bar_panel(parts, p, ox, pw):
    series = list(p['series'].items())
    scale = nice_scale(max(max(v) for _, v in series))
    x0, w = ox + 62, pw - 80
    axes(parts, scale, p['unit'], x0, PANEL_Y, w, PANEL_H, ox + 16)
    group_w = w / len(p['categories'])
    bar_w = min(30, (group_w - 14) / len(series))
    for c, cat in enumerate(p['categories']):
        gx = x0 + group_w * c + (group_w - bar_w * len(series)) / 2
        for s, (_, values) in enumerate(series):
            bh = PANEL_H * values[c] / scale[0]
            parts.append(f'<rect x="{gx + s * bar_w:.1f}" y="{PANEL_Y + PANEL_H - bh:.1f}" width="{bar_w - 2:.1f}" height="{bh:.1f}" fill="{PALETTE[s]}"/>')
        for k, line in enumerate(wrap(cat, group_w - 4, 12, False, 2, p['title'])):
            parts.append(f'<text x="{x0 + group_w * c + group_w / 2:.1f}" y="{PANEL_Y + PANEL_H + 18 + k * 14}" font-size="12" text-anchor="middle">{escape(line)}</text>')
    if len(series) > 1:
        row_legend(parts, [n for n, _ in series], x0, PANEL_Y + PANEL_H + 50, w, p['title'])


def line_panel(parts, p, ox, pw):
    series = list(p['series'].items())
    scale = nice_scale(max(max(v) for _, v in series))
    x0, w = ox + 62, pw - 90
    axes(parts, scale, p['unit'], x0, PANEL_Y, w, PANEL_H, ox + 16)
    line_marks(parts, [v for _, v in series], p['x'], scale, x0, PANEL_Y, w, PANEL_H)
    row_legend(parts, [n for n, _ in series], x0, PANEL_Y + PANEL_H + 36, w, p['title'])


def pie_panel(parts, p, ox, pw):
    (_, pie), = p['pies'].items()
    assert sum(pie.values()) == 100, p['title']
    r, cx, cy = 82, ox + 100, PANEL_Y + 110
    angle = -math.pi / 2
    for i, (cat, pct) in enumerate(pie.items()):
        sweep = 2 * math.pi * pct / 100
        x1, y1 = cx + r * math.cos(angle), cy + r * math.sin(angle)
        x2, y2 = cx + r * math.cos(angle + sweep), cy + r * math.sin(angle + sweep)
        parts.append(f'<path d="M{cx},{cy} L{x1:.1f},{y1:.1f} A{r},{r} 0 {1 if sweep > math.pi else 0} 1 {x2:.1f},{y2:.1f} Z" fill="{PALETTE[i]}" stroke="#ffffff" stroke-width="2"/>')
        mid = angle + sweep / 2
        if pct >= 8:
            parts.append(f'<text x="{cx + r * 0.64 * math.cos(mid):.1f}" y="{cy + r * 0.64 * math.sin(mid) + 4:.1f}" font-size="12" fill="#ffffff" font-weight="bold" text-anchor="middle">{pct}%</text>')
        else:
            parts.append(f'<text x="{cx + (r + 16) * math.cos(mid):.1f}" y="{cy + (r + 16) * math.sin(mid) + 4:.1f}" font-size="12" font-weight="bold" text-anchor="middle">{pct}%</text>')
        angle += sweep
    side_legend(parts, list(pie), ox + 205, PANEL_Y + 30, pw - 215, p['title'])


def table_panel(parts, p, ox, pw):
    cols = len(p['header'])
    col_w = (pw - 30) / cols
    x0, y = ox + 15, PANEL_Y
    for r, row in enumerate([p['header']] + p['rows']):
        rh = 44 if r == 0 else 32
        fill = '#dbeafe' if r == 0 else ('#f8fafc' if r % 2 else '#ffffff')
        for c, cell in enumerate(row):
            x = x0 + col_w * c
            parts.append(f'<rect x="{x:.1f}" y="{y}" width="{col_w:.1f}" height="{rh}" fill="{fill}" stroke="#94a3b8"/>')
            bold = r == 0 or c == 0
            lines = wrap(cell, col_w - 10, 12 if r == 0 else 13, bold, 2, p['title'])
            for k, line in enumerate(lines):
                dy = rh / 2 + 4 + (k - (len(lines) - 1) / 2) * 14
                parts.append(f'<text x="{x + col_w / 2:.1f}" y="{y + dy:.1f}" font-size="{12 if r == 0 else 13}" font-weight="{"bold" if bold else "normal"}" text-anchor="middle">{escape(line)}</text>')
        y += rh
    assert y <= 430, p['title']


PANELS = {'bar': bar_panel, 'line': line_panel, 'pie': pie_panel, 'table': table_panel}


def mixed_svg(f):
    """Two charts side by side, so the figure fills the 760x440 frame at full size."""
    assert len(f['parts']) == 2, f['title']
    parts = frame(f['title'])
    parts.append('<line x1="380" y1="52" x2="380" y2="420" stroke="#e2e8f0"/>')
    for i, p in enumerate(f['parts']):
        ox = i * 380
        for k, line in enumerate(wrap(p['title'], 350, 14, True, 1, f['title'])):
            parts.append(f'<text x="{ox + 190}" y="{66 + k * 16}" font-size="14" font-weight="bold" text-anchor="middle">{escape(line)}</text>')
        PANELS[p['kind']](parts, p, ox, 380)
    parts.append('</g></svg>')
    return ''.join(parts)


# ---------------------------------------------------------------- island map

MAP_FILL = {'water': '#bfdbfe', 'sand': '#fde68a', 'building': '#cbd5e1', 'road': '#d1d5db', 'green': '#bbf7d0'}
CELL = 32
# (name, x, y, w, h, kind, label): label None centres the name in the area,
# (x, y) puts it at that grid point (left-aligned), '' draws no label.
ISLAND = {
    'Before': [('Sea', 0, 0, 10, 8, 'water', (0.3, 0.65)), ('Island', 1, 1, 8, 5.5, 'green', (1.3, 1.7)),
               ('Hills', 5.5, 1.5, 3, 2, 'sand', None), ('Village', 2, 4, 2.5, 1.5, 'building', None),
               ('Pier', 2.75, 6.5, 0.75, 1, 'road', (3.75, 7.25))],
    'After': [('Sea', 0, 0, 10, 8, 'water', (0.3, 0.65)), ('Island', 1, 1, 8, 5.5, 'green', (1.3, 1.7)),
              ('Wind turbines', 5.5, 1.5, 3, 2, 'sand', (6.0, 3.25)), ('Village', 2, 4, 2.5, 1.5, 'building', None),
              ('Access road', 4.5, 4.55, 2.5, 0.4, 'road', (4.6, 4.35)), ('', 6.6, 3.5, 0.4, 1.45, 'road', ''),
              ('Control building', 7.1, 4.1, 1.8, 1.1, 'building', None),
              ('Longer pier', 2.75, 6.5, 0.75, 1.5, 'road', (3.75, 7.5))],
}


def turbine(x, y):
    """A small three-bladed turbine on a mast, hub at (x, y)."""
    out = [f'<line x1="{x}" y1="{y}" x2="{x}" y2="{y + 16}" stroke="#334155" stroke-width="2"/>']
    for a in (-90, 30, 150):
        rad = math.radians(a)
        out.append(f'<line x1="{x}" y1="{y}" x2="{x + 9 * math.cos(rad):.1f}" y2="{y + 9 * math.sin(rad):.1f}" stroke="#334155" stroke-width="2"/>')
    return ''.join(out)


def island_svg(f):
    parts = frame(f['title'], height=380)
    for p, (label, areas) in enumerate(ISLAND.items()):
        ox, oy = 30 + p * 380, 70
        parts.append(f'<text x="{ox + 160}" y="{oy - 12}" font-size="15" font-weight="bold" text-anchor="middle">{escape(label)}</text>')
        for name, x, y, w, h, kind, _ in areas:
            assert 0 <= x and x + w <= 10 and 0 <= y and y + h <= 8, (label, name)
            parts.append(f'<rect x="{ox + x * CELL:.1f}" y="{oy + y * CELL:.1f}" width="{w * CELL:.1f}" height="{h * CELL:.1f}" fill="{MAP_FILL[kind]}" stroke="#475569"/>')
        parts.append(f'<rect x="{ox}" y="{oy}" width="{10 * CELL}" height="{8 * CELL}" fill="none" stroke="#172554"/>')
        if label == 'After':
            parts += [turbine(ox + gx * CELL, oy + 1.95 * CELL) for gx in (6.2, 7.0, 7.8)]
        for name, x, y, w, h, kind, at in areas:       # labels last, so no area covers one
            if at == '':
                continue
            if at is None:
                lines = wrap(name, w * CELL - 6, 11, False, 2, name)
                for k, line in enumerate(lines):
                    parts.append(f'<text x="{ox + (x + w / 2) * CELL:.1f}" y="{oy + (y + h / 2) * CELL + 4 + (k - (len(lines) - 1) / 2) * 13:.1f}" font-size="11" text-anchor="middle">{escape(line)}</text>')
            else:
                size, style = (13, ' font-style="italic"') if name in ('Sea', 'Island') else (11, '')
                assert at[0] * CELL + text_width(name, size) <= 10 * CELL - 4, (label, name)
                parts.append(f'<text x="{ox + at[0] * CELL:.1f}" y="{oy + at[1] * CELL:.1f}" font-size="{size}"{style}>{escape(name)}</text>')
        parts.append(f'<text x="{ox + 10 * CELL - 22}" y="{oy + 18}" font-size="13" font-weight="bold">N&#8593;</text>')
    parts.append('</g></svg>')
    return ''.join(parts)


# ------------------------------------------------------- PTE process diagrams

def process_svg(title, steps):
    """The migration 000087 diagram on a 760 canvas, with names wrapped inside each box."""
    width, height = 760, 280
    svg = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}" style="background:#ffffff;font-family:Arial,sans-serif;">',
           f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
           f'<text x="30" y="36" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>',
           '<defs><marker id="arrow" viewBox="0 0 10 10" refX="6" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">'
           '<path d="M 0 1 L 10 5 L 0 9 z" fill="#3b82f6"/></marker></defs>']
    n = len(steps)
    assert 2 <= n <= 5, title
    step_w, step_h, y = 122, 88, 110
    gap = (width - 60 - n * step_w) / (n - 1)
    for idx, text in enumerate(steps):
        x = 30 + idx * (step_w + gap)
        svg.append(f'<rect x="{x:.1f}" y="{y}" width="{step_w}" height="{step_h}" fill="#eff6ff" stroke="#3b82f6" stroke-width="1.5" rx="6"/>')
        svg.append(f'<circle cx="{x + 18:.1f}" cy="{y + 18}" r="11" fill="#3b82f6"/>')
        svg.append(f'<text x="{x + 18:.1f}" y="{y + 22}" font-size="11" font-weight="bold" fill="#ffffff" text-anchor="middle">{idx + 1}</text>')
        size = next(s for s in (12, 11, 10) if fits(text, step_w - 14, s))
        lines = wrap(text, step_w - 14, size, True, 3, title)
        for k, line in enumerate(lines):
            svg.append(f'<text x="{x + step_w / 2:.1f}" y="{y + 58 - (len(lines) - 1) * 7 + k * 14}" font-size="{size}" font-weight="bold" fill="#1e293b" text-anchor="middle">{escape(line)}</text>')
        if idx < n - 1:
            ax1, ay = x + step_w + 3, y + step_h / 2
            svg.append(f'<line x1="{ax1:.1f}" y1="{ay}" x2="{ax1 + gap - 9:.1f}" y2="{ay}" stroke="#3b82f6" stroke-width="2" marker-end="url(#arrow)"/>')
    svg.append('</svg>')
    return ''.join(svg)


def fits(text, box_w, size):
    try:
        wrap(text, box_w, size, True, 3)
        return True
    except AssertionError:
        return False


# ---------------------------------------------------------------------- SQL

# md5 of each value as migrations 000052, 000083, 000087 and 000093 wrote it.
OLD_IMAGE = {
    'ielts-audit-001': '10be9fc2ded5603f0362b4fbc9edca2b', 'ielts-audit-002': '29f4a6e09523e50eb19ea303c2b99d15',
    'ielts-audit-003': '1349de6284cf6515668437d63aa5fdd8', 'ielts-audit-004': '3c87d72019f35762237a1839ff2d38f3',
    'ielts-audit-005': '0349f428285a1774992681c660ee0690', 'ielts-b2-wt1-013': '8604d8a2718114e51badeb333e4d3c51',
    'ielts-b2-wt1-050': '53b2da3caad534be6daeb663ee04592e', 'ielts-b4-fig-035': '2111b22ca4c9a5870066c962553420d3',
    'ielts-b4-fig-036': '5a4398fbce1d7362a7b7d7d6c9a745ee', 'ielts-b4-fig-037': '8f2836ad45230cb089d0743445d8c38e',
    'ielts-b4-fig-038': 'dbe3edde5accc6edbc1dd854222c87af', 'pte-b3-di-041': '8d3289fc624556a33ce786b94c483327',
    'pte-b3-di-042': 'b5e34c2e4e82c7cc73c7506d81cadf8b', 'pte-b3-di-043': '6c022f2ecab272952bb9c6ae7ee7ec2e',
    'pte-b3-di-044': '6406b9c086de728f8e5cdd4e164f805c', 'pte-b3-di-045': '71a1f3ce7ac780f08648fbffdd8a2e5d',
    'pte-b3-di-046': '337b6c5c9146dfc40a0f4fd5c9245c3b', 'pte-b3-di-047': '1ee8dd2d5f8f3ce8b439018733233ab6',
    'pte-b3-di-048': '5e0d1c31348dc4c929b64b28c9a77451', 'pte-b3-di-049': 'e46bab4979cdeee4474212c5c640346f',
    'pte-b3-di-050': 'fa6a4f6a8c4819b7804e3a411a8f532e',
}
OLD_PROMPT = {
    'ielts-audit-001': '50565cd43582687844f0f58ea86d6757', 'ielts-audit-002': 'de4addf4583799f23604751f81b02ea9',
    'ielts-audit-003': 'c1cd5783a5bd8fbe2704efb76ad818b1', 'ielts-audit-004': 'e3b9b4e2d5228e362c3499832ea6c501',
    'ielts-audit-005': 'a0684cc5cfd0d73a364a1dc810680057',
}
OLD_DATA = {
    'ielts-audit-001': '2814cf34d5f948b31968853b5bfdf849', 'ielts-audit-002': '9ab75746d91b470e42d9353c5e445c67',
    'ielts-audit-003': 'bbae5065e9ef43b7e2059e14a84db2a4', 'ielts-audit-004': '12ec48dbc7dbad16e01b4da524b4015c',
    'ielts-audit-005': '88cba552ece480eff198d9ac3c62db7e',
}
OLD = {'image_url': OLD_IMAGE, 'prompt': OLD_PROMPT, 'figure_data': OLD_DATA}


def lit(value):
    return "'" + value.replace("'", "''") + "'"


def by_title(figures, title):
    (f,) = [f for f in figures if f['title'] == title]
    return f


updates = []                                     # (id, column, new value)
for qid, title, desc, header, rows, total in TABLES:
    updates.append((qid, 'image_url', data_uri(table_svg(title, header, rows, total))))
    updates.append((qid, 'prompt', f'The table below shows {desc}.' + TASK1_TAIL))
    updates.append((qid, 'figure_data', table_data(title, desc, header, rows, total, TABLE_FEATURES[qid])))
updates.append(('ielts-b2-wt1-013', 'image_url', data_uri(line_svg(by_title(B2, 'Cinema tickets and streaming subscriptions')))))
updates.append(('ielts-b2-wt1-050', 'image_url', data_uri(island_svg(by_title(B2, 'An island before and after a wind farm')))))
for qid, title in [('ielts-b4-fig-035', 'Visitors to the Lake District National Park'), ('ielts-b4-fig-036', 'Coffee shops in the UK'),
                   ('ielts-b4-fig-037', 'Plastic packaging thrown away and recycled'), ('ielts-b4-fig-038', 'Bus and car journeys in Brighton and Hove')]:
    updates.append((qid, 'image_url', data_uri(mixed_svg(by_title(B4, title)))))
assert len(PROCESSES) == 10
for i, p in enumerate(PROCESSES, 41):
    updates.append((f'pte-b3-di-{i:03}', 'image_url', data_uri(process_svg(p['data']['title'], p['data']['steps']))))
assert all(qid in OLD[col] for qid, col, _ in updates) and len(updates) == 31

body = ['-- Generated by scripts/fix_figure_images.py.',
        '-- Redraws practice figures the 2026-10-16 image audit found broken. Each',
        '-- update applies only while the row still holds the original value.', '']
for qid, col, value in updates:
    body.append(f"UPDATE questions SET {col} = {lit(value)} WHERE id = {lit(qid)} AND md5({col}) = {lit(OLD[col][qid])};")

KEEP = '-- The redrawn figures are kept on rollback; the originals were broken.\nSELECT 1;\n'
(ROOT / f'migrations/{MIGRATION}.up.sql').write_text('\n'.join(body) + '\n', encoding='utf-8', newline='\n')
(ROOT / f'migrations/{MIGRATION}.down.sql').write_text(KEEP, encoding='utf-8', newline='\n')
print('updates:', len(updates))
