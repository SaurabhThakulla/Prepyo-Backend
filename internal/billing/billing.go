// Package billing owns plans, entitlements and purchase rewards.
//
// Entitlements are read from the database on every check rather than cached on
// the user row, so a plan change takes effect immediately and a stale counter
// cannot grant access nobody paid for.
package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/notifications"
)

var (
	ErrPlanNotFound      = errors.New("plan not found")
	ErrLimitReached      = errors.New("daily evaluation limit reached")
	ErrMockLimitReached  = errors.New("mock test allowance reached for your plan")
	ErrInvalidPayment    = errors.New("invalid payment confirmation")
)

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

const planFields = `id, name, price_npr, duration_months, duration_days, bonus_days, features, ai_evaluations_per_day, mock_tests_included, is_popular`

func (r *Repository) Plans(ctx context.Context) ([]models.Plan, error) {
	rows, err := r.db.Query(ctx, `SELECT `+planFields+` FROM plans ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()

	plans := []models.Plan{}
	for rows.Next() {
		var p models.Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.PriceNPR, &p.DurationMonths, &p.DurationDays, &p.BonusDays,
			&p.Features, &p.AIEvaluationsPerDay, &p.MockTestsIncluded, &p.IsPopular); err != nil {
			return nil, fmt.Errorf("scan plan: %w", err)
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *Repository) Plan(ctx context.Context, id string) (models.Plan, error) {
	var p models.Plan
	err := r.db.QueryRow(ctx, `SELECT `+planFields+` FROM plans WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.PriceNPR, &p.DurationMonths, &p.DurationDays, &p.BonusDays,
			&p.Features, &p.AIEvaluationsPerDay, &p.MockTestsIncluded, &p.IsPopular)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Plan{}, ErrPlanNotFound
	}
	if err != nil {
		return models.Plan{}, fmt.Errorf("get plan: %w", err)
	}
	return p, nil
}

type Service struct {
	repo          *Repository
	notifications *notifications.Repository
}

func NewService(repo *Repository, notifs *notifications.Repository) *Service {
	return &Service{repo: repo, notifications: notifs}
}

// State returns the learner's plan together with today's usage and mock test entitlement.
func (s *Service) State(ctx context.Context, db database.DB, user models.User) (models.SubscriptionState, error) {
	plan, err := s.effectivePlan(ctx, user)
	if err != nil {
		return models.SubscriptionState{}, err
	}

	// Counting usage:
	// AI evaluations are counted for the current day.
	// For free plans, full mocks are counted lifetime (1 free mock test included on signup + bonuses).
	// For paid active plans, full mocks are counted monthly.
	var evaluationsToday, mocksUsed int
	if plan.ID == "free" {
		err = db.QueryRow(ctx, `
			SELECT
				(SELECT count(*) FROM ai_evaluations
				  WHERE user_id = $1 AND created_at >= date_trunc('day', now())),
				(SELECT count(*) FROM mock_attempts ma
				  JOIN mocks m ON ma.mock_id = m.id
				  WHERE ma.user_id = $1 AND NOT m.is_diagnostic)`,
			user.ID).Scan(&evaluationsToday, &mocksUsed)
	} else {
		err = db.QueryRow(ctx, `
			SELECT
				(SELECT count(*) FROM ai_evaluations
				  WHERE user_id = $1 AND created_at >= date_trunc('day', now())),
				(SELECT count(*) FROM mock_attempts ma
				  JOIN mocks m ON ma.mock_id = m.id
				  WHERE ma.user_id = $1 AND NOT m.is_diagnostic AND ma.completed_at >= date_trunc('month', now()))`,
			user.ID).Scan(&evaluationsToday, &mocksUsed)
	}

	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("read usage: %w", err)
	}

	totalMockTestsAllowed := plan.MockTestsIncluded + user.BonusMockTests

	state := models.SubscriptionState{
		PlanID:                plan.ID,
		PlanName:              plan.Name,
		IsActive:              planIsActive(user),
		DailyEvaluationsUsed:  evaluationsToday,
		DailyEvaluationsLimit: plan.AIEvaluationsPerDay,
		MockTestsIncluded:     plan.MockTestsIncluded,
		BonusMockTests:        user.BonusMockTests,
		TotalMockTestsAllowed: totalMockTestsAllowed,
		MockTestsUsed:         mocksUsed,
		BonusDays:             plan.BonusDays,
	}
	if user.PlanValidUntil != nil {
		state.ValidUntil = user.PlanValidUntil.Format(time.DateOnly)
	}
	return state, nil
}

// CheckEvaluationAllowance returns an error when the learner has used up today's AI evaluations.
func (s *Service) CheckEvaluationAllowance(ctx context.Context, db database.DB, user models.User) (models.SubscriptionState, error) {
	state, err := s.State(ctx, db, user)
	if err != nil {
		return state, err
	}
	if state.DailyEvaluationsUsed >= state.DailyEvaluationsLimit {
		return state, ErrLimitReached
	}
	return state, nil
}

// CheckMockAllowance checks if the learner has available full mock tests in their quota.
func (s *Service) CheckMockAllowance(ctx context.Context, db database.DB, user models.User) (models.SubscriptionState, error) {
	state, err := s.State(ctx, db, user)
	if err != nil {
		return state, err
	}
	if state.MockTestsUsed >= state.TotalMockTestsAllowed {
		return state, ErrMockLimitReached
	}
	return state, nil
}

type ConfirmPaymentParams struct {
	UserID         string
	PlanID         string
	PaymentGateway string
	TransactionID  string
	AmountNPR      int
}

