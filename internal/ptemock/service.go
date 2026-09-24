// Package ptemock runs PTE Academic mock tests: the full test and one
// sectional test per skill, on one engine.
//
// A paper is taken the way the real test is taken: one item per screen, in the
// order of the test, with no way back. Where a learner is, and how much time
// each clock has left, is kept here rather than in the browser, so a reload
// resumes the same item with the same time and cannot buy any. Speaking items
// run their own preparation and recording windows in the player; the Essay,
// Summarize Written Text and Summarize Spoken Text each have a clock of their
// own; the rest of Reading and of Listening each share one clock per part.
//
// When the last item is done (or its time runs out) the paper is scored in the
// background: answer-key items at once, spoken and written items by the
// evaluators, each mark saved as it arrives so an interrupted run resumes. The
// report gives the 10-90 skill scores, and the overall for a full test.
package ptemock

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/reading"
	"github.com/prepyo/backend/internal/scoring"
)

// ExamVersionID is the PTE version the tasks and scale are written for.
const ExamVersionID = "pte-2026-01"

// SubmitGrace is how long after a clock runs out an answer is still taken as
// sent: the browser sends it when its countdown reaches zero, and the request
// needs time to arrive. Past it, the item is closed with whatever was saved.
const SubmitGrace = 20 * time.Second

// StaleAfter is how long an unfinished paper stays open. A test is one sitting;
// one left for a day is closed when the learner next starts that kind.
const StaleAfter = 24 * time.Hour

var (
	ErrNotPTE          = errors.New("PTE mocks are PTE tests")
	ErrUnknownKind     = errors.New("unknown PTE mock kind")
	ErrBankTooSmall    = errors.New("the bank cannot deal this paper")
	ErrSessionNotFound = errors.New("pte mock not found")
	ErrNotInProgress   = errors.New("this pte mock is not in progress")
	ErrOutOfOrder      = errors.New("that is not the item on screen")
	ErrNoDraft         = errors.New("this item's answer is recorded, not drafted")
	ErrNotFinished     = errors.New("this pte mock has not been scored")
)

// Evaluator rates spoken and written items. evaluations.Service implements it.
type Evaluator interface {
	EvaluatePTEMockSpeaking(ctx context.Context, userID string, question models.Question,
		transcript string, durationSeconds int, delivery scoring.Delivery, sessionID string) (models.Evaluation, error)
	EvaluatePTEMockWriting(ctx context.Context, userID string, question models.Question, text, sessionID string) (models.Evaluation, error)
}

// Transcriber turns a recording into text on the server: Whisper on the audio
// provider (Groq). ai.Gateway implements it. Spoken answers are transcribed
// the way the IELTS speaking mock does it (speakingmock.SaveAnswer): by the
// server wherever it can, with the browser's transcript as the fallback.
type Transcriber interface {
	SpeakingAvailable() bool
	Transcribe(ctx context.Context, audio []byte, format string) (string, ai.Usage, error)
}

type Service struct {
	db        *pgxpool.Pool
	listener  Transcriber
	questions *questions.Repository
	reading   *reading.Repository
	mocks     *mocks.Repository
	billing   *billing.Service
	xp        *gamification.Service
	evaluator Evaluator
	log       *slog.Logger

	// scoring holds the papers being scored by this process, so a second
	// request for the same report does not start a second run.
	scoring sync.Map
	// now is the clock; tests move it.
	now func() time.Time
	// manualScoring leaves scoring to an explicit ScorePaper call; tests set
	// it so that a run does not race the assertions.
	manualScoring bool
}

func NewService(db *pgxpool.Pool, questionRepo *questions.Repository, readingRepo *reading.Repository,
	mockRepo *mocks.Repository, billingService *billing.Service, xp *gamification.Service,
	evaluator Evaluator, listener Transcriber, log *slog.Logger) *Service {
	return &Service{db: db, questions: questionRepo, reading: readingRepo, mocks: mockRepo,
		billing: billingService, xp: xp, evaluator: evaluator, listener: listener, log: log, now: time.Now}
}

