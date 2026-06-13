package authz

import "strings"

const (
	RoleAdmin                = "admin"
	RoleUser                 = "user"
	RoleManager              = "manager"
	RoleDocumentsAdmin       = "documents_admin"
	RoleSurveillanceAdmin    = "surveillance_admin"
	RoleSurveillanceImporter = "surveillance_importer"
	RoleEmailAdmin           = "email_admin"
	RoleAuditReader          = "audit_reader"
	RoleMetricsReader        = "metrics_reader"
)

var rolePermissions = map[string][]Permission{
	RoleAdmin: AllPermissions,
	RoleUser: {
		PermissionPortalAccess,
		PermissionSystemsRead,
		PermissionSystemsLaunch,
	},
	RoleManager: {
		PermissionPortalAccess,
		PermissionSystemsRead,
		PermissionSystemsLaunch,
		PermissionUsersRead,
		PermissionClientsRead,
		PermissionAnnouncementsRead,
		PermissionDocumentsRead,
		PermissionDocumentsWrite,
		PermissionDocumentTemplatesRead,
		PermissionSurveillanceRead,
		PermissionStorageLocationsRead,
		PermissionDataQualityRead,
		PermissionDataQualityWrite,
		PermissionNotificationsRead,
	},
	RoleDocumentsAdmin: {
		PermissionDocumentsRead,
		PermissionDocumentsWrite,
		PermissionDocumentsProcess,
		PermissionDocumentTemplatesRead,
		PermissionDocumentTemplatesWrite,
		PermissionDocumentTemplatesPublish,
		PermissionStorageLocationsRead,
		PermissionStorageLocationsWrite,
	},
	RoleSurveillanceAdmin: {
		PermissionSurveillanceRead,
		PermissionSurveillanceImport,
		PermissionSurveillanceManageLocations,
		PermissionSurveillanceManageAlerts,
	},
	RoleSurveillanceImporter: {
		PermissionSurveillanceRead,
		PermissionSurveillanceImport,
	},
	RoleEmailAdmin: {
		PermissionEmailRead,
		PermissionEmailSend,
		PermissionEmailManage,
	},
	RoleAuditReader: {
		PermissionAuditRead,
	},
	RoleMetricsReader: {
		PermissionMetricsRead,
	},
}

var systemRolePermissions = map[string]map[string][]Permission{
	SystemDashboardWeb: {
		DashboardWebAccess: {
			PermissionPortalAccess,
			PermissionSystemsRead,
			PermissionSystemsLaunch,
		},
		DashboardWebIntegratedOutbreakAccess: {
			PermissionOutbreakAccess,
			PermissionSurveillanceRead,
		},
	},
	SystemIntegratedOutbreak: {
		IntegratedOutbreakAccess: {
			PermissionOutbreakAccess,
			PermissionSurveillanceRead,
		},
		IntegratedOutbreakSuperAdmin: {
			PermissionOutbreakAccess,
			PermissionOutbreakManage,
			PermissionSurveillanceRead,
			PermissionSurveillanceImport,
			PermissionSurveillanceManageLocations,
			PermissionSurveillanceManageAlerts,
			PermissionDataQualityRead,
			PermissionDataQualityWrite,
			PermissionDataQualityResolve,
			PermissionDocumentsRead,
			PermissionDocumentsWrite,
			PermissionDocumentsProcess,
		},
		IntegratedOutbreakAdmin: {
			PermissionOutbreakAccess,
			PermissionOutbreakManage,
			PermissionSurveillanceRead,
			PermissionSurveillanceImport,
			PermissionSurveillanceManageLocations,
			PermissionSurveillanceManageAlerts,
			PermissionDataQualityRead,
			PermissionDataQualityWrite,
			PermissionDataQualityResolve,
		},
		IntegratedOutbreakManager: {
			PermissionOutbreakAccess,
			PermissionSurveillanceRead,
			PermissionDataQualityRead,
			PermissionDataQualityWrite,
			PermissionDocumentsRead,
		},
		IntegratedOutbreakViewer: {
			PermissionOutbreakAccess,
			PermissionSurveillanceRead,
			PermissionDataQualityRead,
			PermissionDocumentsRead,
		},
		IntegratedOutbreakDataEntry: {
			PermissionOutbreakAccess,
			PermissionSurveillanceRead,
			PermissionSurveillanceImport,
			PermissionDataQualityRead,
			PermissionDataQualityWrite,
		},
		IntegratedOutbreakLabTechnician: {
			PermissionOutbreakAccess,
			PermissionSurveillanceRead,
			PermissionDataQualityRead,
			PermissionDataQualityWrite,
		},
		IntegratedOutbreakSurveillanceOfficer: {
			PermissionOutbreakAccess,
			PermissionSurveillanceRead,
			PermissionSurveillanceImport,
			PermissionSurveillanceManageAlerts,
		},
	},
	SystemReportBrowser: {
		ReportBrowserAccess: {
			PermissionReportBrowserRead,
		},
	},
}

