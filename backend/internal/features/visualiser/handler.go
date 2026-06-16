package visualiser

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/moh-sso-dashboard/internal/config"
	"github.com/moh-sso-dashboard/internal/dto"
	"github.com/moh-sso-dashboard/internal/http/response"
)

type Handler struct {
	config *config.Config
	db     *sql.DB
}

func NewHandler(
	config *config.Config,
	db *sql.DB,
) *Handler {
	return &Handler{
		config: config,
		db:     db,
	}
}

// GetDatasets gets all datasets
func (h *Handler) GetDatasets(c *gin.Context) {
	ctx := c.Request.Context()
	query := `
		SELECT dataset_key, dataset_id, display_name, is_current, create_date
		FROM dwh.dim_dataset
		WHERE is_current = true
		AND dataset_key <> -1
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
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

	response.OK(c, http.StatusOK, results)
}

// GetDataElements gets data elements filtered by data_set_id
func (h *Handler) GetDataElements(c *gin.Context) {
	ctx := c.Request.Context()

	dataSetID := c.Query("data_set_id")
	if dataSetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data_set_id is required"})
		return
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
		AND data_set_id = $1
	`

	rows, err := h.db.QueryContext(ctx, query, dataSetID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
		return
	}
	defer rows.Close()

	results := make([]dto.DataElement, 0)
	for rows.Next() {
		var de dto.DataElement
		err := rows.Scan(&de.DimDataElementMapKey, &de.DataElementID, &de.DataSetID, &de.DataElementShortName, &de.DataElementLongName, &de.RowVersion, &de.IsCurrent)
		if err != nil {
			continue
		}
		results = append(results, de)
	}

	response.OK(c, http.StatusOK, results)
}

