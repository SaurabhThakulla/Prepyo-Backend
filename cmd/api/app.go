package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/prepyo/backend/internal/admin"
	"github.com/prepyo/backend/internal/ai"
	"github.com/prepyo/backend/internal/auth"
	"github.com/prepyo/backend/internal/billing"
	"github.com/prepyo/backend/internal/evaluations"
	"github.com/prepyo/backend/internal/exams"
	"github.com/prepyo/backend/internal/gamification"
	"github.com/prepyo/backend/internal/leaderboards"
	"github.com/prepyo/backend/internal/mistakes"
	"github.com/prepyo/backend/internal/mocks"
	"github.com/prepyo/backend/internal/notifications"
	"github.com/prepyo/backend/internal/practice"
	"github.com/prepyo/backend/internal/progress"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/internal/referrals"
	"github.com/prepyo/backend/internal/users"
	"github.com/prepyo/backend/pkg/config"
	"github.com/prepyo/backend/pkg/httpx"
)

// app holds everything the router needs. Building it in one place makes the
// dependency direction obvious: repositories, then services, then handlers.
type app struct {
	cfg  *config.Config
	pool *pgxpool.Pool
	log  *slog.Logger

	authService *auth.Service

	authHandler         *auth.Handler
	userHandler         *users.Handler
	examHandler         *exams.Handler
	questionHandler     *questions.Handler
	practiceHandler     *practice.Handler
	mockHandler         *mocks.Handler
	mistakeHandler      *mistakes.Handler
	evaluationHandler   *evaluations.Handler
	aiHandler           *ai.Handler
	progressHandler     *progress.Handler
	gamificationHandler *gamification.Handler
	leaderboardHandler  *leaderboards.Handler
	notificationHandler *notifications.Handler
	billingHandler      *billing.Handler
	referralHandler     *referrals.Handler
	adminHandler        *admin.Handler
}

func newApp(cfg *config.Config, pool *pgxpool.Pool, log *slog.Logger) *app {
	// Repositories.
	userRepo := users.NewRepository(pool)
	examRepo := exams.NewRepository(pool)
	questionRepo := questions.NewRepository(pool)
	practiceRepo := practice.NewRepository(pool)
	mockRepo := mocks.NewRepository(pool)
	mistakeRepo := mistakes.NewRepository(pool)
	evaluationRepo := evaluations.NewRepository(pool)
	notificationRepo := notifications.NewRepository(pool)
	leaderboardRepo := leaderboards.NewRepository(pool)
	planRepo := billing.NewRepository(pool)
	referralRepo := referrals.NewRepository(pool)

	// Services.
	xpService := gamification.NewService()
	referralService := referrals.NewService(pool, referralRepo, xpService, notificationRepo, cfg.WebAppURL, log)
	authService := auth.NewService(pool, userRepo, referralService, cfg.SessionTTL, log)
	billingService := billing.NewService(planRepo, notificationRepo)
	progressService := progress.NewService(examRepo)
	gateway := ai.NewGateway(cfg, log)
	evaluationService := evaluations.NewService(pool, evaluationRepo, questionRepo, examRepo, billingService, gateway, xpService)

	return &app{
		cfg:         cfg,
		pool:        pool,
		log:         log,
		authService: authService,

		authHandler:         auth.NewHandler(authService, cfg.SecureCookies, cfg.SessionTTL),
		userHandler:         users.NewHandler(pool, userRepo, progressService, billingService, log),
		examHandler:         exams.NewHandler(examRepo, log),
		questionHandler:     questions.NewHandler(questionRepo, log),
		practiceHandler:     practice.NewHandler(pool, practiceRepo, questionRepo, mistakeRepo, xpService, referralService, log),
		mockHandler:         mocks.NewHandler(pool, mockRepo, questionRepo, examRepo, xpService, billingService, referralService, log),
		mistakeHandler:      mistakes.NewHandler(pool, mistakeRepo, xpService, log),
		evaluationHandler:   evaluations.NewHandler(evaluationService, evaluationRepo, log),
		aiHandler:           ai.NewHandler(gateway, log),
		progressHandler:     progress.NewHandler(pool, progressService, log),
		gamificationHandler: gamification.NewHandler(pool, xpService, log),
		leaderboardHandler:  leaderboards.NewHandler(leaderboardRepo, log),
		notificationHandler: notifications.NewHandler(notificationRepo, log),
		billingHandler:      billing.NewHandler(pool, planRepo, billingService, log),
		referralHandler:     referrals.NewHandler(referralService, log),
		adminHandler:        admin.NewHandler(pool, log),
	}
}

