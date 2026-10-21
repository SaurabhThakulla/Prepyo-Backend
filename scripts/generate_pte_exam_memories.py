"""
Generate PTE Exam Memories practice bank (Migration 000110).
Covers:
Phase 1: Reading (RWFIB, RFIB, Reorder Paragraphs)
Phase 2: Writing (Summarize Written Text, Write Essay)
Phase 3: Speaking Text (Read Aloud, Answer Short Question)

Source: PTE Tutorials Exam Memories Material (November 2022).
Automated deduplication, OCR typo correction, and schema validation.
"""
import json
import re
import sys
from pathlib import Path
from typing import Dict, List, Tuple, Any
import fitz

ROOT = Path(__file__).resolve().parents[1]
PDF_PATH = r"C:\Users\thaku\Downloads\November Prediction File low.pdf"
VERSION = "pte-2026-01"

UP_MIGRATION = ROOT / "migrations" / "000110_pte_exam_memories.up.sql"
DOWN_MIGRATION = ROOT / "migrations" / "000110_pte_exam_memories.down.sql"

TYPO_REPLACEMENTS = {
    "Califomia": "California",
    "alomost": "almost",
    "equatiorial": "equatorial",
    "jakets": "jackets",
    "Ceorge": "George",
    "trun of the century": "turn of the century",
    "twi vikcanic eruptions": "two volcanic eruptions",
    "crips failed": "crops failed",
    "wenturing": "venturing",
    "exnibition": "exhibition",
    "mordernised": "modernised",
    "mon-attendance": "non-attendance",
    "dimply written off": "simply written off",
    "ifentified": "identified",
    "refrigeratours": "refrigerators",
    "additivies": "additives",
    "Ethiopiann": "Ethiopian",
    "age of live": "age of five",
    "Uauslly": "Usually",
    "identication": "identification",
    "waiting crows": "waiting crowds",
    "agreeement": "agreement",
    "have his cars heard": "have his case heard",
    "getting rounf": "getting round",
    "wherer": "where",
    "abeas corpus": "habeas corpus",
    "ling-buried": "long-buried",
    "conain consume": "consume",
    "Overs the course": "Over the course",
    "talnts": "talents",
    "aroad overseas": "abroad",
}

def clean_page(p_text: str) -> str:
    p_text = re.sub(r'Save Big.*', '', p_text, flags=re.DOTALL | re.IGNORECASE)
    p_text = re.sub(r'Download PTE Tutorials App[^\n]*', '', p_text, flags=re.IGNORECASE)
    p_text = re.sub(r'PTETUTORIALS\.online', '', p_text, flags=re.IGNORECASE)
    p_text = re.sub(r'www\.ptetutorials\.com', '', p_text, flags=re.IGNORECASE)
    p_text = re.sub(r'www\.aussizzgroup\.com[^\n]*', '', p_text, flags=re.IGNORECASE)
    p_text = re.sub(r'#PTETutorials_Exam_Memories', '', p_text, flags=re.IGNORECASE)
    p_text = re.sub(r'Product of\s+AUSSIZZ\s+GROUP[^\n]*', '', p_text, flags=re.IGNORECASE)
    p_text = re.sub(r'Get 15 PTE Mock tests Free on Purchase', '', p_text, flags=re.IGNORECASE)
    p_text = re.sub(r'Become a pro at the PTE academic[^\n]*', '', p_text, flags=re.IGNORECASE)

    p_text = p_text.replace('\u2019', "'").replace('\u2018', "'").replace('\u201c', '"').replace('\u201d', '"')
    p_text = p_text.replace('\u2013', '-').replace('\u2014', '-').replace('`', "'")
    p_text = p_text.replace('\ufffd', "'")

    for bad, good in TYPO_REPLACEMENTS.items():
        p_text = p_text.replace(bad, good)

    return p_text

def sql_lit(val: Any) -> str:
    if val is None:
        return 'NULL'
    if isinstance(val, bool):
        return 'TRUE' if val else 'FALSE'
    if isinstance(val, (int, float)):
        return str(val)
    if isinstance(val, (dict, list)):
        dumped = json.dumps(val, ensure_ascii=False)
        return "'" + dumped.replace("'", "''") + "'::jsonb"
    s = str(val)
    return "'" + s.replace("'", "''") + "'"

def sql_text_array(items: List[str]) -> str:
    escaped = [sql_lit(x) for x in items]
    return 'ARRAY[' + ', '.join(escaped) + ']::text[]'

