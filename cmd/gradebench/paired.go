package main

// Paired mode: evaluates PTE benchmark answers against reference scores
// (from human examiners or benchmark rubrics) using the application's scoring engine
// (evaluations.ScoreSpeaking / ScorePTEWriting), and reports bias and variance per trait.

import (
	"context"
	"encoding/json"
	"errors"
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
	ID       string             `json:"id"`
	RefID    string             `json:"refId"`
	LegacyID string             `json:"alfaId"`
	Audio    string             `json:"audio"`
	Text     string             `json:"text"`
	Source   string             `json:"source"`
	Marks    map[string]any     `json:"alfa"`
	Scores   map[string]any     `json:"scores"`
	Marker   string             `json:"marker"`
	Question *benchmarkQuestion `json:"question"`
}

type benchmarkQuestion struct {
	RefID           string          `json:"alfaId"`
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
	Marker     string             `json:"marker"`
	Task       string             `json:"task"`
	Run        int                `json:"run"`
	Ref        map[string]float64 `json:"reference"`
	Ours       map[string]float64 `json:"ours"`
	Delivery   string             `json:"delivery,omitempty"`
	Transcript string             `json:"transcript,omitempty"`
	Model      string             `json:"model,omitempty"`
	CostUSD    float64            `json:"costUsd"`
	Error      string             `json:"error,omitempty"`
}

var pteVersion = models.ExamVersion{ID: "pte-2026-01", Exam: models.ExamPTE, MinScore: 10, MaxScore: 90, ScoreStep: 1}

func runPaired(g *ai.Gateway, dir, pattern string, runs, workers int, out string) {
	questions, err := loadBenchmarkQuestions(
		filepath.Join(dir, "paired", "benchmark_questions.json"),
		filepath.Join(dir, "paired", "alfa_questions.json"),
	)
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
				fmt.Printf("  %-24s run %d  %-28s %-7s %5s  ours %5s  %s%s\n", r.ID, r.Run, r.Task, r.Marker,
					fmtNum(r.Ref["overall"]), fmtNum(r.Ours["overall"]), r.Delivery, r.Error)
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

func gradePaired(g *ai.Gateway, questions map[string]benchmarkQuestion, s pairedSample) pairedResult {
	id := firstNonEmpty(s.ID, s.RefID, s.LegacyID)
	r := pairedResult{
		ID:     id,
		Marker: firstNonEmpty(s.Marker, "reference"),
		Ref:    map[string]float64{},
		Ours:   map[string]float64{},
	}
	marks := s.Scores
	if len(marks) == 0 {
		marks = s.Marks
	}
	for name, v := range marks {
		if score, ok := parseReferenceScore(v); ok {
			r.Ref[traitKey(name)] = score
		}
	}
	bq, ok := questions[id]
	if s.Question != nil {
		bq, ok = *s.Question, true
	}
	if !ok {
		r.Error = "unknown question: " + id + " (provide question definition inline)"
		return r
	}
	r.Task = bq.TypeName
	q := bq.question(s.Source)
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

func (bq benchmarkQuestion) question(source string) models.Question {
	q := models.Question{
		ID: bq.ID, ExamVersionID: pteVersion.ID, Exam: models.ExamPTE, Skill: models.SkillType(bq.Skill),
		TypeID: bq.TypeID, TypeName: bq.TypeName, Title: bq.Title, Prompt: bq.Prompt,
		ContextPassage: bq.ContextPassage, AudioTranscript: bq.AudioTranscript,
		ModelAnswer: bq.ModelAnswer, FigureData: bq.FigureData,
	}
	_ = json.Unmarshal(bq.CorrectAnswers, &q.CorrectAnswers)
	if source != "" {
		q.AudioTranscript = source
	}
	return q
}

func loadBenchmarkQuestions(paths ...string) (map[string]benchmarkQuestion, error) {
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if errors.Is(err, os.ErrNotExist) || len(data) == 0 {
		return map[string]benchmarkQuestion{}, nil
	}
	if err != nil {
		return nil, err
	}
	var list []benchmarkQuestion
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("load benchmark questions: %w", err)
	}
	byID := map[string]benchmarkQuestion{}
	for _, q := range list {
		key := firstNonEmpty(q.ID, q.RefID)
		byID[key] = q
	}
	return byID, nil
}

// parseReferenceScore parses a 10-90 score or a fractional "a/b" score scaled to 10-90.
func parseReferenceScore(v any) (float64, bool) {
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

// traitKey normalizes trait names across scoring rubrics.
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

// reportPaired prints mean reference scores, engine scores, bias, and mean absolute error per task and trait.
func reportPaired(results []pairedResult) {
	type agg struct{ n, ref, ours, abs float64 }
	byTask := map[string]map[string]*agg{}
	for _, r := range results {
		if r.Error != "" {
			continue
		}
		key := r.Marker + " | " + r.Task
		traits := byTask[key]
		if traits == nil {
			traits = map[string]*agg{}
			byTask[key] = traits
		}
		for name, a := range r.Ref {
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
			t.ref += a
			t.ours += o
			t.abs += math.Abs(o - a)
		}
	}
	fmt.Println("\nmarker | task                         trait                                  n  theirs  ours   bias  mean|gap|")
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
			fmt.Printf("%-36s %-36s %4.0f  %5.1f  %5.1f  %+5.1f  %7.1f\n", task, name, a.n, a.ref/a.n, a.ours/a.n, (a.ours-a.ref)/a.n, a.abs/a.n)
		}
	}
	unmatched := map[string]bool{}
	for _, r := range results {
		for name := range r.Ref {
			if _, ok := r.Ours[name]; !ok && r.Error == "" {
				unmatched[r.Task+": "+name] = true
			}
		}
	}
	for k := range unmatched {
		fmt.Printf("  no matching trait for %s\n", k)
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
