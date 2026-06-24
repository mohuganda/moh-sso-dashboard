package main

import (
	"fmt"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect and validate backend configuration",
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate backend application configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig(".")
		if err != nil {
			return err
		}

		fmt.Println("Backend config is valid")
		fmt.Printf("Environment: %s\n", cfg.Environment)
		fmt.Printf("Database: %s@%s:%s/%s\n", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)
		fmt.Printf("Keycloak realm: %s\n", cfg.KeycloakRealm)
		fmt.Printf("Storage provider: %s\n", cfg.StorageProvider)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}
