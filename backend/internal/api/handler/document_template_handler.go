package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/service"
)

type DocumentTemplateHandler struct {
	service       service.DocumentTemplateService
	sheetService  service.DocumentTemplateSheetService
	columnService service.DocumentTemplateColumnService
}

func NewDocumentTemplateHandler(
	service service.DocumentTemplateService,
	sheetService service.DocumentTemplateSheetService,
	columnService service.DocumentTemplateColumnService,
) *DocumentTemplateHandler {

	return &DocumentTemplateHandler{
		service: service,
	}
}

/* =========================================================
 * Create Template
 * ========================================================= */

func (h *DocumentTemplateHandler) CreateTemplate(c *gin.Context) {

	var req model.CreateTemplateRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)

		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {

		response.Fail(
			c,
			http.StatusUnauthorized,
			"INVALID_USER",
			"Invalid authenticated user",
		)

		return
	}

	req.CreatedBy = userID

	template, err := h.service.CreateTemplate(
		c.Request.Context(),
		req,
	)
	if err != nil {

		var apiErr *apierror.APIError

		if errors.As(err, &apiErr) {

			response.Fail(
				c,
				apiErr.HTTPStatus,
				apiErr.Code,
				apiErr.Message,
			)

			return
		}

		response.Fail(
			c,
			http.StatusInternalServerError,
			"CREATE_TEMPLATE_FAILED",
			"Failed to create template",
		)

		return
	}

	response.OK(c, http.StatusCreated, template)
}

/* =========================================================
 * Get Template
 * ========================================================= */

func (h *DocumentTemplateHandler) GetTemplate(c *gin.Context) {

	code := c.Param("code")

	if code == "" {

		response.Fail(
			c,
			http.StatusBadRequest,
			"MISSING_TEMPLATE_CODE",
			"Template code is required",
		)

		return
	}

	template, err := h.service.GetTemplateRuntime(
		c.Request.Context(),
		code,
	)
	if err != nil {

		response.Fail(
			c,
			http.StatusNotFound,
			"TEMPLATE_NOT_FOUND",
			"Template not found",
		)

		return
	}

	response.OK(c, http.StatusOK, template)
}

/* =========================================================
 * Get Template Structure
 * ========================================================= */

func (h *DocumentTemplateHandler) GetTemplateStructure(c *gin.Context) {

	code := c.Param("code")

	if code == "" {

		response.Fail(
			c,
			http.StatusBadRequest,
			"MISSING_TEMPLATE_CODE",
			"Template code is required",
		)

		return
	}

	structure, err := h.service.GetTemplateStructure(
		c.Request.Context(),
		code,
	)
	if err != nil {

		response.Fail(
			c,
			http.StatusNotFound,
			"TEMPLATE_NOT_FOUND",
			"Template structure not found",
		)

		return
	}

	response.OK(c, http.StatusOK, structure)
}

/* =========================================================
 * List Templates
 * ========================================================= */

func (h *DocumentTemplateHandler) ListTemplates(c *gin.Context) {

	templates, err := h.service.ListTemplates(
		c.Request.Context(),
	)
	if err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"LIST_TEMPLATES_FAILED",
			"Failed to list templates",
		)

		return
	}

	response.OK(c, http.StatusOK, templates)
}

/* =========================================================
 * Update Template
 * ========================================================= */

func (h *DocumentTemplateHandler) UpdateTemplate(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_TEMPLATE_ID",
			"Invalid template ID",
		)

		return
	}

	var req model.UpdateTemplateRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)

		return
	}

	req.ID = id

	template, err := h.service.UpdateTemplate(
		c.Request.Context(),
		req,
	)
	if err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"UPDATE_TEMPLATE_FAILED",
			"Failed to update template",
		)

		return
	}

	response.OK(c, http.StatusOK, template)
}

/* =========================================================
 * Publish Template
 * ========================================================= */

func (h *DocumentTemplateHandler) PublishTemplate(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_TEMPLATE_ID",
			"Invalid template ID",
		)

		return
	}

	if err := h.service.PublishTemplate(
		c.Request.Context(),
		id,
	); err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"PUBLISH_TEMPLATE_FAILED",
			"Failed to publish template",
		)

		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Archive Template
 * ========================================================= */

func (h *DocumentTemplateHandler) ArchiveTemplate(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_TEMPLATE_ID",
			"Invalid template ID",
		)

		return
	}

	if err := h.service.ArchiveTemplate(
		c.Request.Context(),
		id,
	); err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"ARCHIVE_TEMPLATE_FAILED",
			"Failed to archive template",
		)

		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * Delete Template
 * ========================================================= */

func (h *DocumentTemplateHandler) DeleteTemplate(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_TEMPLATE_ID",
			"Invalid template ID",
		)

		return
	}

	if err := h.service.DeleteTemplate(
		c.Request.Context(),
		id,
	); err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"DELETE_TEMPLATE_FAILED",
			"Failed to delete template",
		)

		return
	}

	c.Status(http.StatusNoContent)
}

/* =========================================================
 * List Sheets
 * ========================================================= */

func (h *DocumentTemplateHandler) ListSheets(c *gin.Context) {

	templateID, err := uuid.Parse(c.Param("templateId"))
	if err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_TEMPLATE_ID",
			"Invalid template ID",
		)

		return
	}

	sheets, err := h.sheetService.ListSheets(
		c.Request.Context(),
		templateID,
	)
	if err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"LIST_SHEETS_FAILED",
			"Failed to list sheets",
		)

		return
	}

	response.OK(c, http.StatusOK, sheets)
}

/* =========================================================
 * List Columns
 * ========================================================= */

func (h *DocumentTemplateHandler) ListColumns(c *gin.Context) {

	sheetID, err := uuid.Parse(c.Param("sheetId"))
	if err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_SHEET_ID",
			"Invalid sheet ID",
		)

		return
	}

	columns, err := h.columnService.ListColumns(
		c.Request.Context(),
		sheetID,
	)
	if err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"LIST_COLUMNS_FAILED",
			"Failed to list columns",
		)

		return
	}

	response.OK(c, http.StatusOK, columns)
}

func (h *DocumentTemplateHandler) CreateTemplateWithStructure(c *gin.Context) {
	var req model.CreateTemplateStructureRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(
			c,
			http.StatusBadRequest,
			"VALIDATION_FAILED",
			"Invalid request payload",
		)

		return
	}

	userID, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		response.Fail(
			c,
			http.StatusUnauthorized,
			"INVALID_USER",
			"Invalid authenticated user",
		)

		return
	}

	req.Template.CreatedBy = userID

	template, err := h.service.CreateTemplateStructure(
		c.Request.Context(),
		req,
	)
	if err != nil {
		var apiErr *apierror.APIError

		if errors.As(err, &apiErr) {
			response.Fail(
				c,
				apiErr.HTTPStatus,
				apiErr.Code,
				apiErr.Message,
			)

			return
		}

		response.Fail(
			c,
			http.StatusInternalServerError,
			"CREATE_TEMPLATE_FAILED",
			err.Error(),
		)

		return
	}

	response.OK(c, http.StatusCreated, template)
}
