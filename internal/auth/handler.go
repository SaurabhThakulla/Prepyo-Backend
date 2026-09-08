package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	service       *Service
	secureCookies bool
	sessionTTL    time.Duration
}

func NewHandler(s *Service, secureCookies bool, sessionTTL time.Duration) *Handler {
	return &Handler{service: s, secureCookies: secureCookies, sessionTTL: sessionTTL}
}

// Routes returns the public auth routes. GET /session and DELETE /account need
// a session and are mounted behind RequireUser by the caller.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/google", h.googleSignIn)
	r.Post("/logout", h.logout)

	r.Group(func(private chi.Router) {
		private.Use(h.service.RequireUser)
		private.Get("/session", h.session)
		private.Post("/logout-everywhere", h.logoutEverywhere)
		private.Delete("/account", h.deleteAccount)
	})
	return r
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(SessionCookieName); err == nil && cookie.Value != "" {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			httpx.Internal(w, h.service.log, "auth.logout", err)
			return
		}
	}
	// Clearing the cookie happens whether or not a session was found, so a
	// stale cookie always ends up removed.
	h.clearSessionCookie(w)
	httpx.JSON(w, http.StatusOK, nil)
}

func (h *Handler) logoutEverywhere(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	if err := h.service.LogoutEverywhere(r.Context(), user.ID); err != nil {
		httpx.Internal(w, h.service.log, "auth.logoutEverywhere", err)
		return
	}
	h.clearSessionCookie(w)
	httpx.JSON(w, http.StatusOK, nil)
}

func (h *Handler) session(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	httpx.JSON(w, http.StatusOK, map[string]any{"user": models.NewUserProfile(user)})
}

func (h *Handler) deleteAccount(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	if err := h.service.DeleteAccount(r.Context(), user.ID); err != nil {
		httpx.Internal(w, h.service.log, "auth.deleteAccount", err)
		return
	}

	h.clearSessionCookie(w)
	httpx.JSON(w, http.StatusOK, nil)
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:  SessionCookieName,
		Value: token,
		Path:  "/",
		// HttpOnly keeps the token out of reach of page scripts, so an XSS bug
		// cannot read it.
		HttpOnly: true,
		// Lax still sends the cookie on normal navigation but not on
		// cross-site form posts, which covers the common CSRF cases.
		SameSite: http.SameSiteLaxMode,
		Secure:   h.secureCookies,
		Expires:  time.Now().Add(h.sessionTTL),
		MaxAge:   int(h.sessionTTL.Seconds()),
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.secureCookies,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

type googleSignInRequest struct {
	Credential   string `json:"credential"`
	ReferralCode string `json:"referralCode"`
}

func (h *Handler) googleSignIn(w http.ResponseWriter, r *http.Request) {
	var req googleSignInRequest
	if !httpx.Decode(w, r, &req, h.service.log, "auth.googleSignIn") {
		return
	}

	user, token, err := h.service.SignInWithGoogle(r.Context(), req.Credential, req.ReferralCode)
	switch {
	case errors.Is(err, ErrGoogleNotConfigured):
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeNotConfigured,
			"Google sign-in is not available right now. Please try again later.")
		return
	case errors.Is(err, ErrGoogleTokenInvalid):
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized,
			"That Google sign-in could not be verified. Please try again.")
		return
	case errors.Is(err, ErrGoogleEmailUnusable):
		httpx.ValidationError(w, map[string]string{
			"email": "That Google account has no verified email address.",
		})
		return
	case errors.Is(err, ErrEmailTaken):
		httpx.ValidationError(w, map[string]string{
			"email": "That email is already used by another account.",
		})
		return
	case err != nil:
		httpx.Internal(w, h.service.log, "auth.googleSignIn", err)
		return
	}

	h.setSessionCookie(w, token)
	httpx.JSON(w, http.StatusOK, map[string]any{"user": models.NewUserProfile(user)})
}
