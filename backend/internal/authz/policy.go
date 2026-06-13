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
	seen := map[string]bool{}
	systems := make([]string, 0)

	add := func(system string) {
		system = strings.TrimSpace(system)
		if system == "" || seen[system] {
			return
		}
		seen[system] = true
		systems = append(systems, system)
	}

	for system, roles := range normalizeClientRoles(clientRoles) {
		if len(roles) == 0 {
			continue
		}

		switch system {
		case SystemDashboardWeb:
			add(system)
			if hasRole(roles, DashboardWebIntegratedOutbreakAccess) {
				add(SystemIntegratedOutbreak)
			}
		case SystemIntegratedOutbreak, SystemReportBrowser:
			add(system)
		}
	}

	return systems
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
