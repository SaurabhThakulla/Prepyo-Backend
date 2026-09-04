package reading

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/scoring"
)

// Reading task types.
//
// These are the ids stored in questions.type_id and reading_question_groups
// .type_id. A type appearing twice on one passage — two separate Yes/No/Not
// Given sets, a sentence completion at the top and a summary completion at the
// bottom — is two groups of the same type, not two types. That keeps the
// practice menu a list of things a learner can practise rather than a list of
// the places a task happens to appear.
const (
	TypeSentenceCompletion  = "reading-sentence-completion"
	TypeTrueFalse           = "reading-true-false"
	TypeFindTheWriter       = "reading-find-the-writer"
	TypeArrangePassage      = "reading-arrange-passage"
	TypeYesNoNotGiven       = "reading-yes-no-not-given"
	TypeMatchingInformation = "reading-matching-information"
)

// MockRequiredTypes is the spread a passage must carry to appear in a generated
// mock. A generated paper promises every task type on every passage, so a
// passage missing one is not eligible — it is not "mostly fine", it would leave
// a learner who sat that paper untested on that type.
var MockRequiredTypes = []string{
	TypeSentenceCompletion,
	TypeTrueFalse,
	TypeFindTheWriter,
	TypeArrangePassage,
	TypeYesNoNotGiven,
	TypeMatchingInformation,
}

// MockPassageCount is how many passages one generated reading paper carries.
const MockPassageCount = 3

var (
	// ErrBankTooSmall means the passage bank cannot fill a paper at all. It is
	// a content problem, not something the learner did, and it is reported
	// rather than papered over with a short mock that would score unfairly.
	ErrBankTooSmall = errors.New("not enough eligible passages for a mock")

	// ErrAlreadySubmitted guards against a double submit grading one paper
	// twice.
	ErrAlreadySubmitted = errors.New("this reading mock has already been submitted")

	// ErrNoAnswers means nothing in the submission belonged to the paper.
	ErrNoAnswers = errors.New("no submitted answers belong to this mock")
)

type Service struct {
	db        *pgxpool.Pool
	repo      *Repository
	questions *questions.Repository
	mocks     *mocks.Repository
	exams     *exams.Repository
	xp        *gamification.Service
	billing   *billing.Service
}

func NewService(
	db *pgxpool.Pool,
	repo *Repository,
	questionRepo *questions.Repository,
	mockRepo *mocks.Repository,
	examRepo *exams.Repository,
	xp *gamification.Service,
	billingService *billing.Service,
) *Service {
	return &Service{
		db:        db,
		repo:      repo,
		questions: questionRepo,
		mocks:     mockRepo,
		exams:     examRepo,
		xp:        xp,
		billing:   billingService,
	}
}

// ---------------------------------------------------------------------------
// Practice
// ---------------------------------------------------------------------------

type PracticeParams struct {
	Exam models.ExamType
	// TypeID is the task the learner chose to work on.
	TypeID string
	// Limit trims the set to the first questions after shuffling. Zero deals
	// the whole group, which is the normal case.
	Limit int
}

// PracticeSet deals one task set: a passage chosen for this learner, and one
// group of the requested type from it, questions in a random order.
//
// The same question may come back on a later set. That is deliberate — a
// learner working on Matching Information should meet the same passage's
// questions again once the rest of the bank has been through — so nothing here
// tracks questions, only passages.
func (s *Service) PracticeSet(ctx context.Context, user models.User, p PracticeParams) (models.ReadingSet, error) {
	group, err := s.repo.PickPracticeGroup(ctx, user.ID, p.Exam, p.TypeID)
	if err != nil {
		return models.ReadingSet{}, err
	}

	passage, err := s.repo.PassageByID(ctx, group.PassageID)
	if err != nil {
		return models.ReadingSet{}, err
	}

	byGroup, err := s.questions.ByGroupIDs(ctx, []string{group.ID})
	if err != nil {
		return models.ReadingSet{}, err
	}

	built := buildGroup(group, byGroup[group.ID])
	if p.Limit > 0 && p.Limit < len(built.Questions) {
		built.Questions = built.Questions[:p.Limit]
	}

	// Recorded after the set is built, so a passage is never marked as read
	// because of a request that failed before the learner saw anything.
	if err := s.repo.RecordExposure(ctx, s.db, user.ID, []string{passage.ID}, ContextPractice); err != nil {
		return models.ReadingSet{}, err
	}

	return models.ReadingSet{
		Passage:        passage,
		Groups:         []models.ReadingGroup{built},
		TotalQuestions: len(built.Questions),
	}, nil
}

