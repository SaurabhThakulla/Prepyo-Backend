package billing

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// A learner may have only MaxPendingPayments waiting: each holds a
// screenshot, so the cap keeps one account from filling the disk.
func TestPendingPaymentsAreCapped(t *testing.T) {
	pool := gradingPool(t)
	ctx := context.Background()
	learner := newLearner(t, pool)
	svc := newService(pool)
	plan, err := svc.repo.Plan(ctx, "pro")
	if err != nil {
		t.Fatal(err)
	}
	request := func(n int) error {
		return svc.RequestPayment(ctx, pool, RequestPaymentParams{
			UserID: learner.ID, Plan: plan, PaymentGateway: "ESEWA",
			TransactionID: fmt.Sprintf("cap-%s-%d", learner.ID, n), PhoneNumber: "9800000000",
			ProofImage: []byte("proof"), ProofImageType: "image/png",
		})
	}
	for n := 1; n <= MaxPendingPayments; n++ {
		if err := request(n); err != nil {
			t.Fatalf("payment %d of %d refused: %v", n, MaxPendingPayments, err)
		}
	}
	if err := request(MaxPendingPayments + 1); !errors.Is(err, ErrTooManyPending) {
		t.Fatalf("payment over the cap: err = %v, want ErrTooManyPending", err)
	}

	// Once one is reviewed, the learner may send another.
	if _, err := pool.Exec(ctx, `
		UPDATE subscription_payments SET status = 'failed'
		 WHERE id = (SELECT id FROM subscription_payments WHERE user_id = $1 ORDER BY created_at LIMIT 1)`, learner.ID); err != nil {
		t.Fatal(err)
	}
	if err := request(MaxPendingPayments + 2); err != nil {
		t.Fatalf("payment after a review refused: %v", err)
	}
}
