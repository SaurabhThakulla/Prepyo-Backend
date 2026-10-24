package institution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PlanForm is the plan and end date a Prepyo admin gives an institution.
type PlanForm struct {
	Plan       string `json:"plan"`
	ValidUntil string `json:"validUntil"` // YYYY-MM-DD, the last day is the day before
}

// Check returns the parsed end date, or a message per rejected field keyed by its JSON name.
func (f PlanForm) Check(now time.Time) (time.Time, map[string]string) {
	problems := map[string]string{}
	if !ValidPlan(strings.ToLower(f.Plan)) {
		problems["plan"] = "Choose Basic, Extended or Ultimate."
	}
	until, err := time.Parse(time.DateOnly, f.ValidUntil)
	if err != nil {
		problems["validUntil"] = "Choose the date the plan ends."
	} else if !until.After(now) {
		problems["validUntil"] = "The end date must be in the future."
	}
	return until, problems
}

// SetPlan gives an institution its plan and end date; Prepyo admins use it.
func SetPlan(ctx context.Context, pool *pgxpool.Pool, institutionID, plan string, validUntil time.Time) error {
	tag, err := pool.Exec(ctx, `UPDATE institutions SET plan = $2, valid_until = $3 WHERE id::text = $1`,
		institutionID, strings.ToLower(plan), validUntil)
	if err != nil {
		return fmt.Errorf("set institution plan: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
