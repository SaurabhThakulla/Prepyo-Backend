// Package reading serves passage-driven reading practice and mock sessions.
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

	// ErrNoReorderItem means the bank holds no item this learner can be dealt.
	ErrNoReorderItem = errors.New("no re-order item available")
)

// Exposure contexts.
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
	id, exam_version_id, title, subtitle, paragraphs, sources,
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

// ListPassages returns passages matching the filter parameters.
func (r *Repository) ListPassages(ctx context.Context, p ListPassagesParams) ([]models.ReadingPassage, int, error) {
	const where = `
		WHERE is_published
		  AND ($1 = '' OR EXISTS (
			  SELECT 1 FROM questions q
			  WHERE q.passage_id = reading_passages.id AND q.is_published
			    AND $1 = ANY(q.supported_exams)))
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
	err := row.Scan(&p.ID, &p.ExamVersionID, &p.Title, &p.Subtitle,
		&p.Paragraphs, &p.Sources, &p.WordCount, &p.Difficulty, &p.Topic, &p.Tags)
	return p, err
}

// ---------------------------------------------------------------------------
// Groups
// ---------------------------------------------------------------------------

const groupFields = `
	id, passage_id, position, type_id, type_name, instructions, resources,
	passage_display, shuffle_questions, time_limit_seconds`

// Group represents a stored reading question group.
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
			&g.Instructions, &g.Resources, &g.PassageDisplay,
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
			&g.Instructions, &g.Resources, &g.PassageDisplay,
			&g.ShuffleQuestions, &g.TimeLimitSeconds)
	if errors.Is(err, pgx.ErrNoRows) {
		return Group{}, ErrNoPassage
	}
	if err != nil {
		return Group{}, fmt.Errorf("get group: %w", err)
	}
	return g, nil
}

// TaskTypes returns reading task types available for an exam with their passage and question counts.
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
			                AND ($1 = '' OR $1 = ANY(q.supported_exams))
			GROUP BY g.type_id

			UNION ALL

			-- Re-order Paragraphs has no passages; each item counts as one.
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

// PickPracticeGroup chooses one group of the requested type for this learner,
// prioritizing unpractised passages then least-recently seen passages.
func (r *Repository) PickPracticeGroup(ctx context.Context, userID string, exam models.ExamType, typeID string) (Group, error) {
	var id string
	err := r.db.QueryRow(ctx, `
		SELECT g.id
		FROM reading_question_groups g
		JOIN reading_passages p ON p.id = g.passage_id AND p.is_published
		LEFT JOIN user_passage_exposures e
		       ON e.user_id = $1 AND e.passage_id = p.id AND e.context = 'practice'
		      AND ($3 = '' OR e.exam = $3)
		WHERE g.type_id = $2
		  AND EXISTS (SELECT 1 FROM questions q
		               WHERE q.group_id = g.id AND q.is_published
		                 AND ($3 = '' OR $3 = ANY(q.supported_exams)))
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

// PickReorderItem chooses one item for this learner, with its backing question.
// Items from previously seen passages are skipped; unseen items are prioritized.
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

// RecordReorderExposure marks items as dealt to this learner.
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

// RecordExposure marks passages as seen by this learner.
func (r *Repository) RecordExposure(ctx context.Context, db database.DB, userID string, exam models.ExamType, passageIDs []string, exposureContext string) error {
	if len(passageIDs) == 0 {
		return nil
	}

	_, err := db.Exec(ctx, `
		INSERT INTO user_passage_exposures (user_id, passage_id, exam, context)
		SELECT $1, unnest($2::text[]), $3, $4
		ON CONFLICT (user_id, passage_id, exam, context) DO UPDATE
		SET seen_count   = user_passage_exposures.seen_count + 1,
		    last_seen_at = now()`, userID, passageIDs, exam, exposureContext)
	if err != nil {
		return fmt.Errorf("record passage exposure: %w", err)
	}
	return nil
}

// SeenPassageIDs returns the passages this learner has met in a context.
func (r *Repository) SeenPassageIDs(ctx context.Context, userID string, exam models.ExamType, exposureContext string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT passage_id FROM user_passage_exposures
		WHERE user_id = $1 AND exam = $2 AND context = $3
		ORDER BY last_seen_at DESC`, userID, exam, exposureContext)
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

// Blueprint represents the structure of a generated reading mock exam.
type Blueprint struct {
	ID              string
	MockID          string
	ExamVersionID   string
	Title           string
	DurationMinutes int
	PassageCount    int
	TotalQuestions  int
	Slots           []BlueprintSlot
}

// BlueprintSlot is one section of a paper, filled from one passage.
type BlueprintSlot struct {
	Position int
	Source   string
	Tasks    []BlueprintTask
}

// BlueprintTask is one task set within a section.
type BlueprintTask struct {
	TypeID        string
	QuestionCount int
}

// TypeIDs is the task types this section draws, in order.
func (s BlueprintSlot) TypeIDs() []string {
	ids := make([]string, len(s.Tasks))
	for i, t := range s.Tasks {
		ids[i] = t.TypeID
	}
	return ids
}

// Counts is the number of questions wanted per task, aligned with TypeIDs.
func (s BlueprintSlot) Counts() []int {
	counts := make([]int, len(s.Tasks))
	for i, t := range s.Tasks {
		counts[i] = t.QuestionCount
	}
	return counts
}

// QuestionCount is the whole section.
func (s BlueprintSlot) QuestionCount() int {
	total := 0
	for _, t := range s.Tasks {
		total += t.QuestionCount
	}
	return total
}

// SourcePassage and SourceReorder are where a slot's content comes from.
const (
	SourcePassage = "passage"
	SourceReorder = "reorder"
)

// ReorderPick is one Re-order Paragraphs item chosen for a paper, with the
// question that carries its boxes and its answer key.
type ReorderPick struct {
	ItemID     string
	QuestionID string
}

func (r *Repository) GeneratedBlueprint(ctx context.Context, exam models.ExamType) (Blueprint, error) {
	var b Blueprint
	err := r.db.QueryRow(ctx, `
		SELECT b.id, m.id, m.exam_version_id, m.title, b.duration_minutes,
		       b.passage_count, b.total_questions
		FROM reading_mock_blueprints b
		JOIN mocks m ON m.id = b.mock_id
		WHERE b.exam = $1 AND b.is_active`, exam).
		Scan(&b.ID, &b.MockID, &b.ExamVersionID, &b.Title, &b.DurationMinutes,
			&b.PassageCount, &b.TotalQuestions)
	if errors.Is(err, pgx.ErrNoRows) {
		return Blueprint{}, ErrNoBlueprint
	}
	if err != nil {
		return Blueprint{}, fmt.Errorf("get generated blueprint: %w", err)
	}

	slots, err := r.blueprintSlots(ctx, b.ID)
	if err != nil {
		return Blueprint{}, err
	}
	b.Slots = slots
	return b, nil
}

func (r *Repository) blueprintSlots(ctx context.Context, blueprintID string) ([]BlueprintSlot, error) {
	rows, err := r.db.Query(ctx, `
		SELECT position, type_id, question_count, source
		FROM reading_mock_blueprint_slots
		WHERE blueprint_id = $1
		ORDER BY position, ordinal`, blueprintID)
	if err != nil {
		return nil, fmt.Errorf("list blueprint slots: %w", err)
	}
	defer rows.Close()

	var list []BlueprintSlot
	for rows.Next() {
		var position, count int
		var typeID, source string
		if err := rows.Scan(&position, &typeID, &count, &source); err != nil {
			return nil, fmt.Errorf("scan blueprint slot: %w", err)
		}

		// Rows arrive grouped by position, so a new section is a change of
		// position rather than a lookup.
		if len(list) == 0 || list[len(list)-1].Position != position {
			list = append(list, BlueprintSlot{Position: position, Source: source})
		}
		last := &list[len(list)-1]
		last.Tasks = append(last.Tasks, BlueprintTask{TypeID: typeID, QuestionCount: count})
	}
	return list, rows.Err()
}

// SlotPassageCandidates returns all passages eligible to fill a section for this
// learner, ordered by exposure priority.
func (r *Repository) SlotPassageCandidates(
	ctx context.Context,
	userID string,
	exam models.ExamType,
	typeIDs []string,
	counts []int,
) ([]MockCandidate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT p.id, (seen.passage_id IS NOT NULL)
		FROM reading_passages p
		LEFT JOIN user_passage_exposures seen
		       ON seen.user_id = $1 AND seen.passage_id = p.id
		      AND seen.exam = $2 AND seen.context = 'mock'
		LEFT JOIN user_passage_exposures practised
		       ON practised.user_id = $1 AND practised.passage_id = p.id
		      AND practised.exam = $2 AND practised.context = 'practice'
		WHERE p.is_published
		  AND NOT EXISTS (
			  SELECT 1
			  FROM unnest($3::text[], $4::int[]) AS want(type_id, n)
			  WHERE (
				  SELECT count(*)
				  FROM questions q
				  JOIN reading_question_groups g ON g.id = q.group_id
				  WHERE g.passage_id = p.id
				    AND g.type_id = want.type_id
				    AND q.is_published
				    AND $2 = ANY(q.supported_exams)
			  ) < want.n
		  )
		ORDER BY (seen.passage_id IS NOT NULL),
		         (practised.passage_id IS NOT NULL),
		         seen.last_seen_at ASC NULLS FIRST,
		         random()`, userID, exam, typeIDs, counts)
	if err != nil {
		return nil, fmt.Errorf("list slot passages: %w", err)
	}
	defer rows.Close()

	var list []MockCandidate
	for rows.Next() {
		var c MockCandidate
		if err := rows.Scan(&c.PassageID, &c.SeenInMock); err != nil {
			return nil, fmt.Errorf("scan slot passage: %w", err)
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// PickReorderItems chooses items for a Re-order Paragraphs slot, with backing questions.
func (r *Repository) PickReorderItems(
	ctx context.Context,
	userID string,
	exam models.ExamType,
	count int,
) ([]ReorderPick, error) {
	rows, err := r.db.Query(ctx, `
		SELECT i.id, q.id
		FROM reading_reorder_items i
		JOIN questions q ON q.reorder_item_id = i.id AND q.is_published
		                AND $2 = ANY(q.supported_exams)
		LEFT JOIN user_reorder_exposures e
		       ON e.user_id = $1 AND e.item_id = i.id AND e.context = 'mock'
		WHERE i.is_published
		  AND (
			  i.source_passage_id IS NULL
			  OR NOT EXISTS (
				  SELECT 1 FROM user_passage_exposures seen
				  WHERE seen.user_id = $1 AND seen.passage_id = i.source_passage_id
			  )
		  )
		ORDER BY e.last_seen_at ASC NULLS FIRST, random()
		LIMIT $3`, userID, exam, count)
	if err != nil {
		return nil, fmt.Errorf("pick re-order items: %w", err)
	}
	defer rows.Close()

	var picks []ReorderPick
	for rows.Next() {
		var p ReorderPick
		if err := rows.Scan(&p.ItemID, &p.QuestionID); err != nil {
			return nil, fmt.Errorf("scan re-order pick: %w", err)
		}
		picks = append(picks, p)
	}
	return picks, rows.Err()
}

// GroupsByIDs loads groups by ID.
func (r *Repository) GroupsByIDs(ctx context.Context, groupIDs []string) (map[string]Group, error) {
	byID := map[string]Group{}
	if len(groupIDs) == 0 {
		return byID, nil
	}

	rows, err := r.db.Query(ctx, `SELECT `+groupFields+`
		FROM reading_question_groups
		WHERE id = ANY($1)`, groupIDs)
	if err != nil {
		return nil, fmt.Errorf("list groups by id: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.PassageID, &g.Position, &g.TypeID, &g.TypeName,
			&g.Instructions, &g.Resources, &g.PassageDisplay,
			&g.ShuffleQuestions, &g.TimeLimitSeconds); err != nil {
			return nil, fmt.Errorf("scan group: %w", err)
		}
		byID[g.ID] = g
	}
	return byID, rows.Err()
}

// Session represents a stored reading mock paper session.
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

// CreateSession stores a dealt reading mock session.
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

// SessionByID loads a reading mock session for a specific user.
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

// ListSessions returns a user's reading mock sessions ordered newest first.
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

// CloseSession transitions an in-progress session to the specified status.
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