// ---------------------------------------------------------------------------
// Generated mocks
// ---------------------------------------------------------------------------

// StartMock deals a reading paper: three passages this learner has not sat, each
// with every task type, questions shuffled within each set.
//
// A learner who already has a live paper gets that one back. Starting is not
// idempotent in the usual sense — it spends passages — so the second call must
// resume rather than deal again.
func (s *Service) StartMock(ctx context.Context, user models.User, exam models.ExamType) (models.ReadingMockSession, error) {
	blueprint, err := s.repo.GeneratedBlueprint(ctx, exam)
	if err != nil {
		return models.ReadingMockSession{}, err
	}

	if live, err := s.repo.LiveSession(ctx, user.ID, exam); err == nil {
		return s.hydrate(ctx, live)
	} else if !errors.Is(err, ErrSessionNotFound) {
		return models.ReadingMockSession{}, err
	}

	// Checked before any passage is spent. Finding out at submit time that the
	// paper was never allowed would burn three fresh passages on a result the
	// learner cannot keep.
	if s.billing != nil {
		if _, err := s.billing.CheckMockAllowance(ctx, s.db, user); err != nil {
			return models.ReadingMockSession{}, err
		}
	}

	candidates, err := s.repo.PickMockPassages(ctx, user.ID, exam, MockRequiredTypes, MockPassageCount)
	if err != nil {
		return models.ReadingMockSession{}, err
	}
	if len(candidates) < MockPassageCount {
		return models.ReadingMockSession{}, fmt.Errorf("%w: found %d of %d",
			ErrBankTooSmall, len(candidates), MockPassageCount)
	}

	passageIDs := make([]string, len(candidates))
	reused := false
	for i, c := range candidates {
		passageIDs[i] = c.PassageID
		// The bank ran out of passages this learner had not sat. They get one
		// back rather than no mock at all, and the paper says so.
		reused = reused || c.SeenInMock
	}

	sets, err := s.buildSets(ctx, passageIDs, "")
	if err != nil {
		return models.ReadingMockSession{}, err
	}

	questionIDs := []string{}
	for _, set := range sets {
		for _, group := range set.Groups {
			for _, q := range group.Questions {
				questionIDs = append(questionIDs, q.ID)
			}
		}
	}
	if len(questionIDs) == 0 {
		return models.ReadingMockSession{}, ErrBankTooSmall
	}

	// Spending the passages and recording the paper are one unit. Either the
	// learner has a paper and those passages are used up, or neither happened.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.ReadingMockSession{}, fmt.Errorf("begin reading mock: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.RecordExposure(ctx, tx, user.ID, passageIDs, ContextMock); err != nil {
		return models.ReadingMockSession{}, err
	}

	session, err := s.repo.CreateSession(ctx, tx, CreateSessionParams{
		UserID:          user.ID,
		MockID:          blueprint.ID,
		Exam:            exam,
		ExamVersionID:   blueprint.ExamVersionID,
		PassageIDs:      passageIDs,
		QuestionIDs:     questionIDs,
		ReusedPassages:  reused,
		DurationMinutes: blueprint.DurationMinutes,
	})
	if err != nil {
		// Two starts raced. The other one won and its paper stands; this one's
		// passage spend rolls back with the transaction.
		if errors.Is(err, ErrSessionOpen) {
			live, liveErr := s.repo.LiveSession(ctx, user.ID, exam)
			if liveErr != nil {
				return models.ReadingMockSession{}, liveErr
			}
			return s.hydrate(ctx, live)
		}
		return models.ReadingMockSession{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return models.ReadingMockSession{}, fmt.Errorf("commit reading mock: %w", err)
	}

	session.Sets = sets
	return session.ReadingMockSession, nil
}

// ResumeMock returns a paper the learner already holds, with the questions in
// the order they were dealt.
func (s *Service) ResumeMock(ctx context.Context, user models.User, sessionID string) (models.ReadingMockSession, error) {
	session, err := s.repo.SessionByID(ctx, s.db, user.ID, sessionID)
	if err != nil {
		return models.ReadingMockSession{}, err
	}
	return s.hydrate(ctx, session)
}

