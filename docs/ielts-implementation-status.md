# IELTS audit implementation checkpoint

## IELTS conformance fixes (2026-09-23)

Changes made against the IELTS conformance audit. Official references are
the ielts.org format pages, "IELTS scoring in detail", and the public Writing
(May 2023) and Speaking band descriptors.

- Content: 000052 (Listening, Academic Task 1, Speaking drills), 000054 (one
  Reading passage) and 000055 (Speaking pilot) are restored exactly as
  committed before 76e414e. They are Prepyo-original IELTS content that tests
  still read. Other content removed in 76e414e was deliberately not restored.
- Reading paper v2 (000068): the old Academic blueprint could not be filled
  by any passage in the repo. The new shape is still 13/13/14 = 40 one-mark
  questions in 60 minutes, and every original passage can fill any section.
  Ordered sets are put back into text order: official for multiple choice
  and sentence completion, practice convention for TFNG/YNNG. Ordered types
  are never shuffled.
- Reading mock timing (000069): the server keeps the deadline (`expires_at`),
  and the browser saves drafts to `PUT /reading/mocks/{id}/answers`. A
  submission after deadline + 90 s grace is graded from the saved drafts. A
  blank paper gets no band (`not_attempted`); an expired blank paper is
  closed (`paper_expired`). Mock attempts now store their answers.
- General Training (000070, 000071): new columns `users.target_module`,
  `reading_passages.modules`, `reading_mock_blueprints.module` and
  `reading_mock_blueprint_slots.passage_tag`. The GT Reading blueprint is
  everyday texts, workplace texts, then a general-interest article (2,150-2,375
  words), banded on the GT table (anchors 15/23/30/35 -> 4/5/6/7). Content: 12
  Task 1 letters and four original GT Section 1/2 texts, from
  `scripts/generate_ielts_general_training.py`, which checks its own keys.
  Each module is dealt only its own Writing Task 1.
- Scoring: IELTS marking gives one mark per correct answer with no negative
  marking; "Choose TWO" is two marks, and each blank is a mark. Raw scores
  for bands are marks, and any paper size is scaled to 40 before the
  indicative table. Blanks use the same punctuation-insensitive comparison as
  short answers, plus `acceptedAnswers`. PTE marking is unchanged.
- Writing evaluator (writing.v4): IELTS guidance is rewritten from the public
  descriptors in Prepyo's own words, with no population prior and no
  "native-like" demands. Criteria must be whole bands. Task type comes from
  `type_id`. Prepyo's word count (hyphenated words are one word, copied
  prompt runs of 5+ words are excluded) is passed to the model. Responses of
  20 words or fewer are Band 1 on every criterion without a model call.
- Speaking evaluators (speaking.v4, speaking.transcript.v3): accent is
  assessed as intelligibility, not nativeness. The invented pause caps and
  words-per-minute penalties are removed. Content matching against a sample
  answer is disabled for IELTS. A barely heard IELTS answer gets no band;
  it used to get Band 0 with a PTE-style "Content" criterion.
- Overall band: only from all four skill bands, using the official
  .25/.75 rounding. Skill sources:
  - Reading: the latest mock, or practice through the table.
  - Listening: the latest mock (since 2026-09-24), or practice through the
    table.
  - Writing: recent tasks, with Task 2 weighted x2.
  - Speaking: recent answers.
  Until all four exist the estimate is null, with `skillsCovered` and
  `skillsRequired` saying what is missing.
- Listening: IELTS scripts are no longer in question payloads
  (`scriptOnRequest`). They are served at play time from
  `GET /questions/{id}/playback`, and review responses carry the transcript.
  IELTS items play once before submission. Each speaker gets their own
  voice, and speaker labels are not spoken.
- Writing mock (000072, `internal/writingmock`, `/writing/mocks`): one Task 1
  for the module plus one Task 2, 60 minutes on the server clock, autosaved
  drafts. Writing band = RoundIELTSBand((T1 + 2*T2)/3). The 1:2 weighting is
  official; applying the half-band rule to it is Prepyo's convention. A blank
  task is Band 0 (not attempted).

