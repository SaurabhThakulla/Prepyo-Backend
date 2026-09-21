package evaluations

import (
	"context"
	"testing"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/testdb"
)

// Database-backed test for storing an evaluation. Run with:
//
//	TEST_DATABASE_URL=postgres://postgres@localhost:5432/prepyo?sslmode=disable go test ./internal/evaluations/
//
// It exists because the bug it covers was invisible everywhere else: strengths
// and weaknesses are jsonb columns, and a Go slice sent straight at one is
// encoded as a Postgres array literal, {"one","two"}, which the column rejects
// as invalid JSON. Every evaluation that listed a strength - which is nearly
// all of them - failed to save, and the learner was told something had gone
// wrong on our side. An empty list happens to encode as {}, valid JSON, so the
// path looked fine whenever the model returned nothing to praise.
func TestSaveStoresStrengthsAndWeaknesses(t *testing.T) {
	url := testdb.URL(t)
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping database-backed evaluation tests")
	}
	ctx := context.Background()
	// The pool has to be built the way the app builds it. database.Connect
	// falls back to QueryExecModeExec for pooler compatibility, and that is the
	// mode where this bug lives: without a statement description pgx has no
	// parameter types to go on, so a Go slice is encoded as a Postgres array
	// rather than as JSON. A default pool describes its statements, learns the
	// column is jsonb, and quietly does the right thing - which is why this
	// never reproduced outside the running server.
	pool, err := database.Connect(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)

	var userID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (email, name, plan_id, timezone, referral_code)
		VALUES ('evaluations-' || gen_random_uuid() || '@test.local', 'Evaluation Test', 'free',
		        'Asia/Kathmandu', 'TEST-' || substr(replace(gen_random_uuid()::text, '-', ''), 1, 10))
		RETURNING id::text`).Scan(&userID); err != nil {
		t.Fatalf("create learner: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID) })

	repo := NewRepository(pool)
	score := 6.5
	saved, err := repo.Save(ctx, pool, SaveParams{
		UserID:      userID,
		Fingerprint: []byte("strengths-round-trip"),
		Evaluation: models.Evaluation{
			Exam:              models.ExamIELTS,
			Skill:             models.SkillSpeaking,
			EvaluationVersion: "v1",
			EstimatedScore:    &score,
			ScoreConfidence:   "low",
			Summary:           "A provisional estimate from a transcript.",
			Strengths:         []string{"Answers both parts of the question", "Ideas are ordered clearly"},
			Weaknesses:        []string{"Tense errors repeat"},
			Criteria: []models.EvaluationCriterion{
				{Name: "Fluency and Coherence", Score: 6.5, MaxScore: 9, Feedback: "Clear enough to follow."},
			},
			Transcript: "i live near the ring road and it is very peaceful",
		},
		Usage: models.EvaluationUsage{Provider: "test", Model: "test-model", PromptVersion: "test.v1"},
	})
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	if len(saved.Strengths) != 2 || saved.Strengths[0] != "Answers both parts of the question" {
		t.Errorf("strengths came back as %#v", saved.Strengths)
	}
	if len(saved.Weaknesses) != 1 || saved.Weaknesses[0] != "Tense errors repeat" {
		t.Errorf("weaknesses came back as %#v", saved.Weaknesses)
	}
	if saved.Transcript == "" || len(saved.Criteria) != 1 {
		t.Errorf("evaluation came back incomplete: %+v", saved)
	}
}