func (a *app) router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(a.requestLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: a.cfg.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
		// The session lives in a cookie, so the browser must be allowed to
		// send it. This is why AllowedOrigins can never be "*".
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", a.health)

	r.Route("/api/v1", func(v1 chi.Router) {
		// Public.
		v1.Group(func(public chi.Router) {
			// Credential endpoints are the ones worth guessing at, so they get
			// a tighter limit than everything else.
			public.With(rateLimit(10, time.Minute)).
				Mount("/auth", a.authHandler.Routes())
			public.Mount("/exams", a.examHandler.Routes())
			public.Mount("/subscriptions", a.billingHandler.Routes(a.authService.RequireUser))
			public.Mount("/referrals", a.referralHandler.Routes(a.authService.RequireUser))
		})

		// Signed in.
		v1.Group(func(private chi.Router) {
			private.Use(a.authService.RequireUser)
			private.Use(rateLimit(120, time.Minute))

			private.Mount("/profile", a.userHandler.Routes())
			private.Mount("/questions", a.questionHandler.Routes())
			private.Mount("/practice", a.practiceHandler.Routes())
			private.Mount("/mocks", a.mockHandler.Routes())
			private.Mount("/mistakes", a.mistakeHandler.Routes())
			private.Mount("/progress", a.progressHandler.Routes())
			private.Mount("/gamification", a.gamificationHandler.Routes())
			private.Mount("/leaderboards", a.leaderboardHandler.Routes())
			private.Mount("/notifications", a.notificationHandler.Routes())

			// AI endpoints cost money per call, so they are limited harder
			// than ordinary reads. Plan allowances are enforced separately in
			// the evaluations service.
			private.With(rateLimit(20, time.Minute)).
				Mount("/evaluations", a.evaluationHandler.Routes())
			private.With(rateLimit(20, time.Minute)).
				Mount("/ai", a.aiHandler.Routes())
		})

		// Admin only.
		v1.Group(func(adminOnly chi.Router) {
			adminOnly.Use(a.authService.RequireUser, a.authService.RequireAdmin)
			adminOnly.Mount("/admin", a.adminHandler.Routes())
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No such endpoint.")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeBadRequest, "That method is not allowed here.")
	})

	return r
}

// rateLimit throttles by client IP and answers in the same JSON envelope as
// every other error. httprate's own default writes plain text, which the
// browser cannot parse and which surfaces to a learner as "the server returned
// an unreadable response" instead of "you are going too fast".
//
// NOTE: httprate.KeyByIP reads r.RemoteAddr, which middleware.RealIP above has
// already rewritten from the client-supplied X-Forwarded-For header. Behind a
// trusted proxy that is correct; directly exposed it lets a caller pick their
// own bucket. See GHSA-9g5q-2w5x-hmxf. Fixing it properly needs to know which
// proxies to trust.
func rateLimit(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(requests, window, httprate.KeyByIP,
		httprate.WithLimitHandler(httpx.RateLimited))
}

// health reports whether the API can reach its database, so a load balancer
// does not keep sending traffic to an instance that cannot serve it.
func (a *app) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := a.pool.Ping(ctx); err != nil {
		a.log.Error("health check failed", "error", err)
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeInternal, "Database unreachable.")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"status": "healthy"})
}

// requestLogger records one line per request with the fields the spec asks for:
// request id, method, path, status and duration.
func (a *app) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(wrapped, r)

		level := slog.LevelInfo
		if wrapped.Status() >= 500 {
			level = slog.LevelError
		}
		// Only the path is logged, never the query string or body: those can
		// carry a learner's own writing.
		a.log.Log(r.Context(), level, "request",
			"requestId", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.Status(),
			"durationMs", time.Since(started).Milliseconds(),
		)
	})
}

// cleanExpiredSessions removes dead session rows in the background.
func (a *app) cleanExpiredSessions(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	a.authService.PurgeExpiredSessions(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.authService.PurgeExpiredSessions(ctx)
		}
	}
}
