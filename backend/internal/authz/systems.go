package authz

const (
	SystemDashboardWeb       = "dashboard-web"
	SystemOutbreakManagement = "outbreak-management"
	SystemDataStatistics     = "data-statistics"
	SystemUtilities          = "utilities"
	SystemSettings           = "settings"
)

const (
	DashboardWebAccess  = "dashboard-web_access"
	DashboardWebAdmin   = "portal_admin"
	DashboardWebManager = "portal_manager"
	DashboardWebUser    = "portal_user"

	OutbreakManagementAccess              = "outbreak-management_access"
	OutbreakManagementSuperAdmin          = "super_admin"
	OutbreakManagementAdmin               = "admin"
	OutbreakManagementManager             = "manager"
	OutbreakManagementViewer              = "viewer"
	OutbreakManagementDataEntry           = "data_entry"
	OutbreakManagementLabTechnician       = "lab_technician"
	OutbreakManagementSurveillanceOfficer = "surveillance_officer"

	ReportBrowserAdmin   = "report_admin"
	ReportBrowserManager = "report_manager"
	ReportBrowserAnalyst = "report_analyst"
	ReportBrowserViewer  = "report_viewer"

	DataStatisticsAccess = "data-statistics_access"
	UtilitiesAccess      = "utilities_access"
	SettingsAccess       = "settings_access"
)
