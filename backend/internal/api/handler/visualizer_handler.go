package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/dto"
)

type VisualiserHandler struct {
	config *config.Config
	db     *sql.DB
}

func NewVisualiserHandler(
	config *config.Config,
	db *sql.DB,
) *VisualiserHandler {
	return &VisualiserHandler{
		config: config,
		db:     db,
	}
}

// GetDatasets gets all datasets
func (h *VisualiserHandler) GetDatasets(c *gin.Context) {
	ctx := c.Request.Context()
	query := `
		SELECT dataset_key, dataset_id, display_name, is_current, create_date
		FROM dwh.dim_dataset
		WHERE is_current = true
		AND dataset_key <> -1
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.Dataset
	for rows.Next() {
		var d dto.Dataset
		err := rows.Scan(&d.DatasetKey, &d.DatasetID, &d.DisplayName, &d.IsCurrent, &d.CreateDate)
		if err != nil {
			continue
		}
		results = append(results, d)
	}

	c.JSON(http.StatusOK, results)
}

// GetDataElements gets data elements, optionally filtered by data_set_id
func (h *VisualiserHandler) GetDataElements(c *gin.Context) {
	ctx := c.Request.Context()
	type Request struct {
		DataSetID *string `json:"data_set_id"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		req = Request{}
	}

	query := `
		SELECT 
			dim_data_element_map_key, 
			data_element_id, 
			data_set_id, 
			data_element_short_name, 
			data_element_long_name,
			row_version, 
			is_current
		FROM dwh.dim_hmis_data_element_map
		WHERE dim_data_element_map_key <> -1
		AND is_current = true
	`

	var rows *sql.Rows
	var err error

	if req.DataSetID != nil && *req.DataSetID != "" {
		query += ` AND data_set_id = $1`
		rows, err = h.db.QueryContext(ctx, query, *req.DataSetID)
	} else {
		rows, err = h.db.QueryContext(ctx, query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.DataElement
	for rows.Next() {
		var de dto.DataElement
		err := rows.Scan(&de.DimDataElementMapKey, &de.DataElementID, &de.DataSetID, &de.DataElementShortName, &de.DataElementLongName, &de.RowVersion, &de.IsCurrent)
		if err != nil {
			continue
		}
		results = append(results, de)
	}

	c.JSON(http.StatusOK, results)
}

// GetDataValues gets data values with optional filters
func (h *VisualiserHandler) GetDataValues(c *gin.Context) {
	ctx := c.Request.Context()
	type Request struct {
		OU           []string `json:"ou"`
		PE           []string `json:"pe"`
		DX           []string `json:"dx"`
		StartDate    string   `json:"startDate"`
		EndDate      string   `json:"endDate"`
		OrgunitLevel *string  `json:"orgunitLevel"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	var conditions []string
	var values []interface{}
	paramCounter := 0

	if len(req.OU) > 0 {
		paramCounter++
		values = append(values, req.OU)
		conditions = append(conditions, fmt.Sprintf("org_unit_id = ANY($%d)", paramCounter))
	}

	if len(req.DX) > 0 {
		paramCounter++
		values = append(values, req.DX)
		conditions = append(conditions, fmt.Sprintf("data_element_id = ANY($%d)", paramCounter))
	}

	if len(req.PE) > 0 {
		paramCounter++
		values = append(values, req.PE)
		conditions = append(conditions, fmt.Sprintf(`"period" = ANY($%d)`, paramCounter))
	} else if req.StartDate != "" && req.EndDate != "" {
		paramCounter++
		values = append(values, req.StartDate)
		conditions = append(conditions, fmt.Sprintf("tperiod >= $%d", paramCounter))
		paramCounter++
		values = append(values, req.EndDate)
		conditions = append(conditions, fmt.Sprintf("tperiod <= $%d", paramCounter))
	}

	if req.OrgunitLevel != nil && *req.OrgunitLevel != "" {
		paramCounter++
		values = append(values, *req.OrgunitLevel)
		conditions = append(conditions, fmt.Sprintf(`"level" = $%d`, paramCounter))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT 
			org_unit_id,
			data_element_id,
			"period",
			category_combo,
			value,
			facility,
			"level",
			region,
			district,
			sub_county,
			dataelement
		FROM report.hmis_summary
		` + whereClause + `
		ORDER BY 
			"period",
			org_unit_id,
			data_element_id,
			category_combo
	`

	rows, err := h.db.QueryContext(ctx, query, values...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var rowsList []dto.DataValueRow
	for rows.Next() {
		var row dto.DataValueRow
		err := rows.Scan(&row.OrgUnitID, &row.DataElementID, &row.Period, &row.CategoryCombo, &row.Value, &row.Facility, &row.Level, &row.Region, &row.District, &row.SubCounty, &row.Dataelement)
		if err != nil {
			continue
		}
		rowsList = append(rowsList, row)
	}

	c.JSON(http.StatusOK, gin.H{"rows": rowsList})
}

// GetThemes gets all themes
func (h *VisualiserHandler) GetThemes(c *gin.Context) {
	ctx := c.Request.Context()
	query := `
		SELECT theme_id, theme_name 
		FROM report.themes
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.Theme
	for rows.Next() {
		var t dto.Theme
		err := rows.Scan(&t.ThemeID, &t.ThemeName)
		if err != nil {
			continue
		}
		results = append(results, t)
	}

	c.JSON(http.StatusOK, results)
}

// GetDataElementsByTheme gets data elements for a specific theme
func (h *VisualiserHandler) GetDataElementsByTheme(c *gin.Context) {
	ctx := c.Request.Context()
	type Request struct {
		ThemeID string `json:"theme_id" binding:"required"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "theme_id is required"})
		return
	}

	query := `
		SELECT 
			t.theme_category_id, 
			t.theme_id,
			n.theme_name,
			t.data_element_id as data_element_key,
			m.data_element_id, 
			m.data_element_short_name
		FROM report.theme_category t
		INNER JOIN dwh.dim_hmis_data_element_map_current m ON m.dim_data_element_map_current_key = t.data_element_id
		INNER JOIN report.themes n ON n.theme_id = t.theme_id 
		WHERE t.theme_id = $1
	`

	rows, err := h.db.QueryContext(ctx, query, req.ThemeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.DataElementByTheme
	for rows.Next() {
		var det dto.DataElementByTheme
		err := rows.Scan(&det.ThemeCategoryID, &det.ThemeID, &det.ThemeName, &det.DataElementKey, &det.DataElementID, &det.DataElementShortName)
		if err != nil {
			continue
		}
		results = append(results, det)
	}

	c.JSON(http.StatusOK, results)
}

// GetHIVSummary gets HIV summary data
func (h *VisualiserHandler) GetHIVSummary(c *gin.Context) {
	ctx := c.Request.Context()
	query := `
		SELECT
			year,
			quarter,
			SUM(tested) AS total_tested,
			SUM(hiv_pos) AS total_hiv_positive,
			SUM(linked_care) AS total_linked_care,
			SUM(enrolled_care) AS total_enrolled_care,
			SUM(art_starts) AS total_art_starts,
			SUM(art_with_cd4) AS total_art_with_cd4,
			SUM(tx_curr) AS total_tx_curr,
			SUM(tx_1st_line) AS total_tx_1st_line,
			SUM(tx_2nd_line) AS total_tx_2nd_line,
			SUM(tx_3rd_plus) AS total_tx_3rd_plus,
			SUM(tb_screened) AS total_tb_screened,
			SUM(maln_assessed) AS total_malnutrition_assessed
		FROM hiv.mv_hiv_quarterly_agg
		GROUP BY year, quarter
		ORDER BY year
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.HIVSummary
	for rows.Next() {
		var hiv dto.HIVSummary
		err := rows.Scan(&hiv.Year, &hiv.Quarter, &hiv.TotalTested, &hiv.TotalHIVPositive, &hiv.TotalLinkedCare, &hiv.TotalEnrolledCare, &hiv.TotalARTStarts, &hiv.TotalARTWithCD4, &hiv.TotalTXCurr, &hiv.TotalTX1stLine, &hiv.TotalTX2ndLine, &hiv.TotalTX3rdPlus, &hiv.TotalTBScreened, &hiv.TotalMalnutritionAssessed)
		if err != nil {
			continue
		}
		results = append(results, hiv)
	}

	c.JSON(http.StatusOK, results)
}

// GetHIVTested gets HIV tested data
func (h *VisualiserHandler) GetHIVTested(c *gin.Context) {
	ctx := c.Request.Context()
	query := `
		SELECT year,
			quarter,
			sum(total_tested) as tested_hiv,
			sum(total_positive) as tested_hiv_positive,
			sum(total_linked_to_care) as total_linked_care
		FROM hiv.mv_hmis_105_flat
		GROUP BY year, quarter
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.HIVTested
	for rows.Next() {
		var hiv dto.HIVTested
		err := rows.Scan(&hiv.Year, &hiv.Quarter, &hiv.TestedHIV, &hiv.TestedHIVPositive, &hiv.TotalLinkedCare)
		if err != nil {
			continue
		}
		results = append(results, hiv)
	}

	c.JSON(http.StatusOK, results)
}

// GetHIVRegimen gets HIV regimen data
func (h *VisualiserHandler) GetHIVRegimen(c *gin.Context) {

	ctx := c.Request.Context()

	query := `
		SELECT
			year,
			quarter,
			SUM(
				COALESCE(active_first_line_regimen, 0)
				+ COALESCE(active_on_art_2nd_line_regimen, 0)
				+ COALESCE(active_on_art_3rd_line, 0)
			) AS actives,
			SUM(total_tested_vl_past_12months) as tested_viral_load,
			SUM(clients_suppressed_12months) as total_suppressed
		FROM hiv.mv_hmis_106_flat
		GROUP BY year, quarter
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var results []dto.HIVRegimen
	for rows.Next() {
		var hiv dto.HIVRegimen
		err := rows.Scan(&hiv.Year, &hiv.Quarter, &hiv.Actives, &hiv.TestedViralLoad, &hiv.TotalSuppressed)
		if err != nil {
			continue
		}
		results = append(results, hiv)
	}

	c.JSON(http.StatusOK, results)
}
