// Package speakingmock runs the IELTS Speaking mock: Part 1, the Part 2 long
// turn and the Part 3 discussion in one sitting, rated as one performance.
//
// The examiner's questions are spoken by the browser's voice (or the server's,
// where the browser has none). Each answer is recorded and transcribed as soon
// as it ends: on the audio provider (Whisper on Groq) where one is configured,
// otherwise from the browser's own recogniser. The whole test is then rated
// once, across all three parts, as IELTS rates it. Transcripts cannot show
// pronunciation, so the estimate covers the other three criteria and says so.
package speakingmock

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
)

// MockID is the mocks row speaking attempts are recorded against.
const MockID = "mock-ielts-speaking-gen"

// DurationMinutes is the time a paper stays open: the test itself runs 11-14
// minutes, and recording, uploading and a slow connection need room.
const DurationMinutes = 40

var (
	ErrNotIELTS         = errors.New("the speaking mock is an IELTS paper")
	ErrNoSet            = errors.New("no speaking test is available")
	ErrSessionNotFound  = errors.New("speaking mock not found")
	ErrAlreadySubmitted = errors.New("this speaking mock has already been submitted")
	ErrPaperClosed      = errors.New("this speaking mock is closed")
	ErrUnknownStep      = errors.New("that question is not on this test")
	ErrTooLittle        = errors.New("too few questions have been answered to rate the test")
	// ErrExpiredTooLittle closes a test whose time ran out before enough was
	// said to rate it, so it no longer blocks a new one.
	ErrExpiredTooLittle = errors.New("time ran out before enough was said to rate the test")
)

// Answer time limits. Part 1 answers are short; the Part 2 long turn is up to
// two minutes after one minute's preparation; Part 3 answers are extended.
const (
	part1Seconds    = 40
	cueCardPrep     = 60
	cueCardSeconds  = 120
	roundingSeconds = 30
	part3Seconds    = 60
)

// Step is one thing the examiner says, and the answer it asks for.
type Step struct {
	Key  string `json:"key"`
	Part int    `json:"part"`
	// Lead is said before the question: a part's introduction or a new topic.
	Lead     string `json:"lead,omitempty"`
	Question string `json:"question"`
	// CueCard is shown, not read, for the Part 2 long turn.
	CueCard       *CueCard `json:"cueCard,omitempty"`
	PrepSeconds   int      `json:"prepSeconds,omitempty"`
	AnswerSeconds int      `json:"answerSeconds"`
}

// CueCard is the Part 2 task card.
type CueCard struct {
	Task    string   `json:"task"`
	Points  []string `json:"points"`
	Explain string   `json:"explain"`
}

// Answer is one recorded answer, transcribed.
type Answer struct {
	Key             string `json:"key"`
	Transcript      string `json:"transcript"`
	DurationSeconds int    `json:"durationSeconds"`
	// Source says where the transcript came from: "server" (Whisper) or
	// "browser" (the device's recogniser).
	Source string `json:"source"`
}

// Session is a dealt speaking test.
type Session struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	Title            string    `json:"title"`
	SecondsRemaining int       `json:"secondsRemaining"`
	Steps            []Step    `json:"steps"`
	Answers          []Answer  `json:"answers"`
	CreatedAt        time.Time `json:"createdAt"`

	setID     string
	expiresAt time.Time
}

// Result is a rated speaking test.
type Result struct {
	Attempt    *models.MockAttempt `json:"attempt"`
	Evaluation models.Evaluation   `json:"evaluation"`
	XPAwarded  int                 `json:"xpAwarded"`
}

// Transcriber turns a recording into text. ai.Gateway implements it.
type Transcriber interface {
	SpeakingAvailable() bool
	Transcribe(ctx context.Context, audio []byte, format string) (string, ai.Usage, error)
}

// Rater rates a whole test. evaluations.Service implements it.
type Rater interface {
	EvaluateSpeakingTest(ctx context.Context, user models.User, sessionID string, turns []ai.SpeakingTestTurn) (models.Evaluation, error)
}

type Service struct {
	db       *pgxpool.Pool
	mocks    *mocks.Repository
	billing  *billing.Service
	xp       *gamification.Service
	listener Transcriber
	rater    Rater
}

func NewService(db *pgxpool.Pool, mockRepo *mocks.Repository, billingService *billing.Service,
	xp *gamification.Service, listener Transcriber, rater Rater) *Service {
	return &Service{db: db, mocks: mockRepo, billing: billingService, xp: xp, listener: listener, rater: rater}
}

