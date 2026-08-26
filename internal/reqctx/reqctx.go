// Package reqctx carries the authenticated user on the request context.
//
// It sits on its own so that any module can read the current user without
// importing internal/auth, which would create a cycle: auth already imports
// users, and users needs the current user in its handlers.
package reqctx

import (
	"context"

	"github.com/prepyo/backend/internal/models"
)

// contextKey is unexported, so only this package can put a user on a context.
type contextKey struct{}

var userKey contextKey

// WithUser returns a copy of ctx carrying the user. Only the auth middleware
// should call this.
func WithUser(ctx context.Context, user models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// User returns the authenticated user. The second result is false on public
// routes, where no auth middleware ran.
func User(ctx context.Context) (models.User, bool) {
	user, ok := ctx.Value(userKey).(models.User)
	return user, ok
}

// MustUser returns the authenticated user and panics if there is none.
//
// Handlers mounted behind the auth middleware can call this without an error
// branch: a missing user there is a routing mistake, not a runtime condition.
// The panic is caught by chi's Recoverer and surfaced as a 500.
func MustUser(ctx context.Context) models.User {
	user, ok := User(ctx)
	if !ok {
		panic("reqctx: handler requires a user but was not mounted behind auth.RequireUser")
	}
	return user
}
