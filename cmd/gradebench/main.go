// Command gradebench grades fixed sample answers with the live AI gateway and
// reports scores against each sample's target level, plus token use. It never
// opens a database. Configure the provider with AI_BASE_URL, AI_API_KEY and
// AI_MODEL_WRITING.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/pkg/config"
)

type sample struct {
	ID              string  `json:"id"`
	Kind            string  `json:"kind"`
	Exam            string  `json:"exam"`
	TypeID          string  `json:"typeId"`
	TaskName        string  `json:"taskName"`
	Prompt          string  `json:"prompt"`
	Context         string  `json:"context"`
	Figure          string  `json:"figure"`
	Source          string  `json:"source"`
	MinimumWords    int     `json:"minimumWords"`
	DurationSeconds int     `json:"durationSeconds"`
	Target          float64 `json:"target"`
	Text            string  `json:"text"`
	Reference       string  `json:"reference"`
	Delivery        string  `json:"delivery"`
	Turns           []struct {
		Part     int    `json:"part"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
		Seconds  int    `json:"seconds"`
	} `json:"turns"`
}

type result struct {
	ID               string             `json:"id"`
	Exam             string             `json:"exam"`
	Kind             string             `json:"kind"`
	Run              int                `json:"run"`
	Target           float64            `json:"target"`
	Score            *float64           `json:"score"`
	Criteria         map[string]float64 `json:"criteria"`
	Fixes            int                `json:"fixes"`
	FeedbackChars    int                `json:"feedbackChars"`
	Model            string             `json:"model"`
	PromptTokens     int                `json:"promptTokens"`
	CompletionTokens int                `json:"completionTokens"`
	LatencyMS        int                `json:"latencyMs"`
	Error            string             `json:"error,omitempty"`
}

// validationFailures counts replies the gateway rejected and asked for again.
var validationFailures atomic.Int64

type countingHandler struct{ slog.Handler }

func (h countingHandler) Handle(ctx context.Context, r slog.Record) error {
	if strings.Contains(r.Message, "failed validation") {
		validationFailures.Add(1)
	}
	return h.Handler.Handle(ctx, r)
}

func (h countingHandler) WithAttrs(a []slog.Attr) slog.Handler {
	return countingHandler{h.Handler.WithAttrs(a)}
}

func main() {
	dir := flag.String("dir", "cmd/gradebench", "folder holding samples_*.json")
	runs := flag.Int("runs", 1, "times to grade each sample")
	workers := flag.Int("workers", 4, "parallel requests")
	only := flag.String("only", "", "grade only sample ids containing this text")
	out := flag.String("out", "", "write results JSON here")
	audio := flag.String("audio", "", "comma-separated recordings to transcribe with Whisper instead of grading")
	flag.Parse()

	key := os.Getenv("AI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "AI_API_KEY is not set")
		os.Exit(2)
	}
	model := firstNonEmpty(os.Getenv("AI_MODEL_WRITING"), config.DefaultAIModel)
	cfg := &config.Config{
		AIBaseURL:        strings.TrimRight(firstNonEmpty(os.Getenv("AI_BASE_URL"), "https://openrouter.ai/api/v1"), "/"),
		AIAPIKey:         key,
		AIAudioBaseURL:   strings.TrimRight(firstNonEmpty(os.Getenv("AI_BASE_URL"), "https://openrouter.ai/api/v1"), "/"),
		AIAudioAPIKey:    key,
		AIModels:         config.AIModels{Writing: model, Speaking: model, Tutoring: model},
		AIRequestTimeout: 120 * time.Second,
		AIMaxTokens:      8000,
	}
	log := slog.New(countingHandler{slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})})
	g := ai.NewGateway(cfg, log)
	if *audio != "" {
		transcribe(g, strings.Split(*audio, ","))
		return
	}

	samples, err := loadSamples(*dir, *only)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("grading %d samples x %d run(s) with %s on %s\n\n", len(samples), *runs, model, hostOf(cfg.AIBaseURL))

	type job struct {
		s   sample
		run int
	}
	jobs := make(chan job)
	var mu sync.Mutex
	var results []result
	var wg sync.WaitGroup
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				r := grade(g, j.s)
				r.Run = j.run
				mu.Lock()
				results = append(results, r)
				fmt.Printf("  %-28s run %d  target %5.1f  score %5s  tokens %5d/%5d  %5.1fs %s\n",
					r.ID, r.Run, r.Target, fmtScore(r.Score), r.PromptTokens, r.CompletionTokens,
					float64(r.LatencyMS)/1000, r.Error)
				mu.Unlock()
			}
		}()
	}
	for run := 1; run <= *runs; run++ {
		for _, s := range samples {
			jobs <- job{s, run}
		}
	}
	close(jobs)
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].ID != results[j].ID {
			return results[i].ID < results[j].ID
		}
		return results[i].Run < results[j].Run
	})
	report(results, *runs)

	if *out != "" {
		data, _ := json.MarshalIndent(map[string]any{
			"model": model, "provider": hostOf(cfg.AIBaseURL), "ranAt": time.Now().Format(time.RFC3339),
			"validationRetries": validationFailures.Load(), "results": results,
		}, "", "  ")
		if err := os.WriteFile(*out, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("\nresults written to %s\n", *out)
	}
}

func grade(g *ai.Gateway, s sample) result {
	exam := models.ExamType(s.Exam)
	minScore, maxScore := 0.0, 9.0
	if exam == models.ExamPTE {
		minScore, maxScore = 10, 90
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	var ev models.Evaluation
	var usage ai.Usage
	var err error
	switch s.Kind {
	case "speaking-test":
		turns := make([]ai.SpeakingTestTurn, 0, len(s.Turns))
		for _, t := range s.Turns {
			turns = append(turns, ai.SpeakingTestTurn{Part: t.Part, Question: t.Question, Answer: t.Answer, Seconds: t.Seconds})
		}
		ev, usage, err = g.EvaluateSpeakingTest(ctx, ai.SpeakingTestRequest{Turns: turns, MinScore: minScore, MaxScore: maxScore})
	case "writing":
		ev, usage, err = g.EvaluateWriting(ctx, ai.WritingRequest{
			Exam: exam, TypeID: s.TypeID, TaskName: s.TaskName, Prompt: s.Prompt,
			ContextPassage: s.Context, FigureData: s.Figure, LearnerText: s.Text,
			WordCount: len(strings.Fields(s.Text)), MinimumWords: s.MinimumWords,
			MinScore: minScore, MaxScore: maxScore,
		})
	default:
		ev, usage, err = g.EvaluateSpokenTranscript(ctx, ai.SpokenTranscriptRequest{
			Exam: exam, TaskName: s.TaskName, Prompt: s.Prompt, SourceText: s.Source, ReferenceAnswer: s.Reference,
			Transcript: s.Text, DurationSeconds: s.DurationSeconds, WordCount: len(strings.Fields(s.Text)), Delivery: s.Delivery,
			MinScore: minScore, MaxScore: maxScore,
		})
	}

	r := result{ID: s.ID, Exam: s.Exam, Kind: s.Kind, Target: s.Target, Model: usage.Model,
		PromptTokens: usage.PromptTokens, CompletionTokens: usage.CompletionTokens, LatencyMS: usage.LatencyMS}
	if err != nil {
		r.Error = err.Error()
		return r
	}
	r.Score = ev.EstimatedScore
	r.Fixes = len(ev.SentenceFeedback)
	r.Criteria = map[string]float64{}
	r.FeedbackChars = len(ev.Summary)
	for _, c := range ev.Criteria {
		r.Criteria[c.Name] = c.Score
		r.FeedbackChars += len(c.Feedback)
	}
	for _, f := range ev.SentenceFeedback {
		r.FeedbackChars += len(f.Correction) + len(f.Explanation)
	}
	return r
}

// report prints accuracy against targets and average token use per exam.
func report(results []result, runs int) {
	type agg struct {
		n, ok, within   int
		absErr, in, out float64
		latency         float64
		fixes           float64
	}
	byExam := map[string]*agg{}
	spread := map[string][]float64{}
	for _, r := range results {
		a := byExam[r.Exam]
		if a == nil {
			a = &agg{}
			byExam[r.Exam] = a
		}
		a.n++
		a.in += float64(r.PromptTokens)
		a.out += float64(r.CompletionTokens)
		a.latency += float64(r.LatencyMS)
		if r.Score == nil {
			continue
		}
		a.ok++
		a.fixes += float64(r.Fixes)
		diff := math.Abs(*r.Score - r.Target)
		a.absErr += diff
		tolerance := 0.5
		if r.Exam == "PTE" {
			tolerance = 7
		}
		if diff <= tolerance {
			a.within++
		}
		spread[r.ID] = append(spread[r.ID], *r.Score)
	}

	fmt.Println("\nexam   graded  mean |score-target|  within tolerance  avg fixes  avg tokens in/out  avg latency")
	for _, exam := range []string{"IELTS", "PTE"} {
		a := byExam[exam]
		if a == nil || a.n == 0 {
			continue
		}
		ok := math.Max(float64(a.ok), 1)
		tol := "±0.5 band"
		if exam == "PTE" {
			tol = "±7 points"
		}
		fmt.Printf("%-6s %3d/%-3d %10.2f %12d/%d (%s) %8.1f %10.0f/%-6.0f %8.1fs\n",
			exam, a.ok, a.n, a.absErr/ok, a.within, a.ok, tol, a.fixes/ok, a.in/float64(a.n), a.out/float64(a.n), a.latency/float64(a.n)/1000)
	}

	if runs > 1 {
		var worst float64
		var worstID string
		for id, scores := range spread {
			lo, hi := scores[0], scores[0]
			for _, s := range scores {
				lo, hi = math.Min(lo, s), math.Max(hi, s)
			}
			if hi-lo > worst {
				worst, worstID = hi-lo, id
			}
		}
		fmt.Printf("\nlargest score change between runs: %.1f (%s)\n", worst, worstID)
	}
	fmt.Printf("replies rejected by validation and re-asked: %d\n", validationFailures.Load())
}

// transcribe sends each recording to Whisper and prints what it heard.
func transcribe(g *ai.Gateway, files []string) {
	for _, f := range files {
		data, err := os.ReadFile(strings.TrimSpace(f))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		started := time.Now()
		text, _, err := g.Transcribe(ctx, data, strings.TrimPrefix(filepath.Ext(f), "."))
		cancel()
		fmt.Printf("%s (%d KB, %.1fs)\n", filepath.Base(f), len(data)/1024, time.Since(started).Seconds())
		if err != nil {
			fmt.Printf("  error: %v\n\n", err)
			continue
		}
		fmt.Printf("  %s\n\n", text)
	}
}

func loadSamples(dir, only string) ([]sample, error) {
	files, err := filepath.Glob(filepath.Join(dir, "samples_*.json"))
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("no samples_*.json in %s", dir)
	}
	var all []sample
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var list []sample
		if err := json.Unmarshal(data, &list); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		for _, s := range list {
			if only == "" || strings.Contains(s.ID, only) {
				all = append(all, s)
			}
		}
	}
	return all, nil
}

func fmtScore(s *float64) string {
	if s == nil {
		return "-"
	}
	return fmt.Sprintf("%.1f", *s)
}

func hostOf(u string) string {
	u = strings.TrimPrefix(strings.TrimPrefix(u, "https://"), "http://")
	if i := strings.Index(u, "/"); i >= 0 {
		return u[:i]
	}
	return u
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