// ServerTranscription reports whether spoken answers are transcribed here.
// Without it they are scored from the browser's transcript.
func (s *Service) ServerTranscription() bool {
	return s.listener != nil && s.listener.SpeakingAvailable()
}

// ---------------------------------------------------------------------------
// Starting a paper
// ---------------------------------------------------------------------------

// Start deals a paper of a kind, or returns the learner's open one of that
// kind at no cost. A full test spends one from the plan's full-mock allowance;
// a sectional test spends SectionMockSubTests sub-tests.
func (s *Service) Start(ctx context.Context, user models.User, kind Kind) (View, error) {
	if user.TargetExam != models.ExamPTE {
		return View{}, ErrNotPTE
	}
	blueprint, ok := BlueprintFor(kind)
	if !ok {
		return View{}, ErrUnknownKind
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE pte_mock_sessions SET status = 'abandoned', completed_at = now()
		 WHERE user_id = $1 AND kind = $2 AND status = 'in_progress' AND created_at < $3`,
		user.ID, kind, s.now().Add(-StaleAfter)); err != nil {
		return View{}, fmt.Errorf("close stale pte mock: %w", err)
	}
	if live, err := liveSession(ctx, s.db, user.ID, kind); err == nil {
		return s.Get(ctx, user, live.ID)
	} else if !errors.Is(err, ErrSessionNotFound) {
		return View{}, err
	}

	// An early check saves dealing a paper that cannot be started; the binding
	// one runs under the user lock below.
	if err := s.checkAllowance(ctx, s.db, user, blueprint); err != nil {
		return View{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return View{}, fmt.Errorf("begin pte mock: %w", err)
	}
	defer tx.Rollback(ctx)

	if s.billing != nil {
		if err := billing.LockUserForQuota(ctx, tx, user.ID); err != nil {
			return View{}, err
		}
		if err := s.checkAllowance(ctx, tx, user, blueprint); err != nil {
			return View{}, err
		}
	}

	// Dealt in the transaction that writes the paper, with the chosen
	// questions locked, so none can be deleted between choosing it and
	// writing it down.
	items, missing, err := s.deal(ctx, tx, user.ID, blueprint)
	if err != nil {
		return View{}, err
	}

	id, err := insertPaper(ctx, tx, user.ID, blueprint, ExamVersionID, items, missing)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// A second start raced the first; resume what it dealt.
			tx.Rollback(ctx)
			live, liveErr := liveSession(ctx, s.db, user.ID, kind)
			if liveErr != nil {
				return View{}, liveErr
			}
			return s.Get(ctx, user, live.ID)
		}
		return View{}, fmt.Errorf("create pte mock: %w", err)
	}

	// A sectional test is charged in the same transaction as its paper, so a
	// failed start costs nothing. A full test is counted by its session row.
	if !blueprint.FullAllowance && s.billing != nil {
		if _, err := s.billing.RecordSessionStartCredits(ctx, tx, user, string(models.ExamPTE),
			string(blueprint.Skills[0]), "pte-mock:"+id, billing.SectionMockSubTests); err != nil {
			return View{}, err
		}
		// A sectional test the AI marks takes that cost from the grading pool.
		if err := s.billing.RecordAIGrading(ctx, tx, user.ID, blueprint.GradingUnits,
			"mock-pte-"+string(blueprint.Kind), id); err != nil {
			return View{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return View{}, fmt.Errorf("commit pte mock: %w", err)
	}
	return s.Get(ctx, user, id)
}

func (s *Service) checkAllowance(ctx context.Context, db database.DB, user models.User, blueprint Blueprint) error {
	if s.billing == nil {
		return nil
	}
	if blueprint.FullAllowance {
		_, err := s.billing.CheckMockAllowance(ctx, db, user)
		return err
	}
	if _, err := s.billing.CheckSubTestCredits(ctx, db, user, billing.SectionMockSubTests); err != nil {
		return err
	}
	if blueprint.GradingUnits > 0 {
		if _, err := s.billing.CheckAIGradings(ctx, db, user, blueprint.GradingUnits); err != nil {
			return err
		}
	}
	return nil
}

// deal chooses a paper's items: for each slot of the blueprint, questions of
// that task the learner has not met in a mock, then not in practice, then at
// random. A task the bank has none of is left out and named.
func (s *Service) deal(ctx context.Context, db database.DB, userID string, blueprint Blueprint) ([]dealt, []string, error) {
	var items []dealt
	var missing []string
	taken := []string{}
	perPart := map[Part]int{}

	for _, slot := range blueprint.Slots {
		task := Tasks[slot.Task]
		rows, err := db.Query(ctx, `
			SELECT q.id
			  FROM questions q
			 WHERE q.is_published AND q.skill = $2 AND q.type_id = ANY($3)
			   AND 'PTE' = ANY(q.supported_exams) AND NOT (q.id = ANY($5))
			 ORDER BY (SELECT count(*) FROM pte_mock_items i
			             JOIN pte_mock_sessions p ON p.id = i.session_id
			            WHERE p.user_id = $1 AND i.question_id = q.id),
			          EXISTS (SELECT 1 FROM practice_attempts a WHERE a.user_id = $1 AND a.question_id = q.id),
			          random()
			 LIMIT $4
			   FOR SHARE OF q`, userID, task.Skill, task.TypeIDs, slot.Count, taken)
		if err != nil {
			return nil, nil, fmt.Errorf("deal %s: %w", slot.Task, err)
		}
		var ids []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, nil, fmt.Errorf("scan %s: %w", slot.Task, err)
			}
			ids = append(ids, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
		if len(ids) == 0 {
			missing = append(missing, task.Name)
			continue
		}

		var bank map[string]models.Question
		if slot.Task == "RO" {
			if bank, err = questions.NewRepository(db).ByIDs(ctx, ids); err != nil {
				return nil, nil, err
			}
		}
		for _, id := range ids {
			item := dealt{QuestionID: id, Task: slot.Task, Part: task.Part}
			// The bank stores the boxes in the right order; the paper keeps
			// the order it dealt them in, so a reload shows the same screen.
			if q, ok := bank[id]; ok {
				for _, o := range q.Options {
					item.OptionOrder = append(item.OptionOrder, o.ID)
				}
				rand.Shuffle(len(item.OptionOrder), func(i, j int) {
					item.OptionOrder[i], item.OptionOrder[j] = item.OptionOrder[j], item.OptionOrder[i]
				})
			}
			items = append(items, item)
			taken = append(taken, id)
			perPart[task.Part]++
		}
	}

	// Every part the kind covers must have something in it, or the paper is
	// not the test it says it is.
	for _, part := range blueprint.Parts() {
		if perPart[part] == 0 {
			return nil, nil, fmt.Errorf("%w: no %s items", ErrBankTooSmall, part)
		}
	}
	return items, missing, nil
}

// ---------------------------------------------------------------------------
// Taking a paper
// ---------------------------------------------------------------------------

// Get is a paper as the learner sees it now: its clocks brought up to date,
// and the item on screen.
func (s *Service) Get(ctx context.Context, user models.User, id string) (View, error) {
	session, items, kick, err := s.progress(ctx, user.ID, id, func(sessionRow, []itemRow) error { return nil })
	if err != nil {
		return View{}, err
	}
	if kick || s.stale(session) {
		s.kick(user, session.ID)
	}
	return s.view(ctx, session, items)
}

// Answer is what the learner sends for the item on screen.
type Answer struct {
	Response        *models.AnswerSubmission `json:"response"`
	Transcript      string                   `json:"transcript"`
	DurationSeconds int                      `json:"durationSeconds"`
	Delivery        *Delivery                `json:"delivery"`
	// Audio is a spoken answer's recording (base64, in Format: "mp3" or
	// "wav"), sent with every spoken answer as the IELTS speaking mock sends
	// it. It is transcribed and dropped, never stored; Transcript is what the
	// browser heard, kept if the server cannot transcribe.
	Audio  string `json:"audio,omitempty"`
	Format string `json:"format,omitempty"`
}

// Where a spoken answer's transcript came from.
const (
	SourceBrowser = "browser"
	SourceServer  = "server"
)

// transcribeTimeout bounds the Whisper call made while the learner waits to
// move on.
const transcribeTimeout = 30 * time.Second

// Limits on what a browser may send.
const (
	maxTranscriptChars = 6000
	maxRecordSeconds   = 180
	maxTextChars       = 20000
)

// Submit records the answer to the item on screen and moves to the next one.
// An item that is not on screen - one already passed, or not yet reached - is
// refused: the test only goes forward.
func (s *Service) Submit(ctx context.Context, user models.User, id string, position int, answer Answer) (View, error) {
	// As speakingmock.SaveAnswer: the server's transcript where it has one,
	// otherwise the browser's. A failed transcription loses nothing.
	source := SourceBrowser
	if answer.Audio != "" {
		if text, ok := s.transcribe(ctx, user.ID, id, position, answer); ok {
			answer.Transcript, source = text, SourceServer
		}
	}
	answer.Audio = ""

	session, items, kick, err := s.progress(ctx, user.ID, id, func(session sessionRow, items []itemRow) error {
		if session.Status != StatusInProgress {
			return ErrNotInProgress
		}
		i := itemIndex(items, position)
		if position != session.Current || i < 0 {
			return ErrOutOfOrder
		}
		it := &items[i]
		task := Tasks[it.Task]
		now := s.now()
		if task.Marking == MarkSpoken {
			it.Transcript = clip(strings.TrimSpace(answer.Transcript), maxTranscriptChars)
			it.Duration = min(max(answer.DurationSeconds, 0), maxRecordSeconds)
			it.Delivery = saneDelivery(answer.Delivery)
			it.TranscriptSource = ""
			if it.Transcript != "" {
				it.TranscriptSource = source
			}
			it.Response = nil
			it.Status = ItemAnswered
			if it.Transcript == "" && it.Duration == 0 {
				it.Status = ItemSkipped
			}
		} else {
			it.Response = cleanResponse(answer.Response, it.QuestionID)
			it.Status = ItemAnswered
			if !answered(it.Response) {
				it.Status = ItemSkipped
			}
		}
		it.AnsweredAt = &now
		it.dirty = true
		return nil
	})
	if err != nil {
		return View{}, err
	}
	if kick {
		s.kick(user, session.ID)
	}
	return s.view(ctx, session, items)
}

// transcribe turns an uploaded recording into text with Whisper. It runs
// before the paper is locked,
// so a slow provider never holds the learner's paper, and only for the spoken
// item actually on screen, so the endpoint cannot be used as a free
// transcription service. Any failure leaves the answer as the browser sent it.
func (s *Service) transcribe(ctx context.Context, userID, id string, position int, answer Answer) (string, bool) {
	if !s.ServerTranscription() || !ai.AudioFormats[answer.Format] {
		return "", false
	}
	session, err := sessionByID(ctx, s.db, userID, id, false)
	if err != nil || session.Status != StatusInProgress || session.Current != position {
		return "", false
	}
	var task string
	if err := s.db.QueryRow(ctx, `SELECT task FROM pte_mock_items WHERE session_id = $1 AND position = $2`,
		session.ID, position).Scan(&task); err != nil || Tasks[task].Marking != MarkSpoken {
		return "", false
	}
	audio, err := base64.StdEncoding.DecodeString(answer.Audio)
	if err != nil || len(audio) == 0 {
		return "", false
	}
	callCtx, cancel := context.WithTimeout(ctx, transcribeTimeout)
	defer cancel()
	text, _, err := s.listener.Transcribe(callCtx, audio, answer.Format)
	if err != nil {
		if s.log != nil {
			s.log.Warn("pte mock answer not transcribed", "session", id, "position", position, "err", err)
		}
		return "", false
	}
	return clip(strings.TrimSpace(text), maxTranscriptChars), true
}

// SaveDraft keeps the answer so far to the item on screen, which is what is
// marked if its time runs out. Spoken answers have no draft.
func (s *Service) SaveDraft(ctx context.Context, user models.User, id string, position int, response *models.AnswerSubmission) (View, error) {
	session, items, kick, err := s.progress(ctx, user.ID, id, func(session sessionRow, items []itemRow) error {
		if session.Status != StatusInProgress {
			return ErrNotInProgress
		}
		i := itemIndex(items, position)
		if position != session.Current || i < 0 {
			return ErrOutOfOrder
		}
		it := &items[i]
		if Tasks[it.Task].Marking == MarkSpoken {
			return ErrNoDraft
		}
		it.Response = cleanResponse(response, it.QuestionID)
		it.dirty = true
		return nil
	})
	if err != nil {
		return View{}, err
	}
	if kick {
		s.kick(user, session.ID)
	}
	return s.view(ctx, session, items)
}

// Finish ends the test where it stands. Items not yet reached are skipped and
// score zero, as unanswered items do; the paper is then scored.
func (s *Service) Finish(ctx context.Context, user models.User, id string) (View, error) {
	session, items, _, err := s.progress(ctx, user.ID, id, func(session sessionRow, items []itemRow) error {
		if session.Status == StatusScoring || session.Status == StatusCompleted {
			return nil
		}
		if session.Status != StatusInProgress {
			return ErrNotInProgress
		}
		now := s.now()
		for i := range items {
			if items[i].Position >= session.Current && items[i].Status == ItemPending {
				closeItem(&items[i], now)
				items[i].Status = ItemSkipped
				if answered(items[i].Response) {
					items[i].Status = ItemAnswered
				}
			}
		}
		return nil
	})
	if err != nil {
		return View{}, err
	}
	if session.Status == StatusScoring {
		s.kick(user, session.ID)
	}
	return s.view(ctx, session, items)
}

// Abandon closes an open paper without a score. What it cost is not
// returned, as a test that was booked and not sat is not.
func (s *Service) Abandon(ctx context.Context, user models.User, id string) error {
	if !isUUID(id) {
		return ErrSessionNotFound
	}
	tag, err := s.db.Exec(ctx, `
		UPDATE pte_mock_sessions SET status = 'abandoned', completed_at = now()
		 WHERE id = $1 AND user_id = $2 AND status = 'in_progress'`, id, user.ID)
	if err != nil {
		return fmt.Errorf("abandon pte mock: %w", err)
	}
	if tag.RowsAffected() == 0 {
		if _, err := sessionByID(ctx, s.db, user.ID, id, false); err != nil {
			return err
		}
		return ErrNotInProgress
	}
	return nil
}

// progress runs one change to a paper under its lock: it brings the clocks up
// to date, applies change, then moves on to the next item, and writes it all.
// kick reports that the paper has just run out of items and needs scoring.
func (s *Service) progress(ctx context.Context, userID, id string,
	change func(sessionRow, []itemRow) error) (sessionRow, []itemRow, bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return sessionRow{}, nil, false, fmt.Errorf("begin pte mock: %w", err)
	}
	defer tx.Rollback(ctx)

	session, err := sessionByID(ctx, tx, userID, id, true)
	if err != nil {
		return sessionRow{}, nil, false, err
	}
	items, err := loadItems(ctx, tx, session.ID)
	if err != nil {
		return sessionRow{}, nil, false, err
	}

	wasScoring := session.Status == StatusScoring
	s.advance(&session, items)
	if err := change(session, items); err != nil {
		return sessionRow{}, nil, false, err
	}
	s.advance(&session, items)
	kick := session.Status == StatusScoring && !wasScoring

	if err := saveProgress(ctx, tx, session, items); err != nil {
		return sessionRow{}, nil, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return sessionRow{}, nil, false, fmt.Errorf("commit pte mock: %w", err)
	}
	return session, items, kick, nil
}

// advance brings a paper up to date with the clock. It closes every item whose
// countdown has run out (keeping a saved draft as the answer), moves the
// learner past answered and closed items, starts the countdown of the item now
// on screen, and sends a paper with nothing left to scoring.
func (s *Service) advance(session *sessionRow, items []itemRow) {
	if session.Status != StatusInProgress {
		return
	}
	now := s.now()
	for session.Current <= session.Total {
		i := itemIndex(items, session.Current)
		if i < 0 {
			// The question was deleted from the bank, and its item with it.
			session.Current++
			continue
		}
		it := &items[i]
		if it.Status != ItemPending {
			session.Current++
			continue
		}
		key := clockKey(*it)
		if key != "" {
			deadline, started := session.Deadlines[key]
			// The grace period lets an answer already on its way arrive for the
			// item on screen. It never opens a new item on a clock that has run
			// out: that item would appear with no time at all.
			cutoff := deadline.Add(SubmitGrace)
			if it.ShownAt == nil {
				cutoff = deadline
			}
			if !started {
				session.Deadlines[key] = now.Add(time.Duration(clockSeconds(key, items)) * time.Second)
			} else if now.After(cutoff) {
				// Time is up for this clock: every item left on it closes
				// with what was saved, and the test moves on.
				for j := i; j < len(items) && clockKey(items[j]) == key; j++ {
					if items[j].Status == ItemPending {
						closeItem(&items[j], deadline)
					}
				}
				continue
			}
		}
		if it.ShownAt == nil {
			it.ShownAt = &now
			it.dirty = true
		}
		break
	}
	if session.Current > session.Total {
		session.Status = StatusScoring
		started := now
		session.ScoringStartedAt = &started
	}
}

// itemIndex is where the item at a position sits in a paper's items, or -1
// when there is none: an item goes with its question if that is deleted.
func itemIndex(items []itemRow, position int) int {
	if i := position - 1; i >= 0 && i < len(items) && items[i].Position == position {
		return i
	}
	for i := range items {
		if items[i].Position == position {
			return i
		}
	}
	return -1
}

// closeItem ends an item whose time is up: a saved draft is its answer, and
// without one it timed out.
func closeItem(it *itemRow, at time.Time) {
	it.Status = ItemTimedOut
	if answered(it.Response) {
		it.Status = ItemAnswered
	}
	it.AnsweredAt = &at
	it.dirty = true
}

// cleanResponse keeps only what an answer may carry, for the item's own
// question.
func cleanResponse(sub *models.AnswerSubmission, questionID string) *models.AnswerSubmission {
	if sub == nil {
		return nil
	}
	clean := models.AnswerSubmission{
		QuestionID:      questionID,
		Exam:            models.ExamPTE,
		TextResponse:    clip(sub.TextResponse, maxTextChars),
		SelectedOptions: sub.SelectedOptions,
		BlankResponses:  sub.BlankResponses,
	}
	if len(clean.SelectedOptions) > 50 {
		clean.SelectedOptions = clean.SelectedOptions[:50]
	}
	if len(clean.BlankResponses) > 50 {
		clean.BlankResponses = nil
	}
	return &clean
}

// answered reports whether a response carries anything at all.
func answered(sub *models.AnswerSubmission) bool {
	if sub == nil {
		return false
	}
	if strings.TrimSpace(sub.TextResponse) != "" || len(sub.SelectedOptions) > 0 {
		return true
	}
	for _, v := range sub.BlankResponses {
		if strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

// saneDelivery drops measurements that cannot describe a recording, as the
// practice endpoint does: they arrive from the learner's browser.
func saneDelivery(d *Delivery) *Delivery {
	if d == nil {
		return nil
	}
	ok := d.DurationSeconds > 0 && d.DurationSeconds <= maxRecordSeconds &&
		d.PauseCount >= 0 && d.PauseCount <= d.DurationSeconds &&
		d.LongestPauseSeconds >= 0 && d.LongestPauseSeconds <= float64(d.DurationSeconds) &&
		d.SpeakingRatio >= 0 && d.SpeakingRatio <= 1
	if !ok {
		return nil
	}
	return d
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
