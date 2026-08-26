// Package questions serves the shared question bank.
//
// Questions are content, not learner data: the same row is served to everyone.
// The answer key never reaches the browser while a question is being answered
// (see models.Question.PublicQuestion).
package questions

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var ErrNotFound = errors.New("question not found")

const selectFields = `
	id, exam_version_id, exam, skill, type_id, type_name, title, prompt,
	COALESCE(context_passage, ''), COALESCE(audio_url, ''), COALESCE(audio_transcript, ''),
	COALESCE(image_url, ''), prep_time_seconds, time_limit_seconds,
	COALESCE(options, '[]'::jsonb), COALESCE(correct_answers, '[]'::jsonb),
	COALESCE(blanks, '[]'::jsonb), COALESCE(model_answer, ''), COALESCE(explanation, ''),
	difficulty, tags, points`

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

type ListParams struct {
	Exam   models.ExamType
	Skill  models.SkillType
	TypeID string
	Limit  int
	Offset int
}

// List returns published questions matching the filters. Empty filter fields
// mean "any".
func (r *Repository) List(ctx context.Context, p ListParams) ([]models.Question, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `
		SELECT count(*) FROM questions
		WHERE is_published
		  AND ($1 = '' OR exam = $1)
		  AND ($2 = '' OR skill = $2)
		  AND ($3 = '' OR type_id = $3)`,
		p.Exam, p.Skill, p.TypeID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count questions: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+selectFields+` FROM questions
		WHERE is_published
		  AND ($1 = '' OR exam = $1)
		  AND ($2 = '' OR skill = $2)
		  AND ($3 = '' OR type_id = $3)
		ORDER BY id
		LIMIT $4 OFFSET $5`,
		p.Exam, p.Skill, p.TypeID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list questions: %w", err)
	}
	defer rows.Close()

	list := []models.Question{}
	for rows.Next() {
		q, err := scan(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, q)
	}
	return list, total, rows.Err()
}

// ByID returns one question including its answer key. Callers serving a
// learner must call PublicQuestion before sending it out.
func (r *Repository) ByID(ctx context.Context, id string) (models.Question, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectFields+` FROM questions WHERE id = $1 AND is_published`, id)

	q, err := scan(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Question{}, ErrNotFound
	}
	if err != nil {
		return models.Question{}, fmt.Errorf("get question: %w", err)
	}
	return q, nil
}

// ByIDs returns several questions in one round trip, keyed by id.
func (r *Repository) ByIDs(ctx context.Context, ids []string) (map[string]models.Question, error) {
	if len(ids) == 0 {
		return map[string]models.Question{}, nil
	}

	rows, err := r.db.Query(ctx, `SELECT `+selectFields+` FROM questions WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, fmt.Errorf("get questions: %w", err)
	}
	defer rows.Close()

	found := make(map[string]models.Question, len(ids))
	for rows.Next() {
		q, err := scan(rows)
		if err != nil {
			return nil, err
		}
		found[q.ID] = q
	}
	return found, rows.Err()
}

func scan(row pgx.Row) (models.Question, error) {
	var q models.Question
	err := row.Scan(
		&q.ID, &q.ExamVersionID, &q.Exam, &q.Skill, &q.TypeID, &q.TypeName,
		&q.Title, &q.Prompt, &q.ContextPassage, &q.AudioURL, &q.AudioTranscript,
		&q.ImageURL, &q.PrepTimeSeconds, &q.TimeLimitSeconds,
		&q.Options, &q.CorrectAnswers, &q.Blanks, &q.ModelAnswer, &q.Explanation,
		&q.Difficulty, &q.Tags, &q.Points,
	)
	if err != nil {
		return models.Question{}, err
	}
	return q, nil
}
