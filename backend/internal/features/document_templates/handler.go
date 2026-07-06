package document_templates

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/moh-sso-dashboard/internal/http/apierror"
	"github.com/moh-sso-dashboard/internal/http/response"
	"github.com/moh-sso-dashboard/internal/model"
)

type Handler struct {
	service       Service
	sheetService  SheetService
	columnService ColumnService
	db            *sql.DB
	remoteDB      *sql.DB
}

func NewHandler(
	service Service,
	sheetService SheetService,
	columnService ColumnService,
	db *sql.DB,
	remoteDB *sql.DB,
) *Handler {

	return &Handler{
		service:       service,
		sheetService:  sheetService,
		columnService: columnService,
		db:            db,
		remoteDB:      remoteDB,
	}
}

/* =========================================================
 * Create Template
 * ========================================================= */

func (h *Handler) CreateTemplate(c *gin.Context) {

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

	response.OK(c, http.StatusCreated, toTemplateResponse(template))
}

/* =========================================================
 * Get Template
 * ========================================================= */

func (h *Handler) GetTemplate(c *gin.Context) {

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

	response.OK(c, http.StatusOK, toTemplateRuntimeResponse(template))
}

/* =========================================================
 * Get Template Structure
 * ========================================================= */

func (h *Handler) GetTemplateStructure(c *gin.Context) {

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

	response.OK(c, http.StatusOK, toTemplateStructureResponse(structure))
}

/* =========================================================
 * List Templates
 * ========================================================= */

func (h *Handler) ListTemplates(c *gin.Context) {

	ctx := c.Request.Context()
	activeOnly := strings.EqualFold(c.Query("active"), "true")

	var (
		templates []model.DocumentTemplate
		err       error
	)

	if activeOnly {
		templates, err = h.service.ListActiveTemplates(ctx)
	} else {
		templates, err = h.service.ListTemplates(ctx)
	}

	if err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"LIST_TEMPLATES_FAILED",
			"Failed to list templates",
		)

		return
	}

	if templates == nil {
		templates = []model.DocumentTemplate{}
	}

	response.OK(c, http.StatusOK, templates)
}

/* =========================================================
 * Update Template
 * ========================================================= */

func (h *Handler) UpdateTemplate(c *gin.Context) {

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

	response.OK(c, http.StatusOK, toTemplateResponse(template))
}

/* =========================================================
 * Publish Template
 * ========================================================= */

func (h *Handler) PublishTemplate(c *gin.Context) {

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

func (h *Handler) ArchiveTemplate(c *gin.Context) {

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

func (h *Handler) DeleteTemplate(c *gin.Context) {

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

func (h *Handler) ListSheets(c *gin.Context) {

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

	response.OK(c, http.StatusOK, toSheetResponses(sheets))
}

/* =========================================================
 * Update Column
 * ========================================================= */

func (h *Handler) UpdateColumn(c *gin.Context) {

	columnID, err := uuid.Parse(c.Param("columnId"))
	if err != nil {

		response.Fail(
			c,
			http.StatusBadRequest,
			"INVALID_COLUMN_ID",
			"Invalid column ID",
		)

		return
	}

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

	req.ID = columnID

	col, err := h.columnService.UpdateColumn(c.Request.Context(), req)
	if err != nil {

		response.Fail(
			c,
			http.StatusInternalServerError,
			"UPDATE_COLUMN_FAILED",
			"Failed to update column",
		)

		return
	}

	// Update column_key separately when provided (not part of SQLC-generated update)
	if strings.TrimSpace(req.ColumnKey) != "" {
		if _, err := h.db.ExecContext(c.Request.Context(),
			`UPDATE document_template_columns SET column_key = $1, updated_at = NOW() WHERE id = $2`,
			strings.TrimSpace(req.ColumnKey), columnID,
		); err != nil {
			response.Fail(c, http.StatusInternalServerError, "UPDATE_COLUMN_KEY_FAILED", "Failed to update column key")
			return
		}
		col.ColumnKey = strings.TrimSpace(req.ColumnKey)
	}

	response.OK(c, http.StatusOK, col)
}

/* =========================================================
 * List Columns
 * ========================================================= */

func (h *Handler) ListColumns(c *gin.Context) {

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

	response.OK(c, http.StatusOK, toColumnResponses(columns))
}

/* =========================================================
 * Has Data
 * ========================================================= */

func (h *Handler) HasData(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.Fail(c, http.StatusBadRequest, "MISSING_TEMPLATE_CODE", "Template code is required")
		return
	}

	ctx := c.Request.Context()
	var hasData bool

	err := h.remoteDB.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM import.template_row_data
			WHERE template_code = $1 AND is_valid = TRUE
		)
	`, strings.ToUpper(strings.TrimSpace(code))).Scan(&hasData)

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "HAS_DATA_CHECK_FAILED", "Failed to check template data")
		return
	}

	response.OK(c, http.StatusOK, gin.H{"has_data": hasData})
}

/* =========================================================
 * Replace Structure
 * ========================================================= */

func (h *Handler) ReplaceStructure(c *gin.Context) {
	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "INVALID_ID", "invalid template id")
		return
	}

	var req struct {
		Sheets []model.CreateTemplateSheetWithColumnsRequest `json:"sheets"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "VALIDATION_FAILED", "invalid request payload")
		return
	}

	structure, err := h.service.ReplaceStructure(c.Request.Context(), templateID, req.Sheets)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "REPLACE_FAILED", err.Error())
		return
	}

	response.OK(c, http.StatusOK, structure)
}

/* =========================================================
 * Create Template With Structure
 * ========================================================= */

func (h *Handler) CreateTemplateWithStructure(c *gin.Context) {
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

		if errors.Is(err, ErrTemplateCodeExists) {
			response.Fail(
				c,
				http.StatusConflict,
				"TEMPLATE_CODE_EXISTS",
				"A template with this code already exists",
			)

			return
		}

		response.Fail(
			c,
			http.StatusInternalServerError,
			"CREATE_TEMPLATE_FAILED",
			"request failed",
		)

		return
	}

	response.OK(c, http.StatusCreated, toTemplateStructureResponse(template))
}
