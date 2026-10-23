package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/prepyo/backend/internal/database"
	"github.com/prepyo/backend/internal/media"
	"github.com/prepyo/backend/internal/questions"
	"github.com/prepyo/backend/pkg/config"
	"github.com/prepyo/backend/pkg/logger"
)

// runAssets is `server assets [-restore] [-yes] [-vacuum] [-dir folder]`: it moves question
// audio and images from the database to the media bucket (or back with -restore).
// Without -yes it only counts what it would move.
func runAssets(args []string) error {
	fs := flag.NewFlagSet("assets", flag.ContinueOnError)
	restore := fs.Bool("restore", false, "copy files back from the bucket into the database")
	yes := fs.Bool("yes", false, "make the change (otherwise a dry run that only counts)")
	dir := fs.String("dir", "", "use this local folder instead of the bucket (for testing)")
	vacuum := fs.Bool("vacuum", false, "after moving, compact the table so the database gets the space back")
	if err := fs.Parse(args); err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log := logger.New(cfg.Env)
	store := media.New(cfg)
	if *dir != "" {
		store = media.Dir(*dir)
	}
	if store == nil {
		return errors.New("no media bucket configured (S3_*), and no -dir given")
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if *restore {
		n, err := questions.RestoreAssetsFromStore(ctx, pool, store, *yes, log)
		fmt.Printf("restore: %d files, write=%v\n", n, *yes)
		return err
	}
	fmt.Println("table size before:", tableSize(ctx, pool))
	n, size, err := questions.MoveAssetsToStore(ctx, pool, store, *yes, log)
	fmt.Printf("move: %d files, %.1f MB, write=%v\n", n, float64(size)/1e6, *yes)
	if err != nil {
		return err
	}
	if *vacuum && *yes {
		// Rewrites the table so the freed space goes back to the database; locks it for a moment.
		if _, err := pool.Exec(ctx, "VACUUM FULL question_assets"); err != nil {
			return fmt.Errorf("vacuum: %w", err)
		}
	}
	fmt.Println("table size after:", tableSize(ctx, pool))
	return nil
}

func tableSize(ctx context.Context, pool *pgxpool.Pool) string {
	var size string
	if err := pool.QueryRow(ctx, "SELECT pg_size_pretty(pg_total_relation_size('question_assets'))").Scan(&size); err != nil {
		return "unknown (" + err.Error() + ")"
	}
	return size
}
