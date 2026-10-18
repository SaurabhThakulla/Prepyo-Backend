package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// WritingPromptVersion identifies the active writing evaluation prompt version.
const WritingPromptVersion = "writing.v6" // band ladders, Pearson trait points, score worked out in code.

// EvaluationVersion is the evaluation schema version.
const EvaluationVersion = "v1"

type WritingRequest struct {
	Exam models.ExamType
	// TypeID identifies the task (for IELTS: Academic Task 1, General
	// Training Task 1 letter, or Task 2), which decides the first criterion.
	TypeID      string
	TaskName    string
	Prompt      string
	LearnerText string
	// ContextPassage is source material, not part of the learner's response.
	ContextPassage string
	// FigureData is the tabular data behind an image prompt, if applicable.
	FigureData string
	// WordCount is Prepyo's count of the learner's words, with wording copied
	// from the task left out; MinimumWords is what the task asks for. Both are
	// zero when not counted.
	WordCount    int
	MinimumWords int
	MinScore     float64
	MaxScore     float64
}

// evaluationPayload mirrors the raw JSON returned by the model.
type evaluationPayload struct {
	Summary        string   `json:"summary"`
	Strengths      []string `json:"strengths"`
	Weaknesses     []string `json:"weaknesses"`
	EstimatedScore struct {
		Value      *float64 `json:"value"`
		Confidence string   `json:"confidence"`
	} `json:"estimatedScore"`
	Criteria []struct {
		Name     string  `json:"name"`
		Score    float64 `json:"score"`
		MaxScore float64 `json:"maxScore"`
		Feedback string  `json:"feedback"`
	} `json:"criteria"`
	SentenceFeedback []struct {
		Original    string `json:"original"`
		Correction  string `json:"correction"`
		IssueType   string `json:"issueType"`
		Explanation string `json:"explanation"`
	} `json:"sentenceFeedback"`
}

const maxValidationAttempts = 2

// EvaluateWriting evaluates a writing task submission and returns feedback with an estimated score.
func (g *Gateway) EvaluateWriting(ctx context.Context, req WritingRequest) (models.Evaluation, Usage, error) {
	messages := []chatMessage{
		{Role: "system", Content: writingSystemPrompt(req)},
		{Role: "user", Content: writingUserPrompt(req)},
	}

	var usage Usage

	for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
		raw, attemptUsage, err := g.complete(ctx, g.text, g.models.Writing, WritingPromptVersion, messages, true)
		usage.add(attemptUsage)
		if err != nil {
			return models.Evaluation{}, Usage{}, err
		}

		evaluation, problem := parseWriting(raw, req)
		if problem == nil {
			return evaluation, usage, nil
		}

		g.log.Error("ai output failed validation",
			"model", g.models.Writing, "attempt", attempt, "error", problem)

		if attempt == maxValidationAttempts || ctx.Err() != nil {
			break
		}

		messages = append(messages,
			chatMessage{Role: "assistant", Content: raw},
			chatMessage{Role: "user", Content: fmt.Sprintf(
				"That reply was rejected: %s. Send the complete JSON again with that fixed and everything else unchanged.",
				problem)},
		)
	}

	return models.Evaluation{}, usage, ErrBadOutput
}

// parseWriting decodes and validates a raw response string.
func parseWriting(raw string, req WritingRequest) (models.Evaluation, error) {
	var payload evaluationPayload
	if err := json.Unmarshal([]byte(unfence(raw)), &payload); err != nil {
		return models.Evaluation{}, fmt.Errorf("the reply was not valid JSON (%w)", err)
	}
	return validateWriting(payload, req)
}

// knownMaxScore is a criterion's fixed maximum: 9 for an IELTS criterion, the
// trait's points for PTE, or 0 when the name is not one we know.
func knownMaxScore(name string, spec feedbackSpec) float64 {
	if spec.Exam == models.ExamIELTS {
		return 9
	}
	for _, t := range spec.PTETraits {
		if traitKey(t.Name) == traitKey(name) {
			return float64(t.Max)
		}
	}
	return 0
}

// unfence removes the markdown code fence (```json ... ```) some models wrap
// their JSON in, so the reply parses without asking the model again.
func unfence(raw string) string {
	s := strings.TrimSpace(raw)
	if !strings.HasPrefix(s, "```") || !strings.HasSuffix(s, "```") || len(s) < 6 {
		return raw
	}
	s = strings.TrimSpace(s[3 : len(s)-3])
	if tag, rest, ok := strings.Cut(s, "\n"); ok && !strings.ContainsAny(tag, "{[") {
		s = rest
	} else if strings.HasPrefix(strings.ToLower(s), "json") {
		s = s[4:]
	}
	return strings.TrimSpace(s)
}

