package reading

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/mistakes"
	"github.com/prepyo/backend/internal/mockpapers"
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
	TypeMatchHeading        = "reading-arrange-passage"
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

	// ErrNotAttempted means the paper has no answers at all. A blank paper is
	// not a result; it is not recorded and does not set a band.
	ErrNotAttempted = errors.New("no question on this paper has been answered")

	// ErrExpiredUnattempted means time ran out on a paper with no answers.
	// The paper is closed without a result.
	ErrExpiredUnattempted = errors.New("time ran out before any question was answered")

	// ErrPaperClosed means answers arrived for a paper that is no longer open.
	ErrPaperClosed = errors.New("this paper is closed")
)

// SubmitGrace is how long after the deadline a submission is still taken as
// sent: the browser submits when its timer reaches zero, and the request needs
// time to arrive. Past it, the paper is graded from the answers saved in time.
const SubmitGrace = 90 * time.Second

// keepsTextOrder reports whether a task's questions must stay in the order the
// author set. The official format says multiple choice, sentence completion
// and short-answer questions follow the order of the text, and TFNG/YNNG sets
// do in official and published practice papers; summary gaps sit in a printed
// summary. Shuffling any of them would misrepresent the task, whatever the
// group's own flag says.
func keepsTextOrder(typeID string) bool {
	switch typeID {
	case TypeTrueFalse, TypeYesNoNotGiven, TypeSentenceCompletion,
		"reading-summary-completion", "reading-short-answer",
		TypeMCQSingle, TypeMCQMultiple:
		return true
	}
	return false
}

// moduleFor is the IELTS module a learner's reading is drawn from. Modules are
// an IELTS idea; every other exam reads as Academic, which filters nothing.
func moduleFor(exam models.ExamType, user models.User) string {
	if exam == models.ExamIELTS {
		return user.IELTSModule()
	}
	return models.ModuleAcademic
}

// mayShuffle reports whether a group's questions may be dealt in random order.
func mayShuffle(g Group) bool {
	return g.ShuffleQuestions && !keepsTextOrder(g.TypeID)
}

type Service struct {
	db        *pgxpool.Pool
	repo      *Repository
	questions *questions.Repository
	mocks     *mocks.Repository
	exams     *exams.Repository
	xp        *gamification.Service
	billing   *billing.Service
	mistakes  *mistakes.Repository
}

func NewService(
	db *pgxpool.Pool,
	repo *Repository,
	questionRepo *questions.Repository,
	mockRepo *mocks.Repository,
	examRepo *exams.Repository,
	xp *gamification.Service,
	billingService *billing.Service,
	mistakeRepo *mistakes.Repository,
) *Service {
	return &Service{
		db:        db,
		repo:      repo,
		questions: questionRepo,
		mocks:     mockRepo,
		exams:     examRepo,
		xp:        xp,
		billing:   billingService,
		mistakes:  mistakeRepo,
	}
}

// ---------------------------------------------------------------------------
// Practice
// ---------------------------------------------------------------------------

type PracticeParams struct {
	Exam models.ExamType
	// TypeID is the task the learner chose to work on.
	TypeID string
	// TypeIDs is the whole family behind that choice. A dashboard heading can
	// stand for more than one server type — Multiple Choice is one heading and
	// two types — and practice deals all of them together rather than making
	// the learner meet half the questions and never see the rest.
	TypeIDs []string
	// Limit trims the set to the first questions after shuffling. Zero deals
	// the whole group, which is the normal case.
	Limit int
}

