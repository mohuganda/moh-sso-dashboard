package system_rbac

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	ClientID          string     `json:"clientId" yaml:"clientId"`
	DisplayName       string     `json:"displayName" yaml:"displayName"`
	Description       string     `json:"description,omitempty" yaml:"description,omitempty"`
	Icon              string     `json:"icon,omitempty" yaml:"icon,omitempty"`
	LaunchURL         string     `json:"launchUrl,omitempty" yaml:"launchUrl,omitempty"`
	Category          string     `json:"category,omitempty" yaml:"category,omitempty"`
	OwnerTeam         string     `json:"ownerTeam,omitempty" yaml:"ownerTeam,omitempty"`
	OwnerName         string     `json:"ownerName,omitempty" yaml:"ownerName,omitempty"`
	OwnerEmail        string     `json:"ownerEmail,omitempty" yaml:"ownerEmail,omitempty"`
	SupportURL        string     `json:"supportUrl,omitempty" yaml:"supportUrl,omitempty"`
	DocumentationURL  string     `json:"documentationUrl,omitempty" yaml:"documentationUrl,omitempty"`
	Environment       string     `json:"environment,omitempty" yaml:"environment,omitempty"`
	Criticality       string     `json:"criticality,omitempty" yaml:"criticality,omitempty"`
	Navigation        string     `json:"navigation,omitempty" yaml:"navigation,omitempty"`
	SystemType        string     `json:"systemType,omitempty" yaml:"systemType,omitempty"`
	DisplayInLauncher *bool      `json:"displayInLauncher,omitempty" yaml:"displayInLauncher,omitempty"`
	DisplayInSideNav  *bool      `json:"displayInSideNav,omitempty" yaml:"displayInSideNav,omitempty"`
	LaunchMode        string     `json:"launchMode,omitempty" yaml:"launchMode,omitempty"`
	Enabled           *bool      `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	SortOrder         int32      `json:"sortOrder,omitempty" yaml:"sortOrder,omitempty"`
	AccessRoles       []string   `json:"accessRoles,omitempty" yaml:"accessRoles,omitempty"`
	Roles             []SeedRole `json:"roles,omitempty" yaml:"roles,omitempty"`
}

type SeedRole struct {
	Name        string   `json:"name" yaml:"name"`
	DisplayName string   `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty" yaml:"permissions,omitempty"`
}

