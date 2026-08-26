package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// WritingPromptVersion is stamped on every stored evaluation. Change it
// whenever the prompt below changes, so old feedback stays traceable to the
// wording that produced it.
const WritingPromptVersion = "writing.v1"

// EvaluationVersion is the shape of the JSON contract, stored alongside each
// result so a later reader knows how to interpret it.
const EvaluationVersion = "v1"

type WritingRequest struct {
	Exam        models.ExamType
	TaskName    string
	Prompt      string
	LearnerText string
	// MinScore and MaxScore come from the exam version, so PTE is validated
	// against 10-90 and IELTS against 0-9.
	MinScore float64
	MaxScore float64
}

// evaluationPayload mirrors the JSON the model is told to return. It is
// deliberately separate from models.Evaluation: the model fills this, and only
// validated fields are copied across.
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

// maxValidationAttempts covers one correction round. A reply can be perfectly
// well-formed JSON and still be unusable — most often a score on the wrong
// exam's scale — and telling the model exactly what was wrong fixes it far
// more cheaply than failing the learner's submission.
const maxValidationAttempts = 2

// EvaluateWriting returns qualitative feedback and an estimated score.
//
// The estimate is a practice estimate. Neither this package nor its callers
// present it as an official Pearson or IELTS result.
func (g *Gateway) EvaluateWriting(ctx context.Context, req WritingRequest) (models.Evaluation, Usage, error) {
	messages := []chatMessage{
		{Role: "system", Content: writingSystemPrompt(req)},
		{Role: "user", Content: writingUserPrompt(req)},
	}

	// Usage accumulates across attempts: a correction round costs real tokens,
	// and cost reporting that hid them would understate what evaluations spend.
	var usage Usage

	for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
		raw, attemptUsage, err := g.complete(ctx, g.models.Writing, WritingPromptVersion, messages, true)
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

		// Hand back the exact reply and the exact complaint, so the retry
		// corrects the one field that was wrong instead of starting over and
		// risking a different mistake.
		messages = append(messages,
			chatMessage{Role: "assistant", Content: raw},
			chatMessage{Role: "user", Content: fmt.Sprintf(
				"That reply was rejected: %s. Send the complete JSON again with that fixed and everything else unchanged.",
				problem)},
		)
	}

	return models.Evaluation{}, usage, ErrBadOutput
}

// parseWriting decodes and validates one reply. It returns the underlying
// problem rather than ErrBadOutput so the caller can quote it back to the model.
func parseWriting(raw string, req WritingRequest) (models.Evaluation, error) {
	var payload evaluationPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return models.Evaluation{}, fmt.Errorf("the reply was not valid JSON (%w)", err)
	}
	return validateWriting(payload, req)
}

// add folds one attempt's usage into a running total. Provider, model and
// prompt version are the same for every attempt, so the last one wins.
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

// validateWriting checks the model's reply before any of it is stored.
//
// Anything out of range is rejected outright rather than clamped: a score the
// model could not produce correctly is not one to guess at.
func validateWriting(p evaluationPayload, req WritingRequest) (models.Evaluation, error) {
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

	// An off-scale score is rejected, never converted onto the right scale.
	// Reading a stray 8.5 as an IELTS band and mapping it to PTE assumes what
	// the model meant, and any linear PTE/IELTS mapping overstates the middle
	// of the range badly against the published concordance. The retry in
	// EvaluateWriting asks for a corrected score instead.
	score := p.EstimatedScore.Value
	if score != nil {
		if *score < req.MinScore || *score > req.MaxScore {
			return models.Evaluation{}, fmt.Errorf("score %.2f outside %.1f-%.1f for %s", *score, req.MinScore, req.MaxScore, req.Exam)
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

	// Sentence feedback must quote text the learner actually wrote. Dropping
	// unmatched entries stops the model from "correcting" invented sentences,
	// which is confusing and makes the whole report look untrustworthy.
	haystack := normaliseSpace(req.LearnerText)
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
		Exam:              req.Exam,
		Skill:             models.SkillWriting,
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

// exampleScore picks a plausible value on this exam's own scale for the worked
// example in the system prompt.
//
// A fixed number cannot work here. 7.0 reads as a sensible IELTS band and as a
// nonsense PTE score, and an example on the wrong scale drags the model's real
// answer onto that scale with it — which is exactly how PTE evaluations started
// coming back as 8.5 and getting thrown out by validateWriting.
func exampleScore(min, max float64) float64 {
	v := min + 0.75*(max-min)
	// Wide point scales like PTE's 10-90 do not use fractions; narrow band
	// scales like IELTS's 0-9 move in half points.
	if max-min > 20 {
		return math.Round(v)
	}
	return math.Round(v*2) / 2
}

func writingUserPrompt(req WritingRequest) string {
	return fmt.Sprintf("Task: %s\n\nPrompt:\n%s\n\nLearner's response:\n%s",
		req.TaskName, req.Prompt, req.LearnerText)
}

// normaliseSpace collapses runs of whitespace so a quote that differs only in
// line breaks still matches the learner's text.
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
