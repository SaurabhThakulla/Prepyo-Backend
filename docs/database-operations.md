# Database operations: writing and speaking

## Repair and integrity

Migration 000050 inserts 212 missing canonical writing questions and two speaking
questions with the 000048/000049 corrections already applied. Existing rows are
left unchanged. With 000049's four summaries intact the bank is 216 writing and
two speaking questions. This restores starter content, NOT a complete speaking
syllabus. The reason rows disappeared is not established by migration history.

The down migration deliberately does not delete restored rows: they can have
learner references, and there is no safe way to identify which existed beforehand.
Back up before rollout. Test restoration on a disposable database first. Never
replay historical seeds verbatim against an already-migrated environment.

Startup and the migration command check each canonical ID, skill and eligibility;
missing content fails explicitly, including when there are no pending migrations.
This detects drift at startup, not continuous runtime deletion. Monitor startup
errors; investigate and restore through a reviewed new migration, not auto-healing
that could conceal data loss or overwrite editorial work.

## Separate schema owner and API identity

1. Obtain the direct/session connection from the provider dashboard (usually
   port 5432). Do not just change the transaction-pooler port or guess its host.
2. Use it as MIGRATION_DATABASE_URL in a release job running `go run ./cmd/migrate`.
   That job also needs the ordinary config required by pkg/config.
3. Run scripts/database-runtime-role.sql as the schema owner to provision the
   server-only `prepyo_api` role. Set its password securely in the provider or
   interactive psql (`\password prepyo_api`); the script stores no credential.
4. Set API DATABASE_URL to that role's connection and AUTO_MIGRATE=false. Do not
   give MIGRATION_DATABASE_URL to the long-running API in this configuration.
5. Verify login, authoring, metering, evaluation persistence, practice, and jobs
   on staging before switching production. Keep the old configuration for rollback.

The role script grants DML, not schema ownership, TRUNCATE, CREATE or BYPASSRLS.
For RLS-enabled application tables it adds a policy for this server-only role.
The server must access all learners to run the existing repositories/background
jobs; authorization still lives in the API. This is **not per-user RLS isolation**.
Do not distribute this credential to browsers or Supabase anon/authenticated
clients. Existing RLS and policies for other roles are not removed. Audit those
roles separately; adding an API policy does not prove all other grants are safe.
Reapply the role script after schema changes to grant new tables and policies.

## Pooling and credentials

Runtime pgx already replaces the default cached-statement mode with extended
`exec`; this avoids PgBouncer transaction-pooler statement-name collisions and
retains parameterized JSON encoding. Migration file bodies explicitly use the
simple protocol because they are trusted, argument-free multi-statement SQL.
Parameterized repository calls must not be globally switched to simple protocol.

Percent-encode credentials as URL components exactly once. Do not log complete
URLs or put credentials in shell commands, committed files, or tickets. Changing
an ignored local .env does not change deployment secrets. Update deployment
secrets through the hosting provider and rotate exposed credentials separately.

## Tests

Every DB-backed test uses internal/testdb. TEST_DATABASE_URL is mandatory; there
is no fallback to DATABASE_URL or .env. It requires localhost/loopback, rejects
port 6543 and remote fallback hosts, and requires a database name ending `_test`.
Never tunnel production onto a local port: a URL guard cannot detect that.

Create an isolated PostgreSQL cluster/database, apply all migrations, clear any
inherited read-only PGOPTIONS, and run `go test ./...` with TEST_DATABASE_URL.
Without that variable DB tests skip. CI containers should expose their disposable
PostgreSQL on loopback. Delete/stop the disposable database after validation.
