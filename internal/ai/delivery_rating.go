package ai

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
)

// DeliveryPromptVersion is stamped on usage for audio-rated Pronunciation and Oral Fluency.
const DeliveryPromptVersion = "pte.delivery.v3"

// DeliveryRating is Pronunciation and Oral Fluency on Pearson's 10-90 scale, heard
// from the audio, with the given text's words (numbered from 1) the speaker said
// less than well: Average and Bad.
type DeliveryRating struct {
	Pronunciation         int    `json:"pronunciation"`
	PronunciationFeedback string `json:"pronunciationFeedback"`
	Fluency               int    `json:"oralFluency"`
	FluencyFeedback       string `json:"oralFluencyFeedback"`
	Average               []int  `json:"averageWords"`
	Bad                   []int  `json:"badWords"`
	// WordScored marks a rating from the pronunciation scorer (per-word sounds,
	// no fluency): Pronunciation then comes from the word grades alone.
	WordScored bool `json:"-"`
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
			{Type: "text", Text: "The text the learner was asked to say, word by word:\n" + numberedWords(expected) + "\nRate the recording."},
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
		rating, problem := parseDeliveryRating(raw, len(strings.Fields(expected)))
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

// numberedWords lists the text's words as "1 The 2 best ...", so the rating can name a word unambiguously.
func numberedWords(text string) string {
	var b strings.Builder
	for i, w := range strings.Fields(text) {
		fmt.Fprintf(&b, "%d %s ", i+1, w)
	}
	return strings.TrimSpace(b.String())
}

// wordNumbers keeps valid word numbers (1..n) once each, dropping any in skip.
func wordNumbers(in []int, n int, skip map[int]bool) []int {
	seen := map[int]bool{}
	var out []int
	for _, w := range in {
		if w >= 1 && w <= n && !seen[w] && !skip[w] {
			seen[w] = true
			out = append(out, w)
		}
	}
	return out
}

func parseDeliveryRating(raw string, words int) (DeliveryRating, error) {
	var r DeliveryRating
	if err := decodeFirstJSON(raw, &r); err != nil {
		return DeliveryRating{}, fmt.Errorf("the reply was not valid JSON (%w)", err)
	}
	if r.Pronunciation < 10 || r.Pronunciation > 90 || r.Fluency < 10 || r.Fluency > 90 {
		return DeliveryRating{}, fmt.Errorf("pronunciation and oralFluency must be whole numbers from 10 to 90")
	}
	if strings.TrimSpace(r.PronunciationFeedback) == "" || strings.TrimSpace(r.FluencyFeedback) == "" {
		return DeliveryRating{}, fmt.Errorf("pronunciationFeedback and oralFluencyFeedback must not be empty")
	}
	r.Bad = wordNumbers(r.Bad, words, nil)
	bad := map[int]bool{}
	for _, w := range r.Bad {
		bad[w] = true
	}
	r.Average = wordNumbers(r.Average, words, bad)
	return r, nil
}

var deliverySystemPrompt = `You are a strict PTE Academic examiner rating one spoken response to a Read Aloud or Repeat Sentence task, the way Pearson's automated scoring does.
Listen to the audio recording. Rate HOW it was spoken. Which words were said is checked separately.

Judge every word of the given text against Pearson's top level: each sound produced the way an expert speaker would in connected speech, correct word stress, natural linking.
- good: said clearly and correctly, with the right stress, as an expert speaker would.
- average: understandable, but with a noticeable flaw: a distorted vowel or consonant, a dropped or added ending, wrong word stress, or over-separated, syllable-by-syllable delivery.
- bad: mispronounced so a listener would hesitate or mishear it, swallowed, mumbled, or not said.
Judge each word on its own; never mark a word good by default. A learner who has not practised usually has a quarter or more of the words average or bad. A regional accent is fine only where every sound in the word is still clear and correctly stressed.

Pronunciation (10-90), on Pearson's scale:
- 79-90: native-like; all sounds clear, stress fully appropriate, natural connected speech. Very few average words, no bad ones.
- 65-78: advanced; most sounds correct, a few consistent errors in sounds or stress.
- 50-64: good; clearly understandable but many words average, some bad.
- 30-49: intermediate; frequent distortions, a listener must work to follow.
- 10-29: mostly hard to understand.

Oral Fluency (10-90), on Pearson's scale:
- 79-90: smooth rhythm and phrasing, natural sentence stress and intonation, no hesitations, repetitions or false starts.
- 65-78: acceptable rhythm with some uneven phrasing or one or two hesitations.
- 50-64: uneven or flat, monotone delivery; word-by-word reading; several hesitations or self-corrections.
- 30-49: choppy, frequent hesitations, long pauses or repeated words.
- 10-29: halting, isolated words, mostly silence.
Speed alone is not fluency: a fast, flat reading with no phrasing is at most 64.

Reply with JSON only, giving word numbers from the numbered text:
{"pronunciation": <whole number 10-90>, "pronunciationFeedback": "<one or two sentences on the sounds or stress to fix>", "oralFluency": <whole number 10-90>, "oralFluencyFeedback": "<one or two sentences on rhythm, phrasing and intonation>", "averageWords": [<numbers of the words said with a flaw>], "badWords": [<numbers of the words mispronounced or not said>]}`
