package auth

import (
	"errors"
	"net/http"

	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// SessionCookieName is the cookie the browser sends back on every request.
const SessionCookieName = "prepyo_session"

// RequireUser rejects requests without a valid session and puts the user on
// the request context for the handlers behind it.
func (s *Service) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Please sign in to continue.")
			return
		}

		user, err := s.Authenticate(r.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, ErrSessionNotFound) {
				httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Your session has expired. Please sign in again.")
				return
			}
			httpx.Internal(w, s.log, "auth.RequireUser", err)
			return
		}

		next.ServeHTTP(w, r.WithContext(reqctx.WithUser(r.Context(), user)))
	})
}

// RequireAdmin allows only admin accounts through. Mount it after RequireUser.
func (s *Service) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := reqctx.User(r.Context())
		if !ok {
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Please sign in to continue.")
			return
		}
		if !user.IsAdmin() {
			// Deliberately the same shape as any other forbidden response: it
			// should not confirm that an admin area exists.
			httpx.Error(w, http.StatusForbidden, httpx.CodeForbidden, "You do not have access to this resource.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
