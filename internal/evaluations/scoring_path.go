package evaluations

import (
	"context"
	"encoding/base64"
	"log/slog"
	"strings"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/models"
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
		// Gemini rates delivery while it also transcribes; without the rating, scoring falls back to assumed values.
		var rating *ai.DeliveryRating
		var ratingUsage ai.Usage
		rated := make(chan struct{})
		go func() {
			defer close(rated)
			if !rate {
				return
			}
			r, u, err := g.RateDelivery(ctx, audio, format, material.expected)
			ratingUsage = u
			if err != nil {
				slog.Warn("delivery rating failed, using assumed values", "error", err)
				return
			}
			rating = &r
		}()
		transcript, heard, usage, err := g.TranscribeMarked(ctx, audio, format)
		<-rated
		if err != nil {
			return speakingScore{}, err
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
// rating whenever a delivery model is set.
func ScoreSpeaking(ctx context.Context, g *ai.Gateway, q models.Question, v models.ExamVersion,
	audio []byte, format string, delivery scoring.Delivery, seconds int) (models.Evaluation, ai.Usage, error) {
	s, err := scoreSpeaking(ctx, g, q, v, audio, format, delivery, seconds, g.DeliveryRatingAvailable())
	return s.evaluation, s.usage, err
}
