package reading

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
)

// Database-backed integration tests for reading mock composition and passage eligibility.
// Run with:
//   TEST_DATABASE_URL=postgres://postgres@localhost:5432/prepyo?sslmode=disable go test ./internal/reading/

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed reading tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func testService(t *testing.T, pool *pgxpool.Pool) *Service {
	t.Helper()
	return NewService(pool, NewRepository(pool), questions.NewRepository(pool),
		mocks.NewRepository(pool), exams.NewRepository(pool), gamification.NewService(), nil)
}

// newLearner makes a throwaway user and removes them, and everything that
// cascades from them, when the test ends.
func newLearner(t *testing.T, pool *pgxpool.Pool, exam models.ExamType) models.User {
	t.Helper()

	var user models.User
	err := pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, referral_code)
		VALUES ('reading-' || gen_random_uuid() || '@test.local', 'Reading Test', 'free',
		        'Asia/Kathmandu', $1,
		        'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam`, exam).
		Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}

	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID); err != nil {
			t.Errorf("cleanup learner: %v", err)
		}
	})
	return user
}

func TestOnePassageServesBothExams(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	rows, err := pool.Query(ctx, `
		SELECT p.id
		  FROM reading_passages p
		 WHERE EXISTS (SELECT 1 FROM questions q
		                WHERE q.passage_id = p.id AND 'IELTS' = ANY(q.supported_exams))
		   AND EXISTS (SELECT 1 FROM questions q
		                WHERE q.passage_id = p.id AND 'PTE' = ANY(q.supported_exams))`)
	if err != nil {
		t.Fatalf("query shared passages: %v", err)
	}
	defer rows.Close()

	shared := 0
	for rows.Next() {
		shared++
	}
	if shared == 0 {
		t.Fatal("no passage carries questions for both exams; the bank is still split by exam")
	}
}