type setContent struct {
	Part1 []struct {
		Topic     string   `json:"topic"`
		Questions []string `json:"questions"`
	} `json:"part1"`
	Part2 struct {
		Topic    string   `json:"topic"`
		Cue      string   `json:"cue"`
		Points   []string `json:"points"`
		Explain  string   `json:"explain"`
		Rounding string   `json:"rounding"`
	} `json:"part2"`
	Part3 struct {
		Topic     string   `json:"topic"`
		Questions []string `json:"questions"`
	} `json:"part3"`
}

// StepsFor builds the examiner's script for a set, in the order of the test.
func StepsFor(content setContent) []Step {
	var steps []Step
	for t, topic := range content.Part1 {
		for q, question := range topic.Questions {
			step := Step{Key: fmt.Sprintf("p1-%d-%d", t+1, q+1), Part: 1, Question: question, AnswerSeconds: part1Seconds}
			if q == 0 {
				if t == 0 {
					step.Lead = "In this first part, I'd like to ask you some questions about yourself. Let's talk about " + topic.Topic + "."
				} else {
					step.Lead = "Now let's talk about " + topic.Topic + "."
				}
			}
			steps = append(steps, step)
		}
	}
	steps = append(steps, Step{
		Key:  "p2-talk",
		Part: 2,
		Lead: "Now I'm going to give you a topic, and I'd like you to talk about it for one to two minutes. " +
			"Before you talk, you'll have one minute to think about what you're going to say. You can make some notes if you wish.",
		Question:      "Remember, you have one to two minutes for this. I'll tell you when the time is up. Can you start speaking now, please?",
		CueCard:       &CueCard{Task: content.Part2.Cue, Points: content.Part2.Points, Explain: content.Part2.Explain},
		PrepSeconds:   cueCardPrep,
		AnswerSeconds: cueCardSeconds,
	})
	if content.Part2.Rounding != "" {
		steps = append(steps, Step{Key: "p2-round", Part: 2, Question: content.Part2.Rounding, AnswerSeconds: roundingSeconds})
	}
	for q, question := range content.Part3.Questions {
		step := Step{Key: fmt.Sprintf("p3-%d", q+1), Part: 3, Question: question, AnswerSeconds: part3Seconds}
		if q == 0 {
			step.Lead = "We've been talking about " + content.Part2.Topic + ", and I'd like to discuss with you one or two more general questions related to this. Let's consider " + content.Part3.Topic + "."
		}
		steps = append(steps, step)
	}
	return steps
}

// Start deals a test, or returns the learner's open one.
func (s *Service) Start(ctx context.Context, user models.User, charge bool) (Session, error) {
	return s.start(ctx, user, charge, false)
}

// StartForFullMock deals a fresh test for a full mock's Speaking section. It
// is not charged (the full mock was), and it is kept apart from any speaking
// mock the learner has open.
func (s *Service) StartForFullMock(ctx context.Context, user models.User) (Session, error) {
	return s.start(ctx, user, false, true)
}

