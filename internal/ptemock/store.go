package ptemock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

// Session statuses.
const (
	StatusInProgress = "in_progress"
	StatusScoring    = "scoring"
	StatusCompleted  = "completed"
	StatusAbandoned  = "abandoned"
)

// Item statuses.
const (
	ItemPending  = "pending"
	ItemAnswered = "answered"
	ItemSkipped  = "skipped"
	ItemTimedOut = "timed_out"
)

// sessionRow is a pte_mock_sessions row.
type sessionRow struct {
	ID               string
	UserID           string
	Kind             Kind
	MockID           string
	ExamVersionID    string
	Status           string
	Current          int
	Total            int
	Deadlines        map[string]time.Time
	Missing          []string
	Result           *Result
	AttemptID        *string
	PaperID          *string
	CreatedAt        time.Time
	ScoringStartedAt *time.Time
	CompletedAt      *time.Time
}

// itemRow is a pte_mock_items row.
type itemRow struct {
	Position    int
	QuestionID  string
	Task        string
	Part        Part
	OptionOrder []string
	Status      string
	Response    *models.AnswerSubmission
	Transcript  string
	// TranscriptSource is "browser" or "server" (Whisper).
	TranscriptSource string
	Duration         int
	Delivery         *Delivery
	ShownAt          *time.Time
	AnsweredAt       *time.Time
	Scored           bool
	Score            *float64
	MaxScore         *float64
	UnscoredReason   string
	Feedback         string
	CorrectDisplay   string
	UserDisplay      string
	EvaluationID     string

	// dirty marks a row changed in memory and not yet written.
	dirty bool
}

// Delivery is what the browser measured while a spoken answer was recorded,
// in the shape the practice endpoints take it.
type Delivery struct {
	DurationSeconds     int     `json:"durationSeconds"`
	PauseCount          int     `json:"pauseCount"`
	LongestPauseSeconds float64 `json:"longestPauseSeconds"`
	SpeakingRatio       float64 `json:"speakingRatio"`
}

const sessionColumns = `id::text, user_id::text, kind, mock_id, exam_version_id, status, current_position,
	total_items, deadlines, missing_tasks, result, mock_attempt_id::text, created_at, scoring_started_at, completed_at, paper_id::text`

func scanSession(row pgx.Row) (sessionRow, error) {
	var s sessionRow
	var deadlines map[string]string
	var result []byte
	err := row.Scan(&s.ID, &s.UserID, &s.Kind, &s.MockID, &s.ExamVersionID, &s.Status, &s.Current,
		&s.Total, &deadlines, &s.Missing, &result, &s.AttemptID, &s.CreatedAt, &s.ScoringStartedAt, &s.CompletedAt, &s.PaperID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sessionRow{}, ErrSessionNotFound
	}
	if err != nil {
		return sessionRow{}, fmt.Errorf("read pte mock: %w", err)
	}
	s.Deadlines = make(map[string]time.Time, len(deadlines))
	for key, raw := range deadlines {
		at, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			return sessionRow{}, fmt.Errorf("read pte mock deadline %q: %w", key, err)
		}
		s.Deadlines[key] = at
	}
	if len(result) > 0 {
		var r Result
		if err := json.Unmarshal(result, &r); err != nil {
			return sessionRow{}, fmt.Errorf("read pte mock result: %w", err)
		}
		s.Result = &r
	}
	return s, nil
}

// sessionByID reads one of the learner's papers. forUpdate locks it, which
// every write to a paper does first so that two requests cannot both move it.
func sessionByID(ctx context.Context, db database.DB, userID, id string, forUpdate bool) (sessionRow, error) {
	if !isUUID(id) {
		return sessionRow{}, ErrSessionNotFound
	}
	lock := ""
	if forUpdate {
		lock = " FOR UPDATE"
	}
	return scanSession(db.QueryRow(ctx, `SELECT `+sessionColumns+` FROM pte_mock_sessions
		WHERE id = $1 AND user_id = $2`+lock, id, userID))
}

func liveSession(ctx context.Context, db database.DB, userID string, kind Kind) (sessionRow, error) {
	return scanSession(db.QueryRow(ctx, `SELECT `+sessionColumns+` FROM pte_mock_sessions
		WHERE user_id = $1 AND kind = $2 AND status = 'in_progress'`, userID, kind))
}

