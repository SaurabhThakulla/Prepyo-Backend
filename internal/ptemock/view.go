package ptemock

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/models"
)

// PartView is one part of a paper and where its items sit.
type PartView struct {
	Part  Part   `json:"part"`
	Label string `json:"label"`
	First int    `json:"first"`
	Last  int    `json:"last"`
	Items int    `json:"items"`
}

// PartLabels are the parts' names on screen.
var PartLabels = map[Part]string{
	PartSpeakingWriting: "Speaking & Writing",
	PartReading:         "Reading",
	PartListening:       "Listening",
}

// ClockView is the countdown the item on screen runs on.
type ClockView struct {
	Key string `json:"key"`
	// Scope is "item" for an item's own clock, "part" for its part's.
	Scope            string    `json:"scope"`
	TotalSeconds     int       `json:"totalSeconds"`
	SecondsRemaining int       `json:"secondsRemaining"`
	Deadline         time.Time `json:"deadline"`
}

// ItemView is the item on screen, with what it needs and nothing it must not
// show: no answer key, and a recording's script only where the browser speaks
// it (see models.Question.PublicQuestion).
type ItemView struct {
	Position int             `json:"position"`
	Task     Task            `json:"task"`
	Part     Part            `json:"part"`
	Question models.Question `json:"question"`
	// Passage is the text a reading question is asked about, when its set
	// shows one. Resources is the set's own material (a word bank).
	Passage      *models.ReadingPassage    `json:"passage,omitempty"`
	Resources    []models.ReadingParagraph `json:"resources,omitempty"`
	Draft        *models.AnswerSubmission  `json:"draft,omitempty"`
	PartPosition int                       `json:"partPosition"`
	PartItems    int                       `json:"partItems"`
}

// View is a paper as the learner sees it.
type View struct {
	ID          string     `json:"id"`
	Kind        Kind       `json:"kind"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	Current     int        `json:"current"`
	TotalItems  int        `json:"totalItems"`
	Answered    int        `json:"answered"`
	Parts       []PartView `json:"parts"`
	Item        *ItemView  `json:"item,omitempty"`
	Clock       *ClockView `json:"clock,omitempty"`
	Missing     []string   `json:"missing"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	// Scored counts items marked so far, while the paper is being scored.
	Scored int     `json:"scored"`
	Result *Result `json:"result,omitempty"`
}

func (s *Service) view(ctx context.Context, session sessionRow, items []itemRow) (View, error) {
	blueprint, _ := BlueprintFor(session.Kind)
	v := View{
		ID: session.ID, Kind: session.Kind, Title: blueprint.Title, Status: session.Status,
		Current: session.Current, TotalItems: session.Total, Missing: session.Missing,
		CreatedAt: session.CreatedAt, CompletedAt: session.CompletedAt, Result: session.Result,
	}
	if v.Missing == nil {
		v.Missing = []string{}
	}
	partIndex := map[Part]int{}
	for _, it := range items {
		if it.Status == ItemAnswered {
			v.Answered++
		}
		if it.Scored || it.UnscoredReason != "" {
			v.Scored++
		}
		i, ok := partIndex[it.Part]
		if !ok {
			i = len(v.Parts)
			partIndex[it.Part] = i
			v.Parts = append(v.Parts, PartView{Part: it.Part, Label: PartLabels[it.Part], First: it.Position})
		}
		v.Parts[i].Last = it.Position
		v.Parts[i].Items++
	}

	current := itemIndex(items, session.Current)
	if session.Status != StatusInProgress || current < 0 {
		return v, nil
	}
	it := items[current]
	item, err := s.itemView(ctx, it)
	if err != nil {
		return View{}, err
	}
	part := v.Parts[partIndex[it.Part]]
	item.PartPosition = it.Position - part.First + 1
	item.PartItems = part.Items
	v.Item = &item

	if key := clockKey(it); key != "" {
		if deadline, ok := session.Deadlines[key]; ok {
			scope := "part"
			if Tasks[it.Task].Clock == ClockItem {
				scope = "item"
			}
			remaining := int(math.Ceil(deadline.Sub(s.now()).Seconds()))
			v.Clock = &ClockView{Key: key, Scope: scope, TotalSeconds: clockSeconds(key, items),
				SecondsRemaining: max(remaining, 0), Deadline: deadline}
		}
	}
	return v, nil
}

