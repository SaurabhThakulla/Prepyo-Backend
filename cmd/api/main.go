// Command api is the Prepyo backend: one Go binary serving the whole REST API.
//
// main does startup and shutdown only. Wiring lives in app.go and every rule
// about how the product behaves lives in internal/.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/pkg/config"
	"github.com/prepyo/backend/pkg/logger"
)

func main() {
	// run returns an error rather than exiting inline, so deferred cleanup
	// such as closing the connection pool actually runs.
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "prepyo api: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.Env)
	log.Info("starting prepyo api", "env", cfg.Env, "port", cfg.Port, "aiEnabled", cfg.AIEnabled())

	if !cfg.AIEnabled() {
		log.Warn("OPENROUTER_API_KEY is not set: evaluation and tutor endpoints will return 503")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, log); err != nil {
		return err
	}

	app := newApp(cfg, pool, log)
	go app.cleanExpiredSessions(ctx)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           app.router(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		// Long enough for a synchronous evaluation to finish while the learner
		// waits. Moving evaluation onto a queue is what would bring this down.
		WriteTimeout: cfg.AIRequestTimeout + 30*time.Second,
		IdleTimeout:  120 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	// Stop accepting new connections but let in-flight requests finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed, closing now", "error", err)
		return server.Close()
	}
	log.Info("stopped cleanly")
	return nil
}
