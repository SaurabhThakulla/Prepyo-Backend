package ptemock

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// Scoring limits.
const (
	// scoringTimeout bounds one run over a paper. A full test has around
	// thirty rated items; at the concurrency below that is a few minutes at
	// worst.
	scoringTimeout = 15 * time.Minute
	// scoringConcurrency is how many items are rated at once, so a full test
	// does not send thirty requests to the provider in one burst.
	scoringConcurrency = 4
	// scoringStaleAfter is when a paper still marked as scoring is taken to
	// have lost its run (a restart, say) and is started again. Marks already
	// saved are kept, so a rerun only rates what is left.
	scoringStaleAfter = 3 * time.Minute
	// ratingAttempts is how many times one item's rating is tried before it is
	// left unscored.
	ratingAttempts = 2
)

// Reasons an item has no mark.
const reasonUnavailable = "Rating was unavailable, so this item is left out of your scores."

// stale reports whether a paper has been scoring for too long to still be
// running.
func (s *Service) stale(session sessionRow) bool {
	return session.Status == StatusScoring && session.ScoringStartedAt != nil &&
		s.now().Sub(*session.ScoringStartedAt) > scoringStaleAfter
}

// kick scores a paper in the background, unless this process already is.
func (s *Service) kick(user models.User, id string) {
	if s.manualScoring {
		return
	}
	if _, running := s.scoring.LoadOrStore(id, true); running {
		return
	}
	go func() {
		defer s.scoring.Delete(id)
		ctx, cancel := context.WithTimeout(context.Background(), scoringTimeout)
		defer cancel()
		if err := s.ScorePaper(ctx, user, id); err != nil && s.log != nil {
			s.log.Error("pte mock scoring failed", "session", id, "err", err)
		}
	}()
}

// ScorePaper marks every item not yet marked, then records the result. It is
// safe to run again on the same paper: marks are saved as they arrive and the
// result is written once.
func (s *Service) ScorePaper(ctx context.Context, user models.User, id string) error {
	session, err := sessionByID(ctx, s.db, user.ID, id, false)
	if err != nil {
		return err
	}
	if session.Status != StatusScoring {
		return nil
	}
	if _, err := s.db.Exec(ctx, `UPDATE pte_mock_sessions SET scoring_started_at = now()
		WHERE id = $1 AND status = 'scoring'`, id); err != nil {
		return fmt.Errorf("touch pte mock scoring: %w", err)
	}

	items, err := loadItems(ctx, s.db, session.ID)
	if err != nil {
		return err
	}
	ids := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.QuestionID)
	}
	bank, err := s.questions.ByIDs(ctx, ids)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	slots := make(chan struct{}, scoringConcurrency)
	for i := range items {
		it := &items[i]
		if it.Scored || it.UnscoredReason != "" {
			continue
		}
		q, ok := bank[it.QuestionID]
		if !ok {
			it.UnscoredReason = "This question has been removed from the bank."
			if err := saveMark(ctx, s.db, session.ID, *it); err != nil {
				return err
			}
			continue
		}
		if !s.needsRating(*it) {
			s.markLocally(it, q)
			if err := saveMark(ctx, s.db, session.ID, *it); err != nil {
				return err
			}
			continue
		}
		wg.Add(1)
		slots <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-slots }()
			s.rate(ctx, user.ID, session.ID, it, q)
			if err := saveMark(ctx, s.db, session.ID, *it); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if firstErr != nil {
		return firstErr
	}
	return s.complete(ctx, user, session, items)
}

// needsRating reports whether an item goes to an evaluator: a spoken or
// written answer with something in it.
func (s *Service) needsRating(it itemRow) bool {
	task := Tasks[it.Task]
	switch task.Marking {
	case MarkSpoken:
		return s.evaluator != nil && (strings.TrimSpace(it.Transcript) != "" || it.Duration > 0)
	case MarkWritten:
		if s.evaluator == nil || it.Response == nil {
			return false
		}
		words := len(strings.Fields(it.Response.TextResponse))
		return words > 0 && words >= task.MinWords && (task.MaxWords == 0 || words <= task.MaxWords)
	default:
		return false
	}
}

// markLocally marks an item without an evaluator: against its answer key, or
// at zero for a response with nothing in it or outside the task's form.
func (s *Service) markLocally(it *itemRow, q models.Question) {
	task := Tasks[it.Task]
	zero := func(feedback string) {
		score, max := 0.0, 1.0
		it.Scored, it.Score, it.MaxScore, it.Feedback = true, &score, &max, feedback
	}

	switch task.Marking {
	case MarkSpoken:
		switch {
		case s.evaluator == nil:
			it.UnscoredReason = reasonUnavailable
		default:
			zero("Nothing was recorded for this item, so it scores zero.")
		}
	case MarkWritten:
		words := 0
		if it.Response != nil {
			words = len(strings.Fields(it.Response.TextResponse))
		}
		switch {
		case words == 0:
			zero("Nothing was written for this item, so it scores zero.")
		case words < task.MinWords || (task.MaxWords > 0 && words > task.MaxWords):
			zero(fmt.Sprintf("Your response has %d words. %s must be %d to %d words; outside that range PTE scores it zero on every trait.",
				words, task.Name, task.MinWords, task.MaxWords))
		default:
			it.UnscoredReason = reasonUnavailable
		}
	default:
		sub := models.AnswerSubmission{QuestionID: q.ID, Exam: models.ExamPTE}
		if it.Response != nil {
			sub = *it.Response
			sub.QuestionID, sub.Exam = q.ID, models.ExamPTE
		}
		graded, ok := scoring.Grade(q, sub)
		if !ok {
			it.UnscoredReason = "This item has no answer key to mark against."
			return
		}
		score, max := graded.Score, graded.MaxScore
		if max <= 0 {
			max = 1
		}
		it.Scored, it.Score, it.MaxScore = true, &score, &max
		it.Feedback, it.CorrectDisplay, it.UserDisplay = graded.Feedback, graded.CorrectDisplay, graded.UserDisplay
		if !answered(it.Response) {
			it.Feedback = "Not answered."
		}
	}
}

