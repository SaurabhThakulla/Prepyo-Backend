package main

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
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
	"github.com/prepyo/backend/internal/reading"
	"github.com/prepyo/backend/internal/referrals"
	"github.com/prepyo/backend/internal/report"
	"github.com/prepyo/backend/internal/users"
	"github.com/prepyo/backend/internal/web"
	"github.com/prepyo/backend/pkg/config"
	"github.com/prepyo/backend/pkg/httpx"
)

// requestTimeout is the deadline for a route that only talks to our own
// database. Anything slower than this is a bug, not a slow provider.
const requestTimeout = 60 * time.Second

// AIRouteTimeout is the deadline for the routes that wait on a model provider.
//
// One evaluation can spend AIRequestTimeout per provider call and a speaking
// one makes more than a single call: the audio attempt, then a retry, then the
// text fallback. Sized under the old blanket 60s, the request was cancelled
// while the provider was still answering, and the learner was told something
// had gone wrong on our side when nothing had.
func AIRouteTimeout(cfg *config.Config) time.Duration {
	return 3*cfg.AIRequestTimeout + 15*time.Second
}

type app struct {
	cfg  *config.Config
	pool *pgxpool.Pool
	log  *slog.Logger

	authService *auth.Service
	userRepo    *users.Repository

	authHandler         *auth.Handler
	userHandler         *users.Handler
	examHandler         *exams.Handler
	questionHandler     *questions.Handler
	readingHandler      *reading.Handler
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
	reportHandler       *report.Handler
	adminHandler        *admin.Handler
}

func newApp(cfg *config.Config, pool *pgxpool.Pool, log *slog.Logger) *app {
	// Repositories.
	userRepo := users.NewRepository(pool)
	examRepo := exams.NewRepository(pool)
	questionRepo := questions.NewRepository(pool)
	practiceRepo := practice.NewRepository(pool)
	mockRepo := mocks.NewRepository(pool)
	readingRepo := reading.NewRepository(pool)
	mistakeRepo := mistakes.NewRepository(pool)
	evaluationRepo := evaluations.NewRepository(pool)
	notificationRepo := notifications.NewRepository(pool)
	leaderboardRepo := leaderboards.NewRepository(pool)
	planRepo := billing.NewRepository(pool)
	referralRepo := referrals.NewRepository(pool)

	// Services.
	xpService := gamification.NewService()
	referralService := referrals.NewService(pool, referralRepo, xpService, notificationRepo, cfg.WebAppURL, log)
	authService := auth.NewService(pool, userRepo, referralService, cfg.SessionTTL, log,
		auth.NewGoogleVerifier(cfg.GoogleClientID), cfg.AdminEmail, cfg.AdminPassword)
	billingService := billing.NewService(planRepo, notificationRepo)
	progressService := progress.NewService(examRepo)
	gateway := ai.NewGateway(cfg, log)
	evaluationService := evaluations.NewService(pool, evaluationRepo, questionRepo, examRepo, billingService, gateway, xpService)
	readingService := reading.NewService(pool, readingRepo, questionRepo, mockRepo, examRepo, xpService, billingService, mistakeRepo)

	return &app{
		cfg:         cfg,
		pool:        pool,
		log:         log,
		authService: authService,

		userRepo:            userRepo,
		authHandler:         auth.NewHandler(authService, cfg.SecureCookies, cfg.SessionTTL),
		userHandler:         users.NewHandler(pool, userRepo, progressService, billingService, log),
		examHandler:         exams.NewHandler(examRepo, log),
		questionHandler:     questions.NewHandler(questionRepo, log),
		readingHandler:      reading.NewHandler(readingService, readingRepo, log),
		practiceHandler:     practice.NewHandler(pool, practiceRepo, questionRepo, mistakeRepo, examRepo, xpService, billingService, speakingScorable(gateway), log),
		mockHandler:         mocks.NewHandler(pool, mockRepo, questionRepo, examRepo, xpService, billingService, mistakeRepo, log),
		mistakeHandler:      mistakes.NewHandler(pool, mistakeRepo, xpService, log),
		evaluationHandler:   evaluations.NewHandler(evaluationService, evaluationRepo, log),
		aiHandler:           ai.NewHandler(gateway, pool, log),
		progressHandler:     progress.NewHandler(pool, progressService, log),
		gamificationHandler: gamification.NewHandler(pool, xpService, log),
		leaderboardHandler:  leaderboards.NewHandler(leaderboardRepo, log),
		notificationHandler: notifications.NewHandler(notificationRepo, log),
		billingHandler:      billing.NewHandler(pool, planRepo, billingService, log),
		referralHandler:     referrals.NewHandler(referralService, log),
		reportHandler:       report.NewHandler(pool, log),
		adminHandler:        admin.NewHandler(pool, log),
	}
}

