package db

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool is the global database connection pool.
var Pool *pgxpool.Pool

// Connect initialises the PostgreSQL connection pool.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Verify the connection is alive.
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to reach database: %w", err)
	}

	slog.Info("database connected")
	return pool, nil
}

// RunMigrations applies all pending up-migrations from the given directory.
func RunMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	slog.Info("migrations applied successfully")
	return nil
}

// RunSeed runs the seed SQL file if it exists.
func RunSeed(ctx context.Context, pool *pgxpool.Pool, seedPath string) error {
	data, err := os.ReadFile(seedPath)
	if err != nil {
		// Seed file is optional — skip silently.
		slog.Info("no seed file found, skipping", "path", seedPath)
		return nil
	}

	if _, err := pool.Exec(ctx, string(data)); err != nil {
		return fmt.Errorf("seed failed: %w", err)
	}

	slog.Info("database seeded successfully")
	return nil
}
