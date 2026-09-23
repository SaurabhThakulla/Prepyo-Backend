// Package listeningmock runs the IELTS Listening mock: four parts and forty
// questions, each recording heard once, marked one mark per answer and banded
// on the indicative Listening table.
//
// The recordings are scripts, spoken by the browser's voice where it has one
// and by the server (speech package) where it does not. A script is the
// answer key, so it is never part of the paper; it is served part by part at
// play time. The server keeps the time and the draft answers; a submission
// after the deadline is graded from what was saved in time.
package listeningmock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// MockID is the mocks row listening attempts are recorded against.
const MockID = "mock-ielts-listening-gen"

// DurationMinutes covers the four recordings, the reading time before each
// and two minutes to check answers at the end, with room for a slow voice.
const DurationMinutes = 45

// SubmitGrace is how long after the deadline a submission is still taken as sent.
const SubmitGrace = 90 * time.Second

var (
	ErrNotIELTS         = errors.New("the listening mock is an IELTS paper")
	ErrNoTest           = errors.New("no listening test is available")
	ErrSessionNotFound  = errors.New("listening mock not found")
	ErrAlreadySubmitted = errors.New("this listening mock has already been submitted")
	ErrPaperClosed      = errors.New("this listening mock is closed")
	ErrNotAttempted     = errors.New("no question on this paper has been answered")
	ErrNoPart           = errors.New("that part is not on this paper")
)

// Group is a set of questions with shared instructions: a form, a map, a list.
type Group struct {
	ID           string            `json:"id"`
	TypeID       string            `json:"typeId"`
	Instructions string            `json:"instructions"`
	Heading      string            `json:"heading,omitempty"`
	ImageURL     string            `json:"imageUrl,omitempty"`
	FirstNumber  int               `json:"firstNumber"`
	Questions    []models.Question `json:"questions"`
}

// Part is one recording and its questions.
type Part struct {
	PartNo         int     `json:"partNo"`
	Setting        string  `json:"setting"`
	ReadingSeconds int     `json:"readingSeconds"`
	FirstNumber    int     `json:"firstNumber"`
	LastNumber     int     `json:"lastNumber"`
	Groups         []Group `json:"groups"`
}

// Session is a dealt listening paper.
type Session struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	TestTitle        string    `json:"testTitle"`
	DurationMinutes  int       `json:"durationMinutes"`
	CreatedAt        time.Time `json:"createdAt"`
	ExpiresAt        time.Time `json:"expiresAt"`
	SecondsRemaining int       `json:"secondsRemaining"`
	ReusedTest       bool      `json:"reusedTest"`
	// PartsPlayed is how many recordings have started; resuming continues
	// with the next one.
	PartsPlayed  int                       `json:"partsPlayed"`
	Parts        []Part                    `json:"parts,omitempty"`
	DraftAnswers []models.AnswerSubmission `json:"draftAnswers"`

	testID      string
	questionIDs []string
}

// Result is a graded listening paper.
type Result struct {
	Attempt   models.MockAttempt      `json:"attempt"`
	XPAwarded int                     `json:"xpAwarded"`
	Review    []models.ReviewQuestion `json:"review"`
	// Scripts are the four transcripts, for review after marking.
	Scripts []string `json:"scripts"`
}

type Service struct {
	db        *pgxpool.Pool
	questions *questions.Repository
	mocks     *mocks.Repository
	billing   *billing.Service
	xp        *gamification.Service
}

func NewService(db *pgxpool.Pool, questionRepo *questions.Repository, mockRepo *mocks.Repository,
	billingService *billing.Service, xp *gamification.Service) *Service {
	return &Service{db: db, questions: questionRepo, mocks: mockRepo, billing: billingService, xp: xp}
}

// StartOptions changes how a paper is started.
type StartOptions struct {
	// Charge spends section-mock credits.
	Charge bool
	// FullMock deals a fresh paper for a full mock's Listening section. It is
	// not charged (the full mock was), and it is kept apart from any section
	// mock the learner has open.
	FullMock bool
}

