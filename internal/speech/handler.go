package speech

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	svc *Service
	log *slog.Logger
}

func NewHandler(svc *Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Routes serves clips by id. Ids come only from segment lists built from known
// scripts, so this cannot be used to speak arbitrary text.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{clipID}", h.clip)
	return r
}

func (h *Handler) clip(w http.ResponseWriter, r *http.Request) {
	audio, contentType, err := h.svc.Clip(r.Context(), chi.URLParam(r, "clipID"))
	if err != nil {
		switch {
		case errors.Is(err, ErrClipNotFound):
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That audio clip does not exist.")
		case errors.Is(err, ai.ErrSpeechUnavailable), errors.Is(err, ai.ErrUnavailable):
			httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeAIUnavailable,
				"Audio could not be generated right now. Try Chrome or Edge, which can read the recording aloud.")
		default:
			httpx.Internal(w, h.log, "speech.clip", err)
		}
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(audio)))
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio)
}

// WriteSegments answers a request for a script's clips, or says server speech
// is not available.
func WriteSegments(w http.ResponseWriter, r *http.Request, svc *Service, log *slog.Logger, script string, fixed map[string]string) {
	if svc == nil || !svc.Available() {
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeNotConfigured,
			"Server audio is not available. Try Chrome or Edge, which can read the recording aloud.")
		return
	}
	segments, err := svc.Segments(r.Context(), script, fixed)
	if err != nil {
		httpx.Internal(w, log, "speech.segments", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"segments": segments})
}