const itemColumns = `position, question_id, task, part, option_order, status, response,
	COALESCE(transcript, ''), COALESCE(transcript_source, ''), COALESCE(duration_seconds, 0), delivery, shown_at, answered_at,
	scored, score, max_score, COALESCE(unscored_reason, ''), COALESCE(feedback, ''),
	COALESCE(correct_display, ''), COALESCE(user_display, ''), COALESCE(evaluation_id::text, '')`

func loadItems(ctx context.Context, db database.DB, sessionID string) ([]itemRow, error) {
	rows, err := db.Query(ctx, `SELECT `+itemColumns+` FROM pte_mock_items
		WHERE session_id = $1 ORDER BY position`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list pte mock items: %w", err)
	}
	defer rows.Close()

	var items []itemRow
	for rows.Next() {
		var it itemRow
		var response, delivery []byte
		if err := rows.Scan(&it.Position, &it.QuestionID, &it.Task, &it.Part, &it.OptionOrder, &it.Status,
			&response, &it.Transcript, &it.TranscriptSource, &it.Duration, &delivery, &it.ShownAt, &it.AnsweredAt,
			&it.Scored, &it.Score, &it.MaxScore, &it.UnscoredReason, &it.Feedback,
			&it.CorrectDisplay, &it.UserDisplay, &it.EvaluationID); err != nil {
			return nil, fmt.Errorf("scan pte mock item: %w", err)
		}
		if len(response) > 0 {
			var sub models.AnswerSubmission
			if err := json.Unmarshal(response, &sub); err != nil {
				return nil, fmt.Errorf("read pte mock response: %w", err)
			}
			it.Response = &sub
		}
		if len(delivery) > 0 {
			var d Delivery
			if err := json.Unmarshal(delivery, &d); err != nil {
				return nil, fmt.Errorf("read pte mock delivery: %w", err)
			}
			it.Delivery = &d
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// saveProgress writes a paper's place, clocks and status, and every item
// changed in memory.
func saveProgress(ctx context.Context, db database.DB, s sessionRow, items []itemRow) error {
	deadlines := make(map[string]string, len(s.Deadlines))
	for key, at := range s.Deadlines {
		deadlines[key] = at.UTC().Format(time.RFC3339Nano)
	}
	deadlinesJSON, err := jsonText(deadlines)
	if err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `
		UPDATE pte_mock_sessions
		   SET current_position = $2, deadlines = $3::jsonb, status = $4, scoring_started_at = $5
		 WHERE id = $1`, s.ID, s.Current, deadlinesJSON, s.Status, s.ScoringStartedAt); err != nil {
		return fmt.Errorf("save pte mock progress: %w", err)
	}
	for i := range items {
		it := &items[i]
		if !it.dirty {
			continue
		}
		var response, delivery *string
		if it.Response != nil {
			text, err := jsonText(it.Response)
			if err != nil {
				return err
			}
			response = &text
		}
		if it.Delivery != nil {
			text, err := jsonText(it.Delivery)
			if err != nil {
				return err
			}
			delivery = &text
		}
		if _, err := db.Exec(ctx, `
			UPDATE pte_mock_items
			   SET status = $3, response = $4::jsonb, transcript = NULLIF($5, ''), duration_seconds = NULLIF($6, 0),
			       delivery = $7::jsonb, shown_at = $8, answered_at = $9, transcript_source = NULLIF($10, '')
			 WHERE session_id = $1 AND position = $2`,
			s.ID, it.Position, it.Status, response, it.Transcript, it.Duration, delivery, it.ShownAt, it.AnsweredAt,
			it.TranscriptSource); err != nil {
			return fmt.Errorf("save pte mock item %d: %w", it.Position, err)
		}
		it.dirty = false
	}
	return nil
}

// jsonText encodes a value for a jsonb parameter. The API's pool runs
// statements unprepared (database.Connect), where pgx cannot infer a Go map or
// struct's type, so JSON goes over as text and is cast in the statement.
func jsonText(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode json: %w", err)
	}
	return string(b), nil
}

// saveMark writes one item's mark as soon as it has one, so scoring that is
// interrupted picks up where it stopped rather than rating anything twice.
func saveMark(ctx context.Context, db database.DB, sessionID string, it itemRow) error {
	_, err := db.Exec(ctx, `
		UPDATE pte_mock_items
		   SET scored = $3, score = $4, max_score = $5, unscored_reason = NULLIF($6, ''),
		       feedback = NULLIF($7, ''), correct_display = NULLIF($8, ''), user_display = NULLIF($9, ''),
		       evaluation_id = NULLIF($10, '')::uuid
		 WHERE session_id = $1 AND position = $2`,
		sessionID, it.Position, it.Scored, it.Score, it.MaxScore, it.UnscoredReason,
		it.Feedback, it.CorrectDisplay, it.UserDisplay, it.EvaluationID)
	if err != nil {
		return fmt.Errorf("save pte mock mark %d: %w", it.Position, err)
	}
	return nil
}

// dealt is one item chosen for a new paper.
type dealt struct {
	QuestionID  string
	Task        string
	Part        Part
	OptionOrder []string
}

func insertPaper(ctx context.Context, tx pgx.Tx, userID string, blueprint Blueprint, examVersionID string,
	items []dealt, missing []string, paperID *string) (string, error) {
	if missing == nil {
		missing = []string{} // a nil slice is sent as NULL
	}
	var id string
	err := tx.QueryRow(ctx, `
		INSERT INTO pte_mock_sessions (user_id, kind, mock_id, exam_version_id, total_items, missing_tasks, paper_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text`, userID, blueprint.Kind, blueprint.MockID, examVersionID, len(items), missing, paperID).Scan(&id)
	if err != nil {
		return "", err
	}
	rows := make([][]any, 0, len(items))
	for i, it := range items {
		rows = append(rows, []any{id, i + 1, it.QuestionID, it.Task, string(it.Part), it.OptionOrder})
	}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"pte_mock_items"},
		[]string{"session_id", "position", "question_id", "task", "part", "option_order"},
		pgx.CopyFromRows(rows)); err != nil {
		return "", fmt.Errorf("deal pte mock items: %w", err)
	}
	return id, nil
}