// speakingScorable reports whether a speaking answer can be graded at all, by
// either route: the recording itself, or the transcript a device produced. Only
// when neither works is a speaking task free practice that spends no credit.
func speakingScorable(gateway *ai.Gateway) bool {
	return gateway.SpeakingAvailable() || gateway.Available()
}

func (a *app) router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	// HSTS only where cookies are already HTTPS-only, so a plain-HTTP dev
	// server never tells a browser to refuse it.
	r.Use(securityHeaders(a.cfg.SecureCookies))
	r.Use(trustedClientIP(a.cfg.TrustedProxyCIDRs))
	r.Use(a.requestLogger)
	r.Use(middleware.Recoverer)
	// The router's own deadline is the longest any route may take, which is an
	// AI-backed one. Everything else is held to requestTimeout below; giving the
	// whole router the shorter deadline used to cancel speaking evaluations
	// mid-flight, and a cancellation that landed after the provider replied cost
	// the learner feedback that had already been paid for.
	r.Use(middleware.Timeout(AIRouteTimeout(a.cfg)))

	r.Use(cors.Handler(apiCORSOptions(a.cfg.AllowedOrigins)))

	r.Get("/health", a.health)

	r.Route("/api/v1", func(v1 chi.Router) {
		// Rate limits count per device (see deviceKey), and a browser needs an
		// id before it has signed in.
		v1.Use(deviceCookie(a.cfg.SecureCookies))

		// Public.
		v1.Group(func(public chi.Router) {
			public.Use(middleware.Timeout(requestTimeout))

			// Nobody is signed in yet, so the device is the device cookie. That
			// cookie can be thrown away, so a per-IP ceiling sits underneath,
			// sized so a whole classroom signing in on one Wi-Fi never meets
			// it. Guessing is stopped elsewhere: admin login locks per email,
			// and Google verifies its own sign-ins.
			public.With(rateLimit(600, time.Minute), rateLimitByDevice(60, time.Minute)).
				Mount("/auth", a.authHandler.Routes())
			public.Mount("/exams", a.examHandler.Routes())

			// Plans are public; everything else here needs a session, and is
			// limited per device once the session is known.
			signedIn := func(next http.Handler) http.Handler {
				return a.authService.RequireUser(rateLimitByDevice(60, time.Minute)(next))
			}
			public.Mount("/subscriptions", a.billingHandler.Routes(signedIn))
			public.Mount("/referrals", a.referralHandler.Routes(signedIn))
		})

		// Signed in.
		v1.Group(func(private chi.Router) {
			// The per-IP ceiling is a backstop against one address running many
			// accounts; it is high enough that a shared network never meets it
			// through normal use. The real limit is per device.
			private.Use(rateLimit(600, time.Minute))
			private.Use(a.authService.RequireUser)
			private.Use(rateLimitByDevice(60, time.Minute))

			private.Group(func(quick chi.Router) {
				quick.Use(middleware.Timeout(requestTimeout))

				quick.Mount("/profile", a.userHandler.Routes())
				quick.Mount("/questions", a.questionHandler.Routes())
				quick.Mount("/reading", a.readingHandler.Routes())
				quick.Mount("/practice", a.practiceHandler.Routes())
				quick.Mount("/mocks", a.mockHandler.Routes())
				quick.With(a.authService.RequirePremium).
					Mount("/mistakes", a.mistakeHandler.Routes())
				quick.Mount("/progress", a.progressHandler.Routes())
				quick.Mount("/gamification", a.gamificationHandler.Routes())
				quick.Mount("/leaderboards", a.leaderboardHandler.Routes())
				quick.Mount("/notifications", a.notificationHandler.Routes())

				// Only raising a report or replying is throttled; reading your
				// own conversation is not.
				quick.Mount("/report", a.reportHandler.Routes(rateLimitByDevice(5, 10*time.Minute)))
			})

			// These wait on a provider call, so they keep the router's longer
			// deadline rather than the one every other route gets.
			private.With(rateLimitByDevice(20, time.Minute)).
				Mount("/evaluations", a.evaluationHandler.Routes())
			private.With(rateLimitByDevice(20, time.Minute)).
				Mount("/ai", a.aiHandler.Routes())
		})

		// Admin only.
		v1.Group(func(adminOnly chi.Router) {
			adminOnly.Use(middleware.Timeout(requestTimeout))
			adminOnly.Use(a.authService.RequireUser, a.authService.RequireAdmin)
			adminOnly.Mount("/admin", a.adminHandler.Routes())
		})
	})

	var serveWeb http.Handler
	if a.cfg.WebDistDir != "" {
		serveWeb = web.Handler(a.cfg.WebDistDir)
	}

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if serveWeb == nil || strings.HasPrefix(r.URL.Path, "/api/") {
			httpx.Error(w, http.StatusNotFound, httpx.CodeNotFound, "No such endpoint.")
			return
		}
		serveWeb.ServeHTTP(w, r)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, http.StatusMethodNotAllowed, httpx.CodeBadRequest, "That method is not allowed here.")
	})

	return r
}

