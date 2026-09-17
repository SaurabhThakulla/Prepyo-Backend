package billing

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/referrals"
)

func referralService(pool *pgxpool.Pool) *referrals.Service {
	return referrals.NewService(pool, referrals.NewRepository(pool), nil, nil, "https://test.local", slog.Default())
}

func linkFriends(t *testing.T, pool *pgxpool.Pool) (models.User, models.User) {
	t.Helper()
	owner, friend := newLearner(t, pool), newLearner(t, pool)
	if err := pool.QueryRow(context.Background(), `SELECT referral_code FROM users WHERE id=$1`, owner.ID).Scan(&owner.ReferralCode); err != nil {
		t.Fatal(err)
	}
	if _, err := referralService(pool).Redeem(context.Background(), friend.ID, "  "+strings.ToLower(owner.ReferralCode)+"  "); err != nil {
		t.Fatal(err)
	}
	return owner, friend
}

func assertReferralDays(t *testing.T, pool *pgxpool.Pool, owner models.User, want int) {
	t.Helper()
	var days int
	if err := pool.QueryRow(context.Background(), `SELECT bonus_pro_days FROM users WHERE id=$1`, owner.ID).Scan(&days); err != nil {
		t.Fatal(err)
	}
	if days != want {
		t.Fatalf("bonus days=%d, want %d", days, want)
	}
}

func TestReferralConfirmedPurchaseExactlyOnce(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	owner, friend := linkFriends(t, pool)
	svc := newService(pool)
	plan, err := svc.repo.Plan(ctx, "pro")
	if err != nil {
		t.Fatal(err)
	}
	assertReferralDays(t, pool, owner, 0)
	if _, err := referralService(pool).Redeem(ctx, friend.ID, owner.ReferralCode); !errors.Is(err, referrals.ErrAlreadyReferred) {
		t.Fatalf("repeat redemption: %v", err)
	}
	p := ConfirmPaymentParams{UserID: friend.ID, PlanID: plan.ID, AmountNPR: plan.PriceNPR, PaymentGateway: "esewa", TransactionID: "referral-" + friend.ID}
	if err := svc.RequestPayment(ctx, pool, RequestPaymentParams{UserID: friend.ID, Plan: plan, PaymentGateway: p.PaymentGateway, TransactionID: p.TransactionID}); err != nil {
		t.Fatal(err)
	}
	assertReferralDays(t, pool, owner, 0)
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := svc.ConfirmPayment(ctx, pool, p); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	assertReferralDays(t, pool, owner, 3)
	var active bool
	if err := pool.QueryRow(ctx, `SELECT plan_id='pro' AND role='taiyari' AND plan_valid_until=CURRENT_DATE+3 FROM users WHERE id=$1`, owner.ID).Scan(&active); err != nil || !active {
		t.Fatalf("bonus must activate entitlement: %v, %v", active, err)
	}
	overview, err := referralService(pool).Overview(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	if overview.Stats.Completed != 1 || overview.Stats.TotalBonusDaysEarned != 3 || len(overview.RecentReferrals) != 1 || overview.RecentReferrals[0].RewardDays != 3 {
		t.Fatalf("incorrect overview: %+v", overview)
	}
	p.TransactionID += "-renewal"
	if _, err := svc.ConfirmPayment(ctx, pool, p); err != nil {
		t.Fatal(err)
	}
	assertReferralDays(t, pool, owner, 3)
	var queued, notices int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM queued_plans WHERE user_id=$1), (SELECT count(*) FROM notifications WHERE user_id=$2 AND type='referral')`, friend.ID, owner.ID).Scan(&queued, &notices); err != nil {
		t.Fatal(err)
	}
	if queued != 1 || notices != 1 {
		t.Fatalf("queued=%d notices=%d; want 1 each", queued, notices)
	}
}

func TestReferralRejectsInvalidRedemption(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	owner, _ := linkFriends(t, pool)
	svc := referralService(pool)
	if _, err := svc.Redeem(ctx, owner.ID, owner.ReferralCode); !errors.Is(err, referrals.ErrSelfReferral) {
		t.Fatalf("self: %v", err)
	}
	if _, err := svc.Redeem(ctx, owner.ID, "missing-code"); !errors.Is(err, referrals.ErrCodeNotFound) {
		t.Fatalf("unknown: %v", err)
	}
	learner := newLearner(t, pool)
	bill := newService(pool)
	plan, err := bill.repo.Plan(ctx, "pro")
	if err != nil {
		t.Fatal(err)
	}
	if err := bill.RequestPayment(ctx, pool, RequestPaymentParams{UserID: learner.ID, Plan: plan, PaymentGateway: "esewa", TransactionID: "prior-" + learner.ID}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Redeem(ctx, learner.ID, owner.ReferralCode); !errors.Is(err, referrals.ErrPurchaseExists) {
		t.Fatalf("after payment: %v", err)
	}
}
