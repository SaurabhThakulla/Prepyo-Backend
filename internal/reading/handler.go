package reading

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	svc  *Service
	repo *Repository
	log  *slog.Logger
}

func NewHandler(svc *Service, repo *Repository, log *slog.Logger) *Handler {
	return &Handler{svc: svc, repo: repo, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/types", h.types)
	r.Get("/passages", h.listPassages)
	r.Get("/passages/{passageID}", h.getPassage)

	// POST, not GET: dealing a set records that this learner has now read the
	// passage, which is what stops the next set repeating it.
	r.Post("/practice", h.practice)

	r.Route("/mocks", func(m chi.Router) {
		m.Get("/", h.listSessions)
		m.Post("/", h.startMock)
		m.Get("/{sessionID}", h.getMock)
		m.Delete("/{sessionID}", h.abandonMock)
		m.Post("/{sessionID}/submit", h.submitMock)
	})

	return r
}

// examFor resolves the exam a request is about, defaulting to the learner's
// own target exam so the common case needs no parameter.
func examFor(raw string, user models.User) (models.ExamType, bool) {
	if strings.TrimSpace(raw) == "" {
		return user.TargetExam, true
	}
	exam := models.ExamType(raw)
	return exam, exam.Valid()
}

func (h *Handler) types(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	exam, ok := examFor(r.URL.Query().Get("exam"), user)
	if !ok {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Unknown exam. Use PTE or IELTS.")
		return
	}

	list, err := h.repo.TaskTypes(r.Context(), exam)
	if err != nil {
		httpx.Internal(w, h.log, "reading.types", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"types": list,
		// What a generated mock will ask for, so the client can show which
		// types are covered by a mock rather than hardcoding the list.
		"mockRequiredTypes": MockRequiredTypes,
		"mockPassageCount":  MockPassageCount,
	})
}

func (h *Handler) listPassages(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	query := r.URL.Query()

	exam, ok := examFor(query.Get("exam"), user)
	if !ok {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Unknown exam. Use PTE or IELTS.")
		return
	}

	page := httpx.ReadPage(r)
	list, total, err := h.repo.ListPassages(r.Context(), ListPassagesParams{
		Exam:   exam,
		TypeID: query.Get("typeId"),
		Limit:  page.Limit,
		Offset: page.Offset,
	})
	if err != nil {
		httpx.Internal(w, h.log, "reading.listPassages", err)
		return
	}

	seen, err := h.repo.SeenPassageIDs(r.Context(), user.ID, ContextMock)
	if err != nil {
		httpx.Internal(w, h.log, "reading.listPassages.seen", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"passages": list,
		// Which of these the learner has already sat in a mock, and so will not
		// be dealt again. Listing the ids rather than a flag per passage keeps
		// the passage payload the same shape everywhere it appears.
		"satInMockPassageIds": seen,
		"pagination":          page.Meta(total),
	})
}

// getPassage returns one passage with every task set on it, answer keys
// stripped. It is a direct read, so it does not count as having been dealt.
func (h *Handler) getPassage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "passageID")

	sets, err := h.svc.buildSets(r.Context(), []string{id}, r.URL.Query().Get("typeId"))
	if err != nil {
		httpx.Internal(w, h.log, "reading.getPassage", err)
		return
	}
	if len(sets) == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That passage does not exist.")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"set": sets[0]})
}

type practiceRequest struct {
	Exam   string `json:"exam,omitempty"`
	TypeID string `json:"typeId"`
	Limit  int    `json:"limit,omitempty"`
}

// practice deals one task set of the chosen type from a passage picked for this
// learner.
func (h *Handler) practice(w http.ResponseWriter, r *http.Request) {
	var req practiceRequest
	if !httpx.Decode(w, r, &req, h.log, "reading.practice") {
		return
	}
	if strings.TrimSpace(req.TypeID) == "" {
		httpx.ValidationError(w, map[string]string{"typeId": "Choose a task type to practise."})
		return
	}
	if req.Limit < 0 {
		httpx.ValidationError(w, map[string]string{"limit": "Must be zero or more."})
		return
	}

	user := reqctx.MustUser(r.Context())
	exam, ok := examFor(req.Exam, user)
	if !ok {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Unknown exam. Use PTE or IELTS.")
		return
	}

	set, err := h.svc.PracticeSet(r.Context(), user, PracticeParams{
		Exam:   exam,
		TypeID: req.TypeID,
		Limit:  req.Limit,
	})
	if err != nil {
		if errors.Is(err, ErrNoPassage) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound,
				"There are no passages with that task type yet. Try another one.")
			return
		}
		httpx.Internal(w, h.log, "reading.practice", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"set": set})
}

