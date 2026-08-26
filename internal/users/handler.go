package users

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/progress"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	db       *pgxpool.Pool
	repo     *Repository
	progress *progress.Service
	billing  *billing.Service
	log      *slog.Logger
}

func NewHandler(
	db *pgxpool.Pool,
	repo *Repository,
	progressService *progress.Service,
	billingService *billing.Service,
	log *slog.Logger,
) *Handler {
	return &Handler{db: db, repo: repo, progress: progressService, billing: billingService, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.get)
	r.Patch("/", h.update)
	return r
}

// get returns the profile with the derived estimate and plan state attached,
// which is what the dashboard needs in one call.
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	profile := models.NewUserProfile(user)

	estimate, err := h.progress.Estimate(r.Context(), h.db, user)
	if err != nil {
		httpx.Internal(w, h.log, "users.get.estimate", err)
		return
	}
	profile.Estimate = &estimate

	state, err := h.billing.State(r.Context(), h.db, user)
	if err != nil {
		httpx.Internal(w, h.log, "users.get.billing", err)
		return
	}
	profile.Subscription = &state

	httpx.JSON(w, http.StatusOK, map[string]any{"user": profile})
}

type updateRequest struct {
	Name        *string  `json:"name"`
	TargetExam  *string  `json:"targetExam"`
	TargetScore *float64 `json:"targetScore"`
	ExamDate    *string  `json:"examDate"`
	NepalRegion *string  `json:"nepalRegion"`
	Timezone    *string  `json:"timezone"`
}

// update changes onboarding and goal fields.
//
// Only these fields are writable. XP, streak, plan and role are not part of
// the request shape at all, so there is no way to send them.
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req updateRequest
	if !httpx.Decode(w, r, &req, h.log, "users.update") {
		return
	}

	user := reqctx.MustUser(r.Context())
	params := UpdateProfileParams{
		Name:        trimmed(req.Name),
		TargetScore: req.TargetScore,
		NepalRegion: trimmed(req.NepalRegion),
		Timezone:    trimmed(req.Timezone),
	}
	problems := map[string]string{}

	if req.Name != nil && strings.TrimSpace(*req.Name) == "" {
		problems["name"] = "Enter your name."
	}

	if req.TargetExam != nil {
		exam := models.ExamType(*req.TargetExam)
		if !exam.Valid() {
			problems["targetExam"] = "Choose PTE or IELTS."
		} else {
			params.TargetExam = &exam
		}
	}

	if req.ExamDate != nil && *req.ExamDate != "" {
		date, err := time.Parse(time.DateOnly, *req.ExamDate)
		if err != nil {
			problems["examDate"] = "Use the format YYYY-MM-DD."
		} else {
			params.ExamDate = &date
		}
	}

	if req.Timezone != nil && *req.Timezone != "" {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			problems["timezone"] = "Unknown timezone."
		}
	}

	// A target outside the exam's own scale would make every gap calculation
	// nonsense, so it is rejected here rather than stored.
	if req.TargetScore != nil {
		exam := user.TargetExam
		if params.TargetExam != nil {
			exam = *params.TargetExam
		}
		if !validTargetScore(exam, *req.TargetScore) {
			problems["targetScore"] = "That score is outside the range for this exam."
		}
	}

	if len(problems) > 0 {
		httpx.ValidationError(w, problems)
		return
	}

	updated, err := h.repo.UpdateProfile(r.Context(), user.ID, params)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Your account could not be found.")
			return
		}
		httpx.Internal(w, h.log, "users.update", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"user": models.NewUserProfile(updated)})
}

func validTargetScore(exam models.ExamType, score float64) bool {
	if exam == models.ExamIELTS {
		return score >= 0 && score <= 9
	}
	return score >= 10 && score <= 90
}

func trimmed(value *string) *string {
	if value == nil {
		return nil
	}
	t := strings.TrimSpace(*value)
	return &t
}
