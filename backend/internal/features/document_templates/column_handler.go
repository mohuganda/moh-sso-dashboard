package document_templates

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
)

type ColumnHandler struct {
	service ColumnService
}

func NewColumnHandler(
	columnService ColumnService,
) *ColumnHandler {

	return &ColumnHandler{
		service: columnService,
	}
}

/* =========================================================
 * CREATE COLUMN
 * ========================================================= */
func (h *ColumnHandler) CreateColumn(c *gin.Context) {

	var req model.CreateColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)
		return
	}

	col, err := h.service.CreateColumn(c.Request.Context(), req)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"CREATE_FAILED",
			"request failed",
		)
		return
	}

	response.OK(c, http.StatusCreated, col)
}

/* =========================================================
 * GET COLUMN
 * ========================================================= */
func (h *ColumnHandler) GetColumn(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid column ID")
		return
	}

	col, err := h.service.GetColumn(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, http.StatusNotFound, "COLUMN_NOT_FOUND", "Column not found")
		return
	}

	response.OK(c, http.StatusOK, col)
}

/* =========================================================
 * LIST COLUMNS
 * ========================================================= */
func (h *ColumnHandler) ListColumns(c *gin.Context) {

	sheetID, err := uuid.Parse(c.Param("sheetId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid sheet ID")
		return
	}

	cols, err := h.service.ListColumns(c.Request.Context(), sheetID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list columns")
		return
	}

	response.OK(c, http.StatusOK, cols)
}

/* =========================================================
 * LIST REQUIRED COLUMNS
 * ========================================================= */
func (h *ColumnHandler) ListRequiredColumns(c *gin.Context) {

	sheetID, err := uuid.Parse(c.Param("sheetId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid sheet ID")
		return
	}

	cols, err := h.service.ListRequiredColumns(c.Request.Context(), sheetID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list required columns")
		return
	}

	response.OK(c, http.StatusOK, cols)
}

/* =========================================================
 * LIST UNIQUE COLUMNS
 * ========================================================= */
func (h *ColumnHandler) ListUniqueColumns(c *gin.Context) {

	sheetID, err := uuid.Parse(c.Param("sheetId"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid sheet ID")
		return
	}

	cols, err := h.service.ListUniqueColumns(c.Request.Context(), sheetID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list unique columns")
		return
	}

	response.OK(c, http.StatusOK, cols)
}

/* =========================================================
 * UPDATE COLUMN
 * ========================================================= */
func (h *ColumnHandler) UpdateColumn(c *gin.Context) {

	var req model.UpdateColumnRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid column ID")
		return
	}

	req.ID = id

	col, err := h.service.UpdateColumn(c.Request.Context(), req)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"UPDATE_FAILED",
			"request failed",
		)
		return
	}

	response.OK(c, http.StatusOK, col)
}

/* =========================================================
 * ARCHIVE COLUMN
 * ========================================================= */
func (h *ColumnHandler) ArchiveColumn(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid column ID")
		return
	}

	if err := h.service.ArchiveColumn(c.Request.Context(), id); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"ARCHIVE_FAILED",
			"Failed to archive column",
		)
		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * DELETE COLUMN
 * ========================================================= */
func (h *ColumnHandler) DeleteColumn(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_UUID", "Invalid column ID")
		return
	}

	if err := h.service.DeleteColumn(c.Request.Context(), id); err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"DELETE_FAILED",
			"Failed to delete column",
		)
		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * PROCESS VALUE (runtime helper endpoint)
 * ========================================================= */
func (h *ColumnHandler) ProcessValue(c *gin.Context) {

	var body processValueRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)
		return
	}

	result, err := h.service.Handle(c.Request.Context(), body.Column, body.Value)
	if err != nil {
		response.Fail(
			c,
			http.StatusInternalServerError,
			"PROCESS_FAILED",
			"Failed to process value",
		)
		return
	}

	response.OK(c, http.StatusOK, result)
}
