package system_rbac

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moh-sso-dashboard/internal/authz"
	"go.yaml.in/yaml/v3"
)

type SeedFile struct {
	Systems    []SeedSystem    `json:"systems" yaml:"systems"`
	RealmRoles []SeedRealmRole `json:"realmRoles" yaml:"realmRoles"`
}

type SeedSystem struct {
	ClientID    string     `json:"clientId" yaml:"clientId"`
	DisplayName string     `json:"displayName" yaml:"displayName"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Icon        string     `json:"icon,omitempty" yaml:"icon,omitempty"`
	LaunchURL   string     `json:"launchUrl,omitempty" yaml:"launchUrl,omitempty"`
	Category    string     `json:"category,omitempty" yaml:"category,omitempty"`
	Enabled     *bool      `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	AccessRoles []string   `json:"accessRoles,omitempty" yaml:"accessRoles,omitempty"`
	Roles       []SeedRole `json:"roles,omitempty" yaml:"roles,omitempty"`
}

type SeedRole struct {
	Name        string   `json:"name" yaml:"name"`
	DisplayName string   `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty" yaml:"permissions,omitempty"`
}

type SeedRealmRole struct {
	Name        string   `json:"name" yaml:"name"`
	Permissions []string `json:"permissions,omitempty" yaml:"permissions,omitempty"`
}

func LoadSeedFile(path string) (SeedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SeedFile{}, err
	}

	var seed SeedFile
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		err = json.Unmarshal(data, &seed)
	default:
		err = yaml.Unmarshal(data, &seed)
	}
	if err != nil {
		return SeedFile{}, err
	}

	return seed, nil
}

func WriteSeedFile(path string, seed SeedFile) error {
	data, err := yaml.Marshal(seed)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func ValidateSeed(seed SeedFile) error {
	knownPermissions := map[string]bool{"*": true}
	for _, permission := range authz.AllPermissions {
		knownPermissions[string(permission)] = true
	}

	seenSystems := map[string]bool{}
	for _, system := range seed.Systems {
		clientID := strings.TrimSpace(system.ClientID)
		if clientID == "" {
			return fmt.Errorf("system clientId is required")
		}
		if seenSystems[clientID] {
			return fmt.Errorf("duplicate system clientId %q", clientID)
		}
		seenSystems[clientID] = true

		if strings.TrimSpace(system.DisplayName) == "" {
			return fmt.Errorf("system %q displayName is required", clientID)
		}

		seenRoles := map[string]bool{}
		for _, role := range system.Roles {
			roleName := strings.TrimSpace(role.Name)
			if roleName == "" {
				return fmt.Errorf("system %q has a role without a name", clientID)
			}
			if seenRoles[roleName] {
				return fmt.Errorf("system %q has duplicate role %q", clientID, roleName)
			}
			seenRoles[roleName] = true
			if err := validatePermissions(knownPermissions, role.Permissions); err != nil {
				return fmt.Errorf("system %q role %q: %w", clientID, roleName, err)
			}
		}
	}

	seenRealmRoles := map[string]bool{}
	for _, role := range seed.RealmRoles {
		roleName := strings.TrimSpace(role.Name)
		if roleName == "" {
			return fmt.Errorf("realm role name is required")
		}
		if seenRealmRoles[roleName] {
			return fmt.Errorf("duplicate realm role %q", roleName)
		}
		seenRealmRoles[roleName] = true
		if err := validatePermissions(knownPermissions, role.Permissions); err != nil {
			return fmt.Errorf("realm role %q: %w", roleName, err)
		}
	}

	return nil
}

func validatePermissions(knownPermissions map[string]bool, permissions []string) error {
	for _, permission := range permissions {
		permission = strings.TrimSpace(permission)
		if permission == "" {
			return fmt.Errorf("permission cannot be empty")
		}
		if !knownPermissions[permission] {
			return fmt.Errorf("unknown permission %q", permission)
		}
	}

	return nil
}

func DefaultSeed() SeedFile {
	enabled := true

	return SeedFile{
		Systems: []SeedSystem{
			{
				ClientID:    authz.SystemDashboardWeb,
				DisplayName: "Integrated Health Portal",
				Description: "Central portal shell",
				Icon:        "dashboard",
				LaunchURL:   "/portal",
				Category:    "platform",
				Enabled:     &enabled,
				AccessRoles: []string{authz.DashboardWebAccess},
				Roles: []SeedRole{
					{
						Name:        authz.DashboardWebAccess,
						DisplayName: "Dashboard Access",
						Permissions: []string{
							string(authz.PermissionPortalAccess),
							string(authz.PermissionSystemsRead),
							string(authz.PermissionSystemsLaunch),
						},
					},
					{
						Name:        authz.DashboardWebIntegratedOutbreakAccess,
						DisplayName: "Outbreak System Access",
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
						},
					},
				},
			},
			{
				ClientID:    authz.SystemIntegratedOutbreak,
				DisplayName: "Integrated Outbreak System",
				Description: "Outbreak surveillance and response",
				Icon:        "outbreak",
				LaunchURL:   "/portal/apps/dwh/surveillance",
				Category:    "surveillance",
				Enabled:     &enabled,
				AccessRoles: []string{
					authz.IntegratedOutbreakAccess,
					authz.IntegratedOutbreakViewer,
					authz.IntegratedOutbreakDataEntry,
					authz.IntegratedOutbreakSurveillanceOfficer,
					authz.IntegratedOutbreakManager,
					authz.IntegratedOutbreakAdmin,
					authz.IntegratedOutbreakSuperAdmin,
				},
				Roles: []SeedRole{
					{
						Name: authz.IntegratedOutbreakSuperAdmin,
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionOutbreakManage),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionSurveillanceImport),
							string(authz.PermissionSurveillanceManageLocations),
							string(authz.PermissionSurveillanceManageAlerts),
							string(authz.PermissionDataQualityRead),
							string(authz.PermissionDataQualityWrite),
							string(authz.PermissionDataQualityResolve),
							string(authz.PermissionDocumentsRead),
							string(authz.PermissionDocumentsWrite),
							string(authz.PermissionDocumentsProcess),
						},
					},
					{
						Name: authz.IntegratedOutbreakViewer,
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionDataQualityRead),
							string(authz.PermissionDocumentsRead),
						},
					},
				},
			},
			{
				ClientID:    authz.SystemReportBrowser,
				DisplayName: "Report Browser",
				Icon:        "reporting",
				LaunchURL:   "/portal/apps/dwh/reports",
				Category:    "reports",
				Enabled:     &enabled,
				AccessRoles: []string{authz.ReportBrowserAccess},
				Roles: []SeedRole{
					{
						Name:        authz.ReportBrowserAccess,
						DisplayName: "Report Browser Access",
						Permissions: []string{
							string(authz.PermissionReportBrowserRead),
						},
					},
				},
			},
		},
		RealmRoles: []SeedRealmRole{
			{Name: authz.RoleAdmin, Permissions: []string{"*"}},
			{
				Name: authz.RoleUser,
				Permissions: []string{
					string(authz.PermissionPortalAccess),
					string(authz.PermissionSystemsRead),
					string(authz.PermissionSystemsLaunch),
				},
			},
		},
	}
}
