"""SVG generation utilities for PTE Describe Image (DI) charts.
Outputs standalone SVG elements converted to data:image/svg+xml URIs.
"""

import math
from urllib.parse import quote
from html import escape


def data_uri(svg_string: str) -> str:
    """Convert an SVG string into a valid data:image/svg+xml URI."""
    return "data:image/svg+xml," + quote(svg_string, safe="")


def nice_max(max_val):
    """Top of a four-step axis: the smallest round step (1, 1.5, 2, 2.5, 3, 4, 5, 6 or 8 x 10^k) whose
    four steps leave some headroom above the largest value."""
    if max_val <= 0:
        return 4
    target = max_val * 1.05 / 4
    power = 10 ** math.floor(math.log10(target))
    step = next(m * power for m in (1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10) if m * power >= target)
    return step * 4


def bar_svg(title: str, categories: list, series: dict, unit: str = "") -> str:
    """Generate a clean, high-contrast bar chart SVG."""
    width = 720
    height = 400
    margin_left = 130
    margin_bottom = 70
    margin_top = 60
    margin_right = 40

    plot_w = width - margin_left - margin_right
    plot_h = height - margin_top - margin_bottom

    palette = ["#2563eb", "#059669", "#d97706", "#7c3aed", "#db2777"]
    series_names = list(series.keys())
    num_series = len(series_names)
    num_cats = len(categories)

    all_vals = [v for s in series.values() for v in s]
    max_val = max(all_vals) if all_vals else 100
    grid_max = nice_max(max_val)

    svg = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}" style="background:#ffffff;font-family:Arial,sans-serif;">',
        f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
        f'<text x="{margin_left}" y="32" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>',
    ]
    if unit:
        svg.append(f'<text x="{margin_left}" y="50" font-size="12" fill="#64748b">Unit: {escape(unit)}</text>')

    for i in range(5):
        val = f"{round(grid_max * i / 4, 6):g}"
        y = margin_top + plot_h - (i / 4) * plot_h
        svg.append(f'<line x1="{margin_left}" y1="{y:.1f}" x2="{width - margin_right}" y2="{y:.1f}" stroke="#e2e8f0" stroke-width="1"/>')
        svg.append(f'<text x="{margin_left - 10}" y="{y + 4:.1f}" font-size="11" text-anchor="end" fill="#64748b">{val}</text>')

    svg.append(f'<line x1="{margin_left}" y1="{margin_top}" x2="{margin_left}" y2="{margin_top + plot_h}" stroke="#94a3b8" stroke-width="1.5"/>')
    svg.append(f'<line x1="{margin_left}" y1="{margin_top + plot_h}" x2="{width - margin_right}" y2="{margin_top + plot_h}" stroke="#94a3b8" stroke-width="1.5"/>')

    group_w = plot_w / num_cats
    bar_w = (group_w * 0.7) / num_series

    for c_idx, cat in enumerate(categories):
        group_x = margin_left + c_idx * group_w + (group_w * 0.15)
        for s_idx, sname in enumerate(series_names):
            val = series[sname][c_idx]
            bar_h = (val / grid_max) * plot_h
            bx = group_x + s_idx * bar_w
            by = margin_top + plot_h - bar_h
            color = palette[s_idx % len(palette)]
            svg.append(f'<rect x="{bx:.1f}" y="{by:.1f}" width="{bar_w - 2:.1f}" height="{bar_h:.1f}" fill="{color}" rx="2"/>')
            svg.append(f'<text x="{bx + (bar_w - 2)/2:.1f}" y="{by - 4:.1f}" font-size="10" text-anchor="middle" fill="#334155">{val}</text>')

        cat_center = margin_left + (c_idx + 0.5) * group_w
        svg.append(f'<text x="{cat_center:.1f}" y="{margin_top + plot_h + 20}" font-size="11" text-anchor="middle" fill="#334155">{escape(cat)}</text>')

    leg_x = margin_left
    leg_y = height - 15
    for s_idx, sname in enumerate(series_names):
        color = palette[s_idx % len(palette)]
        svg.append(f'<rect x="{leg_x}" y="{leg_y - 10}" width="12" height="12" fill="{color}" rx="2"/>')
        svg.append(f'<text x="{leg_x + 16}" y="{leg_y}" font-size="11" fill="#334155">{escape(sname)}</text>')
        leg_x += len(sname) * 7 + 40

    svg.append('</svg>')
    return "".join(svg)