type SeedRealmRole struct {
	Name        string              `json:"name" yaml:"name"`
	Permissions []string            `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	SystemRoles map[string][]string `json:"systemRoles,omitempty" yaml:"systemRoles,omitempty"`
}

const dataStatisticsNavigation = `[
  {"id":"dashboards","label":"Dashboards","path":"/apps/dwh/dashboards","permission":"report_browser:read"},
  {"id":"data-validation","label":"Data Validation","path":"/apps/dwh/data-validation","permission":"data_quality:read"},
  {"id":"data-visualizer","label":"Data Visualizer","path":"/apps/dwh/data-visualizer","permission":"data_quality:read"},
  {"id":"documents","label":"Document Management","path":"/apps/dwh/documents","permission":"documents:read"},
  {"id":"surveillance","label":"Surveillance","path":"/apps/dwh/surveillance","permission":"surveillance:read"},
  {"id":"issue-tracker","label":"Issue Tracking","path":"/apps/dwh/issue-tracker","permission":"issue_tracker:read"}
]`

const utilitiesNavigation = `[
  {"id":"self-service","label":"Self Service","children":[
    {"id":"my-timesheet","label":"My Timesheet","path":"/apps/utilities/self-service/timesheet"},
    {"id":"elearning","label":"eLearning","path":"/apps/utilities/self-service/elearning"},
    {"id":"leave-plan","label":"Leave Plan","path":"/apps/utilities/self-service/leave-plan"},
    {"id":"absence-requests","label":"Absence Requests","path":"/apps/utilities/self-service/absence-requests"},
    {"id":"absence-dashboard","label":"My Absence Dashboard","path":"/apps/utilities/self-service/absence-dashboard"},
    {"id":"eservice-requests","label":"eService Requests","children":[
      {"id":"document-upload","label":"Document Upload","path":"/apps/utilities/self-service/eservice/document-upload","permission":"documents:write"},
      {"id":"service-access","label":"Service Access","path":"/apps/utilities/self-service/eservice/service-access"},
      {"id":"equipment-request","label":"Equipment Request","path":"/apps/utilities/self-service/eservice/equipment-request"}
    ]}
  ]}
]`

const caseRegistersNavigation = `[
  {"id":"external-referrals","label":"External Referrals","path":"/apps/case-registers/external-referrals"},
  {"id":"disease-registers","label":"Disease Registers","path":"/apps/case-registers/disease-registers"}
]`

const outbreakManagementNavigation = `[
  {"id":"signals-alerts","label":"Signals & Alerts","path":"/apps/outbreak-management/signals-alerts"},
  {"id":"poe-management","label":"PoE Management","path":"/apps/outbreak-management/poe-management"},
  {"id":"case-management","label":"Case Management","path":"/apps/outbreak-management/case-management"}
]`

const referenceRegistersNavigation = `[
  {"id":"facility-register","label":"Facility Register","path":"/apps/reference-registers/facility-register"},
  {"id":"terminology-service","label":"Terminology Service","children":[
    {"id":"test-menu","label":"Test Menu","path":"/apps/reference-registers/terminology/test-menu"},
    {"id":"pharmaceuticals","label":"Pharmaceuticals","path":"/apps/reference-registers/terminology/pharmaceuticals"},
    {"id":"procedures","label":"Procedures","path":"/apps/reference-registers/terminology/procedures"},
    {"id":"equipment","label":"Equipment","path":"/apps/reference-registers/terminology/equipment"}
  ]}
]`

const eServicesNavigation = `[
  {"id":"ihris","label":"iHRIS","path":"/apps/eservices/ihris"},
  {"id":"meeting-manager","label":"Meeting Manager","path":"/apps/eservices/meeting-manager"},
  {"id":"action-tracker","label":"Action Tracker","path":"/apps/eservices/action-tracker"},
  {"id":"clinician-outputs","label":"Clinician Outputs","path":"/apps/eservices/clinician-outputs"},
  {"id":"leave-absence","label":"Leave & Absence Management","path":"/apps/eservices/leave-absence"},
  {"id":"workplans","label":"Workplans","path":"/apps/eservices/workplans"},
  {"id":"budget-tracker","label":"Budget Tracker","path":"/apps/eservices/budget-tracker"},
  {"id":"activity-reporting","label":"Activity Reporting","path":"/apps/eservices/activity-reporting"},
  {"id":"partner-management","label":"Partner Management","path":"/apps/eservices/partner-management"},
  {"id":"observatory-uploads","label":"Observatory Uploads","path":"/apps/eservices/observatory-uploads"}
]`

const researchStudiesNavigation = `[
  {"id":"studies","label":"Studies","path":"/apps/research-studies/studies"},
  {"id":"datasets","label":"Datasets","path":"/apps/research-studies/datasets"},
  {"id":"ethics-approvals","label":"Ethics & Approvals","path":"/apps/research-studies/ethics"},
  {"id":"publications","label":"Publications","path":"/apps/research-studies/publications"}
]`

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
	systemRoles := map[string]map[string]bool{}
	systemAccessRoles := map[string]map[string]bool{}
	for _, system := range seed.Systems {
		system = NormalizeSystemBehavior(system)
		clientID := strings.ToLower(strings.TrimSpace(system.ClientID))
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
		if err := ValidateSystemBehavior(system); err != nil {
			return fmt.Errorf("system %q: %w", clientID, err)
		}

		seenRoles := map[string]bool{}
		for _, role := range system.Roles {
			roleName := strings.ToLower(strings.TrimSpace(role.Name))
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
		systemRoles[clientID] = seenRoles
		accessRoles := map[string]bool{}
		for _, roleName := range system.AccessRoles {
			accessRoles[strings.ToLower(strings.TrimSpace(roleName))] = true
		}
		systemAccessRoles[clientID] = accessRoles
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
		for clientID, roles := range role.SystemRoles {
			clientID = strings.ToLower(strings.TrimSpace(clientID))
			if !seenSystems[clientID] {
				return fmt.Errorf("realm role %q references unknown system %q", roleName, clientID)
			}
			for _, systemRole := range roles {
				systemRole = strings.ToLower(strings.TrimSpace(systemRole))
				if !systemRoles[clientID][systemRole] {
					return fmt.Errorf("realm role %q references unknown role %q for system %q", roleName, systemRole, clientID)
				}
				if !systemAccessRoles[clientID][systemRole] {
					return fmt.Errorf("realm role %q default role %q is not an access role for system %q", roleName, systemRole, clientID)
				}
			}
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

func NormalizeSystemBehavior(system SeedSystem) SeedSystem {
	system.SystemType = strings.ToLower(strings.TrimSpace(system.SystemType))
	system.LaunchMode = strings.ToLower(strings.TrimSpace(system.LaunchMode))
	system.LaunchURL = strings.TrimSpace(system.LaunchURL)
	system.Navigation = strings.TrimSpace(system.Navigation)

	if system.SystemType == "" {
		if isHTTPURL(system.LaunchURL) && system.Navigation == "" {
			system.SystemType = "external"
		} else {
			system.SystemType = "platform"
		}
	}
	if system.LaunchMode == "" {
		if system.SystemType == "external" {
			system.LaunchMode = "new_tab"
		} else {
			system.LaunchMode = "internal"
		}
	}
	if system.DisplayInLauncher == nil {
		value := true
		system.DisplayInLauncher = &value
	}
	if system.DisplayInSideNav == nil {
		value := system.SystemType == "platform" && system.Navigation != ""
		system.DisplayInSideNav = &value
	}
	return system
}

func ValidateSystemBehavior(system SeedSystem) error {
	system = NormalizeSystemBehavior(system)
	if system.SystemType != "platform" && system.SystemType != "external" {
		return fmt.Errorf("systemType must be platform or external")
	}
	if system.LaunchMode != "internal" && system.LaunchMode != "new_tab" && system.LaunchMode != "same_tab" {
		return fmt.Errorf("launchMode must be internal, new_tab, or same_tab")
	}
	if system.SystemType == "platform" {
		if system.LaunchMode != "internal" {
			return fmt.Errorf("platform systems must use internal launchMode")
		}
		if system.LaunchURL != "" && !isPortalPath(system.LaunchURL) {
			return fmt.Errorf("platform launchUrl must begin with /portal or /apps")
		}
	} else {
		if system.LaunchMode == "internal" {
			return fmt.Errorf("external systems must use new_tab or same_tab launchMode")
		}
		if !isHTTPURL(system.LaunchURL) {
			return fmt.Errorf("external launchUrl must be an absolute HTTP or HTTPS URL")
		}
		if system.Navigation != "" || (system.DisplayInSideNav != nil && *system.DisplayInSideNav) {
			return fmt.Errorf("external systems cannot define portal side navigation")
		}
	}
	if system.DisplayInSideNav != nil && *system.DisplayInSideNav {
		if system.SystemType != "platform" {
			return fmt.Errorf("side navigation requires a platform system")
		}
		var items []map[string]any
		if err := json.Unmarshal([]byte(system.Navigation), &items); err != nil || len(items) == 0 {
			return fmt.Errorf("side navigation requires a non-empty navigation JSON array")
		}
	}
	return nil
}

func isPortalPath(value string) bool {
	return value == "/portal" || strings.HasPrefix(value, "/portal/") || value == "/apps" || strings.HasPrefix(value, "/apps/")
}

func isHTTPURL(value string) bool {
	if strings.HasPrefix(value, "//") {
		return false
	}
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != "" && parsed.User == nil
}

func DefaultSeed() SeedFile {
	enabled := true

	seed := SeedFile{
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
						Name:        authz.DashboardWebAdmin,
						DisplayName: "Portal Admin",
						Permissions: []string{
							"*",
						},
					},
					{
						Name:        authz.DashboardWebManager,
						DisplayName: "Portal Manager",
						Permissions: []string{
							string(authz.PermissionPortalAccess),
							string(authz.PermissionSystemsRead),
							string(authz.PermissionSystemsLaunch),
							string(authz.PermissionUsersRead),
							string(authz.PermissionClientsRead),
							string(authz.PermissionAnnouncementsRead),
							string(authz.PermissionDocumentsRead),
							string(authz.PermissionDocumentTemplatesRead),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionReportBrowserRead),
							string(authz.PermissionNotificationsRead),
						},
					},
					{
						Name:        authz.DashboardWebUser,
						DisplayName: "Portal User",
						Permissions: []string{
							string(authz.PermissionPortalAccess),
							string(authz.PermissionSystemsRead),
							string(authz.PermissionSystemsLaunch),
						},
					},
				},
			},
			{
				ClientID:    authz.SystemOutbreakManagement,
				DisplayName: "Outbreak Management",
				Description: "Signals, alerts, surveillance, points of entry, and case management.",
				Icon:        "warning-alt",
				LaunchURL:   "/portal/apps/outbreak-management",
				Category:    "surveillance",
				Navigation:  outbreakManagementNavigation,
				Enabled:     &enabled,
				AccessRoles: []string{
					authz.OutbreakManagementAccess,
					authz.OutbreakManagementViewer,
					authz.OutbreakManagementDataEntry,
					authz.OutbreakManagementSurveillanceOfficer,
					authz.OutbreakManagementManager,
					authz.OutbreakManagementAdmin,
					authz.OutbreakManagementSuperAdmin,
				},
				Roles: []SeedRole{
					{
						Name:        authz.OutbreakManagementAccess,
						DisplayName: "Outbreak Management Access",
						Permissions: []string{
							string(authz.PermissionPortalAccess),
							string(authz.PermissionSystemsRead),
							string(authz.PermissionSystemsLaunch),
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
						},
					},
					{
						Name: authz.OutbreakManagementSuperAdmin,
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
						Name: authz.OutbreakManagementAdmin,
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
						},
					},
					{
						Name: authz.OutbreakManagementManager,
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionDataQualityRead),
							string(authz.PermissionDataQualityWrite),
							string(authz.PermissionDocumentsRead),
						},
					},
					{
						Name: authz.OutbreakManagementViewer,
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionDataQualityRead),
							string(authz.PermissionDocumentsRead),
						},
					},
					{
						Name: authz.OutbreakManagementDataEntry,
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionSurveillanceImport),
							string(authz.PermissionDataQualityRead),
							string(authz.PermissionDataQualityWrite),
						},
					},
					{
						Name: authz.OutbreakManagementLabTechnician,
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionDataQualityRead),
							string(authz.PermissionDataQualityWrite),
						},
					},
					{
						Name: authz.OutbreakManagementSurveillanceOfficer,
						Permissions: []string{
							string(authz.PermissionOutbreakAccess),
							string(authz.PermissionSurveillanceRead),
							string(authz.PermissionSurveillanceImport),
							string(authz.PermissionSurveillanceManageAlerts),
						},
					},
				},
			},
			defaultDataStatisticsSystem(enabled),
			defaultPortalSystem("eservices", "eServices", "Digital service requests and operational tools.", "application", "/portal/apps/eservices", "services", eServicesNavigation, nil, enabled),
			defaultPortalSystem("research-studies", "Research & Studies", "Research studies, datasets, ethics approvals, and publications.", "microscope", "/portal/apps/research-studies", "research", researchStudiesNavigation, nil, enabled),
			defaultPortalSystem("case-registers", "Case Registers", "Case registers and referral tracking.", "document", "/portal/apps/case-registers", "clinical", caseRegistersNavigation, []string{
				string(authz.PermissionDocumentsRead),
			}, enabled),
			defaultPortalSystem("reference-registers", "Reference Registers", "Facility, terminology, and reference data registers.", "catalog", "/portal/apps/reference-registers", "registry", referenceRegistersNavigation, []string{
				string(authz.PermissionSystemsRead),
			}, enabled),
			defaultPortalSystem("utilities", "Utilities", "Self-service utilities and staff tools.", "tools", "/portal/apps/utilities/self-service", "utilities", utilitiesNavigation, nil, enabled),
			defaultSettingsSystem(enabled),
		},
		RealmRoles: []SeedRealmRole{
			{Name: authz.RoleAdmin, Permissions: []string{"*"}},
			{
				Name: authz.RoleUser,
				Permissions: []string{
					string(authz.PermissionPortalAccess),
					string(authz.PermissionSystemsRead),
					string(authz.PermissionSystemsLaunch),
					string(authz.PermissionDataQualityRead),
					string(authz.PermissionDocumentsRead),
					string(authz.PermissionSurveillanceRead),
					string(authz.PermissionReportBrowserRead),
				},
				SystemRoles: map[string][]string{
					authz.SystemDataStatistics: {authz.DataStatisticsAccess},
					authz.SystemUtilities:      {authz.UtilitiesAccess},
					authz.SystemSettings:       {authz.SettingsAccess},
				},
			},
			{
				Name: authz.RoleManager,
				Permissions: []string{
					string(authz.PermissionPortalAccess),
					string(authz.PermissionSystemsRead),
					string(authz.PermissionSystemsLaunch),
					string(authz.PermissionUsersRead),
					string(authz.PermissionClientsRead),
					string(authz.PermissionAnnouncementsRead),
					string(authz.PermissionDocumentsRead),
					string(authz.PermissionSurveillanceRead),
					string(authz.PermissionDataQualityRead),
					string(authz.PermissionDataQualityWrite),
					string(authz.PermissionNotificationsRead),
				},
			},
		},
	}
	applyDefaultSortOrder(seed.Systems)
	for index := range seed.Systems {
		seed.Systems[index] = NormalizeSystemBehavior(seed.Systems[index])
	}
	return seed
}

func applyDefaultSortOrder(systems []SeedSystem) {
	sortOrders := map[string]int32{
		authz.SystemDashboardWeb:    0,
		"outbreak-management":       10,
		authz.SystemDataStatistics:  20,
		"eservices":                 30,
		"research-studies":          40,
		"case-registers":            50,
		"reference-registers":       60,
		"utilities":                 90,
		"settings":                  100,
		"demo-platform-system":      900,
		"external-knowledge-system": 910,
		"external-lab-portal":       920,
	}

	for index := range systems {
		if systems[index].SortOrder != 0 {
			continue
		}
		if sortOrder, ok := sortOrders[systems[index].ClientID]; ok {
			systems[index].SortOrder = sortOrder
		}
	}
}

func defaultPortalSystem(
	clientID string,
	displayName string,
	description string,
	icon string,
	launchURL string,
	category string,
	navigation string,
	extraPermissions []string,
	enabled bool,
) SeedSystem {
	accessRole := clientID + "_access"
	permissions := []string{
		string(authz.PermissionPortalAccess),
		string(authz.PermissionSystemsRead),
		string(authz.PermissionSystemsLaunch),
	}
	permissions = append(permissions, extraPermissions...)

	return SeedSystem{
		ClientID:    clientID,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		LaunchURL:   launchURL,
		Category:    category,
		Navigation:  navigation,
		Enabled:     &enabled,
		AccessRoles: []string{accessRole},
		Roles: []SeedRole{
			{
				Name:        accessRole,
				DisplayName: displayName + " Access",
				Permissions: permissions,
			},
		},
	}
}

func defaultSettingsSystem(enabled bool) SeedSystem {
	system := defaultPortalSystem(
		"settings",
		"Settings",
		"User profile, session, and security settings.",
		"settings",
		"/portal/apps/settings",
		"platform",
		"[]",
		nil,
		enabled,
	)
	displayInLauncher := false
	displayInSideNav := false
	system.DisplayInLauncher = &displayInLauncher
	system.DisplayInSideNav = &displayInSideNav
	return system
}

func defaultDataStatisticsSystem(enabled bool) SeedSystem {
	system := defaultPortalSystem(
		authz.SystemDataStatistics,
		"Data & Statistics",
		"Data quality, document upload, reports, and surveillance tools.",
		"home",
		"/portal/apps/dwh",
		"platform",
		dataStatisticsNavigation,
		[]string{
			string(authz.PermissionDataQualityRead),
			string(authz.PermissionDocumentsRead),
			string(authz.PermissionSurveillanceRead),
			string(authz.PermissionReportBrowserRead),
		},
		enabled,
	)
	system.AccessRoles = append(system.AccessRoles,
		authz.ReportBrowserViewer,
		authz.ReportBrowserAnalyst,
		authz.ReportBrowserManager,
		authz.ReportBrowserAdmin,
	)
	system.Roles = append(system.Roles,
		SeedRole{
			Name:        authz.ReportBrowserAdmin,
			DisplayName: "Report Admin",
			Permissions: []string{
				string(authz.PermissionReportBrowserRead),
				string(authz.PermissionMetricsRead),
				string(authz.PermissionAuditRead),
			},
		},
		SeedRole{
			Name:        authz.ReportBrowserManager,
			DisplayName: "Report Manager",
			Permissions: []string{
				string(authz.PermissionReportBrowserRead),
				string(authz.PermissionMetricsRead),
			},
		},
		SeedRole{
			Name:        authz.ReportBrowserAnalyst,
			DisplayName: "Report Analyst",
			Permissions: []string{
				string(authz.PermissionReportBrowserRead),
				string(authz.PermissionDataQualityRead),
			},
		},
		SeedRole{
			Name:        authz.ReportBrowserViewer,
			DisplayName: "Report Viewer",
			Permissions: []string{
				string(authz.PermissionReportBrowserRead),
			},
		},
	)
	return system
}