func PermissionsForContext(realmRoles []string, clientRoles map[string][]string) []Permission {
	seen := map[Permission]bool{}
	permissions := make([]Permission, 0)

	add := func(values []Permission) {
		for _, permission := range values {
			if seen[permission] {
				continue
			}
			seen[permission] = true
			permissions = append(permissions, permission)
		}
	}

	for _, role := range NormalizeRoles(realmRoles) {
		add(rolePermissions[role])
	}

	for system, roles := range normalizeClientRoles(clientRoles) {
		roleMap := systemRolePermissions[system]
		for _, role := range roles {
			add(roleMap[role])
		}
	}

	return permissions
}

func AccessibleSystemsForContext(clientRoles map[string][]string) []string {
	return systemIDs(AccessibleSystemDetailsForContext(clientRoles))
}

func AccessibleSystemDetailsForContext(clientRoles map[string][]string) []SystemAccess {
	seen := map[string]bool{}
	systems := make([]SystemAccess, 0)

	add := func(system SystemAccess) {
		clientID := strings.TrimSpace(system.ClientID)
		if clientID == "" || seen[clientID] {
			return
		}
		seen[clientID] = true
		systems = append(systems, system)
	}

	systemAccess := func(system string, roles []string) SystemAccess {
		system = strings.TrimSpace(system)
		metadata := staticSystemMetadata[system]
		return SystemAccess{
			ClientID:    system,
			DisplayName: metadata.DisplayName,
			LaunchURL:   metadata.LaunchURL,
			Icon:        metadata.Icon,
			Category:    metadata.Category,
			Roles:       roles,
		}
	}

	for system, roles := range normalizeClientRoles(clientRoles) {
		if len(roles) == 0 {
			continue
		}

		switch system {
		case SystemDashboardWeb:
			add(systemAccess(system, roles))
			if hasRole(roles, DashboardWebIntegratedOutbreakAccess) {
				add(systemAccess(SystemIntegratedOutbreak, []string{DashboardWebIntegratedOutbreakAccess}))
			}
		case SystemIntegratedOutbreak, SystemReportBrowser:
			add(systemAccess(system, roles))
		}
	}

	return systems
}

type systemMetadata struct {
	DisplayName string
	LaunchURL   string
	Icon        string
	Category    string
}

var staticSystemMetadata = map[string]systemMetadata{
	SystemDashboardWeb: {
		DisplayName: "Integrated Health Portal",
		LaunchURL:   "/portal",
		Icon:        "dashboard",
		Category:    "platform",
	},
	SystemIntegratedOutbreak: {
		DisplayName: "Integrated Outbreak System",
		LaunchURL:   "/portal/apps/dwh/surveillance",
		Icon:        "outbreak",
		Category:    "surveillance",
	},
	SystemReportBrowser: {
		DisplayName: "Report Browser",
		LaunchURL:   "/portal/apps/dwh/reports",
		Icon:        "reporting",
		Category:    "reports",
	},
}

func PermissionsForRoles(roles []string) []Permission {
	seen := map[Permission]bool{}
	permissions := make([]Permission, 0)

	for _, role := range NormalizeRoles(roles) {
		for _, permission := range rolePermissions[role] {
			if seen[permission] {
				continue
			}
			seen[permission] = true
			permissions = append(permissions, permission)
		}
	}

	return permissions
}

func NormalizeRoles(roles []string) []string {
	seen := map[string]bool{}
	normalized := make([]string, 0, len(roles))

	for _, role := range roles {
		role = strings.ToLower(strings.TrimSpace(role))
		if role == "" || seen[role] {
			continue
		}
		seen[role] = true
		normalized = append(normalized, role)
	}

	return normalized
}

func hasRole(roles []string, role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return false
	}

	for _, current := range roles {
		if strings.EqualFold(strings.TrimSpace(current), role) {
			return true
		}
	}

	return false
}
