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

// EPS-TOPIK tests reading and listening only, so onboarding cannot save
// speaking or writing as its focus.
func TestOnboardingFocusFollowsTheExam(t *testing.T) {
	for _, c := range []struct {
		exam  models.ExamType
		skill string
		ok    bool
	}{
		{models.ExamEPSTOPIK, "listening", true},
		{models.ExamEPSTOPIK, "reading", true},
		{models.ExamEPSTOPIK, "speaking", false},
		{models.ExamEPSTOPIK, "writing", false},
		{models.ExamIELTS, "speaking", true},
		{models.ExamPTE, "writing", true},
	} {
		params := UpdateProfileParams{}
		problems := map[string]string{}
		skill := c.skill
		applyOnboardingAnswers(updateRequest{FocusSkill: &skill}, c.exam, &params, problems)
		if ok := problems["focusSkill"] == "" && params.FocusSkill != nil; ok != c.ok {
			t.Errorf("%s focus %s: accepted = %v, want %v (%v)", c.exam, c.skill, ok, c.ok, problems)
		}
	}
}

func TestTargetScoreFollowsTheExamScale(t *testing.T) {
	cases := []struct {
		exam  models.ExamType
		score float64
		ok    bool
	}{
		{models.ExamIELTS, 7.5, true},
		{models.ExamIELTS, 10, false},
		{models.ExamPTE, 79, true},
		{models.ExamPTE, 5, false},
		{models.ExamEPSTOPIK, 60, true},
		{models.ExamEPSTOPIK, 0, true},
		{models.ExamEPSTOPIK, 100, true},
		{models.ExamEPSTOPIK, 120, false},
	}
	for _, c := range cases {
		if got := validTargetScore(c.exam, c.score); got != c.ok {
			t.Errorf("validTargetScore(%s, %v) = %v, want %v", c.exam, c.score, got, c.ok)
		}
	}
}