// Start deals a paper, or returns the learner's open one.
func (s *Service) Start(ctx context.Context, user models.User, opts StartOptions) (Session, error) {
	if user.TargetExam != models.ExamIELTS {
		return Session{}, ErrNotIELTS
	}
	if !opts.FullMock {
		if live, err := s.live(ctx, s.db, user.ID); err == nil {
			return s.hydrate(ctx, live)
		} else if !errors.Is(err, ErrSessionNotFound) {
			return Session{}, err
		}
	}

	charge := opts.Charge && !opts.FullMock && s.billing != nil
	if charge {
		if _, err := s.billing.CheckSubTestCredits(ctx, s.db, user, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
	}

	// A test the learner has not sat comes first.
	var testID string
	var reused bool
	err := s.db.QueryRow(ctx, `
		SELECT t.id, EXISTS (SELECT 1 FROM listening_mock_sessions m WHERE m.user_id = $1 AND m.test_id = t.id)
		  FROM listening_tests t
		 WHERE t.is_published
		   AND (SELECT count(*) FROM questions q
		          JOIN listening_question_groups g ON g.id = q.listening_group_id
		          JOIN listening_parts p ON p.id = g.part_id
		         WHERE p.test_id = t.id AND q.is_published) = 40
		 ORDER BY 2, random()
		 LIMIT 1`, user.ID).Scan(&testID, &reused)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNoTest
	}
	if err != nil {
		return Session{}, fmt.Errorf("pick listening test: %w", err)
	}
	ids, err := s.questionIDs(ctx, testID)
	if err != nil {
		return Session{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, fmt.Errorf("begin listening mock: %w", err)
	}
	defer tx.Rollback(ctx)
	if charge {
		if err := billing.LockUserForQuota(ctx, tx, user.ID); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckSubTestCredits(ctx, tx, user, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
	}
	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO listening_mock_sessions (user_id, test_id, question_ids, duration_minutes, expires_at, reused_test, in_full_mock)
		VALUES ($1, $2, $3, $4, now() + make_interval(mins => $4), $5, $6)
		RETURNING id::text`, user.ID, testID, ids, DurationMinutes, reused, opts.FullMock).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			tx.Rollback(ctx)
			live, liveErr := s.live(ctx, s.db, user.ID)
			if liveErr != nil {
				return Session{}, liveErr
			}
			return s.hydrate(ctx, live)
		}
		return Session{}, fmt.Errorf("create listening mock: %w", err)
	}
	if charge {
		if _, err := s.billing.RecordSessionStartCredits(ctx, tx, user, string(models.ExamIELTS),
			string(models.SkillListening), "listening-mock:"+id, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, fmt.Errorf("commit listening mock: %w", err)
	}
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.hydrate(ctx, session)
}

// questionIDs lists a test's questions in paper order: part, group, position.
func (s *Service) questionIDs(ctx context.Context, testID string) ([]string, error) {
	rows, err := s.db.Query(ctx, `
		SELECT q.id
		  FROM questions q
		  JOIN listening_question_groups g ON g.id = q.listening_group_id
		  JOIN listening_parts p ON p.id = g.part_id
		 WHERE p.test_id = $1 AND q.is_published
		 ORDER BY p.part_no, g.position, q.group_position`, testID)
	if err != nil {
		return nil, fmt.Errorf("list listening questions: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Resume returns a paper by id.
func (s *Service) Resume(ctx context.Context, user models.User, id string) (Session, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.hydrate(ctx, session)
}

// Script is one part's recording script, for the browser to speak. It is
// served only while the paper is open.
func (s *Service) Script(ctx context.Context, user models.User, id string, partNo int) (string, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return "", err
	}
	if session.Status != "in_progress" {
		return "", ErrPaperClosed
	}
	var script string
	err = s.db.QueryRow(ctx, `SELECT script FROM listening_parts WHERE test_id = $1 AND part_no = $2`,
		session.testID, partNo).Scan(&script)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoPart
	}
	return script, err
}

// SaveDrafts stores the answers so far, while time remains.
func (s *Service) SaveDrafts(ctx context.Context, user models.User, id string, answers []models.AnswerSubmission, partsPlayed int) (int, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return 0, err
	}
	kept := onPaper(session.questionIDs, answers)
	body, err := json.Marshal(kept)
	if err != nil {
		return 0, err
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE listening_mock_sessions
		   SET draft_answers = $3,
		       parts_played = GREATEST(parts_played, LEAST(GREATEST($5, 0), 4))
		 WHERE id::text = $1 AND user_id = $2 AND status = 'in_progress'
		   AND now() <= expires_at + make_interval(secs => $4)`,
		id, user.ID, string(body), SubmitGrace.Seconds(), partsPlayed)
	if err != nil {
		return 0, fmt.Errorf("save listening drafts: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return 0, ErrPaperClosed
	}
	fresh, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return 0, err
	}
	return fresh.SecondsRemaining, nil
}

func onPaper(ids []string, answers []models.AnswerSubmission) []models.AnswerSubmission {
	allowed := make(map[string]bool, len(ids))
	for _, id := range ids {
		allowed[id] = true
	}
	seen := map[string]bool{}
	kept := make([]models.AnswerSubmission, 0, len(answers))
	for _, a := range answers {
		if allowed[a.QuestionID] && !seen[a.QuestionID] {
			seen[a.QuestionID] = true
			kept = append(kept, a)
		}
	}
	return kept
}

// Submit marks the paper: one mark per answer, banded on the Listening table.
func (s *Service) Submit(ctx context.Context, user models.User, id string, answers []models.AnswerSubmission) (Result, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Result{}, err
	}
	if session.Status != "in_progress" {
		return Result{}, ErrAlreadySubmitted
	}
	late := time.Now().After(session.ExpiresAt.Add(SubmitGrace))
	if late {
		answers = session.DraftAnswers
	}
	answers = onPaper(session.questionIDs, answers)

	bank, err := s.questions.ByIDs(ctx, session.questionIDs)
	if err != nil {
		return Result{}, err
	}
	graded, err := mocks.GradeAnswersFor(models.ExamIELTS, bank, session.questionIDs, answers)
	if err != nil {
		return Result{}, err
	}
	if graded.Answered == 0 {
		if late || time.Now().After(session.ExpiresAt) {
			_, _ = s.db.Exec(ctx, `UPDATE listening_mock_sessions SET status = 'abandoned' WHERE id::text = $1 AND status = 'in_progress'`, id)
		}
		return Result{}, ErrNotAttempted
	}

	band := scoring.IELTSListeningBand(scaleTo40(graded.Raw, graded.RawMax))
	elapsed := int(time.Since(session.CreatedAt).Seconds())
	if limit := session.DurationMinutes * 60; elapsed > limit {
		elapsed = limit
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("begin listening submit: %w", err)
	}
	defer tx.Rollback(ctx)
	attempt, err := s.mocks.SaveAttempt(ctx, tx, mocks.SaveAttemptParams{
		UserID:          user.ID,
		MockID:          MockID,
		ExamVersionID:   "ielts-2026-01",
		Exam:            models.ExamIELTS,
		UserScore:       band,
		SkillScores:     map[models.SkillType]float64{models.SkillListening: band},
		TotalCorrect:    graded.Raw,
		TotalQuestions:  graded.RawMax,
		DurationSeconds: max(elapsed, 1),
		Answers:         answers,
	})
	if err != nil {
		return Result{}, err
	}
	awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
		UserID: user.ID, Amount: gamification.XPMockCompleted,
		Reason: "Completed listening mock", SourceKey: "listening-mock:" + session.ID,
	})
	if err != nil {
		return Result{}, err
	}
	if _, err := s.xp.TouchStreak(ctx, tx, user); err != nil {
		return Result{}, err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE listening_mock_sessions SET status = 'submitted', submitted_at = now(), mock_attempt_id = $3,
		       draft_answers = $4
		 WHERE id::text = $1 AND user_id = $2 AND status = 'in_progress'`,
		id, user.ID, attempt.ID, mustJSON(answers))
	if err != nil {
		return Result{}, fmt.Errorf("close listening mock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Result{}, ErrAlreadySubmitted
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit listening submit: %w", err)
	}

	review := make([]models.ReviewQuestion, 0, len(session.questionIDs))
	for _, qid := range session.questionIDs {
		if q, ok := bank[qid]; ok {
			review = append(review, q.ForReview())
		}
	}
	scripts, err := s.scripts(ctx, session.testID)
	if err != nil {
		return Result{}, err
	}
	return Result{Attempt: attempt, XPAwarded: awarded, Review: review, Scripts: scripts}, nil
}

func scaleTo40(raw, max int) int {
	if max <= 0 {
		return 0
	}
	if max == 40 {
		return raw
	}
	return int(float64(raw)*40/float64(max) + 0.5)
}

func mustJSON(v any) string {
	body, _ := json.Marshal(v)
	return string(body)
}

func (s *Service) scripts(ctx context.Context, testID string) ([]string, error) {
	rows, err := s.db.Query(ctx, `SELECT script FROM listening_parts WHERE test_id = $1 ORDER BY part_no`, testID)
	if err != nil {
		return nil, fmt.Errorf("read listening scripts: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var script string
		if err := rows.Scan(&script); err != nil {
			return nil, err
		}
		out = append(out, script)
	}
	return out, rows.Err()
}

const sessionFields = `
	s.id::text, s.status, t.title, s.duration_minutes, s.created_at, s.expires_at,
	GREATEST(0, CEIL(EXTRACT(EPOCH FROM (s.expires_at - now()))))::int,
	s.reused_test, s.draft_answers, s.test_id, s.question_ids, s.parts_played`

func scan(row pgx.Row) (Session, error) {
	var s Session
	var drafts []byte
	err := row.Scan(&s.ID, &s.Status, &s.TestTitle, &s.DurationMinutes, &s.CreatedAt, &s.ExpiresAt,
		&s.SecondsRemaining, &s.ReusedTest, &drafts, &s.testID, &s.questionIDs, &s.PartsPlayed)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("read listening mock: %w", err)
	}
	s.DraftAnswers = []models.AnswerSubmission{}
	if len(drafts) > 0 {
		if err := json.Unmarshal(drafts, &s.DraftAnswers); err != nil {
			return Session{}, fmt.Errorf("decode listening drafts: %w", err)
		}
	}
	if s.Status != "in_progress" {
		s.SecondsRemaining = 0
	}
	return s, nil
}

func (s *Service) live(ctx context.Context, db database.DB, userID string) (Session, error) {
	return scan(db.QueryRow(ctx, `SELECT `+sessionFields+`
		FROM listening_mock_sessions s JOIN listening_tests t ON t.id = s.test_id
		WHERE s.user_id = $1 AND s.status = 'in_progress' AND NOT s.in_full_mock`, userID))
}

func (s *Service) byID(ctx context.Context, db database.DB, userID, id string) (Session, error) {
	return scan(db.QueryRow(ctx, `SELECT `+sessionFields+`
		FROM listening_mock_sessions s JOIN listening_tests t ON t.id = s.test_id
		WHERE s.id::text = $1 AND s.user_id = $2`, id, userID))
}

// hydrate attaches the parts, groups and public questions, numbered 1-40.
func (s *Service) hydrate(ctx context.Context, session Session) (Session, error) {
	bank, err := s.questions.ByIDs(ctx, session.questionIDs)
	if err != nil {
		return Session{}, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT p.part_no, p.setting, p.reading_seconds, g.id, g.type_id, g.instructions, g.heading,
		       COALESCE(g.image_url, ''), q.id
		  FROM listening_parts p
		  JOIN listening_question_groups g ON g.part_id = p.id
		  JOIN questions q ON q.listening_group_id = g.id
		 WHERE p.test_id = $1 AND q.id = ANY($2)
		 ORDER BY p.part_no, g.position, q.group_position`, session.testID, session.questionIDs)
	if err != nil {
		return Session{}, fmt.Errorf("read listening paper: %w", err)
	}
	defer rows.Close()

	number := 0
	for rows.Next() {
		var partNo, reading int
		var setting, groupID, typeID, instructions, heading, image, qid string
		if err := rows.Scan(&partNo, &setting, &reading, &groupID, &typeID, &instructions, &heading, &image, &qid); err != nil {
			return Session{}, err
		}
		number++
		if len(session.Parts) == 0 || session.Parts[len(session.Parts)-1].PartNo != partNo {
			session.Parts = append(session.Parts, Part{PartNo: partNo, Setting: setting, ReadingSeconds: reading, FirstNumber: number})
		}
		part := &session.Parts[len(session.Parts)-1]
		part.LastNumber = number
		if len(part.Groups) == 0 || part.Groups[len(part.Groups)-1].ID != groupID {
			part.Groups = append(part.Groups, Group{ID: groupID, TypeID: typeID, Instructions: instructions,
				Heading: heading, ImageURL: image, FirstNumber: number})
		}
		group := &part.Groups[len(part.Groups)-1]
		if q, ok := bank[qid]; ok {
			public := q.PublicQuestion()
			// The part's script is the recording; a question carries none.
			public.AudioTranscript, public.ScriptOnRequest = "", false
			group.Questions = append(group.Questions, public)
		}
	}
	return session, rows.Err()
}
