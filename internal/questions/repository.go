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

	// IncludePassageQuestions brings reading questions that belong to a passage
	// into the results. Off by default: see List.
	IncludePassageQuestions bool
}

// List returns published questions matching the filters. Empty filter fields
// mean "any".
//
// Questions attached to a reading passage are left out unless asked for. They
// are not standalone tasks — the text they are about lives in
// reading_passages, not in the question row — so serving one here would hand a
// learner a question with nothing to read. They are served with their passage
// by /api/v1/reading instead.
func (r *Repository) List(ctx context.Context, p ListParams) ([]models.Question, int, error) {
	const where = `
		WHERE is_published
		  AND ($1 = '' OR exam = $1)
		  AND ($2 = '' OR skill = $2)
		  AND ($3 = '' OR type_id = $3)
		  AND ($4 OR passage_id IS NULL)`

	var total int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM questions`+where,
		p.Exam, p.Skill, p.TypeID, p.IncludePassageQuestions).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count questions: %w", err)
	}

	rows, err := r.db.Query(ctx, `
		SELECT `+selectFields+` FROM questions`+where+`
		ORDER BY id
		LIMIT $5 OFFSET $6`,
		p.Exam, p.Skill, p.TypeID, p.IncludePassageQuestions, p.Limit, p.Offset)
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

// ByGroupIDs returns the questions belonging to several reading groups in one
// round trip, keyed by group id and in their authored order within each group.
//
// It lives here rather than in internal/reading so that there stays exactly one
// piece of code that knows how a question row is shaped. Answer keys come back
// attached: the caller decides whether it is grading (keep them) or serving
// (strip them with PublicQuestion).
func (r *Repository) ByGroupIDs(ctx context.Context, groupIDs []string) (map[string][]models.Question, error) {
	byGroup := map[string][]models.Question{}
	if len(groupIDs) == 0 {
		return byGroup, nil
	}

	rows, err := r.db.Query(ctx, `
		SELECT group_id, `+selectFields+`
		FROM questions
		WHERE group_id = ANY($1) AND is_published
		ORDER BY group_id, group_position, id`, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("list group questions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var groupID string
		q, err := scanWithPrefix(rows, &groupID)
		if err != nil {
			return nil, err
		}
		byGroup[groupID] = append(byGroup[groupID], q)
	}
	return byGroup, rows.Err()
}

func scan(row pgx.Row) (models.Question, error) {
	var q models.Question
	err := row.Scan(fieldsOf(&q)...)
	if err != nil {
		return models.Question{}, err
	}
	return q, nil
}

// scanWithPrefix scans a row that selects one extra column before selectFields.
func scanWithPrefix(row pgx.Row, prefix any) (models.Question, error) {
	var q models.Question
	if err := row.Scan(append([]any{prefix}, fieldsOf(&q)...)...); err != nil {
		return models.Question{}, err
	}
	return q, nil
}

// fieldsOf lists scan targets in the order of selectFields. The two are a pair:
// change one and change the other.
func fieldsOf(q *models.Question) []any {
	return []any{
		&q.ID, &q.ExamVersionID, &q.Exam, &q.Skill, &q.TypeID, &q.TypeName,
		&q.Title, &q.Prompt, &q.ContextPassage, &q.AudioURL, &q.AudioTranscript,
		&q.ImageURL, &q.PrepTimeSeconds, &q.TimeLimitSeconds,
		&q.Options, &q.CorrectAnswers, &q.Blanks, &q.ModelAnswer, &q.Explanation,
		&q.Difficulty, &q.Tags, &q.Points,
	}
}
