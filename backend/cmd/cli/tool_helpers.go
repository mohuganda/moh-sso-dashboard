package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/moh-sso-dashboard/internal/config"
)

type checkResult struct {
	Name    string
	OK      bool
	Message string
}

func printCheck(result checkResult) {
	status := "FAIL"
	if result.OK {
		status = "OK"
	}
	fmt.Printf("[%s] %s: %s\n", status, result.Name, result.Message)
}

func openToolDB(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	db, err := sql.Open(cfg.DBDriver, cfg.DbSource())
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func tableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`, table).Scan(&exists)
	return exists, err
}

func countRows(ctx context.Context, db *sql.DB, table string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
