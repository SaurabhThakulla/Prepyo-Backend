// Package models holds the types shared across modules and sent to the client.
//
// Fields that can be worked out from other data are marked "derived": they are
// computed when a response is built, never stored. Level always agrees with XP
// because there is only one place it can come from.
package models

import "time"

type ExamType string

const (
	ExamPTE   ExamType = "PTE"
	ExamIELTS ExamType = "IELTS"
)

func (e ExamType) Valid() bool { return e == ExamPTE || e == ExamIELTS }

type SkillType string

const (
	SkillSpeaking  SkillType = "speaking"
	SkillWriting   SkillType = "writing"
	SkillReading   SkillType = "reading"
	SkillListening SkillType = "listening"
)

func (s SkillType) Valid() bool {
	switch s {
	case SkillSpeaking, SkillWriting, SkillReading, SkillListening:
		return true
	}
	return false
}

// AllSkills is the canonical order used by dashboards and progress views.
var AllSkills = []SkillType{SkillSpeaking, SkillWriting, SkillReading, SkillListening}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

// User is the stored row. It never leaves the backend, because it carries the
// password hash and the role.
type User struct {
	ID                   string
	Email                string
	PasswordHash         string
	Name                 string
	Role                 string
	TargetExam           ExamType
	TargetScore          *float64
	ExamDate             *time.Time
	NepalRegion          string
	XP                   int
	StreakDays           int
	StreakLastActiveDate *time.Time
	Timezone             string
	PlanID               string
	PlanValidUntil       *time.Time
	ReferralCode         string
	BonusMockTests       int
	BonusProDays         int
	CreatedAt            time.Time

	// Set when a profile image exists. The bytes themselves are not loaded
	// here: the user row is read on every authenticated request and must stay
	// cheap. See users.Repository.Image.
	AvatarUpdatedAt *time.Time
	CoverUpdatedAt  *time.Time
}

const RoleAdmin = "admin"

func (u User) IsAdmin() bool { return u.Role == RoleAdmin }

// UserProfile is what the client receives. Build it with NewUserProfile so the
// derived fields are always filled the same way.
type UserProfile struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	TargetExam     ExamType  `json:"targetExam"`
	TargetScore    *float64  `json:"targetScore"`
	ExamDate       string    `json:"examDate,omitempty"`
	NepalRegion    string    `json:"nepalRegion"`
	XP             int       `json:"xp"`
	StreakDays     int       `json:"streakDays"`
	Timezone       string    `json:"timezone"`
	ReferralCode   string    `json:"referralCode"`
	BonusMockTests int       `json:"bonusMockTests"`
	BonusProDays   int       `json:"bonusProDays"`
	CreatedAt      time.Time `json:"createdAt"`

	// Derived.
	Level         int `json:"level"`
	XPToNextLevel int `json:"xpToNextLevel"`

	// When a profile image was last set, or absent when there is none. The
	// client uses it both to decide whether to render an image and to bust its
	// own cache when one changes.
	AvatarUpdatedAt *time.Time `json:"avatarUpdatedAt,omitempty"`
	CoverUpdatedAt  *time.Time `json:"coverUpdatedAt,omitempty"`

	// Filled from other tables by the caller. Nil when not requested, so a
	// cheap endpoint does not pay for an estimate it will not use.
	Estimate     *ScoreEstimate     `json:"estimate,omitempty"`
	Subscription *SubscriptionState `json:"subscription,omitempty"`
}

// XPPerLevel is the width of one level. Levels start at 1.
const XPPerLevel = 400

func LevelForXP(xp int) int { return xp/XPPerLevel + 1 }

func NewUserProfile(u User) UserProfile {
	level := LevelForXP(u.XP)
	profile := UserProfile{
		ID:             u.ID,
		Email:          u.Email,
		Name:           u.Name,
		Role:           u.Role,
		TargetExam:     u.TargetExam,
		TargetScore:    u.TargetScore,
		NepalRegion:    u.NepalRegion,
		XP:             u.XP,
		StreakDays:     u.StreakDays,
		Timezone:       u.Timezone,
		ReferralCode:   u.ReferralCode,
		BonusMockTests: u.BonusMockTests,
		BonusProDays:   u.BonusProDays,
		CreatedAt:      u.CreatedAt,
		Level:          level,
		XPToNextLevel:  level*XPPerLevel - u.XP,

		AvatarUpdatedAt: u.AvatarUpdatedAt,
		CoverUpdatedAt:  u.CoverUpdatedAt,
	}
	if u.ExamDate != nil {
		profile.ExamDate = u.ExamDate.Format(time.DateOnly)
	}
	return profile
}

