package authz

// DefaultUserClientRoles is the standard user's client-role baseline. Explicit
// grants and administrator/manager mappings are managed separately.
func DefaultUserClientRoles() map[string][]string {
	return map[string][]string{
		SystemDashboardWeb: {DashboardWebAccess, DashboardWebUser},
		SystemDataStatistics: {
			ReportSchedulerAccess, ReportBrowserViewer, DataStatisticsAccess,
			IssueTrackerViewer, DocumentViewer,
		},
		SystemSettings: {SettingsAccess},
	}
}
