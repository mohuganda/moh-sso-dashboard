package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/config"
	"github.com/spf13/cobra"
)

var doctorTimeout time.Duration

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run operational readiness checks for the backend",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), doctorTimeout)
		defer cancel()

		results := runDoctorChecks(ctx)
		failed := false
		for _, result := range results {
			printCheck(result)
			if !result.OK {
				failed = true
			}
		}
		if failed {
			return fmt.Errorf("one or more doctor checks failed")
		}
		return nil
	},
}

func runDoctorChecks(ctx context.Context) []checkResult {
	results := make([]checkResult, 0)

	cfg, err := config.LoadConfig(".")
	if err != nil {
		return []checkResult{{
			Name:    "config",
			OK:      false,
			Message: err.Error(),
		}}
	}
	results = append(results, checkResult{Name: "config", OK: true, Message: "application config loaded"})

	primaryDB, err := openToolDB(ctx, cfg)
	if err != nil {
		results = append(results, checkResult{Name: "database", OK: false, Message: err.Error()})
	} else {
		defer primaryDB.Close()
		results = append(results, checkResult{Name: "database", OK: true, Message: cfg.DBDriver + " connection is reachable"})
		results = append(results, migrationTableCheck(ctx, primaryDB))
		results = append(results, systemRBACTableChecks(ctx, primaryDB)...)
	}

	results = append(results, tcpCheck(ctx, "redis", net.JoinHostPort(cfg.RedisHost, cfg.RedisPort)))
	results = append(results, keycloakConfigCheck(cfg))
	results = append(results, emailConfigCheck(cfg))
	results = append(results, storageConfigCheck(cfg))

	return results
}

func migrationTableCheck(ctx context.Context, db *sql.DB) checkResult {
	exists, err := tableExists(ctx, db, "schema_migrations")
	if err != nil {
		return checkResult{Name: "migrations", OK: false, Message: err.Error()}
	}
	if !exists {
		return checkResult{Name: "migrations", OK: false, Message: "schema_migrations table is missing"}
	}
	return checkResult{Name: "migrations", OK: true, Message: "schema_migrations table exists"}
}

func systemRBACTableChecks(ctx context.Context, db *sql.DB) []checkResult {
	tables := []string{
		"ihp_systems",
		"ihp_system_roles",
		"ihp_permissions",
		"ihp_system_role_permissions",
		"ihp_realm_role_permissions",
	}
	results := make([]checkResult, 0, len(tables)+1)
	for _, table := range tables {
		exists, err := tableExists(ctx, db, table)
		if err != nil {
			results = append(results, checkResult{Name: "system-rbac:" + table, OK: false, Message: err.Error()})
			continue
		}
		if !exists {
			results = append(results, checkResult{Name: "system-rbac:" + table, OK: false, Message: "table is missing"})
			continue
		}
		results = append(results, checkResult{Name: "system-rbac:" + table, OK: true, Message: "table exists"})
	}
	if count, err := countRows(ctx, db, "ihp_systems"); err == nil {
		results = append(results, checkResult{Name: "system-rbac:seed", OK: count > 0, Message: fmt.Sprintf("%d systems configured", count)})
	}
	return results
}

func tcpCheck(ctx context.Context, name string, address string) checkResult {
	if strings.TrimSpace(address) == "" || strings.HasSuffix(address, ":") {
		return checkResult{Name: name, OK: false, Message: "host/port is not configured"}
	}

	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return checkResult{Name: name, OK: false, Message: err.Error()}
	}
	_ = conn.Close()
	return checkResult{Name: name, OK: true, Message: address + " is reachable"}
}

func keycloakConfigCheck(cfg *config.Config) checkResult {
	missing := missingValues(map[string]string{
		"KEYCLOAK_BASE_URL":            cfg.KeycloakBaseURL,
		"KEYCLOAK_REALM":               cfg.KeycloakRealm,
		"KEYCLOAK_ADMIN_CLIENT_ID":     cfg.KeycloakAdminClientID,
		"KEYCLOAK_ADMIN_CLIENT_SECRET": cfg.KeycloakAdminClientSecret,
		"KEYCLOAK_WEB_CLIENT_ID":       cfg.KeycloakWebClientID,
	})
	if len(missing) > 0 {
		return checkResult{Name: "keycloak", OK: false, Message: "missing " + strings.Join(missing, ", ")}
	}
	return checkResult{Name: "keycloak", OK: true, Message: "admin and web client config present"}
}

func emailConfigCheck(cfg *config.Config) checkResult {
	missing := missingValues(map[string]string{
		"SMTP_HOST":       cfg.SMTP.Host,
		"SMTP_FROM_EMAIL": cfg.SMTP.FromEmail,
	})
	if len(missing) > 0 || cfg.SMTP.Port <= 0 {
		if cfg.SMTP.Port <= 0 {
			missing = append(missing, "SMTP_PORT")
		}
		return checkResult{Name: "email", OK: false, Message: "missing " + strings.Join(missing, ", ")}
	}
	return checkResult{Name: "email", OK: true, Message: fmt.Sprintf("%s:%d configured", cfg.SMTP.Host, cfg.SMTP.Port)}
}

func storageConfigCheck(cfg *config.Config) checkResult {
	switch cfg.StorageProvider {
	case "", "local":
		if strings.TrimSpace(cfg.LocalBasePath) == "" {
			return checkResult{Name: "storage", OK: false, Message: "LOCAL_BASE_PATH is required for local storage"}
		}
	case "nfs":
		if strings.TrimSpace(cfg.NFSBasePath) == "" {
			return checkResult{Name: "storage", OK: false, Message: "NFS_BASE_PATH is required for nfs storage"}
		}
	case "s3":
		missing := missingValues(map[string]string{"S3_BUCKET": cfg.S3Bucket, "S3_REGION": cfg.S3Region})
		if len(missing) > 0 {
			return checkResult{Name: "storage", OK: false, Message: "missing " + strings.Join(missing, ", ")}
		}
	case "minio":
		missing := missingValues(map[string]string{"MINIO_ENDPOINT": cfg.MinioEndpoint, "MINIO_BUCKET": cfg.MinioBucket})
		if len(missing) > 0 {
			return checkResult{Name: "storage", OK: false, Message: "missing " + strings.Join(missing, ", ")}
		}
	default:
		return checkResult{Name: "storage", OK: false, Message: "unknown provider " + cfg.StorageProvider}
	}
	return checkResult{Name: "storage", OK: true, Message: cfg.StorageProvider + " storage configured"}
}

func missingValues(values map[string]string) []string {
	missing := make([]string, 0)
	for key, value := range values {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

func init() {
	doctorCmd.Flags().DurationVar(&doctorTimeout, "timeout", 15*time.Second, "Maximum time to spend on doctor checks")
}
