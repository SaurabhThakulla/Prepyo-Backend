# Numbered Mock Tests Architecture

## Overview

Numbered mock tests (Test 1, 2, 3...) provide learners with permanent, reproducible, directly comparable mock test papers across every sectional test in both IELTS and PTE:
- **IELTS**: Listening, Speaking, Reading, Writing (Academic and General Training)
- **PTE**: Speaking, Writing, Reading, Listening, Full Mock

A test's content is fixed upon publication, its test number is permanent, and retakes of a test present identical questions so that score progression is directly comparable.

---

## Core Rules

1. **Published papers are immutable.**
   A database trigger (`trg_mock_papers_immutable`) rejects any change to `content`, `number`, `exam`, `section`, or `module` on a published paper. The only allowed modification to a published row is marking `status = 'retired'` and recording `retired_at`. Deletions of published rows are blocked (`trg_mock_papers_no_delete`). Any correction to a published paper must be published as a new **revision**.

2. **No generation in the catalog GET.**
   The `GET /api/v1/mock-papers` endpoint is strictly read-only. Papers are assembled by a background worker (`Builder`), triggered at API startup, by an admin request, or when the catalog notices that a section has fewer published papers than its target (default 10). The builder operates off the HTTP request path under PostgreSQL advisory locks (`pg_try_advisory_lock`), and the GET endpoint returns what currently exists.

3. **`module` is `NOT NULL`.**
   The column uses values `'any' | 'academic' | 'general_training'` (default `'any'`). This ensures composite unique constraints hold reliably under PostgreSQL without NULL-handling ambiguities.

4. **Numbers are assigned in a transaction and never renumbered.**
   Numbers are allocated per scope (`exam, section, module`) using a dedicated counter table (`mock_paper_counters`). The counter row is locked `FOR UPDATE` within the publishing transaction. If a test is retired, its number is retained (creating a gap), and numbers are never reused.

5. **Versioning & Integrity.**
   - `revision` (1, 2, ...): A revised paper is a new row with the **same number** and an incremented revision. Only one published revision per test number is allowed at any time (enforced by a partial unique index).
   - `supersedes_id`: References the preceding revision that this paper replaced.
   - `content_schema`: Specifies the JSON shape of `content`.
   - `content_hash`: Deterministic SHA-256 of the content JSON, unique per scope (`exam, section, module`), preventing duplicate paper creation.
   - Sessions reference the exact revision sat on via `paper_id`. Past attempts always review against their original revision.

6. **Question reuse & overlap constraints.**
   When building papers:
   - Slots are filled prioritizing unused questions first, then least-used questions across published papers in that scope.
   - Overlap exceeding **50%** with any existing published paper of the same scope is rejected.
   - For PTE full mocks, the paper must cover all three parts (Speaking & Writing, Reading, Listening). If any part cannot be populated, the paper is refused.

7. **User status is derived from sessions.**
   The `mock_papers` table stores no per-user data. Learner status is computed per test from session tables (`listening_mock_sessions`, `speaking_mock_sessions`, `writing_mock_sessions`, `reading_mock_sessions`, `pte_mock_sessions`).
   - States shown: `not_started`, `in_progress`, `completed`.
   - Abandoned sessions increment the attempt count but are never shown as a separate status. If an attempt is abandoned after a completed attempt, the test remains in `completed` state showing the latest completed score.
   - An "Updated" badge is displayed if the published paper revision is higher than the revision sat on in the learner's last attempt.

8. **`paper_id` on every attempt.**
   Every start (first attempt, retake, or resume) sets or keeps `paper_id`. If `paperId` is omitted in the request (e.g. legacy clients or quick start), the engine selects the lowest-numbered published test the learner has not yet completed.

---

## Schema & Tables

### `mock_papers`
- `id` (TEXT PRIMARY KEY)
- `exam` (TEXT NOT NULL, e.g. `'ielts'`, `'pte'`)
- `section` (TEXT NOT NULL, e.g. `'speaking'`, `'writing'`, `'reading'`, `'listening'`, `'full'`)
- `module` (TEXT NOT NULL DEFAULT `'any'`)
- `number` (INT NOT NULL)
- `revision` (INT NOT NULL DEFAULT 1)
- `supersedes_id` (TEXT REFERENCES mock_papers(id))
- `status` (`'draft' | 'published' | 'retired'`)
- `title` (TEXT NOT NULL)
- `content_schema` (INT NOT NULL DEFAULT 1)
- `content` (JSONB NOT NULL)
- `content_hash` (TEXT NOT NULL)
- `created_at`, `published_at`, `retired_at` (TIMESTAMPTZ)

### `mock_paper_counters`
- `(exam, section, module)` (PRIMARY KEY)
- `next_number` (INT NOT NULL)

---

## Content Shapes (`content_schema = 1`)

- **IELTS Writing**: `{"task1Id": "...", "task2Id": "..."}`
- **IELTS Reading**: `{"passageIds": [...], "questionIds": [...], "reorderIds": [...]}`
- **IELTS Listening**: `{"testId": "..."}`
- **IELTS Speaking**: `{"setId": "..."}`
- **PTE (Sectional & Full)**:
  ```json
  {
    "items": [
      { "questionId": "...", "task": "RA", "part": "speaking_writing" },
      { "questionId": "...", "task": "RO", "part": "reading" }
    ],
    "missing": []
  }
  ```

---

## Endpoints

### Learner Endpoints
- `GET /api/v1/mock-papers?exam=ielts&section=speaking&module=academic`
  Returns the numbered list of published tests with the learner's derived status, score, attempt count, and `isUpdated` flag.
  A completed IELTS test also carries `attemptId`, the `mock_attempts` row its Review opens (`GET /api/v1/mocks/attempts/{id}`).
  A PTE test's Review opens its session's report instead.
- Each engine's start (`POST /listening/mocks`, `/speaking/mocks`, `/writing/mocks`, `/reading/mocks`, `/pte/mocks`) takes an optional `paperId`.

### Admin Endpoints (Require Admin)
- `POST /api/v1/admin/mock-papers/build`
  Payload: `{"exam": "ielts", "section": "writing", "module": "academic", "count": 10}`
  Enqueues background generation; returns `202 Accepted`.
- `POST /api/v1/admin/mock-papers/{id}/revise`
  Payload: `{"content": {...}, "title": "Optional new title"}`
  Retires old revision and publishes new revision with the same test number.
- `POST /api/v1/admin/mock-papers/{id}/retire`
  Retires paper.

## Starting a test

- **One open test per section.** Asking for a test while another test of the section is open is refused with `409`, never answered by resuming the other one. Asking for the open test, or for no test, resumes it (`mockpapers.CheckOpenPaper`).
- **Module.** An IELTS Reading or Writing paper for one module is refused (`409`) to a learner of the other (`mockpapers.CheckModule`).
- **Unknown, draft and retired papers.** An unknown or unpublished paper is `404`; a retired one is `409`. Past attempts on a retired paper still review.
- **A PTE attempt being scored counts as completed.** It shows as completed in the list, and a start without `paperId` moves on to the next test.
- **Deleted questions.** A question deleted from the bank after its paper was published is left out of new attempts on that paper. A paper with a whole part left empty cannot be started (`ErrBankTooSmall`).
- **The builder's lock.** Its `pg_try_advisory_lock` is taken and released on one dedicated connection, because a session lock released on another pooled connection stays held.
