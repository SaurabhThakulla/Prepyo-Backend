// Package institution lets a consultancy sign up, sign in and see its dashboard.
// It is self-contained: own tables, own session cookie, own routes.
package institution

import "time"

// Institution is one consultancy.
type Institution struct {
	ID    string
	Name  string
	Code  string // the institution's address on the site: /<code>
	Phone string
	// Plan is basic, extended or ultimate; empty until Prepyo sets one.
	Plan string
	// ValidUntil is the end of the paid period; nil until Prepyo sets one.
	ValidUntil *time.Time
	Profile    Profile
	// OnboardedAt is when the setup was finished; nil keeps the dashboard locked.
	OnboardedAt *time.Time
	// LogoUpdatedAt is nil when no logo was uploaded.
	LogoUpdatedAt *time.Time
}

// Admin is the one person who signs in for an institution.
type Admin struct {
	ID    string
	Name  string
	Email string
}

// Account is a signed-in admin together with their institution.
type Account struct {
	Admin       Admin
	Institution Institution
}

// The plans an institution can buy. Only the seats differ; every student gets the same allowance.
const (
	PlanBasic    = "basic"
	PlanExtended = "extended"
	PlanUltimate = "ultimate"
)

var planSeats = map[string]int{PlanBasic: 10, PlanExtended: 25, PlanUltimate: 50}

// ValidPlan reports whether plan is one of the three plans.
func ValidPlan(plan string) bool { return planSeats[plan] > 0 }

// Seats is how many students the plan allows; 0 with no plan.
func (i Institution) Seats() int { return planSeats[i.Plan] }

const (
	StatusNone    = "none"    // no plan yet: no students can be added
	StatusActive  = "active"  // inside a paid period
	StatusExpired = "expired" // the paid period has ended: students get reading and listening only
)

// Status works out the institution's state at the given moment.
func (i Institution) Status(now time.Time) string {
	switch {
	case i.Plan == "" || i.ValidUntil == nil:
		return StatusNone
	case i.ValidUntil.After(now):
		return StatusActive
	default:
		return StatusExpired
	}
}
