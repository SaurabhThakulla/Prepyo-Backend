package mocks

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/testdb"
)

func TestMockResponsesPersistAndUnavailableMocksAreHidden(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testdb.URL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	// Temporary tables remove unrelated foreign keys without modifying real users.
	for _, sql := range []string{
		`CREATE TEMP TABLE mock_attempts (LIKE public.mock_attempts INCLUDING ALL) ON COMMIT DROP`,
		`CREATE TEMP TABLE mocks (LIKE public.mocks INCLUDING ALL) ON COMMIT DROP`,
		`INSERT INTO mocks SELECT * FROM public.mocks`,
	} {
		if _, err = tx.Exec(ctx, sql); err != nil {
			t.Fatal(err)
		}
	}
	repo := NewRepository(tx)
	var unavailable string
	if err = tx.QueryRow(ctx, `SELECT id FROM mocks WHERE NOT is_available LIMIT 1`).Scan(&unavailable); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ByID(ctx, unavailable); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unavailable mock exposed: %v", err)
	}
	list, err := repo.List(ctx, models.ExamIELTS)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range list {
		if m.ID == unavailable {
			t.Fatal("unavailable mock listed")
		}
	}
	answers := []models.AnswerSubmission{{QuestionID: "saved-question", TextResponse: "My original response", SelectedOptions: []string{"B"}}}
	a, err := repo.SaveAttempt(ctx, tx, SaveAttemptParams{UserID: "00000000-0000-0000-0000-000000000001", MockID: "test", ExamVersionID: "ielts-2026-01", Exam: models.ExamIELTS, SkillScores: map[models.SkillType]float64{}, Answers: answers})
	if err != nil {
		t.Fatal(err)
	}
	var raw []byte
	if err = tx.QueryRow(ctx, `SELECT answers FROM mock_attempts WHERE id=$1`, a.ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var got []models.AnswerSubmission
	if err = json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].TextResponse != answers[0].TextResponse || got[0].SelectedOptions[0] != "B" {
		t.Fatalf("lost response: %s", raw)
	}
}