func (s *Service) AbandonMock(ctx context.Context, user models.User, sessionID string) error {
	session, err := s.repo.SessionByID(ctx, s.db, user.ID, sessionID)
	if err != nil {
		return err
	}
	if session.Status != StatusInProgress {
		return ErrAlreadySubmitted
	}
	// The passages stay spent. The learner has seen them, and abandoning a
	// paper is not a way to get them dealt again.
	_, err = s.repo.CloseSession(ctx, s.db, session.ID, StatusAbandoned, nil)
	return err
}

// MockResult is what a graded reading paper returns.
type MockResult struct {
	Attempt           models.MockAttempt      `json:"attempt"`
	XPAwarded         int                     `json:"xpAwarded"`
	Streak            int                     `json:"streak"`
	ScoredQuestions   int                     `json:"scoredQuestions"`
	UngradedQuestions int                     `json:"ungradedQuestions"`
	ScoreConfidence   string                  `json:"scoreConfidence"`
	Review            []models.ReviewQuestion `json:"review"`
}

// SubmitMock grades a dealt paper and records the result.
//
// Grading runs over the question ids stored when the paper was dealt, never
// over the ids in the request, so extra answers cannot widen the paper. The
// answers themselves go through internal/mocks, which is the same grader the
// fixed blueprints use.
func (s *Service) SubmitMock(
	ctx context.Context,
	user models.User,
	sessionID string,
	answers []models.AnswerSubmission,
	durationSeconds int,
) (MockResult, error) {
	session, err := s.repo.SessionByID(ctx, s.db, user.ID, sessionID)
	if err != nil {
		return MockResult{}, err
	}
	if session.Status != StatusInProgress {
		return MockResult{}, ErrAlreadySubmitted
	}

	bank, err := s.questions.ByIDs(ctx, session.QuestionIDs)
	if err != nil {
		return MockResult{}, err
	}

	graded := mocks.GradeAnswers(bank, answers)
	if graded.Total == 0 {
		return MockResult{}, ErrNoAnswers
	}

	version, err := s.exams.ByID(ctx, session.ExamVersionID)
	if err != nil {
		return MockResult{}, err
	}

	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}
	skillScores := make(map[models.SkillType]float64, len(graded.BySkill))
	for skill, accuracy := range graded.BySkill {
		skillScores[skill] = scale.EstimateFromAccuracy(accuracy)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return MockResult{}, fmt.Errorf("begin reading mock submit: %w", err)
	}
	defer tx.Rollback(ctx)

	attempt, err := s.mocks.SaveAttempt(ctx, tx, mocks.SaveAttemptParams{
		UserID:          user.ID,
		MockID:          session.MockID,
		ExamVersionID:   session.ExamVersionID,
		Exam:            session.Exam,
		UserScore:       scale.EstimateFromAccuracy(graded.Accuracy),
		SkillScores:     skillScores,
		TotalCorrect:    graded.Correct,
		TotalQuestions:  graded.Total,
		DurationSeconds: durationSeconds,
	})
	if err != nil {
		return MockResult{}, err
	}

	// Keyed to the session rather than to the mock and the day, because every
	// generated paper is new work: a learner who sits two reading mocks in one
	// afternoon has done two mocks, not one.
	awarded, err := s.xp.Award(ctx, tx, gamification.AwardParams{
		UserID:    user.ID,
		Amount:    gamification.XPMockCompleted,
		Reason:    "Completed reading mock",
		SourceKey: "reading-mock:" + session.ID,
	})
	if err != nil {
		return MockResult{}, err
	}

	streak, err := s.xp.TouchStreak(ctx, tx, user)
	if err != nil {
		return MockResult{}, err
	}

	// The status guard is inside the update. If another request submitted this
	// paper between the read above and here, nothing is written and the
	// attempt rolls back with the transaction.
	closed, err := s.repo.CloseSession(ctx, tx, session.ID, StatusSubmitted, &attempt.ID)
	if err != nil {
		return MockResult{}, err
	}
	if !closed {
		return MockResult{}, ErrAlreadySubmitted
	}

	if err := tx.Commit(ctx); err != nil {
		return MockResult{}, fmt.Errorf("commit reading mock submit: %w", err)
	}

	return MockResult{
		Attempt:           attempt,
		XPAwarded:         awarded,
		Streak:            streak,
		ScoredQuestions:   graded.Total,
		UngradedQuestions: graded.Ungraded,
		ScoreConfidence:   scoring.Confidence(graded.Total),
		Review:            reviewOf(session.QuestionIDs, bank),
	}, nil
}

