package mockpapers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WritingComposer interface {
	PickTasksForBuilder(ctx context.Context, module string) (string, string, error)
}

type ReadingComposer interface {
	ComposeForBuilder(ctx context.Context, module string) ([]string, []string, []string, error)
}

type PTEItemResult struct {
	QuestionID string
	Task       string
	Part       string
}

type PTEComposer interface {
	DealForBuilder(ctx context.Context, kind string) ([]PTEItemResult, []string, error)
}

type Builder struct {
	pool    *pgxpool.Pool
	repo    *Repository
	writing WritingComposer
	reading ReadingComposer
	pte     PTEComposer
	log     *slog.Logger

	reqCh  chan Scope
	scopes map[string]bool
	mu     sync.Mutex
	stopCh chan struct{}
}

func NewBuilder(pool *pgxpool.Pool, repo *Repository, writing WritingComposer, reading ReadingComposer, pte PTEComposer, log *slog.Logger) *Builder {
	return &Builder{
		pool:    pool,
		repo:    repo,
		writing: writing,
		reading: reading,
		pte:     pte,
		log:     log,
		reqCh:   make(chan Scope, 50),
		scopes:  make(map[string]bool),
		stopCh:  make(chan struct{}),
	}
}

// Start launches the background worker. It returns immediately.
func (b *Builder) Start(ctx context.Context) {
	go b.worker(ctx)
}

// Stop terminates the builder.
func (b *Builder) Stop() {
	close(b.stopCh)
}

// QueueAll queues all standard scopes for background building.
func (b *Builder) QueueAll() {
	allScopes := []Scope{
		{Exam: ExamIELTS, Section: SectionWriting, Module: ModuleAcademic},
		{Exam: ExamIELTS, Section: SectionWriting, Module: ModuleGeneralTraining},
		{Exam: ExamIELTS, Section: SectionReading, Module: ModuleAcademic},
		{Exam: ExamIELTS, Section: SectionReading, Module: ModuleGeneralTraining},
		{Exam: ExamIELTS, Section: SectionListening, Module: ModuleAny},
		{Exam: ExamIELTS, Section: SectionSpeaking, Module: ModuleAny},
		{Exam: ExamPTE, Section: SectionSpeaking, Module: ModuleAny},
		{Exam: ExamPTE, Section: SectionWriting, Module: ModuleAny},
		{Exam: ExamPTE, Section: SectionReading, Module: ModuleAny},
		{Exam: ExamPTE, Section: SectionListening, Module: ModuleAny},
		{Exam: ExamPTE, Section: SectionFull, Module: ModuleAny},
	}
	for _, sc := range allScopes {
		b.Request(sc)
	}
}

// Request deduplicates and enqueues a build request for a scope.
func (b *Builder) Request(scope Scope) {
	scope.Exam = strings.ToLower(strings.TrimSpace(scope.Exam))
	scope.Section = strings.ToLower(strings.TrimSpace(scope.Section))
	scope.Module = strings.ToLower(strings.TrimSpace(scope.Module))
	if scope.Module == "" {
		scope.Module = ModuleAny
	}

	key := fmt.Sprintf("%s:%s:%s", scope.Exam, scope.Section, scope.Module)

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.scopes[key] {
		return
	}
	b.scopes[key] = true

	select {
	case b.reqCh <- scope:
	default:
		// Queue full; let next trigger enqueue
		delete(b.scopes, key)
	}
}

func (b *Builder) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-b.stopCh:
			return
		case scope := <-b.reqCh:
			key := fmt.Sprintf("%s:%s:%s", scope.Exam, scope.Section, scope.Module)
			b.mu.Lock()
			delete(b.scopes, key)
			b.mu.Unlock()

			if err := b.BuildScope(ctx, scope, DefaultTargetPapers); err != nil {
				b.log.Warn("builder error for scope", "scope", key, "error", err)
			}
		}
	}
}

func lockKey(scope Scope) int64 {
	h := fnv.New64a()
	h.Write([]byte(fmt.Sprintf("mockpapers:%s:%s:%s", scope.Exam, scope.Section, scope.Module)))
	return int64(h.Sum64())
}

