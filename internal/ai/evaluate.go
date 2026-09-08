package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// WritingPromptVersion identifies the active writing evaluation prompt version.
const WritingPromptVersion = "writing.v1"

// EvaluationVersion is the evaluation schema version.
const EvaluationVersion = "v1"

type WritingRequest struct {
	Exam        models.ExamType
	TaskName    string
	Prompt      string
	LearnerText string
	// FigureData is the tabular data behind an image prompt, if applicable.
	FigureData string
	MinScore   float64
	MaxScore   float64
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
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return models.Evaluation{}, fmt.Errorf("the reply was not valid JSON (%w)", err)
	}
	return validateWriting(payload, req)
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
}

func validateWriting(p evaluationPayload, req WritingRequest) (models.Evaluation, error) {
	return validateFeedback(p, feedbackSpec{
		Exam:     req.Exam,
		Skill:    models.SkillWriting,
		MinScore: req.MinScore,
		MaxScore: req.MaxScore,
		Quotable: req.LearnerText,
	})
}

type feedbackSpec struct {
	Exam     models.ExamType
	Skill    models.SkillType
	MinScore float64
	MaxScore float64
	Quotable string
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
	if score != nil {
		if *score < spec.MinScore || *score > spec.MaxScore {
			return models.Evaluation{}, fmt.Errorf("score %.2f outside %.1f-%.1f for %s", *score, spec.MinScore, spec.MaxScore, spec.Exam)
		}
	}

	criteria := make([]models.EvaluationCriterion, 0, len(p.Criteria))
	for _, c := range p.Criteria {
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
	var b strings.Builder
	b.WriteString("You are an experienced ")
	b.WriteString(string(req.Exam))
	b.WriteString(" writing examiner giving practice feedback to a learner in Nepal.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Judge only the text the learner wrote. Never quote or invent a sentence they did not write.\n")
	b.WriteString("- Every entry in sentenceFeedback must copy an exact sentence from the learner's text into `original`.\n")
	b.WriteString("- If the response is too short or off-topic to judge, set estimatedScore.value to null and say so in the summary.\n")
	if req.Exam == models.ExamPTE {
		b.WriteString("- CRITICAL FOR PTE: estimatedScore.value MUST be on the 10-90 PTE points scale (e.g. 65, 79, 85). DO NOT output 0-9 IELTS band numbers.\n")
	} else {
		b.WriteString("- CRITICAL FOR IELTS: estimatedScore.value MUST be on the 0.0-9.0 IELTS band scale in 0.5 steps (e.g. 6.5, 7.0, 7.5).\n")
	}
	b.WriteString("- Set estimatedScore.confidence to low, medium or high based on how much evidence the response gives you.\n")
	b.WriteString("- Use the published assessment criteria for this exam. Do not invent weightings.\n")
	b.WriteString("- criteria[].maxScore is the maximum the published rubric gives that one criterion. It is not the exam's overall scale.\n")
	b.WriteString("- The example below already uses this exam's scale. Copy its shape, never its numbers.\n\n")
	b.WriteString(fmt.Sprintf(`Reply with JSON only, in this shape:
{
  "summary": "two or three sentences",
  "estimatedScore": {"value": %.1f, "confidence": "medium"},
  "criteria": [{"name": "...", "score": <this criterion's score>, "maxScore": <this criterion's maximum>, "feedback": "..."}],
  "strengths": ["..."],
  "weaknesses": ["..."],
  "sentenceFeedback": [{"original": "...", "correction": "...", "issueType": "grammar", "explanation": "..."}]
}`, exampleScore(req.MinScore, req.MaxScore)))
	return b.String()
}

// exampleScore returns a representative score value formatted for the given scale.
func exampleScore(min, max float64) float64 {
	v := min + 0.75*(max-min)
	if max-min > 20 {
		return math.Round(v)
	}
	return math.Round(v*2) / 2
}

func writingUserPrompt(req WritingRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Task: %s\n\nPrompt:\n%s\n", req.TaskName, req.Prompt)
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
