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
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/notifications"
	"github.com/prepyo/backend/internal/referrals"
	"github.com/prepyo/backend/internal/report"
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
	// ReplyCount and LastFromLearner let the queue show which conversations
	// are waiting on the team.
	ReplyCount      int  `json:"replyCount"`
	LastFromLearner bool `json:"lastFromLearner"`
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
		SELECT r.id, r.user_id, u.name, COALESCE(u.email, ''), r.path, r.message, r.status, r.created_at,
		       (SELECT count(*) FROM issue_report_replies p WHERE p.report_id = r.id),
		       COALESCE((SELECT NOT p.from_staff FROM issue_report_replies p
		                 WHERE p.report_id = r.id ORDER BY p.created_at DESC LIMIT 1), true)
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
			&item.Path, &item.Message, &item.Status, &created, &item.ReplyCount, &item.LastFromLearner); err != nil {
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

	ctx := r.Context()
	tx, err := h.db.Begin(ctx)
	if err != nil {
		httpx.Internal(w, h.log, "admin.resolveReport.begin", err)
		return
	}
	defer tx.Rollback(ctx)

	if err := resolveReportTx(ctx, tx, chi.URLParam(r, "id"), actor.ID); err != nil {
		if errors.Is(err, report.ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound,
				"That report no longer exists, or is already resolved.")
			return
		}
		httpx.Internal(w, h.log, "admin.resolveReport", err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		httpx.Internal(w, h.log, "admin.resolveReport.commit", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"resolved": true})
}

// resolveReportTx closes an open report and tells the learner.
func resolveReportTx(ctx context.Context, tx pgx.Tx, reportID, actorID string) error {
	var learnerID string
	err := tx.QueryRow(ctx, `
		UPDATE issue_reports
		SET status = 'resolved', resolved_by = $2, resolved_at = now()
		WHERE id = $1 AND status = 'open'
		RETURNING user_id`, reportID, actorID).Scan(&learnerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return report.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("resolve report: %w", err)
	}
	_, err = notifications.Send(ctx, tx, notifications.CreateParams{
		UserID: learnerID, Type: notifications.TypeReport,
		Title:     "Your report has been resolved",
		Message:   "Thanks for flagging it. If something is still not right, send us a new report.",
		ActionURL: "/reports?report=" + reportID,
	})
	return err
}

// report returns one report with its whole conversation.
func (h *Handler) report(w http.ResponseWriter, r *http.Request) {
	rep, err := report.Load(r.Context(), h.db, chi.URLParam(r, "id"), "")
	if errors.Is(err, report.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That report no longer exists.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "admin.report", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"report": rep})
}

type replyReportRequest struct {
	Message string `json:"message"`
	// Resolve closes the report in the same step, for "here is the fix".
	Resolve bool `json:"resolve"`
}

// replyReport answers a learner, and tells them in their notifications.
func (h *Handler) replyReport(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())

	var req replyReportRequest
	if !httpx.Decode(w, r, &req, h.log, "admin.replyReport") {
		return
	}
	message := report.Clip(strings.TrimSpace(req.Message))
	if message == "" {
		httpx.ValidationError(w, map[string]string{"message": "Write a reply first."})
		return
	}

	ctx := r.Context()
	id := chi.URLParam(r, "id")
	rep, err := report.Load(ctx, h.db, id, "")
	if errors.Is(err, report.ErrNotFound) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That report no longer exists.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "admin.replyReport.load", err)
		return
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		httpx.Internal(w, h.log, "admin.replyReport.begin", err)
		return
	}
	defer tx.Rollback(ctx)

	if err := report.AddReply(ctx, tx, id, actor.ID, true, message); err != nil {
		httpx.Internal(w, h.log, "admin.replyReport", err)
		return
	}
	if _, err := notifications.Send(ctx, tx, notifications.CreateParams{
		UserID: rep.UserID, Type: notifications.TypeReport,
		Title:     "Prepyo team replied to your report",
		Message:   report.Snippet(message),
		ActionURL: "/reports?report=" + id,
	}); err != nil {
		httpx.Internal(w, h.log, "admin.replyReport.notify", err)
		return
	}
	if req.Resolve && rep.Status == "open" {
		if err := resolveReportTx(ctx, tx, id, actor.ID); err != nil {
			httpx.Internal(w, h.log, "admin.replyReport.resolve", err)
			return
		}
	}
	if err := tx.Commit(ctx); err != nil {
		httpx.Internal(w, h.log, "admin.replyReport.commit", err)
		return
	}

	updated, err := report.Load(ctx, h.db, id, "")
	if err != nil {
		httpx.Internal(w, h.log, "admin.replyReport.reload", err)
		return
	}
	h.log.Info("admin replied to a report", "actor", actor.Email, "report", id, "resolved", req.Resolve)
	httpx.JSON(w, http.StatusOK, map[string]any{"report": updated})
}

type announcementRequest struct {
	Title     string `json:"title"`
	Message   string `json:"message"`
	ActionURL string `json:"actionUrl"`
}