func apiCORSOptions(origins []string) cors.Options {
	return cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// rateLimit throttles requests by client IP. Use it only as a generous
// ceiling; the real limits are per device (rateLimitByDevice).
func rateLimit(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(requests, window, clientIPKey,
		httprate.WithLimitHandler(httpx.RateLimited))
}

// health reports whether the API can reach its database.
func (a *app) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := a.pool.Ping(ctx); err != nil {
		a.log.Error("health check failed", "error", err)
		httpx.Error(w, http.StatusServiceUnavailable, httpx.CodeInternal, "Database unreachable.")
		return
	}
	httpx.JSON(w, http.StatusOK, healthPayload())
}

// requestLogger logs incoming HTTP requests with latency and status.
func (a *app) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(wrapped, r)

		level := slog.LevelInfo
		if wrapped.Status() >= 500 {
			level = slog.LevelError
		}
		a.log.Log(r.Context(), level, "request",
			"requestId", middleware.GetReqID(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.Status(),
			"durationMs", time.Since(started).Milliseconds(),
		)
	})
}

// reconcileRoles updates user roles for expired subscriptions on a periodic
// schedule, then sends the date-driven notifications that depend on them.
func (a *app) reconcileRoles(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	run := func() {
		// Queued plans start first: a plan that came due this tick should be
		// live before roles are derived from plan state, or the learner drops
		// to the free tier until the next pass.
		started, err := a.userRepo.ActivateQueuedPlans(ctx)
		if err != nil {
			a.log.Error("queued plan activation failed", "error", err)
		} else if started > 0 {
			a.log.Info("activated queued plans", "count", started)
		}

		changed, err := a.userRepo.ReconcileRoles(ctx)
		if err != nil {
			a.log.Error("role reconcile failed", "error", err)
			return
		}
		if changed > 0 {
			a.log.Info("reconciled subscription roles", "count", changed)
		}

		// After plans have moved on, so "your plan has ended" is never sent to
		// someone whose queued plan has just started.
		sent, err := notifications.NewRepository(a.pool).RunScheduled(ctx)
		if err != nil {
			a.log.Error("scheduled notifications failed", "error", err)
		} else if sent > 0 {
			a.log.Info("sent scheduled notifications", "count", sent)
		}
	}

	run()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
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
