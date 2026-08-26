package referrals

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	service *Service
	log     *slog.Logger
}

func NewHandler(service *Service, log *slog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

func (h *Handler) Routes(requireUser func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Public endpoint for pre-validating code on signup page
	r.Get("/validate", h.validate)

	// Authenticated endpoint for referral overview dashboard
	r.Group(func(private chi.Router) {
		private.Use(requireUser)
		private.Get("/me", h.overview)
	})

	return r
}

func (h *Handler) validate(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	res, err := h.service.ValidateCode(r.Context(), code)
	if err != nil {
		httpx.Internal(w, h.log, "referrals.validate", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"valid":        res.Valid,
		"referrerName": res.ReferrerName,
		"message":      res.Message,
	})
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())
	res, err := h.service.Overview(r.Context(), user)
	if err != nil {
		httpx.Internal(w, h.log, "referrals.overview", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"referralCode":    res.ReferralCode,
		"shareLink":       res.ShareLink,
		"stats":           res.Stats,
		"milestones":      res.Milestones,
		"recentReferrals": res.RecentReferrals,
	})
}
