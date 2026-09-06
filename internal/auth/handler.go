package auth

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

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
	r.Post("/otp/request", h.requestOTP)
	r.Post("/otp/verify", h.verifyOTP)
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
	var req struct {
		Password string `json:"password"`
		Code     string `json:"code"`
	}
	if !httpx.Decode(w, r, &req, h.service.log, "auth.deleteAccount") {
		return
	}

	user := reqctx.MustUser(r.Context())
	if err := h.service.DeleteAccount(r.Context(), user.ID, req.Password, req.Code); err != nil {
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

type otpRequest struct {
	Phone string `json:"phone"`
}

func (h *Handler) requestOTP(w http.ResponseWriter, r *http.Request) {
	var req otpRequest
	if !httpx.Decode(w, r, &req, h.service.log, "auth.requestOTP") {
		return
	}

	registered, err := h.service.RequestOTP(r.Context(), req.Phone)
	switch {
	case errors.Is(err, ErrInvalidPhone):
		httpx.ValidationError(w, map[string]string{"phone": "Enter a Nepali mobile number, for example 9801234567."})
		return
	case errors.Is(err, ErrOTPRateLimited):
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeLimitReached,
			"Too many codes requested. Wait a minute and try again.")
		return
	case err != nil:
		httpx.Internal(w, h.service.log, "auth.requestOTP", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"registered": registered})
}

type otpVerifyRequest struct {
	Phone        string `json:"phone"`
	Code         string `json:"code"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	ReferralCode string `json:"referralCode"`
}

func (h *Handler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var req otpVerifyRequest
	if !httpx.Decode(w, r, &req, h.service.log, "auth.verifyOTP") {
		return
	}

	user, token, err := h.service.VerifyOTP(r.Context(), req.Phone, req.Code, req.Name, req.Email, req.ReferralCode)
	switch {
	case errors.Is(err, ErrInvalidPhone):
		httpx.ValidationError(w, map[string]string{"phone": "Enter a Nepali mobile number, for example 9801234567."})
		return
	case errors.Is(err, ErrNameRequired):
		httpx.ValidationError(w, map[string]string{"name": "Enter your name to finish creating your account."})
		return
	case errors.Is(err, ErrEmailRequired):
		httpx.ValidationError(w, map[string]string{"email": "Enter your email address."})
		return
	case errors.Is(err, ErrEmailInvalid):
		httpx.ValidationError(w, map[string]string{"email": "Enter a valid email address."})
		return
	case errors.Is(err, ErrEmailTaken):
		httpx.ValidationError(w, map[string]string{"email": "That email is already used by another account."})
		return
	case errors.Is(err, ErrPhoneTaken):
		httpx.ValidationError(w, map[string]string{"phone": "That number already has an account. Sign in instead."})
		return
	case errors.Is(err, ErrOTPExpired):
		httpx.ValidationError(w, map[string]string{"code": "That code has expired. Ask for a new one."})
		return
	case errors.Is(err, ErrOTPInvalid):
		httpx.ValidationError(w, map[string]string{"code": "That code is not right."})
		return
	case errors.Is(err, ErrOTPRateLimited):
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeLimitReached,
			"Too many attempts on that code. Ask for a new one.")
		return
	case err != nil:
		httpx.Internal(w, h.service.log, "auth.verifyOTP", err)
		return
	}

	h.setSessionCookie(w, token)
	httpx.JSON(w, http.StatusOK, map[string]any{"user": models.NewUserProfile(user)})
}
