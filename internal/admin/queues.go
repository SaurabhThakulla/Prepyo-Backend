package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// The two things an admin works through: issues learners raised, and payments
// waiting to be believed.
//
// Both are queues rather than notifications. A report that was emailed and a
// payment that was granted on trust both had the same problem — nothing was
// left behind to check later.

type issueReport struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
	Path      string `json:"path"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

// reports lists issue reports, open first. ?status=resolved shows the closed
// ones instead.
func (h *Handler) reports(w http.ResponseWriter, r *http.Request) {
	page := httpx.ReadPage(r)
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && status != "open" && status != "resolved" {
		httpx.ValidationError(w, map[string]string{"status": "Use open or resolved."})
		return
	}

	const where = `WHERE ($1 = '' OR r.status = $1)`

	var total int
	if err := h.db.QueryRow(r.Context(),
		`SELECT count(*) FROM issue_reports r `+where, status).Scan(&total); err != nil {
		httpx.Internal(w, h.log, "admin.reports.count", err)
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT r.id, r.user_id, u.name, COALESCE(u.email, ''), r.path, r.message, r.status, r.created_at
		FROM issue_reports r
		JOIN users u ON u.id = r.user_id `+where+`
		ORDER BY r.status = 'open' DESC, r.created_at DESC
		LIMIT $2 OFFSET $3`, status, page.Limit, page.Offset)
	if err != nil {
		httpx.Internal(w, h.log, "admin.reports", err)
		return
	}
	defer rows.Close()

	list := make([]issueReport, 0, page.Limit)
	for rows.Next() {
		var item issueReport
		var created time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.UserEmail,
			&item.Path, &item.Message, &item.Status, &created); err != nil {
			httpx.Internal(w, h.log, "admin.reports.scan", err)
			return
		}
		item.CreatedAt = created.Format(time.RFC3339)
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		httpx.Internal(w, h.log, "admin.reports.rows", err)
		return
	}

	var open int
	_ = h.db.QueryRow(r.Context(),
		`SELECT count(*) FROM issue_reports WHERE status = 'open'`).Scan(&open)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"reports":   list,
		"openCount": open,
		"meta":      page.Meta(total),
	})
}

// resolveReport closes one report. There is no re-open: a report that turns out
// to be unfinished is better raised again with what was learned.
func (h *Handler) resolveReport(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())

	tag, err := h.db.Exec(r.Context(), `
		UPDATE issue_reports
		SET status = 'resolved', resolved_by = $2, resolved_at = now()
		WHERE id = $1 AND status = 'open'`, chi.URLParam(r, "id"), actor.ID)
	if err != nil {
		httpx.Internal(w, h.log, "admin.resolveReport", err)
		return
	}
	if tag.RowsAffected() == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound,
			"That report no longer exists, or is already resolved.")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{"resolved": true})
}

type paymentRequest struct {
	ID            string `json:"id"`
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	UserEmail     string `json:"userEmail"`
	PlanID        string `json:"planId"`
	PlanName      string `json:"planName"`
	Gateway       string `json:"gateway"`
	TransactionID string `json:"transactionId"`
	AmountNPR     int    `json:"amountNpr"`
	Status        string `json:"status"`
	EffectiveDays int    `json:"effectiveDays"`
	HasProof      bool   `json:"hasProof"`
	ReviewNote    string `json:"reviewNote,omitempty"`
	CreatedAt     string `json:"createdAt"`
	ReviewedAt    string `json:"reviewedAt,omitempty"`
}

// payments lists purchase requests, pending first.
func (h *Handler) payments(w http.ResponseWriter, r *http.Request) {
	page := httpx.ReadPage(r)
	status := strings.TrimSpace(r.URL.Query().Get("status"))

	const where = `WHERE ($1 = '' OR p.status = $1)`

	var total int
	if err := h.db.QueryRow(r.Context(),
		`SELECT count(*) FROM subscription_payments p `+where, status).Scan(&total); err != nil {
		httpx.Internal(w, h.log, "admin.payments.count", err)
		return
	}

	rows, err := h.db.Query(r.Context(), `
		SELECT p.id, p.user_id, u.name, COALESCE(u.email, ''), p.plan_id, pl.name,
		       p.payment_gateway, p.transaction_id, p.amount_npr, p.status,
		       p.effective_days, p.proof_image IS NOT NULL, p.review_note,
		       p.created_at, p.reviewed_at
		FROM subscription_payments p
		JOIN users u ON u.id = p.user_id
		JOIN plans pl ON pl.id = p.plan_id `+where+`
		ORDER BY p.status = 'pending' DESC, p.created_at DESC
		LIMIT $2 OFFSET $3`, status, page.Limit, page.Offset)
	if err != nil {
		httpx.Internal(w, h.log, "admin.payments", err)
		return
	}
	defer rows.Close()

	list := make([]paymentRequest, 0, page.Limit)
	for rows.Next() {
		var item paymentRequest
		var created time.Time
		var reviewed *time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.UserEmail,
			&item.PlanID, &item.PlanName, &item.Gateway, &item.TransactionID,
			&item.AmountNPR, &item.Status, &item.EffectiveDays, &item.HasProof,
			&item.ReviewNote, &created, &reviewed); err != nil {
			httpx.Internal(w, h.log, "admin.payments.scan", err)
			return
		}
		item.CreatedAt = created.Format(time.RFC3339)
		if reviewed != nil {
			item.ReviewedAt = reviewed.Format(time.RFC3339)
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		httpx.Internal(w, h.log, "admin.payments.rows", err)
		return
	}

	var pending int
	_ = h.db.QueryRow(r.Context(),
		`SELECT count(*) FROM subscription_payments WHERE status = 'pending'`).Scan(&pending)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"payments":     list,
		"pendingCount": pending,
		"meta":         page.Meta(total),
	})
}

