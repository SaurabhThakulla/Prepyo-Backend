package referrals

import (
	"context"
	"errors"

	"github.com/prepyo/backend/internal/models"
)

var ErrPurchaseExists = errors.New("enter a referral code before creating your first subscription purchase")

func (s *Service) Redeem(ctx context.Context, userID, code string) (*models.Referral, error) {
	if NormalizeCode(code) == "" {
		return nil, ErrCodeNotFound
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if err = LockLifecycle(ctx, tx); err != nil {
		return nil, err
	}
	var exists bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM referrals WHERE referee_id = $1)`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrAlreadyReferred
	}
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM subscription_payments WHERE user_id = $1 AND status IN ('pending', 'success', 'refunded'))`, userID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrPurchaseExists
	}
	ref, err := s.LinkReferralOnRegister(ctx, tx, userID, code)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return ref, nil
}
