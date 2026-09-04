# Prepyo backend

Go API for the Prepyo PTE/IELTS practice platform. One binary, PostgreSQL, no
microservices.

## Running it

```bash
docker compose up
```

That starts Postgres, Redis and the API on port 8080. The API applies its own
migrations at startup, so there is no separate migration step.

To run the API directly against the compose database:

```bash
cp backend/.env.example backend/.env
```

then, from `backend/`:

```bash
DATABASE_URL="postgres://prepyo:prepyo_pass@localhost:5432/prepyo_db?sslmode=disable" go run ./cmd/api
```

The API refuses to start if required configuration is missing, and reports
every problem at once rather than one per restart.

```bash
go test ./...
```

## Layout

```text
cmd/api/          main.go starts and stops; app.go wires everything together
internal/
  auth/           registration, login, sessions, RequireUser middleware
  users/          the users table and profile updates
  reqctx/         carries the current user on the request context
  exams/          exam versions (scoring scales)
  questions/      the shared question bank
  scoring/        deterministic grading - the only place a score is decided
  practice/       single-question attempts
  reading/        passages, passage-driven practice, generated reading mocks
  mocks/          full mock exams
  mistakes/       the mistake bank
  evaluations/    AI writing feedback: allowance, dedupe, persistence
  ai/             the only package that talks to a model provider
  progress/       score estimates derived from attempts
  gamification/   XP ledger, streaks, daily missions
  leaderboards/   ranking
  notifications/  in-app messages
  billing/        plans and entitlements
  admin/          operational metrics
  database/       connection pool and migration runner
migrations/       .sql files, embedded into the binary
pkg/
  config/         environment loading with startup validation
  httpx/          JSON responses, error codes, pagination
  logger/         structured logging
```

Most modules follow the same two files: `repository.go` holds the SQL,
`handler.go` holds the HTTP. Where there is real logic between them, it lives in
`service.go`.

## Rules worth knowing before changing anything

**Scores are decided on the server.** A submission carries what the learner
typed or selected and nothing else. `internal/scoring` turns that into a result.
There is no code path where a score arrives from the client.

**An unknown task type fails loudly.** `scoring.Grade` returns `ok=false` rather
than defaulting to full marks, and the handler returns 501. A made-up score in a
learner's history is worse than a visible gap.

**XP is a ledger.** Every award has a `source_key` naming what it pays for, with
a unique index behind it. Awards are keyed by question or mock plus the day, so
repeating a task is still recorded but only pays once per day.

**Nothing derived is stored.** Level comes from XP, the score estimate comes from
attempts, and evaluation usage comes from counting rows. Storing them would let
them drift from the evidence.

**The AI gateway never invents output.** With no API key configured, evaluation
and tutor endpoints return 503. Model replies are validated against the expected
shape and score range before anything is persisted, and sentence-level feedback
is dropped unless it quotes text the learner actually wrote.

**Ownership is enforced in the SQL.** Learner queries carry `user_id` in the
WHERE clause, so another learner's id matches no rows instead of relying on a
check that someone can forget to write.

**A reading passage is spent when it is dealt, not when it is finished.**
`internal/reading` records the exposure at the moment a paper is handed over, so
a learner who opens a mock and closes the tab does not get the same three
passages next time. A generated paper stores the exact question ids it dealt and
grades only those, which is why extra answers in a submission cannot widen it.

## Not built yet

These are known gaps, not oversights:

- **Payments.** `POST /subscriptions/checkout` returns 501. Wiring it up needs a
  provider, a signature-verified webhook, and only then a plan change. It
  deliberately does not grant a plan for free.
- **Speaking evaluation.** Needs audio capture and object storage first. There is
  no endpoint that pretends to score a recording.
- **Background jobs.** Writing evaluation runs synchronously while the learner
  waits. Redis is in compose for this but is not used yet.
- **Rate limiting is per-instance.** In-memory, so it does not hold across
  replicas. Redis would fix that.
- **The reading bank holds three passages.** A generated mock needs three, so
  the first one a learner sits uses all of them and the second has to repeat.
  The response says so (`reusedPassages`) rather than hiding it, but the real
  fix is more passages: see `migrations/000008_reading_seed.up.sql` for the
  shape one has to have to be eligible.
- **Reading mocks are IELTS only.** The code is exam-agnostic; what is missing
  is a PTE blueprint row and PTE passages.