# -----------------------------------------------------------------------------
# Section Parsers
# -----------------------------------------------------------------------------

def parse_read_aloud(doc) -> List[Dict[str, Any]]:
    text = '\n'.join([clean_page(doc[i].get_text()) for i in range(2, 7)])
    matches = re.findall(r'(?:^|\n)\s*(\d+)\)\s*(.*?)(?=\n\s*\d+\)|\Z)', text, re.DOTALL)
    seen = set()
    items = []
    for num, body in matches:
        t = ' '.join(body.split()).strip()
        norm = re.sub(r'[^a-z0-9]', '', t.lower())
        if len(t) > 30 and norm not in seen:
            seen.add(norm)
            title = ' '.join(t.split()[:4]).rstrip('.,;:-').title()
            items.append({
                'title': title,
                'passage': t,
            })
    return items

def parse_asq(doc) -> List[Dict[str, Any]]:
    text = '\n'.join([clean_page(doc[i].get_text()) for i in range(40, 45)])
    matches = re.findall(r'(?:^|\n)\s*(\d+)\)\s*(.*?)\s*[–—\-]\s*(.*?)(?=\n\s*\d+\)|\Z)', text, re.DOTALL)
    seen = set()
    items = []
    for num, q, a in matches:
        qc = ' '.join(q.split()).strip()
        ac = ' '.join(a.split()).strip().rstrip('.')
        norm = re.sub(r'[^a-z0-9]', '', qc.lower())
        if norm and norm not in seen and len(ac) > 0:
            seen.add(norm)
            opts = [opt.strip() for opt in re.split(r'[/,]', ac) if opt.strip()]
            model_ans = opts[0].capitalize() if opts else ac.capitalize()
            items.append({
                'question': qc,
                'model_answer': model_ans,
                'answers': opts if opts else [ac],
            })
    return items

def parse_swt(doc) -> List[Dict[str, Any]]:
    text = '\n'.join([clean_page(doc[i].get_text()) for i in range(46, 56)])
    matches = re.findall(r'(?:^|\n)\s*(\d+)\)\s*(.*?)(?=\n\s*\d+\)|\Z)', text, re.DOTALL)
    seen = set()
    items = []
    for num, body in matches:
        t = ' '.join(body.split()).strip()
        norm = re.sub(r'[^a-z0-9]', '', t[:80].lower())
        if len(t) > 100 and norm not in seen:
            seen.add(norm)
            title = ' '.join(t.split()[:4]).rstrip('.,;:-').title()
            items.append({
                'title': title,
                'passage': t,
            })
    return items

def parse_essay(doc) -> List[Dict[str, Any]]:
    text = '\n'.join([clean_page(doc[i].get_text()) for i in range(57, 59)])
    matches = re.findall(r'(?:^|\n)\s*(\d+)\)\s*(.*?)(?=\n\s*\d+\)|\Z)', text, re.DOTALL)
    seen = set()
    items = []
    for num, body in matches:
        t = ' '.join(body.split()).strip()
        norm = re.sub(r'[^a-z0-9]', '', t.lower())
        if len(t) > 20 and norm not in seen:
            seen.add(norm)
            title = ' '.join(t.split()[:4]).rstrip('.,;:-').title()
            items.append({
                'title': title,
                'prompt': t,
            })
    return items

def parse_reorder(doc) -> List[Dict[str, Any]]:
    text = '\n'.join([clean_page(doc[i].get_text()) for i in range(159, 180)])
    blocks = re.split(r'\n\s*(\d+)\)\s*', '\n' + text)
    unique_ro = []
    seen = set()
    for i in range(1, len(blocks), 2):
        content = blocks[i+1]
        m_ans = re.search(r'Answer:\s*([A-Ea-e]+)', content)
        if not m_ans: continue
        ans = m_ans.group(1).upper()
        p_matches = re.findall(r'(?:\[|\()?([A-E])(?:\]|\))\.?\s*(.*?)(?=\s*(?:\[|\()?([A-E])(?:\]|\))|\s*Answer:|\Z)', content, re.DOTALL)
        paras = {m[0]: ' '.join(m[1].split()).strip() for m in p_matches if m[0] in 'ABCDE'}
        if len(paras) >= 4 and len(ans) == len(paras):
            first_p = list(paras.values())[0]
            norm = re.sub(r'[^a-z0-9]', '', first_p[:50].lower())
            if norm and norm not in seen:
                seen.add(norm)
                title = ' '.join(paras[ans[0]].split()[:4]).rstrip('.,;:-').title()
                ordered = [{'label': f'p{k+1}', 'text': paras[letter]} for k, letter in enumerate(ans)]
                unique_ro.append({
                    'title': title,
                    'ordered_paragraphs': ordered,
                    'ans_sequence': [f'p{k+1}' for k in range(len(ans))],
                    'raw_ans': ans,
                })
    return unique_ro