// ---------------------------------------------------------------------------
// Composition
// ---------------------------------------------------------------------------

// hydrate fills a stored session with its passages and questions, in the order
// they were dealt. The stored question list is the authority: a paper reopened
// tomorrow is the same paper, not a freshly shuffled one.
func (s *Service) hydrate(ctx context.Context, session Session) (models.ReadingMockSession, error) {
	sets, err := s.buildSets(ctx, session.PassageIDs, "")
	if err != nil {
		return models.ReadingMockSession{}, err
	}

	order := make(map[string]int, len(session.QuestionIDs))
	for i, id := range session.QuestionIDs {
		order[id] = i
	}
	for i := range sets {
		for j := range sets[i].Groups {
			questions := sets[i].Groups[j].Questions
			// A question added to the group after this paper was dealt is not
			// on it, and one removed since is simply gone.
			kept := questions[:0]
			for _, q := range questions {
				if _, ok := order[q.ID]; ok {
					kept = append(kept, q)
				}
			}
			sortByOrder(kept, order)
			sets[i].Groups[j].Questions = kept
		}
	}

	session.Sets = sets
	return session.ReadingMockSession, nil
}

// buildSets loads passages and their groups and assembles them, shuffling each
// group that allows it. An empty typeID takes every group on the passage.
func (s *Service) buildSets(ctx context.Context, passageIDs []string, typeID string) ([]models.ReadingSet, error) {
	passages, err := s.repo.PassagesByIDs(ctx, passageIDs)
	if err != nil {
		return nil, err
	}

	groups, err := s.repo.GroupsForPassages(ctx, passageIDs, typeID)
	if err != nil {
		return nil, err
	}

	groupIDs := make([]string, len(groups))
	for i, g := range groups {
		groupIDs[i] = g.ID
	}
	byGroup, err := s.questions.ByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	built := make(map[string][]models.ReadingGroup, len(passages))
	for _, g := range groups {
		built[g.PassageID] = append(built[g.PassageID], buildGroup(g, byGroup[g.ID]))
	}

	// passageIDs order is the order the paper was dealt in, so it drives the
	// result rather than whatever order the database returned rows in.
	sets := make([]models.ReadingSet, 0, len(passageIDs))
	for _, id := range passageIDs {
		passage, ok := passages[id]
		if !ok {
			continue
		}
		set := models.ReadingSet{Passage: passage, Groups: built[id]}
		for _, g := range set.Groups {
			set.TotalQuestions += len(g.Questions)
		}
		sets = append(sets, set)
	}
	return sets, nil
}

// buildGroup attaches questions to a group, with the answer key stripped and
// the order randomised where the task allows it.
func buildGroup(g Group, list []models.Question) models.ReadingGroup {
	safe := make([]models.Question, len(list))
	for i, q := range list {
		safe[i] = q.PublicQuestion()
	}
	if g.ShuffleQuestions {
		rand.Shuffle(len(safe), func(i, j int) { safe[i], safe[j] = safe[j], safe[i] })
	}

	group := g.ReadingGroup
	group.Questions = safe
	return group
}

// sortByOrder puts questions back into the dealt order. Insertion sort: a group
// holds a handful of questions, and this keeps the dependency at zero.
func sortByOrder(list []models.Question, order map[string]int) {
	for i := 1; i < len(list); i++ {
		for j := i; j > 0 && order[list[j].ID] < order[list[j-1].ID]; j-- {
			list[j], list[j-1] = list[j-1], list[j]
		}
	}
}

// reviewOf returns the paper with its answer key restored, in dealt order, for
// the results screen. Safe here and only here: the paper has been graded and
// closed, so the answers can no longer change what the learner scored.
func reviewOf(questionIDs []string, bank map[string]models.Question) []models.ReviewQuestion {
	review := make([]models.ReviewQuestion, 0, len(questionIDs))
	for _, id := range questionIDs {
		if q, ok := bank[id]; ok {
			review = append(review, q.ForReview())
		}
	}
	return review
}
