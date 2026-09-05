// Package reading serves passage-driven reading work.
//
// It answers two requests that the flat question bank cannot:
//
//   - "give me a Matching Information set" — one random passage that carries
//     that task type, dealt with its questions in a random order;
//   - "give me a reading mock" — three passages this learner has never sat,
//     each carrying the full spread of task types.
//
// Both rest on the same rule: a passage is the unit of content, and what a
// learner has already read is remembered. Practice comes back to a passage only
// once the rest of the bank has been used; a mock does not come back to one at
// all until there is nothing left to deal.
//
// Grading is not reimplemented here. A generated mock produces an ordinary
// mock_attempts row scored by internal/mocks, and a single practice answer goes
// through /api/v1/practice like every other question, because a reading
// question is a row in `questions` like every other question.
package reading

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

var (
	ErrNoPassage       = errors.New("no passage available")
	ErrSessionNotFound = errors.New("reading mock session not found")
	ErrNoBlueprint     = errors.New("no generated reading mock blueprint for this exam")
)

// Exposure contexts. They are stored separately because the rules differ: a
// mock must not re-deal a passage the learner has sat, while practice may.
const (
	ContextPractice = "practice"
	ContextMock     = "mock"
)

// Session statuses.
const (
	StatusInProgress = "in_progress"
	StatusSubmitted  = "submitted"
	StatusAbandoned  = "abandoned"
)

type Repository struct {
	db database.DB
}

func NewRepository(db database.DB) *Repository {
	return &Repository{db: db}
}

const passageFields = `
	id, exam_version_id, exam, title, subtitle, paragraphs, sources,
	word_count, difficulty, topic, tags`

// ---------------------------------------------------------------------------
// Passages
// ---------------------------------------------------------------------------

func (r *Repository) PassageByID(ctx context.Context, id string) (models.ReadingPassage, error) {
	row := r.db.QueryRow(ctx, `SELECT `+passageFields+`
		FROM reading_passages WHERE id = $1 AND is_published`, id)

	p, err := scanPassage(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ReadingPassage{}, ErrNoPassage
	}
	if err != nil {
		return models.ReadingPassage{}, fmt.Errorf("get passage: %w", err)
	}
	return p, nil
}