// PracticeSet deals every set of the requested types that one passage carries.
//
// A mock is composed slot by slot and takes a fixed count from each type, which
// is why it builds its own sets. Practice is the opposite: the learner picked a
// task and wants all of it that this passage has.
func (s *Service) PracticeSet(ctx context.Context, user models.User, p PracticeParams) (models.ReadingSet, error) {
	types := p.TypeIDs
	if len(types) == 0 && p.TypeID != "" {
		types = []string{p.TypeID}
	}
	for _, id := range types {
		if id == TypeReorderParagraphs {
			return s.practiceReorder(ctx, user, p.Exam)
		}
	}

	anchor, err := s.repo.PickPracticeGroup(ctx, user.ID, p.Exam, moduleFor(p.Exam, user), types)
	if err != nil {
		return models.ReadingSet{}, err
	}

	sets, err := s.buildSets(ctx, []string{anchor.PassageID}, types, p.Exam)
	if err != nil {
		return models.ReadingSet{}, err
	}
	if len(sets) == 0 || len(sets[0].Groups) == 0 {
		return models.ReadingSet{}, ErrNoPassage
	}
	set := sets[0]

	// A limit trims the sitting as a whole, so a capped set still reads as
	// consecutive questions rather than the first few of every group.
	if p.Limit > 0 && p.Limit < set.TotalQuestions {
		remaining := p.Limit
		kept := make([]models.ReadingGroup, 0, len(set.Groups))
		for _, group := range set.Groups {
			if remaining <= 0 {
				break
			}
			if len(group.Questions) > remaining {
				group.Questions = group.Questions[:remaining]
			}
			remaining -= len(group.Questions)
			kept = append(kept, group)
		}
		set.Groups = kept
		set.TotalQuestions = p.Limit
	}

	if err := s.repo.RecordExposure(ctx, s.db, user.ID, p.Exam, []string{anchor.PassageID}, ContextPractice); err != nil {
		return models.ReadingSet{}, err
	}

	return set, nil
}

// practiceReorder deals one Re-order Paragraphs item as a set with no passage.
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

// StartMock deals or resumes a reading paper for the learner.
func (s *Service) StartMock(ctx context.Context, user models.User, exam models.ExamType, paperID ...string) (models.ReadingMockSession, error) {
	return s.startMock(ctx, user, exam, false, paperID...)
}

// StartMockForFullMock deals a fresh IELTS paper for a full mock's Reading
// section. It is not charged (the full mock was), and it is kept apart from
// any reading mock the learner has open.
func (s *Service) StartMockForFullMock(ctx context.Context, user models.User) (models.ReadingMockSession, error) {
	return s.startMock(ctx, user, models.ExamIELTS, true)
}

