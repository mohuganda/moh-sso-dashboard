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
		PermissionAnnouncementsRead,
		PermissionClientsRead,
		PermissionDocumentsRead,
		PermissionDocumentTemplatesRead,
		PermissionSurveillanceRead,
		PermissionStorageLocationsRead,
		PermissionDataQualityRead,
	},
	RoleManager: {
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
