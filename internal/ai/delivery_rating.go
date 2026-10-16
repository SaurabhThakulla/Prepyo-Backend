package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// DeliveryPromptVersion is stamped on usage for audio-rated Pronunciation and Oral Fluency.
const DeliveryPromptVersion = "pte.delivery.v1"

// DeliveryRating is Pronunciation and Oral Fluency on Pearson's 0-5 ladders, heard from the audio.
type DeliveryRating struct {
	Pronunciation         int      `json:"pronunciation"`
	PronunciationFeedback string   `json:"pronunciationFeedback"`
	Fluency               int      `json:"oralFluency"`
	FluencyFeedback       string   `json:"oralFluencyFeedback"`
	Unclear               []string `json:"unclearWords"`
}

// DeliveryRatingAvailable reports whether an audio model is set to rate delivery.
func (g *Gateway) DeliveryRatingAvailable() bool {
	m := strings.ToLower(strings.TrimSpace(g.models.Delivery))
	return m != "" && m != "off" && g.audio.configured()
}

// RateDelivery asks an audio model to rate how a fixed text was spoken; content is not its job.
func (g *Gateway) RateDelivery(ctx context.Context, audio []byte, format, expected string) (DeliveryRating, Usage, error) {
	if !g.DeliveryRatingAvailable() {
		return DeliveryRating{}, Usage{}, fmt.Errorf("%w: no delivery model is configured", ErrUnavailable)
	}
	if !AudioFormats[format] {
		return DeliveryRating{}, Usage{}, fmt.Errorf("%w: audio format %q is not supported", ErrBadOutput, format)
	}
	messages := []chatMessage{
		{Role: "system", Content: deliverySystemPrompt},
		{Role: "user", Content: []contentPart{
			{Type: "text", Text: "The text the learner was asked to read aloud:\n\"\"\"\n" + strings.TrimSpace(expected) + "\n\"\"\"\nRate the recording."},
			{Type: "input_audio", Audio: &audioInput{Data: base64.StdEncoding.EncodeToString(audio), Format: format}},
		}},
	}

	var usage Usage
	for attempt := 1; attempt <= maxValidationAttempts; attempt++ {
		raw, attemptUsage, err := g.complete(ctx, g.audio, g.models.Delivery, DeliveryPromptVersion, messages, true)
		usage.add(attemptUsage)
		if err != nil {
			return DeliveryRating{}, usage, err
		}
		rating, problem := parseDeliveryRating(raw)
		if problem == nil {
			return rating, usage, nil
		}
		g.log.Warn("delivery rating failed validation", "model", g.models.Delivery, "attempt", attempt, "error", problem)
		messages = append(messages,
			chatMessage{Role: "assistant", Content: raw},
			chatMessage{Role: "user", Content: "That reply was rejected: " + problem.Error() + ". Send the complete JSON again."})
	}
	return DeliveryRating{}, usage, ErrBadOutput
}

func parseDeliveryRating(raw string) (DeliveryRating, error) {
	var r DeliveryRating
	if err := json.Unmarshal([]byte(unfence(raw)), &r); err != nil {
		return DeliveryRating{}, fmt.Errorf("the reply was not valid JSON (%w)", err)
	}
	if r.Pronunciation < 0 || r.Pronunciation > 5 || r.Fluency < 0 || r.Fluency > 5 {
		return DeliveryRating{}, fmt.Errorf("pronunciation and oralFluency must be whole numbers from 0 to 5")
	}
	if strings.TrimSpace(r.PronunciationFeedback) == "" || strings.TrimSpace(r.FluencyFeedback) == "" {
		return DeliveryRating{}, fmt.Errorf("pronunciationFeedback and oralFluencyFeedback must not be empty")
	}
	if len(r.Unclear) > 12 {
		r.Unclear = r.Unclear[:12]
	}
	return r, nil
}

var deliverySystemPrompt = `You are a strict PTE Academic examiner rating one spoken response to a Read Aloud or Repeat Sentence task.
Listen to the audio itself. Rate only HOW it was spoken, never which words were said: content is marked elsewhere.

Oral Fluency (0-5), Pearson's scale. Judge rhythm, phrasing, stress and smoothness across the whole response; reading word by word, uneven pacing, hesitations, repetitions and false starts all count against it.
` + pteFluencyLadder + `

Pronunciation (0-5), Pearson's scale. Judge vowels, consonants, word stress and sentence stress against what a wide range of listeners would understand.
An accent is not a fault on its own; lower the score for sounds, stress or intonation that make words harder to understand or that the listener has to adjust to.
` + ptePronunciationLadder + `

Give 4 or 5 only when the recording clearly meets that band; when between two bands, give the lower.

Reply with JSON only:
{"pronunciation": <0-5>, "pronunciationFeedback": "<one or two sentences naming the sounds or words to fix>", "oralFluency": <0-5>, "oralFluencyFeedback": "<one or two sentences on rhythm, phrasing and pauses>", "unclearWords": ["<words from the given text that were mispronounced or hard to understand>"]}`
