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

var systemRBACFile string

var systemRBACCmd = &cobra.Command{
	Use:   "system-rbac",
	Short: "Manage Integrated Health Portal system RBAC mappings",
}

var systemRBACValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a system RBAC seed file",
	Run: func(cmd *cobra.Command, args []string) {
		seed, err := systemrbac.LoadSeedFile(systemRBACFile)
		if err != nil {
			log.Fatalf("❌ Failed to load system RBAC seed file: %v", err)
		}

		if err := systemrbac.ValidateSeed(seed); err != nil {
			log.Fatalf("❌ System RBAC seed is invalid: %v", err)
		}

		fmt.Printf("✅ System RBAC seed '%s' is valid\n", systemRBACFile)
	},
}

var systemRBACSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Apply a system RBAC seed file to the portal database",
	Run: func(cmd *cobra.Command, args []string) {
		seed, err := systemrbac.LoadSeedFile(systemRBACFile)
		if err != nil {
			log.Fatalf("❌ Failed to load system RBAC seed file: %v", err)
		}

		if err := systemrbac.ValidateSeed(seed); err != nil {
			log.Fatalf("❌ System RBAC seed is invalid: %v", err)
		}

		cfg, err := config.LoadConfig(".")
		if err != nil {
			log.Fatalf("❌ Failed to load app config: %v", err)
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
			log.Fatalf("❌ Failed to connect to database: %v", err)
		}
		defer primaryDB.Close()

		if err := db.MigrateDB(primaryDB, "file://internal/db/migrations"); err != nil {
			log.Fatalf("❌ Failed to migrate database: %v", err)
		}

		if err := systemrbac.ApplySeed(ctx, primaryDB, seed); err != nil {
			log.Fatalf("❌ Failed to apply system RBAC seed: %v", err)
		}

		fmt.Printf("✅ System RBAC seed '%s' applied successfully\n", systemRBACFile)
	},
}

var systemRBACExportDefaultCmd = &cobra.Command{
	Use:   "export-default",
	Short: "Write the built-in default system RBAC seed file",
	Run: func(cmd *cobra.Command, args []string) {
		if err := systemrbac.WriteSeedFile(systemRBACFile, systemrbac.DefaultSeed()); err != nil {
			log.Fatalf("❌ Failed to write default system RBAC seed file: %v", err)
		}

		fmt.Printf("✅ Default system RBAC seed written to '%s'\n", systemRBACFile)
	},
}

func init() {
	systemRBACCmd.AddCommand(systemRBACValidateCmd)
	systemRBACCmd.AddCommand(systemRBACSeedCmd)
	systemRBACCmd.AddCommand(systemRBACExportDefaultCmd)

	for _, cmd := range []*cobra.Command{
		systemRBACValidateCmd,
		systemRBACSeedCmd,
		systemRBACExportDefaultCmd,
	} {
		cmd.Flags().StringVar(&systemRBACFile, "file", "config/system-rbac.seed.yaml", "System RBAC seed file")
	}
}