// BuildScope acquires the advisory lock and builds papers up to targetCount.
func (b *Builder) BuildScope(ctx context.Context, scope Scope, targetCount int) error {
	lKey := lockKey(scope)

	// A session advisory lock belongs to the connection that took it, so the
	// lock and the unlock must run on one connection held for the whole build.
	// Taken and released through the pool, the unlock can land on another
	// connection, fail silently, and leave the scope locked for good.
	conn, err := b.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire lock connection: %w", err)
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(ctx, "SELECT pg_try_advisory_lock($1)", lKey).Scan(&locked); err != nil {
		return fmt.Errorf("try advisory lock: %w", err)
	}
	if !locked {
		// Another worker/instance is building this scope right now
		return nil
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", lKey)
	}()

	// IELTS Listening tests and Speaking sets are written whole, not composed:
	// each published one is a numbered test, however many there are.
	if scope.Exam == ExamIELTS && (scope.Section == SectionListening || scope.Section == SectionSpeaking) {
		return b.registerFixed(ctx, scope)
	}

	existing, err := b.repo.ListPublished(ctx, scope.Exam, scope.Section, scope.Module)
	if err != nil {
		return fmt.Errorf("list existing: %w", err)
	}

	needed := targetCount - len(existing)
	if needed <= 0 {
		return nil
	}

	for i := 0; i < needed; i++ {
		published, err := b.buildOne(ctx, scope, existing)
		if err != nil {
			if errors.Is(err, ErrBankTooSmall) || errors.Is(err, ErrOverlapExceeded) {
				b.log.Info("stopping builder for scope", "scope", scope, "reason", err.Error(), "papersCount", len(existing))
				return nil
			}
			return err
		}
		existing = append(existing, published)
	}

	return nil
}

