// Package writingmock runs the IELTS Writing mock: Task 1 and Task 2 in one
// timed sitting, rated task by task and combined into a Writing band.
//
// IELTS rates both tasks and states that Task 2 contributes twice as much as
// Task 1 to the Writing score. The band here is the 1:2 weighted mean of the
// two task estimates, rounded with the IELTS half-band rule. IELTS does not
// publish how the weighted mean is rounded; using the overall-band rule is
// Prepyo's convention, and the result is a practice estimate.
package writingmock

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/scoring"
)

// MockID is the mocks row every writing mock attempt is recorded against.
const MockID = "mock-ielts-writing-gen"

// DurationMinutes is the official time for IELTS Writing: 60 minutes for both
// tasks, about 20 for Task 1 and 40 for Task 2.
const DurationMinutes = 60

// SubmitGrace is how long after the deadline a submission is still taken as
// sent; past it the paper is graded from the drafts saved in time.
const SubmitGrace = 90 * time.Second

var (
	ErrNotIELTS         = errors.New("the writing mock is an IELTS paper")
	ErrNoTasks          = errors.New("not enough writing tasks to build a mock")
	ErrSessionNotFound  = errors.New("writing mock not found")
	ErrAlreadySubmitted = errors.New("this writing mock has already been submitted")
	ErrPaperClosed      = errors.New("this writing mock is closed")
)

// Session is a dealt writing paper as the learner sees it.
type Session struct {
	ID               string          `json:"id"`
	Status           string          `json:"status"`
	Module           string          `json:"module"`
	DurationMinutes  int             `json:"durationMinutes"`
	CreatedAt        time.Time       `json:"createdAt"`
	ExpiresAt        time.Time       `json:"expiresAt"`
	SecondsRemaining int             `json:"secondsRemaining"`
	Task1            models.Question `json:"task1"`
	Task2            models.Question `json:"task2"`
	Task1Draft       string          `json:"task1Draft"`
	Task2Draft       string          `json:"task2Draft"`

	task1ID, task2ID string
}

// Result is a graded writing mock.
type Result struct {
	// Attempt is nil when the band could not be computed because a task could
	// not be rated; the task evaluations say why.
	Attempt         *models.MockAttempt `json:"attempt"`
	WritingBand     *float64            `json:"writingBand"`
	Task1Evaluation models.Evaluation   `json:"task1Evaluation"`
	Task2Evaluation models.Evaluation   `json:"task2Evaluation"`
	XPAwarded       int                 `json:"xpAwarded"`
	Streak          int                 `json:"streak"`
}

// TaskEvaluator rates one writing task. evaluations.Service implements it.
type TaskEvaluator interface {
	EvaluateMockTask(ctx context.Context, user models.User, question models.Question, text, sessionID string) (models.Evaluation, error)
}

type Service struct {
	db        *pgxpool.Pool
	questions *questions.Repository
	mocks     *mocks.Repository
	billing   *billing.Service
	xp        *gamification.Service
	evaluator TaskEvaluator
}

func NewService(db *pgxpool.Pool, questionRepo *questions.Repository, mockRepo *mocks.Repository,
	billingService *billing.Service, xp *gamification.Service, evaluator TaskEvaluator) *Service {
	return &Service{db: db, questions: questionRepo, mocks: mockRepo, billing: billingService, xp: xp, evaluator: evaluator}
}

// task1Type is the Writing Task 1 each module sets.
func task1Type(module string) string {
	if module == models.ModuleGeneralTraining {
		return "ielts-writing-task1-letter"
	}
	return "ielts-writing-task1-figure"
}

// Band combines the two task estimates the way the IELTS Writing score does:
// Task 2 counts twice as much as Task 1.
func Band(task1, task2 float64) float64 {
	return scoring.RoundIELTSBand((task1 + 2*task2) / 3)
}

// Start deals a paper, or returns the one this learner already has open.
func (s *Service) Start(ctx context.Context, user models.User) (Session, error) {
	return s.start(ctx, user, false)
}

// StartForFullMock deals a fresh paper for a full mock's Writing section. It
// is not charged (the full mock was), and it is kept apart from any writing
// mock the learner has open.
func (s *Service) StartForFullMock(ctx context.Context, user models.User) (Session, error) {
	return s.start(ctx, user, true)
}

