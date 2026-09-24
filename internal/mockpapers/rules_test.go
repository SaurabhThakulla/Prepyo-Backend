package mockpapers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prepyo/backend/internal/models"
)

func TestCheckOpenPaper(t *testing.T) {
	open := "paper-2"
	if err := CheckOpenPaper("", &open); err != nil {
		t.Fatalf("no test asked for resumes the open one: %v", err)
	}
	if err := CheckOpenPaper("paper-2", &open); err != nil {
		t.Fatalf("asking for the open test resumes it: %v", err)
	}
	if err := CheckOpenPaper("paper-5", &open); !errors.Is(err, ErrOtherPaperOpen) {
		t.Fatalf("asking for another test while one is open: %v", err)
	}
	if err := CheckOpenPaper("paper-5", nil); !errors.Is(err, ErrOtherPaperOpen) {
		t.Fatalf("an open attempt from before numbered tests still blocks another: %v", err)
	}
}

func TestCheckModule(t *testing.T) {
	for _, tc := range []struct {
		paper, learner string
		ok             bool
	}{
		{ModuleAny, "academic", true},
		{"", "general_training", true},
		{ModuleAcademic, "academic", true},
		{ModuleGeneralTraining, "academic", false},
		{ModuleAcademic, "general_training", false},
	} {
		if err := CheckModule(tc.paper, tc.learner); (err == nil) != tc.ok {
			t.Errorf("paper %q, learner %q: %v", tc.paper, tc.learner, err)
		}
	}
}

func TestWriteErrorMapsPaperErrors(t *testing.T) {
	for err, status := range map[error]int{
		ErrPaperNotFound: 404, ErrPaperNotPublished: 404, ErrPaperRetired: 409,
		ErrOtherPaperOpen: 409, ErrWrongModule: 409,
	} {
		rec := httptest.NewRecorder()
		if !WriteError(rec, err) || rec.Code != status {
			t.Errorf("%v: status %d, want %d", err, rec.Code, status)
		}
	}
	if WriteError(httptest.NewRecorder(), errors.New("something else")) {
		t.Error("an unrelated error is left to the caller")
	}
}

// The builder's advisory lock must be released when a build ends, however it
// ends. Taken on one pooled connection and released on another, it used to
// stay held, and the scope never built again.
func TestBuildReleasesItsLock(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	b := NewBuilder(pool, NewRepository(pool), nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	scope := Scope{Exam: ExamPTE, Section: "lock_test_section", Module: ModuleAny}

	for i := 0; i < 3; i++ {
		// No PTE composer is configured, so the build fails after taking the lock.
		if err := b.BuildScope(ctx, scope, 1); err == nil {
			t.Fatal("a build with no composer should fail")
		}
		var free bool
		if err := pool.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lockKey(scope)).Scan(&free); err != nil {
			t.Fatal(err)
		}
		if !free {
			t.Fatalf("run %d: the scope is still locked after its build ended", i+1)
		}
		// Released on the same connection? Not guaranteed through the pool, so
		// drop every session lock this backend holds.
		if _, err := pool.Exec(ctx, "SELECT pg_advisory_unlock_all()"); err != nil {
			t.Fatal(err)
		}
	}
}

// A PTE attempt waiting on its score is finished for the learner: the test
// reads completed, and a default start moves on to the next test.
func TestCatalogCountsScoringAsCompleted(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)
	section := fmt.Sprintf("pte_scoring_%d", time.Now().UnixNano())
	user := newTestLearner(t, pool, models.ExamPTE)
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM pte_mock_sessions WHERE user_id = $1`, user.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM mock_paper_counters WHERE section = $1`, section)
		_, _ = pool.Exec(ctx, `UPDATE mock_papers SET status = 'retired', retired_at = now() WHERE section = $1`, section)
	})
	paper, err := repo.Publish(ctx, DraftPaper{
		Exam: ExamPTE, Section: section, Module: ModuleAny, Title: "PTE Test 1", ContentSchema: 1,
		Content: json.RawMessage(`{"items":[{"questionId":"q1","task":"RA","part":"speaking_writing"}],"missing":[]}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO pte_mock_sessions (user_id, kind, mock_id, exam_version_id, total_items, paper_id, status)
		VALUES ($1, 'speaking', 'mock-pte-speaking', 'pte-2026-01', 1, $2, 'scoring')`, user.ID, paper.ID); err != nil {
		t.Fatal(err)
	}
	cat, err := NewCatalogService(pool, repo, nil).Catalog(ctx, user, ExamPTE, section, ModuleAny)
	if err != nil {
		t.Fatal(err)
	}
	if len(cat) != 1 || cat[0].Status != "completed" || cat[0].Score != nil {
		t.Fatalf("a paper being scored: %+v", cat)
	}
}
