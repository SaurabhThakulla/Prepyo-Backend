// Package models holds domain models shared across modules.
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

// User represents a user account in the database.
type User struct {
	ID                   string
	Email                string
	GoogleSub            string
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
	PlanStartedAt        time.Time
	PlanValidUntil       *time.Time
	ReferralCode         string
	BonusMockTests       int
	BonusProDays         int
	CreatedAt            time.Time

	AvatarUpdatedAt *time.Time
	CoverUpdatedAt  *time.Time
}

const (
	RoleAdmin   = "admin"
	RoleSuru    = "suru"
	RoleAbhyas  = "abhyas"
	RoleTaiyari = "taiyari"
	RoleUdaan   = "udaan"
)

var planRoles = map[string]string{
	"free":   RoleSuru,
	"weekly": RoleAbhyas,
	"pro":    RoleTaiyari,
	"elite":  RoleUdaan,
}

// RoleForPlan returns the tier role a plan id corresponds to.
func RoleForPlan(planID string) string {
	if role, ok := planRoles[planID]; ok {
		return role
	}
	return RoleSuru
}

// RoleForUser returns the role a user should hold based on their plan validity.
func RoleForUser(u User) string {
	if u.IsAdmin() {
		return RoleAdmin
	}
	if u.PlanValidUntil == nil || !u.PlanValidUntil.After(time.Now()) {
		return RoleSuru
	}
	return RoleForPlan(u.PlanID)
}

func (u User) IsAdmin() bool { return u.Role == RoleAdmin }

// HasActivePaidPlan reports whether the user is inside a live paid subscription.
func (u User) HasActivePaidPlan() bool {
	return u.PlanValidUntil != nil && u.PlanValidUntil.After(time.Now())
}

// DaysRemaining is whole days left in the current period, 0 when it is not a live paid one.
func (u User) DaysRemaining() int {
	if !u.HasActivePaidPlan() {
		return 0
	}
	days := int(time.Until(*u.PlanValidUntil).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// UserProfile is the public profile sent to the client.
type UserProfile struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	PlanID         string    `json:"planId"`
	PlanStartedAt  string    `json:"planStartedAt"`
	PlanValidUntil string    `json:"planValidUntil,omitempty"`
	PaidPlanActive bool      `json:"paidPlanActive"`
	DaysRemaining  int       `json:"daysRemaining"`
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

	Level           int                `json:"level"`
	XPToNextLevel   int                `json:"xpToNextLevel"`
	AvatarUpdatedAt *time.Time         `json:"avatarUpdatedAt,omitempty"`
	CoverUpdatedAt  *time.Time         `json:"coverUpdatedAt,omitempty"`
	Estimate        *ScoreEstimate     `json:"estimate,omitempty"`
	Subscription    *SubscriptionState `json:"subscription,omitempty"`
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
		PlanID:         u.PlanID,
		PlanStartedAt:  u.PlanStartedAt.Format(time.DateOnly),
		PaidPlanActive: u.HasActivePaidPlan(),
		DaysRemaining:  u.DaysRemaining(),
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
	if u.PlanValidUntil != nil {
		profile.PlanValidUntil = u.PlanValidUntil.Format(time.DateOnly)
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
	PlanID   string `json:"planId"`
	PlanName string `json:"planName"`

	// DailySubTestsUsed counts task sets attempted in the learner's own day,
	// across every skill. See billing.Service.State.
	DailySubTestsUsed  int `json:"dailySubTestsUsed"`
	DailySubTestsLimit int `json:"dailySubTestsLimit"`

	// Deprecated: duplicates of DailySubTests* for app builds < 1.1.0, which
	// read the old names and would render an undefined quota without them.
	// Remove once 1.1.0 is live on Vercel.
	DailyEvaluationsUsed  int `json:"dailyEvaluationsUsed"`
	DailyEvaluationsLimit int `json:"dailyEvaluationsLimit"`

	ValidUntil            string `json:"validUntil,omitempty"`
	IsActive              bool   `json:"isActive"`
	MockTestsIncluded     int    `json:"mockTestsIncluded"`
	BonusMockTests        int    `json:"bonusMockTests"`
	TotalMockTestsAllowed int    `json:"totalMockTestsAllowed"`
	MockTestsUsed         int    `json:"mockTestsUsed"`
	BonusDays             int    `json:"bonusDays"`
}

type Plan struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	PriceNPR       int      `json:"priceNPR"`
	DurationMonths int      `json:"durationMonths"`
	DurationDays   int      `json:"durationDays"`
	BonusDays      int      `json:"bonusDays"`
	Features       []string `json:"features"`

	// SubTestsPerDay is the daily allowance of task sets, in any skill.
	SubTestsPerDay int `json:"subTestsPerDay"`

	// Deprecated: duplicate of SubTestsPerDay for app builds < 1.1.0.
	// Remove once 1.1.0 is live on Vercel.
	AIEvaluationsPerDay int `json:"aiEvaluationsPerDay"`

	MockTestsIncluded int  `json:"mockTestsIncluded"`
	IsPopular         bool `json:"isPopular"`
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
	ID            string   `json:"id"`
	Options       []string `json:"options,omitempty"`
	CorrectAnswer string   `json:"correctAnswer,omitempty"`
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
	GroupID          string           `json:"-"`
	Difficulty       string           `json:"difficulty"`
	Tags             []string         `json:"tags"`
	Points           int              `json:"points"`
	SupportedExams   []ExamType       `json:"supportedExams,omitempty"`
	CorrectAnswers   []string         `json:"-"`
	ModelAnswer      string           `json:"-"`
	Explanation      string           `json:"-"`
	FigureData       string           `json:"-"`
}

// SupportsExam reports whether this question may be answered under an exam.
func (q Question) SupportsExam(exam ExamType) bool {
	if len(q.SupportedExams) == 0 {
		return q.Exam == exam
	}
	for _, e := range q.SupportedExams {
		if e == exam {
			return true
		}
	}
	return false
}

// PublicQuestion returns a version of the question with answer keys removed.
func (q Question) PublicQuestion() Question {
	safe := q
	safe.CorrectAnswers = nil
	safe.ModelAnswer = ""
	safe.Explanation = ""
	safe.FigureData = ""

	safe.Blanks = make([]Blank, len(q.Blanks))
	for i, b := range q.Blanks {
		safe.Blanks[i] = Blank{ID: b.ID, Options: b.Options}
	}
	return safe
}

// ReviewQuestion is the question as shown after submission.
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
	ID                   string        `json:"id"`
	ExamVersionID        string        `json:"examVersionId"`
	Exam                 ExamType      `json:"exam"`
	Title                string        `json:"title"`
	Description          string        `json:"description"`
	TotalDurationMinutes int           `json:"totalDurationMinutes"`
	TotalQuestions       int           `json:"totalQuestions"`
	IsDiagnostic         bool          `json:"isDiagnostic"`
	IsGenerated          bool          `json:"isGenerated"`
	Sections             []MockSection `json:"sections"`
}

