"""Generator for PTE Listening Bank 3 (Batch E)
Generates migrations/000090_pte_listening_bank_3.up.sql and down.sql
Enforces:
- 400 original listening items (8 tasks x 50 questions each)
- Exact type IDs and live type names
- Deterministic byte-identical generation with SHA-256 validation
- Unix newlines ('\\n')
"""

import hashlib
import json
import re
import sys
from pathlib import Path

# Add repo root to sys.path
REPO_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(REPO_ROOT))

from scripts.pte_listening_b3.wfd import WFD_ITEMS
from scripts.pte_listening_b3.fib import FIB_ITEMS
from scripts.pte_listening_b3.hiw import HIW_ITEMS
from scripts.pte_listening_b3.smw import SMW_ITEMS
from scripts.pte_listening_b3.mcsa import LMCSA_ITEMS
from scripts.pte_listening_b3.mcma import LMCMA_ITEMS
from scripts.pte_listening_b3.hcs import HCS_ITEMS
from scripts.pte_listening_b3.sst import SST_ITEMS
from scripts.pte_listening_b3.validator import (
    validate_wfd,
    validate_fib,
    validate_hiw,
    validate_smw,
    validate_mcsa,
    validate_mcma,
    validate_hcs,
    validate_sst,
    validate_all
)

def sql_str(val):
    if val is None:
        return "NULL"
    escaped = val.replace("'", "''")
    return f"'{escaped}'"

def sql_json(val):
    if val is None:
        return "NULL"
    dumped = json.dumps(val, ensure_ascii=False)
    escaped = dumped.replace("'", "''")
    return f"'{escaped}'::jsonb"

def sql_array(arr):
    if not arr:
        return "ARRAY[]::text[]"
    elements = ", ".join("'" + item.replace("'", "''") + "'" for item in arr)
    return f"ARRAY[{elements}]::text[]"

def generate_up_sql():
    # 1. Run all validations first
    print("Validating all 400 PTE Listening Batch E items...")
    validate_wfd(WFD_ITEMS)
    validate_fib(FIB_ITEMS)
    validate_hiw(HIW_ITEMS)
    validate_smw(SMW_ITEMS)
    validate_mcsa(LMCSA_ITEMS)
    validate_mcma(LMCMA_ITEMS)
    validate_hcs(HCS_ITEMS)
    validate_sst(SST_ITEMS)
    
    all_tasks = {
        'WFD': WFD_ITEMS,
        'FIB': FIB_ITEMS,
        'HIW': HIW_ITEMS,
        'SMW': SMW_ITEMS,
        'LMCSA': LMCSA_ITEMS,
        'LMCMA': LMCMA_ITEMS,
        'HCS': HCS_ITEMS,
        'SST': SST_ITEMS
    }
    validate_all(all_tasks)
    
    lines = []
    lines.append("-- Migration 000090: PTE Listening Bank 3 (Batch E)")
    lines.append("-- 400 original listening questions across 8 sub-tasks (50 items each):")
    lines.append("--   1. Summarize Spoken Text (50 items, q-pl90-sst-01..50)")
    lines.append("--   2. MCQ Multiple Answer (50 items, q-pl90-mcma-01..50)")
    lines.append("--   3. Listening Fill in the Blanks (50 items, q-pl90-fib-01..50)")
    lines.append("--   4. Highlight Correct Summary (50 items, q-pl90-hcs-01..50)")
    lines.append("--   5. MCQ Single Answer (50 items, q-pl90-mcsa-01..50)")
    lines.append("--   6. Select Missing Word (50 items, q-pl90-smw-01..50)")
    lines.append("--   7. Highlight Incorrect Word (50 items, q-pl90-hiw-01..50)")
    lines.append("--   8. Write from Dictation (50 items, q-pl90-wfd-01..50)")
    lines.append("")
    
    all_items = (
        SST_ITEMS +
        LMCMA_ITEMS +
        FIB_ITEMS +
        HCS_ITEMS +
        LMCSA_ITEMS +
        SMW_ITEMS +
        HIW_ITEMS +
        WFD_ITEMS
    )
    
    lines.append("INSERT INTO questions (")
    lines.append("    id, exam_version_id, exam, supported_exams, skill, type_id, type_name,")
    lines.append("    title, prompt, context_passage, audio_transcript, prep_time_seconds,")
    lines.append("    time_limit_seconds, options, correct_answers, blanks, model_answer,")
    lines.append("    explanation, difficulty, points, tags, is_published")
    lines.append(") VALUES")
    
    value_rows = []
    for it in all_items:
        tags = ["PTE", "Original practice", it["type_name"]]
        row_fields = [
            sql_str(it["id"]),
            "'pte-2026-01'",
            "'PTE'",
            "ARRAY['PTE']::text[]",
            "'listening'",
            sql_str(it["type_id"]),
            sql_str(it["type_name"]),
            sql_str(it["title"]),
            sql_str(it["prompt"]),
            sql_str(it.get("context_passage")),
            sql_str(it.get("audio_transcript")),
            str(it.get("prep_time_seconds", 7)),
            str(it.get("time_limit_seconds", 90)),
            sql_json(it.get("options")),
            sql_json(it.get("correct_answers")),
            sql_json(it.get("blanks")),
            sql_str(it.get("model_answer")),
            sql_str(it.get("explanation")),
            "'medium'",
            str(it.get("points", 10)),
            sql_array(tags),
            "true"
        ]
        value_rows.append(f"    ({', '.join(row_fields)})")
        
    lines.append(",\n".join(value_rows) + "\nON CONFLICT (id) DO NOTHING;\n")
    return "\n".join(lines)

def generate_down_sql():
    return "-- Migration 000090 Down: No-op down migration\nSELECT 1;\n"

def main():
    up_sql = generate_up_sql()
    down_sql = generate_down_sql()
    
    migrations_dir = REPO_ROOT / "migrations"
    up_path = migrations_dir / "000090_pte_listening_bank_3.up.sql"
    down_path = migrations_dir / "000090_pte_listening_bank_3.down.sql"
    
    with open(up_path, "w", encoding="utf-8", newline="\n") as f:
        f.write(up_sql)
        
    with open(down_path, "w", encoding="utf-8", newline="\n") as f:
        f.write(down_sql)
        
    up_bytes = up_path.read_bytes()
    up_sha = hashlib.sha256(up_bytes).hexdigest()
    
    down_bytes = down_path.read_bytes()
    down_sha = hashlib.sha256(down_bytes).hexdigest()
    
    print("\n--- Summary Statistics ---")
    print(f"Total questions generated: 400")
    print(f"Wrote {up_path} ({len(up_bytes):,} bytes, sha256={up_sha})")
    print(f"Wrote {down_path} ({len(down_bytes):,} bytes, sha256={down_sha})")

if __name__ == "__main__":
    main()
