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
	// ValidUntil is the end of the paid period; nil on the free plan.
	ValidUntil *time.Time
	SeatLimit  int // how many students it may add; 2 on the free plan
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

const (
	StatusFree    = "free"    // no paid plan: 2 students, reading and listening only
	StatusActive  = "active"  // inside a paid period
	StatusExpired = "expired" // the paid period has ended, so back to the free limits
)

// Status works out the institution's state at the given moment.
func (i Institution) Status(now time.Time) string {
	switch {
	case i.ValidUntil == nil:
		return StatusFree
	case i.ValidUntil.After(now):
		return StatusActive
	default:
		return StatusExpired
	}
}
