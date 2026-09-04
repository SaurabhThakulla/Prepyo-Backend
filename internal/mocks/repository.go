// Package mocks runs full-length mock exams.
//
// A mock is graded from the answers the learner submitted, question by
// question, using the same scoring package as practice. Results are stamped
// with the exam version they ran under.
package mocks

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("mock not found")

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

// List returns blueprints with their sections.
func (r *Repository) List(ctx context.Context, exam models.ExamType) ([]models.Mock, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, exam_version_id, exam, title, description, total_duration_minutes,
		       is_diagnostic, is_generated
		FROM mocks
		WHERE ($1 = '' OR exam = $1)
		ORDER BY is_diagnostic DESC, id`, exam)
	if err != nil {
		return nil, fmt.Errorf("list mocks: %w", err)
	}
	defer rows.Close()

	list := []models.Mock{}
	ids := []string{}
	for rows.Next() {
		var m models.Mock
		if err := rows.Scan(&m.ID, &m.ExamVersionID, &m.Exam, &m.Title, &m.Description,
			&m.TotalDurationMinutes, &m.IsDiagnostic, &m.IsGenerated); err != nil {
			return nil, fmt.Errorf("scan mock: %w", err)
		}
		list = append(list, m)
		ids = append(ids, m.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sections, err := r.sectionsFor(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Sections = sections[list[i].ID]
		for _, s := range list[i].Sections {
			list[i].TotalQuestions += len(s.QuestionIDs)
		}
	}
	return list, nil
}

func (r *Repository) ByID(ctx context.Context, id string) (models.Mock, error) {
	var m models.Mock
	err := r.db.QueryRow(ctx, `
		SELECT id, exam_version_id, exam, title, description, total_duration_minutes,
		       is_diagnostic, is_generated
		FROM mocks WHERE id = $1`, id).
		Scan(&m.ID, &m.ExamVersionID, &m.Exam, &m.Title, &m.Description,
			&m.TotalDurationMinutes, &m.IsDiagnostic, &m.IsGenerated)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Mock{}, ErrNotFound
	}
	if err != nil {
		return models.Mock{}, fmt.Errorf("get mock: %w", err)
	}

	sections, err := r.sectionsFor(ctx, []string{id})
	if err != nil {
		return models.Mock{}, err
	}
	m.Sections = sections[id]
	for _, s := range m.Sections {
		m.TotalQuestions += len(s.QuestionIDs)
	}
	return m, nil
}

// sectionsFor loads sections for several mocks at once, keeping List to two
// queries regardless of how many blueprints come back.
func (r *Repository) sectionsFor(ctx context.Context, mockIDs []string) (map[string][]models.MockSection, error) {
	byMock := map[string][]models.MockSection{}
	if len(mockIDs) == 0 {
		return byMock, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT mock_id, id, name, skill, duration_minutes, question_ids
		FROM mock_sections
		WHERE mock_id = ANY($1)
		ORDER BY mock_id, position`, mockIDs)
	if err != nil {
		return nil, fmt.Errorf("list mock sections: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var mockID string
		var s models.MockSection
		if err := rows.Scan(&mockID, &s.ID, &s.Name, &s.Skill, &s.DurationMinutes, &s.QuestionIDs); err != nil {
			return nil, fmt.Errorf("scan mock section: %w", err)
		}
		byMock[mockID] = append(byMock[mockID], s)
	}
	return byMock, rows.Err()
}

type SaveAttemptParams struct {
	UserID          string
	MockID          string
	ExamVersionID   string
	Exam            models.ExamType
	UserScore       float64
	SkillScores     map[models.SkillType]float64
	TotalCorrect    int
	TotalQuestions  int
	DurationSeconds int
}

func (r *Repository) SaveAttempt(ctx context.Context, db database.DB, p SaveAttemptParams) (models.MockAttempt, error) {
	var a models.MockAttempt
	err := db.QueryRow(ctx, `
		INSERT INTO mock_attempts (
			user_id, mock_id, exam_version_id, exam, user_score, skill_scores,
			total_correct, total_questions, duration_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, mock_id, exam_version_id, exam, user_score, skill_scores,
		          total_correct, total_questions, duration_seconds, completed_at`,
		p.UserID, p.MockID, p.ExamVersionID, p.Exam, p.UserScore, p.SkillScores,
		p.TotalCorrect, p.TotalQuestions, p.DurationSeconds,
	).Scan(&a.ID, &a.MockID, &a.ExamVersionID, &a.Exam, &a.UserScore, &a.SkillScores,
		&a.TotalCorrect, &a.TotalQuestions, &a.DurationSeconds, &a.CompletedAt)
	if err != nil {
		return models.MockAttempt{}, fmt.Errorf("save mock attempt: %w", err)
	}
	return a, nil
}

func (r *Repository) Attempts(ctx context.Context, userID string, limit, offset int) ([]models.MockAttempt, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM mock_attempts WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count mock attempts: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, mock_id, exam_version_id, exam, user_score, skill_scores,
		       total_correct, total_questions, duration_seconds, completed_at
		FROM mock_attempts
		WHERE user_id = $1
		ORDER BY completed_at DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list mock attempts: %w", err)
	}
	defer rows.Close()

	list := []models.MockAttempt{}
	for rows.Next() {
		var a models.MockAttempt
		if err := rows.Scan(&a.ID, &a.MockID, &a.ExamVersionID, &a.Exam, &a.UserScore,
			&a.SkillScores, &a.TotalCorrect, &a.TotalQuestions, &a.DurationSeconds, &a.CompletedAt); err != nil {
			return nil, 0, fmt.Errorf("scan mock attempt: %w", err)
		}
		list = append(list, a)
	}
	return list, total, rows.Err()
}
