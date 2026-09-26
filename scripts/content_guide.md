# Writing a content batch (bank 4 and later)

Every new batch of practice content follows these rules. They exist so that the
bank reads like real exam material and never repeats itself.

## 1. It must read like a real exam paper

- **British English throughout**, as in Cambridge IELTS books and British Council
  material: colour, centre, organise, programme, travelling, licence (noun),
  practise (verb), neighbour, grey, metre, litre, aluminium, jewellery, enrol,
  holiday (not vacation), flat (not apartment), mobile phone, petrol, pavement,
  motorway, timetable, term, marks, CV, queue, lorry, post code, £ prices.
  Dates as "12 March", times as "9.30 am" or "9:30".
- **Plain, specific, natural.** Exam texts state facts, reasons and examples in
  ordinary words. No flourishes, clichés or motivational tone. Vary sentence
  length the way a real writer does.
- **Never use these marks of machine-written text:** delve, tapestry, testament,
  "fast-paced world", "in today's world", ever-evolving, plethora, myriad,
  bustling, vibrant, "navigate the complexities", "it is important to note",
  embark, realm, unlock, seamless, holistic, synergy, paradigm, cutting-edge,
  pivotal, "crucial role", foster, harness, underscore, meticulous, intricate,
  nuanced, "landscape of", "digital age", showcase, elevate, empower, robust,
  leverage, nestled, boasts, "hidden gem", beacon, resonate, and em dashes (—).
  `scripts/bank_check.py` rejects them, with the American spellings it knows.
- **Real, ordinary settings.** Use real countries and cities with invented but
  plausible figures ("The chart below shows the percentage of households with
  internet access in five European countries in 2005 and 2020"), or plain
  descriptions ("a town in the north of England", "a coastal city"). Listening
  uses British names (Mrs Patel, Tom Walker), UK-style addresses and £ prices.
  **No invented fantasy place names** (Oakhaven, Silverdale, Crestmont, Valoria...).
- **Exam topics, exam tone.** Education, work, health, environment, technology,
  cities and transport, family, culture and the arts, travel, science, sport,
  food, media. Avoid what the exams avoid: religion, party politics, war,
  disasters in detail, crime in detail, sex, drugs, death, and anything that
  could upset or disadvantage a candidate.
- **Original.** Written from scratch. Never copy, paraphrase or "recall" any
  Cambridge, British Council, IDP or Pearson material.
- **Exam difficulty.** IELTS roughly band 5.5-8; PTE at the level of a
  first-year university text.
- **Correct answer keys.** Every objective answer is clearly and uniquely
  supported by the text (with the evidence quoted where the generator asks for
  it). Distractors are plausible but definitely wrong. Fill-in answers obey the
  word limit and are spelled exactly as in the text or script.

## 2. It must not repeat anything already in the bank

`scripts/bank_snapshot.json` holds every question, passage, reorder text,
listening script and speaking set in the bank (exported by
`scripts/export_bank_snapshot.py` from a fully migrated database). Every batch
generator runs every new item through `scripts/bank_check.py`:

```python
from scripts.bank_check import Batch
batch = Batch()
batch.question(exam, type_id, title, full_learner_text_and_transcript, where=item_id)
batch.passage(title, full_text, where=passage_id)     # passages, scripts, reorders, sets
batch.style(other_learner_facing_text, where=item_id) # options, explanations, etc.
batch.assert_clean()                                   # before writing any SQL
```

It fails on a same or near-same title in the same task type, heavy overlap of
distinctive words, any shared run of eight words, and the style problems in
section 1. Fix the content; only if a flagged pair is genuinely a different
topic, pass `allowed={...}` for that pair with a comment saying why.

## 3. How a batch is built and tested

- New files only: a data package and a `generate_*_b4*.py` script modelled on
  the batch-3 generator for the same content, writing its own migration pair
  (`.up.sql` with `INSERT ... ON CONFLICT (id) DO NOTHING`, `.down.sql` with
  `SELECT 1;`), Unix line endings, deterministic output.
- IDs use the batch prefix (`ielts-b4-...`, `pte-b4-...`) and never collide.
- Run the generator; every check must pass before SQL is written.
- Test the SQL on a copy of the fully migrated template database:
  ```
  createdb -h localhost -U postgres -T prepyo_b4_base prepyo_b4_<name>_test
  psql -h localhost -U postgres -v ON_ERROR_STOP=1 -d prepyo_b4_<name>_test -f migrations/<file>.up.sql
  ```
  then query the rows back (counts per type, answers, links), and
  `dropdb -h localhost -U postgres prepyo_b4_<name>_test` when done. Never
  connect to any other database.
