// Command officialbench compares learners' official PTE / IELTS score reports
// with their Prepyo full-mock scores and prints the gap per skill.
//
// Each report either carries the Prepyo scores itself ("prepyo"), or the tool
// looks up the learner's latest full mock in DATABASE_URL,
// inside a read-only transaction. Output names learners by "id", never email.
//
//	go run ./cmd/officialbench reports.json
//	DATABASE_URL=postgres://... go run ./cmd/officialbench "reports/*.json"
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// The production Supabase project: read only when -prod-read is passed.
const prodProject = "tkncyktpjsjsrpfvoyiy"

type report struct {
	ID        string             `json:"id"`        // label shown in output
	Email     string             `json:"email"`     // to find the learner's mock; never printed
	Exam      string             `json:"exam"`      // PTE or IELTS
	TestDate  string             `json:"testDate"`  // YYYY-MM-DD of the official test, optional
	AttemptID string             `json:"attemptId"` // a specific mock_attempts row, instead of the latest
	Official  map[string]float64 `json:"official"`  // overall, speaking, writing, reading, listening
	Prepyo    map[string]float64 `json:"prepyo"`    // filled from the database when empty
}

var skills = []string{"overall", "listening", "reading", "writing", "speaking"}

func main() {
	prodRead := flag.Bool("prod-read", false, "allow DATABASE_URL to be the production database (read-only)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: officialbench [-prod-read] <reports.json or glob>...")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	reports, err := load(flag.Args())
	if err != nil {
		fail(err)
	}

	if needsDB(reports) {
		url := os.Getenv("DATABASE_URL")
		if url == "" {
			fail(errors.New("some reports have no \"prepyo\" scores: set DATABASE_URL explicitly to look them up"))
		}
		if strings.Contains(url, prodProject) && !*prodRead {
			fail(errors.New("DATABASE_URL is the production database: pass -prod-read to read it"))
		}
		if err := lookUp(url, reports); err != nil {
			fail(err)
		}
	}
	printReport(reports)
}

func load(patterns []string) ([]*report, error) {
	var all []*report
	for _, pattern := range patterns {
		files, _ := filepath.Glob(pattern)
		if len(files) == 0 {
			return nil, fmt.Errorf("no files match %s", pattern)
		}
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			var list []*report
			if err := json.Unmarshal(data, &list); err != nil {
				return nil, fmt.Errorf("%s: %w", f, err)
			}
			for i, r := range list {
				r.Exam = strings.ToUpper(strings.TrimSpace(r.Exam))
				if r.Exam != "PTE" && r.Exam != "IELTS" {
					return nil, fmt.Errorf("%s #%d: exam must be PTE or IELTS", f, i+1)
				}
				if r.ID == "" {
					r.ID = fmt.Sprintf("%s#%d", filepath.Base(f), i+1)
				}
				r.Official, r.Prepyo = lower(r.Official), lower(r.Prepyo)
			}
			all = append(all, list...)
		}
	}
	return all, nil
}

func lower(m map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range m {
		out[strings.ToLower(strings.TrimSpace(k))] = v
	}
	return out
}

func needsDB(reports []*report) bool {
	for _, r := range reports {
		if len(r.Prepyo) == 0 {
			return true
		}
	}
	return false
}

// lookUp fills each report's Prepyo scores from its learner's full mock.
func lookUp(url string, reports []*report) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, r := range reports {
		if len(r.Prepyo) > 0 {
			continue
		}
		var score float64
		var skillJSON []byte
		var at time.Time
		if r.AttemptID != "" {
			err = tx.QueryRow(ctx, `SELECT user_score, skill_scores, completed_at FROM mock_attempts WHERE id::text = $1`,
				r.AttemptID).Scan(&score, &skillJSON, &at)
		} else {
			mockID := map[string]string{"PTE": "mock-pte-full", "IELTS": "mock-ielts-full"}[r.Exam]
			err = tx.QueryRow(ctx, `
				SELECT a.user_score, a.skill_scores, a.completed_at
				FROM mock_attempts a JOIN users u ON u.id = a.user_id
				WHERE lower(u.email) = lower($1) AND a.mock_id = $2
				ORDER BY a.completed_at DESC
				LIMIT 1`, r.Email, mockID).Scan(&score, &skillJSON, &at)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Printf("  %-16s no full mock found\n", r.ID)
			continue
		}
		if err != nil {
			return fmt.Errorf("%s: %w", r.ID, err)
		}
		r.Prepyo = map[string]float64{}
		if err := json.Unmarshal(skillJSON, &r.Prepyo); err != nil {
			return fmt.Errorf("%s: skill scores: %w", r.ID, err)
		}
		r.Prepyo = lower(r.Prepyo)
		r.Prepyo["overall"] = score
		if test, err := time.Parse("2006-01-02", r.TestDate); err == nil {
			fmt.Printf("  %-16s latest mock %s, %+d days from the test\n", r.ID, at.Format("2006-01-02"), int(math.Round(at.Sub(test).Hours()/24)))
		} else {
			fmt.Printf("  %-16s latest mock %s\n", r.ID, at.Format("2006-01-02"))
		}
	}
	return nil
}

// printReport shows each learner, then per exam and skill: the official mean,
// Prepyo's, the bias (Prepyo minus official), mean gap and share within range.
func printReport(reports []*report) {
	tolerance := map[string]float64{"PTE": 5, "IELTS": 0.5}
	fmt.Printf("\n%-16s %-6s %-10s %8s %8s %7s\n", "learner", "exam", "skill", "official", "prepyo", "gap")
	type agg struct{ n, off, ours, abs, within float64 }
	byKey := map[string]*agg{}
	for _, r := range reports {
		for _, skill := range skills {
			off, ok1 := r.Official[skill]
			ours, ok2 := r.Prepyo[skill]
			if !ok1 || !ok2 {
				continue
			}
			fmt.Printf("%-16s %-6s %-10s %8.1f %8.1f %+7.1f\n", r.ID, r.Exam, skill, off, ours, ours-off)
			k := r.Exam + " " + skill
			a := byKey[k]
			if a == nil {
				a = &agg{}
				byKey[k] = a
			}
			a.n++
			a.off += off
			a.ours += ours
			a.abs += math.Abs(ours - off)
			if math.Abs(ours-off) <= tolerance[r.Exam]+1e-9 {
				a.within++
			}
		}
	}
	if len(byKey) == 0 {
		fmt.Println("\nno learner has both official and Prepyo scores")
		return
	}
	fmt.Printf("\n%-6s %-10s %4s %8s %8s %7s %9s %8s\n", "exam", "skill", "n", "official", "prepyo", "bias", "mean|gap|", "within")
	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	order := map[string]int{}
	for i, s := range skills {
		order[s] = i
	}
	sort.Slice(keys, func(i, j int) bool {
		ei, si, _ := strings.Cut(keys[i], " ")
		ej, sj, _ := strings.Cut(keys[j], " ")
		if ei != ej {
			return ei < ej
		}
		return order[si] < order[sj]
	})
	for _, k := range keys {
		a := byKey[k]
		exam, skill, _ := strings.Cut(k, " ")
		fmt.Printf("%-6s %-10s %4.0f %8.1f %8.1f %+7.1f %9.2f %4.0f/%-3.0f (±%g)\n", exam, skill, a.n, a.off/a.n, a.ours/a.n,
			(a.ours-a.off)/a.n, a.abs/a.n, a.within, a.n, tolerance[exam])
	}
	fmt.Println("\nbias > 0: Prepyo scores higher than the official test. Fewer than ~10 learners per exam is too few to tune on.")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
