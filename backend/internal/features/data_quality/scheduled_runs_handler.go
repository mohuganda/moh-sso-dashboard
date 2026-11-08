package data_quality

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
)

type scheduleRunRequest struct {
	TableID     string    `json:"table_id" binding:"required"`
	Scope       RunScope  `json:"scope"`
	ScheduledAt time.Time `json:"scheduled_at" binding:"required"`
}

func (h *Handler) ScheduleDQARun(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req scheduleRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	if req.ScheduledAt.Before(time.Now()) {
		response.Fail(c, http.StatusBadRequest, "INVALID_SCHEDULE", "scheduled_at is in the past")
		return
	}
	scheduled, err := h.dqaStore.ScheduleRun(c.Request.Context(), strings.TrimSpace(req.TableID),
		req.Scope, req.ScheduledAt, getUserEmailOrUsername(c))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "DQA_SCHEDULE_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusCreated, scheduled)
}

func (h *Handler) ListDQAScheduledRuns(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	status := c.Query("status")
	switch status {
	case "", "pending", "running", "done", "failed", "cancelled":
	default:
		response.Fail(c, http.StatusBadRequest, "INVALID_STATUS", "unknown status")
		return
	}
	runs, err := h.dqaStore.ListScheduledRuns(c.Request.Context(), status,
		parseListLimit(c, 50, 200), parseListOffset(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_SCHEDULES_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, runs)
}

func (h *Handler) CancelDQAScheduledRun(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	id, err := strconv.ParseInt(c.Param("scheduleId"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_SCHEDULE_ID", "scheduleId must be an integer")
		return
	}
	scheduled, err := h.dqaStore.CancelScheduledRun(c.Request.Context(), id)
	if errors.Is(err, errScheduledRunNotFound) {
		response.Fail(c, http.StatusNotFound, "SCHEDULE_NOT_FOUND", "no pending scheduled run with that id")
		return
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "DQA_SCHEDULE_CANCEL_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, scheduled)
}
