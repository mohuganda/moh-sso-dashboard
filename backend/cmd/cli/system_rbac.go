package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/config"
	systemrbac "github.com/moh-sso-dashboard/internal/features/system_rbac"
	"github.com/moh-sso-dashboard/internal/keycloak"
	db "github.com/moh-sso-dashboard/internal/migrate"
	"github.com/spf13/cobra"
)

var systemRBACFile string
var systemRBACRealmRoles []string
var systemRBACClientRoles []string
var systemRBACDraftFile string
var systemRBACApply bool

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

var systemRBACDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system RBAC seed and database readiness",
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, err := systemrbac.LoadSeedFile(systemRBACFile)
		if err != nil {
			return fmt.Errorf("load seed file: %w", err)
		}
		if err := systemrbac.ValidateSeed(seed); err != nil {
			return fmt.Errorf("validate seed file: %w", err)
		}
		printCheck(checkResult{Name: "system-rbac:seed-file", OK: true, Message: systemRBACFile + " is valid"})

		cfg, err := config.LoadConfig(".")
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()

		primaryDB, err := openToolDB(ctx, cfg)
		if err != nil {
			return fmt.Errorf("database: %w", err)
		}
		defer primaryDB.Close()

		failed := false
		for _, result := range systemRBACTableChecks(ctx, primaryDB) {
			printCheck(result)
			if !result.OK {
				failed = true
			}
		}
		if failed {
			return fmt.Errorf("system RBAC database checks failed")
		}
		return nil
	},
}

var systemRBACExplainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain effective permissions for realm/client roles",
	Run: func(cmd *cobra.Command, args []string) {
		clientRoles := parseClientRoleFlags(systemRBACClientRoles)
		access := authz.NewContext("cli-preview", systemRBACRealmRoles, clientRoles)

		fmt.Println("Realm roles:")
		for _, role := range access.RealmRoles {
			fmt.Printf("  - %s\n", role)
		}
		fmt.Println("Client roles:")
		for clientID, roles := range access.ClientRoles {
			fmt.Printf("  - %s: %s\n", clientID, strings.Join(roles, ", "))
		}
		fmt.Println("Permissions:")
		for _, permission := range access.PermissionStrings() {
			fmt.Printf("  - %s\n", permission)
		}
		fmt.Println("Accessible systems:")
		for _, system := range access.SystemAccess() {
			fmt.Printf("  - %s (%s)\n", system.ClientID, strings.Join(system.Roles, ", "))
		}
	},
}

var systemRBACUnmappedCmd = &cobra.Command{
	Use:   "unmapped",
	Short: "List Keycloak clients or roles not represented in the seed file",
	RunE: func(cmd *cobra.Command, args []string) error {
		seed, err := systemrbac.LoadSeedFile(systemRBACFile)
		if err != nil {
			return err
		}
		if err := systemrbac.ValidateSeed(seed); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()

		kc, err := newAdminKCForTool()
		if err != nil {
			return err
		}

		draft, err := draftSystemRBACSeed(ctx, kc)
		if err != nil {
			return err
		}
		reportUnmapped(seed, draft)
		return nil
	},
}

var systemRBACSyncKeycloakCmd = &cobra.Command{
	Use:   "sync-keycloak",
	Short: "Draft or apply system RBAC mappings from Keycloak clients and roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 60*time.Second)
		defer cancel()

		kc, err := newAdminKCForTool()
		if err != nil {
			return err
		}

		draft, err := draftSystemRBACSeed(ctx, kc)
		if err != nil {
			return err
		}
		if err := systemrbac.ValidateSeed(draft); err != nil {
			return err
		}

		if systemRBACDraftFile != "" {
			if err := systemrbac.WriteSeedFile(systemRBACDraftFile, draft); err != nil {
				return err
			}
			fmt.Printf("Draft seed written to %s\n", systemRBACDraftFile)
		}

		if !systemRBACApply {
			fmt.Printf("Dry run complete. Discovered %d systems and %d realm roles.\n", len(draft.Systems), len(draft.RealmRoles))
			return nil
		}

		cfg, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		primaryDB, err := openToolDB(ctx, cfg)
		if err != nil {
			return err
		}
		defer primaryDB.Close()

		if err := systemrbac.ApplySeed(ctx, primaryDB, draft); err != nil {
			return err
		}
		fmt.Println("Keycloak RBAC draft applied to the portal database")
		return nil
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
	systemRBACCmd.AddCommand(systemRBACDoctorCmd)
	systemRBACCmd.AddCommand(systemRBACExplainCmd)
	systemRBACCmd.AddCommand(systemRBACUnmappedCmd)
	systemRBACCmd.AddCommand(systemRBACSyncKeycloakCmd)
	systemRBACCmd.AddCommand(systemRBACExportDefaultCmd)

	for _, cmd := range []*cobra.Command{
		systemRBACValidateCmd,
		systemRBACSeedCmd,
		systemRBACDoctorCmd,
		systemRBACUnmappedCmd,
		systemRBACExportDefaultCmd,
	} {
		cmd.Flags().StringVar(&systemRBACFile, "file", "config/system-rbac.seed.yaml", "System RBAC seed file")
	}

	systemRBACExplainCmd.Flags().StringArrayVar(&systemRBACRealmRoles, "realm-role", nil, "Realm role to include in the explanation")
	systemRBACExplainCmd.Flags().StringArrayVar(&systemRBACClientRoles, "client-role", nil, "Client role in the form client-id:role")

	systemRBACSyncKeycloakCmd.Flags().StringVar(&systemRBACDraftFile, "draft-file", "", "Optional path to write the generated seed draft")
	systemRBACSyncKeycloakCmd.Flags().BoolVar(&systemRBACApply, "apply", false, "Apply generated mappings to the database")
}

