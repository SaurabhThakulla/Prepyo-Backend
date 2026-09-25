import re
import sys
from pathlib import Path
from collections import Counter

REPO_ROOT = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(REPO_ROOT))
from scripts.live_topics import question_clashes, topic_clashes

STOP_WORDS = {
    "a", "an", "the", "and", "or", "in", "on", "at", "to", "for", "of", "with", "by",
    "from", "up", "about", "into", "over", "after", "is", "are", "was", "were", "be",
    "been", "being", "have", "has", "had", "do", "does", "did", "but", "if", "then",
    "else", "when", "where", "why", "how", "all", "any", "both", "each", "few", "more",
    "most", "other", "some", "such", "no", "nor", "not", "only", "own", "same", "so",
    "than", "too", "very", "can", "will", "just", "should", "now", "it", "its",
    # Generic title words that say nothing about the topic.
    "first", "new", "that", "we", "our", "you", "your", "they", "their", "what", "who", "which"
}

# Themes that the first draft of this batch repeated many times. Checked on the
# recordings themselves, so renaming a title cannot hide a repeat. No theme may
# appear in more than three of the 400 items.
THEMES = {
    'tree rings': r'tree.ring|growth ring|dendro',
    'solar cells': r'solar cell|photovoltaic',
    'biofilm': r'quorum|biofilm',
    'vents': r'hydrothermal|chemosynth',
    'aquifer': r'aquifer|groundwater',
    'volcanic': r'caldera|volcanic ash|tephra',
    'pottery': r'sherd|pottery|ceramic',
    'stomata': r'stomata',
    'trench': r'trench|hadal|bathyal',
    'irrigation': r'terrac|irrigation|sluice',
    'bioluminescence': r'biolumin|fluoresc',
    'ozone': r'ozone',
}

# Earth-science recordings are capped at ten across the batch.
EARTH_SCIENCE = r'volcan|tectonic|sediment|geolog|erosion|erode|earthquake|glacier|geyser|hailstone|quicksand|el niño|tropical storm|topsoil|soft and uneven'

FORBIDDEN_JARGON = [
    "chemolithoautotrophy",
    "tephrochronology",
    "paleoceanography",
    "dendrochronological"
]

def count_words(text):
    return len(re.sub(r"[^\w\s']+", " ", text).split())

def validate_wfd(items):
    assert len(items) == 50, f"WFD expected 50 items, got {len(items)}"
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-wfd-{i:02d}", f"WFD ID mismatch: {it['id']}"
        sentence = it["sentence"]
        wc = count_words(sentence)
        assert 8 <= wc <= 14, f"WFD item {it['id']} word count {wc} out of range [8, 14]: {sentence}"
        assert sentence[0].isupper(), f"WFD item {it['id']} does not start with capital: {sentence}"
        assert sentence.endswith(('.', '?')), f"WFD item {it['id']} does not end with punctuation: {sentence}"
        clashes = question_clashes('PTE', it['title'], type_name='Write from Dictation')
        assert not clashes, f"WFD item {it['id']} title {it['title']} clashes with: {clashes}"
    print("[OK] WFD validated: 50 items (word counts 8-14, correct punctuation, no clashes)")

def validate_sst(items):
    assert len(items) == 50, f"SST expected 50 items, got {len(items)}"
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-sst-{i:02d}", f"SST ID mismatch: {it['id']}"
        script = it.get("audio_transcript") or it.get("script")
        wc_script = count_words(script)
        assert 120 <= wc_script <= 200, f"SST {it['id']} script word count {wc_script} out of range [120, 200]"
        
        # Pearson word count check: 50-70 words
        model = it["model_answer"]
        wc_model = count_words(model)
        assert 50 <= wc_model <= 70, f"SST {it['id']} model answer word count {wc_model} out of range [50, 70]:\n{model}"
        
        kws = it.get("correct_answers") or it.get("keywords")
        assert len(kws) >= 3, f"SST {it['id']} has fewer than 3 keywords"
        model_lower = model.lower()
        for kw in kws:
            assert kw.lower() in model_lower, f"SST {it['id']} keyword {kw!r} not in model answer: {model}"
            
        clashes = question_clashes('PTE', it['title'], type_name='Summarize Spoken Text')
        assert not clashes, f"SST {it['id']} title {it['title']} clashes with: {clashes}"
    print("[OK] SST validated: 50 items (transcripts 120-200 words, model answers 50-70 words, keywords present, no clashes)")

