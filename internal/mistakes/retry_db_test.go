package mistakes

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/testdb"
	"github.com/prepyo/backend/migrations"
)

func TestRetryRequiresNewCorrectAttempt(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := tx.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`CREATE TEMP TABLE mistakes (id text, user_id text, question_id text, exam text,
 resolved boolean DEFAULT false, last_attempted_at timestamptz) ON COMMIT DROP`)
	exec(`CREATE TEMP TABLE practice_attempts (user_id text, question_id text, exam text,
 is_correct boolean, created_at timestamptz) ON COMMIT DROP`)
	exec(`INSERT INTO mistakes VALUES ('m', 'u', 'q', 'IELTS', false, '2026-09-18 10:00:00Z')`)
	repo := NewRepository(tx)
	reject := func(user string) {
		t.Helper()
		if err := repo.Resolve(ctx, tx, user, "m"); !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected unresolved, got %v", err)
		}
	}
	reject("u")
	for _, sql := range []string{
		`INSERT INTO practice_attempts VALUES ('u','q','IELTS',true,'2026-09-18 09:59:00Z')`,
		`INSERT INTO practice_attempts VALUES ('other','q','IELTS',true,'2026-09-18 10:01:00Z')`,
		`INSERT INTO practice_attempts VALUES ('u','other','IELTS',true,'2026-09-18 10:01:00Z')`,
		`INSERT INTO practice_attempts VALUES ('u','q','PTE',true,'2026-09-18 10:01:00Z')`,
		`INSERT INTO practice_attempts VALUES ('u','q','IELTS',false,'2026-09-18 10:01:00Z')`,
	} {
		exec(sql)
		reject("u")
	}
	exec(`INSERT INTO practice_attempts VALUES ('u','q','IELTS',true,'2026-09-18 10:02:00Z')`)
	reject("other")
	if err := repo.Resolve(ctx, tx, "u", "m"); err != nil {
		t.Fatal(err)
	}
	reject("u") // No duplicate resolution reward.
	exec(`UPDATE mistakes SET resolved=false, last_attempted_at='2026-09-18 10:03:00Z'`)
	reject("u") // A later failure reopens the mistake.
	exec(`INSERT INTO practice_attempts VALUES ('u','q','IELTS',true,'2026-09-18 10:04:00Z')`)
	exec(`INSERT INTO practice_attempts VALUES ('u','q','IELTS',false,'2026-09-18 10:05:00Z')`)
	reject("u") // A superseded correct attempt does not count.
}

func TestRetryTaskMaterialAndOwnership(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	exec := func(sql string) {
		t.Helper()
		if _, err := tx.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	for _, table := range []string{"questions", "reading_passages", "reading_question_groups"} {
		exec("CREATE TEMP TABLE " + table + " (LIKE public." + table + " INCLUDING ALL) ON COMMIT DROP")
	}
	body, err := migrations.Files.ReadFile("000054_ielts_reading_subtasks.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
		t.Fatal(err)
	}
	exec(`CREATE TEMP TABLE mistakes (id text, user_id text, question_id text, exam text) ON COMMIT DROP`)
	exec(`INSERT INTO mistakes SELECT 'm','u',id,'IELTS' FROM questions WHERE group_id='g-ir54-single' LIMIT 1`)
	repo := NewRepository(tx)
	task, err := repo.Task(ctx, "u", "m")
	if err != nil {
		t.Fatal(err)
	}
	if task.Passage == nil || len(task.Passage.Paragraphs) == 0 || task.Group == nil || task.Group.Instructions == "" {
		t.Fatal("retry is missing passage or instructions")
	}
	if len(task.Question.CorrectAnswers) != 0 || task.Question.Explanation != "" || task.Question.ModelAnswer != "" {
		t.Fatal("retry leaked answer key")
	}
	if _, err := repo.Task(ctx, "other", "m"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ownership: %v", err)
	}
	exec(`UPDATE reading_question_groups SET passage_display='hidden'`)
	task, err = repo.Task(ctx, "u", "m")
	if err != nil {
		t.Fatal(err)
	}
	if task.Passage != nil {
		t.Fatal("hidden task exposes clean passage")
	}
	exec(`UPDATE questions SET is_published=false`)
	if _, err := repo.Task(ctx, "u", "m"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unpublished: %v", err)
	}
}
