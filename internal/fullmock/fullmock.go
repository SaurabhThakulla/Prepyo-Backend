// Package fullmock runs the IELTS full mock: the four section mocks as one
// test, in the order of the test, with an overall band.
//
// A full mock is paid for from the plan's full-mock allowance when it starts.
// Its sections are the section mocks themselves, opened without charging
// sub-tests; each records its own result as a section mock does, and the full
// mock records the four bands with the overall once all four are done.
package fullmock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// MockID is the attempt row a finished full mock is recorded against.
const MockID = billing.FullMockID

// examVersionID is the IELTS version the section mocks are written for.
const examVersionID = "ielts-2026-01"

// Order is the order of the test: Listening, Reading and Writing are taken
// together, and Speaking last (in IELTS it may be on another day).
var Order = []models.SkillType{models.SkillListening, models.SkillReading, models.SkillWriting, models.SkillSpeaking}

var (
	ErrNotIELTS           = errors.New("the full mock is an IELTS test")
	ErrSessionNotFound    = errors.New("full mock not found")
	ErrNotInProgress      = errors.New("full mock is finished")
	ErrOutOfOrder         = errors.New("section is not the next one in the test")
	ErrUnknownSection     = errors.New("unknown section")
	ErrSectionsUnfinished = errors.New("a section is not finished")
)

// Section statuses.
const (
	SectionNotStarted = "not_started"
	SectionInProgress = "in_progress"
	// SectionSubmitted has a band.
	SectionSubmitted = "submitted"
	// SectionClosed ended without a band: abandoned, or a writing paper that
	// could not be rated.
	SectionClosed = "closed"
)

// SectionState is where one section of the test stands.
type SectionState struct {
	Skill     models.SkillType `json:"skill"`
	SessionID string           `json:"sessionId,omitempty"`
	Status    string           `json:"status"`
	Band      *float64         `json:"band"`
}

// Session is one full mock.
type Session struct {
	ID          string           `json:"id"`
	Module      string           `json:"module"`
	Status      string           `json:"status"`
	Stage       models.SkillType `json:"stage,omitempty"`
	Sections    []SectionState   `json:"sections"`
	OverallBand *float64         `json:"overallBand"`
	AttemptID   string           `json:"attemptId,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
	CompletedAt *time.Time       `json:"completedAt,omitempty"`

	sessionIDs map[models.SkillType]*string
}

// Starter deals a section's paper for a full mock, uncharged and apart from
// any section mock the learner has open, and reopens it by id. A paper comes
// back as its section mock's own start returns it, for that section's runner.
type Starter interface {
	Deal(ctx context.Context, user models.User) (id string, paper any, err error)
	Resume(ctx context.Context, user models.User, id string) (paper any, err error)
}

// sectionTables is where each section's papers are kept.
var sectionTables = map[models.SkillType]string{
	models.SkillListening: "listening_mock_sessions",
	models.SkillReading:   "reading_mock_sessions",
	models.SkillWriting:   "writing_mock_sessions",
	models.SkillSpeaking:  "speaking_mock_sessions",
}

type Service struct {
	db       *pgxpool.Pool
	mocks    *mocks.Repository
	billing  *billing.Service
	starters map[models.SkillType]Starter
}

func NewService(db *pgxpool.Pool, mockRepo *mocks.Repository, billingService *billing.Service, starters map[models.SkillType]Starter) *Service {
	return &Service{db: db, mocks: mockRepo, billing: billingService, starters: starters}
}

// Start opens a full mock, spending one from the allowance, or returns the
// learner's open one at no cost.
func (s *Service) Start(ctx context.Context, user models.User) (Session, error) {
	if user.TargetExam != models.ExamIELTS {
		return Session{}, ErrNotIELTS
	}
	if live, err := s.live(ctx, s.db, user.ID); err == nil {
		return s.state(ctx, live)
	} else if !errors.Is(err, ErrSessionNotFound) {
		return Session{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, fmt.Errorf("begin full mock: %w", err)
	}
	defer tx.Rollback(ctx)

	if s.billing != nil {
		if err := billing.LockUserForQuota(ctx, tx, user.ID); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckMockAllowance(ctx, tx, user); err != nil {
			return Session{}, err
		}
	}

	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO full_mock_sessions (user_id, module) VALUES ($1, $2)
		RETURNING id::text`, user.ID, user.IELTSModule()).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			tx.Rollback(ctx)
			live, liveErr := s.live(ctx, s.db, user.ID)
			if liveErr != nil {
				return Session{}, liveErr
			}
			return s.state(ctx, live)
		}
		return Session{}, fmt.Errorf("create full mock: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("commit full mock: %w", err)
	}
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.state(ctx, session)
}

