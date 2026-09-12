package users

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
)

func testRepository(t *testing.T) *Repository {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed user tests")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewRepository(pool)
}

func newTestUser(t *testing.T, repo *Repository) models.User {
	t.Helper()

	suffix := time.Now().UnixNano()
	user, err := repo.Create(context.Background(), CreateParams{
		Email:        fmt.Sprintf("onboarding-%d@example.test", suffix),
		Name:         "Onboarding Check",
		NepalRegion:  "Kathmandu",
		Timezone:     "Asia/Kathmandu",
		ReferralCode: fmt.Sprintf("OB%d", suffix),
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { _ = repo.Delete(context.Background(), user.ID) })
	return user
}

func TestCompletingOnboardingIsRecordedOnce(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	user := newTestUser(t, repo)

	if user.OnboardingCompletedAt != nil {
		t.Fatalf("a new account starts with onboarding done: %v", user.OnboardingCompletedAt)
	}

	untouched, err := repo.UpdateProfile(ctx, user.ID, UpdateProfileParams{})
	if err != nil {
		t.Fatalf("plain update: %v", err)
	}
	if untouched.OnboardingCompletedAt != nil {
		t.Fatalf("an unrelated update finished onboarding")
	}

	first, err := repo.UpdateProfile(ctx, user.ID, UpdateProfileParams{CompleteOnboarding: true})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if first.OnboardingCompletedAt == nil {
		t.Fatalf("completing onboarding stored nothing")
	}

	again, err := repo.UpdateProfile(ctx, user.ID, UpdateProfileParams{CompleteOnboarding: true})
	if err != nil {
		t.Fatalf("complete again: %v", err)
	}
	if again.OnboardingCompletedAt == nil || !again.OnboardingCompletedAt.Equal(*first.OnboardingCompletedAt) {
		t.Fatalf("second completion moved the time: %v then %v", first.OnboardingCompletedAt, again.OnboardingCompletedAt)
	}

	later, err := repo.UpdateProfile(ctx, user.ID, UpdateProfileParams{})
	if err != nil {
		t.Fatalf("later update: %v", err)
	}
	if later.OnboardingCompletedAt == nil {
		t.Fatalf("a later update cleared the completion")
	}
}

func TestOnboardingAnswersAreStored(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()
	user := newTestUser(t, repo)

	goal, place, retake, skill := "university", "uk", "retake", "writing"
	previous, minutes := 6.5, 45

	saved, err := repo.UpdateProfile(ctx, user.ID, UpdateProfileParams{
		StudyGoal:          &goal,
		Destination:        &place,
		PriorAttempt:       &retake,
		PreviousScore:      &previous,
		FocusSkill:         &skill,
		DailyMinutes:       &minutes,
		CompleteOnboarding: true,
	})
	if err != nil {
		t.Fatalf("save answers: %v", err)
	}
	if saved.StudyGoal != goal || saved.Destination != place || saved.PriorAttempt != retake || saved.FocusSkill != skill {
		t.Fatalf("text answers not stored: %+v", saved)
	}
	if saved.PreviousScore == nil || *saved.PreviousScore != previous || saved.DailyMinutes == nil || *saved.DailyMinutes != minutes {
		t.Fatalf("numeric answers not stored: previous=%v minutes=%v", saved.PreviousScore, saved.DailyMinutes)
	}

	reread, err := repo.ByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("reread: %v", err)
	}
	if reread.FocusSkill != skill || reread.DailyMinutes == nil {
		t.Fatalf("answers lost on read: %+v", reread)
	}

	first := "first"
	cleared, err := repo.UpdateProfile(ctx, user.ID, UpdateProfileParams{PriorAttempt: &first})
	if err != nil {
		t.Fatalf("switch to first attempt: %v", err)
	}
	if cleared.PriorAttempt != first || cleared.PreviousScore != nil {
		t.Fatalf("a first attempt kept a previous score: %v", cleared.PreviousScore)
	}
}