def line_svg(title: str, categories: list, series: dict, unit: str = "") -> str:
    """Generate a clean line graph SVG."""
    width = 720
    height = 400
    margin_left = 130
    margin_bottom = 70
    margin_top = 60
    margin_right = 40

    plot_w = width - margin_left - margin_right
    plot_h = height - margin_top - margin_bottom

    palette = ["#2563eb", "#dc2626", "#059669", "#d97706"]
    series_names = list(series.keys())
    num_cats = len(categories)

    all_vals = [v for s in series.values() for v in s]
    max_val = max(all_vals) if all_vals else 100
    min_val = min(all_vals) if all_vals else 0
    grid_max = nice_max(max_val)

    svg = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}" style="background:#ffffff;font-family:Arial,sans-serif;">',
        f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
        f'<text x="{margin_left}" y="32" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>',
    ]
    if unit:
        svg.append(f'<text x="{margin_left}" y="50" font-size="12" fill="#64748b">Unit: {escape(unit)}</text>')

    for i in range(5):
        val = f"{round(grid_max * i / 4, 6):g}"
        y = margin_top + plot_h - (i / 4) * plot_h
        svg.append(f'<line x1="{margin_left}" y1="{y:.1f}" x2="{width - margin_right}" y2="{y:.1f}" stroke="#e2e8f0" stroke-width="1"/>')
        svg.append(f'<text x="{margin_left - 10}" y="{y + 4:.1f}" font-size="11" text-anchor="end" fill="#64748b">{val}</text>')

    svg.append(f'<line x1="{margin_left}" y1="{margin_top}" x2="{margin_left}" y2="{margin_top + plot_h}" stroke="#94a3b8" stroke-width="1.5"/>')
    svg.append(f'<line x1="{margin_left}" y1="{margin_top + plot_h}" x2="{width - margin_right}" y2="{margin_top + plot_h}" stroke="#94a3b8" stroke-width="1.5"/>')

    step_x = plot_w / (num_cats - 1) if num_cats > 1 else plot_w

    for s_idx, sname in enumerate(series_names):
        vals = series[sname]
        color = palette[s_idx % len(palette)]
        points = []
        for c_idx, val in enumerate(vals):
            x = margin_left + c_idx * step_x
            y = margin_top + plot_h - (val / grid_max) * plot_h
            points.append((x, y, val))

        path_d = "M " + " L ".join(f"{p[0]:.1f} {p[1]:.1f}" for p in points)
        svg.append(f'<path d="{path_d}" fill="none" stroke="{color}" stroke-width="2.5"/>')
        for px, py, val in points:
            svg.append(f'<circle cx="{px:.1f}" cy="{py:.1f}" r="4" fill="{color}" stroke="#ffffff" stroke-width="1.5"/>')
            svg.append(f'<text x="{px:.1f}" y="{py - 8:.1f}" font-size="10" text-anchor="middle" fill="#334155">{val}</text>')

    for c_idx, cat in enumerate(categories):
        x = margin_left + c_idx * step_x
        svg.append(f'<text x="{x:.1f}" y="{margin_top + plot_h + 20}" font-size="11" text-anchor="middle" fill="#334155">{escape(cat)}</text>')

    leg_x = margin_left
    leg_y = height - 15
    for s_idx, sname in enumerate(series_names):
        color = palette[s_idx % len(palette)]
        svg.append(f'<line x1="{leg_x}" y1="{leg_y - 4}" x2="{leg_x + 16}" y2="{leg_y - 4}" stroke="{color}" stroke-width="2.5"/>')
        svg.append(f'<circle cx="{leg_x + 8}" cy="{leg_y - 4}" r="3.5" fill="{color}"/>')
        svg.append(f'<text x="{leg_x + 24}" y="{leg_y}" font-size="11" fill="#334155">{escape(sname)}</text>')
        leg_x += len(sname) * 8 + 45

    svg.append('</svg>')
    return "".join(svg)


def pie_svg(title: str, slices: dict, unit: str = "%") -> str:
    """Generate a pie chart SVG with clear labels and legend."""
    import math

    width = 680
    height = 380
    cx = 240
    cy = 200
    r = 130

    palette = ["#2563eb", "#059669", "#d97706", "#7c3aed", "#db2777", "#0891b2"]
    total = sum(slices.values())

    svg = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}" style="background:#ffffff;font-family:Arial,sans-serif;">',
        f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
        f'<text x="40" y="36" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>',
    ]

    current_angle = -math.pi / 2
    leg_x = 420
    leg_y = 110

    for idx, (label, val) in enumerate(slices.items()):
        fraction = val / total
        slice_angle = fraction * 2 * math.pi
        end_angle = current_angle + slice_angle

        x1 = cx + r * math.cos(current_angle)
        y1 = cy + r * math.sin(current_angle)
        x2 = cx + r * math.cos(end_angle)
        y2 = cy + r * math.sin(end_angle)

        large_arc = 1 if slice_angle > math.pi else 0
        color = palette[idx % len(palette)]

        d = f"M {cx} {cy} L {x1:.2f} {y1:.2f} A {r} {r} 0 {large_arc} 1 {x2:.2f} {y2:.2f} Z"
        svg.append(f'<path d="{d}" fill="{color}" stroke="#ffffff" stroke-width="2"/>')

        mid_angle = current_angle + slice_angle / 2
        label_r = r * 0.65
        lx = cx + label_r * math.cos(mid_angle)
        ly = cy + label_r * math.sin(mid_angle)
        svg.append(f'<text x="{lx:.1f}" y="{ly + 4:.1f}" font-size="11" font-weight="bold" text-anchor="middle" fill="#ffffff">{val}{unit}</text>')

        svg.append(f'<rect x="{leg_x}" y="{leg_y - 12}" width="14" height="14" fill="{color}" rx="2"/>')
        svg.append(f'<text x="{leg_x + 22}" y="{leg_y}" font-size="12" fill="#334155">{escape(label)}: {val}{unit}</text>')
        leg_y += 32

        current_angle = end_angle

    svg.append('</svg>')
    return "".join(svg)