def validate_fib(items):
    assert len(items) == 50, f"LFIB expected 50 items, got {len(items)}"
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-fib-{i:02d}", f"LFIB ID mismatch: {it['id']}"
        script = it.get("audio_transcript") or it.get("script")
        passage = it["context_passage"]
        blanks = it["blanks"]
        assert len(blanks) == 3, f"LFIB {it['id']} must have exactly 3 blanks, got {len(blanks)}"
        
        reconstructed = passage
        for b in blanks:
            marker = f"[[{b['id']}]]"
            assert marker in reconstructed, f"LFIB {it['id']} marker {marker} missing from passage"
            reconstructed = reconstructed.replace(marker, b["correctAnswer"], 1)
        assert reconstructed == script, f"LFIB {it['id']} reconstructed text does not match script!\nReconstructed: {reconstructed}\nScript:        {script}"
        
        clashes = question_clashes('PTE', it['title'], type_name='Listening Fill in the Blanks')
        assert not clashes, f"LFIB {it['id']} title {it['title']} clashes with: {clashes}"
    print("[OK] LFIB validated: 50 items (3 blanks each, exact script reconstruction, no clashes)")

def validate_hiw(items):
    assert len(items) == 50, f"HIW expected 50 items, got {len(items)}"
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-hiw-{i:02d}", f"HIW ID mismatch: {it['id']}"
        script = it.get("audio_transcript") or it.get("script")
        passage = it["context_passage"]
        spoken = script.split()
        displayed = passage.split()
        assert len(spoken) == len(displayed), f"HIW {it['id']} word counts differ: spoken={len(spoken)}, displayed={len(displayed)}"
        
        diffs = [f"w{idx+1}" for idx, (s, d) in enumerate(zip(spoken, displayed)) if s != d]
        assert len(diffs) == 3, f"HIW {it['id']} must have exactly 3 substituted words, got {len(diffs)} ({diffs})"
        
        clashes = question_clashes('PTE', it['title'], type_name='Highlight Incorrect Word')
        assert not clashes, f"HIW {it['id']} title {it['title']} clashes with: {clashes}"
    print("[OK] HIW validated: 50 items (exact word alignment, exactly 3 substitutions, no clashes)")

def validate_hcs(items):
    assert len(items) == 50, f"HCS expected 50 items, got {len(items)}"
    keys = Counter()
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-hcs-{i:02d}", f"HCS ID mismatch: {it['id']}"
        script = it.get("audio_transcript") or it.get("script")
        wc_script = count_words(script)
        assert 120 <= wc_script <= 200, f"HCS {it['id']} script word count {wc_script} out of range [120, 200]"
        
        assert len(it["options"]) == 4, f"HCS {it['id']} must have 4 options"
        assert len(it["correct_answers"]) == 1, f"HCS {it['id']} must have 1 correct answer"
        key = it["correct_answers"][0]
        keys[key] += 1
        
        for opt in it["options"]:
            wc_opt = count_words(opt["text"])
            assert 30 <= wc_opt <= 50, f"HCS {it['id']} option {opt['id']} word count {wc_opt} out of range [30, 50]"
            
        opt_ids = [opt["id"] for opt in it["options"]]
        assert opt_ids == ["A", "B", "C", "D"], f"HCS {it['id']} option IDs mismatch: {opt_ids}"
        
        clashes = question_clashes('PTE', it['title'], type_name='Highlight Correct Summary')
        assert not clashes, f"HCS {it['id']} title {it['title']} clashes with: {clashes}"
    print(f"[OK] HCS validated: 50 items (transcripts 120-200, 4 options strictly 30-50 words, key distribution: {dict(keys)}, no clashes)")

def validate_smw(items):
    assert len(items) == 50, f"SMW expected 50 items, got {len(items)}"
    keys = Counter()
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-smw-{i:02d}", f"SMW ID mismatch: {it['id']}"
        script = it.get("audio_transcript") or it.get("script")
        assert script.endswith("[beep]"), f"SMW {it['id']} script must end with '[beep]': {script}"
        assert script.count("[beep]") == 1, f"SMW {it['id']} script must have exactly one '[beep]': {script}"
        
        assert len(it["options"]) == 4, f"SMW {it['id']} must have 4 options"
        assert len(it["correct_answers"]) == 1, f"SMW {it['id']} must have 1 correct answer"
        key = it["correct_answers"][0]
        keys[key] += 1
        
        opt_ids = [opt["id"] for opt in it["options"]]
        assert opt_ids == ["A", "B", "C", "D"], f"SMW {it['id']} option IDs mismatch: {opt_ids}"
        
        clashes = question_clashes('PTE', it['title'], type_name='Select Missing Word')
        assert not clashes, f"SMW {it['id']} title {it['title']} clashes with: {clashes}"
    print(f"[OK] SMW validated: 50 items (ends with single [beep], key distribution: {dict(keys)}, no clashes)")

