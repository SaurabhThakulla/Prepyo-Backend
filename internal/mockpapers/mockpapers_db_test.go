package mockpapers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/testdb"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := testdb.URL(t)
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping mockpapers tests")
	}
	// The same statement mode as the API's pool (database.Connect), which is
	// stricter about parameter types than pgx's default.
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := database.Migrate(context.Background(), pool, slog.New(slog.NewTextHandler(io.Discard, nil))); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pool
}

func newTestLearner(t *testing.T, pool *pgxpool.Pool, exam models.ExamType) models.User {
	t.Helper()
	var user models.User
	err := pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, referral_code)
		VALUES ('mockpaper-' || gen_random_uuid() || '@test.local', 'MockPaper Test', 'free',
		        'Asia/Kathmandu', $1,
		        'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam`, exam).
		Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID)
	})
	return user
}

func TestMockPapers_ImmutabilityTrigger(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)

	uniqueScope := fmt.Sprintf("speaking_immut_%d", time.Now().UnixNano())
	draft := DraftPaper{
		Exam:          ExamIELTS,
		Section:       uniqueScope,
		Module:        ModuleAny,
		Title:         "Immutable Test 1",
		ContentSchema: 1,
		Content:       json.RawMessage(`{"setId":"set-123"}`),
	}

	paper, err := repo.Publish(ctx, draft)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM mock_paper_counters WHERE section = $1`, uniqueScope)
		_, _ = pool.Exec(ctx, `UPDATE mock_papers SET status = 'retired', retired_at = now() WHERE section = $1`, uniqueScope)
	})

	// Rule 1: Attempt to mutate content of published paper must fail
	_, err = pool.Exec(ctx, `UPDATE mock_papers SET content = '{"setId":"set-hacked"}' WHERE id = $1`, paper.ID)
	if err == nil {
		t.Fatalf("expected error updating content of published paper, got nil")
	}

	// Attempt to mutate number must fail
	_, err = pool.Exec(ctx, `UPDATE mock_papers SET number = 999 WHERE id = $1`, paper.ID)
	if err == nil {
		t.Fatalf("expected error updating number of published paper, got nil")
	}

	// Attempt to delete published paper must fail
	_, err = pool.Exec(ctx, `DELETE FROM mock_papers WHERE id = $1`, paper.ID)
	if err == nil {
		t.Fatalf("expected error deleting published paper, got nil")
	}

	// Retiring must succeed
	if err := repo.Retire(ctx, paper.ID); err != nil {
		t.Fatalf("retire should succeed, got: %v", err)
	}

	// Check status is retired
	p, err := repo.ByID(ctx, paper.ID)
	if err != nil {
		t.Fatalf("fetch by id: %v", err)
	}
	if p.Status != StatusRetired || p.RetiredAt == nil {
		t.Fatalf("expected retired status and non-nil retired_at, got status=%s, retiredAt=%v", p.Status, p.RetiredAt)
	}
}

func TestMockPapers_NumberingAndRevision(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)

	uniqueScope := fmt.Sprintf("writing_rev_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM mock_paper_counters WHERE section = $1`, uniqueScope)
		_, _ = pool.Exec(ctx, `UPDATE mock_papers SET status = 'retired', retired_at = now() WHERE section = $1`, uniqueScope)
	})

	// 1. Publish Test 1
	p1, err := repo.Publish(ctx, DraftPaper{
		Exam:          ExamIELTS,
		Section:       uniqueScope,
		Module:        ModuleAcademic,
		Title:         "Academic Test 1",
		ContentSchema: 1,
		Content:       json.RawMessage(`{"task1Id":"t1-1","task2Id":"t2-1"}`),
	})
	if err != nil {
		t.Fatalf("publish p1: %v", err)
	}
	if p1.Number != 1 || p1.Revision != 1 {
		t.Fatalf("expected number=1, rev=1, got number=%d, rev=%d", p1.Number, p1.Revision)
	}

	// 2. Publish Test 2
	p2, err := repo.Publish(ctx, DraftPaper{
		Exam:          ExamIELTS,
		Section:       uniqueScope,
		Module:        ModuleAcademic,
		Title:         "Academic Test 2",
		ContentSchema: 1,
		Content:       json.RawMessage(`{"task1Id":"t1-2","task2Id":"t2-2"}`),
	})
	if err != nil {
		t.Fatalf("publish p2: %v", err)
	}
	if p2.Number != 2 || p2.Revision != 1 {
		t.Fatalf("expected number=2, rev=1, got number=%d, rev=%d", p2.Number, p2.Revision)
	}

	// 3. Revise Test 1 -> new revision with same number (1), rev (2), supersedes_id = p1.ID
	revisedContent := json.RawMessage(`{"task1Id":"t1-1-fixed","task2Id":"t2-1-fixed"}`)
	p1Rev2, err := repo.Revise(ctx, p1.ID, revisedContent, "Academic Test 1 (Rev 2)")
	if err != nil {
		t.Fatalf("revise p1: %v", err)
	}
	if p1Rev2.Number != 1 || p1Rev2.Revision != 2 {
		t.Fatalf("expected number=1, rev=2, got number=%d, rev=%d", p1Rev2.Number, p1Rev2.Revision)
	}
	if p1Rev2.SupersedesID == nil || *p1Rev2.SupersedesID != p1.ID {
		t.Fatalf("expected supersedesId=%s, got %v", p1.ID, p1Rev2.SupersedesID)
	}

	// Verify old p1 is now retired
	oldP1, err := repo.ByID(ctx, p1.ID)
	if err != nil {
		t.Fatalf("fetch old p1: %v", err)
	}
	if oldP1.Status != StatusRetired {
		t.Fatalf("expected old p1 status=retired, got %s", oldP1.Status)
	}

	// ListPublished should only return p1Rev2 (number 1) and p2 (number 2)
	list, err := repo.ListPublished(ctx, ExamIELTS, uniqueScope, ModuleAcademic)
	if err != nil {
		t.Fatalf("list published: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 published papers, got %d", len(list))
	}
	if list[0].ID != p1Rev2.ID || list[1].ID != p2.ID {
		t.Fatalf("unexpected papers order or ids: %v", list)
	}

	// 4. Retire Test 2 -> leave gap, next paper gets Test 3
	if err := repo.Retire(ctx, p2.ID); err != nil {
		t.Fatalf("retire p2: %v", err)
	}
	p3, err := repo.Publish(ctx, DraftPaper{
		Exam:          ExamIELTS,
		Section:       uniqueScope,
		Module:        ModuleAcademic,
		Title:         "Academic Test 3",
		ContentSchema: 1,
		Content:       json.RawMessage(`{"task1Id":"t1-3","task2Id":"t2-3"}`),
	})
	if err != nil {
		t.Fatalf("publish p3: %v", err)
	}
	if p3.Number != 3 {
		t.Fatalf("expected gap retained and number=3 assigned, got number=%d", p3.Number)
	}
}

