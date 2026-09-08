package reading

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"

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
	TypeFindTheParagraph    = "reading-find-the-paragraph"
	TypeArrangePassage      = "reading-arrange-passage"
	TypeYesNoNotGiven       = "reading-yes-no-not-given"
	TypeMatchingInformation = "reading-matching-information"
)

// PTE reading task types.
//
// The two fill-in-the-blanks ids are the ones the standalone seed has always
// used, promoted from string literals so the grader, the seed and this package
// cannot drift apart. Both gap-fills reach scoring.gradeBlanks and both MCQ
// types reach scoring.gradeChoice, so none of them needed a grader written.
const (
	TypeFillBlanksRW      = "fill-in-blanks-rw"
	TypeFillBlanksR       = "fill-in-blanks-r"
	TypeMCQSingle         = "reading-mcq-single"
	TypeMCQMultiple       = "reading-mcq-multiple"
	TypeReorderParagraphs = "reorder-paragraphs"
)

// The shape of a paper is data, not a constant. It lives in
// reading_mock_blueprints and reading_mock_blueprint_slots, and is read through
// Repository.GeneratedBlueprint.
//
// It used to be a Go var here — three sections of 13/13/14 over six task types —
// and that made three assumptions a shared bank cannot keep: that every paper
// has three passages, that every passage carries all three sections, and that
// there is only one paper shape, which left PTE with no reading mock at all.

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
	// Re-order Paragraphs is not a task on a passage, so it is not dealt like
	// one. Its content lives in its own table and its set comes back without a
	// passage at all.
	if p.TypeID == TypeReorderParagraphs {
		return s.practiceReorder(ctx, user, p.Exam)
	}

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

	built := buildGroup(group, byGroup[group.ID], p.Exam)
	if p.Limit > 0 && p.Limit < len(built.Questions) {
		built.Questions = built.Questions[:p.Limit]
	}

	// Recorded after the set is built, so a passage is never marked as read
	// because of a request that failed before the learner saw anything.
	if err := s.repo.RecordExposure(ctx, s.db, user.ID, p.Exam, []string{passage.ID}, ContextPractice); err != nil {
		return models.ReadingSet{}, err
	}

	return models.ReadingSet{
		Passage:        &passage,
		Groups:         []models.ReadingGroup{built},
		TotalQuestions: len(built.Questions),
	}, nil
}

