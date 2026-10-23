package evaluations

import (
	"context"
	"encoding/base64"
	"log/slog"
	"sort"
	"strings"
	"unicode"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/pronounce"
	"github.com/prepyo/backend/internal/scoring"
)

// The marking the app runs on an answer, without billing, storage or rewards.
// The service and cmd/gradebench both call it, so the bench grades exactly as
// learners are graded.

// writingRequest is what the grader is sent for a written answer.
func writingRequest(q models.Question, v models.ExamVersion, text string, wordCount, minimumWords int) ai.WritingRequest {
	return ai.WritingRequest{
		Exam:           q.Exam,
		TypeID:         q.TypeID,
		TaskName:       q.TypeName,
		Prompt:         q.Prompt,
		FigureData:     q.FigureData,
		ContextPassage: gradingSource(q.ContextPassage, q.AudioTranscript),
		LearnerText:    text,
		WordCount:      wordCount,
		MinimumWords:   minimumWords,
		MinScore:       v.MinScore,
		MaxScore:       v.MaxScore,
	}
}

// ScorePTEWriting grades a PTE written answer as EvaluateWriting does.
func ScorePTEWriting(ctx context.Context, g *ai.Gateway, q models.Question, v models.ExamVersion, text string) (models.Evaluation, ai.Usage, error) {
	// The app sets a word count for IELTS only.
	return g.EvaluateWriting(ctx, writingRequest(q, v, strings.TrimSpace(text), 0, 0))
}

type speakingRoute int

const (
	routeOpen        speakingRoute = iota // model grades the recording
	routeVerbatim                         // Read Aloud, Repeat Sentence: aligned against the text
	routeShortAnswer                      // Answer Short Question: matched against the key
)

func speakingRouteOf(q models.Question) speakingRoute {
	switch {
	case strings.TrimSpace(speakingMaterialOf(q).expected) != "":
		return routeVerbatim
	case matchesShortAnswer(q):
		return routeShortAnswer
	}
	return routeOpen
}

type speakingScore struct {
	evaluation models.Evaluation
	usage      ai.Usage
	units      int
	source     string
}

// scoreSpeaking marks one recording. rate asks for the audio delivery rating
// on Read Aloud and Repeat Sentence (the caller has checked it is paid for).
func scoreSpeaking(ctx context.Context, g *ai.Gateway, q models.Question, v models.ExamVersion,
	audio []byte, format string, delivery scoring.Delivery, seconds int, rate bool) (speakingScore, error) {
	material := speakingMaterialOf(q)
	switch speakingRouteOf(q) {
	case routeVerbatim:
		// Delivery is rated by the pronunciation scorer once the transcript shows which words were said,
		// or by Gemini alongside the transcription (also the scorer's fallback). No rating: assumed values.
		var rating *ai.DeliveryRating
		var ratingUsage ai.Usage
		geminiRate := func() {
			r, u, err := g.RateDelivery(ctx, audio, format, material.expected)
			ratingUsage = u
			if err != nil {
				slog.Warn("delivery rating failed, using assumed values", "error", err)
				return
			}
			rating = &r
		}
		useScorer := rate && pronunciationScorer != nil
		rated := make(chan struct{})
		go func() {
			defer close(rated)
			if rate && !useScorer {
				geminiRate()
			}
		}()
		transcript, heard, usage, err := g.TranscribeMarked(ctx, audio, format)
		<-rated
		if err != nil {
			return speakingScore{}, err
		}
		if useScorer {
			text, positions := scorerWords(material.expected, transcript)
			res, err := pronunciationScorer.Score(ctx, audio, text)
			if err == nil {
				r := ratingFromWords(positions, res)
				rating, ratingUsage = &r, ai.Usage{Provider: "prepyo", Model: "lambda-gop"}
			} else {
				slog.Warn("pronunciation scorer failed, using the Gemini rating", "error", err)
				if g.DeliveryRatingAvailable() {
					geminiRate()
				}
			}
		}
		// Pauses come from the browser; the transcription has no timings.
		if heard.Measured() {
			delivery = heard
		}
		delivery.WordsSpoken = len(strings.Fields(strings.ReplaceAll(transcript, scoring.PauseMark, " ")))
		evaluation := verbatimEvaluation(q, v, material.expected, strings.TrimSpace(transcript), delivery, rating)
		usage.Model, usage.PromptVersion = usage.Model+"+verbatim-alignment", verbatimScoringVersion
		units := billing.WordMatchUnits
		if rating != nil {
			usage.Model += "+" + ratingUsage.Model
			usage.PromptTokens += ratingUsage.PromptTokens
			usage.CompletionTokens += ratingUsage.CompletionTokens
			usage.CostUSD += ratingUsage.CostUSD
			units = billing.DeliveryRatedUnits
		}
		slog.Info("read aloud scored", "deliveryModel", ratingUsage.Model, "promptTokens", ratingUsage.PromptTokens,
			"completionTokens", ratingUsage.CompletionTokens, "deliveryCostUSD", ratingUsage.CostUSD, "rated", rating != nil)
		return speakingScore{evaluation, usage, units, "practice-read-repeat"}, nil

	case routeShortAnswer:
		// Answer Short Question is right or wrong: the transcript is matched against the answer key, not graded by a model.
		transcript, _, usage, err := g.TranscribeMarked(ctx, audio, format)
		if err != nil {
			return speakingScore{}, err
		}
		usage.Model, usage.PromptVersion = usage.Model+"+short-answer-match", shortAnswerScoringVersion
		return speakingScore{shortAnswerEvaluation(q, v, transcript), usage, billing.WordMatchUnits, "practice-short-answer"}, nil
	}

	evaluation, usage, err := g.EvaluateSpeaking(ctx, ai.SpeakingRequest{
		Exam:            q.Exam,
		TaskName:        q.TypeName,
		Prompt:          q.Prompt,
		ExpectedText:    material.expected,
		SourceText:      material.source,
		ReferenceAnswer: material.reference,
		AudioBase64:     base64.StdEncoding.EncodeToString(audio),
		AudioFormat:     format,
		DurationSeconds: seconds,
		Delivery:        delivery,
		MinScore:        v.MinScore,
		MaxScore:        v.MaxScore,
	})
	if err != nil {
		return speakingScore{}, err
	}
	return speakingScore{evaluation, usage, billing.UnitsPerGrading, "practice-speaking"}, nil
}

