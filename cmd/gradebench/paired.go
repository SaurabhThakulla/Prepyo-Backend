package main

// Paired mode: answers you also submitted on Alfa PTE, graded here through the
// app's own scoring path (evaluations.ScoreSpeaking / ScorePTEWriting), with
// the gap to Alfa's overall and trait scores printed per task and trait.

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/evaluations"
	"github.com/prepyo/backend/internal/models"
)

type pairedSample struct {
	ID     string `json:"id"`
	AlfaID string `json:"alfaId"`
	Audio  string `json:"audio"`  // WAV recording, for speaking tasks
	Text   string `json:"text"`   // typed answer, for writing tasks
	Source string `json:"source"` // lecture transcript or image description when the question has none
	// Alfa's scores as shown: numbers on 10-90, or "2/3" for raw trait points.
	Alfa map[string]any `json:"alfa"`
}

type alfaQuestion struct {
	AlfaID          string          `json:"alfaId"`
	ID              string          `json:"id"`
	Skill           string          `json:"skill"`
	TypeID          string          `json:"typeId"`
	TypeName        string          `json:"typeName"`
	Title           string          `json:"title"`
	Prompt          string          `json:"prompt"`
	ContextPassage  string          `json:"contextPassage"`
	AudioTranscript string          `json:"audioTranscript"`
	ModelAnswer     string          `json:"modelAnswer"`
	FigureData      string          `json:"figureData"`
	CorrectAnswers  json.RawMessage `json:"correctAnswers"`
}

type pairedResult struct {
	ID         string             `json:"id"`
	Task       string             `json:"task"`
	Run        int                `json:"run"`
	Alfa       map[string]float64 `json:"alfa"`
	Ours       map[string]float64 `json:"ours"`
	Delivery   string             `json:"delivery,omitempty"`
	Transcript string             `json:"transcript,omitempty"`
	Model      string             `json:"model,omitempty"`
	CostUSD    float64            `json:"costUsd"`
	Error      string             `json:"error,omitempty"`
}

var pteVersion = models.ExamVersion{ID: "pte-2026-01", Exam: models.ExamPTE, MinScore: 10, MaxScore: 90, ScoreStep: 1}

func runPaired(g *ai.Gateway, dir, pattern string, runs, workers int, out string) {
	questions, err := loadAlfaQuestions(filepath.Join(dir, "paired", "alfa_questions.json"))
	if err != nil {
		fail(err)
	}
	files, _ := filepath.Glob(pattern)
	if len(files) == 0 {
		fail(fmt.Errorf("no paired sample files match %s", pattern))
	}
	var samples []pairedSample
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fail(err)
		}
		var list []pairedSample
		if err := json.Unmarshal(data, &list); err != nil {
			fail(fmt.Errorf("%s: %w", f, err))
		}
		for i := range list {
			if list[i].Audio != "" && !filepath.IsAbs(list[i].Audio) {
				list[i].Audio = filepath.Join(filepath.Dir(f), list[i].Audio)
			}
		}
		samples = append(samples, list...)
	}
	fmt.Printf("paired: %d answers x %d run(s), delivery model %v\n\n", len(samples), runs, g.DeliveryRatingAvailable())

	type job struct {
		s   pairedSample
		run int
	}
	jobs := make(chan job)
	var mu sync.Mutex
	var results []pairedResult
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := gradePaired(g, questions, j.s)
				r.Run = j.run
				mu.Lock()
				results = append(results, r)
				fmt.Printf("  %-24s run %d  %-28s alfa %5s  ours %5s  %s%s\n", r.ID, r.Run, r.Task,
					fmtNum(r.Alfa["overall"]), fmtNum(r.Ours["overall"]), r.Delivery, r.Error)
				mu.Unlock()
			}
		}()
	}
	for run := 1; run <= runs; run++ {
		for _, s := range samples {
			jobs <- job{s, run}
		}
	}
	close(jobs)
	wg.Wait()

	reportPaired(results)
	if out != "" {
		data, _ := json.MarshalIndent(map[string]any{"ranAt": time.Now().Format(time.RFC3339), "results": results}, "", "  ")
		if err := os.WriteFile(out, data, 0o644); err != nil {
			fail(err)
		}
		fmt.Printf("\nresults written to %s\n", out)
	}
}

