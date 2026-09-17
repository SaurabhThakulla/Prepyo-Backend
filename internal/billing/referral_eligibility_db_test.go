package billing

import (
	"context"
	"github.com/prepyo/backend/internal/referrals"
	"testing"
)

func TestReferralIneligiblePaymentsAndRollback(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	cases := []struct {
		name, status, plan string
		amount, days       int
		processed          bool
	}{
		{"pending", "pending", "pro", 299, 30, true},
		{"failed", "failed", "pro", 299, 30, true},
		{"cancelled", "cancelled", "pro", 299, 30, true},
		{"refunded", "refunded", "pro", 299, 30, true},
		{"unprocessed", "success", "pro", 299, 30, false},
		{"zero amount", "success", "pro", 0, 30, true},
		{"free plan", "success", "free", 299, 30, true},
		{"zero days", "success", "pro", 299, 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			owner, friend := linkFriends(t, pool)
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if err := referrals.LockLifecycle(ctx, tx); err != nil {
				t.Fatal(err)
			}
			var id string
			err = tx.QueryRow(ctx, `INSERT INTO subscription_payments
   (user_id,plan_id,payment_gateway,transaction_id,amount_npr,status,base_days,effective_days,processed_at)
   VALUES ($1::uuid,$2,'esewa',$1::text,$3,$4,$5,$5,CASE WHEN $6 THEN now() END) RETURNING id`, friend.ID, tc.plan, tc.amount, tc.status, tc.days, tc.processed).Scan(&id)
			if err != nil {
				t.Fatal(err)
			}
			if err := referrals.QualifyPayment(ctx, tx, id); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			assertReferralDays(t, pool, owner, 0)
		})
	}
	t.Run("reward rolls back with payment", func(t *testing.T) {
		owner, friend := linkFriends(t, pool)
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if err := referrals.LockLifecycle(ctx, tx); err != nil {
			t.Fatal(err)
		}
		var id string
		if err := tx.QueryRow(ctx, `INSERT INTO subscription_payments
   (user_id,plan_id,payment_gateway,transaction_id,amount_npr,status,base_days,effective_days,processed_at)
   VALUES ($1::uuid,'pro','esewa',$1::text,299,'success',30,30,now()) RETURNING id`, friend.ID).Scan(&id); err != nil {
			t.Fatal(err)
		}
		if err := referrals.QualifyPayment(ctx, tx, id); err != nil {
			t.Fatal(err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		assertReferralDays(t, pool, owner, 0)
		var status string
		if err := pool.QueryRow(ctx, `SELECT status FROM referrals WHERE referee_id=$1`, friend.ID).Scan(&status); err != nil || status != "pending" {
			t.Fatalf("rollback status=%s err=%v", status, err)
		}
	})
}

func TestReferralBonusPreservesPaidTierAndRepairsExpired(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	for _, tc := range []struct {
		plan, role, wantPlan, wantRole string
		remaining, wantDays            int
	}{
		{"elite", "udaan", "elite", "udaan", 10, 13},
		{"weekly", "abhyas", "weekly", "abhyas", 10, 13},
		{"elite", "udaan", "pro", "taiyari", -10, 3},
		{"free", "admin", "pro", "admin", -10, 3},
	} {
		t.Run(tc.role+tc.wantRole, func(t *testing.T) {
			owner := newLearner(t, pool)
			if _, err := pool.Exec(ctx, `UPDATE users SET plan_id=$2,role=$3,plan_valid_until=CURRENT_DATE+$4::int WHERE id=$1`, owner.ID, tc.plan, tc.role, tc.remaining); err != nil {
				t.Fatal(err)
			}
			if err := referrals.NewRepository(pool).AddBonusProDays(ctx, pool, owner.ID, 3); err != nil {
				t.Fatal(err)
			}
			var plan, role string
			var days int
			if err := pool.QueryRow(ctx, `SELECT plan_id,role,plan_valid_until-CURRENT_DATE FROM users WHERE id=$1`, owner.ID).Scan(&plan, &role, &days); err != nil {
				t.Fatal(err)
			}
			if plan != tc.wantPlan || role != tc.wantRole || days != tc.wantDays {
				t.Fatalf("got %s/%s/%d want %s/%s/%d", plan, role, days, tc.wantPlan, tc.wantRole, tc.wantDays)
			}
		})
	}
}
