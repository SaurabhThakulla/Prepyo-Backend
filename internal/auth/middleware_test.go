package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
)

func TestRequirePremium(t *testing.T) {
	svc := &Service{}
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name       string
		user       *models.User
		wantStatus int
	}{
		{
			name:       "unauthenticated",
			user:       nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "free user without active plan",
			user: &models.User{
				ID:     "user-1",
				PlanID: "free",
				Role:   models.RoleSuru,
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "paid plan but expired",
			user: &models.User{
				ID:             "user-2",
				PlanID:         "pro",
				Role:           models.RoleTaiyari,
				PlanValidUntil: &past,
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "active paid plan",
			user: &models.User{
				ID:             "user-3",
				PlanID:         "pro",
				Role:           models.RoleTaiyari,
				PlanValidUntil: &future,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "admin user bypasses plan requirement",
			user: &models.User{
				ID:   "admin-1",
				Role: models.RoleAdmin,
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := svc.RequirePremium(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/mistakes", nil)
			if tc.user != nil {
				req = req.WithContext(reqctx.WithUser(req.Context(), *tc.user))
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}