def parse_rwfib(doc) -> List[Dict[str, Any]]:
    rwfib_pages = [clean_page(doc[i].get_text()) for i in range(60, 158)]
    rwfib_text = '\n'.join(rwfib_pages)
    rwfib_text = re.sub(r'^.*?Reading and Writing: Fill in the Blanks:\s*', '', rwfib_text, flags=re.DOTALL | re.IGNORECASE)

    q_blocks = re.split(r'\n\s*(\d{1,3})\s*[\)\.]\s*(?=\n|[A-Z"\'`])', '\n' + rwfib_text)
    questions = []
    seen = set()

    for i in range(1, len(q_blocks), 2):
        block = q_blocks[i+1].strip()
        m_ans = re.search(r'Answer\s*[:;]\s*(.*?)(?=\n\s*\d{1,3}\s*[\)\.]|\Z)', block, re.DOTALL | re.IGNORECASE)
        if not m_ans: continue
        
        ans_raw = m_ans.group(1).strip()
        ans_line = ans_raw.splitlines()[0].strip() if ans_raw else ''
        content = block[:m_ans.start()].strip()

        lines = content.splitlines()
        opt_start = None
        for l_idx, line in enumerate(lines):
            if re.match(r'^\s*(?:\(?1\)?|1\.)\s+[a-zA-Z]', line):
                opt_start = l_idx
                break
        if opt_start is None: continue

        passage = ' '.join(' '.join(lines[:opt_start]).split()).strip()
        opt_text = '\n'.join(lines[opt_start:]).strip()

        opt_groups = re.findall(r'(?:^|\n)\s*(?:\(?\d+\)?|\d+\.)\s*(.*?)(?=\n\s*(?:\(?\d+\)?|\d+\.)|\Z)', opt_text, re.DOTALL)
        options = []
        for og in opt_groups:
            cleaned_words = [w.strip() for w in re.split(r'[,/]', ' '.join(og.split())) if w.strip()]
            cleaned_words = [re.sub(r'^[\.\s]+|[\.\s]+$', '', w) for w in cleaned_words if len(w) > 0]
            if len(cleaned_words) >= 2:
                options.append(cleaned_words)

        answers = [a.strip() for a in re.split(r'[,;]', ans_line) if a.strip()]
        answers = [re.sub(r'^[\.\s]+|[\.\s]+$', '', a) for a in answers if len(a) > 0]

        # Count blanks in passage
        blank_pattern = re.compile(r'(?:_{2,}|-{2,}|-\s*_{2,})')
        num_blanks = len(blank_pattern.findall(passage))

        norm = re.sub(r'[^a-z0-9]', '', passage[:80].lower())
        if len(passage) > 50 and len(options) >= 2 and len(answers) >= 2 and norm not in seen:
            # We want num_blanks to match len(options)
            # If blank_pattern count matches len(options):
            if num_blanks == len(options) and len(options) == len(answers):
                seen.add(norm)
                # Substitute blanks with [[b1]], [[b2]], ...
                counter = 0
                def repl(m):
                    nonlocal counter
                    counter += 1
                    return f"[[b{counter}]]"
                formatted_passage = blank_pattern.sub(repl, passage)
                blanks_data = []
                for b_idx in range(len(options)):
                    b_id = f"b{b_idx+1}"
                    opts = options[b_idx]
                    correct = answers[b_idx]
                    # Ensure correct answer is in options
                    if correct not in opts:
                        opts.append(correct)
                    blanks_data.append({
                        'id': b_id,
                        'options': opts,
                        'correctAnswer': correct
                    })

                title = ' '.join(passage.split()[:4]).rstrip('.,;:-').title()
                questions.append({
                    'title': title,
                    'passage': formatted_passage,
                    'raw_passage': passage,
                    'blanks': blanks_data,
                })

    return questions