func (s *Service) start(ctx context.Context, user models.User, fullMock bool) (Session, error) {
	charge := !fullMock && s.billing != nil
	if user.TargetExam != models.ExamIELTS {
		return Session{}, ErrNotIELTS
	}
	if !fullMock {
		if live, err := s.live(ctx, s.db, user.ID); err == nil {
			return s.hydrate(ctx, live)
		} else if !errors.Is(err, ErrSessionNotFound) {
			return Session{}, err
		}
	}

	if charge {
		if _, err := s.billing.CheckSubTestCredits(ctx, s.db, user, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckAIGradings(ctx, s.db, user, billing.IELTSWritingMockUnits); err != nil {
			return Session{}, err
		}
	}

	module := user.IELTSModule()
	task1, err := s.pickTask(ctx, user.ID, "type_id = $2", task1Type(module))
	if err != nil {
		return Session{}, err
	}
	task2, err := s.pickTask(ctx, user.ID, "type_id LIKE $2", "ielts-writing-task2%")
	if err != nil {
		return Session{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, fmt.Errorf("begin writing mock: %w", err)
	}
	defer tx.Rollback(ctx)

	if charge {
		if err := billing.LockUserForQuota(ctx, tx, user.ID); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckSubTestCredits(ctx, tx, user, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckAIGradings(ctx, tx, user, billing.IELTSWritingMockUnits); err != nil {
			return Session{}, err
		}
	}

	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO writing_mock_sessions (user_id, exam, module, task1_id, task2_id, duration_minutes, expires_at, in_full_mock)
		VALUES ($1, 'IELTS', $2, $3, $4, $5, now() + make_interval(mins => $5), $6)
		RETURNING id::text`, user.ID, module, task1, task2, DurationMinutes, fullMock).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// A second start raced the first; resume what it dealt.
			tx.Rollback(ctx)
			live, liveErr := s.live(ctx, s.db, user.ID)
			if liveErr != nil {
				return Session{}, liveErr
			}
			return s.hydrate(ctx, live)
		}
		return Session{}, fmt.Errorf("create writing mock: %w", err)
	}

	if charge {
		if _, err := s.billing.RecordSessionStartCredits(ctx, tx, user, string(models.ExamIELTS),
			string(models.SkillWriting), "writing-mock:"+id, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
		// The AI marks this paper, so it takes its cost from the grading pool.
		if err := s.billing.RecordAIGrading(ctx, tx, user.ID, billing.IELTSWritingMockUnits, "mock-ielts-writing", id); err != nil {
			return Session{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("commit writing mock: %w", err)
	}

	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.hydrate(ctx, session)
}

// pickTask prefers a task the learner has neither had rated nor been dealt in
// a mock, then the one they met longest ago.
func (s *Service) pickTask(ctx context.Context, userID, filter, value string) (string, error) {
	var id string
	err := s.db.QueryRow(ctx, `
		SELECT q.id
		  FROM questions q
		 WHERE q.is_published AND q.skill = 'writing' AND 'IELTS' = ANY(q.supported_exams)
		   AND q.`+filter+`
		 ORDER BY EXISTS (SELECT 1 FROM ai_evaluations e WHERE e.user_id = $1 AND e.question_id = q.id),
		          EXISTS (SELECT 1 FROM writing_mock_sessions w
		                   WHERE w.user_id = $1 AND (w.task1_id = q.id OR w.task2_id = q.id)),
		          random()
		 LIMIT 1`, userID, value).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoTasks
	}
	if err != nil {
		return "", fmt.Errorf("pick writing task: %w", err)
	}
	return id, nil
}

// Resume returns a paper by id.
func (s *Service) Resume(ctx context.Context, user models.User, id string) (Session, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.hydrate(ctx, session)
}

// SaveDrafts stores what the learner has written so far. It refuses once time
// and the grace period have run out.
func (s *Service) SaveDrafts(ctx context.Context, user models.User, id, task1, task2 string) (Session, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE writing_mock_sessions
		   SET task1_draft = $3, task2_draft = $4
		 WHERE id = $1 AND user_id = $2 AND status = 'in_progress'
		   AND now() <= expires_at + make_interval(secs => $5)`,
		id, user.ID, task1, task2, SubmitGrace.Seconds())
	if err != nil {
		return Session{}, fmt.Errorf("save writing drafts: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Session{}, ErrPaperClosed
	}
	return s.byID(ctx, s.db, user.ID, id)
}

// Submit rates both tasks and records the Writing band.
func (s *Service) Submit(ctx context.Context, user models.User, id, task1Text, task2Text string) (Result, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Result{}, err
	}
	if session.Status != "in_progress" {
		return Result{}, ErrAlreadySubmitted
	}
	// Time is the server's: after the deadline and its grace period, only
	// what was saved in time is graded.
	if time.Now().After(session.ExpiresAt.Add(SubmitGrace)) {
		task1Text, task2Text = session.Task1Draft, session.Task2Draft
	}

	bank, err := s.questions.ByIDs(ctx, []string{session.task1ID, session.task2ID})
	if err != nil {
		return Result{}, err
	}
	task1, ok1 := bank[session.task1ID]
	task2, ok2 := bank[session.task2ID]
	if !ok1 || !ok2 {
		return Result{}, ErrNoTasks
	}

	// The two ratings are independent; running them together halves the wait.
	var wg sync.WaitGroup
	var eval1, eval2 models.Evaluation
	var err1, err2 error
	wg.Add(2)
	go func() {
		defer wg.Done()
		eval1, err1 = s.evaluator.EvaluateMockTask(ctx, user, task1, task1Text, session.ID)
	}()
	go func() {
		defer wg.Done()
		eval2, err2 = s.evaluator.EvaluateMockTask(ctx, user, task2, task2Text, session.ID)
	}()
	wg.Wait()
	// Nothing is closed when a rating fails: the drafts are safe and the
	// learner can submit again. A task already rated is not rated twice.
	if err1 != nil {
		return Result{}, err1
	}
	if err2 != nil {
		return Result{}, err2
	}

	result := Result{Task1Evaluation: eval1, Task2Evaluation: eval2}
	elapsed := int(time.Since(session.CreatedAt).Seconds())
	if limit := session.DurationMinutes * 60; elapsed > limit {
		elapsed = limit
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("begin writing mock submit: %w", err)
	}
	defer tx.Rollback(ctx)

	var attemptID *string
	if eval1.EstimatedScore != nil && eval2.EstimatedScore != nil {
		band := Band(*eval1.EstimatedScore, *eval2.EstimatedScore)
		result.WritingBand = &band
		attempt, err := s.mocks.SaveAttempt(ctx, tx, mocks.SaveAttemptParams{
			UserID:          user.ID,
			MockID:          MockID,
			ExamVersionID:   task2.ExamVersionID,
			Exam:            models.ExamIELTS,
			UserScore:       band,
			SkillScores:     map[models.SkillType]float64{models.SkillWriting: band},
			TotalCorrect:    0,
			TotalQuestions:  0,
			DurationSeconds: max(elapsed, 1),
			Answers: []models.AnswerSubmission{
				{QuestionID: task1.ID, TextResponse: strings.TrimSpace(task1Text)},
				{QuestionID: task2.ID, TextResponse: strings.TrimSpace(task2Text)},
			},
		})
		if err != nil {
			return Result{}, err
		}
		result.Attempt = &attempt
		attemptID = &attempt.ID

		awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
			UserID:    user.ID,
			Amount:    gamification.XPMockCompleted,
			Reason:    "Completed writing mock",
			SourceKey: "writing-mock:" + session.ID,
		})
		if err != nil {
			return Result{}, err
		}
		result.XPAwarded = awarded
		if result.Streak, err = s.xp.TouchStreak(ctx, tx, user); err != nil {
			return Result{}, err
		}
	}

	tag, err := tx.Exec(ctx, `
		UPDATE writing_mock_sessions
		   SET status = 'submitted', submitted_at = now(), mock_attempt_id = $3,
		       task1_draft = $4, task2_draft = $5,
		       task1_evaluation_id = NULLIF($6, '')::uuid, task2_evaluation_id = NULLIF($7, '')::uuid
		 WHERE id = $1 AND user_id = $2 AND status = 'in_progress'`,
		session.ID, user.ID, attemptID, task1Text, task2Text, eval1.ID, eval2.ID)
	if err != nil {
		return Result{}, fmt.Errorf("close writing mock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Result{}, ErrAlreadySubmitted
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit writing mock submit: %w", err)
	}
	return result, nil
}

const sessionFields = `
	id::text, status, module, duration_minutes, created_at, expires_at,
	GREATEST(0, CEIL(EXTRACT(EPOCH FROM (expires_at - now()))))::int,
	task1_id, task2_id, task1_draft, task2_draft`

func scan(row pgx.Row) (Session, error) {
	var s Session
	err := row.Scan(&s.ID, &s.Status, &s.Module, &s.DurationMinutes, &s.CreatedAt, &s.ExpiresAt,
		&s.SecondsRemaining, &s.task1ID, &s.task2ID, &s.Task1Draft, &s.Task2Draft)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("read writing mock: %w", err)
	}
	if s.Status != "in_progress" {
		s.SecondsRemaining = 0
	}
	return s, nil
}

func (s *Service) live(ctx context.Context, db database.DB, userID string) (Session, error) {
	return scan(db.QueryRow(ctx, `SELECT `+sessionFields+` FROM writing_mock_sessions
		WHERE user_id = $1 AND status = 'in_progress' AND NOT in_full_mock`, userID))
}

func (s *Service) byID(ctx context.Context, db database.DB, userID, id string) (Session, error) {
	return scan(db.QueryRow(ctx, `SELECT `+sessionFields+` FROM writing_mock_sessions
		WHERE id::text = $1 AND user_id = $2`, id, userID))
}

// hydrate attaches the two tasks, with answer material stripped.
func (s *Service) hydrate(ctx context.Context, session Session) (Session, error) {
	bank, err := s.questions.ByIDs(ctx, []string{session.task1ID, session.task2ID})
	if err != nil {
		return Session{}, err
	}
	task1, ok1 := bank[session.task1ID]
	task2, ok2 := bank[session.task2ID]
	if !ok1 || !ok2 {
		return Session{}, ErrNoTasks
	}
	session.Task1 = task1.PublicQuestion()
	session.Task2 = task2.PublicQuestion()
	return session, nil
}
