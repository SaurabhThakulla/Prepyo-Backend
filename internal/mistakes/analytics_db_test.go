package mistakes

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/testdb"
)

// Monthly is four weekly buckets, weekly is seven daily ones, and the totals
// count exactly what those buckets hold.
func TestAnalyticsPeriodBuckets(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	var userID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, name, plan_id, timezone, referral_code)
		VALUES ('mistakes-' || gen_random_uuid() || '@test.local', 'Buckets', 'free', 'Asia/Kathmandu',
		        'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("create user: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID) })

	// One mistake per offset, each on its own question (mistakes are unique per question).
	offsets := []int{0, 3, 10, 20, 27, 40, 200}
	rows, err := pool.Query(ctx, `SELECT id, exam FROM questions ORDER BY id LIMIT $1`, len(offsets))
	if err != nil {
		t.Fatalf("read questions: %v", err)
	}
	type q struct{ id, exam string }
	var qs []q
	for rows.Next() {
		var item q
		if err := rows.Scan(&item.id, &item.exam); err != nil {
			t.Fatal(err)
		}
		qs = append(qs, item)
	}
	rows.Close()
	if len(qs) < len(offsets) {
		t.Skipf("need %d seeded questions, found %d", len(offsets), len(qs))
	}
	for i, days := range offsets {
		if _, err := pool.Exec(ctx, `
			INSERT INTO mistakes (user_id, question_id, exam, error_tag, user_response, correct_response, explanation, last_attempted_at)
			VALUES ($1, $2, $3, '', '', '', '', CURRENT_DATE - $4::int + INTERVAL '12 hours')`,
			userID, qs[i].id, qs[i].exam, days); err != nil {
			t.Fatalf("insert mistake: %v", err)
		}
	}

	repo := NewRepository(pool)

	monthly, err := repo.Analytics(ctx, AnalyticsParams{UserID: userID, Period: "monthly"})
	if err != nil {
		t.Fatalf("monthly: %v", err)
	}
	if len(monthly.Timeline) != 4 {
		t.Fatalf("monthly buckets = %d, want 4", len(monthly.Timeline))
	}
	// The 4th week is the last 7 days (0, 3), the 3rd holds 10, the 2nd 20, the 1st 27.
	want := []int{1, 1, 1, 2}
	sum := 0
	for i, point := range monthly.Timeline {
		if point.Label != []string{"1st week", "2nd week", "3rd week", "4th week"}[i] {
			t.Errorf("bucket %d label = %q", i, point.Label)
		}
		if point.Count != want[i] {
			t.Errorf("%s count = %d, want %d", point.Label, point.Count, want[i])
		}
		sum += point.Count
	}
	if monthly.Total != sum {
		t.Errorf("monthly total %d does not match its buckets (%d)", monthly.Total, sum)
	}

	weekly, err := repo.Analytics(ctx, AnalyticsParams{UserID: userID, Period: "weekly"})
	if err != nil {
		t.Fatalf("weekly: %v", err)
	}
	if len(weekly.Timeline) != 7 {
		t.Fatalf("weekly buckets = %d, want 7", len(weekly.Timeline))
	}
	sum = 0
	for _, point := range weekly.Timeline {
		sum += point.Count
	}
	// Only the mistakes from today and 3 days ago fall in the last 7 days.
	if sum != 2 || weekly.Total != 2 {
		t.Errorf("weekly buckets sum %d, total %d, want 2 each", sum, weekly.Total)
	}

	// A period the bank no longer offers falls back to weekly rather than all-time.
	fallback, err := repo.Analytics(ctx, AnalyticsParams{UserID: userID, Period: "lifetime"})
	if err != nil {
		t.Fatalf("fallback: %v", err)
	}
	if fallback.Period != "weekly" || fallback.Total != 2 {
		t.Errorf("unknown period gave %q with %d, want weekly with 2", fallback.Period, fallback.Total)
	}
}
