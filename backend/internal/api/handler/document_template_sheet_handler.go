package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
)

var (
	ErrSheetNotFound = errors.New("sheet not found in template")
)

/* =========================================================
 * Handler
 * ========================================================= */
type DocumentTemplateSheetHandler struct {
	service service.DocumentTemplateSheetService
}

/* =========================================================
 * Constructor
 * ========================================================= */
func NewDocumentTemplateSheetHandler(
	service service.DocumentTemplateSheetService,
) *DocumentTemplateSheetHandler {

	return &DocumentTemplateSheetHandler{
		service: service,
	}
}

func (h *DocumentTemplateSheetHandler) CreateSheet(c *gin.Context) {
	var req model.CreateSheetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
		})
		return
	}

	sheet, err := h.service.CreateSheet(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, sheet)
}

func (h *DocumentTemplateSheetHandler) GetSheet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid sheet id",
		})
		return
	}

	sheet, err := h.service.GetSheet(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": ErrSheetNotFound.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sheet)
}

func (h *DocumentTemplateSheetHandler) GetSheetByCode(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid template id",
		})
		return
	}

	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "sheet code required",
		})
		return
	}

	sheet, err := h.service.GetSheetByCode(c.Request.Context(), templateID, code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": ErrSheetNotFound.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sheet)
}

func (h *DocumentTemplateSheetHandler) ListSheets(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid template id",
		})
		return
	}

	sheets, err := h.service.ListSheets(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sheets)
}

func (h *DocumentTemplateSheetHandler) ListRequiredSheets(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid template id",
		})
		return
	}

	sheets, err := h.service.ListRequiredSheets(c.Request.Context(), templateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sheets)
}

func (h *DocumentTemplateSheetHandler) UpdateSheet(c *gin.Context) {
	var req model.UpdateSheetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request payload",
		})
		return
	}

	sheet, err := h.service.UpdateSheet(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, sheet)
}

func (h *DocumentTemplateSheetHandler) DeleteSheet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid sheet id",
		})
		return
	}

	if err := h.service.DeleteSheet(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================

 * ARCHIVE SHEET

 * ========================================================= */

func (h *DocumentTemplateSheetHandler) ArchiveSheet(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid sheet ID")
		return
	}
	if err := h.service.ArchiveSheet(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "ARCHIVE_FAILED", "Failed to archive sheet")
		return
	}
	c.Status(http.StatusNoContent)

}

/* =========================================================

 * RUNTIME HELPERS

 * ========================================================= */

func (h *DocumentTemplateSheetHandler) GetSheetRuntime(c *gin.Context) {

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid template ID")
		return
	}
	sheetCode := c.Param("code")
	tpl := &model.TemplateRuntime{
		ID: templateID,
	}
	sheet, err := h.service.GetSheetRuntime(c.Request.Context(), tpl, sheetCode)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "SHEET_NOT_FOUND", "Sheet not found in template")
		return
	}
	response.OK(c, http.StatusOK, sheet)

}

func (h *DocumentTemplateSheetHandler) ValidateSheetExists(c *gin.Context) {

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid template ID")
		return
	}
	sheetCode := c.Param("code")
	tpl := &model.TemplateRuntime{
		ID: templateID,
	}
	err = h.service.ValidateSheetExists(c.Request.Context(), tpl, sheetCode)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "SHEET_NOT_FOUND", "Sheet does not exist")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"exists": true})

}
