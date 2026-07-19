package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/moh-sso-dashboard/internal/config"
	db "github.com/moh-sso-dashboard/internal/migrate"
)

const migrationSource = "file://internal/db/migrations"

// RunMigrations applies only the primary application database migrations. It
// intentionally avoids Redis, Keycloak, remote databases, and HTTP startup so
// deployment systems can use it as a deterministic pre-start step.
func RunMigrations() error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	primaryDB, err := db.InitDB(ctx, db.DBConfig{
		Driver:          cfg.DBDriver,
		DSN:             cfg.DbSource(),
		MaxOpenConns:    2,
		MaxIdleConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
		WaitTimeout:     2 * time.Minute,
	})
	if err != nil {
		return fmt.Errorf("connect to primary database: %w", err)
	}
	defer primaryDB.Close()

	if err := db.MigrateDB(primaryDB, migrationSource); err != nil {
		return err
	}

	return nil
}