// GetDataValues gets data values with optional filters
func (h *Handler) GetDataValues(c *gin.Context) {
	ctx := c.Request.Context()

	type Request struct {
		OU           []string `json:"ou"`
		PE           []string `json:"pe"`
		DX           []string `json:"dx"`
		StartDate    string   `json:"startDate"`
		EndDate      string   `json:"endDate"`
		OrgunitLevel any      `json:"orgunitLevel"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON")
		return
	}

	var conditions []string
	var values []interface{}
	paramCounter := 0
	requestedLevel := normalizeOrgunitLevel(req.OrgunitLevel)

	if len(req.OU) > 0 {
		paramCounter++
		values = append(values, pq.Array(req.OU))
	}

	if len(req.DX) > 0 {
		paramCounter++
		values = append(values, pq.Array(req.DX))
		conditions = append(conditions, fmt.Sprintf("hs.data_element_id = ANY($%d)", paramCounter))
	}

	if len(req.PE) > 0 {
		paramCounter++
		values = append(values, pq.Array(req.PE))
		conditions = append(conditions, fmt.Sprintf(`hs."period" = ANY($%d)`, paramCounter))
	} else if req.StartDate != "" && req.EndDate != "" {
		paramCounter++
		values = append(values, req.StartDate)
		conditions = append(conditions, fmt.Sprintf("hs.tperiod >= $%d", paramCounter))

		paramCounter++
		values = append(values, req.EndDate)
		conditions = append(conditions, fmt.Sprintf("hs.tperiod <= $%d", paramCounter))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	var query string
	if len(req.OU) > 0 {
		selectedLevelClause := ""
		if requestedLevel != nil {
			paramCounter++
			values = append(values, *requestedLevel)
			selectedLevelClause = fmt.Sprintf("WHERE su.selected_level = $%d", paramCounter)
		}

		query = `
		   WITH selected_input AS (
		     SELECT DISTINCT unnest($1::text[]) AS selected_uid
		   ),
		   selected_candidates AS (
		     SELECT
		       si.selected_uid,
		       '6'::text AS selected_level,
		       COALESCE(MIN(NULLIF(h.facility_name, '')), MIN(NULLIF(h.org_unit_name, '')), si.selected_uid) AS selected_name,
		       1 AS priority
		     FROM selected_input si
		     JOIN dwh.dim_org_hierarchy h
		       ON h.is_current = true
		      AND (h.facility_uid = si.selected_uid OR (h.level = '6' AND h.org_unit_id = si.selected_uid))
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '5'::text AS selected_level,
		       COALESCE(MIN(NULLIF(h.sub_county, '')), MIN(NULLIF(h.org_unit_name, '')), si.selected_uid) AS selected_name,
		       2 AS priority
		     FROM selected_input si
		     JOIN dwh.dim_org_hierarchy h
		       ON h.is_current = true
		      AND (h.sub_county_uid = si.selected_uid OR (h.level = '5' AND h.org_unit_id = si.selected_uid))
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '3'::text AS selected_level,
		       COALESCE(MIN(NULLIF(h.district, '')), si.selected_uid) AS selected_name,
		       3 AS priority
		     FROM selected_input si
		     JOIN dwh.dim_org_hierarchy h
		       ON h.is_current = true
		      AND h.district_uid = si.selected_uid
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '2'::text AS selected_level,
		       COALESCE(MIN(NULLIF(h.region, '')), si.selected_uid) AS selected_name,
		       4 AS priority
		     FROM selected_input si
		     JOIN dwh.dim_org_hierarchy h
		       ON h.is_current = true
		      AND h.region_uid = si.selected_uid
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '1'::text AS selected_level,
		       COALESCE(MIN(CASE WHEN h.level = '1' THEN NULLIF(h.org_unit_name, '') END), 'MoH - Uganda') AS selected_name,
		       5 AS priority
		     FROM selected_input si
		     JOIN dwh.dim_org_hierarchy h
		       ON h.is_current = true
		      AND (h.country_uid = si.selected_uid OR (h.level = '1' AND h.org_unit_id = si.selected_uid))
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '6'::text AS selected_level,
		       si.selected_uid AS selected_name,
		       99 AS priority
		     FROM selected_input si
		   ),
		   selected_units AS (
		     SELECT DISTINCT ON (su.selected_uid)
		       su.selected_uid,
		       su.selected_level,
		       su.selected_name
		     FROM selected_candidates su
		     ` + selectedLevelClause + `
		     ORDER BY su.selected_uid, su.priority
		   ),
		   selected_facilities AS (
		     SELECT DISTINCT
		       su.selected_uid,
		       su.selected_level,
		       su.selected_name,
		       h.dim_org_hierarchy_key
		     FROM selected_units su
		     JOIN dwh.dim_org_hierarchy h
		       ON h.is_current = true
		      AND (
		        (su.selected_level = '6' AND (h.facility_uid = su.selected_uid OR (h.level = '6' AND h.org_unit_id = su.selected_uid)))
		        OR (su.selected_level = '2' AND h.region_uid = su.selected_uid)
		        OR (su.selected_level = '3' AND h.district_uid = su.selected_uid)
		        OR (su.selected_level = '5' AND (h.sub_county_uid = su.selected_uid OR (h.level = '5' AND h.org_unit_id = su.selected_uid)))
		      )
		     WHERE su.selected_level IN ('2', '3', '5', '6')
		   )
		   SELECT
		     x.org_unit_id,
		     x.data_element_id,
		     x."period",
		     x.category_combo,
		     x.value,
		     x.facility,
		     x."level",
		     x.region,
		     x.district,
		     x.sub_county,
		     x.dataelement
		   FROM (
		     SELECT
		       sf.selected_uid AS org_unit_id,
		       hs.data_element_id,
		       hs."period",
		       hs.category_combo,
		       SUM(
		         CASE
		           WHEN hs.value ~ '^\s*-?\d+(\.\d+)?\s*$' THEN TRIM(hs.value)::numeric
		           ELSE 0
		         END
		       )::bigint AS value,
		       sf.selected_name AS facility,
		       sf.selected_level AS "level",
		       CASE
		         WHEN sf.selected_level = '2' THEN sf.selected_name
		         WHEN sf.selected_level IN ('3', '5', '6') THEN COALESCE(MAX(hs.region), '')
		         ELSE ''
		       END AS region,
		       CASE
		         WHEN sf.selected_level = '3' THEN sf.selected_name
		         WHEN sf.selected_level IN ('5', '6') THEN COALESCE(MAX(hs.district), '')
		         ELSE ''
		       END AS district,
		       CASE
		         WHEN sf.selected_level = '5' THEN sf.selected_name
		         WHEN sf.selected_level = '6' THEN COALESCE(MAX(hs.sub_county), '')
		         ELSE ''
		       END AS sub_county,
		       hs.dataelement
		     FROM report.hmis_summary hs
		     JOIN selected_facilities sf
		       ON sf.dim_org_hierarchy_key = hs.dim_org_hierarchy_key
		     ` + whereClause + `
		     GROUP BY
		       sf.selected_uid,
		       sf.selected_name,
		       sf.selected_level,
		       hs.data_element_id,
		       hs."period",
		       hs.category_combo,
		       hs.dataelement

		     UNION ALL

		     SELECT
		       su.selected_uid AS org_unit_id,
		       hs.data_element_id,
		       hs."period",
		       hs.category_combo,
		       SUM(
		         CASE
		           WHEN hs.value ~ '^\s*-?\d+(\.\d+)?\s*$' THEN TRIM(hs.value)::numeric
		           ELSE 0
		         END
		       )::bigint AS value,
		       su.selected_name AS facility,
		       su.selected_level AS "level",
		       '' AS region,
		       '' AS district,
		       '' AS sub_county,
		       hs.dataelement
		     FROM report.hmis_summary hs
		     JOIN selected_units su
		       ON su.selected_level = '1'
		     ` + whereClause + `
		     GROUP BY
		       su.selected_uid,
		       su.selected_name,
		       su.selected_level,
		       hs.data_element_id,
		       hs."period",
		       hs.category_combo,
		       hs.dataelement
		   ) x
		   ORDER BY
		     x."period",
		     x."level",
		     x.facility,
		     x.data_element_id,
		     x.category_combo
		`
	} else {
		aggregationLevel := h.resolveAggregationLevel(ctx, requestedLevel, req.OU)
		orgUnitExpr, facilityExpr, levelExpr, regionExpr, districtExpr, subCountyExpr := aggregationExpressions(aggregationLevel)

		query = `
		   SELECT 
		      ` + orgUnitExpr + ` AS org_unit_id,
		      hs.data_element_id,
		      hs."period",
		      hs.category_combo,
		      SUM(
		        CASE
		          WHEN hs.value ~ '^\s*-?\d+(\.\d+)?\s*$' THEN TRIM(hs.value)::numeric
		          ELSE 0
		        END
		      )::bigint AS value,
		      ` + facilityExpr + ` AS facility,
		      ` + levelExpr + ` AS "level",
		      ` + regionExpr + ` AS region,
	          ` + districtExpr + ` AS district,
	          ` + subCountyExpr + ` AS sub_county,
	          hs.dataelement
	       FROM report.hmis_summary hs
	       ` + whereClause + `
	       GROUP BY 1, 2, 3, 4, 6, 7, 8, 9, 10, 11
	       ORDER BY 
	          3,
	          1,
	          2,
	          4
	    `
	}

	rows, err := h.db.QueryContext(ctx, query, values...)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
		return
	}
	defer rows.Close()

	rowsList := make([]dto.DataValueRow, 0)
	for rows.Next() {
		var row dto.DataValueRow
		err := rows.Scan(
			&row.OrgUnitID,
			&row.DataElementID,
			&row.Period,
			&row.CategoryCombo,
			&row.Value,
			&row.Facility,
			&row.Level,
			&row.Region,
			&row.District,
			&row.SubCounty,
			&row.Dataelement,
		)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
			return
		}
		rowsList = append(rowsList, row)
	}

	if err = rows.Err(); err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
		return
	}

	response.OK(c, http.StatusOK, struct {
		Rows []dto.DataValueRow `json:"rows"`
	}{Rows: rowsList})
}

func (h *Handler) resolveAggregationLevel(ctx context.Context, requested *string, ou []string) string {
	if requested != nil {
		level := strings.TrimSpace(*requested)
		switch level {
		case "1", "2", "3", "5", "6":
			return level
		}
	}

	if len(ou) == 0 {
		return "6"
	}

	query := `
		SELECT CASE
			WHEN EXISTS (
				SELECT 1 FROM dwh.dim_org_hierarchy h
				WHERE h.is_current = true
				  AND h.facility_uid = ANY($1)
			) THEN '6'
			WHEN EXISTS (
				SELECT 1 FROM dwh.dim_org_hierarchy h
				WHERE h.is_current = true
				  AND (h.sub_county_uid = ANY($1) OR h.org_unit_id = ANY($1))
			) THEN '5'
			WHEN EXISTS (
				SELECT 1 FROM dwh.dim_org_hierarchy h
				WHERE h.is_current = true
				  AND h.district_uid = ANY($1)
			) THEN '3'
			WHEN EXISTS (
				SELECT 1 FROM dwh.dim_org_hierarchy h
				WHERE h.is_current = true
				  AND h.region_uid = ANY($1)
			) THEN '2'
			WHEN EXISTS (
				SELECT 1 FROM dwh.dim_org_hierarchy h
				WHERE h.is_current = true
				  AND h.country_uid = ANY($1)
			) THEN '1'
			ELSE '6'
		END
	`

	var level string
	if err := h.db.QueryRowContext(ctx, query, pq.Array(ou)).Scan(&level); err != nil {
		return "6"
	}
	return level
}

func normalizeOrgunitLevel(v any) *string {
	switch x := v.(type) {
	case nil:
		return nil
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		return &s
	case float64:
		// JSON numbers decode to float64 by default.
		s := strconv.FormatInt(int64(x), 10)
		return &s
	default:
		return nil
	}
}

func aggregationExpressions(level string) (orgUnitExpr, facilityExpr, levelExpr, regionExpr, districtExpr, subCountyExpr string) {
	switch level {
	case "1":
		return "'National'", "'National'", "'1'", "''", "''", "''"
	case "2":
		return "COALESCE(NULLIF(hs.region, ''), 'Unknown Region')", "COALESCE(NULLIF(hs.region, ''), 'Unknown Region')", "'2'", "COALESCE(NULLIF(hs.region, ''), 'Unknown Region')", "''", "''"
	case "3":
		return "COALESCE(NULLIF(hs.district, ''), 'Unknown District')", "COALESCE(NULLIF(hs.district, ''), 'Unknown District')", "'3'", "COALESCE(hs.region, '')", "COALESCE(NULLIF(hs.district, ''), 'Unknown District')", "''"
	case "5":
		return "COALESCE(NULLIF(hs.sub_county, ''), 'Unknown Subcounty')", "COALESCE(NULLIF(hs.sub_county, ''), 'Unknown Subcounty')", "'5'", "COALESCE(hs.region, '')", "COALESCE(hs.district, '')", "COALESCE(NULLIF(hs.sub_county, ''), 'Unknown Subcounty')"
	default:
		return "hs.org_unit_id", "COALESCE(hs.facility, hs.org_unit_id)", "COALESCE(NULLIF(hs.\"level\", ''), '6')", "COALESCE(hs.region, '')", "COALESCE(hs.district, '')", "COALESCE(hs.sub_county, '')"
	}
}

// GetThemes gets all themes
func (h *Handler) GetThemes(c *gin.Context) {
	ctx := c.Request.Context()
	query := `
		SELECT theme_id, theme_name
		FROM report.themes
	`

	rows, err := h.db.QueryContext(ctx, query)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
		return
	}
	defer rows.Close()

	var results []dto.Theme
	for rows.Next() {
		var t dto.Theme
		err := rows.Scan(&t.ThemeID, &t.ThemeName, &t.DatasetName)
		if err != nil {
			continue
		}
		results = append(results, t)
	}

	response.OK(c, http.StatusOK, results)
}

// GetDataElementsByTheme gets data elements for a specific theme
func (h *Handler) GetDataElementsByTheme(c *gin.Context) {
	ctx := c.Request.Context()
	type Request struct {
		ThemeID string `json:"theme_id" binding:"required"`
	}

	var req Request
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "theme_id is required")
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
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
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

	response.OK(c, http.StatusOK, results)
}

// GetHIVSummary gets HIV summary data
func (h *Handler) GetHIVSummary(c *gin.Context) {
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
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
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

	response.OK(c, http.StatusOK, results)
}

// GetHIVTested gets HIV tested data
func (h *Handler) GetHIVTested(c *gin.Context) {
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
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
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

	response.OK(c, http.StatusOK, results)
}

// GetHIVRegimen gets HIV regimen data
func (h *Handler) GetHIVRegimen(c *gin.Context) {

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
		response.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "request failed")
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

	response.OK(c, http.StatusOK, results)
}
