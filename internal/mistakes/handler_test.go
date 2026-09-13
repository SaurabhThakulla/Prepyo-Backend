package mistakes

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
)

type dummyDB struct{}

func (d *dummyDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (d *dummyDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, errors.New("db query error")
}

func (d *dummyDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &dummyRow{}
}

type dummyRow struct{}

func (r *dummyRow) Scan(dest ...any) error {
	return errors.New("db row scan error")
}

func TestMistakesHandlerRequiresPremium(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := NewRepository(&dummyDB{})
	handler := NewHandler(nil, repo, nil, log)
	routes := handler.Routes()

	future := time.Now().Add(48 * time.Hour)

	tests := []struct {
		name       string
		user       models.User
		wantStatus int
	}{
		{
			name: "freemium user rejected with 403",
			user: models.User{
				ID:     "free-user",
				PlanID: "free",
				Role:   models.RoleSuru,
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "premium user accepted through middleware to handler",
			user: models.User{
				ID:             "pro-user",
				PlanID:         "pro",
				Role:           models.RoleTaiyari,
				PlanValidUntil: &future,
			},
			// Passed middleware, reached h.list where dummyDB returns error -> 500
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(reqctx.WithUser(req.Context(), tc.user))
			rec := httptest.NewRecorder()

			routes.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