// paymentProof serves the screenshot a learner attached. Behind the admin
// middleware like everything else here, so a proof cannot be read by guessing
// a payment id.
func (h *Handler) paymentProof(w http.ResponseWriter, r *http.Request) {
	var image []byte
	var mime string

	err := h.db.QueryRow(r.Context(),
		`SELECT proof_image, COALESCE(proof_image_type, 'image/jpeg')
		 FROM subscription_payments WHERE id = $1`, chi.URLParam(r, "id")).Scan(&image, &mime)
	if errors.Is(err, pgx.ErrNoRows) || len(image) == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No proof was attached to that payment.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "admin.paymentProof", err)
		return
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "private, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(image)
}

type reviewPaymentRequest struct {
	Approve bool   `json:"approve"`
	Note    string `json:"note"`
}

// reviewPayment approves or rejects one purchase request.
//
// Approving is what actually grants the plan, and it runs through the same
// path a gateway callback would: the entitlement, the role and the period all
// come from one place, so a manually approved subscription is indistinguishable
// from a paid one.
func (h *Handler) reviewPayment(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	paymentID := chi.URLParam(r, "id")

	var req reviewPaymentRequest
	if !httpx.Decode(w, r, &req, h.log, "admin.reviewPayment") {
		return
	}

	granted, err := h.applyPaymentReview(r.Context(), paymentID, actor.ID, req)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound,
			"That request no longer exists, or has already been reviewed.")
		return
	case err != nil:
		httpx.Internal(w, h.log, "admin.reviewPayment", err)
		return
	}

	h.log.Info("admin reviewed a payment",
		"actor", actor.Email, "payment", paymentID, "approved", req.Approve, "grantedDays", granted)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"approved":    req.Approve,
		"grantedDays": granted,
	})
}

func (h *Handler) applyPaymentReview(ctx context.Context, paymentID, actorID string, req reviewPaymentRequest) (int, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin review tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Locked so two admins clicking approve at the same moment cannot grant the
	// same plan twice.
	var userID, planID string
	var days int
	err = tx.QueryRow(ctx, `
		SELECT user_id, plan_id, effective_days
		FROM subscription_payments
		WHERE id = $1 AND status = 'pending'
		FOR UPDATE`, paymentID).Scan(&userID, &planID, &days)
	if err != nil {
		return 0, err
	}

	status := "failed"
	if req.Approve {
		status = "success"
	}

	note := strings.TrimSpace(req.Note)
	if len(note) > 500 {
		note = note[:500]
	}

	if _, err := tx.Exec(ctx, `
		UPDATE subscription_payments
		SET status = $2, reviewed_by = $3, reviewed_at = now(), review_note = $4,
		    processed_at = CASE WHEN $2 = 'success' THEN now() ELSE processed_at END
		WHERE id = $1`, paymentID, status, actorID, note); err != nil {
		return 0, fmt.Errorf("update payment status: %w", err)
	}

	if !req.Approve {
		if err := tx.Commit(ctx); err != nil {
			return 0, fmt.Errorf("commit rejection: %w", err)
		}
		return 0, nil
	}

	// Same shape as a purchase: the period starts today, the expiry extends
	// from whatever was left, and the role follows the plan so the tier is not
	// left behind by the grant.
	if _, err := tx.Exec(ctx, `
		UPDATE users
		SET plan_id = $2,
		    plan_started_at = CURRENT_DATE,
		    -- make_interval keeps the day count an integer. Concatenating it into
		    -- a string forces the parameter to text, which pgx cannot encode.
		    plan_valid_until = COALESCE(GREATEST(plan_valid_until, CURRENT_DATE), CURRENT_DATE) + make_interval(days => $3),
		    role = CASE WHEN role = 'admin' THEN 'admin' ELSE $4 END,
		    updated_at = now()
		WHERE id = $1`, userID, planID, days, roleForPlanID(planID)); err != nil {
		return 0, fmt.Errorf("grant plan: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit approval: %w", err)
	}
	return days, nil
}

// roleForPlanID is planForRole read the other way.
func roleForPlanID(planID string) string {
	for role, plan := range planForRole {
		if plan == planID {
			return role
		}
	}
	return "suru"
}