func newAdminKCForTool() (keycloakRBACClient, error) {
	if baseURL == "" || realm == "" || adminID == "" || adminSecret == "" {
		return nil, fmt.Errorf("provide --base-url, --realm, --admin-client-id, and --admin-client-secret")
	}
	kc := keycloak.NewAdminClient(baseURL, realm, adminID, adminSecret)
	if err := kc.Authenticate(); err != nil {
		return nil, fmt.Errorf("keycloak admin auth failed: %w", err)
	}
	return kc, nil
}

type keycloakRBACClient interface {
	ListClients() ([]keycloak.ClientInfo, error)
	ListRealmRoles(context.Context) ([]keycloak.RoleRep, error)
	ListClientRoles(context.Context, string) ([]keycloak.ClientRoleRep, error)
}

func parseClientRoleFlags(values []string) map[string][]string {
	out := map[string][]string{}
	for _, value := range values {
		clientID, role, ok := strings.Cut(value, ":")
		if !ok {
			continue
		}
		clientID = strings.TrimSpace(clientID)
		role = strings.TrimSpace(role)
		if clientID == "" || role == "" {
			continue
		}
		out[clientID] = append(out[clientID], role)
	}
	return out
}

func draftSystemRBACSeed(ctx context.Context, kc keycloakRBACClient) (systemrbac.SeedFile, error) {
	clients, err := kc.ListClients()
	if err != nil {
		return systemrbac.SeedFile{}, err
	}

	realmRoles, err := kc.ListRealmRoles(ctx)
	if err != nil {
		return systemrbac.SeedFile{}, err
	}

	seed := systemrbac.SeedFile{
		Systems:    make([]systemrbac.SeedSystem, 0, len(clients)),
		RealmRoles: make([]systemrbac.SeedRealmRole, 0, len(realmRoles)),
	}

	for _, role := range realmRoles {
		name := strings.TrimSpace(role.Name)
		if name == "" || strings.HasPrefix(name, "default-roles-") || strings.HasPrefix(name, "uma_") {
			continue
		}
		seed.RealmRoles = append(seed.RealmRoles, systemrbac.SeedRealmRole{Name: name})
	}

	for _, client := range clients {
		clientID := strings.TrimSpace(client.ClientID)
		if clientID == "" || strings.HasPrefix(clientID, "account") || strings.HasPrefix(clientID, "realm-management") || strings.HasPrefix(clientID, "security-admin-console") || strings.HasPrefix(clientID, "admin-cli") {
			continue
		}

		roles, err := kc.ListClientRoles(ctx, clientID)
		if err != nil {
			return systemrbac.SeedFile{}, err
		}

		enabled := client.Enabled
		system := systemrbac.SeedSystem{
			ClientID:    clientID,
			DisplayName: firstNonEmpty(client.Name, clientID),
			Enabled:     &enabled,
			AccessRoles: make([]string, 0),
			Roles:       make([]systemrbac.SeedRole, 0, len(roles)),
		}
		for _, role := range roles {
			roleName := strings.TrimSpace(role.Name)
			if roleName == "" {
				continue
			}
			system.AccessRoles = append(system.AccessRoles, roleName)
			system.Roles = append(system.Roles, systemrbac.SeedRole{
				Name:        roleName,
				Description: role.Description,
			})
		}
		seed.Systems = append(seed.Systems, system)
	}

	sort.Slice(seed.Systems, func(i, j int) bool { return seed.Systems[i].ClientID < seed.Systems[j].ClientID })
	sort.Slice(seed.RealmRoles, func(i, j int) bool { return seed.RealmRoles[i].Name < seed.RealmRoles[j].Name })
	return seed, nil
}

func reportUnmapped(current systemrbac.SeedFile, discovered systemrbac.SeedFile) {
	knownSystems := map[string]bool{}
	for _, system := range current.Systems {
		knownSystems[system.ClientID] = true
	}
	fmt.Println("Unmapped systems:")
	for _, system := range discovered.Systems {
		if !knownSystems[system.ClientID] {
			fmt.Printf("  - %s\n", system.ClientID)
		}
	}

	knownRealmRoles := map[string]bool{}
	for _, role := range current.RealmRoles {
		knownRealmRoles[role.Name] = true
	}
	fmt.Println("Unmapped realm roles:")
	for _, role := range discovered.RealmRoles {
		if !knownRealmRoles[role.Name] {
			fmt.Printf("  - %s\n", role.Name)
		}
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