// ---------------------------------------------------------------------------
// Reading passages
// ---------------------------------------------------------------------------

type ReadingParagraph struct {
	Label string `json:"label"`
	Text  string `json:"text"`
}

type ReadingPassage struct {
	ID            string             `json:"id"`
	ExamVersionID string             `json:"examVersionId"`
	Title         string             `json:"title"`
	Subtitle      string             `json:"subtitle,omitempty"`
	Paragraphs    []ReadingParagraph `json:"paragraphs"`
	Sources       []ReadingParagraph `json:"sources,omitempty"`
	WordCount     int                `json:"wordCount"`
	Difficulty    string             `json:"difficulty"`
	Topic         string             `json:"topic,omitempty"`
	Tags          []string           `json:"tags"`
}

type ReadingGroup struct {
	ID               string             `json:"id"`
	PassageID        string             `json:"passageId"`
	Position         int                `json:"position"`
	TypeID           string             `json:"typeId"`
	TypeName         string             `json:"typeName"`
	Instructions     string             `json:"instructions"`
	Resources        []ReadingParagraph `json:"resources,omitempty"`
	PassageDisplay   string             `json:"passageDisplay"`
	TimeLimitSeconds int                `json:"timeLimitSeconds,omitempty"`
	Questions        []Question         `json:"questions"`
}

type ReadingSet struct {
	Passage        *ReadingPassage `json:"passage,omitempty"`
	Groups         []ReadingGroup  `json:"groups"`
	TotalQuestions int             `json:"totalQuestions"`
}

type ReadingReorderItem struct {
	ID              string             `json:"id"`
	ExamVersionID   string             `json:"examVersionId"`
	Exam            ExamType           `json:"exam"`
	Title           string             `json:"title"`
	Paragraphs      []ReadingParagraph `json:"paragraphs"`
	SourcePassageID string             `json:"sourcePassageId,omitempty"`
	Topic           string             `json:"topic,omitempty"`
	WordCount       int                `json:"wordCount"`
	Difficulty      string             `json:"difficulty"`
	Tags            []string           `json:"tags"`
}

type ReadingTaskType struct {
	TypeID        string `json:"typeId"`
	TypeName      string `json:"typeName"`
	PassageCount  int    `json:"passageCount"`
	QuestionCount int    `json:"questionCount"`
}

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
	ReusedPassages  bool       `json:"reusedPassages"`
	Sets            []ReadingSet `json:"sets,omitempty"`
}

// ---------------------------------------------------------------------------
// Learner activity
// ---------------------------------------------------------------------------

type AnswerSubmission struct {
	QuestionID       string            `json:"questionId"`
	Exam             ExamType          `json:"exam,omitempty"`
	TextResponse     string            `json:"textResponse,omitempty"`
	SelectedOptions  []string          `json:"selectedOptions,omitempty"`
	BlankResponses   map[string]string `json:"blankResponses,omitempty"`
	TimeSpentSeconds int               `json:"timeSpentSeconds,omitempty"`
}

type PracticeAttempt struct {
	ID                 string    `json:"id"`
	QuestionID         string    `json:"questionId"`
	Exam               ExamType  `json:"exam"`
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
	Transcript        string                `json:"transcript,omitempty"`
	CreatedAt         time.Time             `json:"createdAt"`
	Usage             EvaluationUsage       `json:"usage,omitempty"`
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
