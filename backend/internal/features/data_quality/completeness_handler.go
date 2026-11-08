package data_quality

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
)

type completenessRunRequest struct {
	Period string `json:"period"`
}

func (h *Handler) RunCompleteness(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req completenessRunRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	period, err := parseCompletenessPeriod(req.Period, time.Now())
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_PERIOD", err.Error())
		return
	}
	run, err := h.dqaStore.RunCompleteness(c.Request.Context(), period, getUserEmailOrUsername(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "COMPLETENESS_RUN_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusCreated, run)
}

func (h *Handler) ListCompletenessRuns(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	runs, err := h.dqaStore.ListCompletenessRuns(c.Request.Context(), parseListLimit(c, 25, 200), parseListOffset(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "COMPLETENESS_RUNS_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, runs)
}

func (h *Handler) GetCompletenessRun(c *gin.Context) {
	runID, ok := h.completenessRunID(c)
	if !ok {
		return
	}
	run, err := h.dqaStore.GetCompletenessRun(c.Request.Context(), runID)
	if errors.Is(err, errCompletenessRunNotFound) {
		response.Fail(c, http.StatusNotFound, "COMPLETENESS_RUN_NOT_FOUND", err.Error())
		return
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "COMPLETENESS_RUN_GET_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, run)
}

func (h *Handler) ListCompletenessDistricts(c *gin.Context) {
	runID, ok := h.completenessRunID(c)
	if !ok {
		return
	}
	districts, err := h.dqaStore.ListCompletenessDistricts(c.Request.Context(), runID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "COMPLETENESS_DISTRICTS_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, districts)
}

func (h *Handler) ListCompletenessVHTs(c *gin.Context) {
	runID, ok := h.completenessRunID(c)
	if !ok {
		return
	}
	status := c.Query("status")
	if status != "" && status != "missing" && status != "reported" {
		response.Fail(c, http.StatusBadRequest, "INVALID_STATUS", "status must be missing or reported")
		return
	}
	filter := CompletenessVHTFilter{District: c.Query("district"), Status: status}
	vhts, err := h.dqaStore.ListCompletenessVHTs(c.Request.Context(), runID, filter,
		parseListLimit(c, 100, 1000), parseListOffset(c))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "COMPLETENESS_VHTS_LIST_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, vhts)
}

func (h *Handler) completenessRunID(c *gin.Context) (int64, bool) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return 0, false
	}
	runID, err := strconv.ParseInt(c.Param("runId"), 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_RUN_ID", "runId must be an integer")
		return 0, false
	}
	return runID, true
}