func (s *Service) startMock(ctx context.Context, user models.User, exam models.ExamType, fullMock bool, paperID ...string) (models.ReadingMockSession, error) {
	charge := !fullMock && s.billing != nil
	module := moduleFor(exam, user)
	blueprint, err := s.repo.GeneratedBlueprint(ctx, exam, module)
	if err != nil {
		return models.ReadingMockSession{}, err
	}

	if !fullMock {
		if live, err := s.repo.LiveSession(ctx, user.ID, exam); err == nil {
			requested := ""
			if len(paperID) > 0 {
				requested = paperID[0]
			}
			if err := mockpapers.CheckOpenPaper(requested, live.PaperID); err != nil {
				return models.ReadingMockSession{}, err
			}
			return s.hydrate(ctx, live)
		} else if !errors.Is(err, ErrSessionNotFound) {
			return models.ReadingMockSession{}, err
		}
	}

	// A section mock is paid for in sub-tests, not from the full-mock
	// allowance. This early check only saves composing a paper that cannot be
	// started; the binding check runs under the user lock below.
	if charge {
		if _, err := s.billing.CheckSubTestCredits(ctx, s.db, user, billing.SectionMockSubTests); err != nil {
			return models.ReadingMockSession{}, err
		}
	}

	var passageIDs, questionIDs, reorderIDs []string
	var reused bool
	var resolvedPaperID *string

	// Numbered papers here are IELTS Reading's; PTE Reading tests are the PTE
	// mock engine's (ptemock) and never dealt through this service.
	if len(paperID) > 0 && strings.TrimSpace(paperID[0]) != "" {
		reqPaperID := strings.TrimSpace(paperID[0])
		if exam != models.ExamIELTS {
			return models.ReadingMockSession{}, mockpapers.ErrPaperNotFound
		}
		var paperStatus, paperModule string
		var rawContent []byte
		err := s.db.QueryRow(ctx, `
			SELECT status, module, content
			  FROM mock_papers
			 WHERE id = $1 AND exam = 'ielts' AND section = 'reading'`, reqPaperID).Scan(&paperStatus, &paperModule, &rawContent)
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ReadingMockSession{}, mockpapers.ErrPaperNotFound
		}
		if err != nil {
			return models.ReadingMockSession{}, fmt.Errorf("load reading mock paper: %w", err)
		}
		if paperStatus == mockpapers.StatusRetired {
			return models.ReadingMockSession{}, mockpapers.ErrPaperRetired
		}
		if paperStatus != mockpapers.StatusPublished {
			return models.ReadingMockSession{}, mockpapers.ErrPaperNotPublished
		}
		if err := mockpapers.CheckModule(paperModule, module); err != nil {
			return models.ReadingMockSession{}, err
		}

		var c struct {
			PassageIDs  []string `json:"passageIds"`
			QuestionIDs []string `json:"questionIds"`
			ReorderIDs  []string `json:"reorderIds"`
		}
		if err := json.Unmarshal(rawContent, &c); err != nil {
			return models.ReadingMockSession{}, fmt.Errorf("decode paper content: %w", err)
		}
		passageIDs, questionIDs, reorderIDs = c.PassageIDs, c.QuestionIDs, c.ReorderIDs
		resolvedPaperID = &reqPaperID
	} else if exam == models.ExamIELTS {
		// Use lowest-numbered published test the learner has not completed
		var numPaperID string
		var rawContent []byte
		err := s.db.QueryRow(ctx, `
			SELECT p.id, p.content
			  FROM mock_papers p
			 WHERE p.exam = $1 AND p.section = 'reading' AND p.module = $2 AND p.status = 'published'
			 ORDER BY (SELECT count(*) FROM reading_mock_sessions r
			            WHERE r.user_id = $3 AND r.paper_id = p.id AND r.status = 'submitted') ASC,
			          p.number ASC
			 LIMIT 1`, strings.ToLower(string(exam)), strings.ToLower(module), user.ID).Scan(&numPaperID, &rawContent)
		if err == nil && len(rawContent) > 0 {
			var c struct {
				PassageIDs  []string `json:"passageIds"`
				QuestionIDs []string `json:"questionIds"`
				ReorderIDs  []string `json:"reorderIds"`
			}
			if err := json.Unmarshal(rawContent, &c); err == nil && len(c.QuestionIDs) > 0 {
				passageIDs, questionIDs, reorderIDs = c.PassageIDs, c.QuestionIDs, c.ReorderIDs
				resolvedPaperID = &numPaperID
			}
		}

		if len(questionIDs) == 0 {
			composed, err := s.compose(ctx, user.ID, exam, blueprint)
			if err != nil {
				return models.ReadingMockSession{}, err
			}
			passageIDs, questionIDs, reorderIDs, reused = composed.PassageIDs, composed.QuestionIDs, composed.ReorderIDs, composed.Reused
		}
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.ReadingMockSession{}, fmt.Errorf("begin reading mock: %w", err)
	}
	defer tx.Rollback(ctx)

	// Serialise with every other credit spend for this learner, so two starts
	// racing each other cannot both see enough credit left.
	if charge {
		if err := billing.LockUserForQuota(ctx, tx, user.ID); err != nil {
			return models.ReadingMockSession{}, err
		}
		if _, err := s.billing.CheckSubTestCredits(ctx, tx, user, billing.SectionMockSubTests); err != nil {
			return models.ReadingMockSession{}, err
		}
	}

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
		InFullMock:      fullMock,
		PaperID:         resolvedPaperID,
	})
	if err != nil {
		if errors.Is(err, ErrSessionOpen) {
			live, liveErr := s.repo.LiveSession(ctx, user.ID, exam)
			if liveErr != nil {
				return models.ReadingMockSession{}, liveErr
			}
			return s.hydrate(ctx, live)
		}
		return models.ReadingMockSession{}, err
	}

	// Charged in the same transaction as the paper, so a failed start costs
	// nothing and a resumed paper (returned above, before this) is never
	// charged twice.
	if charge {
		if _, err := s.billing.RecordSessionStartCredits(ctx, tx, user, string(exam), string(models.SkillReading),
			"mock:"+session.ID, billing.SectionMockSubTests); err != nil {
			return models.ReadingMockSession{}, err
		}
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

// compose builds a concrete paper from a blueprint for a learner.
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
		out.Reused = out.Reused || candidate.SeenInMock
	}

	if len(out.QuestionIDs) == 0 {
		return composition{}, ErrBankTooSmall
	}
	return out, nil
}

