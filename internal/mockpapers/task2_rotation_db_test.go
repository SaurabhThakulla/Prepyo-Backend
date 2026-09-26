package mockpapers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/prepyo/backend/migrations"
)

// Migration 000091 gives the numbered IELTS Writing tests the six Task 2 essay
// types in turn, as new revisions. This replays it against a scope of ten
// Opinion-only tests (as every scope was before it) and checks the result.
func TestTask2TypesRotateThroughWritingPapers(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	body, err := migrations.Files.ReadFile("000091_ielts_writing_task2_types.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)

	var perType map[string]int
	rows, err := tx.Query(ctx, `SELECT type_id, count(*) FROM questions
		WHERE type_id LIKE 'ielts-writing-task2-%' AND type_id <> 'ielts-writing-task2-opinion'
		  AND is_published AND 'IELTS' = ANY(supported_exams) GROUP BY type_id`)
	if err != nil {
		t.Fatal(err)
	}
	perType = map[string]int{}
	for rows.Next() {
		var id string
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			t.Fatal(err)
		}
		perType[id] = n
	}
	rows.Close()
	for _, id := range []string{"discussion", "advantages", "problem-solution", "cause-effect", "two-part"} {
		if got := perType["ielts-writing-task2-"+id]; got != 50 {
			t.Errorf("%s questions = %d, want 50", id, got)
		}
	}

	// Ten Opinion tests, except Test 2, which already has the type it is due.
	module := fmt.Sprintf("rotation_%d", time.Now().UnixNano())
	for n := 1; n <= 10; n++ {
		task2 := fmt.Sprintf("ielts-wrt-opinion-%03d", 90+n)
		if n == 2 {
			task2 = "ielts-wrt-discussion-050"
		}
		content := fmt.Sprintf(`{"task1Id":"task1-%d","task2Id":"%s"}`, n, task2)
		if _, err := tx.Exec(ctx, `
			INSERT INTO mock_papers (exam, section, module, number, status, title, content, content_hash, published_at)
			VALUES ('ielts', 'writing', $1, $2, 'published', $3, $4::jsonb, $5, now())`,
			module, n, fmt.Sprintf("Test %d", n), content, HashContent(json.RawMessage(content))); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := tx.Conn().PgConn().Exec(ctx, string(body)).ReadAll(); err != nil {
		t.Fatalf("replay migration: %v", err)
	}

	kinds := []string{"opinion", "discussion", "advantages", "problem-solution", "cause-effect", "two-part"}
	rows, err = tx.Query(ctx, `
		SELECT p.number, p.revision, p.supersedes_id IS NOT NULL, p.content_hash,
		       p.content->>'task1Id', p.content->>'task2Id', q.type_id
		  FROM mock_papers p JOIN questions q ON q.id = p.content->>'task2Id'
		 WHERE p.exam = 'ielts' AND p.section = 'writing' AND p.module = $1 AND p.status = 'published'
		 ORDER BY p.number`, module)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	seen := map[string]bool{}
	count := 0
	for rows.Next() {
		var number, revision int
		var superseded bool
		var hash, task1, task2, typeID string
		if err := rows.Scan(&number, &revision, &superseded, &hash, &task1, &task2, &typeID); err != nil {
			t.Fatal(err)
		}
		count++
		want := "ielts-writing-task2-" + kinds[(number-1)%len(kinds)]
		if typeID != want {
			t.Errorf("Test %d Task 2 is %s, want %s", number, typeID, want)
		}
		if task1 != fmt.Sprintf("task1-%d", number) {
			t.Errorf("Test %d Task 1 changed to %s", number, task1)
		}
		unchanged := number == 1 || number == 2 || number == 7
		if unchanged != (revision == 1) || superseded == unchanged {
			t.Errorf("Test %d revision=%d superseded=%v, want unchanged=%v", number, revision, superseded, unchanged)
		}
		if seen[task2] {
			t.Errorf("Test %d repeats Task 2 %s", number, task2)
		}
		seen[task2] = true
		if !unchanged && !strings.HasPrefix(task2, "ielts-wrt-") {
			t.Errorf("Test %d took an unexpected question %s", number, task2)
		}
		// The hash matches what the Go builder would write for this content.
		content, _ := json.Marshal(WritingContent{Task1ID: task1, Task2ID: task2})
		if hash != HashContent(content) {
			t.Errorf("Test %d content hash does not match the builder's", number)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 10 {
		t.Fatalf("published tests = %d, want 10", count)
	}

	var retired int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM mock_papers
		WHERE exam = 'ielts' AND section = 'writing' AND module = $1 AND status = 'retired'`, module).Scan(&retired); err != nil {
		t.Fatal(err)
	}
	if retired != 7 {
		t.Errorf("retired revisions = %d, want 7", retired)
	}
}
