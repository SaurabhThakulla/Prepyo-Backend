package billing

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/referrals"
)

// ConfirmPayment is INTERNAL: callers must independently verify the provider's
// signature and transaction. The public webhook remains fail-closed.
func (s *Service) ConfirmPayment(ctx context.Context, pool *pgxpool.Pool, p ConfirmPaymentParams) (models.SubscriptionState, error) {
	if p.UserID == "" || p.PlanID == "" || p.TransactionID == "" || p.PaymentGateway == "" {
		return models.SubscriptionState{}, ErrInvalidPayment
	}
	plan, err := s.repo.Plan(ctx, p.PlanID)
	if err != nil {
		return models.SubscriptionState{}, err
	}
	if plan.ID == "free" || plan.PriceNPR <= 0 || p.AmountNPR != plan.PriceNPR {
		return models.SubscriptionState{}, ErrInvalidPayment
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.SubscriptionState{}, err
	}
	defer tx.Rollback(ctx)
	if err = referrals.LockLifecycle(ctx, tx); err != nil {
		return models.SubscriptionState{}, err
	}
	var id, status, owner, planID, gateway string
	var amount, days int
	err = tx.QueryRow(ctx, `SELECT id, status, user_id, plan_id, payment_gateway, amount_npr, effective_days
  FROM subscription_payments WHERE transaction_id = $1 FOR UPDATE`, p.TransactionID).Scan(&id, &status, &owner, &planID, &gateway, &amount, &days)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		base := plan.DurationDays
		if base <= 0 {
			base = plan.DurationMonths * 30
		}
		days = base + plan.BonusDays
		if days <= 0 {
			return models.SubscriptionState{}, ErrInvalidPayment
		}
		err = tx.QueryRow(ctx, `INSERT INTO subscription_payments
   (user_id, plan_id, payment_gateway, transaction_id, amount_npr, status, base_days, bonus_days, effective_days)
   VALUES ($1,$2,$3,$4,$5,'pending',$6,$7,$8) RETURNING id`, p.UserID, p.PlanID, p.PaymentGateway, p.TransactionID, p.AmountNPR, base, plan.BonusDays, days).Scan(&id)
		if err != nil {
			return models.SubscriptionState{}, err
		}
	case err != nil:
		return models.SubscriptionState{}, err
	default:
		if owner != p.UserID || planID != p.PlanID || gateway != p.PaymentGateway || amount != p.AmountNPR {
			return models.SubscriptionState{}, ErrInvalidPayment
		}
		if status == "success" {
			u, err := subscriptionUser(ctx, tx, p.UserID)
			if err != nil {
				return models.SubscriptionState{}, err
			}
			return s.State(ctx, tx, u)
		}
		if status != "pending" || days <= 0 {
			return models.SubscriptionState{}, ErrInvalidPayment
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE subscription_payments SET status = 'success', processed_at = now() WHERE id = $1`, id); err != nil {
		return models.SubscriptionState{}, err
	}
	if err = referrals.QualifyPayment(ctx, tx, id); err != nil {
		return models.SubscriptionState{}, err
	}
	if _, err = GrantPurchase(ctx, tx, p.UserID, p.PlanID, id, days); err != nil {
		return models.SubscriptionState{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return models.SubscriptionState{}, fmt.Errorf("commit confirmation: %w", err)
	}
	u, err := subscriptionUser(ctx, pool, p.UserID)
	if err != nil {
		return models.SubscriptionState{}, err
	}
	return s.State(ctx, pool, u)
}

// subscriptionUser reads only the fields used by State, avoiding a dependency
// on the users HTTP package (which itself depends on billing).
func subscriptionUser(ctx context.Context, db database.DB, id string) (models.User, error) {
	var u models.User
	err := db.QueryRow(ctx, `SELECT id, plan_id, plan_started_at, plan_valid_until, timezone,
  bonus_mock_tests FROM users WHERE id = $1`, id).Scan(&u.ID, &u.PlanID, &u.PlanStartedAt, &u.PlanValidUntil, &u.Timezone, &u.BonusMockTests)
	return u, err
}

// GrantPurchase preserves active entitlements and parks paid purchases in the
// existing queue. Both admin approval and verified provider confirmation use it.
func GrantPurchase(ctx context.Context, tx pgx.Tx, userID, planID, paymentID string, days int) (bool, error) {
	if days <= 0 || planID == "free" {
		return false, ErrInvalidPayment
	}
	var active bool
	if err := tx.QueryRow(ctx, `SELECT plan_id <> 'free' AND COALESCE(plan_valid_until > CURRENT_DATE, false)
  FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&active); err != nil {
		return false, err
	}
	if active {
		_, err := tx.Exec(ctx, `INSERT INTO queued_plans (user_id, plan_id, days, payment_id, status) VALUES ($1,$2,$3,$4,'queued')`, userID, planID, days, paymentID)
		return true, err
	}
	_, err := tx.Exec(ctx, `UPDATE users SET plan_id = $2, plan_started_at = CURRENT_DATE,
  plan_valid_until = CURRENT_DATE + make_interval(days => $3),
  role = CASE WHEN role = 'admin' THEN role ELSE $4 END, updated_at = now() WHERE id = $1`, userID, planID, days, models.RoleForPlan(planID))
	return false, err
}