// PassagesByIDs loads several passages at once, keyed by id.
func (r *Repository) PassagesByIDs(ctx context.Context, ids []string) (map[string]models.ReadingPassage, error) {
	found := map[string]models.ReadingPassage{}
	if len(ids) == 0 {
		return found, nil
	}

	rows, err := r.db.Query(ctx, `SELECT `+passageFields+`
		FROM reading_passages WHERE id = ANY($1) AND is_published`, ids)
	if err != nil {
		return nil, fmt.Errorf("get passages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		p, err := scanPassage(rows)
		if err != nil {
			return nil, fmt.Errorf("scan passage: %w", err)
		}
		found[p.ID] = p
	}
	return found, rows.Err()
}

type ListPassagesParams struct {
	Exam   models.ExamType
	TypeID string
	Limit  int
	Offset int
}

// ListPassages is the passage index: what is in the bank, newest ids last.
func (r *Repository) ListPassages(ctx context.Context, p ListPassagesParams) ([]models.ReadingPassage, int, error) {
	// The type filter is a question about the groups on a passage, so both
	// statements share the same EXISTS rather than the count drifting from the
	// page it is counting.
	const where = `
		WHERE is_published
		  AND ($1 = '' OR exam = $1)
		  AND ($2 = '' OR EXISTS (
			  SELECT 1 FROM reading_question_groups g
			  WHERE g.passage_id = reading_passages.id AND g.type_id = $2))`

	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM reading_passages`+where,
		p.Exam, p.TypeID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count passages: %w", err)
	}

	rows, err := r.db.Query(ctx, `SELECT `+passageFields+` FROM reading_passages`+where+`
		ORDER BY id
		LIMIT $3 OFFSET $4`, p.Exam, p.TypeID, p.Limit, p.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list passages: %w", err)
	}
	defer rows.Close()

	list := []models.ReadingPassage{}
	for rows.Next() {
		passage, err := scanPassage(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan passage: %w", err)
		}
		list = append(list, passage)
	}
	return list, total, rows.Err()
}

func scanPassage(row pgx.Row) (models.ReadingPassage, error) {
	var p models.ReadingPassage
	err := row.Scan(&p.ID, &p.ExamVersionID, &p.Exam, &p.Title, &p.Subtitle,
		&p.Paragraphs, &p.Sources, &p.WordCount, &p.Difficulty, &p.Topic, &p.Tags)
	return p, err
}

// ---------------------------------------------------------------------------
// Groups
// ---------------------------------------------------------------------------

const groupFields = `
	id, passage_id, position, type_id, type_name, instructions, resources,
	paper_slot, passage_display, shuffle_questions, time_limit_seconds`

// Group is a stored group. It carries shuffleQuestions, which the service needs
// and the client does not: whether a set may be dealt out of order is a
// property of the task, not something the browser should be asked to respect.
type Group struct {
	models.ReadingGroup
	ShuffleQuestions bool
}

// GroupsForPassages loads the groups on several passages, in passage order then
// position order. An empty typeID means every group.
func (r *Repository) GroupsForPassages(ctx context.Context, passageIDs []string, typeID string) ([]Group, error) {
	if len(passageIDs) == 0 {
		return nil, nil
	}

	rows, err := r.db.Query(ctx, `SELECT `+groupFields+`
		FROM reading_question_groups
		WHERE passage_id = ANY($1) AND ($2 = '' OR type_id = $2)
		ORDER BY passage_id, position`, passageIDs, typeID)
	if err != nil {
		return nil, fmt.Errorf("list groups: %w", err)
	}
	defer rows.Close()

	var list []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.PassageID, &g.Position, &g.TypeID, &g.TypeName,
			&g.Instructions, &g.Resources, &g.PaperSlot, &g.PassageDisplay,
			&g.ShuffleQuestions, &g.TimeLimitSeconds); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		list = append(list, g)
	}
	return list, rows.Err()
}

// GroupByID loads one group.
func (r *Repository) GroupByID(ctx context.Context, id string) (Group, error) {
	var g Group
	err := r.db.QueryRow(ctx, `SELECT `+groupFields+`
		FROM reading_question_groups WHERE id = $1`, id).
		Scan(&g.ID, &g.PassageID, &g.Position, &g.TypeID, &g.TypeName,
			&g.Instructions, &g.Resources, &g.PaperSlot, &g.PassageDisplay,
			&g.ShuffleQuestions, &g.TimeLimitSeconds)
	if errors.Is(err, pgx.ErrNoRows) {
		return Group{}, ErrNoPassage
	}
	if err != nil {
		return Group{}, fmt.Errorf("get group: %w", err)
	}
	return g, nil
}

// TaskTypes is the practice menu: every reading task type in the bank, with the
// number of passages and questions behind it.
//
// The counts are the point. A type with no published questions must be shown as
// unavailable rather than offered and then failing to deal a set.
func (r *Repository) TaskTypes(ctx context.Context, exam models.ExamType) ([]models.ReadingTaskType, error) {
	rows, err := r.db.Query(ctx, `
		SELECT type_id, type_name, passage_count, question_count FROM (
			SELECT g.type_id                      AS type_id,
			       min(g.type_name)               AS type_name,
			       count(DISTINCT g.passage_id)   AS passage_count,
			       count(q.id)                    AS question_count
			FROM reading_question_groups g
			JOIN reading_passages p ON p.id = g.passage_id AND p.is_published
			JOIN questions q ON q.group_id = g.id AND q.is_published
			WHERE ($1 = '' OR p.exam = $1)
			GROUP BY g.type_id

			UNION ALL

			-- Re-order Paragraphs has no passages, so an item counts as one.
			-- The menu is asking "how much is there to work through", and for
			-- this task an item is the unit a learner is dealt.
			SELECT 'reorder-paragraphs',
			       'Re-order Paragraphs',
			       count(DISTINCT i.id),
			       count(q.id)
			FROM reading_reorder_items i
			JOIN questions q ON q.reorder_item_id = i.id AND q.is_published
			WHERE i.is_published AND ($1 = '' OR i.exam = $1)
			HAVING count(q.id) > 0
		) menu
		ORDER BY type_name`, exam)
	if err != nil {
		return nil, fmt.Errorf("list reading types: %w", err)
	}
	defer rows.Close()

	list := []models.ReadingTaskType{}
	for rows.Next() {
		var t models.ReadingTaskType
		if err := rows.Scan(&t.TypeID, &t.TypeName, &t.PassageCount, &t.QuestionCount); err != nil {
			return nil, fmt.Errorf("scan reading type: %w", err)
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------

// PickPracticeGroup chooses one group of the requested type for this learner.
//
// Passages the learner has never practised come first; after that the one they
// met longest ago. Ties break at random, so a learner with a fresh bank gets a
// genuinely random passage and a learner who has worked through it gets an even
// rotation rather than the same passage every time.
func (r *Repository) PickPracticeGroup(ctx context.Context, userID string, exam models.ExamType, typeID string) (Group, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		SELECT g.id
		FROM reading_question_groups g
		JOIN reading_passages p ON p.id = g.passage_id AND p.is_published
		LEFT JOIN user_passage_exposures e
		       ON e.user_id = $1 AND e.passage_id = p.id AND e.context = 'practice'
		WHERE g.type_id = $2
		  AND ($3 = '' OR p.exam = $3)
		  AND EXISTS (SELECT 1 FROM questions q WHERE q.group_id = g.id AND q.is_published)
		ORDER BY e.last_seen_at ASC NULLS FIRST, random()
		LIMIT 1`, userID, typeID, exam).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Group{}, ErrNoPassage
	}
	if err != nil {
		return Group{}, fmt.Errorf("pick practice group: %w", err)
	}
	return r.GroupByID(ctx, id)
}

