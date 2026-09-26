"""Export every question, passage, listening test and speaking set in a fully
migrated database to scripts/bank_snapshot.json, so a new content batch can be
checked against everything already in the bank (scripts/bank_check.py).

Run against a disposable local database that has had every migration applied:
    python scripts/export_bank_snapshot.py "postgres://postgres@localhost:5432/prepyo_b4_base?sslmode=disable"
Requires psql on PATH.
"""
import json
import subprocess
import sys
from pathlib import Path

OUT = Path(__file__).with_name('bank_snapshot.json')


def rows(url, sql):
    out = subprocess.run(['psql', url, '-At', '-c', f"SELECT coalesce(json_agg(t), '[]') FROM ({sql}) t"],
                         check=True, capture_output=True, text=True, encoding='utf-8').stdout
    return json.loads(out)


def main(url):
    snapshot = {
        'questions': rows(url, """
            SELECT id, exam, skill, type_id, type_name, title,
                   concat_ws(' ', prompt, context_passage, audio_transcript, model_answer) AS text
              FROM questions ORDER BY id"""),
        'reading_passages': rows(url, """
            SELECT id, title, topic, paragraphs::text AS text FROM reading_passages ORDER BY id"""),
        'reorder_items': rows(url, """
            SELECT id, exam, title, topic, paragraphs::text AS text FROM reading_reorder_items ORDER BY id"""),
        'listening_parts': rows(url, """
            SELECT p.id, t.title AS test_title, p.part_no, p.setting AS title, p.script AS text
              FROM listening_parts p JOIN listening_tests t ON t.id = p.test_id ORDER BY p.id"""),
        'speaking_sets': rows(url, """
            SELECT id, title, content::text AS text FROM speaking_mock_sets ORDER BY id"""),
    }
    OUT.write_text(json.dumps(snapshot, ensure_ascii=False, indent=0), encoding='utf-8', newline='\n')
    print({k: len(v) for k, v in snapshot.items()})


if __name__ == '__main__':
    main(sys.argv[1])
