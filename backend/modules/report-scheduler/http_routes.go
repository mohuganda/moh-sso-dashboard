package reportscheduler

type endpoint struct {
	method, path string
	public       bool
	required     [][]string
	serve        func(*httpRequest)
}

func (h *httpHandler) endpoints() []endpoint {
	return []endpoint{
		{"GET", "", false, [][]string{{PermissionRead}}, h.GetModule},
		{"GET", "/overview", false, [][]string{{PermissionHistory}}, h.Overview},
		{"GET", "/metrics", false, [][]string{{PermissionMetrics}, {PermissionManage}}, h.Overview},
		{"GET", "/reports", false, [][]string{{PermissionRead}}, h.ListReports},
		{"GET", "/report", false, [][]string{{PermissionRead}}, h.GetReport},
		{"GET", "/report-parameters", false, [][]string{{PermissionRead}}, h.GetReportParameters},
		{"GET", "/reports/{reportId}", false, [][]string{{PermissionRead}}, h.GetReport},
		{"GET", "/reports/{reportId}/parameters", false, [][]string{{PermissionRead}}, h.GetReportParameters},
		{"GET", "/schedules", false, [][]string{{PermissionRead}}, h.ListSchedules},
		{"POST", "/schedules", false, [][]string{{PermissionCreate}}, h.CreateSchedule},
		{"GET", "/schedules/{scheduleId}", false, [][]string{{PermissionRead}}, h.GetSchedule},
		{"PUT", "/schedules/{scheduleId}", false, [][]string{{PermissionUpdate}}, h.UpdateSchedule},
		{"DELETE", "/schedules/{scheduleId}", false, [][]string{{PermissionDelete}}, h.DeleteSchedule},
		{"POST", "/schedules/{scheduleId}/pause", false, [][]string{{PermissionUpdate}}, h.PauseSchedule},
		{"POST", "/schedules/{scheduleId}/resume", false, [][]string{{PermissionUpdate}}, h.ResumeSchedule},
		{"POST", "/schedules/{scheduleId}/duplicate", false, [][]string{{PermissionCreate}}, h.DuplicateSchedule},
		{"POST", "/schedules/{scheduleId}/run", false, [][]string{{PermissionExecute}}, h.RunNow},
		{"GET", "/executions", false, [][]string{{PermissionHistory}}, h.ListExecutions},
		{"GET", "/executions/{executionId}", false, [][]string{{PermissionHistory}}, h.GetExecution},
		{"POST", "/executions/{executionId}/retry", false, [][]string{{PermissionExecute}}, h.RetryExecution},
		{"GET", "/recipients/options", false, [][]string{{PermissionCreate, PermissionUpdate}}, h.RecipientOptions},
		{"POST", "/recipients/preview", false, [][]string{{PermissionCreate, PermissionUpdate}}, h.PreviewRecipients},
		{"GET", "/portal-reports", false, [][]string{{PermissionRead}}, h.ListPortalReports},
		{"GET", "/artifacts/{artifactId}/download", false, [][]string{{PermissionRead}}, h.DownloadArtifact},
		{"GET", "/public/deliveries/{token}/download", true, nil, h.DownloadDeliveryLink},
	}
}