// MockCandidate is a passage a mock could use, and whether this learner has
// already sat it.
type MockCandidate struct {
	PassageID  string
	SeenInMock bool
}

// PickMockPassages chooses the passages for one generated mock.
//
// Only passages carrying all three paper slots are eligible. A paper deals three
// passages and takes one section from each, but which passage lands in which
// slot is decided after they are picked — so every candidate has to be able to
// fill any of the three. Among those, ones the learner has never sat in a mock
// come first — that is the no-repeat rule — then ones they have not met in
// practice either, then the ones sat longest ago.
//
// It returns fewer than `count` when the bank cannot fill the paper. Deciding
// what to do about that is the service's job, not a silent truncation here.
func (r *Repository) PickMockPassages(
	ctx context.Context,
	userID string,
	exam models.ExamType,
	count int,
) ([]MockCandidate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, (seen.passage_id IS NOT NULL)
		FROM reading_passages p
		LEFT JOIN user_passage_exposures seen
		       ON seen.user_id = $1 AND seen.passage_id = p.id AND seen.context = 'mock'
		LEFT JOIN user_passage_exposures practised
		       ON practised.user_id = $1 AND practised.passage_id = p.id AND practised.context = 'practice'
		WHERE p.is_published
		  AND p.exam = $2
		  AND (
			  SELECT count(DISTINCT g.paper_slot)
			  FROM reading_question_groups g
			  WHERE g.passage_id = p.id
			    AND g.paper_slot > 0
			    AND EXISTS (SELECT 1 FROM questions q WHERE q.group_id = g.id AND q.is_published)
		  ) = 3
		ORDER BY (seen.passage_id IS NOT NULL),
		         (practised.passage_id IS NOT NULL),
		         seen.last_seen_at ASC NULLS FIRST,
		         random()
		LIMIT $3`, userID, exam, count)
	if err != nil {
		return nil, fmt.Errorf("pick mock passages: %w", err)
	}
	defer rows.Close()

	var picked []MockCandidate
	for rows.Next() {
		var c MockCandidate
		if err := rows.Scan(&c.PassageID, &c.SeenInMock); err != nil {
			return nil, fmt.Errorf("scan mock candidate: %w", err)
		}
		picked = append(picked, c)
	}
	return picked, rows.Err()
}

// ---------------------------------------------------------------------------
// Re-order Paragraphs
// ---------------------------------------------------------------------------

// ErrNoReorderItem means the bank holds no item this learner can be dealt.
var ErrNoReorderItem = errors.New("no re-order item available")

// PickReorderItem chooses one item for this learner, with its backing question.
//
// Two rules, in order. An item derived from a passage the learner has already
// read is skipped outright: they have seen those sentences in the right order,
// so re-ordering them is not a task any more. Among what is left, unseen items
// come first and otherwise the one met longest ago, which is the same rule
// PickPracticeGroup applies to passages.
func (r *Repository) PickReorderItem(ctx context.Context, userID string, exam models.ExamType) (models.ReadingReorderItem, string, error) {
	var item models.ReadingReorderItem
	var questionID string
	var sourcePassageID *string

	err := r.db.QueryRow(ctx, `
		SELECT i.id, i.exam_version_id, i.exam, i.title, i.paragraphs,
		       i.source_passage_id, i.topic, i.word_count, i.difficulty, i.tags,
		       q.id
		FROM reading_reorder_items i
		JOIN questions q ON q.reorder_item_id = i.id AND q.is_published
		LEFT JOIN user_reorder_exposures e
		       ON e.user_id = $1 AND e.item_id = i.id AND e.context = 'practice'
		WHERE i.is_published
		  AND ($2 = '' OR i.exam = $2)
		  AND (
			  i.source_passage_id IS NULL
			  OR NOT EXISTS (
				  SELECT 1 FROM user_passage_exposures seen
				  WHERE seen.user_id = $1 AND seen.passage_id = i.source_passage_id
			  )
		  )
		ORDER BY e.last_seen_at ASC NULLS FIRST, random()
		LIMIT 1`, userID, exam).
		Scan(&item.ID, &item.ExamVersionID, &item.Exam, &item.Title, &item.Paragraphs,
			&sourcePassageID, &item.Topic, &item.WordCount, &item.Difficulty, &item.Tags,
			&questionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.ReadingReorderItem{}, "", ErrNoReorderItem
	}
	if err != nil {
		return models.ReadingReorderItem{}, "", fmt.Errorf("pick re-order item: %w", err)
	}
	if sourcePassageID != nil {
		item.SourcePassageID = *sourcePassageID
	}
	return item, questionID, nil
}

// RecordReorderExposure marks items as dealt to this learner. The mirror of
// RecordExposure, and called at the same point: when the content is handed over,
// not when it is finished with.
func (r *Repository) RecordReorderExposure(ctx context.Context, db database.DB, userID string, itemIDs []string, exposureContext string) error {
	if len(itemIDs) == 0 {
		return nil
	}

	_, err := db.Exec(ctx, `
		INSERT INTO user_reorder_exposures (user_id, item_id, context)
		SELECT $1, unnest($2::text[]), $3
		ON CONFLICT (user_id, item_id, context) DO UPDATE
		SET seen_count   = user_reorder_exposures.seen_count + 1,
		    last_seen_at = now()`, userID, itemIDs, exposureContext)
	if err != nil {
		return fmt.Errorf("record re-order exposure: %w", err)
	}
	return nil
}

// RecordExposure marks passages as met by this learner.
//
// It is called when the content is handed over, not when it is finished with. A
// learner who opens a mock and closes the tab has still read those passages,
// and giving them the same three next time would defeat the point.
func (r *Repository) RecordExposure(ctx context.Context, db database.DB, userID string, passageIDs []string, exposureContext string) error {
	if len(passageIDs) == 0 {
		return nil
	}

	_, err := db.Exec(ctx, `
		INSERT INTO user_passage_exposures (user_id, passage_id, context)
		SELECT $1, unnest($2::text[]), $3
		ON CONFLICT (user_id, passage_id, context) DO UPDATE
		SET seen_count   = user_passage_exposures.seen_count + 1,
		    last_seen_at = now()`, userID, passageIDs, exposureContext)
	if err != nil {
		return fmt.Errorf("record passage exposure: %w", err)
	}
	return nil
}

// SeenPassageIDs returns the passages this learner has met in a context. It
// backs the "you have sat 6 of 40 passages" line, not the selection itself.
func (r *Repository) SeenPassageIDs(ctx context.Context, userID, exposureContext string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT passage_id FROM user_passage_exposures
		WHERE user_id = $1 AND context = $2
		ORDER BY last_seen_at DESC`, userID, exposureContext)
	if err != nil {
		return nil, fmt.Errorf("list seen passages: %w", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan seen passage: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ---------------------------------------------------------------------------
// Mock sessions
// ---------------------------------------------------------------------------

// Blueprint is the `mocks` row a generated paper is recorded against. Generated
// mocks still produce ordinary mock_attempts, and those reference a mock id.
type Blueprint struct {
	ID              string
	ExamVersionID   string
	Title           string
	DurationMinutes int
}

func (r *Repository) GeneratedBlueprint(ctx context.Context, exam models.ExamType) (Blueprint, error) {
	var b Blueprint
	err := r.db.QueryRow(ctx, `
		SELECT id, exam_version_id, title, total_duration_minutes
		FROM mocks
		WHERE exam = $1 AND is_generated
		ORDER BY id
		LIMIT 1`, exam).Scan(&b.ID, &b.ExamVersionID, &b.Title, &b.DurationMinutes)
	if errors.Is(err, pgx.ErrNoRows) {
		return Blueprint{}, ErrNoBlueprint
	}
	if err != nil {
		return Blueprint{}, fmt.Errorf("get generated blueprint: %w", err)
	}
	return b, nil
}

// Session is a stored paper. QuestionIDs is the whole point of the row: it is
// the paper as dealt, so grading cannot be widened by the client.
type Session struct {
	models.ReadingMockSession
	QuestionIDs []string
}

const sessionFields = `
	s.id, s.mock_id, m.title, s.exam, s.exam_version_id, s.status,
	s.duration_minutes, s.passage_ids, s.question_ids, s.reused_passages,
	s.created_at, s.submitted_at`

const sessionFrom = ` FROM reading_mock_sessions s JOIN mocks m ON m.id = s.mock_id`

type CreateSessionParams struct {
	UserID          string
	MockID          string
	Exam            models.ExamType
	ExamVersionID   string
	PassageIDs      []string
	QuestionIDs     []string
	ReusedPassages  bool
	DurationMinutes int
}

// ErrSessionOpen means this learner already has a live paper for this exam.
var ErrSessionOpen = errors.New("a reading mock is already in progress")

// CreateSession stores a dealt paper.
//
// A learner may hold one live paper per exam. The unique index enforcing that
// is what makes a retried start request resume the first paper instead of
// spending three more passages on a second one; the error is translated here so
// the caller can do the resuming.
func (r *Repository) CreateSession(ctx context.Context, db database.DB, p CreateSessionParams) (Session, error) {
	row := db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO reading_mock_sessions (
				user_id, mock_id, exam, exam_version_id, passage_ids, question_ids,
				reused_passages, duration_minutes)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING *
		)
		SELECT `+sessionFields+` FROM inserted s JOIN mocks m ON m.id = s.mock_id`,
		p.UserID, p.MockID, p.Exam, p.ExamVersionID, p.PassageIDs, p.QuestionIDs,
		p.ReusedPassages, p.DurationMinutes)

	s, err := scanSession(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Session{}, ErrSessionOpen
		}
		return Session{}, fmt.Errorf("create reading mock session: %w", err)
	}
	return s, nil
}

// LiveSession returns this learner's in-progress paper for an exam.
func (r *Repository) LiveSession(ctx context.Context, userID string, exam models.ExamType) (Session, error) {
	row := r.db.QueryRow(ctx, `SELECT `+sessionFields+sessionFrom+`
		WHERE s.user_id = $1 AND s.exam = $2 AND s.status = 'in_progress'`, userID, exam)

	s, err := scanSession(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("get live reading mock session: %w", err)
	}
	return s, nil
}

// SessionByID scopes the lookup to the owner, so one learner cannot read
// another's paper by guessing a uuid.
func (r *Repository) SessionByID(ctx context.Context, db database.DB, userID, sessionID string) (Session, error) {
	row := db.QueryRow(ctx, `SELECT `+sessionFields+sessionFrom+`
		WHERE s.id = $1 AND s.user_id = $2`, sessionID, userID)

	s, err := scanSession(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("get reading mock session: %w", err)
	}
	return s, nil
}

// ListSessions returns this learner's papers, newest first, without their
// question lists filled in.
func (r *Repository) ListSessions(ctx context.Context, userID string, limit, offset int) ([]models.ReadingMockSession, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM reading_mock_sessions WHERE user_id = $1`,
		userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reading mock sessions: %w", err)
	}

	rows, err := r.db.Query(ctx, `SELECT `+sessionFields+sessionFrom+`
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list reading mock sessions: %w", err)
	}
	defer rows.Close()

	list := []models.ReadingMockSession{}
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan reading mock session: %w", err)
		}
		list = append(list, s.ReadingMockSession)
	}
	return list, total, rows.Err()
}

// CloseSession moves a live paper to its final state. The status is part of the
// WHERE clause so a double submit updates nothing and reports it, rather than
// grading the same paper twice.
func (r *Repository) CloseSession(ctx context.Context, db database.DB, sessionID, status string, attemptID *string) (bool, error) {
	tag, err := db.Exec(ctx, `
		UPDATE reading_mock_sessions
		SET status          = $2,
		    mock_attempt_id = $3,
		    submitted_at    = CASE WHEN $2 = 'submitted' THEN now() ELSE submitted_at END
		WHERE id = $1 AND status = 'in_progress'`, sessionID, status, attemptID)
	if err != nil {
		return false, fmt.Errorf("close reading mock session: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func scanSession(row pgx.Row) (Session, error) {
	var s Session
	err := row.Scan(&s.ID, &s.MockID, &s.MockTitle, &s.Exam, &s.ExamVersionID, &s.Status,
		&s.DurationMinutes, &s.PassageIDs, &s.QuestionIDs, &s.ReusedPassages,
		&s.CreatedAt, &s.SubmittedAt)
	if err != nil {
		return Session{}, err
	}
	s.TotalQuestions = len(s.QuestionIDs)
	return s, nil
}
