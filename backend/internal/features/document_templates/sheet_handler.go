package document_templates

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
)

/* =========================================================
 * Handler
 * ========================================================= */
type SheetHandler struct {
	service SheetService
}

/* =========================================================
 * Constructor
 * ========================================================= */
func NewSheetHandler(
	service SheetService,
) *SheetHandler {

	return &SheetHandler{
		service: service,
	}
}

func (h *SheetHandler) CreateSheet(c *gin.Context) {
	var req model.CreateSheetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid request payload")
		return
	}

	sheet, err := h.service.CreateSheet(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "CREATE_FAILED", "request failed")
		return
	}

	response.OK(c, http.StatusCreated, toSheetResponse(sheet))
}

func (h *SheetHandler) GetSheet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "invalid sheet id")
		return
	}

	sheet, err := h.service.GetSheet(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "sheet not found")
		return
	}

	response.OK(c, http.StatusOK, toSheetResponse(sheet))
}

func (h *SheetHandler) GetSheetByCode(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "invalid template id")
		return
	}

	code := c.Param("code")
	if code == "" {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "sheet code required")
		return
	}

	sheet, err := h.service.GetSheetByCode(c.Request.Context(), templateID, code)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "NOT_FOUND", "sheet not found")
		return
	}

	response.OK(c, http.StatusOK, toSheetResponse(sheet))
}

func (h *SheetHandler) ListSheets(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "invalid template id")
		return
	}

	sheets, err := h.service.ListSheets(c.Request.Context(), templateID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toSheetResponses(sheets))
}

func (h *SheetHandler) ListRequiredSheets(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "invalid template id")
		return
	}

	sheets, err := h.service.ListRequiredSheets(c.Request.Context(), templateID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toSheetResponses(sheets))
}

func (h *SheetHandler) UpdateSheet(c *gin.Context) {
	var req model.UpdateSheetRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid request payload")
		return
	}

	sheet, err := h.service.UpdateSheet(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "UPDATE_FAILED", "request failed")
		return
	}

	response.OK(c, http.StatusOK, toSheetResponse(sheet))
}

func (h *SheetHandler) DeleteSheet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "invalid sheet id")
		return
	}

	if err := h.service.DeleteSheet(c.Request.Context(), id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "DELETE_FAILED", "request failed")
		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================

 * ARCHIVE SHEET

 * ========================================================= */

func (h *SheetHandler) ArchiveSheet(c *gin.Context) {

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

func (h *SheetHandler) GetSheetRuntime(c *gin.Context) {

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
	response.OK(c, http.StatusOK, toSheetRuntimeResponse(sheet))

}

func (h *SheetHandler) ValidateSheetExists(c *gin.Context) {

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
	response.OK(c, http.StatusOK, ExistsResponse{Exists: true})

}