func TestMockPapers_ModuleIsolation(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)

	uniqueSection := fmt.Sprintf("reading_mod_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM mock_paper_counters WHERE section = $1`, uniqueSection)
		_, _ = pool.Exec(ctx, `UPDATE mock_papers SET status = 'retired', retired_at = now() WHERE section = $1`, uniqueSection)
	})

	// Academic Test 1
	pAcad, err := repo.Publish(ctx, DraftPaper{
		Exam:          ExamIELTS,
		Section:       uniqueSection,
		Module:        ModuleAcademic,
		Title:         "Academic Test 1",
		ContentSchema: 1,
		Content:       json.RawMessage(`{"passageIds":["p1"]}`),
	})
	if err != nil {
		t.Fatalf("publish academic: %v", err)
	}

	// General Training Test 1 (must be allowed with same number=1 because module differs)
	pGT, err := repo.Publish(ctx, DraftPaper{
		Exam:          ExamIELTS,
		Section:       uniqueSection,
		Module:        ModuleGeneralTraining,
		Title:         "General Training Test 1",
		ContentSchema: 1,
		Content:       json.RawMessage(`{"passageIds":["p2"]}`),
	})
	if err != nil {
		t.Fatalf("publish general training: %v", err)
	}

	if pAcad.Number != 1 || pGT.Number != 1 {
		t.Fatalf("both should have number 1 in their separate modules, got acad=%d, gt=%d", pAcad.Number, pGT.Number)
	}
}

