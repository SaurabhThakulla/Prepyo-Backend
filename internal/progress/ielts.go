package progress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/scoring"
)

// recentEvaluations is how many recent writing or speaking estimates make up
// a skill band. One essay is a weak basis for a band; five is a steadier one
// without letting months-old work hold the estimate back.
const recentEvaluations = 5

// ieltsSkill is one skill's standing for an IELTS learner.
type ieltsSkill struct {
	band     *float64
	attempts int
	accuracy float64
}

// ieltsSkills estimates a band for each IELTS skill from the evidence that
// skill actually has:
//   - Reading and Listening: the latest mock of that skill in the window,
//     else practice accuracy. Accuracy becomes a band through
//     the indicative raw-to-band table for the learner's module (scaled to a
//     40-mark paper), not a straight percentage of nine.
//   - Writing: recent task estimates, with Task 2 counting twice as much as
//     Task 1, as it does in the IELTS Writing band.
//   - Speaking: recent answer estimates.
func (s *Service) ieltsSkills(ctx context.Context, db database.DB, user models.User) (map[models.SkillType]ieltsSkill, error) {
	since := time.Now().Add(-recentWindow)
	module := user.IELTSModule()
	skills := map[models.SkillType]ieltsSkill{}

	rows, err := db.Query(ctx, `
		SELECT q.skill, count(*), COALESCE(sum(a.score), 0), COALESCE(sum(a.max_score), 0)
		  FROM practice_attempts a
		  JOIN questions q ON q.id = a.question_id
		 WHERE a.user_id = $1 AND a.exam = 'IELTS' AND a.created_at > $2
		   AND q.skill IN ('reading', 'listening')
		 GROUP BY q.skill`, user.ID, since)
	if err != nil {
		return nil, fmt.Errorf("read IELTS practice totals: %w", err)
	}
	for rows.Next() {
		var skill models.SkillType
		var attempts int
		var earned, max float64
		if err := rows.Scan(&skill, &attempts, &earned, &max); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan IELTS practice totals: %w", err)
		}
		if max <= 0 {
			continue
		}
		accuracy := earned / max
		band := scoring.IELTSBandFromRawMarks(module, string(skill), int(math.Round(accuracy*40)), 40)
		skills[skill] = ieltsSkill{band: &band, attempts: attempts, accuracy: accuracy}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// A full paper is better evidence than scattered practice items.
	for _, skill := range []models.SkillType{models.SkillReading, models.SkillListening} {
		var mockScores []byte
		var mockCorrect, mockTotal int
		err = db.QueryRow(ctx, `
			SELECT skill_scores, total_correct, total_questions
			  FROM mock_attempts
			 WHERE user_id = $1 AND exam = 'IELTS' AND completed_at > $2
			   AND skill_scores ? $3
			 ORDER BY completed_at DESC
			 LIMIT 1`, user.ID, since, string(skill)).Scan(&mockScores, &mockCorrect, &mockTotal)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			continue
		case err != nil:
			return nil, fmt.Errorf("read latest IELTS %s mock: %w", skill, err)
		}
		var byskill map[models.SkillType]float64
		if err := json.Unmarshal(mockScores, &byskill); err != nil {
			return nil, fmt.Errorf("decode %s mock scores: %w", skill, err)
		}
		if band, ok := byskill[skill]; ok {
			current := skills[skill]
			current.band = &band
			// A full mock's attempt carries four bands and no marks.
			if mockTotal > 0 {
				current.accuracy = float64(mockCorrect) / float64(mockTotal)
			}
			current.attempts += mockTotal
			skills[skill] = current
		}
	}

	evaluations, err := db.Query(ctx, `
		SELECT recent.skill, recent.estimated_score::float8, COALESCE(q.type_id, '')
		  FROM (
			SELECT e.skill, e.estimated_score, e.question_id,
			       row_number() OVER (PARTITION BY e.skill ORDER BY e.created_at DESC) AS recency
			  FROM ai_evaluations e
			 WHERE e.user_id = $1 AND e.exam = 'IELTS' AND e.created_at > $2
			   AND e.estimated_score IS NOT NULL
			   AND e.skill IN ('writing', 'speaking')
		  ) recent
		  LEFT JOIN questions q ON q.id = recent.question_id
		 WHERE recency <= $3`, user.ID, since, recentEvaluations)
	if err != nil {
		return nil, fmt.Errorf("read IELTS evaluations: %w", err)
	}
	defer evaluations.Close()

	type weighted struct{ sum, weight float64 }
	totals := map[models.SkillType]*weighted{}
	counts := map[models.SkillType]int{}
	for evaluations.Next() {
		var skill models.SkillType
		var score float64
		var typeID string
		if err := evaluations.Scan(&skill, &score, &typeID); err != nil {
			return nil, fmt.Errorf("scan IELTS evaluation: %w", err)
		}
		weight := 1.0
		if skill == models.SkillWriting && strings.Contains(typeID, "task2") {
			weight = 2
		}
		if totals[skill] == nil {
			totals[skill] = &weighted{}
		}
		totals[skill].sum += score * weight
		totals[skill].weight += weight
		counts[skill]++
	}
	if err := evaluations.Err(); err != nil {
		return nil, err
	}
	for skill, total := range totals {
		band := scoring.RoundIELTSBand(total.sum / total.weight)
		skills[skill] = ieltsSkill{band: &band, attempts: counts[skill], accuracy: band / 9}
	}
	return skills, nil
}

// ieltsEstimate is the overall band: the mean of all four skill bands with the
// official rounding. With fewer than four there is no overall band, and the
// estimate says how many skills it is still waiting for.
func (s *Service) ieltsEstimate(ctx context.Context, db database.DB, user models.User) (models.ScoreEstimate, error) {
	skills, err := s.ieltsSkills(ctx, db, user)
	if err != nil {
		return models.ScoreEstimate{}, err
	}

	estimate := models.ScoreEstimate{
		TargetScore:    user.TargetScore,
		UpdatedAt:      time.Now(),
		SkillsRequired: len(models.AllSkills),
	}

	var sum float64
	fewest := -1
	for _, skill := range models.AllSkills {
		row, ok := skills[skill]
		if !ok || row.band == nil {
			continue
		}
		estimate.SkillsCovered++
		estimate.BasedOn += row.attempts
		sum += *row.band
		if fewest < 0 || row.attempts < fewest {
			fewest = row.attempts
		}
	}

	if estimate.SkillsCovered < len(models.AllSkills) {
		estimate.Confidence = "low"
		return estimate, nil
	}

	value := scoring.RoundIELTSBand(sum / float64(len(models.AllSkills)))
	estimate.Value = &value
	// The overall band is only as settled as its least-practised skill.
	estimate.Confidence = scoring.Confidence(fewest * 4)
	if estimate.TargetScore != nil {
		gap := math.Max(0, *estimate.TargetScore-value)
		estimate.TargetGap = &gap
		readiness := 100
		if *estimate.TargetScore > 0 {
			readiness = int(math.Round(math.Min(1, value / *estimate.TargetScore) * 100))
		}
		estimate.Readiness = &readiness
	}
	return estimate, nil
}
