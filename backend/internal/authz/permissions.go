package authz

type Permission string

const (
	PermissionUsersRead       Permission = "users:read"
	PermissionUsersWrite      Permission = "users:write"
	PermissionUsersRolesWrite Permission = "users:roles:write"

	PermissionClientsRead       Permission = "clients:read"
	PermissionClientsWrite      Permission = "clients:write"
	PermissionClientsRolesWrite Permission = "clients:roles:write"

	PermissionAnnouncementsRead    Permission = "announcements:read"
	PermissionAnnouncementsWrite   Permission = "announcements:write"
	PermissionAnnouncementsPublish Permission = "announcements:publish"

	PermissionDocumentsRead    Permission = "documents:read"
	PermissionDocumentsWrite   Permission = "documents:write"
	PermissionDocumentsProcess Permission = "documents:process"

	PermissionDocumentTemplatesRead    Permission = "document_templates:read"
	PermissionDocumentTemplatesWrite   Permission = "document_templates:write"
	PermissionDocumentTemplatesPublish Permission = "document_templates:publish"

	PermissionSurveillanceRead            Permission = "surveillance:read"
	PermissionSurveillanceImport          Permission = "surveillance:import"
	PermissionSurveillanceManageLocations Permission = "surveillance:manage_locations"
	PermissionSurveillanceManageAlerts    Permission = "surveillance:manage_alerts"

	PermissionEmailRead   Permission = "email:read"
	PermissionEmailSend   Permission = "email:send"
	PermissionEmailManage Permission = "email:manage"

	PermissionStorageLocationsRead  Permission = "storage_locations:read"
	PermissionStorageLocationsWrite Permission = "storage_locations:write"

	PermissionDataQualityRead    Permission = "data_quality:read"
	PermissionDataQualityWrite   Permission = "data_quality:write"
	PermissionDataQualityResolve Permission = "data_quality:resolve"

	PermissionMetricsRead        Permission = "metrics:read"
	PermissionAuditRead          Permission = "audit:read"
	PermissionNotificationsRead  Permission = "notifications:read"
	PermissionNotificationsWrite Permission = "notifications:write"
)

var AllPermissions = []Permission{
	PermissionUsersRead,
	PermissionUsersWrite,
	PermissionUsersRolesWrite,
	PermissionClientsRead,
	PermissionClientsWrite,
	PermissionClientsRolesWrite,
	PermissionAnnouncementsRead,
	PermissionAnnouncementsWrite,
	PermissionAnnouncementsPublish,
	PermissionDocumentsRead,
	PermissionDocumentsWrite,
	PermissionDocumentsProcess,
	PermissionDocumentTemplatesRead,
	PermissionDocumentTemplatesWrite,
	PermissionDocumentTemplatesPublish,
	PermissionSurveillanceRead,
	PermissionSurveillanceImport,
	PermissionSurveillanceManageLocations,
	PermissionSurveillanceManageAlerts,
	PermissionEmailRead,
	PermissionEmailSend,
	PermissionEmailManage,
	PermissionStorageLocationsRead,
	PermissionStorageLocationsWrite,
	PermissionDataQualityRead,
	PermissionDataQualityWrite,
	PermissionDataQualityResolve,
	PermissionMetricsRead,
	PermissionAuditRead,
	PermissionNotificationsRead,
	PermissionNotificationsWrite,
}
