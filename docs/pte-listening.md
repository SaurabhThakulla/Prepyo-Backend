# PTE listening with browser TTS

Listening questions continue to use the shared `questions` table. For PTE
listening authoring, `audioTranscript` is required, even if an optional
`audioUrl` is supplied. The learner API retains the transcript for browser TTS;
it is not a secret answer channel and should only be hidden visually until review.

## Type compatibility

Migration 000046 updates the eight affected original seed rows to the admin IDs:

| Legacy ID | Canonical ID |
| --- | --- |
| multiple-choice-multiple | pte-listening-mcma |
| fill-in-the-blanks | pte-listening-fib |
| highlight-correct-summary | pte-highlight-correct-summary |
| select-missing-word | pte-select-missing-word |

Question-list filters accept either ID, including before migration. Admin
create/update accepts legacy IDs for PTE and saves the canonical ID. No question
IDs, option IDs, answer keys, timings, or mock references are changed.
Clients should render canonical IDs; alias filters do not change response IDs.

## Authoring contract

Use `GET /api/v1/admin/questions/types` for the catalogue and
`POST /api/v1/admin/questions` to create questions. New answer kinds are
`summary` and `words`; an admin frontend must render controls for them.

### Fill in the blanks

```json
{
  "exam": "PTE",
  "typeId": "pte-listening-fib",
  "title": "Library hours",
  "audioTranscript": "The library closes at nine.",
  "contextPassage": "The [[b1]] closes at [[b2]].",
  "blanks": [{"correctAnswer": "library"}, {"correctAnswer": "nine"}],
  "publish": false
}
```

The displayed passage is required and retained separately from the spoken
script. Blank keys are numbered b1, b2, ... in submitted nonempty order.
Submit learner answers in `blankResponses`, keyed by those IDs.

### Highlight Incorrect Words

```json
{
  "exam": "PTE",
  "typeId": "pte-highlight-incorrect-word",
  "title": "Library hours",
  "audioTranscript": "The library closes at nine.",
  "contextPassage": "The library opens at nine.",
  "correctAnswers": ["w3"],
  "publish": false
}
```

The server generates ordered `options` from whitespace-delimited displayed
words. IDs are w1, w2, ...; punctuation remains attached. Repeated words have
distinct IDs. Render those options as clickable words and submit their IDs in
`selectedOptions`. Admin detail returns correct word IDs, not their spellings,
so editing preserves repeated-word distinctions. Re-select the answer positions
if the displayed text changes. The display supports 2-500 words. There must be
at least one incorrect word and one unmarked word. Legacy MCQ-style highlighting
content needs reauthoring with a displayed transcript and word-position keys.

### Summarize Spoken Text

Use `typeId: "summarize-spoken-text"`, `audioTranscript`, `modelAnswer`, and
`correctAnswers` containing key concepts (not option IDs). Keywords are trimmed,
lowercased and deduplicated; at least one and a model summary are required.
Learners submit `textResponse`. Existing word-length and keyword grading is
approximate practice feedback, not official PTE scoring or semantic evaluation.

### Dictation and choice tasks

Dictation derives `correct_answers` from the spoken transcript. Do not populate
blanks. Learners submit `textResponse`. MCQ authoring still accepts option text
and correct answer text; the server converts these to option IDs, which learners
submit in `selectedOptions`.

## Frontend work outside this repository

Keep the spoken transcript visually hidden during practice. Start speech from a
user interaction, handle voice availability and playback errors, and cancel
speech on navigation. For Select Missing Word, split at `[beep]` and play a real
beep after speech rather than asking TTS to pronounce the marker. Handle speaker
labels before playback. No browser TTS implementation is present in this backend.