// itemView loads the question behind an item and what its screen shows.
func (s *Service) itemView(ctx context.Context, it itemRow) (ItemView, error) {
	bank, err := s.questions.ByIDs(ctx, []string{it.QuestionID})
	if err != nil {
		return ItemView{}, err
	}
	q, ok := bank[it.QuestionID]
	if !ok {
		return ItemView{}, fmt.Errorf("pte mock item %d: question %s is gone", it.Position, it.QuestionID)
	}
	item := ItemView{Position: it.Position, Task: Tasks[it.Task], Part: it.Part, Question: q.PublicQuestion()}
	if len(it.OptionOrder) > 0 {
		item.Question.Options = reorder(item.Question.Options, it.OptionOrder)
	}
	if it.Status == ItemPending && it.Response != nil {
		item.Draft = it.Response
	}

	if q.GroupID != "" && s.reading != nil {
		groups, err := s.reading.GroupsByIDs(ctx, []string{q.GroupID})
		if err != nil {
			return ItemView{}, err
		}
		if group, ok := groups[q.GroupID]; ok {
			item.Resources = group.Resources
			if group.PassageDisplay != "hidden" && group.PassageID != "" {
				passages, err := s.reading.PassagesByIDs(ctx, []string{group.PassageID})
				if err != nil {
					return ItemView{}, err
				}
				if p, ok := passages[group.PassageID]; ok {
					item.Passage = &p
				}
			}
		}
	}
	return item, nil
}

