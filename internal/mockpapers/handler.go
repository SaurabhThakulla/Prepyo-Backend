package mockpapers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	catalog *CatalogService
	builder *Builder
	repo    *Repository
	log     *slog.Logger
}

func NewHandler(catalog *CatalogService, builder *Builder, repo *Repository, log *slog.Logger) *Handler {
	return &Handler{
		catalog: catalog,
		builder: builder,
		repo:    repo,
		log:     log,
	}
}

// Routes returns the learner-facing API router mounted at /api/v1/mock-papers.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.list)
	return r
}

// AdminRoutes returns the admin router mounted at /admin/mock-papers.
func (h *Handler) AdminRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/build", h.adminBuild)
	r.Post("/{id}/revise", h.adminRevise)
	r.Post("/{id}/retire", h.adminRetire)
	return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	exam := strings.TrimSpace(r.URL.Query().Get("exam"))
	if exam == "" {
		exam = string(user.TargetExam)
	}
	section := strings.TrimSpace(r.URL.Query().Get("section"))
	if section == "" {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "The section query parameter is required.")
		return
	}
	module := strings.TrimSpace(r.URL.Query().Get("module"))

	papers, err := h.catalog.Catalog(r.Context(), user, exam, section, module)
	if err != nil {
		httpx.Internal(w, h.log, "mockpapers.list", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"papers": papers,
	})
}

type buildRequest struct {
	Exam    string `json:"exam"`
	Section string `json:"section"`
	Module  string `json:"module"`
	Count   int    `json:"count"`
}

func (h *Handler) adminBuild(w http.ResponseWriter, r *http.Request) {
	var req buildRequest
	if !httpx.Decode(w, r, &req, h.log, "mockpapers.adminBuild") {
		return
	}

	count := req.Count
	if count <= 0 {
		count = DefaultTargetPapers
	}

	scope := Scope{
		Exam:    strings.ToLower(strings.TrimSpace(req.Exam)),
		Section: strings.ToLower(strings.TrimSpace(req.Section)),
		Module:  strings.ToLower(strings.TrimSpace(req.Module)),
	}
	if scope.Module == "" {
		scope.Module = ModuleAny
	}

	if h.builder != nil {
		// The build outlives this request: its context is cancelled as soon as
		// the 202 below is sent, so the build runs on its own.
		go func() {
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), buildTimeout)
			defer cancel()
			if err := h.builder.BuildScope(ctx, scope, count); err != nil {
				h.log.Error("admin build failed", "scope", scope, "error", err)
			}
		}()
	}

	httpx.JSON(w, http.StatusAccepted, map[string]any{
		"queued":  true,
		"scope":   scope,
		"target":  count,
		"message": "Paper generation started in the background",
	})
}

type reviseRequest struct {
	Title   string          `json:"title"`
	Content json.RawMessage `json:"content"`
}

func (h *Handler) adminRevise(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req reviseRequest
	if !httpx.Decode(w, r, &req, h.log, "mockpapers.adminRevise") {
		return
	}
	if len(req.Content) == 0 {
		httpx.Error(w, http.StatusBadRequest, httpx.CodeBadRequest, "content is required")
		return
	}

	paper, err := h.repo.Revise(r.Context(), id, req.Content, req.Title)
	if err != nil {
		if errors.Is(err, ErrPaperNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Mock paper not found.")
			return
		}
		httpx.Internal(w, h.log, "mockpapers.adminRevise", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"paper": paper,
	})
}

func (h *Handler) adminRetire(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Retire(r.Context(), id); err != nil {
		if errors.Is(err, ErrPaperNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Mock paper not found.")
			return
		}
		httpx.Internal(w, h.log, "mockpapers.adminRetire", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"retired": true,
		"id":      id,
	})
}
