package ai

import (
	"context"
	"testing"
	"time"
)

func TestTimeLeftForCall(t *testing.T) {
	tests := []struct {
		name string
		ctx  func() (context.Context, context.CancelFunc)
		want bool
	}{
		{
			name: "no deadline",
			ctx:  func() (context.Context, context.CancelFunc) { return context.Background(), func() {} },
			want: true,
		},
		{
			name: "room for another call",
			ctx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), minCallBudget+time.Minute)
			},
			want: true,
		},
		{
			// Starting a call here would burn the rest of the budget and fail
			// anyway, leaving the caller with a cancellation to report.
			name: "not enough left",
			ctx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), minCallBudget/2)
			},
			want: false,
		},
		{
			name: "already cancelled",
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()
			if got := timeLeftForCall(ctx); got != tc.want {
				t.Fatalf("timeLeftForCall = %v, want %v", got, tc.want)
			}
		})
	}
}
