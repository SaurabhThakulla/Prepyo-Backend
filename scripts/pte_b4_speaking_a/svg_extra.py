"""SVG drawings for the Describe Image kinds batch 3 did not have: plan maps
(one panel or a before-and-after pair), labelled cross-sections and cycle
diagrams. Same look as scripts/pte_speaking_b3/svg_utils.py (white
background, Arial, slate text), and the same data URI encoding.
"""

import math
from html import escape

HEAD = ('<svg xmlns="http://www.w3.org/2000/svg" width="{w}" height="{h}" viewBox="0 0 {w} {h}" '
        'style="background:#ffffff;font-family:Arial,sans-serif;">'
        '<rect width="{w}" height="{h}" fill="#ffffff"/>')

FILLS = {
    'building': ('#e2e8f0', '#64748b'),
    'new': ('#dbeafe', '#2563eb'),
    'green': ('#dcfce7', '#16a34a'),
    'water': ('#bfdbfe', '#3b82f6'),
    'parking': ('#f1f5f9', '#94a3b8'),
    'site': ('#fef3c7', '#d97706'),
    'houses': ('#fde2e4', '#be123c'),
}


def text_width(text, size, bold=False):
    """Rough width in pixels of Arial text, used to check that labels fit."""
    return len(text) * size * (0.6 if bold else 0.55)


def wrap(text, size, max_w, bold=False):
    """Split text into lines no wider than max_w."""
    lines, cur = [], ''
    for word in text.split():
        trial = f'{cur} {word}'.strip()
        if cur and text_width(trial, size, bold) > max_w:
            lines.append(cur)
            cur = word
        else:
            cur = trial
    if cur:
        lines.append(cur)
    return lines


def _label(x, y, text, size=11, bold=False, max_w=None, anchor='middle', fill='#1e293b'):
    lines = wrap(text, size, max_w, bold) if max_w else [text]
    assert not max_w or all(text_width(l, size, bold) <= max_w for l in lines), f'label too wide: {text}'
    weight = ' font-weight="bold"' if bold else ''
    start = y - (len(lines) - 1) * (size + 2) / 2
    return ''.join(
        f'<text x="{x:.1f}" y="{start + i * (size + 2) + size * 0.35:.1f}" font-size="{size}"{weight} '
        f'text-anchor="{anchor}" fill="{fill}">{escape(line)}</text>'
        for i, line in enumerate(lines))


def _feature(f, ox, oy):
    """One map feature, offset by the panel origin (ox, oy)."""
    kind = f['kind']
    if kind in ('road', 'path', 'river'):
        pts = ' '.join(f'{ox + x:.1f},{oy + y:.1f}' for x, y in f['points'])
        if kind == 'road':
            out = (f'<polyline points="{pts}" fill="none" stroke="#9ca3af" stroke-width="{f.get("width", 16)}" '
                   f'stroke-linejoin="round"/>'
                   f'<polyline points="{pts}" fill="none" stroke="#f8fafc" stroke-width="1.5" stroke-dasharray="8 6"/>')
        elif kind == 'path':
            out = f'<polyline points="{pts}" fill="none" stroke="#a16207" stroke-width="2.5" stroke-dasharray="5 4"/>'
        else:
            out = (f'<polyline points="{pts}" fill="none" stroke="#93c5fd" stroke-width="{f.get("width", 22)}" '
                   f'stroke-linejoin="round" stroke-linecap="round"/>')
        if f.get('label'):
            lx, ly = f['label_at']
            out += _label(ox + lx, oy + ly, f['label'], 10, fill=f.get('label_fill', '#334155'))
        return out
    x, y, w, h = f['box']
    fill, stroke = FILLS[f.get('style', 'building')]
    if kind == 'area':
        out = (f'<ellipse cx="{ox + x + w / 2:.1f}" cy="{oy + y + h / 2:.1f}" rx="{w / 2:.1f}" ry="{h / 2:.1f}" '
               f'fill="{fill}" stroke="{stroke}" stroke-width="1.5"/>')
    else:
        dash = ' stroke-dasharray="6 4"' if f.get('style') == 'site' else ''
        out = (f'<rect x="{ox + x:.1f}" y="{oy + y:.1f}" width="{w}" height="{h}" rx="4" fill="{fill}" '
               f'stroke="{stroke}" stroke-width="1.5"{dash}/>')
    if f.get('label'):
        out += _label(ox + x + w / 2, oy + y + h / 2, f['label'], 10, bold=True, max_w=w - 8)
    return out


def _north(x, y):
    return (f'<path d="M {x} {y + 22} L {x + 7} {y + 2} L {x + 14} {y + 22} L {x + 7} {y + 16} Z" fill="#1e293b"/>'
            f'<text x="{x + 7}" y="{y - 3}" font-size="11" font-weight="bold" text-anchor="middle" fill="#1e293b">N</text>')


