package reading

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
)

func loadTestEnv() string {
	if url := os.Getenv("TEST_DATABASE_URL"); url != "" {
		return url
	}
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	for _, p := range []string{".env", "../.env", "../../.env"} {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DATABASE_URL=") {
				return strings.Trim(strings.TrimPrefix(line, "DATABASE_URL="), `"'`)
			}
		}
	}
	return ""
}

func TestPTEReadingPracticeDealsAllFiveTasks(t *testing.T) {
	dbURL := loadTestEnv()
	if dbURL == "" {
		t.Skip("DATABASE_URL not found, skipping integration test")
	}

	ctx := context.Background()
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		t.Fatalf("database.Connect error: %v", err)
	}

	user := newLearner(t, pool, models.ExamPTE)
	t.Cleanup(pool.Close)

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