def validate_mcsa(items):
    assert len(items) == 50, f"LMCSA expected 50 items, got {len(items)}"
    keys = Counter()
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-mcsa-{i:02d}", f"LMCSA ID mismatch: {it['id']}"
        script = it.get("audio_transcript") or it.get("script")
        wc_script = count_words(script)
        assert 120 <= wc_script <= 200, f"LMCSA {it['id']} script word count {wc_script} out of range [120, 200]"
        
        assert len(it["options"]) == 4, f"LMCSA {it['id']} must have 4 options"
        assert len(it["correct_answers"]) == 1, f"LMCSA {it['id']} must have 1 correct answer"
        key = it["correct_answers"][0]
        keys[key] += 1
        
        opt_ids = [opt["id"] for opt in it["options"]]
        assert opt_ids == ["A", "B", "C", "D"], f"LMCSA {it['id']} option IDs mismatch: {opt_ids}"
        
        clashes = question_clashes('PTE', it['title'], type_name='MCQ Single Answer')
        assert not clashes, f"LMCSA {it['id']} title {it['title']} clashes with: {clashes}"
    print(f"[OK] LMCSA validated: 50 items (transcripts 120-200 words, 4 options, key distribution: {dict(keys)}, no clashes)")

def validate_mcma(items):
    assert len(items) == 50, f"LMCMA expected 50 items, got {len(items)}"
    letter_counts = Counter()
    pair_counts = Counter()
    for i, it in enumerate(items, 1):
        assert it["id"] == f"q-pl90-mcma-{i:02d}", f"LMCMA ID mismatch: {it['id']}"
        script = it.get("audio_transcript") or it.get("script")
        wc_script = count_words(script)
        assert 120 <= wc_script <= 200, f"LMCMA {it['id']} script word count {wc_script} out of range [120, 200]"
        
        assert 5 <= len(it["options"]) <= 7, f"LMCMA {it['id']} must have 5-7 options, got {len(it['options'])}"
        assert 2 <= len(it["correct_answers"]) <= 3, f"LMCMA {it['id']} must have 2-3 correct answers, got {len(it['correct_answers'])}"
        
        ans = tuple(sorted(it["correct_answers"]))
        pair_counts[ans] += 1
        for a in ans:
            letter_counts[a] += 1
            
        opt_ids = [opt["id"] for opt in it["options"]]
        expected_ids = [chr(ord('A') + idx) for idx in range(len(it["options"]))]
        assert opt_ids == expected_ids, f"LMCMA {it['id']} option IDs mismatch: {opt_ids}"
        
        clashes = question_clashes('PTE', it['title'], type_name='MCQ Multiple Answer')
        assert not clashes, f"LMCMA {it['id']} title {it['title']} clashes with: {clashes}"
    print(f"[OK] LMCMA validated: 50 items (transcripts 120-200 words, 5-7 options, 2-3 correct keys, letters: {dict(sorted(letter_counts.items()))}, no clashes)")

def validate_all(all_tasks):
    all_titles = []
    all_text = []
    for task_name, items in all_tasks.items():
        titles = [it['title'] for it in items]
        all_titles.extend(titles)
        assert len(set(titles)) == len(titles), f"Duplicate titles within {task_name}"
        for it in items:
            t = (it.get("title") or "") + " " + (it.get("audio_transcript") or "") + " " + (it.get("context_passage") or "")
            all_text.append(t.lower())
    
    # 1. Total titles uniqueness
    assert len(set(all_titles)) == len(all_titles) == 400, f"Expected 400 unique titles across Batch E, got {len(set(all_titles))}"
    print("[OK] All 400 titles in Batch E are distinct and unique!")
    
    # 2. Word frequency check: no topic word in more than three of the 400 titles.
    # (Repeats in the recordings themselves are checked separately below.)
    words = []
    for t in all_titles:
        for w in re.findall(r"[a-zA-Z]+", t.lower().replace("'s", "")):
            if w not in STOP_WORDS:
                words.append(w)
    c = Counter(words)
    high = {w: cnt for w, cnt in c.items() if cnt > 3}
    assert not high, f"Word frequency check failed! Words appearing > 3 times across Batch E titles: {high}"
    print(f"[OK] Title word check passed: no topic word appears in more than three of the 400 titles.")
    
    # 3. Theme repeats, judged on the recordings
    transcripts = [(it.get("audio_transcript") or "") for items in all_tasks.values() for it in items]
    counts = {name: sum(bool(re.search(pat, t, re.I)) for t in transcripts) for name, pat in THEMES.items()}
    over = {k: v for k, v in counts.items() if v > 3}
    assert not over, f"Themes repeated in more than three recordings: {over}"
    earth = sum(bool(re.search(EARTH_SCIENCE, t, re.I)) for t in transcripts)
    assert earth <= 10, f"{earth} earth-science recordings; the batch allows at most 10"
    print(f"[OK] Theme check passed on recordings: {counts}; earth science {earth}/10")

    # 4. Specialist jargon check
    joined_text = " ".join(all_text)
    for j in FORBIDDEN_JARGON:
        assert j not in joined_text, f"Forbidden specialist jargon '{j}' found in Batch E content!"
    print("[OK] Specialist jargon check passed! No forbidden jargon found in any item.")
