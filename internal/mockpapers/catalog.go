package mockpapers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/models"
)

type CatalogService struct {
	db      *pgxpool.Pool
	repo    *Repository
	builder *Builder
}

func NewCatalogService(db *pgxpool.Pool, repo *Repository, builder *Builder) *CatalogService {
	return &CatalogService{db: db, repo: repo, builder: builder}
}

type sessionStats struct {
	AttemptCount        int
	LatestSessionID     *string
	LatestStatus        string
	LatestRevision      int
	LatestCreatedAt     *time.Time
	CompletedSessionID  *string
	CompletedAttemptID  *string
	CompletedScore      *float64
	CompletedAt         *time.Time
	LastAttemptRevision int
}

// Catalog returns the numbered list of published mock tests for a learner,
// with learner status derived from their past sessions.
func (c *CatalogService) Catalog(ctx context.Context, user models.User, exam, section, module string) ([]CatalogPaper, error) {
	exam = strings.ToLower(strings.TrimSpace(exam))
	section = strings.ToLower(strings.TrimSpace(section))
	module = strings.ToLower(strings.TrimSpace(module))

	if module == "" {
		if exam == ExamIELTS && (section == SectionReading || section == SectionWriting) {
			module = strings.ToLower(user.IELTSModule())
		} else {
			module = ModuleAny
		}
	}

	papers, err := c.repo.ListPublished(ctx, exam, section, module)
	if err != nil {
		return nil, fmt.Errorf("list published papers: %w", err)
	}

	// Rule 2: No generation in the catalog GET.
	// If the section has fewer published papers than target, signal the builder non-blocking.
	if len(papers) < DefaultTargetPapers && c.builder != nil {
		c.builder.Request(Scope{Exam: exam, Section: section, Module: module})
	}

	if len(papers) == 0 {
		return []CatalogPaper{}, nil
	}

	statsMap, err := c.querySessionStats(ctx, user.ID, exam, section, module)
	if err != nil {
		return nil, fmt.Errorf("query session stats: %w", err)
	}

	result := make([]CatalogPaper, len(papers))
	for i, paper := range papers {
		item := CatalogPaper{
			ID:           paper.ID,
			Exam:         paper.Exam,
			Section:      paper.Section,
			Module:       paper.Module,
			Number:       paper.Number,
			Revision:     paper.Revision,
			Title:        paper.Title,
			Status:       "not_started",
			AttemptCount: 0,
			IsUpdated:    false,
		}

		if stats, ok := statsMap[paper.Number]; ok {
			item.AttemptCount = stats.AttemptCount

			if stats.LatestStatus == "in_progress" {
				item.Status = "in_progress"
				item.SessionID = stats.LatestSessionID
				item.LastAttemptAt = stats.LatestCreatedAt
			} else if stats.CompletedSessionID != nil {
				item.Status = "completed"
				item.SessionID = stats.CompletedSessionID
				item.LastAttemptSessionID = stats.CompletedSessionID
				item.AttemptID = stats.CompletedAttemptID
				item.Score = stats.CompletedScore
				item.LastAttemptAt = stats.CompletedAt
			} else {
				// Internal status was abandoned or none completed: reads not_started
				item.Status = "not_started"
				item.LastAttemptAt = stats.LatestCreatedAt
			}

			// "Updated" appears when the current published revision number is higher
			// than the revision of the learner's last attempt.
			if stats.LastAttemptRevision > 0 && paper.Revision > stats.LastAttemptRevision {
				item.IsUpdated = true
			}
		}

		result[i] = item
	}

	return result, nil
}