// rate sends a spoken or written answer to its evaluator, trying again once
// before leaving the item unscored.
func (s *Service) rate(ctx context.Context, userID, sessionID string, it *itemRow, q models.Question) {
	var evaluation models.Evaluation
	var err error
	for attempt := 0; attempt < ratingAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
		if ctx.Err() != nil {
			err = ctx.Err()
			break
		}
		if Tasks[it.Task].Marking == MarkSpoken {
			delivery := scoring.Delivery{WordsSpoken: len(strings.Fields(it.Transcript))}
			if d := it.Delivery; d != nil {
				delivery.DurationSeconds, delivery.PauseCount = d.DurationSeconds, d.PauseCount
				delivery.LongestPauseSeconds, delivery.SpeakingRatio = d.LongestPauseSeconds, d.SpeakingRatio
			}
			evaluation, err = s.evaluator.EvaluatePTEMockSpeaking(ctx, userID, q, it.Transcript,
				max(it.Duration, 1), delivery, sessionID)
		} else {
			evaluation, err = s.evaluator.EvaluatePTEMockWriting(ctx, userID, q, it.Response.TextResponse, sessionID)
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		if s.log != nil {
			s.log.Warn("pte mock item not rated", "session", sessionID, "position", it.Position, "err", err)
		}
		it.UnscoredReason = reasonUnavailable
		return
	}
	fraction, ok := FractionOfEvaluation(evaluation)
	if !ok {
		it.UnscoredReason = reasonUnavailable
		return
	}
	score, max := fraction, 1.0
	it.Scored, it.Score, it.MaxScore = true, &score, &max
	it.Feedback = evaluation.Summary
	it.EvaluationID = evaluation.ID
}

// complete aggregates the marks and records the result, once.
func (s *Service) complete(ctx context.Context, user models.User, session sessionRow, items []itemRow) error {
	blueprint, ok := BlueprintFor(session.Kind)
	if !ok {
		return ErrUnknownKind
	}
	marks := make([]Mark, 0, len(items))
	answeredCount, keyCorrect, keyTotal := 0, 0, 0
	answers := make([]models.AnswerSubmission, 0, len(items))
	for _, it := range items {
		m := Mark{Task: it.Task, Scored: it.Scored}
		if it.Score != nil && it.MaxScore != nil {
			m.Score, m.MaxScore = *it.Score, *it.MaxScore
		}
		marks = append(marks, m)
		if it.Status == ItemAnswered {
			answeredCount++
		}
		if Tasks[it.Task].Marking == MarkKey && it.Scored {
			keyTotal++
			if m.MaxScore > 0 && m.Score >= m.MaxScore {
				keyCorrect++
			}
		}
		switch {
		case it.Response != nil:
			answers = append(answers, *it.Response)
		case it.Transcript != "":
			answers = append(answers, models.AnswerSubmission{QuestionID: it.QuestionID, Exam: models.ExamPTE, TextResponse: it.Transcript})
		}
	}
	result := Score(blueprint, marks, answeredCount)

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin pte mock result: %w", err)
	}
	defer tx.Rollback(ctx)

	// Under the paper's lock, so two runs finishing together record it once.
	locked, err := sessionByID(ctx, tx, user.ID, session.ID, true)
	if err != nil {
		return err
	}
	if locked.Status != StatusScoring {
		return nil
	}

	var attemptID *string
	if headline, ok := result.Headline(blueprint); ok {
		skillScores := make(map[models.SkillType]float64, len(result.Skills))
		for skill, score := range result.Skills {
			skillScores[skill] = float64(score)
		}
		attempt, err := s.mocks.SaveAttempt(ctx, tx, mocks.SaveAttemptParams{
			UserID:          user.ID,
			MockID:          blueprint.MockID,
			ExamVersionID:   session.ExamVersionID,
			Exam:            models.ExamPTE,
			UserScore:       headline,
			SkillScores:     skillScores,
			TotalCorrect:    keyCorrect,
			TotalQuestions:  keyTotal,
			DurationSeconds: max(1, int(s.now().Sub(session.CreatedAt).Seconds())),
			Answers:         answers,
		})
		if err != nil {
			return err
		}
		attemptID = &attempt.ID

		if s.xp != nil && answeredCount > 0 {
			if _, err := s.xp.Award(ctx, tx, gamification.AwardParams{
				UserID:    user.ID,
				Amount:    gamification.XPMockCompleted,
				Reason:    "Completed " + blueprint.Title,
				SourceKey: "pte-mock:" + session.ID,
			}); err != nil {
				return err
			}
			if _, err := s.xp.TouchStreak(ctx, tx, user); err != nil {
				return err
			}
		}
	}

	resultJSON, err := jsonText(result)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE pte_mock_sessions
		   SET status = 'completed', completed_at = now(), result = $2::jsonb, mock_attempt_id = $3
		 WHERE id = $1 AND status = 'scoring'`, session.ID, resultJSON, attemptID); err != nil {
		return fmt.Errorf("record pte mock result: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit pte mock result: %w", err)
	}
	return nil
}