// assignPassages finds a valid assignment of distinct passages to slots using backtracking.
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
		candidates, err := s.repo.SlotPassageCandidates(ctx, userID, exam, blueprint.Module, slot.PassageTag, slot.TypeIDs(), slot.Counts())
		if err != nil {
			return nil, err
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("%w: no passage can fill section %d (%v)",
				ErrBankTooSmall, slot.Position, slot.TypeIDs())
		}
		sections = append(sections, section{position: slot.Position, candidates: candidates})
	}

	// Sort most constrained slots first to prune the search space quickly.
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

// slotQuestions selects the required question counts per task from a passage.
func (s *Service) slotQuestions(
	ctx context.Context,
	passageID string,
	exam models.ExamType,
	slot BlueprintSlot,
) ([]string, error) {
	groups, err := s.repo.GroupsForPassages(ctx, []string{passageID}, nil)
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
			if mayShuffle(g) {
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

// SaveDrafts records the answers given so far on an open paper. Answers for
// questions that are not on the paper are dropped rather than stored.
func (s *Service) SaveDrafts(ctx context.Context, user models.User, sessionID string, answers []models.AnswerSubmission) (models.ReadingMockSession, error) {
	session, err := s.repo.SessionByID(ctx, s.db, user.ID, sessionID)
	if err != nil {
		return models.ReadingMockSession{}, err
	}
	if session.Status != StatusInProgress {
		return models.ReadingMockSession{}, ErrPaperClosed
	}

	onPaper := make(map[string]bool, len(session.QuestionIDs))
	for _, id := range session.QuestionIDs {
		onPaper[id] = true
	}
	kept := make([]models.AnswerSubmission, 0, len(answers))
	seen := make(map[string]bool, len(answers))
	for _, answer := range answers {
		if !onPaper[answer.QuestionID] || seen[answer.QuestionID] {
			continue
		}
		seen[answer.QuestionID] = true
		kept = append(kept, answer)
	}

	saved, err := s.repo.SaveDrafts(ctx, user.ID, sessionID, kept, SubmitGrace)
	if err != nil {
		return models.ReadingMockSession{}, err
	}
	if !saved {
		return models.ReadingMockSession{}, ErrPaperClosed
	}

	fresh, err := s.repo.SessionByID(ctx, s.db, user.ID, sessionID)
	if err != nil {
		return models.ReadingMockSession{}, err
	}
	return models.ReadingMockSession{
		ID:               fresh.ID,
		Status:           fresh.Status,
		ExpiresAt:        fresh.ExpiresAt,
		SecondsRemaining: fresh.SecondsRemaining,
	}, nil
}

// ResumeMock returns an in-progress paper for the learner.
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

	// Time is the server's. A submission that arrives after the deadline and
	// its grace period is graded from the answers saved while time remained,
	// not from answers the browser may have kept collecting afterwards.
	late := time.Now().After(session.ExpiresAt.Add(SubmitGrace))
	if late {
		answers = session.DraftAnswers
	}

	elapsed := int(time.Since(session.CreatedAt).Seconds())
	if limit := session.DurationMinutes * 60; elapsed > limit {
		elapsed = limit
	}
	if elapsed < 1 {
		elapsed = 1
	}
	durationSeconds = elapsed

	bank, err := s.questions.ByIDs(ctx, session.QuestionIDs)
	if err != nil {
		return MockResult{}, err
	}

	graded, err := mocks.GradeAnswersFor(session.Exam, bank, session.QuestionIDs, answers)
	if errors.Is(err, mocks.ErrInvalidAnswers) {
		return MockResult{}, ErrNoAnswers
	}
	if err != nil {
		return MockResult{}, err
	}
	if graded.Total == 0 {
		return MockResult{}, ErrNoAnswers
	}
	if graded.Answered == 0 {
		if late || time.Now().After(session.ExpiresAt) {
			if _, err := s.repo.CloseSession(ctx, s.db, session.ID, StatusAbandoned, nil); err != nil {
				return MockResult{}, err
			}
			return MockResult{}, ErrExpiredUnattempted
		}
		return MockResult{}, ErrNotAttempted
	}

	version, err := s.exams.ByID(ctx, session.ExamVersionID)
	if err != nil {
		return MockResult{}, err
	}

	scale := scoring.Scale{Min: version.MinScore, Max: version.MaxScore, Step: version.ScoreStep}
	// IELTS converts marks through the indicative reading table for the
	// learner's module, scaled to 40 when the paper is another size; PTE uses
	// its own scale. Raw is marks under IELTS and correct items under PTE.
	// The paper's own module decides the table, not whichever module the
	// learner has chosen since it was dealt.
	module, err := s.repo.ModuleForMock(ctx, session.MockID)
	if err != nil {
		return MockResult{}, err
	}
	userScore := scale.EstimateForModule(string(session.Exam), module, "reading", graded.Raw, graded.RawMax)

	skillScores := make(map[models.SkillType]float64, len(graded.BySkill))
	for skill := range graded.BySkill {
		skillScores[skill] = userScore
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
		UserScore:       userScore,
		SkillScores:     skillScores,
		TotalCorrect:    graded.Raw,
		TotalQuestions:  graded.RawMax,
		DurationSeconds: durationSeconds,
		Answers:         answers,
	})
	if err != nil {
		return MockResult{}, err
	}

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

	if s.mistakes != nil {
		for _, answer := range answers {
			question, ok := bank[answer.QuestionID]
			if !ok {
				continue
			}
			if !scoring.Deterministic(question.Skill) {
				continue
			}
			gradedRes, ok := scoring.Grade(question, answer)
			if !ok || gradedRes.IsCorrect {
				continue
			}
			_ = s.mistakes.Record(ctx, tx, mistakes.RecordParams{
				UserID:          user.ID,
				QuestionID:      question.ID,
				Exam:            session.Exam,
				ErrorTag:        gradedRes.ErrorTag,
				UserResponse:    gradedRes.UserDisplay,
				CorrectResponse: gradedRes.CorrectDisplay,
				Explanation:     question.Explanation,
			})
		}
	}

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

// hydrate rebuilds a stored paper session from its dealt question IDs.
func (s *Service) hydrate(ctx context.Context, session Session) (models.ReadingMockSession, error) {
	bank, err := s.questions.ByIDs(ctx, session.QuestionIDs)
	if err != nil {
		return models.ReadingMockSession{}, err
	}

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

// buildSets loads passages and groups for display, filtering by type and exam.
func (s *Service) buildSets(ctx context.Context, passageIDs []string, typeIDs []string, exam models.ExamType) ([]models.ReadingSet, error) {
	passages, err := s.repo.PassagesByIDs(ctx, passageIDs)
	if err != nil {
		return nil, err
	}

	groups, err := s.repo.GroupsForPassages(ctx, passageIDs, typeIDs)
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
		if len(group.Questions) == 0 {
			continue
		}
		built[g.PassageID] = append(built[g.PassageID], group)
	}

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

// buildGroup attaches questions to a group, filtering by exam eligibility and stripping answers.
func buildGroup(g Group, list []models.Question, exam models.ExamType) models.ReadingGroup {
	safe := make([]models.Question, 0, len(list))
	for _, q := range list {
		if exam != "" && !q.SupportsExam(exam) {
			continue
		}
		safe = append(safe, q.PublicQuestion())
	}
	if mayShuffle(g) {
		rand.Shuffle(len(safe), func(i, j int) { safe[i], safe[j] = safe[j], safe[i] })
	}

	group := g.ReadingGroup
	group.Questions = safe
	return group
}

// reviewOf returns the paper questions with answer keys restored for review.
func reviewOf(questionIDs []string, bank map[string]models.Question) []models.ReviewQuestion {
	review := make([]models.ReviewQuestion, 0, len(questionIDs))
	for _, id := range questionIDs {
		if q, ok := bank[id]; ok {
			review = append(review, q.ForReview())
		}
	}
	return review
}

// ComposeForBuilder composes an IELTS reading paper without user history,
// preferring passages and items least used in published mock papers.
func (s *Service) ComposeForBuilder(ctx context.Context, module string) ([]string, []string, []string, error) {
	blueprint, err := s.repo.GeneratedBlueprint(ctx, models.ExamIELTS, module)
	if err != nil {
		return nil, nil, nil, err
	}
	comp, err := s.compose(ctx, "", models.ExamIELTS, blueprint)
	if err != nil {
		return nil, nil, nil, err
	}
	return comp.PassageIDs, comp.QuestionIDs, comp.ReorderIDs, nil
}