def map_svg(title, panels, width=720):
    """A plan map. panels is a list of {'caption', 'features'}; each panel's
    features use coordinates inside a 300 x 260 box (one panel: 640 x 300)."""
    assert len(panels) in (1, 2)
    if len(panels) == 1:
        pw, ph, positions = 640, 300, [(40, 70)]
    else:
        pw, ph, positions = 300, 260, [(40, 90), (380, 90)]
    height = positions[0][1] + ph + 30
    svg = [HEAD.format(w=width, h=height),
           f'<text x="40" y="36" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>']
    for (ox, oy), panel in zip(positions, panels):
        if panel.get('caption'):
            svg.append(f'<text x="{ox + pw / 2}" y="{oy - 12}" font-size="14" font-weight="bold" text-anchor="middle" '
                       f'fill="#1e293b">{escape(panel["caption"])}</text>')
        svg.append(f'<rect x="{ox}" y="{oy}" width="{pw}" height="{ph}" fill="#fafaf9" stroke="#cbd5e1" stroke-width="1.5"/>')
        for f in panel['features']:
            svg.append(_feature(f, ox, oy))
        svg.append(_north(ox + pw - 24, oy + 16))
    svg.append('</svg>')
    return ''.join(svg)


def layers_svg(title, layers, note=''):
    """A labelled cross-section: layers from top to bottom, each (name, detail, colour)."""
    width = 720
    band_h = 44
    top = 70
    x0, x1 = 60, 330
    height = top + band_h * len(layers) + (60 if note else 40)
    svg = [HEAD.format(w=width, h=height),
           f'<text x="40" y="36" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>']
    for i, (name, detail, colour) in enumerate(layers):
        y = top + i * band_h
        svg.append(f'<rect x="{x0}" y="{y}" width="{x1 - x0}" height="{band_h}" fill="{colour}" stroke="#475569" stroke-width="1"/>')
        cy = y + band_h / 2
        svg.append(f'<line x1="{x1}" y1="{cy}" x2="{x1 + 40}" y2="{cy}" stroke="#475569" stroke-width="1"/>'
                   f'<circle cx="{x1}" cy="{cy}" r="2.5" fill="#475569"/>')
        assert text_width(name, 13, True) <= 300 and text_width(detail, 11) <= 300, f'layer label too wide: {name}'
        svg.append(f'<text x="{x1 + 48}" y="{cy - 3}" font-size="13" font-weight="bold" fill="#1e293b">{escape(name)}</text>')
        svg.append(f'<text x="{x1 + 48}" y="{cy + 13}" font-size="11" fill="#475569">{escape(detail)}</text>')
    if note:
        svg.append(f'<text x="{x0}" y="{height - 20}" font-size="11" fill="#64748b">{escape(note)}</text>')
    svg.append('</svg>')
    return ''.join(svg)


def cycle_svg(title, steps):
    """A cycle diagram: 3-6 stages placed round a circle, joined by arrows."""
    n = len(steps)
    assert 3 <= n <= 6
    width, height = 720, 460
    cx, cy, r = 360, 255, 150
    bw, bh = 150, 58
    svg = [HEAD.format(w=width, h=height),
           f'<text x="40" y="36" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>',
           '<defs><marker id="cyc" viewBox="0 0 10 10" refX="6" refY="5" markerWidth="7" markerHeight="7" '
           'orient="auto-start-reverse"><path d="M 0 1 L 10 5 L 0 9 z" fill="#3b82f6"/></marker></defs>']
    angles = [-math.pi / 2 + 2 * math.pi * i / n for i in range(n)]
    # arrows along the circle between the boxes
    for i in range(n):
        a1 = angles[i] + 0.42
        a2 = angles[(i + 1) % n] - 0.42 + (2 * math.pi if i == n - 1 else 0)
        x1, y1 = cx + r * math.cos(a1), cy + r * math.sin(a1)
        x2, y2 = cx + r * math.cos(a2), cy + r * math.sin(a2)
        svg.append(f'<path d="M {x1:.1f} {y1:.1f} A {r} {r} 0 0 1 {x2:.1f} {y2:.1f}" fill="none" stroke="#3b82f6" '
                   f'stroke-width="2" marker-end="url(#cyc)"/>')
    for i, (step, a) in enumerate(zip(steps, angles)):
        x, y = cx + r * math.cos(a) - bw / 2, cy + r * math.sin(a) - bh / 2
        svg.append(f'<rect x="{x:.1f}" y="{y:.1f}" width="{bw}" height="{bh}" rx="8" fill="#eff6ff" stroke="#3b82f6" stroke-width="1.5"/>')
        svg.append(f'<circle cx="{x + 14:.1f}" cy="{y + 14:.1f}" r="10" fill="#3b82f6"/>'
                   f'<text x="{x + 14:.1f}" y="{y + 18:.1f}" font-size="11" font-weight="bold" fill="#ffffff" '
                   f'text-anchor="middle">{i + 1}</text>')
        svg.append(_label(x + bw / 2 + 8, y + bh / 2, step, 11, bold=True, max_w=bw - 36))
    svg.append('</svg>')
    return ''.join(svg)