// registerFixed publishes a numbered test for every published IELTS Listening
// test or Speaking set that has none yet, oldest first, so a test added later
// takes the next number.
func (b *Builder) registerFixed(ctx context.Context, scope Scope) error {
	var query, key string
	switch scope.Section {
	case SectionListening:
		key = "testId"
		query = `SELECT t.id, t.title FROM listening_tests t
			WHERE t.is_published
			  AND NOT EXISTS (SELECT 1 FROM mock_papers mp
			                   WHERE mp.exam = 'ielts' AND mp.section = 'listening' AND mp.content->>'testId' = t.id)
			ORDER BY t.created_at, t.id`
	case SectionSpeaking:
		key = "setId"
		query = `SELECT s.id, s.title FROM speaking_mock_sets s
			WHERE s.is_published
			  AND NOT EXISTS (SELECT 1 FROM mock_papers mp
			                   WHERE mp.exam = 'ielts' AND mp.section = 'speaking' AND mp.content->>'setId' = s.id)
			ORDER BY s.created_at, s.id`
	}

	rows, err := b.pool.Query(ctx, query)
	if err != nil {
		return fmt.Errorf("find unregistered %s tests: %w", scope.Section, err)
	}
	type fixed struct{ id, title string }
	var found []fixed
	for rows.Next() {
		var f fixed
		if err := rows.Scan(&f.id, &f.title); err != nil {
			rows.Close()
			return fmt.Errorf("scan %s test: %w", scope.Section, err)
		}
		found = append(found, f)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, f := range found {
		content, err := json.Marshal(map[string]string{key: f.id})
		if err != nil {
			return err
		}
		if _, err := b.repo.Publish(ctx, DraftPaper{
			Exam: ExamIELTS, Section: scope.Section, Module: ModuleAny,
			Title: f.title, ContentSchema: 1, Content: content,
		}); err != nil {
			return fmt.Errorf("register %s %s: %w", scope.Section, f.id, err)
		}
	}
	if len(found) > 0 {
		b.log.Info("registered numbered tests", "scope", scope, "count", len(found))
	}
	return nil
}

func (b *Builder) buildOne(ctx context.Context, scope Scope, existing []MockPaper) (MockPaper, error) {
	const maxRetries = 5

	for attempt := 0; attempt < maxRetries; attempt++ {
		draft, questionIDs, err := b.composeCandidate(ctx, scope, len(existing)+1)
		if err != nil {
			return MockPaper{}, err
		}

		if ExceedsOverlap(questionIDs, existing) {
			continue
		}

		return b.repo.Publish(ctx, draft)
	}

	return MockPaper{}, ErrOverlapExceeded
}

func (b *Builder) composeCandidate(ctx context.Context, scope Scope, nextNum int) (DraftPaper, []string, error) {
	switch {
	case scope.Exam == ExamIELTS && scope.Section == SectionWriting:
		if b.writing == nil {
			return DraftPaper{}, nil, fmt.Errorf("writing composer not configured")
		}
		t1, t2, err := b.writing.PickTasksForBuilder(ctx, scope.Module)
		if err != nil {
			return DraftPaper{}, nil, err
		}
		contentBytes, _ := json.Marshal(WritingContent{Task1ID: t1, Task2ID: t2})
		title := fmt.Sprintf("IELTS %s Writing Practice Test %d", moduleTitle(scope.Module), nextNum)
		return DraftPaper{
			Exam:          scope.Exam,
			Section:       scope.Section,
			Module:        scope.Module,
			Title:         title,
			ContentSchema: 1,
			Content:       contentBytes,
		}, []string{t1, t2}, nil

	case scope.Exam == ExamIELTS && scope.Section == SectionReading:
		if b.reading == nil {
			return DraftPaper{}, nil, fmt.Errorf("reading composer not configured")
		}
		pIDs, qIDs, rIDs, err := b.reading.ComposeForBuilder(ctx, scope.Module)
		if err != nil {
			return DraftPaper{}, nil, err
		}
		if len(qIDs) == 0 {
			return DraftPaper{}, nil, ErrBankTooSmall
		}
		contentBytes, _ := json.Marshal(ReadingContent{PassageIDs: pIDs, QuestionIDs: qIDs, ReorderIDs: rIDs})
		title := fmt.Sprintf("IELTS %s Reading Practice Test %d", moduleTitle(scope.Module), nextNum)
		return DraftPaper{
			Exam:          scope.Exam,
			Section:       scope.Section,
			Module:        scope.Module,
			Title:         title,
			ContentSchema: 1,
			Content:       contentBytes,
		}, qIDs, nil

	case scope.Exam == ExamPTE:
		if b.pte == nil {
			return DraftPaper{}, nil, fmt.Errorf("pte composer not configured")
		}
		dealtItems, missing, err := b.pte.DealForBuilder(ctx, scope.Section)
		if err != nil {
			return DraftPaper{}, nil, err
		}
		if len(dealtItems) == 0 {
			return DraftPaper{}, nil, ErrBankTooSmall
		}

		// PTE full papers check: covers all three parts (Speaking & Writing, Reading, Listening) in blueprint order.
		if scope.Section == SectionFull {
			hasSpk := false
			hasRead := false
			hasList := false
			for _, item := range dealtItems {
				switch item.Part {
				case "speaking_writing", "speaking", "writing":
					hasSpk = true
				case "reading":
					hasRead = true
				case "listening":
					hasList = true
				}
			}
			if !hasSpk || !hasRead || !hasList {
				return DraftPaper{}, nil, fmt.Errorf("%w: full test requires all three parts", ErrBankTooSmall)
			}
		}

		items := make([]PTEItem, len(dealtItems))
		qIDs := make([]string, len(dealtItems))
		for idx, d := range dealtItems {
			items[idx] = PTEItem{QuestionID: d.QuestionID, Task: d.Task, Part: d.Part}
			qIDs[idx] = d.QuestionID
		}

		contentBytes, _ := json.Marshal(PTEContent{Items: items, Missing: missing})
		title := fmt.Sprintf("PTE %s Practice Test %d", sectionTitle(scope.Section), nextNum)
		return DraftPaper{
			Exam:          scope.Exam,
			Section:       scope.Section,
			Module:        ModuleAny,
			Title:         title,
			ContentSchema: 1,
			Content:       contentBytes,
		}, qIDs, nil

	default:
		return DraftPaper{}, nil, fmt.Errorf("unsupported scope for builder: %s/%s/%s", scope.Exam, scope.Section, scope.Module)
	}
}

// ExceedsOverlap checks if candidateIDs overlap by more than 50% with any existing published paper.
func ExceedsOverlap(candidateIDs []string, existing []MockPaper) bool {
	if len(candidateIDs) == 0 {
		return false
	}
	candSet := make(map[string]bool, len(candidateIDs))
	for _, id := range candidateIDs {
		candSet[id] = true
	}

	for _, paper := range existing {
		oldIDs := ExtractQuestionIDs(paper)
		if len(oldIDs) == 0 {
			continue
		}
		common := 0
		for _, id := range oldIDs {
			if candSet[id] {
				common++
			}
		}
		overlap := float64(common) / float64(len(candidateIDs))
		if overlap > MaxAllowedOverlap {
			return true
		}
	}
	return false
}

// ExtractQuestionIDs parses the question IDs stored in a paper's content.
func ExtractQuestionIDs(paper MockPaper) []string {
	switch {
	case paper.Exam == ExamIELTS && paper.Section == SectionWriting:
		var c WritingContent
		if err := json.Unmarshal(paper.Content, &c); err == nil {
			var ids []string
			if c.Task1ID != "" {
				ids = append(ids, c.Task1ID)
			}
			if c.Task2ID != "" {
				ids = append(ids, c.Task2ID)
			}
			return ids
		}
	case paper.Exam == ExamIELTS && paper.Section == SectionReading:
		var c ReadingContent
		if err := json.Unmarshal(paper.Content, &c); err == nil {
			return c.QuestionIDs
		}
	case paper.Exam == ExamPTE:
		var c PTEContent
		if err := json.Unmarshal(paper.Content, &c); err == nil {
			ids := make([]string, len(c.Items))
			for i, item := range c.Items {
				ids[i] = item.QuestionID
			}
			return ids
		}
	}
	return nil
}

func moduleTitle(mod string) string {
	switch mod {
	case ModuleAcademic:
		return "Academic"
	case ModuleGeneralTraining:
		return "General Training"
	default:
		return ""
	}
}

func sectionTitle(sec string) string {
	switch sec {
	case SectionSpeaking:
		return "Speaking"
	case SectionWriting:
		return "Writing"
	case SectionReading:
		return "Reading"
	case SectionListening:
		return "Listening"
	case SectionFull:
		return "Full Mock"
	default:
		return strings.Title(sec)
	}
}
