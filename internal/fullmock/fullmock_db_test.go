package fullmock

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/testdb"
)

// fakeSection deals real rows in a section's table, marked as a full mock's,
// so the full mock reads their status as it would a section mock's paper.
type fakeSection struct {
	pool  *pgxpool.Pool
	skill models.SkillType
	dealt *int
}

func (f fakeSection) Deal(ctx context.Context, user models.User) (string, any, error) {
	*f.dealt++
	var id string
	var err error
	switch f.skill {
	case models.SkillListening:
		err = f.pool.QueryRow(ctx, `
			INSERT INTO listening_mock_sessions (user_id, test_id, question_ids, duration_minutes, expires_at, in_full_mock)
			VALUES ($1, (SELECT id FROM listening_tests LIMIT 1), '{}', 45, now() + interval '45 minutes', TRUE)
			RETURNING id::text`, user.ID).Scan(&id)
	case models.SkillReading:
		err = f.pool.QueryRow(ctx, `
			INSERT INTO reading_mock_sessions (user_id, mock_id, exam, exam_version_id, passage_ids, question_ids, duration_minutes, expires_at, in_full_mock)
			VALUES ($1, 'mock-ielts-full', 'IELTS', 'ielts-2026-01', '{}', '{}', 60, now() + interval '60 minutes', TRUE)
			RETURNING id::text`, user.ID).Scan(&id)
	case models.SkillWriting:
		err = f.pool.QueryRow(ctx, `
			INSERT INTO writing_mock_sessions (user_id, exam, module, task1_id, task2_id, duration_minutes, expires_at, in_full_mock)
			VALUES ($1, 'IELTS', 'academic', (SELECT id FROM questions LIMIT 1), (SELECT id FROM questions LIMIT 1), 60, now() + interval '60 minutes', TRUE)
			RETURNING id::text`, user.ID).Scan(&id)
	case models.SkillSpeaking:
		err = f.pool.QueryRow(ctx, `
			INSERT INTO speaking_mock_sessions (user_id, set_id, expires_at, in_full_mock)
			VALUES ($1, (SELECT id FROM speaking_mock_sets LIMIT 1), now() + interval '40 minutes', TRUE)
			RETURNING id::text`, user.ID).Scan(&id)
	}
	return id, map[string]string{"id": id}, err
}

func (f fakeSection) Resume(_ context.Context, _ models.User, id string) (any, error) {
	return map[string]string{"id": id}, nil
}