// ConfirmPayment processes a successful subscription purchase, granting base duration plus purchase bonus days.
// It is strictly idempotent against duplicate webhooks or retries.
func (s *Service) ConfirmPayment(ctx context.Context, pool *pgxpool.Pool, p ConfirmPaymentParams) (models.SubscriptionState, error) {
	if p.UserID == "" || p.PlanID == "" || p.TransactionID == "" {
		return models.SubscriptionState{}, ErrInvalidPayment
	}

	plan, err := s.repo.Plan(ctx, p.PlanID)
	if err != nil {
		return models.SubscriptionState{}, err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("begin payment confirmation tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if this transaction ID was already confirmed (idempotency)
	var existingStatus string
	err = tx.QueryRow(ctx, `
		SELECT status FROM subscription_payments
		WHERE transaction_id = $1`, p.TransactionID).Scan(&existingStatus)
	if err == nil && existingStatus == "success" {
		// Already processed successfully! Return current state without duplicating entitlement.
		var u models.User
		_ = tx.QueryRow(ctx, `SELECT id, email, password_hash, name, role, target_exam, target_score, exam_date, nepal_region, xp, streak_days, streak_last_active_date, timezone, plan_id, plan_valid_until, referral_code, bonus_mock_tests, bonus_pro_days, created_at FROM users WHERE id = $1`, p.UserID).
			Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.TargetExam, &u.TargetScore, &u.ExamDate, &u.NepalRegion, &u.XP, &u.StreakDays, &u.StreakLastActiveDate, &u.Timezone, &u.PlanID, &u.PlanValidUntil, &u.ReferralCode, &u.BonusMockTests, &u.BonusProDays, &u.CreatedAt)
		return s.State(ctx, tx, u)
	}

	// Calculate base + purchase bonus days:
	// Weekly: 7 days (+0 bonus) = 7 days
	// Normal (pro): 30 days (+3 bonus) = 33 days
	// Max (elite): 90 days (+7 bonus) = 97 days
	baseDays := plan.DurationDays
	if baseDays <= 0 && plan.DurationMonths > 0 {
		baseDays = plan.DurationMonths * 30
	}
	bonusDays := plan.BonusDays
	effectiveDays := baseDays + bonusDays

	// Update user's plan and expiration date transactionally
	var updatedUser models.User
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET plan_id = $2,
		    plan_valid_until = COALESCE(GREATEST(plan_valid_until, CURRENT_DATE), CURRENT_DATE) + ($3 || ' days')::INTERVAL,
		    updated_at = now()
		WHERE id = $1
		RETURNING id, email, password_hash, name, role, target_exam, target_score, exam_date, nepal_region, xp, streak_days, streak_last_active_date, timezone, plan_id, plan_valid_until, referral_code, bonus_mock_tests, bonus_pro_days, created_at`,
		p.UserID, plan.ID, effectiveDays).
		Scan(&updatedUser.ID, &updatedUser.Email, &updatedUser.PasswordHash, &updatedUser.Name,
			&updatedUser.Role, &updatedUser.TargetExam, &updatedUser.TargetScore, &updatedUser.ExamDate,
			&updatedUser.NepalRegion, &updatedUser.XP, &updatedUser.StreakDays, &updatedUser.StreakLastActiveDate,
			&updatedUser.Timezone, &updatedUser.PlanID, &updatedUser.PlanValidUntil,
			&updatedUser.ReferralCode, &updatedUser.BonusMockTests, &updatedUser.BonusProDays, &updatedUser.CreatedAt)
	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("update user plan: %w", err)
	}

	// Record subscription payment in ledger
	_, err = tx.Exec(ctx, `
		INSERT INTO subscription_payments (
			user_id, plan_id, payment_gateway, transaction_id,
			amount_npr, status, base_days, bonus_days, effective_days, processed_at
		)
		VALUES ($1, $2, $3, $4, $5, 'success', $6, $7, $8, now())
		ON CONFLICT (transaction_id) DO UPDATE SET
			status = 'success',
			processed_at = now()`,
		p.UserID, plan.ID, p.PaymentGateway, p.TransactionID,
		p.AmountNPR, baseDays, bonusDays, effectiveDays)
	if err != nil {
		return models.SubscriptionState{}, fmt.Errorf("record subscription payment: %w", err)
	}

	// Send confirmation notification
	if s.notifications != nil {
		msg := fmt.Sprintf("Your %s plan is activated for %d days (%d base + %d purchase bonus days)!", plan.Name, effectiveDays, baseDays, bonusDays)
		if bonusDays == 0 {
			msg = fmt.Sprintf("Your %s plan is activated for %d days!", plan.Name, effectiveDays)
		}
		_ = s.notifications.Create(ctx, tx, notifications.CreateParams{
			UserID:    p.UserID,
			Title:     "Subscription Activated! 🎉",
			Message:   msg,
			Type:      "system",
			ActionURL: "/subscription",
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return models.SubscriptionState{}, fmt.Errorf("commit payment confirmation: %w", err)
	}

	return s.State(ctx, pool, updatedUser)
}

// effectivePlan falls back to the free plan once a paid plan has lapsed, so an
// expired subscription cannot keep its higher limits.
func (s *Service) effectivePlan(ctx context.Context, user models.User) (models.Plan, error) {
	if !planIsActive(user) {
		return s.repo.Plan(ctx, "free")
	}
	plan, err := s.repo.Plan(ctx, user.PlanID)
	if errors.Is(err, ErrPlanNotFound) {
		return s.repo.Plan(ctx, "free")
	}
	return plan, err
}

func planIsActive(user models.User) bool {
	if user.PlanID == "free" {
		return true
	}
	if user.PlanValidUntil == nil {
		return false
	}
	return user.PlanValidUntil.After(time.Now())
}
