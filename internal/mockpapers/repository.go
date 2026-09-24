package mockpapers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// HashContent generates a deterministic sha256 hex string for the given raw JSON.
func HashContent(content json.RawMessage) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

// Publish locks the scope counter, assigns the permanent number, and inserts a published paper.
func (r *Repository) Publish(ctx context.Context, draft DraftPaper) (MockPaper, error) {
	exam := strings.ToLower(strings.TrimSpace(draft.Exam))
	section := strings.ToLower(strings.TrimSpace(draft.Section))
	module := strings.ToLower(strings.TrimSpace(draft.Module))
	if module == "" {
		module = ModuleAny
	}
	schema := draft.ContentSchema
	if schema <= 0 {
		schema = 1
	}

	contentHash := HashContent(draft.Content)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return MockPaper{}, fmt.Errorf("begin publish: %w", err)
	}
	defer tx.Rollback(ctx)

	// Ensure counter row exists
	_, err = tx.Exec(ctx, `
		INSERT INTO mock_paper_counters (exam, section, module, next_number)
		VALUES ($1, $2, $3, COALESCE((SELECT max(number) FROM mock_papers WHERE exam = $1 AND section = $2 AND module = $3), 0) + 1)
		ON CONFLICT (exam, section, module) DO NOTHING`, exam, section, module)
	if err != nil {
		return MockPaper{}, fmt.Errorf("init counter: %w", err)
	}

	// Lock counter row FOR UPDATE
	var nextNumber int
	err = tx.QueryRow(ctx, `
		SELECT next_number
		  FROM mock_paper_counters
		 WHERE exam = $1 AND section = $2 AND module = $3
		   FOR UPDATE`, exam, section, module).Scan(&nextNumber)
	if err != nil {
		return MockPaper{}, fmt.Errorf("lock counter: %w", err)
	}

	// Advance counter
	_, err = tx.Exec(ctx, `
		UPDATE mock_paper_counters
		   SET next_number = next_number + 1
		 WHERE exam = $1 AND section = $2 AND module = $3`, exam, section, module)
	if err != nil {
		return MockPaper{}, fmt.Errorf("advance counter: %w", err)
	}

	title := draft.Title
	if strings.TrimSpace(title) == "" {
		title = fmt.Sprintf("Test %d", nextNumber)
	}

	var paper MockPaper
	err = tx.QueryRow(ctx, `
		INSERT INTO mock_papers (
			exam, section, module, number, revision, status, title,
			content_schema, content, content_hash, created_at, published_at
		) VALUES (
			$1, $2, $3, $4, 1, 'published', $5,
			$6, $7, $8, now(), now()
		) RETURNING id, exam, section, module, number, revision, supersedes_id, status, title,
		            content_schema, content, content_hash, created_at, published_at, retired_at`,
		exam, section, module, nextNumber, title,
		schema, draft.Content, contentHash).Scan(
		&paper.ID, &paper.Exam, &paper.Section, &paper.Module, &paper.Number, &paper.Revision,
		&paper.SupersedesID, &paper.Status, &paper.Title, &paper.ContentSchema, &paper.Content,
		&paper.ContentHash, &paper.CreatedAt, &paper.PublishedAt, &paper.RetiredAt)
	if err != nil {
		return MockPaper{}, fmt.Errorf("insert published paper: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return MockPaper{}, fmt.Errorf("commit published paper: %w", err)
	}

	return paper, nil
}