Not done: recorded multi-accent Listening audio (see the 2026-09-24 audio
notes). Human calibration of the AI estimates. Restoring the PTE and legacy
content that the failing DB tests expect (000046-000049, 000053; left to the
owner).

## Section mocks and the full mock (2026-09-24)

- Audio. The browser's own voices (`speechSynthesis`, e.g. Chrome and Edge)
  read Listening recordings and the Speaking examiner. A browser with no
  voices falls back to server clips from Groq TTS: `AI_TTS_MODEL`, default
  `canopylabs/orpheus-v1-english`, with voices in `AI_TTS_VOICES`. Clips are
  at most 200 characters, cached in `speech_clips` (000073), and served at
  `GET /speech/{id}`. The org admin must accept the Orpheus model terms in the
  Groq console first. Until then the endpoint answers `not_configured`. The
  browser player holds each utterance, since Chrome can drop the end event of
  one it has collected. A voice that is listed but never starts (Brave's
  shields can do this) switches to the server clips after about 2 seconds,
  and the rest of that page uses the server clips too. The browser is
  detected by what it can do, never by its name. The start button unlocks
  audio, because iOS Safari only plays sound that starts from a tap.
- Listening mock (000074, 000075, `internal/listeningmock`,
  `/listening/mocks`): four parts and 40 questions in recording order, each
  recording played once. The learner gets reading time before each part and
  2 minutes to check at the end. The server keeps the time and the drafts,
  and `parts_played` stops a reopened paper replaying a part. Marked on the
  Listening table. One original test, `lt-t1`, from
  `scripts/generate_ielts_listening_tests.py`, which checks that every answer
  is heard in order. Mock items are kept out of practice.
- Speaking mock (000076, `internal/speakingmock`, `/speaking/mocks`): Part 1
  has two topics; Part 2 is a cue card with 1 minute to prepare and up to 2
  minutes to talk, plus a rounding-off question; Part 3 is linked to the
  Part 2 topic. Answer time limits are Prepyo's. Each answer is recorded and
  transcribed with Whisper on the server, or with the browser's transcript
  if the server can't. The whole test is rated once (`speaking.test.v1`) on
  three criteria. Pronunciation is not assessed from text, and the result
  says so. A test that runs out of time with fewer than three heard answers
  is closed.
- Full mock (000077, `internal/fullmock`, `/full-mocks`, Mock Center for
  IELTS): Listening, Reading, Writing, then Speaking, in that order. It
  spends one full mock from the plan when it starts (counted from
  `full_mock_sessions`), and its sections cost no sub-tests. Each section is a
  fresh paper marked `in_full_mock`, kept apart from any section mock the
  learner has open. Each section records its own attempt. The full attempt
  (`mock-ielts-full`) stores the four bands, and the overall is the mean
  rounded with the official .25/.75 rule; there is no overall if a section
  has no band. Ending a full mock closes its open section and does not
  refund the allowance.
- Progress: a Listening mock band is now Listening evidence, as a Reading
  mock band already was.
- Auto-submit on time-out runs once in the Listening, Reading and Writing
  runners. It used to retry in a loop after a failure, and in Writing each
  retry asked for a rating again.

Still not done: only one Listening test and four Speaking sets exist, so
repeats come quickly. Recorded human audio for Listening. Pronunciation
rating.

## PTE practice expansion

- Migration 000053 adds 115 original PTE practice tasks: five per subtask for
  23 categories across all four skills, including the current speaking task
  "Respond to a Situation" (prompt versions bumped; syllabus and hub updated).
  Reading tasks are linked into the real passage/question-group/reorder banks
  (20 group-linked questions, 5 reorder items), not standalone rows. All content
  is original; no Pearson material and no recalled exam questions. The
  "Repeated Essay Questions" prompts are labelled original practice prompts.
