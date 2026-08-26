// Package exams serves exam versions.
//
// A version pins the scoring scale for an exam. Attempts record the version
// they ran under, so changing the current version never rewrites old results.
package exams

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("exam version not found")

const selectFields = `id, exam, label, description, min_score, max_score, score_step, is_current`

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

// Current returns the live version for an exam. Practice and mocks use this to
// stamp new attempts.
func (r *Repository) Current(ctx context.Context, exam models.ExamType) (models.ExamVersion, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectFields+` FROM exam_versions WHERE exam = $1 AND is_current`, exam)
	return scanOne(row)
}

func (r *Repository) ByID(ctx context.Context, id string) (models.ExamVersion, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectFields+` FROM exam_versions WHERE id = $1`, id)
	return scanOne(row)
}

func (r *Repository) List(ctx context.Context) ([]models.ExamVersion, error) {
	rows, err := r.db.Query(ctx, `SELECT `+selectFields+` FROM exam_versions ORDER BY exam, id`)
	if err != nil {
		return nil, fmt.Errorf("list exam versions: %w", err)
	}
	defer rows.Close()

	versions := []models.ExamVersion{}
	for rows.Next() {
		var v models.ExamVersion
		if err := rows.Scan(&v.ID, &v.Exam, &v.Label, &v.Description, &v.MinScore, &v.MaxScore, &v.ScoreStep, &v.IsCurrent); err != nil {
			return nil, fmt.Errorf("scan exam version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func scanOne(row pgx.Row) (models.ExamVersion, error) {
	var v models.ExamVersion
	err := row.Scan(&v.ID, &v.Exam, &v.Label, &v.Description, &v.MinScore, &v.MaxScore, &v.ScoreStep, &v.IsCurrent)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ExamVersion{}, ErrNotFound
	}
	if err != nil {
		return models.ExamVersion{}, fmt.Errorf("get exam version: %w", err)
	}
	return v, nil
}
