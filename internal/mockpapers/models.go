package mockpapers

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	ExamIELTS = "ielts"
	ExamPTE   = "pte"

	SectionSpeaking  = "speaking"
	SectionWriting   = "writing"
	SectionReading   = "reading"
	SectionListening = "listening"
	SectionFull      = "full"

	ModuleAny             = "any"
	ModuleAcademic        = "academic"
	ModuleGeneralTraining = "general_training"

	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusRetired   = "retired"

	// DefaultTargetPapers is how many numbered papers each scope targets. A
	// scope stops short of it when the bank cannot supply another paper
	// within MaxAllowedOverlap, so this is a ceiling the bank grows into.
	// Raised from 10 when bank 4 doubled every sub-task.
	DefaultTargetPapers = 20

	// MaxAllowedOverlap is the maximum question overlap (50%) permitted between papers in a scope.
	MaxAllowedOverlap = 0.50
)

var (
	ErrPaperNotFound     = errors.New("mock paper not found")
	ErrPaperRetired      = errors.New("this test paper has been retired")
	ErrPaperNotPublished = errors.New("mock paper is not published")
	ErrBankTooSmall      = errors.New("question bank too small to build paper")
	ErrOverlapExceeded   = errors.New("cannot build paper without exceeding 50% overlap with existing papers")
	// ErrOtherPaperOpen means the learner asked for one test while another of
	// the same section is still open; it must be finished or left first.
	ErrOtherPaperOpen = errors.New("another test of this section is still open")
	// ErrWrongModule means an IELTS paper is for the other module.
	ErrWrongModule = errors.New("this test is for the other IELTS module")
)

// buildTimeout bounds one background build of a scope.
const buildTimeout = 10 * time.Minute

type Scope struct {
	Exam    string `json:"exam"`
	Section string `json:"section"`
	Module  string `json:"module"`
}

type MockPaper struct {
	ID            string          `json:"id"`
	Exam          string          `json:"exam"`
	Section       string          `json:"section"`
	Module        string          `json:"module"`
	Number        int             `json:"number"`
	Revision      int             `json:"revision"`
	SupersedesID  *string         `json:"supersedesId,omitempty"`
	Status        string          `json:"status"`
	Title         string          `json:"title"`
	ContentSchema int             `json:"contentSchema"`
	Content       json.RawMessage `json:"content"`
	ContentHash   string          `json:"contentHash"`
	CreatedAt     time.Time       `json:"createdAt"`
	PublishedAt   *time.Time      `json:"publishedAt,omitempty"`
	RetiredAt     *time.Time      `json:"retiredAt,omitempty"`
}

// Content shapes for content_schema = 1

type WritingContent struct {
	Task1ID string `json:"task1Id"`
	Task2ID string `json:"task2Id"`
}

type ReadingContent struct {
	PassageIDs  []string `json:"passageIds"`
	QuestionIDs []string `json:"questionIds"`
	ReorderIDs  []string `json:"reorderIds"`
}

type PTEItem struct {
	QuestionID string `json:"questionId"`
	Task       string `json:"task"`
	Part       string `json:"part"`
}

type PTEContent struct {
	Items   []PTEItem `json:"items"`
	Missing []string  `json:"missing"`
}

type ListeningContent struct {
	TestID string `json:"testId"`
}

type SpeakingContent struct {
	SetID string `json:"setId"`
}

// CatalogPaper is what the learner receives in GET /api/v1/mock-papers
type CatalogPaper struct {
	ID                   string  `json:"id"`
	Exam                 string  `json:"exam"`
	Section              string  `json:"section"`
	Module               string  `json:"module"`
	Number               int     `json:"number"`
	Revision             int     `json:"revision"`
	Title                string  `json:"title"`
	Status               string  `json:"status"` // "not_started" | "in_progress" | "completed"
	SessionID            *string `json:"sessionId,omitempty"`
	LastAttemptSessionID *string `json:"lastAttemptSessionId,omitempty"`
	// AttemptID is the score report of the latest completed attempt
	// (mock_attempts), which is what Review opens for an IELTS section.
	AttemptID     *string    `json:"attemptId,omitempty"`
	Score         *float64   `json:"score,omitempty"`
	LastAttemptAt *time.Time `json:"lastAttemptAt,omitempty"`
	AttemptCount  int        `json:"attemptCount"`
	IsUpdated     bool       `json:"isUpdated"`
}

// DraftPaper contains the data needed to publish a new paper.
type DraftPaper struct {
	Exam          string
	Section       string
	Module        string
	Title         string
	ContentSchema int
	Content       json.RawMessage
}