def table_svg(title: str, headers: list, rows: list) -> str:
    """Generate a clean data table SVG."""
    width = 720
    row_height = 36
    header_height = 42
    margin_top = 70
    margin_left = 40
    table_w = width - margin_left * 2
    height = margin_top + header_height + len(rows) * row_height + 40

    col_w = table_w / len(headers)

    svg = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}" style="background:#ffffff;font-family:Arial,sans-serif;">',
        f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
        f'<text x="{margin_left}" y="40" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>',
        f'<rect x="{margin_left}" y="{margin_top}" width="{table_w}" height="{header_height}" fill="#f1f5f9" stroke="#cbd5e1" stroke-width="1"/>',
    ]

    for c_idx, h in enumerate(headers):
        hx = margin_left + c_idx * col_w + 14
        svg.append(f'<text x="{hx:.1f}" y="{margin_top + 26}" font-size="12" font-weight="bold" fill="#1e293b">{escape(h)}</text>')

    for r_idx, row in enumerate(rows):
        ry = margin_top + header_height + r_idx * row_height
        bg = "#ffffff" if r_idx % 2 == 0 else "#f8fafc"
        svg.append(f'<rect x="{margin_left}" y="{ry}" width="{table_w}" height="{row_height}" fill="{bg}" stroke="#e2e8f0" stroke-width="1"/>')
        for c_idx, val in enumerate(row):
            vx = margin_left + c_idx * col_w + 14
            svg.append(f'<text x="{vx:.1f}" y="{ry + 23}" font-size="12" fill="#334155">{escape(str(val))}</text>')

    svg.append('</svg>')
    return "".join(svg)


def process_svg(title: str, steps: list) -> str:
    """Generate a clean linear flow process diagram SVG."""
    width = 720
    height = 280
    svg = [
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}" style="background:#ffffff;font-family:Arial,sans-serif;">',
        f'<rect width="{width}" height="{height}" fill="#ffffff"/>',
        f'<text x="40" y="36" font-size="16" font-weight="bold" fill="#1e293b">{escape(title)}</text>',
        '<defs><marker id="arrow" viewBox="0 0 10 10" refX="6" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">'
        '<path d="M 0 1 L 10 5 L 0 9 z" fill="#3b82f6"/></marker></defs>',
    ]

    num_steps = len(steps)
    assert 2 <= num_steps <= 5, f"{title}: the diagram fits 2-5 steps, got {num_steps}"
    step_w = 115
    step_h = 75
    y = 120
    gap = (width - 80 - (num_steps * step_w)) / (num_steps - 1) if num_steps > 1 else 0

    for idx, step_text in enumerate(steps):
        x = 40 + idx * (step_w + gap)
        svg.append(f'<rect x="{x:.1f}" y="{y}" width="{step_w}" height="{step_h}" fill="#eff6ff" stroke="#3b82f6" stroke-width="1.5" rx="6"/>')
        svg.append(f'<circle cx="{x + 20:.1f}" cy="{y + 20}" r="11" fill="#3b82f6"/>')
        svg.append(f'<text x="{x + 20:.1f}" y="{y + 24}" font-size="11" font-weight="bold" fill="#ffffff" text-anchor="middle">{idx+1}</text>')

        words = step_text.split()
        line1 = " ".join(words[:2])
        line2 = " ".join(words[2:])
        svg.append(f'<text x="{x + step_w/2:.1f}" y="{y + 44}" font-size="11" font-weight="bold" fill="#1e293b" text-anchor="middle">{escape(line1)}</text>')
        if line2:
            svg.append(f'<text x="{x + step_w/2:.1f}" y="{y + 59}" font-size="10" fill="#475569" text-anchor="middle">{escape(line2)}</text>')

        if idx < num_steps - 1:
            ax1 = x + step_w + 4
            ax2 = ax1 + gap - 8
            ay = y + step_h / 2
            svg.append(f'<line x1="{ax1:.1f}" y1="{ay}" x2="{ax2:.1f}" y2="{ay}" stroke="#3b82f6" stroke-width="2" marker-end="url(#arrow)"/>')

    svg.append('</svg>')
    return "".join(svg)
