package visualiser

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type Repository interface {
	ListDatasets(ctx context.Context) ([]DatasetResponse, error)
	ListDataElements(ctx context.Context, dataSetID *string) ([]DataElementResponse, error)
	ListDataValues(ctx context.Context, req DataValuesRequest) (DataValuesResponse, error)
	ListThemes(ctx context.Context) ([]ThemeResponse, error)
	ListDataElementsByTheme(ctx context.Context, themeID string) ([]DataElementByThemeResponse, error)
	ListHIVSummary(ctx context.Context) ([]HIVSummaryResponse, error)
	ListHIVTested(ctx context.Context) ([]HIVTestedResponse, error)
	ListHIVRegimen(ctx context.Context) ([]HIVRegimenResponse, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListDatasets(ctx context.Context) ([]DatasetResponse, error) {
	query := `
		SELECT dataset_key, dataset_id, display_name, is_current, create_date
		FROM dwh.dim_dataset
		WHERE is_current = true
		AND dataset_key <> -1
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]DatasetResponse, 0)
	for rows.Next() {
		var dataset DatasetResponse
		if err := rows.Scan(&dataset.DatasetKey, &dataset.DatasetID, &dataset.DisplayName, &dataset.IsCurrent, &dataset.CreateDate); err != nil {
			continue
		}
		results = append(results, dataset)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListDataElements(ctx context.Context, dataSetID *string) ([]DataElementResponse, error) {
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

	var (
		rows *sql.Rows
		err  error
	)
	if dataSetID != nil && *dataSetID != "" {
		query += ` AND data_set_id = $1`
		rows, err = r.db.QueryContext(ctx, query, *dataSetID)
	} else {
		rows, err = r.db.QueryContext(ctx, query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]DataElementResponse, 0)
	for rows.Next() {
		var element DataElementResponse
		if err := rows.Scan(&element.DimDataElementMapKey, &element.DataElementID, &element.DataSetID, &element.DataElementShortName, &element.DataElementLongName, &element.RowVersion, &element.IsCurrent); err != nil {
			continue
		}
		results = append(results, element)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListDataValues(ctx context.Context, req DataValuesRequest) (DataValuesResponse, error) {
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

	levelOfCareClause := ""
	if len(req.LevelOfCare) > 0 {
		paramCounter++
		values = append(values, pq.Array(req.LevelOfCare))
		levelOfCareClause = fmt.Sprintf(" AND h.level_of_care = ANY($%d)", paramCounter)
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
		countryLevelOfCareJoin := ""
		if len(req.LevelOfCare) > 0 {
			countryLevelOfCareJoin = `
		     JOIN (
		       SELECT DISTINCT COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) AS facility_uid
		       FROM hiv.organisation_unit h
		       WHERE COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) IS NOT NULL
		         ` + levelOfCareClause + `
		     ) ou_filter
		       ON ou_filter.facility_uid = hs.org_unit_id
		     `
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
		     JOIN hiv.organisation_unit h
		       ON (h.facility_uid = si.selected_uid OR (h.level = '6' AND h.org_unit_id = si.selected_uid))
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '5'::text AS selected_level,
		       COALESCE(MIN(NULLIF(h.sub_county, '')), MIN(NULLIF(h.org_unit_name, '')), si.selected_uid) AS selected_name,
		       2 AS priority
		     FROM selected_input si
		     JOIN hiv.organisation_unit h ON
		      (h.sub_county_uid = si.selected_uid OR (h.level = '5' AND h.org_unit_id = si.selected_uid))
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '3'::text AS selected_level,
		       COALESCE(MIN(NULLIF(h.district, '')), si.selected_uid) AS selected_name,
		       3 AS priority
		     FROM selected_input si
		     JOIN hiv.organisation_unit h
		      ON h.district_uid = si.selected_uid
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '2'::text AS selected_level,
		       COALESCE(MIN(NULLIF(h.region, '')), si.selected_uid) AS selected_name,
		       4 AS priority
		     FROM selected_input si
		     JOIN hiv.organisation_unit h
		      ON h.region_uid = si.selected_uid
		     GROUP BY si.selected_uid

		     UNION ALL

		     SELECT
		       si.selected_uid,
		       '1'::text AS selected_level,
		       COALESCE(MIN(CASE WHEN h.level = '1' THEN NULLIF(h.org_unit_name, '') END), 'MoH - Uganda') AS selected_name,
		       5 AS priority
		     FROM selected_input si
		     JOIN hiv.organisation_unit h
		      ON (h.country_uid = si.selected_uid OR (h.level = '1' AND h.org_unit_id = si.selected_uid))
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
		       COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) AS facility_uid
		     FROM selected_units su
		     JOIN hiv.organisation_unit h
		      ON (
		        (su.selected_level = '6' AND (h.facility_uid = su.selected_uid OR (h.level = '6' AND h.org_unit_id = su.selected_uid)))
		        OR (su.selected_level = '2' AND h.region_uid = su.selected_uid)
		        OR (su.selected_level = '3' AND h.district_uid = su.selected_uid)
		        OR (su.selected_level = '5' AND (h.sub_county_uid = su.selected_uid OR (h.level = '5' AND h.org_unit_id = su.selected_uid)))
		      )
		     WHERE su.selected_level IN ('2', '3', '5', '6')
		       AND COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) IS NOT NULL
		       ` + levelOfCareClause + `
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
		       ON sf.facility_uid = hs.org_unit_id
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
		     ` + countryLevelOfCareJoin + `
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
		aggregationLevel := r.resolveAggregationLevel(ctx, requestedLevel, req.OU)
		orgUnitExpr, facilityExpr, levelExpr, regionExpr, districtExpr, subCountyExpr := aggregationExpressions(aggregationLevel)
		levelOfCareJoin := ""
		if len(req.LevelOfCare) > 0 {
			levelOfCareJoin = `
	       JOIN (
	         SELECT DISTINCT COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) AS facility_uid
	         FROM hiv.organisation_unit h
	         WHERE COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) IS NOT NULL
	           ` + levelOfCareClause + `
	       ) ou_filter
	         ON ou_filter.facility_uid = hs.org_unit_id
	       `
		}

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
	       ` + levelOfCareJoin + `
	       ` + whereClause + `
	       GROUP BY 1, 2, 3, 4, 6, 7, 8, 9, 10, 11
	       ORDER BY 
	          3,
	          1,
	          2,
	          4
	    `
	}

	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return DataValuesResponse{}, err
	}
	defer rows.Close()

	rowsList := make([]DataValueRowResponse, 0)
	for rows.Next() {
		var row DataValueRowResponse
		if err := rows.Scan(
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
		); err != nil {
			return DataValuesResponse{}, err
		}
		rowsList = append(rowsList, row)
	}
	if err := rows.Err(); err != nil {
		return DataValuesResponse{}, err
	}
	return DataValuesResponse{Rows: rowsList}, nil
}

