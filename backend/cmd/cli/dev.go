package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/moh-sso-dashboard/internal/config"
	systemrbac "github.com/moh-sso-dashboard/internal/features/system_rbac"
	db "github.com/moh-sso-dashboard/internal/migrate"
	"github.com/spf13/cobra"
)

var devSeedSystemRBAC bool
var devSeedFile string

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development utilities",
}

var devSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed development data",
	Run: func(cmd *cobra.Command, args []string) {
		if !devSeedSystemRBAC {
			fmt.Println("No seed target selected. Use --system-rbac to apply system RBAC seed data.")
			return
		}

		seed, err := systemrbac.LoadSeedFile(devSeedFile)
		if err != nil {
			log.Fatalf("failed to load system RBAC seed: %v", err)
		}
		if err := systemrbac.ValidateSeed(seed); err != nil {
			log.Fatalf("system RBAC seed is invalid: %v", err)
		}

		cfg, err := config.LoadConfig(".")
		if err != nil {
			log.Fatalf("failed to load app config: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		primaryDB, err := db.InitDB(ctx, db.DBConfig{
			Driver:          cfg.DBDriver,
			DSN:             cfg.DbSource(),
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 30 * time.Minute,
			WaitTimeout:     30 * time.Second,
		})
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		defer primaryDB.Close()

		if err := db.MigrateDB(primaryDB, "file://internal/db/migrations"); err != nil {
			log.Fatalf("failed to migrate database: %v", err)
		}
		if err := systemrbac.ApplySeed(ctx, primaryDB, seed); err != nil {
			log.Fatalf("failed to apply system RBAC seed: %v", err)
		}

		fmt.Printf("System RBAC seed '%s' applied successfully\n", devSeedFile)
	},
}

func init() {
	devCmd.AddCommand(devSeedCmd)
	devSeedCmd.Flags().BoolVar(&devSeedSystemRBAC, "system-rbac", false, "Apply system RBAC seed data")
	devSeedCmd.Flags().StringVar(&devSeedFile, "file", "config/system-rbac.seed.yaml", "Seed file to apply")
}