func TestPassagesCarryNoExam(t *testing.T) {
	pool := testPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	var exists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM information_schema.columns
		                WHERE table_name = 'reading_passages' AND column_name = 'exam')`).
		Scan(&exists); err != nil {
		t.Fatalf("check column: %v", err)
	}
	if exists {
		t.Error("reading_passages.exam is back; a passage has an exam again")
	}

	// rp-time-01 was authored for PTE. An IELTS learner must reach it, because
	// 000027 put IELTS tasks on it.
	ieltsGroup, err := repo.PickPracticeGroup(ctx, newLearner(t, pool, models.ExamIELTS).ID,
		models.ExamIELTS, []string{TypeMatchingInformation})
	if err != nil {
		t.Fatalf("pick IELTS matching information: %v", err)
	}
	if ieltsGroup.ID == "" {
		t.Fatal("no group dealt to IELTS")
	}

	// And a PTE learner must reach the passages that were authored for IELTS.
	pteGroup, err := repo.PickPracticeGroup(ctx, newLearner(t, pool, models.ExamPTE).ID,
		models.ExamPTE, []string{TypeMCQSingle})
	if err != nil {
		t.Fatalf("pick PTE multiple choice: %v", err)
	}
	if pteGroup.PassageID == "" {
		t.Fatal("no group dealt to PTE")
	}
}

func TestPassageIndexIsPerExam(t *testing.T) {
	pool := testPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	for _, exam := range []models.ExamType{models.ExamIELTS, models.ExamPTE} {
		list, total, err := repo.ListPassages(ctx, ListPassagesParams{Exam: exam, Limit: 50})
		if err != nil {
			t.Fatalf("list %s passages: %v", exam, err)
		}
		if total == 0 || len(list) == 0 {
			t.Errorf("%s index is empty; the filter is not finding shared passages", exam)
		}

		for _, passage := range list {
			var sets bool
			if err := pool.QueryRow(ctx, `
				SELECT EXISTS (SELECT 1 FROM questions q
				                WHERE q.passage_id = $1 AND q.is_published
				                  AND $2 = ANY(q.supported_exams))`, passage.ID, exam).Scan(&sets); err != nil {
				t.Fatalf("check eligibility: %v", err)
			}
			if !sets {
				t.Errorf("%s index lists %s, on which %s sets nothing", exam, passage.ID, exam)
			}
		}
	}
}

func TestPracticeDealsOnlyEligibleQuestions(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	ctx := context.Background()

	for _, tc := range []struct {
		exam   models.ExamType
		typeID string
	}{
		{models.ExamIELTS, TypeTrueFalse},
		{models.ExamPTE, TypeMCQSingle},
		{models.ExamPTE, TypeFillBlanksRW},
	} {
		user := newLearner(t, pool, tc.exam)
		set, err := svc.PracticeSet(ctx, user, PracticeParams{Exam: tc.exam, TypeID: tc.typeID})
		if err != nil {
			t.Fatalf("%s %s: %v", tc.exam, tc.typeID, err)
		}

		for _, group := range set.Groups {
			if len(group.Questions) == 0 {
				t.Errorf("%s %s: dealt an empty task set", tc.exam, tc.typeID)
			}
			for _, q := range group.Questions {
				var supported []string
				if err := pool.QueryRow(ctx, `SELECT supported_exams FROM questions WHERE id = $1`, q.ID).
					Scan(&supported); err != nil {
					t.Fatalf("read eligibility: %v", err)
				}
				found := false
				for _, e := range supported {
					found = found || e == string(tc.exam)
				}
				if !found {
					t.Errorf("%s practice dealt %s, which supports %v", tc.exam, q.ID, supported)
				}
			}
		}
	}
}

func TestIELTSPaperKeepsItsShape(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	repo := NewRepository(pool)
	ctx := context.Background()

	blueprint, err := repo.GeneratedBlueprint(ctx, models.ExamIELTS)
	if err != nil {
		t.Fatalf("IELTS blueprint: %v", err)
	}
	if blueprint.PassageCount != 3 || blueprint.TotalQuestions != 40 {
		t.Errorf("blueprint = %d passages / %d questions, want 3 / 40",
			blueprint.PassageCount, blueprint.TotalQuestions)
	}

	user := newLearner(t, pool, models.ExamIELTS)
	composed, err := svc.compose(ctx, user.ID, models.ExamIELTS, blueprint)
	if err != nil {
		t.Fatalf("compose IELTS paper: %v", err)
	}

	if len(composed.QuestionIDs) != 40 {
		t.Errorf("paper = %d questions, want 40", len(composed.QuestionIDs))
	}
	if len(composed.PassageIDs) != 3 {
		t.Errorf("paper = %d passages, want 3", len(composed.PassageIDs))
	}

	seen := map[string]bool{}
	for _, id := range composed.PassageIDs {
		if seen[id] {
			t.Errorf("passage %s dealt twice into one paper", id)
		}
		seen[id] = true
	}

	// Every question on an IELTS paper has to be one IELTS sets.
	for _, id := range composed.QuestionIDs {
		var ok bool
		if err := pool.QueryRow(ctx,
			`SELECT 'IELTS' = ANY(supported_exams) FROM questions WHERE id = $1`, id).Scan(&ok); err != nil {
			t.Fatalf("read eligibility: %v", err)
		}
		if !ok {
			t.Errorf("IELTS paper carries %s, which IELTS does not set", id)
		}
	}
}

func TestPTEPaperComposesIndependently(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	repo := NewRepository(pool)
	ctx := context.Background()

	blueprint, err := repo.GeneratedBlueprint(ctx, models.ExamPTE)
	if err != nil {
		t.Fatalf("PTE blueprint: %v", err)
	}

	user := newLearner(t, pool, models.ExamPTE)
	composed, err := svc.compose(ctx, user.ID, models.ExamPTE, blueprint)
	if err != nil {
		t.Fatalf("compose PTE paper: %v", err)
	}

	if len(composed.QuestionIDs) != blueprint.TotalQuestions {
		t.Errorf("paper = %d questions, want %d", len(composed.QuestionIDs), blueprint.TotalQuestions)
	}
	if len(composed.ReorderIDs) == 0 {
		t.Error("PTE paper carries no Re-order Paragraphs items; the reorder section did not fill")
	}

	seen := map[string]bool{}
	for _, id := range composed.PassageIDs {
		if seen[id] {
			t.Errorf("passage %s dealt twice into one paper", id)
		}
		seen[id] = true
	}

	for _, id := range composed.QuestionIDs {
		var ok bool
		if err := pool.QueryRow(ctx,
			`SELECT 'PTE' = ANY(supported_exams) FROM questions WHERE id = $1`, id).Scan(&ok); err != nil {
			t.Fatalf("read eligibility: %v", err)
		}
		if !ok {
			t.Errorf("PTE paper carries %s, which PTE does not set", id)
		}
	}
}

func TestAddingAQuestionDoesNotChangeADealtPaper(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	repo := NewRepository(pool)
	ctx := context.Background()

	blueprint, err := repo.GeneratedBlueprint(ctx, models.ExamIELTS)
	if err != nil {
		t.Fatalf("blueprint: %v", err)
	}

	user := newLearner(t, pool, models.ExamIELTS)
	composed, err := svc.compose(ctx, user.ID, models.ExamIELTS, blueprint)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}

	session := Session{
		ReadingMockSession: models.ReadingMockSession{
			ID: "test-session", Exam: models.ExamIELTS, PassageIDs: composed.PassageIDs,
		},
		QuestionIDs: composed.QuestionIDs,
	}

	before, err := svc.hydrate(ctx, session)
	if err != nil {
		t.Fatalf("hydrate before: %v", err)
	}

	var groupID, passageID string
	if err := pool.QueryRow(ctx, `
		SELECT group_id, passage_id FROM questions WHERE id = $1`, composed.QuestionIDs[0]).
		Scan(&groupID, &passageID); err != nil {
		t.Fatalf("read group: %v", err)
	}

	const intruder = "q-test-immutability"
	if _, err := pool.Exec(ctx, `
		INSERT INTO questions (id, exam_version_id, exam, supported_exams, skill, type_id,
			type_name, title, prompt, correct_answers, time_limit_seconds, points,
			difficulty, tags, passage_id, group_id, group_position)
		SELECT $1, exam_version_id, exam, supported_exams, skill, type_id, type_name,
		       'Intruder', 'Added after the paper was dealt', correct_answers,
		       time_limit_seconds, points, difficulty, tags, passage_id, group_id, 99
		  FROM questions WHERE id = $2`, intruder, composed.QuestionIDs[0]); err != nil {
		t.Fatalf("insert intruder: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(), `DELETE FROM questions WHERE id = $1`, intruder); err != nil {
			t.Errorf("cleanup intruder: %v", err)
		}
	})

	after, err := svc.hydrate(ctx, session)
	if err != nil {
		t.Fatalf("hydrate after: %v", err)
	}

	if countQuestions(before) != countQuestions(after) {
		t.Errorf("paper changed from %d to %d questions after a question was added to its passage",
			countQuestions(before), countQuestions(after))
	}
	for _, set := range after.Sets {
		for _, group := range set.Groups {
			for _, q := range group.Questions {
				if q.ID == intruder {
					t.Fatal("a question added after the paper was dealt appeared on it")
				}
			}
		}
	}
}

func TestHydrateIsStableAcrossReads(t *testing.T) {
	pool := testPool(t)
	svc := testService(t, pool)
	repo := NewRepository(pool)
	ctx := context.Background()

	blueprint, err := repo.GeneratedBlueprint(ctx, models.ExamIELTS)
	if err != nil {
		t.Fatalf("blueprint: %v", err)
	}
	user := newLearner(t, pool, models.ExamIELTS)
	composed, err := svc.compose(ctx, user.ID, models.ExamIELTS, blueprint)
	if err != nil {
		t.Fatalf("compose: %v", err)
	}

	session := Session{
		ReadingMockSession: models.ReadingMockSession{
			ID: "test-session", Exam: models.ExamIELTS, PassageIDs: composed.PassageIDs,
		},
		QuestionIDs: composed.QuestionIDs,
	}

	first, err := svc.hydrate(ctx, session)
	if err != nil {
		t.Fatalf("hydrate: %v", err)
	}
	second, err := svc.hydrate(ctx, session)
	if err != nil {
		t.Fatalf("hydrate again: %v", err)
	}

	if !equal(dealtOrder(first), dealtOrder(second)) {
		t.Error("two reads of the same paper returned different question orders")
	}
	if !equal(dealtOrder(first), composed.QuestionIDs) {
		t.Errorf("hydrated order = %v, want the dealt order %v", dealtOrder(first), composed.QuestionIDs)
	}
}

func countQuestions(session models.ReadingMockSession) int {
	n := 0
	for _, set := range session.Sets {
		for _, group := range set.Groups {
			n += len(group.Questions)
		}
	}
	return n
}

func dealtOrder(session models.ReadingMockSession) []string {
	var ids []string
	for _, set := range session.Sets {
		for _, group := range set.Groups {
			for _, q := range group.Questions {
				ids = append(ids, q.ID)
			}
		}
	}
	return ids
}

// A dashboard heading can cover more than one server type, and practice deals
// the whole family from one passage. Dealing a single group meant a learner who
// chose Multiple Choice met the single-answer questions and never saw the
// multiple-answer ones sitting on the same passage.
func TestPracticeDealsEveryGroupOfTheFamily(t *testing.T) {
	pool := testPool(t)
	repo := NewRepository(pool)
	ctx := context.Background()

	family := []string{TypeMCQSingle, TypeMCQMultiple}

	// Not every passage carries both variants, so the passage under test is the
	// one the assertion is actually about rather than whichever the shuffle
	// happened to deal.
	var passageID string
	if err := pool.QueryRow(ctx, `
		SELECT passage_id FROM reading_question_groups
		WHERE type_id = ANY($1)
		GROUP BY passage_id
		HAVING count(DISTINCT type_id) = 2
		LIMIT 1`, family).Scan(&passageID); err != nil {
		t.Skipf("no passage carries both multiple-choice types: %v", err)
	}

	whole, err := repo.GroupsForPassages(ctx, []string{passageID}, family)
	if err != nil {
		t.Fatalf("load the family: %v", err)
	}

	seen := map[string]bool{}
	for _, g := range whole {
		seen[g.TypeID] = true
		if g.PassageID != passageID {
			t.Errorf("group %s belongs to %s, not the passage asked for", g.ID, g.PassageID)
		}
	}
	if !seen[TypeMCQSingle] || !seen[TypeMCQMultiple] {
		t.Errorf("dealt types %v, want both multiple-choice types", seen)
	}

	// Asking for one type still deals only that type, which is what a mock slot
	// relies on.
	single, err := repo.GroupsForPassages(ctx, []string{passageID}, []string{TypeMCQSingle})
	if err != nil {
		t.Fatalf("load one type: %v", err)
	}
	for _, g := range single {
		if g.TypeID != TypeMCQSingle {
			t.Errorf("asked for %s, got %s", TypeMCQSingle, g.TypeID)
		}
	}
	if len(single) >= len(whole) {
		t.Errorf("one type returned %d groups and the family %d; the filter is not narrowing", len(single), len(whole))
	}

	// No filter at all is still every group on the passage.
	all, err := repo.GroupsForPassages(ctx, []string{passageID}, nil)
	if err != nil {
		t.Fatalf("load every group: %v", err)
	}
	if len(all) <= len(whole) {
		t.Errorf("unfiltered returned %d groups, want more than the %d in the family", len(all), len(whole))
	}

	// The anchor a learner is dealt still belongs to the family they chose.
	anchor, err := repo.PickPracticeGroup(ctx, newLearner(t, pool, models.ExamPTE).ID, models.ExamPTE, family)
	if err != nil {
		t.Fatalf("pick multiple choice: %v", err)
	}
	if anchor.TypeID != TypeMCQSingle && anchor.TypeID != TypeMCQMultiple {
		t.Errorf("anchor type = %s, want one of the family", anchor.TypeID)
	}
}
