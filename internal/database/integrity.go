package database

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// CheckSeedIntegrity verifies canonical IDs, not just totals (unrelated authored
// questions must not mask missing seed content). It never silently re-seeds data.
func CheckSeedIntegrity(ctx context.Context, db DB) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	rows, err := db.Query(ctx, `WITH expected(id, skill) AS (
 SELECT 'ielts-wrt-op-' || lpad(n::text, 3, '0'), 'writing' FROM generate_series(1,100) n
 UNION ALL SELECT 'pte-wrt-es-' || lpad(n::text, 3, '0'), 'writing' FROM generate_series(1,100) n
 UNION ALL SELECT 'ielts-wrt-fg-' || lpad(n::text, 3, '0'), 'writing' FROM generate_series(1,10) n
 UNION ALL SELECT 'pte-wrt-swt-' || lpad(n::text, 3, '0'), 'writing' FROM generate_series(2,5) n
 UNION ALL SELECT * FROM (VALUES ('pte-wrt-001','writing'), ('ielts-wrt-001','writing'),
 ('pte-spk-001','speaking'), ('ielts-spk-001','speaking')) v
 ) SELECT e.id FROM expected e LEFT JOIN questions q ON q.id=e.id
 WHERE q.id IS NULL OR q.skill<>e.skill OR NOT (q.exam=ANY(q.supported_exams))
 ORDER BY e.id`)
	if err != nil {
		return fmt.Errorf("check canonical content: %w", err)
	}
	defer rows.Close()
	var missing []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan canonical content: %w", err)
		}
		missing = append(missing, id)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read canonical content: %w", err)
	}
	if len(missing) > 0 {
		return fmt.Errorf("canonical writing/speaking content missing or invalid (%d): %s; restore via a reviewed repair migration", len(missing), strings.Join(missing, ", "))
	}
	return nil
}
