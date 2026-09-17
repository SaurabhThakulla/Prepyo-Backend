package reading

import (
	"context"
	"github.com/prepyo/backend/internal/testdb"
	"testing"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
)

func TestPTEReadingPracticeDealsAllFiveTasks(t *testing.T) {
	dbURL := testdb.URL(t)
	if dbURL == "" {
		t.Skip("DATABASE_URL not found, skipping integration test")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("database.Connect error: %v", err)
	}

	t.Cleanup(pool.Close)
	user := newLearner(t, pool, models.ExamPTE)

	svc := NewService(pool, NewRepository(pool), questions.NewRepository(pool),
		mocks.NewRepository(pool), exams.NewRepository(pool), gamification.NewService(), nil, nil)

	taskTypes := []string{
		"fill-in-blanks-rw",
		"reading-mcq-multiple",
		"reorder-paragraphs",
		"fill-in-blanks-r",
		"reading-mcq-single",
	}

	for _, typeID := range taskTypes {
		t.Run(typeID, func(t *testing.T) {
			set, err := svc.PracticeSet(ctx, user, PracticeParams{
				Exam:   models.ExamPTE,
				TypeID: typeID,
			})
			if err != nil {
				t.Fatalf("PracticeSet(%s) failed: %v", typeID, err)
			}
			if len(set.Groups) == 0 {
				t.Fatalf("PracticeSet(%s) returned 0 groups", typeID)
			}
			if len(set.Groups[0].Questions) == 0 {
				t.Fatalf("PracticeSet(%s) returned 0 questions", typeID)
			}
			q := set.Groups[0].Questions[0]
			if q.ID == "" {
				t.Errorf("question ID is empty")
			}
			t.Logf("✅ Successfully dealt %s: Question ID=%s Title=%s Points=%d", typeID, q.ID, q.Title, q.Points)
		})
	}
}

func TestPTEMockPaperComposition(t *testing.T) {
	dbURL := testdb.URL(t)
	if dbURL == "" {
		t.Skip("DATABASE_URL not found, skipping integration test")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("database.Connect error: %v", err)
	}

	t.Cleanup(pool.Close)
	user := newLearner(t, pool, models.ExamPTE)

	repo := NewRepository(pool)
	svc := NewService(pool, repo, questions.NewRepository(pool),
		mocks.NewRepository(pool), exams.NewRepository(pool), gamification.NewService(), nil, nil)

	blueprint, err := repo.GeneratedBlueprint(ctx, models.ExamPTE)
	if err != nil {
		t.Fatalf("get PTE blueprint: %v", err)
	}

	composed, err := svc.compose(ctx, user.ID, models.ExamPTE, blueprint)
	if err != nil {
		t.Fatalf("compose PTE mock: %v", err)
	}

	if len(composed.QuestionIDs) != blueprint.TotalQuestions {
		t.Errorf("got %d questions, want %d", len(composed.QuestionIDs), blueprint.TotalQuestions)
	}
	t.Logf("✅ Successfully composed PTE mock paper with %d questions across %d passages and %d reorders",
		len(composed.QuestionIDs), len(composed.PassageIDs), len(composed.ReorderIDs))
}