// Revise publishes a new revision of an existing published paper and retires the old revision.
func (r *Repository) Revise(ctx context.Context, paperID string, newContent json.RawMessage, newTitle string) (MockPaper, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return MockPaper{}, fmt.Errorf("begin revise: %w", err)
	}
	defer tx.Rollback(ctx)

	var old MockPaper
	err = tx.QueryRow(ctx, `
		SELECT id, exam, section, module, number, revision, status, title, content_schema
		  FROM mock_papers
		 WHERE id = $1
		   FOR UPDATE`, paperID).Scan(
		&old.ID, &old.Exam, &old.Section, &old.Module, &old.Number, &old.Revision,
		&old.Status, &old.Title, &old.ContentSchema)
	if errors.Is(err, pgx.ErrNoRows) {
		return MockPaper{}, ErrPaperNotFound
	}
	if err != nil {
		return MockPaper{}, fmt.Errorf("fetch paper to revise: %w", err)
	}
	if old.Status != StatusPublished {
		return MockPaper{}, fmt.Errorf("cannot revise paper with status %q", old.Status)
	}

	// Retire old revision
	tag, err := tx.Exec(ctx, `
		UPDATE mock_papers
		   SET status = 'retired', retired_at = now()
		 WHERE id = $1 AND status = 'published'`, paperID)
	if err != nil {
		return MockPaper{}, fmt.Errorf("retire old revision: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return MockPaper{}, fmt.Errorf("failed to retire old revision %s", paperID)
	}

	title := newTitle
	if strings.TrimSpace(title) == "" {
		title = old.Title
	}
	contentHash := HashContent(newContent)

	var nextRev MockPaper
	err = tx.QueryRow(ctx, `
		INSERT INTO mock_papers (
			exam, section, module, number, revision, supersedes_id, status, title,
			content_schema, content, content_hash, created_at, published_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, 'published', $7,
			$8, $9, $10, now(), now()
		) RETURNING id, exam, section, module, number, revision, supersedes_id, status, title,
		            content_schema, content, content_hash, created_at, published_at, retired_at`,
		old.Exam, old.Section, old.Module, old.Number, old.Revision+1, old.ID, title,
		old.ContentSchema, newContent, contentHash).Scan(
		&nextRev.ID, &nextRev.Exam, &nextRev.Section, &nextRev.Module, &nextRev.Number, &nextRev.Revision,
		&nextRev.SupersedesID, &nextRev.Status, &nextRev.Title, &nextRev.ContentSchema, &nextRev.Content,
		&nextRev.ContentHash, &nextRev.CreatedAt, &nextRev.PublishedAt, &nextRev.RetiredAt)
	if err != nil {
		return MockPaper{}, fmt.Errorf("insert new revision: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return MockPaper{}, fmt.Errorf("commit revision: %w", err)
	}

	return nextRev, nil
}

// Retire marks a published paper as retired.
func (r *Repository) Retire(ctx context.Context, paperID string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE mock_papers
		   SET status = 'retired', retired_at = now()
		 WHERE id = $1 AND status = 'published'`, paperID)
	if err != nil {
		return fmt.Errorf("retire paper: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrPaperNotFound
	}
	return nil
}

// ByID fetches a mock paper by primary key.
func (r *Repository) ByID(ctx context.Context, id string) (MockPaper, error) {
	var paper MockPaper
	err := r.db.QueryRow(ctx, `
		SELECT id, exam, section, module, number, revision, supersedes_id, status, title,
		       content_schema, content, content_hash, created_at, published_at, retired_at
		  FROM mock_papers
		 WHERE id = $1`, id).Scan(
		&paper.ID, &paper.Exam, &paper.Section, &paper.Module, &paper.Number, &paper.Revision,
		&paper.SupersedesID, &paper.Status, &paper.Title, &paper.ContentSchema, &paper.Content,
		&paper.ContentHash, &paper.CreatedAt, &paper.PublishedAt, &paper.RetiredAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return MockPaper{}, ErrPaperNotFound
	}
	if err != nil {
		return MockPaper{}, fmt.Errorf("by id: %w", err)
	}
	return paper, nil
}

// ListPublished lists all published papers in a scope, ordered by permanent number ASC.
func (r *Repository) ListPublished(ctx context.Context, exam, section, module string) ([]MockPaper, error) {
	exam = strings.ToLower(strings.TrimSpace(exam))
	section = strings.ToLower(strings.TrimSpace(section))
	module = strings.ToLower(strings.TrimSpace(module))
	if module == "" {
		module = ModuleAny
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, exam, section, module, number, revision, supersedes_id, status, title,
		       content_schema, content, content_hash, created_at, published_at, retired_at
		  FROM mock_papers
		 WHERE exam = $1 AND section = $2 AND module = $3 AND status = 'published'
		 ORDER BY number ASC`, exam, section, module)
	if err != nil {
		return nil, fmt.Errorf("list published: %w", err)
	}
	defer rows.Close()

	var papers []MockPaper
	for rows.Next() {
		var p MockPaper
		if err := rows.Scan(
			&p.ID, &p.Exam, &p.Section, &p.Module, &p.Number, &p.Revision,
			&p.SupersedesID, &p.Status, &p.Title, &p.ContentSchema, &p.Content,
			&p.ContentHash, &p.CreatedAt, &p.PublishedAt, &p.RetiredAt); err != nil {
			return nil, err
		}
		papers = append(papers, p)
	}
	return papers, rows.Err()
}

// CountPublished returns the count of currently published papers in a scope.
func (r *Repository) CountPublished(ctx context.Context, exam, section, module string) (int, error) {
	exam = strings.ToLower(strings.TrimSpace(exam))
	section = strings.ToLower(strings.TrimSpace(section))
	module = strings.ToLower(strings.TrimSpace(module))
	if module == "" {
		module = ModuleAny
	}

	var count int
	err := r.db.QueryRow(ctx, `
		SELECT count(*)
		  FROM mock_papers
		 WHERE exam = $1 AND section = $2 AND module = $3 AND status = 'published'`,
		exam, section, module).Scan(&count)
	return count, err
}
