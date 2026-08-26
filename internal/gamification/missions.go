package gamification

import (
	"context"
	"fmt"
	"time"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/models"
)

// TodayMissions lists the active missions with this learner's progress for the
// current day in their timezone.
func (s *Service) TodayMissions(ctx context.Context, db database.DB, user models.User) ([]models.DailyMission, error) {
	day := localDay(user)

	rows, err := db.Query(ctx, `
		SELECT m.id, m.title, m.description, m.skill, m.task_type, m.target_count,
		       m.xp_reward,
		       COALESCE(p.completed_count, 0),
		       p.completed_at IS NOT NULL
		FROM daily_missions m
		LEFT JOIN user_mission_progress p
		       ON p.mission_id = m.id AND p.user_id = $1 AND p.mission_date = $2
		WHERE m.is_active
		  AND (m.exam IS NULL OR m.exam = $3)
		ORDER BY m.id`,
		user.ID, day, user.TargetExam)
	if err != nil {
		return nil, fmt.Errorf("read missions: %w", err)
	}
	defer rows.Close()

	missions := []models.DailyMission{}
	for rows.Next() {
		var m models.DailyMission
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.Skill, &m.TaskType,
			&m.TargetCount, &m.XPReward, &m.CompletedCount, &m.Completed); err != nil {
			return nil, fmt.Errorf("scan mission: %w", err)
		}
		missions = append(missions, m)
	}
	return missions, rows.Err()
}

// RecordActivity credits missions for work the learner just finished and pays
// out any that reached their target.
//
// Progress is driven by real activity rather than a "mark complete" button, so
// a mission cannot be claimed without doing the work.
func (s *Service) RecordActivity(ctx context.Context, db database.DB, user models.User, skill models.SkillType) ([]models.DailyMission, error) {
	day := localDay(user)

	// One statement bumps every active mission for this skill and reports the
	// ones that have just reached their target.
	rows, err := db.Query(ctx, `
		INSERT INTO user_mission_progress (user_id, mission_id, mission_date, completed_count)
		SELECT $1, m.id, $2, 1
		FROM daily_missions m
		WHERE m.is_active AND m.skill = $3 AND (m.exam IS NULL OR m.exam = $4)
		ON CONFLICT (user_id, mission_id, mission_date) DO UPDATE
			SET completed_count = user_mission_progress.completed_count + 1,
			    completed_at = CASE
			        WHEN user_mission_progress.completed_at IS NOT NULL
			            THEN user_mission_progress.completed_at
			        WHEN user_mission_progress.completed_count + 1 >=
			             (SELECT target_count FROM daily_missions WHERE id = user_mission_progress.mission_id)
			            THEN now()
			        ELSE NULL
			    END
		RETURNING mission_id, completed_count, completed_at IS NOT NULL`,
		user.ID, day, skill, user.TargetExam)
	if err != nil {
		return nil, fmt.Errorf("record mission activity: %w", err)
	}

	type progress struct {
		missionID string
		completed bool
	}
	var updated []progress
	for rows.Next() {
		var p progress
		var count int
		if err := rows.Scan(&p.missionID, &count, &p.completed); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan mission progress: %w", err)
		}
		updated = append(updated, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Pay out completed missions. The source key includes the date so a
	// mission pays once per day and no more.
	for _, p := range updated {
		if !p.completed {
			continue
		}
		var title string
		var reward int
		if err := db.QueryRow(ctx, `SELECT title, xp_reward FROM daily_missions WHERE id = $1`, p.missionID).
			Scan(&title, &reward); err != nil {
			return nil, fmt.Errorf("read mission reward: %w", err)
		}
		if _, err := s.Award(ctx, db, AwardParams{
			UserID:    user.ID,
			Amount:    reward,
			Reason:    "Daily mission: " + title,
			SourceKey: fmt.Sprintf("mission:%s:%s", p.missionID, day.Format(time.DateOnly)),
		}); err != nil {
			return nil, err
		}
	}

	return s.TodayMissions(ctx, db, user)
}

// LocalDay is today's date in the learner's own timezone.
//
// Callers use it to build XP source keys that reset daily, so repeating the
// same task cannot be farmed for unlimited XP.
func LocalDay(user models.User) string {
	return localDay(user).Format(time.DateOnly)
}

func localDay(user models.User) time.Time {
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		location = nepalTime()
	}
	return truncateToDay(time.Now().In(location))
}