// History is one of the learner's papers, for the list of past tests.
type History struct {
	ID          string     `json:"id"`
	Kind        Kind       `json:"kind"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Headline    *float64   `json:"headline"`
	Result      *Result    `json:"result,omitempty"`
	TotalItems  int        `json:"totalItems"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

func listSessions(ctx context.Context, db database.DB, userID string, limit, offset int) ([]History, int, error) {
	var total int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM pte_mock_sessions WHERE user_id = $1 AND status <> 'abandoned'`,
		userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count pte mocks: %w", err)
	}
	rows, err := db.Query(ctx, `SELECT `+sessionColumns+` FROM pte_mock_sessions
		WHERE user_id = $1 AND status <> 'abandoned'
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list pte mocks: %w", err)
	}
	defer rows.Close()

	list := []History{}
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, 0, err
		}
		h := History{ID: s.ID, Kind: s.Kind, Status: s.Status, TotalItems: s.Total,
			CreatedAt: s.CreatedAt, CompletedAt: s.CompletedAt, Result: s.Result}
		if bp, ok := BlueprintFor(s.Kind); ok {
			h.Title = bp.Title
			if s.Result != nil {
				if headline, ok := s.Result.Headline(bp); ok {
					h.Headline = &headline
				}
			}
		}
		list = append(list, h)
	}
	return list, total, rows.Err()
}

// clockKey names the countdown an item runs on: "" for an item that keeps its
// own speaking windows, "item:N" for an item with a clock of its own, and
// "part:P" for an item on its part's shared clock.
func clockKey(it itemRow) string {
	switch Tasks[it.Task].Clock {
	case ClockItem:
		return "item:" + strconv.Itoa(it.Position)
	case ClockSection:
		return "part:" + string(it.Part)
	default:
		return ""
	}
}

// clockSeconds is a countdown's length: an item clock's own time, or the sum
// of the shares of every item on a part clock.
func clockSeconds(key string, items []itemRow) int {
	seconds := 0
	for _, it := range items {
		if clockKey(it) == key {
			seconds += Tasks[it.Task].Seconds
		}
	}
	return seconds
}

func isUUID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i, r := range id {
		switch {
		case i == 8 || i == 13 || i == 18 || i == 23:
			if r != '-' {
				return false
			}
		case (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F'):
		default:
			return false
		}
	}
	return true
}