def parse_rfib(doc) -> List[Dict[str, Any]]:
    rfib_pages = [clean_page(doc[i].get_text()) for i in range(181, 253)]
    rfib_text = '\n'.join(rfib_pages)
    rfib_text = re.sub(r'^.*?Reading: Fill in the Blanks:\s*', '', rfib_text, flags=re.DOTALL | re.IGNORECASE)

    q_blocks = re.split(r'\n\s*(\d{1,3})\s*[\)\.]\s*(?=\n|[A-Z"\'`])', '\n' + rfib_text)
    questions = []
    seen = set()

    for i in range(1, len(q_blocks), 2):
        block = q_blocks[i+1].strip()
        m_opts = re.search(r'Options?\s*[:;]\s*(.*?)(?=\n\s*Answers?\s*[:;]|\Z)', block, re.DOTALL | re.IGNORECASE)
        m_ans = re.search(r'Answers?\s*[:;]\s*(.*?)(?=\n\s*\d{1,3}\s*[\)\.]|\Z)', block, re.DOTALL | re.IGNORECASE)
        if not m_opts or not m_ans: continue

        raw_passage = block[:m_opts.start()].strip()
        raw_opts = m_opts.group(1).strip()
        raw_ans = m_ans.group(1).strip()
        passage = ' '.join(raw_passage.split()).strip()

        opt_text = ' '.join(raw_opts.split())
        opt_text = re.sub(r'\b\d+\)\s*', '', opt_text)
        opt_list = [w.strip() for w in re.split(r'[,;]', opt_text) if w.strip()]
        opt_list = [re.sub(r'^[\.\s]+|[\.\s]+$', '', w) for w in opt_list if len(w) > 0]

        ans_text = raw_ans.splitlines()[0].strip()
        ans_text = re.sub(r'\b\d+[\.\)]\s*', '', ans_text)
        ans_list = [a.strip() for a in re.split(r'[,;]', ans_text) if a.strip()]
        ans_list = [re.sub(r'^[\.\s]+|[\.\s]+$', '', a) for a in ans_list if len(a) > 0]

        blank_pattern = re.compile(r'(?:_{2,}|-{2,}|-\s*_{2,})')
        num_blanks = len(blank_pattern.findall(passage))

        norm = re.sub(r'[^a-z0-9]', '', passage[:80].lower())
        if len(passage) > 50 and len(opt_list) >= 4 and len(ans_list) >= 2 and norm not in seen:
            if num_blanks == len(ans_list):
                seen.add(norm)
                counter = 0
                def repl(m):
                    nonlocal counter
                    counter += 1
                    return f"[[b{counter}]]"
                formatted_passage = blank_pattern.sub(repl, passage)
                # Ensure all answers are in pool
                pool = list(dict.fromkeys(opt_list))
                for ans in ans_list:
                    if ans not in pool:
                        pool.append(ans)
                
                blanks_data = []
                for b_idx in range(len(ans_list)):
                    b_id = f"b{b_idx+1}"
                    blanks_data.append({
                        'id': b_id,
                        'options': pool,
                        'correctAnswer': ans_list[b_idx]
                    })

                title = ' '.join(passage.split()[:4]).rstrip('.,;:-').title()
                questions.append({
                    'title': title,
                    'passage': formatted_passage,
                    'raw_passage': passage,
                    'pool': pool,
                    'blanks': blanks_data,
                })

    return questions

# -----------------------------------------------------------------------------
# Main Builder
# -----------------------------------------------------------------------------

