package users

import (
	"testing"
	"time"

	"github.com/prepyo/backend/internal/models"
)

func TestExamIsFixedOnceOnboardingIsDone(t *testing.T) {
	done := time.Now()

	cases := []struct {
		name string
		user models.User
		next models.ExamType
		want bool
	}{
		{"choosing during onboarding", models.User{TargetExam: models.ExamPTE}, models.ExamIELTS, true},
		{"switching after onboarding", models.User{TargetExam: models.ExamPTE, OnboardingCompletedAt: &done}, models.ExamIELTS, false},
		{"resending the same exam", models.User{TargetExam: models.ExamIELTS, OnboardingCompletedAt: &done}, models.ExamIELTS, true},
		{"an admin switching", models.User{Role: models.RoleAdmin, TargetExam: models.ExamPTE, OnboardingCompletedAt: &done}, models.ExamIELTS, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := examChangeAllowed(tc.user, tc.next); got != tc.want {
				t.Fatalf("examChangeAllowed = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOnboardingAnswersAreChecked(t *testing.T) {
	text := func(value string) *string { return &value }
	score := func(value float64) *float64 { return &value }
	minutes := func(value int) *int { return &value }

	t.Run("a complete, valid set", func(t *testing.T) {
		params := UpdateProfileParams{}
		problems := map[string]string{}
		applyOnboardingAnswers(updateRequest{
			StudyGoal:     text("university"),
			Destination:   text("new-zealand"),
			PriorAttempt:  text("retake"),
			PreviousScore: score(6.5),
			FocusSkill:    text("writing"),
			DailyMinutes:  minutes(45),
		}, models.ExamIELTS, &params, problems)

		if len(problems) > 0 {
			t.Fatalf("unexpected problems: %v", problems)
		}
		if *params.StudyGoal != "university" || *params.Destination != "new-zealand" || *params.PriorAttempt != "retake" ||
			*params.PreviousScore != 6.5 || *params.FocusSkill != "writing" || *params.DailyMinutes != 45 {
			t.Fatalf("answers not carried into params: %+v", params)
		}
	})

	t.Run("values outside the lists", func(t *testing.T) {
		params := UpdateProfileParams{}
		problems := map[string]string{}
		applyOnboardingAnswers(updateRequest{
			StudyGoal:     text("holiday"),
			Destination:   text("mars"),
			PriorAttempt:  text("third"),
			PreviousScore: score(75),
			FocusSkill:    text("grammar"),
			DailyMinutes:  minutes(2),
		}, models.ExamIELTS, &params, problems)

		for _, field := range []string{"studyGoal", "destination", "priorAttempt", "previousScore", "focusSkill", "dailyMinutes"} {
			if problems[field] == "" {
				t.Errorf("expected a problem for %s", field)
			}
		}
		if params.StudyGoal != nil || params.PreviousScore != nil || params.DailyMinutes != nil {
			t.Fatalf("invalid answers leaked into params: %+v", params)
		}
	})
}
