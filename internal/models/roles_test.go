package models

import (
	"testing"
	"time"
)

func TestRoleForPlan(t *testing.T) {
	cases := map[string]string{
		"free":    RoleSuru,
		"weekly":  RoleAbhyas,
		"pro":     RoleTaiyari,
		"elite":   RoleUdaan,
		"":        RoleSuru,
		"unknown": RoleSuru,
	}
	for planID, want := range cases {
		if got := RoleForPlan(planID); got != want {
			t.Errorf("RoleForPlan(%q) = %q, want %q", planID, got, want)
		}
	}
}

func TestRoleForUser(t *testing.T) {
	future := time.Now().Add(48 * time.Hour)
	past := time.Now().Add(-48 * time.Hour)

	cases := []struct {
		name string
		user User
		want string
	}{
		{"admin keeps access level", User{Role: RoleAdmin, PlanID: "elite", PlanValidUntil: &future}, RoleAdmin},
		{"admin on no plan stays admin", User{Role: RoleAdmin, PlanID: "free"}, RoleAdmin},
		{"live elite", User{PlanID: "elite", PlanValidUntil: &future}, RoleUdaan},
		{"live pro", User{PlanID: "pro", PlanValidUntil: &future}, RoleTaiyari},
		{"live weekly", User{PlanID: "weekly", PlanValidUntil: &future}, RoleAbhyas},
		{"expired pro falls to free tier", User{PlanID: "pro", PlanValidUntil: &past}, RoleSuru},
		{"paid plan with no expiry is not trusted", User{PlanID: "elite"}, RoleSuru},
		{"free plan", User{PlanID: "free"}, RoleSuru},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := RoleForUser(tc.user); got != tc.want {
				t.Errorf("RoleForUser() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDaysRemaining(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	soon := time.Now().Add(50 * time.Hour)

	if got := (User{PlanID: "pro", PlanValidUntil: &past}).DaysRemaining(); got != 0 {
		t.Errorf("expired DaysRemaining() = %d, want 0", got)
	}
	if got := (User{PlanID: "free"}).DaysRemaining(); got != 0 {
		t.Errorf("free DaysRemaining() = %d, want 0", got)
	}
	if got := (User{PlanID: "pro", PlanValidUntil: &soon}).DaysRemaining(); got != 2 {
		t.Errorf("DaysRemaining() = %d, want 2", got)
	}
}

// The profile is what the client renders the subscription window from, so the
// dates must survive the conversion.
func TestNewUserProfilePlanWindow(t *testing.T) {
	start := time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC)
	until := time.Now().Add(72 * time.Hour)

	profile := NewUserProfile(User{
		Role: RoleTaiyari, PlanID: "pro",
		PlanStartedAt: start, PlanValidUntil: &until,
	})

	if profile.PlanStartedAt != "2026-09-08" {
		t.Errorf("PlanStartedAt = %q, want 2026-09-08", profile.PlanStartedAt)
	}
	if profile.PlanValidUntil != until.Format(time.DateOnly) {
		t.Errorf("PlanValidUntil = %q, want %q", profile.PlanValidUntil, until.Format(time.DateOnly))
	}
	if !profile.PaidPlanActive {
		t.Error("PaidPlanActive = false for a live paid plan")
	}
	if profile.DaysRemaining != 3 {
		t.Errorf("DaysRemaining = %d, want 3", profile.DaysRemaining)
	}
	if profile.PlanID != "pro" {
		t.Errorf("PlanID = %q, want pro", profile.PlanID)
	}
}

// A free profile must not advertise an expiry it does not have.
func TestNewUserProfileFreeHasNoExpiry(t *testing.T) {
	profile := NewUserProfile(User{Role: RoleSuru, PlanID: "free", PlanStartedAt: time.Now()})
	if profile.PlanValidUntil != "" {
		t.Errorf("PlanValidUntil = %q, want empty", profile.PlanValidUntil)
	}
	if profile.PaidPlanActive {
		t.Error("PaidPlanActive = true on the free plan")
	}
}