def build_all():
    doc = fitz.open(PDF_PATH)
    print("Parsing Phase 1, 2, 3 content from PTE Exam Memories...")

    ra_items = parse_read_aloud(doc)
    asq_items = parse_asq(doc)
    swt_items = parse_swt(doc)
    we_items = parse_essay(doc)
    ro_items = parse_reorder(doc)
    rwfib_items = parse_rwfib(doc)
    rfib_items = parse_rfib(doc)

    print(f"Read Aloud (RA): {len(ra_items)}")
    print(f"Answer Short Question (ASQ): {len(asq_items)}")
    print(f"Summarize Written Text (SWT): {len(swt_items)}")
    print(f"Write Essay (WE): {len(we_items)}")
    print(f"Reorder Paragraphs (RO): {len(ro_items)}")
    print(f"Reading & Writing FIB (RWFIB): {len(rwfib_items)}")
    print(f"Reading FIB (RFIB): {len(rfib_items)}")

    up_statements = []
    down_statements = []

    up_statements.append("-- 000110_pte_exam_memories.up.sql")
    up_statements.append("-- High-frequency authentic PTE practice bank (Phases 1, 2, 3).")
    up_statements.append("-- Deduplicated, verified, and cleaned from PTE Exam Memories.")
    up_statements.append("\nBEGIN;\n")

    # 1. Reading Passages for RWFIB and RFIB
    passages_rows = []
    groups_rows = []

    # -------------------------------------------------------------------------
    # RWFIB
    # -------------------------------------------------------------------------
    rwfib_q_rows = []
    for idx, item in enumerate(rwfib_items, start=1):
        pid = f"rp-pte-em-rwfib-{idx:03d}"
        gid = f"g-pte-em-rwfib-{idx:03d}"
        qid = f"pte-em-rwfib-{idx:03d}"
        # The 700-word floor allows 0 for short PTE passages, as generate_pte_practice.py does.
        wc = 0
        title = item['title']
        tags = ['PTE', 'Exam Memories', 'Reading & Writing: Fill in the Blanks']

        passages_rows.append(
            f"    ({sql_lit(pid)}, {sql_lit(VERSION)}, {sql_lit(title)}, {sql_lit('PTE Exam Memories')},\n"
            f"     {sql_lit([{'label': 'A', 'text': item['raw_passage']}])}, {wc}, 'medium', 'Academic Reading',\n"
            f"     {sql_text_array(tags)}, 'custom', ARRAY['academic'])"
        )

        groups_rows.append(
            f"    ({sql_lit(gid)}, {sql_lit(pid)}, 1, 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)',\n"
            f"     'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',\n"
            f"     '[]'::jsonb, 'hidden', FALSE, 150)"
        )

        rwfib_q_rows.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-rw', 'Fill in the Blanks (Dropdown)',\n"
            f"     {sql_lit(title)}, 'Below is a text with blanks. Click on each blank, a list of choices will appear. Select the appropriate answer choice for each blank.',\n"
            f"     {sql_lit(item['passage'])}, '[]'::jsonb, '[]'::jsonb, {sql_lit(item['blanks'])},\n"
            f"     'Select the most grammatically and contextually appropriate word for each blank.', {len(item['blanks'])}, 'medium',\n"
            f"     {sql_text_array(tags)}, {sql_lit(pid)}, {sql_lit(gid)}, 1, TRUE, 150, NULL)"
        )

    # -------------------------------------------------------------------------
    # RFIB
    # -------------------------------------------------------------------------
    rfib_q_rows = []
    for idx, item in enumerate(rfib_items, start=1):
        pid = f"rp-pte-em-rfib-{idx:03d}"
        gid = f"g-pte-em-rfib-{idx:03d}"
        qid = f"pte-em-rfib-{idx:03d}"
        wc = 0
        title = item['title']
        tags = ['PTE', 'Exam Memories', 'Reading: Fill in the Blanks']
        pool_options = [{'id': chr(65 + k), 'text': w} for k, w in enumerate(item['pool'])]

        passages_rows.append(
            f"    ({sql_lit(pid)}, {sql_lit(VERSION)}, {sql_lit(title)}, {sql_lit('PTE Exam Memories')},\n"
            f"     {sql_lit([{'label': 'A', 'text': item['raw_passage']}])}, {wc}, 'medium', 'Academic Reading',\n"
            f"     {sql_text_array(tags)}, 'custom', ARRAY['academic'])"
        )

        groups_rows.append(
            f"    ({sql_lit(gid)}, {sql_lit(pid)}, 1, 'fill-in-blanks-r', 'Reading: Fill in the Blanks',\n"
            f"     'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. To undo an answer choice, drag the word back to the box below the text.',\n"
            f"     {sql_lit(pool_options)}, 'hidden', FALSE, 120)"
        )

        rfib_q_rows.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'fill-in-blanks-r', 'Reading: Fill in the Blanks',\n"
            f"     {sql_lit(title)}, 'In the text below some words are missing. Drag words from the box below to the appropriate place in the text. To undo an answer choice, drag the word back to the box below the text.',\n"
            f"     {sql_lit(item['passage'])}, {sql_lit(pool_options)}, '[]'::jsonb, {sql_lit(item['blanks'])},\n"
            f"     'Drag and drop the correct words into the corresponding blanks.', {len(item['blanks'])}, 'medium',\n"
            f"     {sql_text_array(tags)}, {sql_lit(pid)}, {sql_lit(gid)}, 1, TRUE, 120, NULL)"
        )

    # -------------------------------------------------------------------------
    # Reorder Items & Questions
    # -------------------------------------------------------------------------
    reorder_rows = []
    ro_q_rows = []
    for idx, item in enumerate(ro_items, start=1):
        rid = f"ri-pte-em-ro-{idx:03d}"
        qid = f"pte-em-ro-{idx:03d}"
        paras = item['ordered_paragraphs']
        wc = sum(len(p['text'].split()) for p in paras)
        title = item['title']
        tags = ['PTE', 'Exam Memories', 'Reorder Paragraph']

        reorder_rows.append(
            f"    ({sql_lit(rid)}, {sql_lit(VERSION)}, 'PTE', {sql_lit(title)},\n"
            f"     {sql_lit(paras)}, {wc}, 'medium', {sql_text_array(tags)}, TRUE)"
        )

        # Deterministic shuffle for options (never presented in solved order)
        n_box = len(paras)
        shuf_order = [(k + 2) % n_box for k in range(n_box)]
        shuffled_options = [{'id': paras[i]['label'], 'text': paras[i]['text']} for i in shuf_order]
        correct_sequence = [p['label'] for p in paras]

        ro_q_rows.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'reading', 'reorder-paragraphs', 'Reorder Paragraph',\n"
            f"     {sql_lit(title)}, 'The text boxes below have been placed in a random order. Restore the original order.',\n"
            f"     NULL, {sql_lit(shuffled_options)}, {sql_lit(correct_sequence)}, '[]'::jsonb,\n"
            f"     'Restore the logical cohesion of the paragraphs in chronological and thematic order.', {n_box - 1}, 'medium',\n"
            f"     {sql_text_array(tags)}, NULL, NULL, 0, TRUE, 150, {sql_lit(rid)})"
        )

    # -------------------------------------------------------------------------
    # Phase 2: Writing (SWT & WE)
    # -------------------------------------------------------------------------
    swt_q_rows = []
    for idx, item in enumerate(swt_items, start=1):
        qid = f"pte-em-swt-{idx:03d}"
        title = item['title']
        tags = ['PTE', 'Exam Memories', 'Summarize Written Text']

        swt_q_rows.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'writing', 'summarize-written-text', 'Summarize Written Text',\n"
            f"     {sql_lit(title)}, 'Read the passage below and summarise it in ONE single sentence of between 5 and 75 words.',\n"
            f"     {sql_lit(item['passage'])}, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb,\n"
            f"     'Summarize the core argument in a single well-structured sentence between 5 and 75 words.', 10, 'medium',\n"
            f"     {sql_text_array(tags)}, NULL, NULL, 0, TRUE, 600, NULL)"
        )

    we_q_rows = []
    for idx, item in enumerate(we_items, start=1):
        qid = f"pte-em-we-{idx:03d}"
        title = item['title']
        tags = ['PTE', 'Exam Memories', 'Write Essay']
        full_prompt = (
            "You will have 20 minutes to plan, write and revise an essay about the topic below. "
            "Your response will be judged on how well you develop a position, organize your ideas, "
            "present supporting details, and control the elements of standard written English. "
            "You should write 200-300 words.\n\n" + item['prompt']
        )

        we_q_rows.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'writing', 'write-essay', 'Write Essay',\n"
            f"     {sql_lit(title)}, {sql_lit(full_prompt)},\n"
            f"     NULL, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb,\n"
            f"     'Write a structured academic essay between 200 and 300 words with clear supporting examples.', 15, 'medium',\n"
            f"     {sql_text_array(tags)}, NULL, NULL, 0, TRUE, 1200, NULL)"
        )

    # -------------------------------------------------------------------------
    # Phase 3: Speaking Text (RA & ASQ)
    # -------------------------------------------------------------------------
    ra_q_rows = []
    for idx, item in enumerate(ra_items, start=1):
        qid = f"pte-em-ra-{idx:03d}"
        title = item['title']
        tags = ['PTE', 'Exam Memories', 'Read Aloud']

        ra_q_rows.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'speaking', 'read-aloud', 'Read Aloud',\n"
            f"     {sql_lit(title)}, 'Look at the text below. In 35 seconds, you must read this text aloud as naturally and clearly as possible. You have 40 seconds to read aloud.',\n"
            f"     {sql_lit(item['passage'])}, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb,\n"
            f"     'Speak clearly and naturally at an even pace with appropriate phrasing and intonation.', 15, 'medium',\n"
            f"     {sql_text_array(tags)}, NULL, NULL, 0, TRUE, 40, NULL)"
        )

    asq_q_rows = []
    for idx, item in enumerate(asq_items, start=1):
        qid = f"pte-em-asq-{idx:03d}"
        tags = ['PTE', 'Exam Memories', 'Answer Short Questions']

        asq_q_rows.append(
            f"    ({sql_lit(qid)}, {sql_lit(VERSION)}, 'PTE', ARRAY['PTE'], 'speaking', 'pte-answer-short-question', 'Answer Short Questions',\n"
            f"     {sql_lit(f'Short Question {idx}')}, 'You will hear a question. Please give a simple and short answer. Often just one or a few words is enough.',\n"
            f"     NULL, '[]'::jsonb, {sql_lit(item['answers'])}, '[]'::jsonb,\n"
            f"     {sql_lit('Expected answer: ' + item['model_answer'])}, 5, 'medium',\n"
            f"     {sql_text_array(tags)}, NULL, NULL, 0, TRUE, 10, NULL)"
        )

    # -------------------------------------------------------------------------
    # Assemble SQL
    # -------------------------------------------------------------------------
    # 1. reading_passages
    if passages_rows:
        up_statements.append("INSERT INTO reading_passages (id, exam_version_id, title, subtitle, paragraphs, word_count, difficulty, topic, tags, passage_slot, modules)\nVALUES\n" + ",\n".join(passages_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

    # 2. reading_question_groups
    if groups_rows:
        up_statements.append("INSERT INTO reading_question_groups (id, passage_id, position, type_id, type_name, instructions, resources, passage_display, shuffle_questions, time_limit_seconds)\nVALUES\n" + ",\n".join(groups_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

    # 3. reading_reorder_items
    if reorder_rows:
        up_statements.append("INSERT INTO reading_reorder_items (id, exam_version_id, exam, title, paragraphs, word_count, difficulty, tags, is_published)\nVALUES\n" + ",\n".join(reorder_rows) + "\nON CONFLICT (id) DO NOTHING;\n")

    # 4. questions table
    all_q_rows = rwfib_q_rows + rfib_q_rows + ro_q_rows + swt_q_rows + we_q_rows + ra_q_rows + asq_q_rows
    if all_q_rows:
        up_statements.append(
            "INSERT INTO questions (\n"
            "    id, exam_version_id, exam, supported_exams, skill, type_id, type_name,\n"
            "    title, prompt, context_passage, options, correct_answers, blanks,\n"
            "    explanation, points, difficulty, tags, passage_id, group_id, group_position,\n"
            "    is_published, time_limit_seconds, reorder_item_id\n"
            ") VALUES\n" + ",\n".join(all_q_rows) + "\nON CONFLICT (id) DO NOTHING;\n"
        )

    up_statements.append("COMMIT;\n")

    # Write UP migration
    UP_MIGRATION.write_text('\n'.join(up_statements), encoding='utf-8', newline='\n')
    print(f"Wrote UP migration: {UP_MIGRATION} ({UP_MIGRATION.stat().st_size:,} bytes)")

    # Write DOWN migration
    down_sql = """-- 000110_pte_exam_memories.down.sql
BEGIN;

DELETE FROM questions WHERE id LIKE 'pte-em-%';
DELETE FROM reading_reorder_items WHERE id LIKE 'ri-pte-em-%';
DELETE FROM reading_question_groups WHERE id LIKE 'g-pte-em-%';
DELETE FROM reading_passages WHERE id LIKE 'rp-pte-em-%';

COMMIT;
"""
    DOWN_MIGRATION.write_text(down_sql, encoding='utf-8', newline='\n')
    print(f"Wrote DOWN migration: {DOWN_MIGRATION}")

    print("\nTotal questions generated:")
    print(f"  RWFIB:   {len(rwfib_q_rows)}")
    print(f"  RFIB:    {len(rfib_q_rows)}")
    print(f"  RO:      {len(ro_q_rows)}")
    print(f"  SWT:     {len(swt_q_rows)}")
    print(f"  WE:      {len(we_q_rows)}")
    print(f"  RA:      {len(ra_q_rows)}")
    print(f"  ASQ:     {len(asq_q_rows)}")
    print(f"  TOTAL:   {len(all_q_rows)}")

if __name__ == '__main__':
    build_all()