type startMockRequest struct {
	Exam string `json:"exam,omitempty"`
}

// startMock deals a reading paper, or returns the one the learner already holds.
func (h *Handler) startMock(w http.ResponseWriter, r *http.Request) {
	// The body is optional: an empty POST means "a mock for my exam".
	var req startMockRequest
	if r.ContentLength > 0 && !httpx.Decode(w, r, &req, h.log, "reading.startMock") {
		return
	}

	user := reqctx.MustUser(r.Context())
	exam, ok := examFor(req.Exam, user)
	if !ok {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Unknown exam. Use PTE or IELTS.")
		return
	}

	session, err := h.svc.StartMock(r.Context(), user, exam)
	if err != nil {
		switch {
		case errors.Is(err, billing.ErrMockLimitReached):
			httpx.Error(w, http.StatusForbidden, httpx.CodeLimitReached,
				"You have used all mock tests included in your plan. Upgrade or refer friends to unlock more.")
		case errors.Is(err, ErrNoBlueprint):
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound,
				"Reading mocks are not available for that exam yet.")
		case errors.Is(err, ErrBankTooSmall):
			// A content shortage, not a fault the learner can do anything
			// about, so it says what happened rather than "try again".
			h.log.Warn("reading mock bank too small", "op", "reading.startMock", "error", err, "exam", exam)
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
				"There are not enough reading passages to build a full mock yet. Please try again soon.")
		default:
			httpx.Internal(w, h.log, "reading.startMock", err)
		}
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{"session": session})
}

func (h *Handler) getMock(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	session, err := h.svc.ResumeMock(r.Context(), user, chi.URLParam(r, "sessionID"))
	if err != nil {
		h.writeSessionError(w, "reading.getMock", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"session": session})
}

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	page := httpx.ReadPage(r)

	list, total, err := h.repo.ListSessions(r.Context(), user.ID, page.Limit, page.Offset)
	if err != nil {
		httpx.Internal(w, h.log, "reading.listSessions", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"sessions":   list,
		"pagination": page.Meta(total),
	})
}

func (h *Handler) abandonMock(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	if err := h.svc.AbandonMock(r.Context(), user, chi.URLParam(r, "sessionID")); err != nil {
		h.writeSessionError(w, "reading.abandonMock", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"abandoned": true})
}

type submitMockRequest struct {
	Answers         []models.AnswerSubmission `json:"answers"`
	DurationSeconds int                       `json:"durationSeconds"`
}

func (h *Handler) submitMock(w http.ResponseWriter, r *http.Request) {
	var req submitMockRequest
	if !httpx.Decode(w, r, &req, h.log, "reading.submitMock") {
		return
	}

	user := reqctx.MustUser(r.Context())
	result, err := h.svc.SubmitMock(r.Context(), user, chi.URLParam(r, "sessionID"),
		req.Answers, req.DurationSeconds)
	if err != nil {
		if errors.Is(err, ErrNoAnswers) {
			httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest,
				"None of the submitted answers belong to this mock.")
			return
		}
		h.writeSessionError(w, "reading.submitMock", err)
		return
	}

	httpx.JSON(w, http.StatusCreated, map[string]any{
		"attempt":           result.Attempt,
		"xpAwarded":         result.XPAwarded,
		"streak":            result.Streak,
		"scoredQuestions":   result.ScoredQuestions,
		"ungradedQuestions": result.UngradedQuestions,
		"scoreConfidence":   result.ScoreConfidence,
		"review":            result.Review,
	})
}

func (h *Handler) writeSessionError(w http.ResponseWriter, op string, err error) {
	switch {
	case errors.Is(err, ErrSessionNotFound):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That reading mock does not exist.")
	case errors.Is(err, ErrAlreadySubmitted):
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"This reading mock has already been finished.")
	default:
		httpx.Internal(w, h.log, op, err)
	}
}
