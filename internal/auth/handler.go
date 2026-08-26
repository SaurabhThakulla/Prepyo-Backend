package auth

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// minPasswordLength follows current NIST guidance: length matters, composition
// rules ("must contain a symbol") mostly produce predictable passwords.
const minPasswordLength = 10

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
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/logout", h.logout)

	r.Group(func(private chi.Router) {
		private.Use(h.service.RequireUser)
		private.Get("/session", h.session)
		private.Post("/logout-everywhere", h.logoutEverywhere)
		private.Delete("/account", h.deleteAccount)
	})
	return r
}

type registerRequest struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	Name         string `json:"name"`
	NepalRegion  string `json:"nepalRegion"`
	Timezone     string `json:"timezone"`
	ReferralCode string `json:"referralCode"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !httpx.Decode(w, r, &req, h.service.log, "auth.register") {
		return
	}

	problems := map[string]string{}
	if !validEmail(req.Email) {
		problems["email"] = "Enter a valid email address."
	}
	if utf8.RuneCountInString(req.Password) < minPasswordLength {
		problems["password"] = "Use at least 10 characters."
	}
	if strings.TrimSpace(req.Name) == "" {
		problems["name"] = "Enter your name."
	}
	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	user, token, err := h.service.Register(r.Context(), RegisterParams{
		Email:        req.Email,
		Password:     req.Password,
		Name:         req.Name,
		NepalRegion:  req.NepalRegion,
		Timezone:     req.Timezone,
		ReferralCode: req.ReferralCode,
	})
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict, "That email is already registered. Try signing in.")
			return
		}
		httpx.Internal(w, h.service.log, "auth.register", err)
		return
	}

	h.setSessionCookie(w, token)
	httpx.JSON(w, http.StatusCreated, map[string]any{"user": models.NewUserProfile(user)})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !httpx.Decode(w, r, &req, h.service.log, "auth.login") {
		return
	}

	user, token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			// One message for both a wrong password and an unknown address, so
			// this endpoint cannot be used to discover who has an account.
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Incorrect email or password.")
			return
		}
		httpx.Internal(w, h.service.log, "auth.login", err)
		return
	}

	h.setSessionCookie(w, token)
	httpx.JSON(w, http.StatusOK, map[string]any{"user": models.NewUserProfile(user)})
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
	var req struct {
		Password string `json:"password"`
	}
	if !httpx.Decode(w, r, &req, h.service.log, "auth.deleteAccount") {
		return
	}

	user := reqctx.MustUser(r.Context())
	if err := h.service.DeleteAccount(r.Context(), user.ID, req.Password); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Incorrect password.")
			return
		}
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

func validEmail(input string) bool {
	address := strings.TrimSpace(input)
	if address == "" || len(address) > 254 {
		return false
	}
	parsed, err := mail.ParseAddress(address)
	// ParseAddress accepts `Name <a@b.c>`; for a signup field we want the bare
	// address only.
	return err == nil && parsed.Address == address
}