// ScoreEstimate is the learner's current standing. Nil Value means there is not
// enough evidence yet, which the UI shows as "not enough data" rather than
// inventing a number.
type ScoreEstimate struct {
	Value       *float64  `json:"value"`
	Confidence  string    `json:"confidence"` // low, medium, high
	BasedOn     int       `json:"basedOn"`    // attempts behind the estimate
	TargetScore *float64  `json:"targetScore"`
	TargetGap   *float64  `json:"targetGap"`
	Readiness   *int      `json:"readiness"` // percent of target reached
	UpdatedAt   time.Time `json:"updatedAt"`
}

type SubscriptionState struct {
	PlanID                string `json:"planId"`
	PlanName              string `json:"planName"`
	ValidUntil            string `json:"validUntil,omitempty"`
	IsActive              bool   `json:"isActive"`
	DailyEvaluationsUsed  int    `json:"dailyEvaluationsUsed"`
	DailyEvaluationsLimit int    `json:"dailyEvaluationsLimit"`
	MockTestsIncluded     int    `json:"mockTestsIncluded"`
	BonusMockTests        int    `json:"bonusMockTests"`
	TotalMockTestsAllowed int    `json:"totalMockTestsAllowed"`
	MockTestsUsed         int    `json:"mockTestsUsed"`
	BonusDays             int    `json:"bonusDays"`
}

type Plan struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	PriceNPR            int      `json:"priceNPR"`
	DurationMonths      int      `json:"durationMonths"`
	DurationDays        int      `json:"durationDays"`
	BonusDays           int      `json:"bonusDays"`
	Features            []string `json:"features"`
	AIEvaluationsPerDay int      `json:"aiEvaluationsPerDay"`
	MockTestsIncluded   int      `json:"mockTestsIncluded"`
	IsPopular           bool     `json:"isPopular"`
}

// ---------------------------------------------------------------------------
// Referrals & Subscriptions
// ---------------------------------------------------------------------------

type ReferralStatus string

const (
	ReferralPending   ReferralStatus = "pending"
	ReferralCompleted ReferralStatus = "completed"
	ReferralCancelled ReferralStatus = "cancelled"
)

type Referral struct {
	ID               string         `json:"id"`
	ReferrerID       string         `json:"referrerId"`
	RefereeID        string         `json:"refereeId"`
	ReferralCode     string         `json:"referralCode"`
	Status           ReferralStatus `json:"status"`
	RewardReferrerXP int            `json:"rewardReferrerXP"`
	RewardRefereeXP  int            `json:"rewardRefereeXP"`
	CreatedAt        time.Time      `json:"createdAt"`
	CompletedAt      *time.Time     `json:"completedAt,omitempty"`
}

type ReferralMilestone struct {
	Required  int  `json:"required"`
	Current   int  `json:"current"`
	Completed bool `json:"completed"`
}

type ReferralMilestones struct {
	ThreeReferrals ReferralMilestone `json:"threeReferrals"`
	FiveReferrals  ReferralMilestone `json:"fiveReferrals"`
}

type ReferralStats struct {
	TotalInvited  int `json:"totalInvited"`
	Pending       int `json:"pending"`
	Completed     int `json:"completed"`
	TotalXPEarned int `json:"totalXpEarned"`
}