// Current is the learner's open full mock, if they have one.
func (s *Service) Current(ctx context.Context, user models.User) (Session, error) {
	live, err := s.live(ctx, s.db, user.ID)
	if err != nil {
		return Session{}, err
	}
	return s.state(ctx, live)
}

// Get is one of the learner's full mocks.
func (s *Service) Get(ctx context.Context, user models.User, id string) (Session, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.state(ctx, session)
}

// StartSection opens the next section: its paper if one is open, otherwise a
// fresh one. Sections are taken in the order of the test.
func (s *Service) StartSection(ctx context.Context, user models.User, id string, skill models.SkillType) (Session, any, error) {
	starter, ok := s.starters[skill]
	if !ok {
		return Session{}, nil, ErrUnknownSection
	}
	session, err := s.Get(ctx, user, id)
	if err != nil {
		return Session{}, nil, err
	}
	if session.Status != "in_progress" {
		return Session{}, nil, ErrNotInProgress
	}
	if session.Stage != skill {
		return Session{}, nil, ErrOutOfOrder
	}
	for _, section := range session.Sections {
		if section.Skill == skill && section.Status == SectionInProgress {
			paper, err := starter.Resume(ctx, user, section.SessionID)
			return session, paper, err
		}
	}

	// Held while the paper is dealt, so a start sent twice deals one paper:
	// the second waits here, then finds the first one's paper.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, nil, fmt.Errorf("begin %s section: %w", skill, err)
	}
	defer tx.Rollback(ctx)
	var current *string
	if err := tx.QueryRow(ctx, `SELECT `+string(skill)+`_session_id::text FROM full_mock_sessions
		WHERE id = $1 AND user_id = $2 AND status = 'in_progress' FOR UPDATE`, session.ID, user.ID).Scan(&current); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, nil, ErrNotInProgress
		}
		return Session{}, nil, fmt.Errorf("lock full mock: %w", err)
	}
	if current != nil {
		tx.Rollback(ctx)
		paper, err := starter.Resume(ctx, user, *current)
		if err != nil {
			return Session{}, nil, err
		}
		session, err = s.Get(ctx, user, id)
		return session, paper, err
	}

	sectionID, paper, err := starter.Deal(ctx, user)
	if err != nil {
		return Session{}, nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE full_mock_sessions SET `+string(skill)+`_session_id = $3
		 WHERE id = $1 AND user_id = $2`, session.ID, user.ID, sectionID); err != nil {
		return Session{}, nil, fmt.Errorf("record %s section: %w", skill, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, nil, fmt.Errorf("commit %s section: %w", skill, err)
	}
	session, err = s.Get(ctx, user, id)
	if err != nil {
		return Session{}, nil, err
	}
	return session, paper, nil
}

// Finish closes a full mock whose four sections are done and records the
// overall band. The overall is the mean of the four section bands rounded to
// the nearest half band; with a section missing its band there is none.
func (s *Service) Finish(ctx context.Context, user models.User, id string) (Session, error) {
	session, err := s.Get(ctx, user, id)
	if err != nil {
		return Session{}, err
	}
	if session.Status == "completed" {
		return session, nil
	}
	if session.Status != "in_progress" {
		return Session{}, ErrNotInProgress
	}
	if session.Stage != "" {
		return Session{}, ErrSectionsUnfinished
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, fmt.Errorf("begin full mock finish: %w", err)
	}
	defer tx.Rollback(ctx)

	var attemptID *string
	if overall := Overall(session.Sections); overall != nil {
		scores := map[models.SkillType]float64{}
		for _, section := range session.Sections {
			scores[section.Skill] = *section.Band
		}
		attempt, err := s.mocks.SaveAttempt(ctx, tx, mocks.SaveAttemptParams{
			UserID:          user.ID,
			MockID:          MockID,
			ExamVersionID:   examVersionID,
			Exam:            models.ExamIELTS,
			UserScore:       *overall,
			SkillScores:     scores,
			DurationSeconds: int(time.Since(session.CreatedAt).Seconds()),
		})
		if err != nil {
			return Session{}, err
		}
		attemptID = &attempt.ID
	}
	tag, err := tx.Exec(ctx, `
		UPDATE full_mock_sessions
		   SET status = 'completed', completed_at = now(), mock_attempt_id = $3
		 WHERE id = $1 AND user_id = $2 AND status = 'in_progress'`, session.ID, user.ID, attemptID)
	if err != nil {
		return Session{}, fmt.Errorf("finish full mock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Finished by a racing request; report what it recorded.
		tx.Rollback(ctx)
		return s.Get(ctx, user, id)
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("commit full mock finish: %w", err)
	}
	return s.Get(ctx, user, id)
}

// Abandon closes an open full mock, and the section paper it has open. The
// allowance it spent is not returned, as a test that was booked and not sat
// is not.
func (s *Service) Abandon(ctx context.Context, user models.User, id string) error {
	if !isUUID(id) {
		return ErrSessionNotFound
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin abandon full mock: %w", err)
	}
	defer tx.Rollback(ctx)
	session, err := scan(tx.QueryRow(ctx, `
		UPDATE full_mock_sessions SET status = 'abandoned', completed_at = now()
		 WHERE id = $1 AND user_id = $2 AND status = 'in_progress'
		RETURNING `+columns, id, user.ID))
	if errors.Is(err, ErrSessionNotFound) {
		return ErrNotInProgress
	}
	if err != nil {
		return err
	}
	for _, skill := range Order {
		if paper := session.sessionIDs[skill]; paper != nil {
			if _, err := tx.Exec(ctx, `UPDATE `+sectionTables[skill]+` SET status = 'abandoned'
				WHERE id = $1 AND status = 'in_progress' AND in_full_mock`, *paper); err != nil {
				return fmt.Errorf("close %s section: %w", skill, err)
			}
		}
	}
	return tx.Commit(ctx)
}

// Overall is the overall band: the mean of the four section bands, rounded
// to the nearest half band (a mean ending in .25 rounds up to the next half
// band, and .75 to the next whole band). Nil unless all four have a band.
func Overall(sections []SectionState) *float64 {
	if len(sections) != len(Order) {
		return nil
	}
	var sum float64
	for _, section := range sections {
		if section.Band == nil {
			return nil
		}
		sum += *section.Band
	}
	overall := scoring.RoundIELTSBand(sum / float64(len(sections)))
	return &overall
}

const columns = `id::text, module, status, listening_session_id::text, reading_session_id::text,
	writing_session_id::text, speaking_session_id::text, mock_attempt_id::text, created_at, completed_at`

func scan(row pgx.Row) (Session, error) {
	var session Session
	var listening, reading, writing, speaking, attempt *string
	err := row.Scan(&session.ID, &session.Module, &session.Status, &listening, &reading,
		&writing, &speaking, &attempt, &session.CreatedAt, &session.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("read full mock: %w", err)
	}
	session.sessionIDs = map[models.SkillType]*string{
		models.SkillListening: listening,
		models.SkillReading:   reading,
		models.SkillWriting:   writing,
		models.SkillSpeaking:  speaking,
	}
	if attempt != nil {
		session.AttemptID = *attempt
	}
	return session, nil
}

func (s *Service) live(ctx context.Context, db database.DB, userID string) (Session, error) {
	return scan(db.QueryRow(ctx, `SELECT `+columns+` FROM full_mock_sessions
		WHERE user_id = $1 AND status = 'in_progress'`, userID))
}

func (s *Service) byID(ctx context.Context, db database.DB, userID, id string) (Session, error) {
	if !isUUID(id) {
		return Session{}, ErrSessionNotFound
	}
	return scan(db.QueryRow(ctx, `SELECT `+columns+` FROM full_mock_sessions
		WHERE id = $1 AND user_id = $2`, id, userID))
}

// state reads where each section stands, the section the learner is on and,
// once all four are done, the overall band.
func (s *Service) state(ctx context.Context, session Session) (Session, error) {
	session.Sections = make([]SectionState, 0, len(Order))
	session.Stage = ""
	for _, skill := range Order {
		section := SectionState{Skill: skill, Status: SectionNotStarted}
		if id := session.sessionIDs[skill]; id != nil {
			section.SessionID = *id
			var status string
			var band *float64
			err := s.db.QueryRow(ctx, `
				SELECT p.status, (a.skill_scores ->> $2)::float8
				  FROM `+sectionTables[skill]+` p
				  LEFT JOIN mock_attempts a ON a.id = p.mock_attempt_id
				 WHERE p.id = $1`, *id, string(skill)).Scan(&status, &band)
			switch {
			case errors.Is(err, pgx.ErrNoRows):
				section.Status = SectionClosed
			case err != nil:
				return Session{}, fmt.Errorf("read %s section: %w", skill, err)
			case status == "in_progress":
				section.Status = SectionInProgress
			case status == "submitted" && band != nil:
				section.Status = SectionSubmitted
				section.Band = band
			default:
				section.Status = SectionClosed
			}
		}
		if session.Stage == "" && (section.Status == SectionNotStarted || section.Status == SectionInProgress) {
			session.Stage = skill
		}
		session.Sections = append(session.Sections, section)
	}
	if session.Stage == "" {
		session.OverallBand = Overall(session.Sections)
	}
	return session, nil
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