func gradePaired(g *ai.Gateway, questions map[string]alfaQuestion, s pairedSample) pairedResult {
	r := pairedResult{ID: firstNonEmpty(s.ID, s.AlfaID), Alfa: map[string]float64{}, Ours: map[string]float64{}}
	for name, v := range s.Alfa {
		if score, ok := alfaScore(v); ok {
			r.Alfa[traitKey(name)] = score
		}
	}
	aq, ok := questions[s.AlfaID]
	if !ok {
		r.Error = "unknown alfaId " + s.AlfaID
		return r
	}
	r.Task = aq.TypeName
	q := aq.question(s.Source)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	var ev models.Evaluation
	var err error
	switch {
	case q.Skill == models.SkillSpeaking:
		if s.Audio == "" {
			r.Error = "speaking task needs \"audio\""
			return r
		}
		audio, delivery, seconds, perr := prepareRecording(s.Audio)
		if perr != nil {
			r.Error = perr.Error()
			return r
		}
		r.Delivery = fmt.Sprintf("[%ds, %d pauses, longest %.1fs, voiced %.0f%%] ",
			delivery.DurationSeconds, delivery.PauseCount, delivery.LongestPauseSeconds, delivery.SpeakingRatio*100)
		needsSource := strings.Contains(q.TypeID, "retell") || strings.Contains(q.TypeID, "describe-image")
		if needsSource && strings.TrimSpace(q.ContextPassage+q.AudioTranscript) == "" {
			r.Delivery += "(no source: content judged without the lecture or image) "
		}
		var usage ai.Usage
		ev, usage, err = evaluations.ScoreSpeaking(ctx, g, q, pteVersion, audio, "wav", delivery, seconds)
		r.Model, r.CostUSD = usage.Model, usage.CostUSD
	case strings.TrimSpace(s.Text) != "" && (q.Skill == models.SkillWriting || q.TypeID == "summarize-spoken-text"):
		ev, _, err = evaluations.ScorePTEWriting(ctx, g, q, pteVersion, s.Text)
	default:
		r.Error = "needs \"text\" (writing) or \"audio\" (speaking); reading/listening items are rule-scored"
		return r
	}
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Transcript = ev.Transcript
	if ev.EstimatedScore != nil {
		r.Ours["overall"] = *ev.EstimatedScore
	}
	for _, c := range ev.Criteria {
		r.Ours[traitKey(c.Name)] = c.Score
	}
	return r
}

func (aq alfaQuestion) question(source string) models.Question {
	q := models.Question{
		ID: aq.ID, ExamVersionID: pteVersion.ID, Exam: models.ExamPTE, Skill: models.SkillType(aq.Skill),
		TypeID: aq.TypeID, TypeName: aq.TypeName, Title: aq.Title, Prompt: aq.Prompt,
		ContextPassage: aq.ContextPassage, AudioTranscript: aq.AudioTranscript,
		ModelAnswer: aq.ModelAnswer, FigureData: aq.FigureData,
	}
	_ = json.Unmarshal(aq.CorrectAnswers, &q.CorrectAnswers)
	if source != "" {
		q.AudioTranscript = source
	}
	return q
}

func loadAlfaQuestions(path string) (map[string]alfaQuestion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var list []alfaQuestion
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	byID := map[string]alfaQuestion{}
	for _, q := range list {
		byID[q.AlfaID] = q
	}
	return byID, nil
}

// alfaScore reads a 10-90 number, or "a/b" trait points put on 10-90.
func alfaScore(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		if a, b, ok := strings.Cut(x, "/"); ok {
			num, e1 := strconv.ParseFloat(strings.TrimSpace(a), 64)
			den, e2 := strconv.ParseFloat(strings.TrimSpace(b), 64)
			if e1 == nil && e2 == nil && den > 0 {
				if den == 90 {
					return num, true
				}
				return math.Round(10 + 80*num/den), true
			}
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return n, err == nil
	}
	return 0, false
}

var nonLetters = regexp.MustCompile(`[^a-z]+`)

// traitKey puts Alfa's and Prepyo's trait names on one spelling.
func traitKey(name string) string {
	k := strings.TrimSpace(nonLetters.ReplaceAllString(strings.ToLower(name), " "))
	switch k {
	case "fluency", "oral fluency":
		return "oral fluency"
	case "dsc", "written discourse", "development structure coherence", "development structure and coherence":
		return "development structure and coherence"
	case "glr", "linguistic range", "general linguistic range":
		return "general linguistic range"
	case "total", "score", "overall score":
		return "overall"
	}
	return k
}

// reportPaired prints, per task and trait, Alfa's mean, ours, the bias
// (ours minus Alfa) and the mean distance, over every run.
func reportPaired(results []pairedResult) {
	type agg struct{ n, alfa, ours, abs float64 }
	byTask := map[string]map[string]*agg{}
	for _, r := range results {
		if r.Error != "" {
			continue
		}
		traits := byTask[r.Task]
		if traits == nil {
			traits = map[string]*agg{}
			byTask[r.Task] = traits
		}
		for name, a := range r.Alfa {
			o, ok := r.Ours[name]
			if !ok {
				continue
			}
			t := traits[name]
			if t == nil {
				t = &agg{}
				traits[name] = t
			}
			t.n++
			t.alfa += a
			t.ours += o
			t.abs += math.Abs(o - a)
		}
	}
	fmt.Println("\ntask                          trait                                  n   alfa   ours   bias  mean|gap|")
	tasks := make([]string, 0, len(byTask))
	for t := range byTask {
		tasks = append(tasks, t)
	}
	sort.Strings(tasks)
	for _, task := range tasks {
		names := make([]string, 0, len(byTask[task]))
		for n := range byTask[task] {
			names = append(names, n)
		}
		sort.Slice(names, func(i, j int) bool { return names[i] == "overall" || (names[j] != "overall" && names[i] < names[j]) })
		for _, name := range names {
			a := byTask[task][name]
			fmt.Printf("%-29s %-36s %4.0f  %5.1f  %5.1f  %+5.1f  %7.1f\n", task, name, a.n, a.alfa/a.n, a.ours/a.n, (a.ours-a.alfa)/a.n, a.abs/a.n)
		}
	}
	unmatched := map[string]bool{}
	for _, r := range results {
		for name := range r.Alfa {
			if _, ok := r.Ours[name]; !ok && r.Error == "" {
				unmatched[r.Task+": "+name] = true
			}
		}
	}
	for k := range unmatched {
		fmt.Printf("  no matching trait of ours for Alfa's %s\n", k)
	}
}

func fmtNum(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.0f", v)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
