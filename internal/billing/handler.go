package billing

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

type Handler struct {
	db      *pgxpool.Pool
	repo    *Repository
	service *Service
	log     *slog.Logger
}

func NewHandler(db *pgxpool.Pool, repo *Repository, service *Service, log *slog.Logger) *Handler {
	return &Handler{db: db, repo: repo, service: service, log: log}
}

// Routes: /plans and /webhook are public; the rest need a session.
func (h *Handler) Routes(requireUser func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()
	r.Get("/plans", h.plans)
	r.Post("/webhook", h.webhook)

	r.Group(func(private chi.Router) {
		private.Use(requireUser)
		private.Get("/", h.state)
		private.Post("/checkout", h.checkout)
		private.Post("/confirm", h.confirm)
	})
	return r
}

func (h *Handler) plans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.repo.Plans(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "billing.plans", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"plans": plans})
}

func (h *Handler) state(w http.ResponseWriter, r *http.Request) {
	user := reqctx.MustUser(r.Context())

	state, err := h.service.State(r.Context(), h.db, user)
	if err != nil {
		httpx.Internal(w, h.log, "billing.state", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"subscription": state})
}

type checkoutRequest struct {
	PlanID         string `json:"planId"`
	PaymentGateway string `json:"paymentGateway"`
	TransactionID  string `json:"transactionId"`
}

func (h *Handler) checkout(w http.ResponseWriter, r *http.Request) {
	// Proxies to confirm for instant QR / gateway simulated checkout
	h.confirm(w, r)
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if !httpx.Decode(w, r, &req, h.log, "billing.confirm") {
		return
	}

	if strings.TrimSpace(req.PlanID) == "" {
		httpx.ValidationError(w, map[string]string{"planId": "Required."})
		return
	}

	user := reqctx.MustUser(r.Context())

	// If transaction ID was not provided, generate a simulated idempotent tracking ID
	txID := strings.TrimSpace(req.TransactionID)
	if txID == "" {
		buf := make([]byte, 12)
		_, _ = rand.Read(buf)
		txID = "SIM-" + hex.EncodeToString(buf)
	}

	gw := strings.ToUpper(strings.TrimSpace(req.PaymentGateway))
	if gw == "" {
		gw = "ESEWA"
	}

	plan, err := h.repo.Plan(r.Context(), req.PlanID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Plan not found.")
		return
	}

	state, err := h.service.ConfirmPayment(r.Context(), h.db, ConfirmPaymentParams{
		UserID:         user.ID,
		PlanID:         plan.ID,
		PaymentGateway: gw,
		TransactionID:  txID,
		AmountNPR:      plan.PriceNPR,
	})
	if err != nil {
		httpx.Internal(w, h.log, "billing.confirm", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"success":      true,
		"subscription": state,
		"message":      "Payment confirmed successfully! Bonus days have been applied.",
	})
}

type webhookPayload struct {
	UserID         string `json:"userId"`
	PlanID         string `json:"planId"`
	PaymentGateway string `json:"paymentGateway"`
	TransactionID  string `json:"transactionId"`
	Status         string `json:"status"`
	AmountNPR      int    `json:"amountNPR"`
}

func (h *Handler) webhook(w http.ResponseWriter, r *http.Request) {
	var payload webhookPayload
	if !httpx.Decode(w, r, &payload, h.log, "billing.webhook") {
		return
	}

	if payload.Status != "success" && payload.Status != "COMPLETE" {
		// Non-successful events do not grant bonuses
		httpx.JSON(w, http.StatusOK, map[string]any{"received": true, "applied": false})
		return
	}

	state, err := h.service.ConfirmPayment(r.Context(), h.db, ConfirmPaymentParams{
		UserID:         payload.UserID,
		PlanID:         payload.PlanID,
		PaymentGateway: payload.PaymentGateway,
		TransactionID:  payload.TransactionID,
		AmountNPR:      payload.AmountNPR,
	})
	if err != nil {
		httpx.Internal(w, h.log, "billing.webhook", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"received":     true,
		"applied":      true,
		"subscription": state,
	})
}