func (c *CatalogService) querySessionStats(ctx context.Context, userID string, exam, section, module string) (map[int]sessionStats, error) {
	out := make(map[int]sessionStats)

	var query string
	var args []any

	if exam == ExamPTE {
		// All PTE sections and full mocks use pte_mock_sessions
		query = `
			SELECT
				p.number,
				COUNT(*) AS attempt_count,
				(ARRAY_AGG(s.id::text ORDER BY s.created_at DESC))[1] AS latest_session_id,
				(ARRAY_AGG(s.status ORDER BY s.created_at DESC))[1] AS latest_status,
				(ARRAY_AGG(p.revision ORDER BY s.created_at DESC))[1] AS latest_revision,
				(ARRAY_AGG(s.created_at ORDER BY s.created_at DESC))[1] AS latest_created_at,
				(ARRAY_AGG(s.id::text ORDER BY s.created_at DESC) FILTER (WHERE s.status IN ('completed', 'scoring')))[1] AS completed_session_id,
				(ARRAY_AGG(s.mock_attempt_id::text ORDER BY s.created_at DESC) FILTER (WHERE s.status IN ('completed', 'scoring')))[1] AS completed_attempt_id,
				(ARRAY_AGG(COALESCE(a.user_score::float8, (s.result->>'overall')::float8) ORDER BY s.created_at DESC) FILTER (WHERE s.status IN ('completed', 'scoring')))[1] AS completed_score,
				(ARRAY_AGG(COALESCE(s.completed_at, s.created_at) ORDER BY s.created_at DESC) FILTER (WHERE s.status IN ('completed', 'scoring')))[1] AS completed_at
			FROM pte_mock_sessions s
			JOIN mock_papers p ON p.id = s.paper_id
			LEFT JOIN mock_attempts a ON a.id = s.mock_attempt_id
			WHERE s.user_id = $1 AND p.exam = $2 AND p.section = $3
			GROUP BY p.number`
		args = []any{userID, exam, section}
	} else {
		// IELTS sections
		var sessionTable string
		var submittedField string = "submitted_at"
		var completedStatus string = "submitted"

		switch section {
		case SectionListening:
			sessionTable = "listening_mock_sessions"
		case SectionSpeaking:
			sessionTable = "speaking_mock_sessions"
		case SectionWriting:
			sessionTable = "writing_mock_sessions"
		case SectionReading:
			sessionTable = "reading_mock_sessions"
		default:
			return out, nil
		}

		query = fmt.Sprintf(`
			SELECT
				p.number,
				COUNT(*) AS attempt_count,
				(ARRAY_AGG(s.id::text ORDER BY s.created_at DESC))[1] AS latest_session_id,
				(ARRAY_AGG(s.status ORDER BY s.created_at DESC))[1] AS latest_status,
				(ARRAY_AGG(p.revision ORDER BY s.created_at DESC))[1] AS latest_revision,
				(ARRAY_AGG(s.created_at ORDER BY s.created_at DESC))[1] AS latest_created_at,
				(ARRAY_AGG(s.id::text ORDER BY s.created_at DESC) FILTER (WHERE s.status = '%s'))[1] AS completed_session_id,
				(ARRAY_AGG(s.mock_attempt_id::text ORDER BY s.created_at DESC) FILTER (WHERE s.status = '%s'))[1] AS completed_attempt_id,
				(ARRAY_AGG(a.user_score::float8 ORDER BY s.created_at DESC) FILTER (WHERE s.status = '%s'))[1] AS completed_score,
				(ARRAY_AGG(s.%s ORDER BY s.created_at DESC) FILTER (WHERE s.status = '%s'))[1] AS completed_at
			FROM %s s
			JOIN mock_papers p ON p.id = s.paper_id
			LEFT JOIN mock_attempts a ON a.id = s.mock_attempt_id
			WHERE s.user_id = $1 AND p.exam = $2 AND p.section = $3 AND p.module = $4
			GROUP BY p.number`,
			completedStatus, completedStatus, completedStatus, submittedField, completedStatus, sessionTable)
		args = []any{userID, exam, section, module}
	}

	rows, err := c.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("aggregate sessions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var number int
		var stats sessionStats
		var score *float64
		if err := rows.Scan(
			&number,
			&stats.AttemptCount,
			&stats.LatestSessionID,
			&stats.LatestStatus,
			&stats.LatestRevision,
			&stats.LatestCreatedAt,
			&stats.CompletedSessionID,
			&stats.CompletedAttemptID,
			&score,
			&stats.CompletedAt,
		); err != nil {
			return nil, err
		}
		stats.CompletedScore = score
		stats.LastAttemptRevision = stats.LatestRevision
		out[number] = stats
	}

	return out, rows.Err()
}
