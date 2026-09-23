package questions

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	repo *Repository
	log  *slog.Logger
}

func NewHandler(repo *Repository, log *slog.Logger) *Handler {
	return &Handler{repo: repo, log: log}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	r.Get("/assets/{assetID}", h.getAsset)
	r.Get("/{questionID}", h.get)
	return r
}

func (h *Handler) getAsset(w http.ResponseWriter, r *http.Request) {
	assetID := chi.URLParam(r, "assetID")
	asset, err := h.repo.GetAsset(r.Context(), assetID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Asset not found.")
			return
		}
		httpx.Internal(w, h.log, "questions.getAsset", err)
		return
	}

	w.Header().Set("Content-Type", asset.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(asset.Data)))
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Question images may be SVG, which can carry script. An <img> never runs
	// it, but the same URL opened directly would, on our origin; the sandbox
	// stops that without affecting how the image or audio is embedded.
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; sandbox")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(asset.Data)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	exam := models.ExamType(query.Get("exam"))
	skill := models.SkillType(query.Get("skill"))

	if exam != "" && !exam.Valid() {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Unknown exam. Use PTE or IELTS.")
		return
	}
	if skill != "" && !skill.Valid() {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "Unknown skill.")
		return
	}

	page := httpx.ReadPage(r)
	list, total, err := h.repo.List(r.Context(), ListParams{
		Exam:   exam,
		Skill:  skill,
		TypeID: query.Get("typeId"),
		Limit:  page.Limit,
		Offset: page.Offset,
		// Opt-in, for tooling that wants to see the whole bank. A learner-facing
		// caller wants /api/v1/reading, which serves these with their passage.
		IncludePassageQuestions: query.Get("includePassageQuestions") == "true",
		Random:                  query.Get("random") == "true",
	})
	if err != nil {
		httpx.Internal(w, h.log, "questions.list", err)
		return
	}

	// Strip the answer key: this list feeds the practice screen.
	safe := make([]models.Question, len(list))
	for i, q := range list {
		safe[i] = q.PublicQuestion()
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"questions":  safe,
		"pagination": page.Meta(total),
	})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	q, err := h.repo.ByID(r.Context(), chi.URLParam(r, "questionID"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That question does not exist.")
			return
		}
		httpx.Internal(w, h.log, "questions.get", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"question": q.PublicQuestion()})
}