type RecentReferralItem struct {
	ID          string     `json:"id"`
	FriendName  string     `json:"friendName"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type ReferralOverview struct {
	ReferralCode    string               `json:"referralCode"`
	ShareLink       string               `json:"shareLink"`
	Stats           ReferralStats        `json:"stats"`
	Milestones      ReferralMilestones   `json:"milestones"`
	RecentReferrals []RecentReferralItem `json:"recentReferrals"`
}

type ReferralValidation struct {
	Valid        bool   `json:"valid"`
	ReferrerName string `json:"referrerName,omitempty"`
	Message      string `json:"message"`
}

type SubscriptionPayment struct {
	ID             string     `json:"id"`
	UserID         string     `json:"userId"`
	PlanID         string     `json:"planId"`
	PaymentGateway string     `json:"paymentGateway"`
	TransactionID  string     `json:"transactionId"`
	AmountNPR      int        `json:"amountNPR"`
	Status         string     `json:"status"`
	BaseDays       int        `json:"baseDays"`
	BonusDays      int        `json:"bonusDays"`
	EffectiveDays  int        `json:"effectiveDays"`
	ProcessedAt    *time.Time `json:"processedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// ---------------------------------------------------------------------------
// Exam content
// ---------------------------------------------------------------------------

// ExamVersion freezes the scoring scale for an exam. Attempts store the version
// they ran under so an old result keeps its original meaning.
type ExamVersion struct {
	ID          string   `json:"id"`
	Exam        ExamType `json:"exam"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	MinScore    float64  `json:"minScore"`
	MaxScore    float64  `json:"maxScore"`
	ScoreStep   float64  `json:"scoreStep"`
	IsCurrent   bool     `json:"isCurrent"`
}

type QuestionOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Blank struct {
	ID      string   `json:"id"`
	Options []string `json:"options,omitempty"`
	// CorrectAnswer is stripped before a question is sent to a learner.
	CorrectAnswer string `json:"correctAnswer,omitempty"`
}

type Question struct {
	ID               string           `json:"id"`
	ExamVersionID    string           `json:"examVersionId"`
	Exam             ExamType         `json:"exam"`
	Skill            SkillType        `json:"skill"`
	TypeID           string           `json:"typeId"`
	TypeName         string           `json:"typeName"`
	Title            string           `json:"title"`
	Prompt           string           `json:"prompt"`
	ContextPassage   string           `json:"contextPassage,omitempty"`
	AudioURL         string           `json:"audioUrl,omitempty"`
	AudioTranscript  string           `json:"audioTranscript,omitempty"`
	ImageURL         string           `json:"imageUrl,omitempty"`
	PrepTimeSeconds  int              `json:"prepTimeSeconds,omitempty"`
	TimeLimitSeconds int              `json:"timeLimitSeconds"`
	Options          []QuestionOption `json:"options,omitempty"`
	Blanks           []Blank          `json:"blanks,omitempty"`
	Difficulty       string           `json:"difficulty"`
	Tags             []string         `json:"tags"`
	Points           int              `json:"points"`

	// Answer key. Never serialised: PublicQuestion drops these before the
	// question reaches a learner, so the browser cannot read the answers.
	CorrectAnswers []string `json:"-"`
	ModelAnswer    string   `json:"-"`
	Explanation    string   `json:"-"`
}

// PublicQuestion is the question as a learner sees it while answering: no
// answer key, no explanation.
func (q Question) PublicQuestion() Question {
	safe := q
	safe.CorrectAnswers = nil
	safe.ModelAnswer = ""
	safe.Explanation = ""

	safe.Blanks = make([]Blank, len(q.Blanks))
	for i, b := range q.Blanks {
		safe.Blanks[i] = Blank{ID: b.ID, Options: b.Options}
	}
	return safe
}

// ReviewQuestion is the question as shown after submission, with the answer key
// and explanation restored.
type ReviewQuestion struct {
	Question
	CorrectAnswers []string `json:"correctAnswers,omitempty"`
	ModelAnswer    string   `json:"modelAnswer,omitempty"`
	Explanation    string   `json:"explanation,omitempty"`
}

func (q Question) ForReview() ReviewQuestion {
	return ReviewQuestion{
		Question:       q,
		CorrectAnswers: q.CorrectAnswers,
		ModelAnswer:    q.ModelAnswer,
		Explanation:    q.Explanation,
	}
}

type MockSection struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Skill           SkillType `json:"skill"`
	DurationMinutes int       `json:"durationMinutes"`
	QuestionIDs     []string  `json:"questionIds"`
}

type Mock struct {
	ID                   string   `json:"id"`
	ExamVersionID        string   `json:"examVersionId"`
	Exam                 ExamType `json:"exam"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	TotalDurationMinutes int      `json:"totalDurationMinutes"`
	TotalQuestions       int      `json:"totalQuestions"`
	IsDiagnostic         bool     `json:"isDiagnostic"`

	// IsGenerated marks a blueprint whose paper is composed per learner when
	// they start it. Sections and TotalQuestions stay empty for these, because
	// there is no fixed question list to report: the client sends the learner
	// to the reading mock endpoints instead of reading questions from here.
	IsGenerated bool `json:"isGenerated"`

	Sections []MockSection `json:"sections"`
}

// ---------------------------------------------------------------------------
// Reading passages
// ---------------------------------------------------------------------------

// ReadingParagraph is one labelled block of a passage. The label is not
// decoration: a Matching Information answer is a paragraph label.
type ReadingParagraph struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

// ReadingPassage is the text a set of reading questions is written about.
type ReadingPassage struct {
	ID            string             `json:"id"`
	ExamVersionID string             `json:"examVersionId"`
	Exam          ExamType           `json:"exam"`
	Title         string             `json:"title"`
	Subtitle      string             `json:"subtitle,omitempty"`
	Paragraphs    []ReadingParagraph `json:"paragraphs"`

	// Sources carries attributed excerpts for passages written in several
	// voices, which is what Find the Writer questions match against. Empty for
	// single-author passages.
	Sources []ReadingParagraph `json:"sources,omitempty"`

	WordCount  int      `json:"wordCount"`
	Difficulty string   `json:"difficulty"`
	Topic      string   `json:"topic,omitempty"`
	Tags       []string `json:"tags"`
}

// ReadingGroup is one task set on a passage: a type, the instruction line above
// it, and the questions in it.
type ReadingGroup struct {
	ID           string `json:"id"`
	PassageID    string `json:"passageId"`
	Position     int    `json:"position"`
	TypeID       string `json:"typeId"`
	TypeName     string `json:"typeName"`
	Instructions string `json:"instructions"`

	// Resources is material belonging to the task rather than to the passage:
	// the boxes of an ordering task, a summary with gaps in it.
	Resources []ReadingParagraph `json:"resources,omitempty"`

	TimeLimitSeconds int        `json:"timeLimitSeconds,omitempty"`
	Questions        []Question `json:"questions"`
}

// ReadingSet is a passage with some of its groups attached. Practice returns
// one group; a mock returns every group on the passage.
type ReadingSet struct {
	Passage        ReadingPassage `json:"passage"`
	Groups         []ReadingGroup `json:"groups"`
	TotalQuestions int            `json:"totalQuestions"`
}

// ReadingTaskType is one entry in the "what can I practise?" menu, carrying the
// size of the bank behind it so the client can grey out an empty choice.
type ReadingTaskType struct {
	TypeID        string `json:"typeId"`
	TypeName      string `json:"typeName"`
	PassageCount  int    `json:"passageCount"`
	QuestionCount int    `json:"questionCount"`
}

// ReadingMockSession is one generated reading paper. Sets are filled only when
// the paper itself is served, not when a session is listed.
type ReadingMockSession struct {
	ID              string     `json:"id"`
	MockID          string     `json:"mockId"`
	MockTitle       string     `json:"mockTitle,omitempty"`
	Exam            ExamType   `json:"exam"`
	ExamVersionID   string     `json:"examVersionId"`
	Status          string     `json:"status"`
	DurationMinutes int        `json:"durationMinutes"`
	TotalQuestions  int        `json:"totalQuestions"`
	PassageIDs      []string   `json:"passageIds"`
	CreatedAt       time.Time  `json:"createdAt"`
	SubmittedAt     *time.Time `json:"submittedAt,omitempty"`

	// ReusedPassages is true when the bank held fewer unseen passages than the
	// paper needed and one had to be repeated. It is reported rather than
	// hidden, so a repeat does not read as a bug.
	ReusedPassages bool `json:"reusedPassages"`

	Sets []ReadingSet `json:"sets,omitempty"`
}

// ---------------------------------------------------------------------------
// Learner activity
// ---------------------------------------------------------------------------

// AnswerSubmission is one answer sent by the client. Note what is absent: no
// score, no correctness flag. The server works those out.
type AnswerSubmission struct {
	QuestionID       string            `json:"questionId"`
	TextResponse     string            `json:"textResponse,omitempty"`
	SelectedOptions  []string          `json:"selectedOptions,omitempty"`
	BlankResponses   map[string]string `json:"blankResponses,omitempty"`
	TimeSpentSeconds int               `json:"timeSpentSeconds,omitempty"`
}

type PracticeAttempt struct {
	ID                 string    `json:"id"`
	QuestionID         string    `json:"questionId"`
	ExamVersionID      string    `json:"examVersionId"`
	IsCorrect          bool      `json:"isCorrect"`
	Score              float64   `json:"score"`
	MaxScore           float64   `json:"maxScore"`
	AccuracyPercentage int       `json:"accuracyPercentage"`
	Feedback           string    `json:"feedback"`
	UserResponse       string    `json:"userResponse,omitempty"`
	TimeSpentSeconds   int       `json:"timeSpentSeconds"`
	CreatedAt          time.Time `json:"createdAt"`
}

type MockAttempt struct {
	ID              string                `json:"id"`
	MockID          string                `json:"mockId"`
	ExamVersionID   string                `json:"examVersionId"`
	Exam            ExamType              `json:"exam"`
	UserScore       float64               `json:"userScore"`
	SkillScores     map[SkillType]float64 `json:"skillScores"`
	TotalCorrect    int                   `json:"totalCorrect"`
	TotalQuestions  int                   `json:"totalQuestions"`
	DurationSeconds int                   `json:"durationSeconds"`
	CompletedAt     time.Time             `json:"completedAt"`
}

type Mistake struct {
	ID              string    `json:"id"`
	QuestionID      string    `json:"questionId"`
	QuestionTitle   string    `json:"questionTitle"`
	Exam            ExamType  `json:"exam"`
	Skill           SkillType `json:"skill"`
	TypeName        string    `json:"typeName"`
	Prompt          string    `json:"prompt"`
	UserResponse    string    `json:"userResponse"`
	CorrectResponse string    `json:"correctResponse"`
	Explanation     string    `json:"explanation"`
	ErrorTag        string    `json:"errorTag"`
	FailedCount     int       `json:"failedCount"`
	Resolved        bool      `json:"resolved"`
	LastAttemptedAt time.Time `json:"lastAttemptedAt"`
}

// ---------------------------------------------------------------------------
// AI evaluation
// ---------------------------------------------------------------------------

type EvaluationCriterion struct {
	Name     string  `json:"name"`
	Score    float64 `json:"score"`
	MaxScore float64 `json:"maxScore"`
	Feedback string  `json:"feedback"`
}

type SentenceFeedback struct {
	Original    string `json:"original"`
	Correction  string `json:"correction"`
	IssueType   string `json:"issueType"`
	Explanation string `json:"explanation"`
}

// Evaluation is qualitative AI feedback plus an estimate. EstimatedScore is a
// practice estimate, never an official result; the client labels it that way.
type Evaluation struct {
	ID                string                `json:"id"`
	QuestionID        string                `json:"questionId,omitempty"`
	Exam              ExamType              `json:"exam"`
	Skill             SkillType             `json:"skill"`
	EvaluationVersion string                `json:"evaluationVersion"`
	EstimatedScore    *float64              `json:"estimatedScore"`
	ScoreConfidence   string                `json:"scoreConfidence"`
	Summary           string                `json:"summary"`
	Criteria          []EvaluationCriterion `json:"criteria"`
	Strengths         []string              `json:"strengths"`
	Weaknesses        []string              `json:"weaknesses"`
	SentenceFeedback  []SentenceFeedback    `json:"sentenceFeedback"`
	ModelRewrite      string                `json:"modelRewrite,omitempty"`

	// Transcript is what the learner was heard to say. Speaking only: it is the
	// evidence every other field here rests on, so it is stored and shown with
	// them rather than thrown away after scoring.
	Transcript string `json:"transcript,omitempty"`

	CreatedAt time.Time `json:"createdAt"`

	// Usage is recorded for cost tracking and shown only to admins.
	Usage EvaluationUsage `json:"usage,omitempty"`
}

type EvaluationUsage struct {
	Provider         string `json:"provider"`
	Model            string `json:"model"`
	PromptVersion    string `json:"promptVersion"`
	PromptTokens     int    `json:"promptTokens"`
	CompletionTokens int    `json:"completionTokens"`
	LatencyMS        int    `json:"latencyMs"`
}

// ---------------------------------------------------------------------------
// Gamification and notifications
// ---------------------------------------------------------------------------

type XPTransaction struct {
	ID        string    `json:"id"`
	Amount    int       `json:"amount"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
}

type DailyMission struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Skill          SkillType `json:"skill"`
	TaskType       string    `json:"taskType,omitempty"`
	TargetCount    int       `json:"targetCount"`
	CompletedCount int       `json:"completedCount"`
	XPReward       int       `json:"xpReward"`
	Completed      bool      `json:"completed"`
}

type LeaderboardEntry struct {
	Rank        int      `json:"rank"`
	UserID      string   `json:"userId"`
	Name        string   `json:"name"`
	NepalRegion string   `json:"nepalRegion"`
	Exam        ExamType `json:"exam"`
	XP          int      `json:"xp"`
	Level       int      `json:"level"`
	StreakDays  int      `json:"streakDays"`
	IsYou       bool     `json:"isYou"`
}

type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Read      bool      `json:"read"`
	ActionURL string    `json:"actionUrl,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
