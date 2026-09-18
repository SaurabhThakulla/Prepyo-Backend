package database

import (
	"context"
	"fmt"
	"time"
)

// CheckSeedIntegrity verifies foundational system configuration (exam_versions, plans).
// It never silently re-seeds data.
func CheckSeedIntegrity(ctx context.Context, db DB) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM exam_versions WHERE id IN ('pte-2026-01', 'ielts-2026-01')`).Scan(&count); err != nil {
		return fmt.Errorf("check exam_versions: %w", err)
	}
	if count < 2 {
		return fmt.Errorf("core exam_versions missing (%d/2 found)", count)
	}

	if err := db.QueryRow(ctx, `SELECT count(*) FROM plans WHERE id IN ('free', 'pro', 'elite')`).Scan(&count); err != nil {
		return fmt.Errorf("check plans: %w", err)
	}
	if count < 3 {
		return fmt.Errorf("core plans missing (%d/3 found)", count)
	}

	return nil
}
