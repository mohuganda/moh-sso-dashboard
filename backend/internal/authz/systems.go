package authz

const (
	SystemDashboardWeb       = "dashboard-web"
	SystemIntegratedOutbreak = "integrated-outbreak-system"
	SystemReportBrowser      = "report-browser"
	SystemDataStatistics     = "data-statistics"
	SystemUtilities          = "utilities"
	SystemSettings           = "settings"
)

const (
	DashboardWebAccess  = "dashboard-web_access"
	DashboardWebAdmin   = "portal_admin"
	DashboardWebManager = "portal_manager"
	DashboardWebUser    = "portal_user"

	IntegratedOutbreakAccess              = "integrated-outbreak-system_access"
	IntegratedOutbreakSuperAdmin          = "super_admin"
	IntegratedOutbreakAdmin               = "admin"
	IntegratedOutbreakManager             = "manager"
	IntegratedOutbreakViewer              = "viewer"
	IntegratedOutbreakDataEntry           = "data_entry"
	IntegratedOutbreakLabTechnician       = "lab_technician"
	IntegratedOutbreakSurveillanceOfficer = "surveillance_officer"

	ReportBrowserAccess  = "report-browser_access"
	ReportBrowserAdmin   = "report_admin"
	ReportBrowserManager = "report_manager"
	ReportBrowserAnalyst = "report_analyst"
	ReportBrowserViewer  = "report_viewer"

	DataStatisticsAccess = "data-statistics_access"
	UtilitiesAccess      = "utilities_access"
	SettingsAccess       = "settings_access"
)