// add accumulates usage metrics from an attempt.
func (u *Usage) add(other Usage) {
	if other.Provider != "" {
		u.Provider = other.Provider
		u.Model = other.Model
		u.PromptVersion = other.PromptVersion
	}
	u.PromptTokens += other.PromptTokens
	u.CompletionTokens += other.CompletionTokens
	u.LatencyMS += other.LatencyMS
	u.CostUSD += other.CostUSD
}

func validateWriting(p evaluationPayload, req WritingRequest) (models.Evaluation, error) {
	var traits []pteTrait
	formCap := -1
	if req.Exam == models.ExamPTE {
		task := pteWritingTask(req.TypeID, req.TaskName)
		traits = pteWritingTraits(task)
		formCap = pteFormFromLength(task, len(strings.Fields(req.LearnerText)))
	}
	return validateFeedback(p, feedbackSpec{
		Exam:       req.Exam,
		Skill:      models.SkillWriting,
		MinScore:   req.MinScore,
		MaxScore:   req.MaxScore,
		Quotable:   req.LearnerText,
		TaskName:   req.TaskName,
		TypeID:     req.TypeID,
		PTETraits:  traits,
		PTEFormCap: formCap,
	})
}

type feedbackSpec struct {
	Exam     models.ExamType
	Skill    models.SkillType
	MinScore float64
	MaxScore float64
	Quotable string
	TaskName string
	TypeID   string
	// WithoutPronunciation marks feedback produced from a transcript rather
	// than a recording. Pronunciation cannot be judged from words on a page, so
	// the usual four-criterion IELTS shape does not apply and a model that
	// returns one anyway is guessing.
	WithoutPronunciation bool
	// PTETraits are Pearson's traits for the task; when set, the 10-90
	// estimate is worked out from the trait points. PTEFormCap is the most
	// Form can score given the response's length, or -1.
	PTETraits  []pteTrait
	PTEFormCap int
}

// validateFeedback validates the parsed evaluation payload against the specification.
func validateFeedback(p evaluationPayload, spec feedbackSpec) (models.Evaluation, error) {
	if strings.TrimSpace(p.Summary) == "" {
		return models.Evaluation{}, fmt.Errorf("summary is empty")
	}

	confidence := strings.ToLower(strings.TrimSpace(p.EstimatedScore.Confidence))
	switch confidence {
	case "low", "medium", "high":
	case "":
		confidence = "low"
	default:
		return models.Evaluation{}, fmt.Errorf("unknown confidence %q", p.EstimatedScore.Confidence)
	}

	score := p.EstimatedScore.Value
	// With PTE traits the model's own number is only "scorable or not"; the
	// estimate comes from the trait points below.
	if score != nil && len(spec.PTETraits) == 0 {
		// Validate before rounding, which otherwise hides out-of-range values.
		if math.IsNaN(*score) || math.IsInf(*score, 0) || *score < spec.MinScore || *score > spec.MaxScore {
			return models.Evaluation{}, fmt.Errorf("score %.2f outside %.1f-%.1f for %s", *score, spec.MinScore, spec.MaxScore, spec.Exam)
		}
		if spec.Exam == models.ExamIELTS {
			rounded := scoring.RoundIELTSBand(*score)
			score = &rounded
		} else if spec.Exam == models.ExamPTE {
			rounded := math.Round(*score)
			score = &rounded
		}
		if *score < spec.MinScore || *score > spec.MaxScore {
			return models.Evaluation{}, fmt.Errorf("score %.2f outside %.1f-%.1f for %s", *score, spec.MinScore, spec.MaxScore, spec.Exam)
		}
	}

	criteria := make([]models.EvaluationCriterion, 0, len(p.Criteria))
	for _, c := range p.Criteria {
		if c.MaxScore <= 0 {
			c.MaxScore = knownMaxScore(c.Name, spec) // a terse reply may leave out a fixed maximum
		}
		if strings.TrimSpace(c.Name) == "" || c.MaxScore <= 0 {
			continue
		}
		if c.Score < 0 || c.Score > c.MaxScore {
			return models.Evaluation{}, fmt.Errorf("criterion %q scored %.2f of %.2f", c.Name, c.Score, c.MaxScore)
		}
		criteria = append(criteria, models.EvaluationCriterion{
			Name:     c.Name,
			Score:    c.Score,
			MaxScore: c.MaxScore,
			Feedback: c.Feedback,
		})
	}

	if spec.Exam == models.ExamIELTS && score != nil {
		band, err := validateIELTSCriteria(criteria, spec)
		if err != nil {
			return models.Evaluation{}, err
		}
		score = &band
	}
	if len(spec.PTETraits) > 0 && score != nil {
		points, rated, err := ptePointsScore(criteria, spec.PTETraits, spec.PTEFormCap)
		if err != nil {
			return models.Evaluation{}, err
		}
		score, criteria = &points, rated
	}

	haystack := normaliseSpace(spec.Quotable)
	sentences := make([]models.SentenceFeedback, 0, len(p.SentenceFeedback))
	for _, s := range p.SentenceFeedback {
		quoted := normaliseSpace(s.Original)
		if quoted == "" || !strings.Contains(haystack, quoted) {
			continue
		}
		sentences = append(sentences, models.SentenceFeedback{
			Original:    s.Original,
			Correction:  s.Correction,
			IssueType:   s.IssueType,
			Explanation: s.Explanation,
		})
	}

	return models.Evaluation{
		Exam:              spec.Exam,
		Skill:             spec.Skill,
		EvaluationVersion: EvaluationVersion,
		EstimatedScore:    score,
		ScoreConfidence:   confidence,
		Summary:           strings.TrimSpace(p.Summary),
		Criteria:          criteria,
		Strengths:         trimAll(p.Strengths),
		Weaknesses:        trimAll(p.Weaknesses),
		SentenceFeedback:  sentences,
	}, nil
}