// reorder puts options in a dealt order; any the order does not name follow.
func reorder(options []models.QuestionOption, order []string) []models.QuestionOption {
	byID := make(map[string]models.QuestionOption, len(options))
	for _, o := range options {
		byID[o.ID] = o
	}
	out := make([]models.QuestionOption, 0, len(options))
	for _, id := range order {
		if o, ok := byID[id]; ok {
			out = append(out, o)
			delete(byID, id)
		}
	}
	for _, o := range options {
		if _, left := byID[o.ID]; left {
			out = append(out, o)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Catalog and history
// ---------------------------------------------------------------------------

// CatalogEntry is one kind of mock as it is offered.
type CatalogEntry struct {
	Blueprint
	Parts            []PartView `json:"parts"`
	Tasks            []SlotView `json:"tasks"`
	ItemCount        int        `json:"itemCount"`
	AvailableItems   int        `json:"availableItems"`
	EstimatedMinutes int        `json:"estimatedMinutes"`
	Missing          []string   `json:"missing"`
	Available        bool       `json:"available"`
	Cost             CostView   `json:"cost"`
	Live             *LiveView  `json:"live,omitempty"`
	Latest           *History   `json:"latest,omitempty"`
}

// SlotView is one task of a kind: how many a paper deals, and how many it can.
type SlotView struct {
	Code      string           `json:"code"`
	Name      string           `json:"name"`
	Skill     models.SkillType `json:"skill"`
	Part      Part             `json:"part"`
	Count     int              `json:"count"`
	Available int              `json:"available"`
}

// CostView is what starting a kind spends.
type CostView struct {
	// Kind is "full_mock" (one from the plan's full-mock allowance) or
	// "sub_tests" (Credits of the day's sub-tests).
	Kind    string `json:"kind"`
	Credits int    `json:"credits"`
}

// LiveView is an unfinished paper of a kind, to resume.
type LiveView struct {
	ID         string    `json:"id"`
	Current    int       `json:"current"`
	TotalItems int       `json:"totalItems"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Catalog is every kind of PTE mock, with what the bank can deal of each and
// the learner's open and latest papers.
func (s *Service) Catalog(ctx context.Context, user models.User) ([]CatalogEntry, error) {
	counts := map[string]int{}
	rows, err := s.db.Query(ctx, `
		SELECT skill, type_id, count(*) FROM questions
		 WHERE is_published AND 'PTE' = ANY(supported_exams)
		 GROUP BY skill, type_id`)
	if err != nil {
		return nil, fmt.Errorf("count pte bank: %w", err)
	}
	for rows.Next() {
		var skill, typeID string
		var n int
		if err := rows.Scan(&skill, &typeID, &n); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan pte bank: %w", err)
		}
		counts[skill+"/"+typeID] = n
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	live := map[Kind]LiveView{}
	liveRows, err := s.db.Query(ctx, `
		SELECT kind, id::text, current_position, total_items, created_at FROM pte_mock_sessions
		 WHERE user_id = $1 AND status = 'in_progress' AND created_at >= $2`, user.ID, s.now().Add(-StaleAfter))
	if err != nil {
		return nil, fmt.Errorf("list open pte mocks: %w", err)
	}
	for liveRows.Next() {
		var kind Kind
		var l LiveView
		if err := liveRows.Scan(&kind, &l.ID, &l.Current, &l.TotalItems, &l.CreatedAt); err != nil {
			liveRows.Close()
			return nil, fmt.Errorf("scan open pte mock: %w", err)
		}
		live[kind] = l
	}
	liveRows.Close()
	if err := liveRows.Err(); err != nil {
		return nil, err
	}

	latest := map[Kind]History{}
	history, _, err := listSessions(ctx, s.db, user.ID, 50, 0)
	if err != nil {
		return nil, err
	}
	for _, h := range history {
		if _, seen := latest[h.Kind]; !seen && h.Status == StatusCompleted {
			h.Result = nil
			latest[h.Kind] = h
		}
	}

	entries := make([]CatalogEntry, 0, len(Blueprints))
	for _, bp := range Blueprints {
		entry := CatalogEntry{Blueprint: bp, ItemCount: bp.ItemCount(), EstimatedMinutes: bp.EstimatedMinutes(),
			Missing: []string{}, Available: true}
		perPart := map[Part]int{}
		for _, slot := range bp.Slots {
			task := Tasks[slot.Task]
			have := 0
			for _, id := range task.TypeIDs {
				have += counts[string(task.Skill)+"/"+id]
			}
			view := SlotView{Code: slot.Task, Name: task.Name, Skill: task.Skill, Part: task.Part, Count: slot.Count, Available: min(have, slot.Count)}
			entry.Tasks = append(entry.Tasks, view)
			entry.AvailableItems += view.Available
			perPart[task.Part] += view.Available
			if have == 0 {
				entry.Missing = append(entry.Missing, task.Name)
			}
		}
		for _, part := range bp.Parts() {
			entry.Parts = append(entry.Parts, PartView{Part: part, Label: PartLabels[part], Items: perPart[part]})
			if perPart[part] == 0 {
				entry.Available = false
			}
		}
		if bp.FullAllowance {
			entry.Cost = CostView{Kind: "full_mock", Credits: 1}
		} else {
			entry.Cost = CostView{Kind: "sub_tests", Credits: billing.SectionMockSubTests}
		}
		if l, ok := live[bp.Kind]; ok {
			entry.Live = &l
		}
		if h, ok := latest[bp.Kind]; ok {
			entry.Latest = &h
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// History is the learner's PTE mocks, newest first.
func (s *Service) History(ctx context.Context, user models.User, limit, offset int) ([]History, int, error) {
	return listSessions(ctx, s.db, user.ID, limit, offset)
}

// ---------------------------------------------------------------------------
// Report
// ---------------------------------------------------------------------------

// ReviewItem is one item of a scored paper, with the answer key.
type ReviewItem struct {
	Position    int                      `json:"position"`
	Task        string                   `json:"task"`
	TaskName    string                   `json:"taskName"`
	Part        Part                     `json:"part"`
	Status      string                   `json:"status"`
	Question    models.Question          `json:"question"`
	Explanation string                   `json:"explanation,omitempty"`
	ModelAnswer string                   `json:"modelAnswer,omitempty"`
	Response    *models.AnswerSubmission `json:"response,omitempty"`
	Transcript  string                   `json:"transcript,omitempty"`
	// TranscriptSource says who transcribed a spoken answer: the server
	// (Whisper), or the learner's browser where the server could not.
	TranscriptSource string   `json:"transcriptSource,omitempty"`
	Scored           bool     `json:"scored"`
	Percent          *int     `json:"percent"`
	Score            *float64 `json:"score"`
	MaxScore         *float64 `json:"maxScore"`
	UnscoredReason   string   `json:"unscoredReason,omitempty"`
	Feedback         string   `json:"feedback,omitempty"`
	CorrectDisplay   string   `json:"correctDisplay,omitempty"`
	UserDisplay      string   `json:"userDisplay,omitempty"`
	EvaluationID     string   `json:"evaluationId,omitempty"`
}

// Report is a paper with its scores and every item reviewed. While the paper
// is being scored it carries the progress only; the review comes once it is
// complete, because the review is the answer key.
type Report struct {
	View
	Skills []models.SkillType `json:"skills"`
	Review []ReviewItem       `json:"review"`
}

// Report is a paper's score report.
func (s *Service) Report(ctx context.Context, user models.User, id string) (Report, error) {
	session, err := sessionByID(ctx, s.db, user.ID, id, false)
	if err != nil {
		return Report{}, err
	}
	if session.Status == StatusInProgress {
		// Bring the clocks up to date: a paper whose time ran out while the
		// learner was away is ready to score.
		view, err := s.Get(ctx, user, id)
		if err != nil {
			return Report{}, err
		}
		if view.Status == StatusInProgress {
			return Report{}, ErrNotFinished
		}
		if session, err = sessionByID(ctx, s.db, user.ID, id, false); err != nil {
			return Report{}, err
		}
	}
	if session.Status == StatusAbandoned {
		return Report{}, ErrNotFinished
	}
	if s.stale(session) {
		s.kick(user, session.ID)
	}

	items, err := loadItems(ctx, s.db, session.ID)
	if err != nil {
		return Report{}, err
	}
	view, err := s.view(ctx, session, items)
	if err != nil {
		return Report{}, err
	}
	blueprint, _ := BlueprintFor(session.Kind)
	report := Report{View: view, Skills: blueprint.Skills, Review: []ReviewItem{}}
	if session.Status != StatusCompleted {
		return report, nil
	}

	ids := make([]string, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.QuestionID)
	}
	bank, err := s.questions.ByIDs(ctx, ids)
	if err != nil {
		return Report{}, err
	}
	for _, it := range items {
		q := bank[it.QuestionID]
		public := q.PublicQuestion()
		// The review is the answer key; the script of a recording is part of
		// it, and so is the passage a gap-fill was cut from.
		public.AudioTranscript = q.AudioTranscript
		if len(it.OptionOrder) > 0 {
			public.Options = reorder(public.Options, it.OptionOrder)
		}
		r := ReviewItem{
			Position: it.Position, Task: it.Task, TaskName: Tasks[it.Task].Name, Part: it.Part, Status: it.Status,
			Question: public, Explanation: q.Explanation, ModelAnswer: q.ModelAnswer,
			Response: it.Response, Transcript: it.Transcript, TranscriptSource: it.TranscriptSource, Scored: it.Scored,
			Score: it.Score, MaxScore: it.MaxScore, UnscoredReason: it.UnscoredReason,
			Feedback: it.Feedback, CorrectDisplay: it.CorrectDisplay, UserDisplay: it.UserDisplay,
			EvaluationID: it.EvaluationID,
		}
		if it.Scored && it.Score != nil && it.MaxScore != nil && *it.MaxScore > 0 {
			p := int(math.Round(math.Max(0, math.Min(1, *it.Score / *it.MaxScore)) * 100))
			r.Percent = &p
		}
		report.Review = append(report.Review, r)
	}
	return report, nil
}