func TestMockPapers_CatalogDerivedStatus(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := NewRepository(pool)
	catalogSvc := NewCatalogService(pool, repo, nil)

	uniqueSection := fmt.Sprintf("pte_cat_%d", time.Now().UnixNano())
	user := newTestLearner(t, pool, models.ExamPTE)

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM pte_mock_sessions WHERE user_id = $1`, user.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM mock_paper_counters WHERE section = $1`, uniqueSection)
		_, _ = pool.Exec(ctx, `UPDATE mock_papers SET status = 'retired', retired_at = now() WHERE section = $1`, uniqueSection)
	})

	// 1. Publish Test 1 (Rev 1)
	p1, err := repo.Publish(ctx, DraftPaper{
		Exam:          ExamPTE,
		Section:       uniqueSection,
		Module:        ModuleAny,
		Title:         "PTE Test 1",
		ContentSchema: 1,
		Content:       json.RawMessage(`{"items":[{"questionId":"q1","task":"RA","part":"speaking_writing"}],"missing":[]}`),
	})
	if err != nil {
		t.Fatalf("publish p1: %v", err)
	}

	// Catalog before any session: should be not_started
	cat, err := catalogSvc.Catalog(ctx, user, ExamPTE, uniqueSection, ModuleAny)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if len(cat) != 1 || cat[0].Status != "not_started" || cat[0].AttemptCount != 0 {
		t.Fatalf("expected 1 paper with status not_started and 0 attempts, got %+v", cat)
	}

	// 2. Start session (in_progress)
	var sessionID string
	err = pool.QueryRow(ctx, `
		INSERT INTO pte_mock_sessions (user_id, kind, mock_id, exam_version_id, total_items, paper_id, status)
		VALUES ($1, 'speaking', 'mock-pte-speaking', 'pte-2026-01', 1, $2, 'in_progress')
		RETURNING id::text`, user.ID, p1.ID).Scan(&sessionID)
	if err != nil {
		t.Fatalf("insert in_progress session: %v", err)
	}

	cat, err = catalogSvc.Catalog(ctx, user, ExamPTE, uniqueSection, ModuleAny)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if cat[0].Status != "in_progress" || cat[0].AttemptCount != 1 {
		t.Fatalf("expected status in_progress and attempt 1, got status=%s, count=%d", cat[0].Status, cat[0].AttemptCount)
	}

	// 3. Mark completed with score 78
	_, err = pool.Exec(ctx, `
		UPDATE pte_mock_sessions
		   SET status = 'completed', completed_at = now(), result = '{"overall": 78}'::jsonb
		 WHERE id = $1`, sessionID)
	if err != nil {
		t.Fatalf("complete session: %v", err)
	}

	cat, err = catalogSvc.Catalog(ctx, user, ExamPTE, uniqueSection, ModuleAny)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if cat[0].Status != "completed" || cat[0].AttemptCount != 1 || cat[0].Score == nil || *cat[0].Score != 78 {
		t.Fatalf("expected status completed with score 78, got %+v", cat[0])
	}
	if cat[0].IsUpdated {
		t.Fatalf("expected IsUpdated=false because revision is same as sat on")
	}

	// 4. Start a second attempt and abandon it
	_, err = pool.Exec(ctx, `
		INSERT INTO pte_mock_sessions (user_id, kind, mock_id, exam_version_id, total_items, paper_id, status, created_at, completed_at)
		VALUES ($1, 'speaking', 'mock-pte-speaking', 'pte-2026-01', 1, $2, 'abandoned', now() + interval '1 minute', now() + interval '2 minutes')`,
		user.ID, p1.ID)
	if err != nil {
		t.Fatalf("insert abandoned session: %v", err)
	}

	cat, err = catalogSvc.Catalog(ctx, user, ExamPTE, uniqueSection, ModuleAny)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	// Rule 7: An abandoned attempt is counted in attempt count, but test keeps completed status and completed score!
	if cat[0].AttemptCount != 2 {
		t.Fatalf("expected attemptCount=2 after abandoned attempt, got %d", cat[0].AttemptCount)
	}
	if cat[0].Status != "completed" {
		t.Fatalf("expected status=completed to persist after subsequent abandonment, got %s", cat[0].Status)
	}
	if cat[0].Score == nil || *cat[0].Score != 78 {
		t.Fatalf("expected score to remain 78 from latest completed attempt, got %v", cat[0].Score)
	}

	// 5. Revise paper to Rev 2 -> learner's card should now show IsUpdated = true!
	_, err = repo.Revise(ctx, p1.ID, json.RawMessage(`{"items":[{"questionId":"q1-fixed","task":"RA","part":"speaking_writing"}],"missing":[]}`), "PTE Test 1 Rev 2")
	if err != nil {
		t.Fatalf("revise p1: %v", err)
	}

	cat, err = catalogSvc.Catalog(ctx, user, ExamPTE, uniqueSection, ModuleAny)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	if !cat[0].IsUpdated {
		t.Fatalf("expected IsUpdated=true after paper revision > sat attempt revision, got false")
	}
}

func TestMockPapers_OverlapExceeded(t *testing.T) {
	existing := []MockPaper{
		{
			ID:            "paper-1",
			Exam:          ExamIELTS,
			Section:       SectionWriting,
			ContentSchema: 1,
			Content:       json.RawMessage(`{"task1Id":"task-1","task2Id":"task-2"}`),
		},
	}

	// 1 shared out of 2 = 50% overlap -> Allowed (<= 50%)
	cand1 := []string{"task-1", "task-3"}
	if ExceedsOverlap(cand1, existing) {
		t.Fatalf("50%% overlap should be allowed, got ExceedsOverlap=true")
	}

	// 2 shared out of 2 = 100% overlap -> Refused (> 50%)
	cand2 := []string{"task-1", "task-2"}
	if !ExceedsOverlap(cand2, existing) {
		t.Fatalf("100%% overlap should be refused, got ExceedsOverlap=false")
	}

	// 3 items: 2 shared = 66.7% overlap -> Refused (> 50%)
	existing3 := []MockPaper{
		{
			ID:            "paper-3",
			Exam:          ExamPTE,
			Section:       SectionSpeaking,
			ContentSchema: 1,
			Content:       json.RawMessage(`{"items":[{"questionId":"q1"},{"questionId":"q2"},{"questionId":"q3"}]}`),
		},
	}
	cand3 := []string{"q1", "q2", "q4"}
	if !ExceedsOverlap(cand3, existing3) {
		t.Fatalf("66.7%% overlap should be refused, got ExceedsOverlap=false")
	}
}