// ScoreSpeaking marks a recording as EvaluateSpeaking does, with the delivery
// rating whenever a delivery model or the pronunciation scorer is set.
func ScoreSpeaking(ctx context.Context, g *ai.Gateway, q models.Question, v models.ExamVersion,
	audio []byte, format string, delivery scoring.Delivery, seconds int) (models.Evaluation, ai.Usage, error) {
	s, err := scoreSpeaking(ctx, g, q, v, audio, format, delivery, seconds, DeliveryRatingAvailable(g))
	return s.evaluation, s.usage, err
}

// pronunciationScorer is the self-hosted scorer; nil leaves Read Aloud and Repeat Sentence on Gemini.
var pronunciationScorer *pronounce.Client

// UsePronunciationScorer routes Read Aloud and Repeat Sentence delivery to the scorer, Gemini as fallback.
func UsePronunciationScorer(c *pronounce.Client) { pronunciationScorer = c }

// DeliveryRatingAvailable reports whether Read Aloud and Repeat Sentence can be rated for delivery.
func DeliveryRatingAvailable(g *ai.Gateway) bool {
	return pronunciationScorer != nil || g.DeliveryRatingAvailable()
}

// scorerWords is the text the scorer should look for and, for each of its words, the
// given text's word index. When the transcript shows words left out, only the words
// said are sent: forcing missing words onto the audio would misplace the ones that were said.
func scorerWords(expected, transcript string) (string, []int) {
	tokens := strings.Fields(expected)
	alignment := scoring.ScoreVerbatim(expected, transcript)
	omitted := alignment.Expected - alignment.Matched - alignment.Inserted
	var positions []int
	var words []string
	for i, token := range tokens {
		if !strings.ContainsFunc(token, func(c rune) bool { return c < 128 && (unicode.IsLetter(c) || unicode.IsDigit(c)) }) {
			continue // the scorer skips tokens with no letter or digit
		}
		if omitted >= 2 && i < len(alignment.Marks) && !alignment.Marks[i].Said {
			continue
		}
		positions = append(positions, i)
		words = append(words, token)
	}
	if len(words) == 0 {
		return expected, nil
	}
	return strings.Join(words, " "), positions
}

// ratingFromWords turns the scorer's per-word labels into a rating on the given text's word numbers.
func ratingFromWords(positions []int, res pronounce.Result) ai.DeliveryRating {
	r := ai.DeliveryRating{WordScored: true, Pronunciation: res.Pronunciation}
	var weak []pronounce.Word
	for k, w := range res.Words {
		if k >= len(positions) {
			break
		}
		i := positions[k]
		switch w.Label {
		case "bad":
			r.Bad = append(r.Bad, i+1)
		case "average":
			r.Average = append(r.Average, i+1)
		}
		if w.Label != "good" {
			weak = append(weak, w)
		}
	}
	sort.Slice(weak, func(a, b int) bool { return weak[a].Score < weak[b].Score })
	r.PronunciationFeedback = "Every word was pronounced clearly."
	if len(weak) > 0 {
		names := make([]string, 0, 5)
		for _, w := range weak[:min(5, len(weak))] {
			names = append(names, strings.Trim(w.Word, ".,;:!?\"'()"))
		}
		r.PronunciationFeedback = "Practise these words, said least clearly: " + strings.Join(names, ", ") + "."
	}
	return r
}