func setup(t *testing.T) (*Service, *billing.Service, *pgxpool.Pool, models.User) {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	var user models.User
	err = pool.QueryRow(context.Background(), `
		INSERT INTO users (email, name, plan_id, timezone, target_exam, referral_code)
		VALUES ('fullmock-' || gen_random_uuid() || '@test.local', 'Full Mock Test', 'free', 'Asia/Kathmandu',
		        'IELTS', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id, plan_id, timezone, target_exam, bonus_mock_tests`).
		Scan(&user.ID, &user.PlanID, &user.Timezone, &user.TargetExam, &user.BonusMockTests)
	if err != nil {
		t.Fatalf("create learner: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, user.ID) })
	billingService := billing.NewService(billing.NewRepository(pool), nil)
	starters := map[models.SkillType]Starter{}
	for _, skill := range Order {
		starters[skill] = fakeSection{pool: pool, skill: skill, dealt: new(int)}
	}
	return NewService(pool, mocks.NewRepository(pool), billingService, starters), billingService, pool, user
}

// finishSection records a section as submitted with a band, as its section
// mock does on submission.
func finishSection(t *testing.T, pool *pgxpool.Pool, user models.User, skill models.SkillType, sessionID string, band float64) {
	t.Helper()
	ctx := context.Background()
	var attemptID string
	err := pool.QueryRow(ctx, `
		INSERT INTO mock_attempts (user_id, mock_id, exam_version_id, exam, user_score, skill_scores,
		                           total_correct, total_questions, duration_seconds)
		VALUES ($1, 'mock-ielts-`+string(skill)+`-gen', 'ielts-2026-01', 'IELTS', $2::float8, jsonb_build_object($3::text, $2::float8), 0, 0, 60)
		RETURNING id::text`, user.ID, band, string(skill)).Scan(&attemptID)
	if err != nil {
		t.Fatalf("record %s attempt: %v", skill, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE `+sectionTables[skill]+` SET status = 'submitted', mock_attempt_id = $2 WHERE id = $1`,
		sessionID, attemptID); err != nil {
		t.Fatalf("submit %s: %v", skill, err)
	}
}

// A full mock spends the allowance when it starts, once however often the
// start is retried, and the attempt it records at the end is not counted again.
func TestFullMockSpendsAllowanceOnceAtStart(t *testing.T) {
	ctx := context.Background()
	svc, billingService, pool, user := setup(t)

	first, err := svc.Start(ctx, user)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	again, err := svc.Start(ctx, user)
	if err != nil || again.ID != first.ID {
		t.Fatalf("second start = %s (%v), want the open mock %s", again.ID, err, first.ID)
	}
	state, err := billingService.State(ctx, pool, user)
	if err != nil {
		t.Fatal(err)
	}
	if state.MockTestsUsed != 1 {
		t.Fatalf("mock tests used = %d after starting, want 1", state.MockTestsUsed)
	}

	for i, skill := range Order {
		session, _, err := svc.StartSection(ctx, user, first.ID, skill)
		if err != nil {
			t.Fatalf("start %s: %v", skill, err)
		}
		finishSection(t, pool, user, skill, session.Sections[i].SessionID, 6)
	}
	if _, err := svc.Finish(ctx, user, first.ID); err != nil {
		t.Fatalf("finish: %v", err)
	}
	state, err = billingService.State(ctx, pool, user)
	if err != nil {
		t.Fatal(err)
	}
	if state.MockTestsUsed != 1 {
		t.Fatalf("mock tests used = %d after finishing, want still 1", state.MockTestsUsed)
	}

	// The free plan includes one full mock; it has been used.
	if state.TotalMockTestsAllowed == 1 {
		if _, err := svc.Start(ctx, user); !errors.Is(err, billing.ErrMockLimitReached) {
			t.Fatalf("start past the allowance = %v, want ErrMockLimitReached", err)
		}
	}
}

// Sections are taken in the order of the test, and the overall band is the
// mean of the four rounded to the nearest half band.
func TestFullMockRunsSectionsInOrder(t *testing.T) {
	ctx := context.Background()
	svc, _, pool, user := setup(t)
	session, err := svc.Start(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if session.Stage != models.SkillListening {
		t.Fatalf("first stage = %q, want listening", session.Stage)
	}
	if _, _, err := svc.StartSection(ctx, user, session.ID, models.SkillReading); !errors.Is(err, ErrOutOfOrder) {
		t.Fatalf("reading before listening = %v, want ErrOutOfOrder", err)
	}
	if _, err := svc.Finish(ctx, user, session.ID); !errors.Is(err, ErrSectionsUnfinished) {
		t.Fatalf("finish with nothing done = %v, want ErrSectionsUnfinished", err)
	}

	bands := map[models.SkillType]float64{
		models.SkillListening: 7, models.SkillReading: 6.5, models.SkillWriting: 6, models.SkillSpeaking: 7.5,
	}
	for i, skill := range Order {
		state, paper, err := svc.StartSection(ctx, user, session.ID, skill)
		if err != nil {
			t.Fatalf("start %s: %v", skill, err)
		}
		if paper == nil || state.Sections[i].Status != SectionInProgress || state.Stage != skill {
			t.Fatalf("%s section = %+v at stage %q, want in progress", skill, state.Sections[i], state.Stage)
		}
		// A retried start reopens the same paper rather than dealing another.
		retry, _, err := svc.StartSection(ctx, user, session.ID, skill)
		if err != nil || retry.Sections[i].SessionID != state.Sections[i].SessionID {
			t.Fatalf("retry %s = %q (%v), want %q", skill, retry.Sections[i].SessionID, err, state.Sections[i].SessionID)
		}
		if dealt := *svc.starters[skill].(fakeSection).dealt; dealt != 1 {
			t.Fatalf("%s papers dealt = %d, want 1", skill, dealt)
		}
		finishSection(t, pool, user, skill, state.Sections[i].SessionID, bands[skill])
	}

	done, err := svc.Finish(ctx, user, session.ID)
	if err != nil {
		t.Fatalf("finish: %v", err)
	}
	// (7 + 6.5 + 6 + 7.5) / 4 = 6.75, which rounds up to 7.
	if done.Status != "completed" || done.OverallBand == nil || *done.OverallBand != 7 {
		t.Fatalf("finished mock = %+v, want completed with an overall of 7", done)
	}
	var score float64
	var scores map[string]float64
	if err := pool.QueryRow(ctx, `SELECT user_score::float8, skill_scores FROM mock_attempts WHERE id = $1`, done.AttemptID).
		Scan(&score, &scores); err != nil {
		t.Fatalf("read attempt: %v", err)
	}
	if score != 7 || len(scores) != 4 || scores["writing"] != 6 {
		t.Fatalf("attempt = %v %v, want 7 with the four section bands", score, scores)
	}
	if again, err := svc.Finish(ctx, user, session.ID); err != nil || again.AttemptID != done.AttemptID {
		t.Fatalf("finishing twice = %+v (%v), want the same result", again, err)
	}
}

// A section that ends without a band leaves the full mock without an overall.
func TestFullMockWithoutAllBandsHasNoOverall(t *testing.T) {
	ctx := context.Background()
	svc, _, pool, user := setup(t)
	session, err := svc.Start(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	for i, skill := range Order {
		state, _, err := svc.StartSection(ctx, user, session.ID, skill)
		if err != nil {
			t.Fatalf("start %s: %v", skill, err)
		}
		id := state.Sections[i].SessionID
		if skill == models.SkillWriting {
			// A writing paper whose tasks could not be rated is submitted
			// without a band.
			if _, err := pool.Exec(ctx, `UPDATE writing_mock_sessions SET status = 'submitted' WHERE id = $1`, id); err != nil {
				t.Fatal(err)
			}
			continue
		}
		finishSection(t, pool, user, skill, id, 6)
	}
	done, err := svc.Finish(ctx, user, session.ID)
	if err != nil {
		t.Fatalf("finish: %v", err)
	}
	if done.OverallBand != nil || done.AttemptID != "" || done.Sections[2].Status != SectionClosed {
		t.Fatalf("finished mock = %+v, want no overall and writing closed", done)
	}
}

func TestOverallRoundsToHalfBands(t *testing.T) {
	band := func(v float64) *float64 { return &v }
	for _, tc := range []struct {
		bands []float64
		want  float64
	}{
		{[]float64{6.5, 6.5, 5, 7}, 6.5}, // 6.25 rounds up to 6.5
		{[]float64{4, 3.5, 4, 4}, 4},     // 3.875 rounds to 4
		{[]float64{6.5, 6.5, 5.5, 6}, 6}, // 6.125 rounds down to 6
		{[]float64{7, 6.5, 6, 7.5}, 7},   // 6.75 rounds up to 7
	} {
		sections := make([]SectionState, len(tc.bands))
		for i, b := range tc.bands {
			sections[i] = SectionState{Band: band(b)}
		}
		if got := Overall(sections); got == nil || *got != tc.want {
			t.Errorf("Overall(%v) = %v, want %v", tc.bands, got, tc.want)
		}
	}
	if got := Overall([]SectionState{{Band: band(6)}, {Band: band(6)}, {Band: band(6)}, {}}); got != nil {
		t.Errorf("Overall with a missing band = %v, want none", *got)
	}
}

// Ending a full mock closes the section paper it has open.
func TestAbandoningAFullMockClosesItsOpenSection(t *testing.T) {
	ctx := context.Background()
	svc, _, pool, user := setup(t)
	session, err := svc.Start(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	state, _, err := svc.StartSection(ctx, user, session.ID, models.SkillListening)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Abandon(ctx, user, session.ID); err != nil {
		t.Fatalf("abandon: %v", err)
	}
	var status string
	if err := pool.QueryRow(ctx, `SELECT status FROM listening_mock_sessions WHERE id = $1`, state.Sections[0].SessionID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "abandoned" {
		t.Fatalf("open listening paper = %s after ending the full mock, want abandoned", status)
	}
	if err := svc.Abandon(ctx, user, session.ID); !errors.Is(err, ErrNotInProgress) {
		t.Fatalf("second abandon = %v, want ErrNotInProgress", err)
	}
}
