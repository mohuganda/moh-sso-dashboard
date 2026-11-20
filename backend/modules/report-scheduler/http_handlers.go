package reportscheduler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func (h *httpHandler) GetModule(c *httpRequest) {
	module := h.service.Module()
	module.HealthContext = c.actor.HealthContext
	c.ok(http.StatusOK, module)
}

func (h *httpHandler) ListReports(c *httpRequest) {
	items, err := h.service.ListReports(c.request.Context())
	if err != nil {
		healthBIFailure(c, err)
		return
	}
	c.ok(http.StatusOK, items)
}
func (h *httpHandler) GetReport(c *httpRequest) {
	id, ok := reportIDFromRequest(c)
	if !ok {
		return
	}
	item, err := h.service.GetReport(c.request.Context(), id)
	if err != nil {
		healthBIFailure(c, err)
		return
	}
	c.ok(http.StatusOK, item)
}
func (h *httpHandler) GetReportParameters(c *httpRequest) {
	id, ok := reportIDFromRequest(c)
	if !ok {
		return
	}
	items, err := h.service.GetParameters(c.request.Context(), id)
	if err != nil {
		healthBIFailure(c, err)
		return
	}
	c.ok(http.StatusOK, items)
}
func reportIDFromRequest(c *httpRequest) (string, bool) {
	id := c.request.PathValue("reportId")
	if id == "" {
		id = c.request.URL.Query().Get("reportId")
	}
	if strings.TrimSpace(id) == "" {
		c.fail(http.StatusBadRequest, "INVALID_REPORT_ID", "reportId is required")
		return "", false
	}
	return id, true
}
func (h *httpHandler) CreateSchedule(c *httpRequest) {
	var req CreateScheduleRequest
	if err := decodeSchedule(c, &req); err != nil {
		c.fail(http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.CreateSchedule(c.request.Context(), req, c.actor.UserID, c.actor.HealthContext)
	if err != nil {
		c.fail(http.StatusBadRequest, "SCHEDULE_CREATE_FAILED", err.Error())
		return
	}
	c.ok(http.StatusCreated, item)
}
func (h *httpHandler) ListSchedules(c *httpRequest) {
	options, ok := parseListOptions(c, true)
	if !ok {
		return
	}
	items, err := h.service.ListSchedules(c.request.Context(), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext, options)
	if err != nil {
		c.fail(http.StatusInternalServerError, "SCHEDULE_LIST_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, items)
}
func (h *httpHandler) GetSchedule(c *httpRequest) {
	item, err := h.service.GetSchedule(c.request.Context(), c.request.PathValue("scheduleId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "report schedule not found")
		return
	}
	if err != nil {
		c.fail(http.StatusInternalServerError, "SCHEDULE_GET_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, item)
}
func (h *httpHandler) UpdateSchedule(c *httpRequest) {
	var req UpdateScheduleRequest
	if err := decodeSchedule(c, &req); err != nil {
		c.fail(http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	item, err := h.service.UpdateSchedule(c.request.Context(), c.request.PathValue("scheduleId"), req, c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "report schedule not found")
		return
	}
	if err != nil {
		c.fail(http.StatusBadRequest, "SCHEDULE_UPDATE_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, item)
}
func (h *httpHandler) DeleteSchedule(c *httpRequest) {
	err := h.service.DeleteSchedule(c.request.Context(), c.request.PathValue("scheduleId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "report schedule not found")
		return
	}
	if err != nil {
		c.fail(http.StatusInternalServerError, "SCHEDULE_DELETE_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, map[string]any{"deleted": true})
}
func (h *httpHandler) PauseSchedule(c *httpRequest) {
	item, err := h.service.PauseSchedule(c.request.Context(), c.request.PathValue("scheduleId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "report schedule not found")
		return
	}
	if err != nil {
		c.fail(http.StatusInternalServerError, "SCHEDULE_PAUSE_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, item)
}
func (h *httpHandler) ResumeSchedule(c *httpRequest) {
	item, err := h.service.ResumeSchedule(c.request.Context(), c.request.PathValue("scheduleId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "report schedule not found")
		return
	}
	if err != nil {
		c.fail(http.StatusInternalServerError, "SCHEDULE_RESUME_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, item)
}
func (h *httpHandler) DuplicateSchedule(c *httpRequest) {
	item, err := h.service.DuplicateSchedule(c.request.Context(), c.request.PathValue("scheduleId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "report schedule not found")
		return
	}
	if err != nil {
		c.fail(http.StatusInternalServerError, "SCHEDULE_DUPLICATE_FAILED", err.Error())
		return
	}
	c.ok(http.StatusCreated, item)
}
func (h *httpHandler) RunNow(c *httpRequest) {
	execution, err := h.service.RunNow(c.request.Context(), c.request.PathValue("scheduleId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "SCHEDULE_NOT_FOUND", "report schedule not found")
		return
	}
	if err != nil {
		c.fail(http.StatusBadGateway, "SCHEDULE_RUN_FAILED", err.Error())
		return
	}
	c.ok(http.StatusAccepted, execution)
}
func (h *httpHandler) ListExecutions(c *httpRequest) {
	options, ok := parseListOptions(c, false)
	if !ok {
		return
	}
	items, err := h.service.ListExecutions(c.request.Context(), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext, options)
	if err != nil {
		c.fail(http.StatusInternalServerError, "EXECUTION_LIST_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, items)
}

func (h *httpHandler) Overview(c *httpRequest) {
	item, err := h.service.Overview(c.request.Context(), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if err != nil {
		c.fail(http.StatusInternalServerError, "SCHEDULER_OVERVIEW_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, item)
}

func (h *httpHandler) GetExecution(c *httpRequest) {
	item, err := h.service.GetExecutionDetail(c.request.Context(), c.request.PathValue("executionId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "EXECUTION_NOT_FOUND", "report execution not found")
		return
	}
	if err != nil {
		c.fail(http.StatusInternalServerError, "EXECUTION_GET_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, item)
}

func (h *httpHandler) RetryExecution(c *httpRequest) {
	item, err := h.service.RetryExecution(c.request.Context(), c.request.PathValue("executionId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "EXECUTION_NOT_FOUND", "report execution not found")
		return
	}
	if err != nil {
		c.fail(http.StatusBadGateway, "EXECUTION_RETRY_FAILED", err.Error())
		return
	}
	c.ok(http.StatusAccepted, item)
}

func (h *httpHandler) PreviewRecipients(c *httpRequest) {
	var req struct {
		Recipients []ScheduleRecipient `json:"recipients"`
	}
	if err := c.decode(&req); err != nil {
		c.fail(http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if req.Recipients == nil {
		c.fail(http.StatusBadRequest, "INVALID_BODY", "recipients is required")
		return
	}
	preview, err := h.service.PreviewRecipients(c.request.Context(), req.Recipients)
	if err != nil {
		c.fail(http.StatusBadRequest, "RECIPIENT_RESOLUTION_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, preview)
}

func (h *httpHandler) ListPortalReports(c *httpRequest) {
	limit, ok := parseLimit(c, 100)
	if !ok {
		return
	}
	items, err := h.service.ListPortalReports(c.request.Context(), c.actor.UserID, limit)
	if err != nil {
		c.fail(http.StatusInternalServerError, "PORTAL_REPORTS_LIST_FAILED", err.Error())
		return
	}
	c.ok(http.StatusOK, items)
}

func parseListOptions(c *httpRequest, allowEnabled bool) (ListOptions, bool) {
	limit, ok := parseLimit(c, 100)
	if !ok {
		return ListOptions{}, false
	}
	options := ListOptions{Limit: limit, Status: strings.TrimSpace(c.request.URL.Query().Get("status")), Search: strings.TrimSpace(c.request.URL.Query().Get("search"))}
	if len(options.Search) > 200 {
		c.fail(http.StatusBadRequest, "VALIDATION_FAILED", "search must be 200 characters or fewer")
		return ListOptions{}, false
	}
	if allowEnabled && strings.TrimSpace(c.request.URL.Query().Get("enabled")) != "" {
		value, err := strconv.ParseBool(c.request.URL.Query().Get("enabled"))
		if err != nil {
			c.fail(http.StatusBadRequest, "VALIDATION_FAILED", "enabled must be true or false")
			return ListOptions{}, false
		}
		options.Enabled = &value
	}
	if !allowEnabled && options.Status != "" {
		valid := map[string]bool{"queued": true, "generating": true, "polling": true, "generated": true, "delivering": true, "completed": true, "failed": true, "cancelled": true, "retrying": true}
		if !valid[options.Status] {
			c.fail(http.StatusBadRequest, "VALIDATION_FAILED", "invalid execution status filter")
			return ListOptions{}, false
		}
	}
	return options, true
}

func parseLimit(c *httpRequest, defaultLimit int) (int, bool) {
	limit := defaultLimit
	if value := strings.TrimSpace(c.request.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 200 {
			c.fail(http.StatusBadRequest, "VALIDATION_FAILED", "limit must be between 1 and 200")
			return 0, false
		}
		limit = parsed
	}
	return limit, true
}

func (h *httpHandler) DownloadArtifact(c *httpRequest) {
	_, target, err := h.service.ArtifactDownloadURL(c.request.Context(), c.request.PathValue("artifactId"), c.actor.UserID, c.actor.HasPermission(PermissionManage), c.actor.HealthContext)
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "ARTIFACT_NOT_FOUND", "report artifact not found")
		return
	}
	if err != nil {
		c.fail(http.StatusBadGateway, "ARTIFACT_DOWNLOAD_FAILED", err.Error())
		return
	}
	c.writer.Header().Set("Cache-Control", "no-store")
	c.writer.Header().Set("Referrer-Policy", "no-referrer")
	http.Redirect(c.writer, c.request, target, http.StatusFound)
}

func (h *httpHandler) DownloadDeliveryLink(c *httpRequest) {
	_, target, err := h.service.PublicDeliveryDownloadURL(c.request.Context(), c.request.PathValue("token"))
	if errors.Is(err, ErrScheduleNotFound) {
		c.fail(http.StatusNotFound, "DELIVERY_LINK_INVALID", "report delivery link is invalid or expired")
		return
	}
	if err != nil {
		c.fail(http.StatusBadGateway, "DELIVERY_DOWNLOAD_FAILED", "report download is currently unavailable")
		return
	}
	c.writer.Header().Set("Cache-Control", "no-store")
	c.writer.Header().Set("Referrer-Policy", "no-referrer")
	c.writer.Header().Set("X-Robots-Tag", "noindex, nofollow")
	http.Redirect(c.writer, c.request, target, http.StatusFound)
}

func healthBIFailure(c *httpRequest, err error) {
	code := "HEALTH_BI_REQUEST_FAILED"
	status := http.StatusBadGateway
	if strings.Contains(strings.ToLower(err.Error()), "not configured") {
		code = "HEALTH_BI_UNAVAILABLE"
		status = http.StatusServiceUnavailable
	}
	c.fail(status, code, err.Error())
}

func decodeSchedule(c *httpRequest, req *CreateScheduleRequest) error {
	if err := c.decode(req); err != nil {
		return err
	}
	for _, field := range []struct{ name, value string }{
		{"healthBiReportId", req.HealthBIReportID},
		{"reportName", req.ReportName},
		{"frequency", req.Frequency},
		{"timezone", req.Timezone},
		{"periodStrategy", req.PeriodStrategy},
		{"outputFormat", req.OutputFormat},
		{"outputConfig.format", req.OutputConfig.Format},
	} {
		if field.value == "" {
			return errors.New(field.name + " is required")
		}
	}
	return nil
}
