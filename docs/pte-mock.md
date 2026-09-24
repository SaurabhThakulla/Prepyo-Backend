# PTE mock tests

Package `internal/ptemock` runs every PTE mock — the full test and the four
sectional tests — on one engine, mounted at `/api/v1/pte/mocks`. The browser
side is `prepyo/src/components/pteMock` and `prepyo/src/pages/PteMockPage.tsx`.

## How a paper runs

A paper is taken the way PTE Academic is taken:

- **One item per screen, forward only.** The server holds `current_position`;
  an answer for any other item is refused (`409`). There is no Back.
- **Three kinds of clock** (`Task.Clock`):
  - `response` — speaking items run their own prep and recording windows in
    the player; the recording starts after a beep and stops at the time limit
    or after 3 s of silence.
  - `item` — Summarize Written Text (10 min), Write Essay (20 min), Summarize
    Spoken Text (10 min) each have a countdown of their own.
  - `section` — the rest of Reading, and of Listening, share one countdown per
    part. Its length is the sum of the dealt items' shares, so a shorter paper
    gets proportionally less time.
- **The server keeps the time.** A clock's deadline is set when its first item
  is shown and stored on the session. When it passes, every item left on that
  clock is closed — a saved draft becomes the answer — and the test moves on.
  `SubmitGrace` lets an answer sent at zero arrive; it never opens a new item
  on a clock that has run out.
- **Drafts** (`PUT …/items/{n}/draft`) are saved while the learner types or
  chooses, so a timeout or a reload loses nothing. A paper can be left and
  resumed for 24 hours (`StaleAfter`).

## Scoring

When the last item is done (or the learner ends the test) the paper moves to
`scoring` and is marked in the background (`scorer.go`):

- answer-key items by `scoring.Grade`, as practice marks them;
- written items by `evaluations.EvaluatePTEMockWriting`, after PTE's form rule
  (outside `MinWords`–`MaxWords` scores zero without a rating);
- spoken items by `evaluations.EvaluatePTEMockSpeaking` from the device
  transcript — Read Aloud and Repeat Sentence by word alignment, the rest by
  the model.

Spoken answers are transcribed the way the IELTS speaking mock does it
(`speakingmock.SaveAnswer`). Every spoken answer is sent with its recording
(mp3, or wav if encoding fails) and the browser's own transcript. `Submit`
transcribes the recording with Whisper on the audio provider
(`AI_AUDIO_BASE_URL` / `AI_AUDIO_API_KEY`, Groq) and keeps that; if the call
fails, or no audio provider is configured, the browser's transcript is kept, so
an answer is never lost to one failed call. The recording is used and dropped,
never stored; `pte_mock_items.transcript_source` records `server` or `browser`.
Whisper only runs for the spoken item on screen, so the endpoint is not a free
transcription service. The catalog's `serverTranscription` tells the pre-test
check whether the browser's own recognition matters.

Each mark is saved as it arrives, so a run cut short (a restart) resumes from
where it stopped; a paper stuck in `scoring` for 3 minutes is picked up again
by the next report request. An item whose rating fails twice is left out of the
scores and flagged, not counted as zero.

Skill scores are on the 10–90 scale (`score.go`): each task's items are
averaged, and each communicative skill is the weighted mean of the tasks that
count towards it (`Task.Weights` — integrated, as in PTE: Read Aloud counts for
Speaking and Reading, Write from Dictation for Listening and Writing, …). The
overall score of a full test is the mean of the four skills. It is a practice
estimate; Pearson's own conversion is not published.

## Billing

A full test spends one from the plan's full-mock allowance when it starts
(counted from `pte_mock_sessions` in `billing.State`); a sectional test spends
`billing.SectionMockSubTests` sub-tests. Resuming an open paper is free.

## Changing or extending it

Everything about a test's shape is data in two files:

- `tasks.go` — one `Task` per PTE item type: bank type ids (with legacy
  aliases), timings, on-screen instructions, skill weights, form limits.
- `blueprint.go` — one `Blueprint` per kind: which tasks, how many, in order.

To change item counts, edit a blueprint. To add a kind (a 30-minute "mini
mock", say), add a `Blueprint` and a `mocks` row for its attempts, and add the
kind to the `pte_mock_sessions.kind` check and to
`prepyo/src/lib/pteMock/catalog.ts`. To add a task type, add a `Task`; the
player picks its component by skill, so a new task only needs a new component
if it answers in a new way.

The catalog endpoint (`GET /pte/mocks/catalog`) serves blueprints with what the
bank can actually deal, so the UI never holds a second copy of any of this.
Where the bank has none of a task, papers are dealt without it and the task is
named in `missing`; a part with nothing in it at all refuses to start.
