package institution

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/pkg/httpx"
)

// Handler serves the institution routes.
type Handler struct {
	store         *store
	sessionTTL    time.Duration
	secureCookies bool
	log           *slog.Logger
	signInLearner LearnerSignIn

	failuresByNetwork *failureLimiter
	failuresByEmail   *failureLimiter
}

func NewHandler(pool *pgxpool.Pool, sessionTTL time.Duration, secureCookies bool, signInLearner LearnerSignIn, log *slog.Logger) *Handler {
	return &Handler{
		signInLearner:     signInLearner,
		store:             &store{pool: pool},
		sessionTTL:        sessionTTL,
		secureCookies:     secureCookies,
		log:               log,
		failuresByNetwork: newFailureLimiter(maxFailuresPerNetwork),
		failuresByEmail:   newFailureLimiter(maxFailuresPerEmail),
	}
}

// Routes are mounted at /api/v1/institution. signupLimit is an extra rate limit for /signup.
func (h *Handler) Routes(signupLimit ...func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.With(signupLimit...).Post("/signup", h.signUp)
	r.Post("/login", h.logIn)
	r.Post("/logout", h.logOut)

	// Students: the page that names their institution, and their sign-in.
	r.Get("/lookup/{code}", h.lookup)
	r.Post("/student-login", h.studentLogIn)
	r.Get("/logo/{code}", h.getLogo)

	r.Group(func(admin chi.Router) {
		admin.Use(h.requireAdmin)
		admin.Get("/me", h.me)
		admin.Put("/profile", h.saveProfile)
		admin.Put("/logo", h.putLogo)
		admin.Get("/students", h.listStudents)
		admin.Post("/students", h.addStudent)
		admin.Post("/students/{id}/reset-password", h.resetStudentPassword)
		admin.Delete("/students/{id}", h.removeStudent)
	})
	return r
}

func (h *Handler) signUp(w http.ResponseWriter, r *http.Request) {
	var form SignupForm
	if !httpx.Decode(w, r, &form, h.log, "institution.signUp") {
		return
	}

	form.Clean()
	if problems := form.Problems(); len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	passwordHash, err := hashPassword(form.Password)
	if err != nil {
		httpx.Internal(w, h.log, "institution.signUp", err)
		return
	}

	account, err := h.store.createAccount(r.Context(), form, passwordHash)
	if errors.Is(err, errEmailTaken) {
		httpx.ValidationError(w, map[string]string{
			"email": "That email already has an institution account. Sign in instead.",
		})
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "institution.signUp", err)
		return
	}
	h.log.Info("institution signed up", "institution", account.Institution.Name, "code", account.Institution.Code)

	if err := h.startSession(r.Context(), w, account.Admin.ID); err != nil {
		httpx.Internal(w, h.log, "institution.signUp", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, accountBody(account))
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) logIn(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !httpx.Decode(w, r, &req, h.log, "institution.logIn") {
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	now := time.Now()
	networkKey := email + "|" + clientIP(r)

	if h.failuresByNetwork.locked(networkKey, now) || h.failuresByEmail.locked(email, now) {
		httpx.Error(w, http.StatusTooManyRequests, httpx.CodeTooManyRequest,
			"Too many wrong attempts. Please wait a few minutes and try again.")
		return
	}

	// A missing account is not an error: the password check below still runs.
	account, passwordHash, err := h.store.accountForLogin(r.Context(), email)
	if err != nil && !errors.Is(err, errNotFound) {
		httpx.Internal(w, h.log, "institution.logIn", err)
		return
	}

	if !passwordMatches(passwordHash, req.Password) {
		// Counted for unknown emails too, so the lock does not reveal which are real.
		h.failuresByNetwork.record(networkKey, now)
		h.failuresByEmail.record(email, now)
		httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized,
			"That email and password do not match an institution account.")
		return
	}
	h.failuresByNetwork.clear(networkKey)
	h.failuresByEmail.clear(email)

	if err := h.startSession(r.Context(), w, account.Admin.ID); err != nil {
		httpx.Internal(w, h.log, "institution.logIn", err)
		return
	}
	httpx.JSON(w, http.StatusOK, accountBody(account))
}

func (h *Handler) logOut(w http.ResponseWriter, r *http.Request) {
	if err := h.endSession(w, r); err != nil {
		httpx.Internal(w, h.log, "institution.logOut", err)
		return
	}
	httpx.JSON(w, http.StatusOK, nil)
}

// me is the dashboard's data: the signed-in admin and their institution.
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, accountBody(signedInAccount(r)))
}

// accountBody is the JSON every route answers with.
func accountBody(account Account) map[string]any {
	institution := map[string]any{
		"name":      account.Institution.Name,
		"code":      account.Institution.Code,
		"status":    account.Institution.Status(time.Now()),
		"seatLimit": account.Institution.SeatLimit,
	}
	if account.Institution.ValidUntil != nil {
		institution["validUntil"] = account.Institution.ValidUntil.Format(time.DateOnly)
	}
	institution["profile"] = account.Institution.Profile
	institution["onboarded"] = account.Institution.OnboardedAt != nil
	if account.Institution.LogoUpdatedAt != nil {
		institution["logoUrl"] = logoURL(account.Institution.Code, *account.Institution.LogoUpdatedAt)
	}

	return map[string]any{
		"admin": map[string]any{
			"name":  account.Admin.Name,
			"email": account.Admin.Email,
		},
		"institution": institution,
	}
}

// clientIP is the address wrong passwords are counted against.
func clientIP(r *http.Request) string {
	if ip := middleware.GetClientIP(r.Context()); ip != "" {
		return ip
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
