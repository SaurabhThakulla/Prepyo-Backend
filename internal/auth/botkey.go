package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/prepyo/backend/internal/models"
	"github.com/prepyo/backend/internal/reqctx"
	"github.com/prepyo/backend/pkg/httpx"
)

// BotActorLoader returns the admin account the bot's actions are recorded
// under. It is looked up on every request, so removing that account's admin
// role turns the bot off without a restart.
type BotActorLoader func(ctx context.Context) (models.User, error)

// BotActor loads the account with this email as the bot's actor.
func (s *Service) BotActor(email string) BotActorLoader {
	return func(ctx context.Context) (models.User, error) {
		return s.users.ByEmail(ctx, email)
	}
}

// RequireBotKey guards the routes the WhatsApp admin bot calls. There is no
// browser session here, only a shared secret, so what is mounted behind it must
// stay narrow: see admin.Handler.BotRoutes.
//
// A request that is not allowed to know the routes exist (no key configured, or
// an address outside allowedCIDRs) gets the same 404 as any unknown path. Past
// that, a wrong key is a plain 401. On success the actor is placed on the
// request context, so the admin handlers behind it run unchanged and record who
// acted exactly as they do for the dashboard.
func RequireBotKey(key string, allowedCIDRs []string, actor BotActorLoader, log *slog.Logger) func(http.Handler) http.Handler {
	wantKey := sha256.Sum256([]byte(key))

	var allowed []*net.IPNet
	for _, prefix := range allowedCIDRs {
		if _, network, err := net.ParseCIDR(prefix); err == nil {
			allowed = append(allowed, network)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" || !botAddressAllowed(r, allowed) {
				httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "Not found.")
				return
			}

			// Hashing both sides first gives them the same length, so the
			// comparison leaks neither the key nor how long it is.
			gotKey := sha256.Sum256([]byte(bearerToken(r)))
			if subtle.ConstantTimeCompare(gotKey[:], wantKey[:]) != 1 {
				httpx.Error(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "Invalid bot key.")
				return
			}

			user, err := actor(r.Context())
			if err != nil || !user.IsAdmin() {
				// A valid key with nobody to act as: a deployment mistake, not
				// something the caller can fix by retrying.
				log.Error("bot actor is not an admin account", "error", err)
				httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeInternal, "Bot access is not configured.")
				return
			}

			next.ServeHTTP(w, r.WithContext(reqctx.WithUser(r.Context(), user)))
		})
	}
}

func bearerToken(r *http.Request) string {
	scheme, token, ok := strings.Cut(strings.TrimSpace(r.Header.Get("Authorization")), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

// botAddressAllowed reports whether the caller may see the bot routes. With no
// allowlist every address may; the key is then the only gate.
func botAddressAllowed(r *http.Request, allowed []*net.IPNet) bool {
	if len(allowed) == 0 {
		return true
	}
	ip := net.ParseIP(middleware.GetClientIP(r.Context()))
	if ip == nil {
		return false
	}
	for _, network := range allowed {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