- Listening uses labelled browser-speech-synthesis playback (no audio_url);
  Select Missing Word uses an explicit pause, not a recorded beep. Timers are
  practice budgets, not official per-item limits. Generated by
  scripts/generate_pte_practice.py (+ pte_practice_content.py), replay-safe
  (ON CONFLICT (id) DO NOTHING preserves edits; down file preserves data).
- Verified on a fresh local DB (prepyo_pte_expansion_test, loopback 55439):
  migration 000001-000053 applies; SQL counts confirm 5 rows per each of the
  23 type_ids; TestPTEExpansionBank checks counts, PTE-only eligibility,
  explanations, blank anchors, transcripts, SVG figures, summary model lengths
  (50-70 words), deterministic grading of every key (100% accuracy, empty
  response scores zero), group/passage/reorder linkage, and replay safety
  (re-running the migration preserves an editorial title change). Backend
  go test ./... green (TEST_DATABASE_URL on the loopback test DB); frontend
  119 tests, tsc --noEmit clean, production build succeeded.
- Not certified: these are practice drills, not full PTE papers; grading is
  deterministic/heuristic (dictation word matching, keyword coverage) or AI
  structural checks, not Pearson scoring; synthetic audio is not authentic
  speech; estimates are practice-only. Human PTE-specialist review and
  browser/media accessibility checks remain outstanding.

## Delivered
# IELTS audit implementation checkpoint

## Delivered

- Raw AI scores are checked before rounding; scored IELTS responses require four
  named criteria, max 9 each, evidence and a consistent equally weighted mean.
  Null scores remain valid for insufficient evidence. Prompt versions bumped.
- Migration 000051 adds mock availability and persisted JSON responses. Existing
  incomplete static blueprints are unavailable without deleting attempt history.
- Canonical mock submission errors are explicitly rejected.
- Reading information matching is no longer called heading matching. Legacy
  sequencing content retains its IDs and is named as legacy practice. The
  generated IELTS blueprint uses sentence completion instead of sequencing;
  existing composition checks still require 3 distinct passages and 40 items.
- Migration 000052 adds 40 original IELTS practice tasks: five each for Task 1
  figures (tables), Task 2 opinion essays, Speaking Parts 1/2/3, listening
  completion, single-choice listening, and map navigation. No reading rows added.
- The deterministic generator is scripts/generate_ielts_practice.py. Inline SVG
  figures/plans do not depend on an external asset host. Listening scripts use
  the existing explicitly labelled browser-synthesised playback.
- Frontend removes payment gating for unavailable mocks, fixes generic single
  selection, and prevents unsupported reading links silently dealing another task.

## Not delivered / not certified

This is NOT a complete four-skill IELTS mock implementation. Writing, listening
and speaking full-section mocks remain unavailable on all plans. The five-item
batches are short practice drills, not five full papers. A two-task writing
session with Task 2 double weighting, full four-part/40-mark listening papers,
examiner-led speaking interactions and authentic recorded audio remain work.
Genuine reading heading-matching content is not present and is not substituted
with paragraph-information or sequencing questions. Other IELTS task families
beyond the current sidebar catalogue are not covered by this content batch.
Missing production reading content was not restored in this change.

AI criterion checks improve structural reliability, not scoring validity. These
are task-level practice estimates, not official bands. Human IELTS-specialist
review and browser/media accessibility checks are still required. Attempts to
fetch official format pages failed; do not describe this as external certification.

## Rollout

Apply migrations 000051, 000052 and 000053 with the migration connection BEFORE
starting this backend version (or use AUTO_MIGRATE). Repositories require the
new columns. 000051 is a one-time schema migration; 000052 and 000053 are
replay-safe and preserve edits. Both down files intentionally preserve data
rather than undo safety changes.

No production migration, deployment, commit or push was performed in this task.
Local validation uses only explicit loopback *_test databases on port 55439.
