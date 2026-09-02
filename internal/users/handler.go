package users

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
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

	// Both pictures behave identically, so they share one implementation and
	// the kind is bound here rather than read off the URL.
	for _, kind := range []ImageKind{ImageAvatar, ImageCover} {
		r.Get("/"+string(kind), h.getImage(kind))
		r.Put("/"+string(kind), h.putImage(kind))
		r.Delete("/"+string(kind), h.deleteImage(kind))
	}
	return r
}

// maxImageBytes caps an uploaded picture. A profile photo that needs more than
// this is a camera original nobody asked to store.
const maxImageBytes = 2 << 20 // 2 MiB

// allowedImageTypes is what the server will store and serve back. The value is
// taken from sniffing the bytes, never from the request header, so a PDF titled
// image/png is rejected on content.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// getImage serves a stored picture to its owner.
func (h *Handler) getImage(kind ImageKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := reqctx.MustUser(r.Context())

		img, err := h.repo.Image(r.Context(), user.ID, kind)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No image set.")
				return
			}
			httpx.Internal(w, h.log, "users.getImage", err)
			return
		}

		// The picture is one user's, so it must not land in a shared cache. The
		// ETag still saves the bytes on a reload.
		etag := fmt.Sprintf(`"%s-%d"`, kind, img.UpdatedAt.UnixNano())
		w.Header().Set("Cache-Control", "private, max-age=0, must-revalidate")
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", img.ContentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(img.Bytes)))
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, err := w.Write(img.Bytes); err != nil {
			h.log.Warn("users.getImage write", "error", err)
		}
	}
}

// putImage stores a picture sent as the raw request body.
func (h *Handler) putImage(kind ImageKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := reqctx.MustUser(r.Context())

		body := http.MaxBytesReader(w, r.Body, maxImageBytes)
		data, err := io.ReadAll(body)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				httpx.Error(w, http.StatusRequestEntityTooLarge, httpx.CodeBadRequest, "That image is larger than 2 MB. Please choose a smaller one.")
				return
			}
			httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "We could not read that upload.")
			return
		}
		if len(data) == 0 {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "That file is empty.")
			return
		}

		// Sniffing beats the declared header: the bytes are what gets served
		// back, so the bytes decide the type.
		contentType := http.DetectContentType(data)
		if !allowedImageTypes[contentType] {
			httpx.Error(w, http.StatusUnsupportedMediaType, httpx.CodeBadRequest, "Please upload a JPEG, PNG, WebP or GIF image.")
			return
		}

		updatedAt, err := h.repo.SetImage(r.Context(), user.ID, kind, data, contentType)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Your account could not be found.")
				return
			}
			httpx.Internal(w, h.log, "users.putImage", err)
			return
		}

		httpx.JSON(w, http.StatusOK, map[string]any{
			"kind":      kind,
			"updatedAt": updatedAt,
		})
	}
}

// deleteImage removes a picture. Removing one that was never set is a success:
// the learner wanted no picture and there is none.
func (h *Handler) deleteImage(kind ImageKind) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := reqctx.MustUser(r.Context())

		if err := h.repo.ClearImage(r.Context(), user.ID, kind); err != nil {
			if errors.Is(err, ErrNotFound) {
				httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Your account could not be found.")
				return
			}
			httpx.Internal(w, h.log, "users.deleteImage", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
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