func writingSystemPrompt(req WritingRequest) string {
	if isIELTS(req.Exam) {
		return ieltsWritingGuidance(req) + "\n" + ieltsReplyShape("", "medium", "")
	}
	return pteWritingGuidance(req) + "\n" + pteReplyShape("", "medium", "")
}

// ieltsReplyShape is the JSON the IELTS prompts ask for. It carries no example
// numbers: an example band in the prompt pulls every rating towards it.
func ieltsReplyShape(transcriptLine, confidence, summaryNote string) string {
	return fmt.Sprintf(`Reply with JSON only, in this shape:
{
%s  "summary": "one concise sentence: overall band level and key improvement%s",
  "criteria": [{"name": "...", "score": <whole band>, "maxScore": 9, "feedback": "one concise sentence citing evidence"}],
  "estimatedScore": {"value": <rounded mean of the criteria, or null if the response cannot be rated>, "confidence": "%s"},
  "strengths": ["two short phrases"],
  "weaknesses": ["two short phrases"],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "grammar", "explanation": "under 15 words"}]
}
`+replyBrevity, transcriptLine, summaryNote, confidence)
}

// replyBrevity keeps the written feedback short: the scores come from the
// criteria, so length adds cost without adding accuracy.
const replyBrevity = "\nKeep every text field concise: give at most two sentenceFeedback entries (the two highest-impact corrections), one concise sentence per criterion feedback, and under 15 words per explanation.\n"

// pteReplyShape is the JSON the PTE prompts ask for: trait points on
// Pearson's scales, from which Prepyo works out the 10-90 estimate.
func pteReplyShape(transcriptLine, confidence, summaryNote string) string {
	return fmt.Sprintf(`Reply with JSON only, in this shape:
{
%s  "summary": "one concise sentence: overall level and key improvement%s",
  "criteria": [{"name": "...", "score": <whole points>, "maxScore": <that trait's maximum>, "feedback": "one concise sentence citing evidence"}],
  "estimatedScore": {"value": <total of the trait points, or null if the response cannot be rated>, "confidence": "%s"},
  "strengths": ["two short phrases"],
  "weaknesses": ["two short phrases"],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "grammar", "explanation": "under 15 words"}]
}
`+replyBrevity, transcriptLine, summaryNote, confidence)
}

func writingUserPrompt(req WritingRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n\nPrompt:\n%s\n", req.TaskName, req.Prompt)
	if passage := strings.TrimSpace(req.ContextPassage); passage != "" {
		fmt.Fprintf(&b, "\nSource passage (reference material, not the learner's response):\n%s\n", passage)
	}
	if req.WordCount > 0 && req.MinimumWords > 0 {
		fmt.Fprintf(&b, "\nWord count: %d (minimum for this task: %d).\n", req.WordCount, req.MinimumWords)
	}
	if figure := strings.TrimSpace(req.FigureData); figure != "" {
		fmt.Fprintf(&b, "\nThe learner was shown this figure as an image. They could not read "+
			"these numbers as text, so do not penalise wording that differs from it. Use it only "+
			"to judge whether what they report is accurate:\n%s\n", figure)
	}
	fmt.Fprintf(&b, "\nLearner's response:\n%s", req.LearnerText)
	return b.String()
}

// normaliseSpace collapses whitespace runs into single spaces.
func normaliseSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func trimAll(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
