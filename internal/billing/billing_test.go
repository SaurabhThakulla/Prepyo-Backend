package billing

import (
	"testing"
	"time"

	"github.com/prepyo/backend/internal/models"
)

func TestPlanEntitlementDurations(t *testing.T) {
	cases := []struct {
		name              string
		plan              models.Plan
		wantBaseDays      int
		wantBonusDays     int
		wantEffectiveDays int
	}{
		{
			name: "Weekly Plan",
			plan: models.Plan{
				ID:             "weekly",
				Name:           "Weekly Sprint",
				DurationDays:   7,
				BonusDays:      0,
				DurationMonths: 0,
			},
			wantBaseDays:      7,
			wantBonusDays:     0,
			wantEffectiveDays: 7,
		},
		{
			name: "Normal Plan (Pro)",
			plan: models.Plan{
				ID:             "pro",
				Name:           "Pro Prep",
				DurationDays:   30,
				BonusDays:      3,
				DurationMonths: 1,
			},
			wantBaseDays:      30,
			wantBonusDays:     3,
			wantEffectiveDays: 33,
		},
		{
			name: "Max Plan (Elite)",
			plan: models.Plan{
				ID:             "elite",
				Name:           "Elite Master",
				DurationDays:   90,
				BonusDays:      7,
				DurationMonths: 3,
			},
			wantBaseDays:      90,
			wantBonusDays:     7,
			wantEffectiveDays: 97,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base := tc.plan.DurationDays
			bonus := tc.plan.BonusDays
			effective := base + bonus

			if base != tc.wantBaseDays {
				t.Errorf("base days = %d, want %d", base, tc.wantBaseDays)
			}
			if bonus != tc.wantBonusDays {
				t.Errorf("bonus days = %d, want %d", bonus, tc.wantBonusDays)
			}
			if effective != tc.wantEffectiveDays {
				t.Errorf("effective days = %d, want %d", effective, tc.wantEffectiveDays)
			}
		})
	}
}

func TestPlanIsActive(t *testing.T) {
	now := time.Now()
	future := now.Add(48 * time.Hour)
	past := now.Add(-48 * time.Hour)

	cases := []struct {
		name string
		user models.User
		want bool
	}{
		{
			name: "Free plan is always active",
			user: models.User{PlanID: "free", PlanValidUntil: nil},
			want: true,
		},
		{
			name: "Paid plan with future date is active",
			user: models.User{PlanID: "pro", PlanValidUntil: &future},
			want: true,
		},
		{
			name: "Paid plan with past date is expired",
			user: models.User{PlanID: "pro", PlanValidUntil: &past},
			want: false,
		},
		{
			name: "Paid plan with nil date is inactive",
			user: models.User{PlanID: "pro", PlanValidUntil: nil},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := planIsActive(tc.user)
			if got != tc.want {
				t.Errorf("planIsActive() = %v, want %v", got, tc.want)
			}
		})
	}
}