// practiceReorder deals one Re-order Paragraphs item as a set with no passage.
//
// The boxes are shuffled here rather than stored shuffled, so the same item
// comes back arranged differently every time. The stored order is the answer
// key: it has to be written down once, and this is the only place that reads it
// without handing it over.
func (s *Service) practiceReorder(ctx context.Context, user models.User, exam models.ExamType) (models.ReadingSet, error) {
	item, questionID, err := s.repo.PickReorderItem(ctx, user.ID, exam)
	if err != nil {
		if errors.Is(err, ErrNoReorderItem) {
			return models.ReadingSet{}, ErrNoPassage
		}
		return models.ReadingSet{}, err
	}

	question, err := s.questions.ByID(ctx, questionID)
	if err != nil {
		return models.ReadingSet{}, err
	}

	safe := question.PublicQuestion()
	rand.Shuffle(len(safe.Options), func(i, j int) {
		safe.Options[i], safe.Options[j] = safe.Options[j], safe.Options[i]
	})

	group := models.ReadingGroup{
		ID:               "rg-" + item.ID,
		Position:         1,
		TypeID:           TypeReorderParagraphs,
		TypeName:         "Re-order Paragraphs",
		Instructions:     "The text boxes below have been placed in a random order. Restore the original order.",
		PassageDisplay:   "hidden",
		TimeLimitSeconds: question.TimeLimitSeconds,
		Questions:        []models.Question{safe},
	}

	// Recorded after the set is built, so an item is never marked as dealt
	// because of a request that failed before the learner saw anything.
	if err := s.repo.RecordReorderExposure(ctx, s.db, user.ID, []string{item.ID}, ContextPractice); err != nil {
		return models.ReadingSet{}, err
	}

	return models.ReadingSet{
		Groups:         []models.ReadingGroup{group},
		TotalQuestions: 1,
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

	composed, err := s.compose(ctx, user.ID, exam, blueprint)
	if err != nil {
		return models.ReadingMockSession{}, err
	}
	passageIDs, questionIDs, reorderIDs, reused := composed.PassageIDs, composed.QuestionIDs, composed.ReorderIDs, composed.Reused

	// Spending the passages and recording the paper are one unit. Either the
	// learner has a paper and those passages are used up, or neither happened.
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.ReadingMockSession{}, fmt.Errorf("begin reading mock: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := s.repo.RecordExposure(ctx, tx, user.ID, exam, passageIDs, ContextMock); err != nil {
		return models.ReadingMockSession{}, err
	}

	if err := s.repo.RecordReorderExposure(ctx, tx, user.ID, reorderIDs, ContextMock); err != nil {
		return models.ReadingMockSession{}, err
	}

	session, err := s.repo.CreateSession(ctx, tx, CreateSessionParams{
		UserID:          user.ID,
		MockID:          blueprint.MockID,
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

	return s.hydrate(ctx, session)
}

// composition is one dealt paper before it is written down.
type composition struct {
	PassageIDs  []string
	QuestionIDs []string
	ReorderIDs  []string
	Reused      bool
}

// compose deals a paper: one slot at a time, a passage that can fill it, and
// that slot's questions from it.
//
// A slot is filled from whatever the passage actually carries. It does not
// require the passage to have been authored into a particular shape, which is
// what the old paper_slot column demanded — under that rule a passage without
// all eight task sets on it could never appear in any paper.
//
// The order questions are taken in is the order they are dealt in, and that
// order is what gets frozen into the session. Shuffling happens here, once.
func (s *Service) compose(
	ctx context.Context,
	userID string,
	exam models.ExamType,
	blueprint Blueprint,
) (composition, error) {
	var out composition

	assigned, err := s.assignPassages(ctx, userID, exam, blueprint)
	if err != nil {
		return composition{}, err
	}

	for _, slot := range blueprint.Slots {
		if slot.Source == SourceReorder {
			want := slot.QuestionCount()
			picks, err := s.repo.PickReorderItems(ctx, userID, exam, want)
			if err != nil {
				return composition{}, err
			}
			if len(picks) < want {
				return composition{}, fmt.Errorf("%w: section %d wanted %d re-order items, found %d",
					ErrBankTooSmall, slot.Position, want, len(picks))
			}
			for _, pick := range picks {
				out.ReorderIDs = append(out.ReorderIDs, pick.ItemID)
				out.QuestionIDs = append(out.QuestionIDs, pick.QuestionID)
			}
			continue
		}

		candidate, ok := assigned[slot.Position]
		if !ok {
			return composition{}, fmt.Errorf("%w: no passage can fill section %d (%v)",
				ErrBankTooSmall, slot.Position, slot.TypeIDs())
		}

		ids, err := s.slotQuestions(ctx, candidate.PassageID, exam, slot)
		if err != nil {
			return composition{}, err
		}
		if want := slot.QuestionCount(); len(ids) < want {
			return composition{}, fmt.Errorf("%w: %s gave %d of %d questions for section %d",
				ErrBankTooSmall, candidate.PassageID, len(ids), want, slot.Position)
		}

		out.PassageIDs = append(out.PassageIDs, candidate.PassageID)
		out.QuestionIDs = append(out.QuestionIDs, ids...)
		// The bank ran out of passages this learner had not sat. They get one
		// back rather than no mock at all, and the paper says so.
		out.Reused = out.Reused || candidate.SeenInMock
	}

	if len(out.QuestionIDs) == 0 {
		return composition{}, ErrBankTooSmall
	}
	return out, nil
}

// assignPassages gives each passage-backed section a passage of its own.
//
// It is done for the whole paper at once, not section by section, because the
// sections compete: a passage that could fill three of them can only fill one,
// and taking the best passage for section one can leave section four with
// nothing. That is not hypothetical — a passage carrying no multiple-answer set
// is a perfectly good passage, and requiring every passage to carry every task
// type is exactly the assumption this refactor removes.
//
// So: candidates per section, then the sections with the fewest candidates
// first, then backtracking. Candidate order is the exposure preference, so the
// first assignment that works is also the one practice would have chosen.
func (s *Service) assignPassages(
	ctx context.Context,
	userID string,
	exam models.ExamType,
	blueprint Blueprint,
) (map[int]MockCandidate, error) {
	type section struct {
		position   int
		candidates []MockCandidate
	}

	var sections []section
	for _, slot := range blueprint.Slots {
		if slot.Source != SourcePassage {
			continue
		}
		candidates, err := s.repo.SlotPassageCandidates(ctx, userID, exam, slot.TypeIDs(), slot.Counts())
		if err != nil {
			return nil, err
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("%w: no passage can fill section %d (%v)",
				ErrBankTooSmall, slot.Position, slot.TypeIDs())
		}
		sections = append(sections, section{position: slot.Position, candidates: candidates})
	}

	// Most constrained first. It is not required for correctness — the search
	// backtracks — but it finds the answer far sooner and fails faster when
	// there is not one.
	sort.SliceStable(sections, func(i, j int) bool {
		return len(sections[i].candidates) < len(sections[j].candidates)
	})

	assigned := make(map[int]MockCandidate, len(sections))
	taken := make(map[string]bool, len(sections))

	var solve func(i int) bool
	solve = func(i int) bool {
		if i == len(sections) {
			return true
		}
		for _, candidate := range sections[i].candidates {
			if taken[candidate.PassageID] {
				continue
			}
			taken[candidate.PassageID] = true
			assigned[sections[i].position] = candidate
			if solve(i + 1) {
				return true
			}
			delete(assigned, sections[i].position)
			delete(taken, candidate.PassageID)
		}
		return false
	}

	if !solve(0) {
		return nil, fmt.Errorf("%w: %d sections need %d different passages and the bank cannot supply them",
			ErrBankTooSmall, len(sections), len(sections))
	}
	return assigned, nil
}

// slotQuestions takes one section's worth of questions from one passage: each
// task in turn, its own count, from the task sets of that type on the passage.
//
// Counting per task rather than over the section is what keeps the mix right. A
// section asking for six sentence completions and four matching-information
// questions gets six and four, not ten of whichever set the passage lists first.
func (s *Service) slotQuestions(
	ctx context.Context,
	passageID string,
	exam models.ExamType,
	slot BlueprintSlot,
) ([]string, error) {
	groups, err := s.repo.GroupsForPassages(ctx, []string{passageID}, "")
	if err != nil {
		return nil, err
	}

	byType := map[string][]Group{}
	groupIDs := make([]string, 0, len(groups))
	for _, g := range groups {
		byType[g.TypeID] = append(byType[g.TypeID], g)
		groupIDs = append(groupIDs, g.ID)
	}

	byGroup, err := s.questions.ByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, slot.QuestionCount())
	for _, task := range slot.Tasks {
		taken := 0
		for _, g := range byType[task.TypeID] {
			if taken == task.QuestionCount {
				break
			}

			set := make([]models.Question, 0, len(byGroup[g.ID]))
			for _, q := range byGroup[g.ID] {
				if q.SupportsExam(exam) {
					set = append(set, q)
				}
			}
			if g.ShuffleQuestions {
				rand.Shuffle(len(set), func(i, j int) { set[i], set[j] = set[j], set[i] })
			}

			for _, q := range set {
				if taken == task.QuestionCount {
					break
				}
				ids = append(ids, q.ID)
				taken++
			}
		}
	}
	return ids, nil
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

// hydrate rebuilds a stored paper from the question ids it was dealt.
//
// Those ids are the paper. Everything else — which groups appear, which
// passages, what order they run in — is derived from them, so a paper cannot be
// changed by anything that happens to the bank afterwards. Adding a question to
// a passage does not add it to a paper already dealt; changing a blueprint does
// not re-cut one; only unpublishing a question removes it, and then it is gone
// rather than silently replaced.
//
// It used to rebuild by re-running composition and then filtering the result
// against the stored ids, which is why 000009 had to abandon every live paper
// when the slots moved. This does not have that dependency.
func (s *Service) hydrate(ctx context.Context, session Session) (models.ReadingMockSession, error) {
	bank, err := s.questions.ByIDs(ctx, session.QuestionIDs)
	if err != nil {
		return models.ReadingMockSession{}, err
	}

	// Groups and passages in the order the paper first reaches them, so the
	// rebuilt paper runs in dealt order rather than in id order.
	var groupOrder []string
	seenGroup := map[string]bool{}
	questionsByGroup := map[string][]models.Question{}
	standalone := []models.Question{}

	for _, id := range session.QuestionIDs {
		q, ok := bank[id]
		if !ok {
			continue
		}
		if q.GroupID == "" {
			standalone = append(standalone, q.PublicQuestion())
			continue
		}
		if !seenGroup[q.GroupID] {
			seenGroup[q.GroupID] = true
			groupOrder = append(groupOrder, q.GroupID)
		}
		questionsByGroup[q.GroupID] = append(questionsByGroup[q.GroupID], q.PublicQuestion())
	}

	groups, err := s.repo.GroupsByIDs(ctx, groupOrder)
	if err != nil {
		return models.ReadingMockSession{}, err
	}

	var passageOrder []string
	seenPassage := map[string]bool{}
	groupsByPassage := map[string][]models.ReadingGroup{}

	for _, groupID := range groupOrder {
		g, ok := groups[groupID]
		if !ok {
			continue
		}
		built := g.ReadingGroup
		built.Questions = questionsByGroup[groupID]

		if !seenPassage[g.PassageID] {
			seenPassage[g.PassageID] = true
			passageOrder = append(passageOrder, g.PassageID)
		}
		groupsByPassage[g.PassageID] = append(groupsByPassage[g.PassageID], built)
	}

	passages, err := s.repo.PassagesByIDs(ctx, passageOrder)
	if err != nil {
		return models.ReadingMockSession{}, err
	}

	sets := make([]models.ReadingSet, 0, len(passageOrder)+1)
	for _, id := range passageOrder {
		passage, ok := passages[id]
		if !ok {
			continue
		}
		set := models.ReadingSet{Passage: &passage, Groups: groupsByPassage[id]}
		for _, g := range set.Groups {
			set.TotalQuestions += len(g.Questions)
		}
		sets = append(sets, set)
	}

	// Re-order Paragraphs has no passage to hang from, so its questions come
	// last as a set of their own. The boxes are shuffled per deal, and the
	// shuffle is not stored, so this is where a resumed paper gets them
	// rearranged again — the answer is the sequence, not the arrangement they
	// happen to start in.
	if len(standalone) > 0 {
		sets = append(sets, reorderSet(standalone))
	}

	session.Sets = sets
	return session.ReadingMockSession, nil
}

// reorderSet wraps standalone Re-order Paragraphs questions as one set.
func reorderSet(list []models.Question) models.ReadingSet {
	for i := range list {
		options := list[i].Options
		rand.Shuffle(len(options), func(a, b int) { options[a], options[b] = options[b], options[a] })
	}

	return models.ReadingSet{
		Groups: []models.ReadingGroup{{
			ID:             "rg-reorder",
			Position:       1,
			TypeID:         TypeReorderParagraphs,
			TypeName:       "Re-order Paragraphs",
			Instructions:   "The text boxes below have been placed in a random order. Restore the original order.",
			PassageDisplay: "hidden",
			Questions:      list,
		}},
		TotalQuestions: len(list),
	}
}

// buildSets loads passages and their groups and assembles them, shuffling each
// group that allows it. An empty typeID takes every group on the passage, and an
// empty exam takes every question regardless of which exams set it.
//
// This is the read path for a single passage — /reading/passages/{id} — not the
// composition path. A paper is composed by compose() and rebuilt by hydrate();
// neither goes through here, because a paper is a frozen list of question ids
// and this function answers a different question: what is on this passage.
func (s *Service) buildSets(ctx context.Context, passageIDs []string, typeID string, exam models.ExamType) ([]models.ReadingSet, error) {
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
		group := buildGroup(g, byGroup[g.ID], exam)
		// A task set this exam does not set has no questions left in it, and an
		// empty set is not something to put on screen.
		if len(group.Questions) == 0 {
			continue
		}
		built[g.PassageID] = append(built[g.PassageID], group)
	}

	// passageIDs order is the order the paper was dealt in, so it drives the
	// result rather than whatever order the database returned rows in.
	sets := make([]models.ReadingSet, 0, len(passageIDs))
	for _, id := range passageIDs {
		passage, ok := passages[id]
		if !ok {
			continue
		}
		set := models.ReadingSet{Passage: &passage, Groups: built[id]}
		for _, g := range set.Groups {
			set.TotalQuestions += len(g.Questions)
		}
		sets = append(sets, set)
	}
	return sets, nil
}

// buildGroup attaches questions to a group, with the answer key stripped, the
// questions this exam does not set left out, and the order randomised where the
// task allows it.
//
// The eligibility filter is here rather than in the query because this is the
// last place every path passes through: practice, a dealt paper and a resumed
// one all build their sets from here. A question the exam does not set is not a
// question the learner can be asked, whichever route reached it.
func buildGroup(g Group, list []models.Question, exam models.ExamType) models.ReadingGroup {
	safe := make([]models.Question, 0, len(list))
	for _, q := range list {
		if exam != "" && !q.SupportsExam(exam) {
			continue
		}
		safe = append(safe, q.PublicQuestion())
	}
	if g.ShuffleQuestions {
		rand.Shuffle(len(safe), func(i, j int) { safe[i], safe[j] = safe[j], safe[i] })
	}

	group := g.ReadingGroup
	group.Questions = safe
	return group
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
