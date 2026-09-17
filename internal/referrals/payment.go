package referrals

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/notifications"
)

const RewardMembershipDays = 3

// LockLifecycle serializes the infrequent redemption/payment operations before
// they lock payment or user rows. This also prevents reciprocal referrals from
// deadlocking when both friends purchase concurrently. Hold until commit.
func LockLifecycle(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(724901234)`)
	return err
}

// QualifyPayment must run INSIDE the authoritative payment transaction, after
// success is recorded and before commit. No activity or client request calls it.
// A relationship earns once, on its first eligible purchase after redemption.
func QualifyPayment(ctx context.Context, tx pgx.Tx, paymentID string) error {
	var referralID, referrerID, refereeID string
	var referrerXP, refereeXP int
	err := tx.QueryRow(ctx, `
  SELECT r.id, r.referrer_id, r.referee_id, r.reward_referrer_xp, r.reward_referee_xp
  FROM referrals r
  JOIN subscription_payments p ON p.user_id = r.referee_id
  JOIN plans pl ON pl.id = p.plan_id
  WHERE p.id = $1 AND p.status = 'success' AND p.processed_at IS NOT NULL
    AND p.amount_npr > 0 AND pl.price_npr > 0 AND p.plan_id <> 'free'
    AND p.effective_days > 0 AND r.created_at <= p.created_at
    AND r.status <> 'cancelled' AND r.reward_payment_id IS NULL
  FOR UPDATE OF r, p`, paymentID).Scan(&referralID, &referrerID, &refereeID, &referrerXP, &refereeXP)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find qualifying referral: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE referrals SET status = 'completed', completed_at = now(),
  reward_payment_id = $2, reward_days = $3 WHERE id = $1`, referralID, paymentID, RewardMembershipDays); err != nil {
		return err
	}
	if err = NewRepository(tx).AddBonusProDays(ctx, tx, referrerID, RewardMembershipDays); err != nil {
		return err
	}
	// Retain the existing XP benefits, using their original idempotency keys so
	// legacy activity-qualified referrals cannot receive the XP a second time.
	xp := gamification.NewService()
	for _, award := range []gamification.AwardParams{
		{UserID: referrerID, Amount: referrerXP, Reason: "Friend completed a paid subscription purchase", SourceKey: "referral:" + referralID + ":referrer"},
		{UserID: refereeID, Amount: refereeXP, Reason: "Welcome referral bonus", SourceKey: "referral:" + referralID + ":referee"},
	} {
		if _, err = xp.Award(ctx, tx, award); err != nil {
			return err
		}
	}
	return notifications.NewRepository(tx).Create(ctx, tx, notifications.CreateParams{
		UserID: referrerID, Title: "Referral reward earned!", Message: "Your friend's subscription payment was confirmed. You earned 3 bonus membership days!",
		Type: "referral", ActionURL: "/referrals",
	})
}
