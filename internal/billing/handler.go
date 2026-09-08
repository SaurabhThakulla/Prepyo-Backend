package billing

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
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
	// ProofImage is the learner's screenshot of the transfer, as a data URL
	// from the file picker. It is what an admin looks at when deciding whether
	// the money arrived.
	ProofImage string `json:"proofImage"`
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

	proof, proofType, err := decodeProof(req.ProofImage)
	if err != nil {
		httpx.ValidationError(w, map[string]string{"proofImage": err.Error()})
		return
	}

	// The plan is NOT granted here. This used to hand out a subscription the
	// moment a learner typed any transaction id, with nothing checking that the
	// id was real or that money had arrived. The request now waits as 'pending'
	// until an admin approves it — see internal/admin.reviewPayment, which is
	// the only thing that grants a plan.
	if err := h.service.RequestPayment(r.Context(), h.db, RequestPaymentParams{
		UserID:         user.ID,
		Plan:           plan,
		PaymentGateway: gw,
		TransactionID:  txID,
		ProofImage:     proof,
		ProofImageType: proofType,
	}); err != nil {
		if errors.Is(err, ErrDuplicateTransaction) {
			httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
				"That transaction has already been submitted. We are reviewing it.")
			return
		}
		httpx.Internal(w, h.log, "billing.confirm", err)
		return
	}

	state, err := h.service.State(r.Context(), h.db, user)
	if err != nil {
		httpx.Internal(w, h.log, "billing.confirm.state", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"success":      true,
		"pending":      true,
		"subscription": state,
		"message":      "Payment submitted. We will activate your plan once it is checked, usually within a few hours.",
	})
}

// decodeProof reads the data URL the file picker produced.
//
// Optional: a learner who cannot screenshot should still be able to submit,
// and an admin can ask for it. When present it must be an image and small
// enough that a payment row stays cheap to read.
func decodeProof(raw string) ([]byte, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, "", nil
	}

	prefix, encoded, found := strings.Cut(raw, ",")
	if !found || !strings.HasPrefix(prefix, "data:image/") {
		return nil, "", errors.New("Attach an image of the payment.")
	}

	mime, _, _ := strings.Cut(strings.TrimPrefix(prefix, "data:"), ";")

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", errors.New("That image could not be read. Try another screenshot.")
	}
	if len(decoded) > maxProofBytes {
		return nil, "", errors.New("That image is too large. Please keep it under 4 MB.")
	}
	return decoded, mime, nil
}

// maxProofBytes caps the screenshot. Payment rows are listed in the admin
// queue, and an unbounded blob per row makes that listing expensive.
const maxProofBytes = 4 << 20

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