// announce sends one notice to every learner.
func (h *Handler) announce(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())

	var req announcementRequest
	if !httpx.Decode(w, r, &req, h.log, "admin.announce") {
		return
	}
	title := strings.TrimSpace(req.Title)
	message := strings.TrimSpace(req.Message)
	fields := map[string]string{}
	if title == "" || len([]rune(title)) > 120 {
		fields["title"] = "Give it a title of up to 120 characters."
	}
	if message == "" || len([]rune(message)) > 1000 {
		fields["message"] = "Write a message of up to 1000 characters."
	}
	// Only links inside the app: an announcement is not a way to send
	// learners somewhere else under Prepyo's name.
	link := strings.TrimSpace(req.ActionURL)
	if link != "" && (!strings.HasPrefix(link, "/") || strings.HasPrefix(link, "//")) {
		fields["actionUrl"] = "Use a page inside Prepyo, starting with /."
	}
	if len(fields) > 0 {
		httpx.ValidationError(w, fields)
		return
	}

	sent, err := notifications.NewRepository(h.db).Announce(r.Context(), title, message, link)
	if err != nil {
		httpx.Internal(w, h.log, "admin.announce", err)
		return
	}
	h.log.Info("admin sent an announcement", "actor", actor.Email, "recipients", sent)
	httpx.JSON(w, http.StatusOK, map[string]any{"sent": sent})
}

type paymentRequest struct {
	ID            string `json:"id"`
	UserID        string `json:"userId"`
	UserName      string `json:"userName"`
	UserEmail     string `json:"userEmail"`
	PhoneNumber   string `json:"phoneNumber,omitempty"`
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
		SELECT p.id, p.user_id, u.name, COALESCE(u.email, ''), COALESCE(p.phone_number, ''), p.plan_id, pl.name,
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
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.UserEmail, &item.PhoneNumber,
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

	err := h.db.QueryRow(r.Context(),
		`SELECT proof_image FROM subscription_payments WHERE id = $1`,
		chi.URLParam(r, "id")).Scan(&image)
	if errors.Is(err, pgx.ErrNoRows) || len(image) == 0 {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No proof was attached to that payment.")
		return
	}
	if err != nil {
		httpx.Internal(w, h.log, "admin.paymentProof", err)
		return
	}

	// Rows stored before uploads were sniffed carry whatever type the learner's
	// request claimed, so the bytes decide here too. nosniff and a sandbox mean
	// that even such a row, opened directly in a tab, cannot run script as the
	// admin.
	w.Header().Set("Content-Type", http.DetectContentType(image))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; sandbox")
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

	granted, queued, err := h.applyPaymentReview(r.Context(), paymentID, actor.ID, req)
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
		"actor", actor.Email, "payment", paymentID, "approved", req.Approve,
		"grantedDays", granted, "queued", queued)

	httpx.JSON(w, http.StatusOK, map[string]any{
		"approved":    req.Approve,
		"grantedDays": granted,
		// queued means the learner already had a live plan, so this one starts
		// when that ends rather than replacing it.
		"queued": queued,
	})
}

// applyPaymentReview returns the days granted and whether the plan had to wait
// behind one that is still running.
func (h *Handler) applyPaymentReview(ctx context.Context, paymentID, actorID string, req reviewPaymentRequest) (int, bool, error) {
	tx, err := h.db.Begin(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("begin review tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := referrals.LockLifecycle(ctx, tx); err != nil {
		return 0, false, err
	}

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
		return 0, false, err
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
		return 0, false, fmt.Errorf("update payment status: %w", err)
	}

	var planName string
	if err := tx.QueryRow(ctx, `SELECT name FROM plans WHERE id = $1`, planID).Scan(&planName); err != nil {
		return 0, false, fmt.Errorf("read plan name: %w", err)
	}

	if !req.Approve {
		message := "We could not verify this payment, so the plan was not activated."
		if note != "" {
			message += " Note from our team: " + note
		}
		if _, err := notifications.Send(ctx, tx, notifications.CreateParams{
			UserID: userID, Type: notifications.TypePayment,
			Title:     "Payment for " + planName + " not approved",
			Message:   message + " Reply through Report an issue if you think this is a mistake.",
			ActionURL: "/subscription",
		}); err != nil {
			return 0, false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return 0, false, fmt.Errorf("commit rejection: %w", err)
		}
		return 0, false, nil
	}

	if err := referrals.QualifyPayment(ctx, tx, paymentID); err != nil {
		return 0, false, err
	}

	queued, err := billing.GrantPurchase(ctx, tx, userID, planID, paymentID, days)
	if err != nil {
		return 0, false, err
	}

	// A queued plan starts when the running one ends; either way the learner
	// is told the date that matters to them.
	var validUntil time.Time
	if err := tx.QueryRow(ctx, `SELECT plan_valid_until FROM users WHERE id = $1`, userID).Scan(&validUntil); err != nil {
		return 0, false, fmt.Errorf("read plan end: %w", err)
	}
	message := fmt.Sprintf("Your %s plan is active until %s. Enjoy your practice!", planName, validUntil.Format("2 Jan 2006"))
	if queued {
		message = fmt.Sprintf("Your %s plan (%d days) will start when your current plan ends on %s.", planName, days, validUntil.Format("2 Jan 2006"))
	}
	if _, err := notifications.Send(ctx, tx, notifications.CreateParams{
		UserID: userID, Type: notifications.TypePayment,
		Title:     "Payment approved",
		Message:   message,
		ActionURL: "/subscription",
	}); err != nil {
		return 0, false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, false, err
	}
	return days, queued, nil
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