func (s *Service) start(ctx context.Context, user models.User, charge, fullMock bool) (Session, error) {
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
	charge = charge && !fullMock && s.billing != nil
	if charge {
		if _, err := s.billing.CheckSubTestCredits(ctx, s.db, user, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckAIGradings(ctx, s.db, user, billing.IELTSSpeakingMockUnits); err != nil {
			return Session{}, err
		}
	}

	var setID string
	err := s.db.QueryRow(ctx, `
		SELECT id FROM speaking_mock_sets st
		 WHERE is_published
		 ORDER BY EXISTS (SELECT 1 FROM speaking_mock_sessions m WHERE m.user_id = $1 AND m.set_id = st.id), random()
		 LIMIT 1`, user.ID).Scan(&setID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNoSet
	}
	if err != nil {
		return Session{}, fmt.Errorf("pick speaking set: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Session{}, err
	}
	defer tx.Rollback(ctx)
	if charge {
		if err := billing.LockUserForQuota(ctx, tx, user.ID); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckSubTestCredits(ctx, tx, user, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
		if _, err := s.billing.CheckAIGradings(ctx, tx, user, billing.IELTSSpeakingMockUnits); err != nil {
			return Session{}, err
		}
	}
	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO speaking_mock_sessions (user_id, set_id, expires_at, in_full_mock)
		VALUES ($1, $2, now() + make_interval(mins => $3), $4) RETURNING id::text`, user.ID, setID, DurationMinutes, fullMock).Scan(&id)
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
		return Session{}, fmt.Errorf("create speaking mock: %w", err)
	}
	if charge {
		if _, err := s.billing.RecordSessionStartCredits(ctx, tx, user, string(models.ExamIELTS),
			string(models.SkillSpeaking), "speaking-mock:"+id, billing.SectionMockSubTests); err != nil {
			return Session{}, err
		}
		// The AI marks this paper, so it takes its cost from the grading pool.
		if err := s.billing.RecordAIGrading(ctx, tx, user.ID, billing.IELTSSpeakingMockUnits, "mock-ielts-speaking", id); err != nil {
			return Session{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Session{}, err
	}
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.hydrate(ctx, session)
}

// Resume returns a test by id.
func (s *Service) Resume(ctx context.Context, user models.User, id string) (Session, error) {
	session, err := s.byID(ctx, s.db, user.ID, id)
	if err != nil {
		return Session{}, err
	}
	return s.hydrate(ctx, session)
}

// Step returns one step of a test, for the examiner's voice.
func (s *Service) Step(ctx context.Context, user models.User, id, key string) (Step, error) {
	session, err := s.Resume(ctx, user, id)
	if err != nil {
		return Step{}, err
	}
	for _, step := range session.Steps {
		if step.Key == key {
			return step, nil
		}
	}
	return Step{}, ErrUnknownStep
}

// AnswerParams is one recorded answer as uploaded.
type AnswerParams struct {
	Key             string
	AudioBase64     string
	Format          string
	DurationSeconds int
	// BrowserTranscript is what the device's recogniser heard, used when the
	// server cannot transcribe.
	BrowserTranscript string
}

// SaveAnswer transcribes one answer and stores it on the test.
func (s *Service) SaveAnswer(ctx context.Context, user models.User, id string, p AnswerParams) (Answer, error) {
	session, err := s.Resume(ctx, user, id)
	if err != nil {
		return Answer{}, err
	}
	if session.Status != "in_progress" || time.Now().After(session.expiresAt) {
		return Answer{}, ErrPaperClosed
	}
	known := false
	for _, step := range session.Steps {
		if step.Key == p.Key {
			known = true
			break
		}
	}
	if !known {
		return Answer{}, ErrUnknownStep
	}

	answer := Answer{Key: p.Key, DurationSeconds: max(p.DurationSeconds, 0), Source: "browser",
		Transcript: strings.TrimSpace(p.BrowserTranscript)}
	if audio, err := base64.StdEncoding.DecodeString(p.AudioBase64); err == nil && len(audio) > 0 &&
		s.listener != nil && s.listener.SpeakingAvailable() && ai.AudioFormats[p.Format] {
		if text, _, err := s.listener.Transcribe(ctx, audio, p.Format); err == nil {
			answer.Transcript = strings.TrimSpace(text)
			answer.Source = "server"
		}
		// A failed transcription keeps the browser's transcript: the answer
		// is not lost because one provider call failed.
	}

	// Replace any earlier take of the same question.
	kept := make([]Answer, 0, len(session.Answers)+1)
	for _, a := range session.Answers {
		if a.Key != p.Key {
			kept = append(kept, a)
		}
	}
	kept = append(kept, answer)
	body, err := json.Marshal(kept)
	if err != nil {
		return Answer{}, err
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE speaking_mock_sessions SET answers = $3
		 WHERE id::text = $1 AND user_id = $2 AND status = 'in_progress'`, id, user.ID, string(body))
	if err != nil {
		return Answer{}, fmt.Errorf("save speaking answer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Answer{}, ErrPaperClosed
	}
	return answer, nil
}

// Submit rates the whole test and records the Speaking estimate.
func (s *Service) Submit(ctx context.Context, user models.User, id string) (Result, error) {
	session, err := s.Resume(ctx, user, id)
	if err != nil {
		return Result{}, err
	}
	if session.Status != "in_progress" {
		return Result{}, ErrAlreadySubmitted
	}

	byKey := map[string]Answer{}
	for _, a := range session.Answers {
		byKey[a.Key] = a
	}
	var turns []ai.SpeakingTestTurn
	heard := 0
	for _, step := range session.Steps {
		a := byKey[step.Key]
		question := step.Question
		if step.CueCard != nil {
			question = step.CueCard.Task + " (Part 2 long turn)"
		}
		if strings.TrimSpace(a.Transcript) != "" {
			heard++
		}
		turns = append(turns, ai.SpeakingTestTurn{Part: step.Part, Question: question, Answer: a.Transcript, Seconds: a.DurationSeconds})
	}
	// A test with almost nothing said cannot be rated as a test.
	if heard < 3 {
		if session.SecondsRemaining <= 0 {
			if _, err := s.db.Exec(ctx, `UPDATE speaking_mock_sessions SET status = 'abandoned'
				WHERE id = $1 AND status = 'in_progress'`, session.ID); err != nil {
				return Result{}, fmt.Errorf("close speaking mock: %w", err)
			}
			return Result{}, ErrExpiredTooLittle
		}
		return Result{}, ErrTooLittle
	}

	evaluation, err := s.rater.EvaluateSpeakingTest(ctx, user, session.ID, turns)
	if err != nil {
		return Result{}, err
	}
	result := Result{Evaluation: evaluation}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback(ctx)
	var attemptID *string
	if evaluation.EstimatedScore != nil {
		band := *evaluation.EstimatedScore
		elapsed := int(time.Since(session.CreatedAt).Seconds())
		attempt, err := s.mocks.SaveAttempt(ctx, tx, mocks.SaveAttemptParams{
			UserID:          user.ID,
			MockID:          MockID,
			ExamVersionID:   "ielts-2026-01",
			Exam:            models.ExamIELTS,
			UserScore:       band,
			SkillScores:     map[models.SkillType]float64{models.SkillSpeaking: band},
			DurationSeconds: max(min(elapsed, DurationMinutes*60), 1),
		})
		if err != nil {
			return Result{}, err
		}
		result.Attempt = &attempt
		attemptID = &attempt.ID
		awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
			UserID: user.ID, Amount: gamification.XPMockCompleted,
			Reason: "Completed speaking mock", SourceKey: "speaking-mock:" + session.ID,
		})
		if err != nil {
			return Result{}, err
		}
		result.XPAwarded = awarded
		if _, err := s.xp.TouchStreak(ctx, tx, user); err != nil {
			return Result{}, err
		}
	}
	tag, err := tx.Exec(ctx, `
		UPDATE speaking_mock_sessions
		   SET status = 'submitted', submitted_at = now(), mock_attempt_id = $3,
		       evaluation_id = NULLIF($4, '')::uuid
		 WHERE id::text = $1 AND user_id = $2 AND status = 'in_progress'`, id, user.ID, attemptID, evaluation.ID)
	if err != nil {
		return Result{}, fmt.Errorf("close speaking mock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Result{}, ErrAlreadySubmitted
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, err
	}
	return result, nil
}

const sessionFields = `
	s.id::text, s.status, st.title, s.set_id, s.answers, s.created_at, s.expires_at,
	GREATEST(0, CEIL(EXTRACT(EPOCH FROM (s.expires_at - now()))))::int`

func scan(row pgx.Row) (Session, error) {
	var s Session
	var answers []byte
	err := row.Scan(&s.ID, &s.Status, &s.Title, &s.setID, &answers, &s.CreatedAt, &s.expiresAt, &s.SecondsRemaining)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("read speaking mock: %w", err)
	}
	s.Answers = []Answer{}
	if len(answers) > 0 {
		if err := json.Unmarshal(answers, &s.Answers); err != nil {
			return Session{}, fmt.Errorf("decode speaking answers: %w", err)
		}
	}
	if s.Status != "in_progress" {
		s.SecondsRemaining = 0
	}
	return s, nil
}

func (s *Service) live(ctx context.Context, db database.DB, userID string) (Session, error) {
	return scan(db.QueryRow(ctx, `SELECT `+sessionFields+`
		FROM speaking_mock_sessions s JOIN speaking_mock_sets st ON st.id = s.set_id
		WHERE s.user_id = $1 AND s.status = 'in_progress' AND NOT s.in_full_mock`, userID))
}

func (s *Service) byID(ctx context.Context, db database.DB, userID, id string) (Session, error) {
	return scan(db.QueryRow(ctx, `SELECT `+sessionFields+`
		FROM speaking_mock_sessions s JOIN speaking_mock_sets st ON st.id = s.set_id
		WHERE s.id::text = $1 AND s.user_id = $2`, id, userID))
}

func (s *Service) hydrate(ctx context.Context, session Session) (Session, error) {
	var raw []byte
	if err := s.db.QueryRow(ctx, `SELECT content FROM speaking_mock_sets WHERE id = $1`, session.setID).Scan(&raw); err != nil {
		return Session{}, fmt.Errorf("read speaking set: %w", err)
	}
	var content setContent
	if err := json.Unmarshal(raw, &content); err != nil {
		return Session{}, fmt.Errorf("decode speaking set: %w", err)
	}
	session.Steps = StepsFor(content)
	return session, nil
}
