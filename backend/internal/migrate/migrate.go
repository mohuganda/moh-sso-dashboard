package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

type DBConfig struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	WaitTimeout     time.Duration
}

func InitDB(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
	start := time.Now()

	db, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	// Pool tuning
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	// Wait for DB readiness
	waitCtx := ctx
	if cfg.WaitTimeout > 0 {
		var cancel context.CancelFunc
		waitCtx, cancel = context.WithTimeout(ctx, cfg.WaitTimeout)
		defer cancel()
	}

	if err := WaitForDB(waitCtx, db); err != nil {
		return nil, fmt.Errorf("database not ready: %w", err)
	}

	log.Info().
		Dur("duration", time.Since(start)).
		Msg("Database initialized successfully")

	return db, nil
}

func MigrateDB(db *sql.DB, migrateDir string) error {
	start := time.Now()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrateDir,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info().Msg("No new migrations to apply")
		} else {
			var dirtyErr migrate.ErrDirty
			if errors.As(err, &dirtyErr) {
				return fmt.Errorf("migration failed: dirty database version %d; repair the migration state before restarting the backend (inspect with `make local-db-migration-status`; for local dirty version 10 run `make local-db-repair-dirty-10`; otherwise reset local PVCs with `make local-reset-data`): %w", dirtyErr.Version, err)
			}
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Info().
		Dur("duration", time.Since(start)).
		Msg("Database migrated successfully")

	return nil
}

func DropDB(db *sql.DB, migrateDir string) error {
	start := time.Now()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		migrateDir,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("drop migration failed: %w", err)
	}

	log.Info().
		Dur("duration", time.Since(start)).
		Msg("Database dropped successfully")

	return nil
}

func PingDB(ctx context.Context, db *sql.DB) error {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return db.PingContext(pingCtx)
}

func WaitForDB(ctx context.Context, db *sql.DB) error {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for database: %w", ctx.Err())

		case <-ticker.C:
			if err := PingDB(ctx, db); err == nil {
				log.Info().Msg("Database is ready")
				return nil
			} else {
				log.Warn().Err(err).Msg("Waiting for database...")
			}
		}
	}
}
