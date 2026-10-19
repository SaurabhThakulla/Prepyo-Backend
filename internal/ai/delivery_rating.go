package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// DeliveryPromptVersion is stamped on usage for audio-rated Pronunciation and Oral Fluency.
const DeliveryPromptVersion = "pte.delivery.v2"

// DeliveryRating is Pronunciation and Oral Fluency on Pearson's 10-90 scale, heard from the audio.
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
	if r.Pronunciation < 10 || r.Pronunciation > 90 || r.Fluency < 10 || r.Fluency > 90 {
		return DeliveryRating{}, fmt.Errorf("pronunciation and oralFluency must be whole numbers from 10 to 90")
	}
	if strings.TrimSpace(r.PronunciationFeedback) == "" || strings.TrimSpace(r.FluencyFeedback) == "" {
		return DeliveryRating{}, fmt.Errorf("pronunciationFeedback and oralFluencyFeedback must not be empty")
	}
	if len(r.Unclear) > 12 {
		r.Unclear = r.Unclear[:12]
	}
	return r, nil
}

var deliverySystemPrompt = `You are an experienced PTE Academic examiner rating one spoken response to a Read Aloud or Repeat Sentence task.
Listen to the audio recording. Rate only HOW it was spoken (delivery), never which words were said: content accuracy is marked separately.

Rate Oral Fluency (10-90) and Pronunciation (10-90) directly on Pearson's official 10-90 PTE scale.

Oral Fluency (10-90):
- 79-90: Smooth, natural exam rhythm (120-150 words per minute); appropriate phrasing; no mid-sentence hesitations or false starts.
- 65-78: Sustained speech with consistent rhythm; natural breath pauses between clauses; minor hesitations without stopping the overall flow.
- 50-64: Noticeable uneven rhythm or hesitations; reading word-by-word; several mid-sentence pauses.
- 30-49: Frequent hesitations, repeated words, choppy phrasing, or long silences.
- 10-29: Extremely slow, halting, isolated single words, or mostly silent.

Pronunciation (10-90):
- 79-90: Every word and sound easily understood by any listener; clear vowel and consonant articulation; standard word and sentence stress.
- 65-78: Easily intelligible throughout; vowels and consonants are clear. Natural regional accents (e.g. South Asian, Latin American, East Asian) are completely acceptable and NOT penalized as long as words are clear.
- 50-64: Speech is mostly intelligible, but some vowel distortions, omitted word endings, or non-standard stress make a few words unclear.
- 30-49: Many distorted sounds; listener must strain to follow; about one third of words hard to decipher.
- 10-29: More than half of the words unintelligible or inaudible.

Reply with JSON only:
{"pronunciation": <whole number 10-90>, "pronunciationFeedback": "<one or two sentences on sounds or word stress to polish>", "oralFluency": <whole number 10-90>, "oralFluencyFeedback": "<one or two sentences on rhythm, phrasing and pacing>", "unclearWords": ["<words from the text that were mispronounced or hard to understand>"]}`
