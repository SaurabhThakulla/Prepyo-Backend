package ai

import (
	"fmt"
	"math"
	"strings"

	"github.com/prepyo/backend/internal/models"
)

// PTE trait scales from Pearson's PTE Academic Test Taker Score Guide (July
// 2025), section 7 "Scoring Criteria". The model rates each trait on
// Pearson's own points; Prepyo turns the points into the 10-90 estimate, so a
// model cannot drift the overall score away from its own trait ratings.

type pteTrait struct {
	Name string
	Max  int
	// Gate marks Content and Form: Pearson gives no points for a response
	// scored 0 on either.
	Gate bool
}

func pteWritingTraits(task string) []pteTrait {
	switch task {
	case pteTaskSummarizeWrittenText:
		return []pteTrait{{"Content", 4, true}, {"Form", 1, true}, {"Grammar", 2, false}, {"Vocabulary", 2, false}}
	case pteTaskSummarizeSpokenText:
		return []pteTrait{{"Content", 4, true}, {"Form", 2, true}, {"Grammar", 2, false}, {"Vocabulary", 2, false}, {"Spelling", 2, false}}
	}
	return []pteTrait{
		{"Content", 6, true}, {"Form", 2, true}, {"Development, Structure and Coherence", 6, false},
		{"Grammar", 2, false}, {"General Linguistic Range", 6, false}, {"Vocabulary Range", 2, false}, {"Spelling", 2, false},
	}
}

// pteSpeakingTraits lists the traits for an AI-rated speaking task. A
// transcript cannot show pronunciation, so that path leaves it out.
func pteSpeakingTraits(task string, withPronunciation bool) []pteTrait {
	switch task {
	case pteTaskAnswerShortQuestion:
		return []pteTrait{{"Content", 1, true}}
	case pteTaskRepeatSentence:
		return withPronunciationTrait([]pteTrait{{"Content", 3, true}, {"Oral Fluency", 5, false}}, withPronunciation)
	case pteTaskRespondToSituation:
		return withPronunciationTrait([]pteTrait{{"Appropriacy", 6, true}, {"Oral Fluency", 5, false}}, withPronunciation)
	}
	return withPronunciationTrait([]pteTrait{{"Content", 6, true}, {"Oral Fluency", 5, false}}, withPronunciation)
}

func withPronunciationTrait(traits []pteTrait, with bool) []pteTrait {
	if with {
		traits = append(traits, pteTrait{"Pronunciation", 5, false})
	}
	return traits
}

// pteTraitList names the traits and their points for the prompt.
func pteTraitList(traits []pteTrait) string {
	parts := make([]string, len(traits))
	for i, t := range traits {
		parts[i] = fmt.Sprintf("%s (0-%d)", t.Name, t.Max)
	}
	return strings.Join(parts, ", ")
}

// pteFormFromLength is Form as Pearson defines it by length alone; -1 when
// length does not decide it (Summarize Written Text inside 5-75 words still
// needs the one-sentence check).
func pteFormFromLength(task string, words int) int {
	switch task {
	case pteTaskSummarizeWrittenText:
		if words < 5 || words > 75 {
			return 0
		}
		return -1
	case pteTaskSummarizeSpokenText:
		return lengthBand(words, 50, 70, 40, 100)
	}
	return lengthBand(words, 200, 300, 120, 380)
}

func lengthBand(words, lo, hi, outerLo, outerHi int) int {
	switch {
	case words >= lo && words <= hi:
		return 2
	case words >= outerLo && words <= outerHi:
		return 1
	}
	return 0
}

// ptePointsScore checks the reply's traits against Pearson's scales, applies
// the Content/Form gate and returns the 10-90 estimate: 10 plus 80 times the
// share of points earned, the same share the PTE mock adds up
// (ptemock.FractionOfEvaluation). Pearson does not publish its points-to-scale
// conversion, so the straight line is Prepyo's convention. formCap, when not
// negative, is the most Form can be given the response's length.
func ptePointsScore(criteria []models.EvaluationCriterion, traits []pteTrait, formCap int) (float64, []models.EvaluationCriterion, error) {
	if len(criteria) != len(traits) {
		return 0, nil, fmt.Errorf("return exactly these %d traits: %s", len(traits), pteTraitList(traits))
	}
	byName := map[string]models.EvaluationCriterion{}
	for _, c := range criteria {
		byName[traitKey(c.Name)] = c
	}
	out := make([]models.EvaluationCriterion, 0, len(traits))
	gated := false
	for _, t := range traits {
		c, ok := byName[traitKey(t.Name)]
		if !ok {
			return 0, nil, fmt.Errorf("trait %q is missing; return exactly: %s", t.Name, pteTraitList(traits))
		}
		if c.MaxScore != float64(t.Max) || c.Score != math.Trunc(c.Score) || c.Score < 0 || c.Score > float64(t.Max) {
			return 0, nil, fmt.Errorf("%s must be a whole number of points from 0 to %d with maxScore %d", t.Name, t.Max, t.Max)
		}
		if t.Name == "Form" && formCap >= 0 && c.Score > float64(formCap) {
			c.Score = float64(formCap)
		}
		c.Name = t.Name
		if t.Gate && c.Score == 0 {
			gated = true
		}
		out = append(out, c)
	}
	sum, total := 0.0, 0.0
	for i, t := range traits {
		if gated {
			out[i].Score = 0
		}
		sum += out[i].Score
		total += float64(t.Max)
	}
	return math.Round(10 + 80*sum/total), out, nil
}

func traitKey(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.ReplaceAll(name, "&", "and")), " "))
}