func (r *postgresRepository) resolveAggregationLevel(ctx context.Context, requested *string, ou []string) string {
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
				SELECT 1 FROM hiv.organisation_unit h
				WHERE h.facility_uid = ANY($1)
				   OR (h.level = '6' AND h.org_unit_id = ANY($1))
			) THEN '6'
			WHEN EXISTS (
				SELECT 1 FROM hiv.organisation_unit h
				WHERE h.sub_county_uid = ANY($1)
				   OR (h.level = '5' AND h.org_unit_id = ANY($1))
			) THEN '5'
			WHEN EXISTS (
				SELECT 1 FROM hiv.organisation_unit h
				WHERE h.district_uid = ANY($1)
			) THEN '3'
			WHEN EXISTS (
				SELECT 1 FROM hiv.organisation_unit h
				WHERE h.region_uid = ANY($1)
			) THEN '2'
			WHEN EXISTS (
				SELECT 1 FROM hiv.organisation_unit h
				WHERE h.country_uid = ANY($1)
				   OR (h.level = '1' AND h.org_unit_id = ANY($1))
			) THEN '1'
			ELSE '6'
		END
	`

	var level string
	if err := r.db.QueryRowContext(ctx, query, pq.Array(ou)).Scan(&level); err != nil {
		return "6"
	}
	return level
}

func (r *postgresRepository) ListThemes(ctx context.Context) ([]ThemeResponse, error) {
	query := `
		SELECT theme_id, theme_name,dataset_name
		FROM report.themes
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ThemeResponse, 0)
	for rows.Next() {
		var theme ThemeResponse
		if err := rows.Scan(&theme.ThemeID, &theme.ThemeName, &theme.DatasetName); err != nil {
			continue
		}
		results = append(results, theme)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListDataElementsByTheme(ctx context.Context, themeID string) ([]DataElementByThemeResponse, error) {
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

	rows, err := r.db.QueryContext(ctx, query, themeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]DataElementByThemeResponse, 0)
	for rows.Next() {
		var element DataElementByThemeResponse
		if err := rows.Scan(&element.ThemeCategoryID, &element.ThemeID, &element.ThemeName, &element.DataElementKey, &element.DataElementID, &element.DataElementShortName); err != nil {
			continue
		}
		results = append(results, element)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListHIVSummary(ctx context.Context) ([]HIVSummaryResponse, error) {
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

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]HIVSummaryResponse, 0)
	for rows.Next() {
		var summary HIVSummaryResponse
		if err := rows.Scan(&summary.Year, &summary.Quarter, &summary.TotalTested, &summary.TotalHIVPositive, &summary.TotalLinkedCare, &summary.TotalEnrolledCare, &summary.TotalARTStarts, &summary.TotalARTWithCD4, &summary.TotalTXCurr, &summary.TotalTX1stLine, &summary.TotalTX2ndLine, &summary.TotalTX3rdPlus, &summary.TotalTBScreened, &summary.TotalMalnutritionAssessed); err != nil {
			continue
		}
		results = append(results, summary)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListHIVTested(ctx context.Context) ([]HIVTestedResponse, error) {
	query := `
		SELECT year,
			quarter,
			sum(total_tested) as tested_hiv,
			sum(total_positive) as tested_hiv_positive,
			sum(total_linked_to_care) as total_linked_care
		FROM hiv.mv_hmis_105_flat
		GROUP BY year, quarter
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]HIVTestedResponse, 0)
	for rows.Next() {
		var tested HIVTestedResponse
		if err := rows.Scan(&tested.Year, &tested.Quarter, &tested.TestedHIV, &tested.TestedHIVPositive, &tested.TotalLinkedCare); err != nil {
			continue
		}
		results = append(results, tested)
	}
	return results, rows.Err()
}

func (r *postgresRepository) ListHIVRegimen(ctx context.Context) ([]HIVRegimenResponse, error) {
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

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]HIVRegimenResponse, 0)
	for rows.Next() {
		var regimen HIVRegimenResponse
		if err := rows.Scan(&regimen.Year, &regimen.Quarter, &regimen.Actives, &regimen.TestedViralLoad, &regimen.TotalSuppressed); err != nil {
			continue
		}
		results = append(results, regimen)
	}
	return results, rows.Err()
}
