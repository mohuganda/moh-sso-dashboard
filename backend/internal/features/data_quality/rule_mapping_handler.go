package data_quality

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/http/response"
)

type ruleMappingPreviewRequest struct {
	TableID string `json:"table_id" binding:"required"`
}

type ruleMappingApplyRequest struct {
	TableID string   `json:"table_id" binding:"required"`
	Codes   []string `json:"codes" binding:"required"`
}

func (h *Handler) PreviewRuleMapping(c *gin.Context) {
	if h.dqaStore == nil || h.dwhDB == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req ruleMappingPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	candidates, err := h.dqaStore.PreviewRuleMapping(c.Request.Context(), h.dwhDB, strings.TrimSpace(req.TableID))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "RULE_MAPPING_PREVIEW_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, candidates)
}

func (h *Handler) ApplyRuleMapping(c *gin.Context) {
	if h.dqaStore == nil {
		response.Fail(c, http.StatusServiceUnavailable, "DWH_UNAVAILABLE", "DWH connection is not configured")
		return
	}
	var req ruleMappingApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
		return
	}
	result, err := h.dqaStore.ApplyRuleMapping(c.Request.Context(), strings.TrimSpace(req.TableID),
		req.Codes, getUserEmailOrUsername(c))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "RULE_MAPPING_APPLY_FAILED", err.Error())
		return
	}
	response.OK(c, http.StatusOK, result)
}
