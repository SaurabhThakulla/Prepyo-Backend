package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// planForRole is the reverse of models.RoleForPlan. Setting a tier role from
// here grants the matching plan, because a role on its own would not survive:
// users.ReconcileRoles recomputes roles from plan state every hour and would
// undo a role its plan does not back.
var planForRole = map[string]string{
	models.RoleSuru:    "free",
	models.RoleAbhyas:  "weekly",
	models.RoleTaiyari: "pro",
	models.RoleUdaan:   "elite",
}

// roleOrder is the order tiers are reported in, cheapest first, with admin
// leading. Used for the breakdown so the dashboard reads the same every time.
var roleOrder = []string{
	models.RoleAdmin,
	models.RoleSuru,
	models.RoleAbhyas,
	models.RoleTaiyari,
	models.RoleUdaan,
}

type adminUser struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Role           string `json:"role"`
	PlanID         string `json:"planId"`
	PlanStartedAt  string `json:"planStartedAt"`
	PlanValidUntil string `json:"planValidUntil,omitempty"`
	TargetExam     string `json:"targetExam"`
	XP             int    `json:"xp"`
	StreakDays     int    `json:"streakDays"`
	CreatedAt      string `json:"createdAt"`
}

type roleCount struct {
	Role  string `json:"role"`
	Plan  string `json:"plan"`
	Count int    `json:"count"`
}

// users lists accounts for the admin table, newest first. ?q filters on name or
// email, ?role narrows to one tier.
func (h *Handler) users(w http.ResponseWriter, r *http.Request) {
	page := httpx.ReadPage(r)
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	role := strings.TrimSpace(r.URL.Query().Get("role"))

	if role != "" && !validRole(role) {
		httpx.ValidationError(w, map[string]string{"role": "Not a role this system uses."})
		return
	}

	list, total, err := h.listUsers(r.Context(), page, query, role)
	if err != nil {
		httpx.Internal(w, h.log, "admin.users", err)
		return
	}

	counts, err := h.countsByRole(r.Context())
	if err != nil {
		httpx.Internal(w, h.log, "admin.countsByRole", err)
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"users":        list,
		"countsByRole": counts,
		"meta":         page.Meta(total),
	})
}

// userFields is shared by the list and the update so a row means the same thing
// whichever way it was fetched.
const userFields = `id, COALESCE(email, ''), name, role, plan_id,
	plan_started_at, plan_valid_until, target_exam, xp, streak_days, created_at`

func (h *Handler) listUsers(ctx context.Context, page httpx.Page, query, role string) ([]adminUser, int, error) {
	// One filter expression for both the count and the page, so the total can
	// never describe a different set than the rows.
	const where = `
		WHERE ($1 = '' OR name ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR role = $2)`

	var total int
	if err := h.db.QueryRow(ctx, `SELECT count(*) FROM users`+where, query, role).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count admin users: %w", err)
	}

	rows, err := h.db.Query(ctx, `SELECT `+userFields+` FROM users`+where+`
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`, query, role, page.Limit, page.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list admin users: %w", err)
	}
	defer rows.Close()

	list := make([]adminUser, 0, page.Limit)
	for rows.Next() {
		u, err := scanAdminUser(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, u)
	}
	return list, total, rows.Err()
}

// countsByRole is the "how many on each plan" breakdown. Every role is listed
// even at zero, so a tier nobody is on is shown as empty rather than missing.
func (h *Handler) countsByRole(ctx context.Context) ([]roleCount, error) {
	rows, err := h.db.Query(ctx, `SELECT role, count(*) FROM users GROUP BY role`)
	if err != nil {
		return nil, fmt.Errorf("count users by role: %w", err)
	}
	defer rows.Close()

	found := map[string]int{}
	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, fmt.Errorf("scan role count: %w", err)
		}
		found[role] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	counts := make([]roleCount, 0, len(roleOrder))
	for _, role := range roleOrder {
		counts = append(counts, roleCount{Role: role, Plan: planForRole[role], Count: found[role]})
	}
	return counts, nil
}

type setRoleRequest struct {
	Role string `json:"role"`
}

// setUserRole changes one account's role.
//
// A tier role is granted by writing the plan behind it, not just the role
// column: the hourly reconcile derives roles from plan state, so a bare role
// write would be undone within the hour. Promoting to admin leaves the plan
// alone, because admin is an access level rather than a tier.
func (h *Handler) setUserRole(w http.ResponseWriter, r *http.Request) {
	actor := reqctx.MustUser(r.Context())
	targetID := chi.URLParam(r, "id")

	var req setRoleRequest
	if !httpx.Decode(w, r, &req, h.log, "admin.setUserRole") {
		return
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	if !validRole(role) {
		httpx.ValidationError(w, map[string]string{"role": "Not a role this system uses."})
		return
	}

	// Dropping your own admin rights locks you out of the page you are standing
	// on, and there may be no other admin left to undo it.
	if targetID == actor.ID && role != models.RoleAdmin {
		httpx.Error(w, http.StatusConflict, httpx.CodeConflict,
			"You cannot remove your own admin access. Ask another admin to do it.")
		return
	}

	updated, err := h.applyRole(r.Context(), targetID, role)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "That account no longer exists.")
		return
	case err != nil:
		httpx.Internal(w, h.log, "admin.setUserRole", err)
		return
	}

	// Role changes hand out paid access, so they are worth a line in the log
	// naming who did it.
	h.log.Info("admin changed a user role",
		"actor", actor.Email, "target", updated.Email, "role", updated.Role, "plan", updated.PlanID)

	httpx.JSON(w, http.StatusOK, map[string]any{"user": updated})
}

func (h *Handler) applyRole(ctx context.Context, userID, role string) (adminUser, error) {
	if role == models.RoleAdmin {
		return scanAdminUser(h.db.QueryRow(ctx, `
			UPDATE users SET role = 'admin', updated_at = now()
			WHERE id = $1
			RETURNING `+userFields, userID))
	}

	// The free tier has no expiry; a paid tier gets a fresh period of that
	// plan's own length, starting today.
	return scanAdminUser(h.db.QueryRow(ctx, `
		UPDATE users SET
		    role = $2,
		    plan_id = $3,
		    plan_started_at = CURRENT_DATE,
		    plan_valid_until = CASE
		        WHEN $3 = 'free' THEN NULL
		        ELSE CURRENT_DATE + GREATEST(
		            COALESCE((SELECT NULLIF(p.duration_days, 0) FROM plans p WHERE p.id = $3),
		                     (SELECT p.duration_months * 30 FROM plans p WHERE p.id = $3)),
		            1)
		    END,
		    updated_at = now()
		WHERE id = $1
		RETURNING `+userFields, userID, role, planForRole[role]))
}

// row is satisfied by both pgx.Row and pgx.Rows, so one scanner serves the
// list and the update.
type row interface {
	Scan(dest ...any) error
}

func scanAdminUser(r row) (adminUser, error) {
	var u adminUser
	var startedAt, createdAt time.Time
	var validUntil *time.Time

	if err := r.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.PlanID,
		&startedAt, &validUntil, &u.TargetExam, &u.XP, &u.StreakDays, &createdAt); err != nil {
		return adminUser{}, err
	}

	u.PlanStartedAt = startedAt.Format(time.DateOnly)
	u.CreatedAt = createdAt.Format(time.DateOnly)
	if validUntil != nil {
		u.PlanValidUntil = validUntil.Format(time.DateOnly)
	}
	return u, nil
}

func validRole(role string) bool {
	if role == models.RoleAdmin {
		return true
	}
	_, ok := planForRole[role]
	return ok
}
