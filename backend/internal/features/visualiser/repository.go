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

	orgUnitFilterClause := ""
	if len(req.LevelOfCare) > 0 {
		paramCounter++
		values = append(values, pq.Array(req.LevelOfCare))
		orgUnitFilterClause += fmt.Sprintf(" AND h.level_of_care = ANY($%d)", paramCounter)
	}
	if len(req.Ownership) > 0 {
		paramCounter++
		values = append(values, pq.Array(req.Ownership))
		orgUnitFilterClause += fmt.Sprintf(" AND h.ownership = ANY($%d)", paramCounter)
	}
	aggregateAllCareAndOwnership := len(req.LevelOfCare) == 0 && len(req.Ownership) == 0
	levelOfCareExpr := "COALESCE(NULLIF(oa.level_of_care, ''), 'ALL')"
	ownershipExpr := "COALESCE(NULLIF(oa.ownership, ''), 'ALL')"
	if aggregateAllCareAndOwnership {
		levelOfCareExpr = "'ALL'"
		ownershipExpr = "'ALL'"
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
		countryWhereClause := whereClause
		if countryWhereClause == "" {
			countryWhereClause = " WHERE su.selected_level = '1'"
		} else {
			countryWhereClause += " AND su.selected_level = '1'"
		}
		query = `
		   WITH org_unit_attrs AS (
		     SELECT
		       COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) AS facility_uid,
		       MAX(NULLIF(h.level_of_care, '')) AS level_of_care,
		       MAX(NULLIF(h.ownership, '')) AS ownership
		     FROM hiv.organisation_unit h
		     WHERE COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) IS NOT NULL
		     GROUP BY 1
		   ),
		   selected_input AS (
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
		        (su.selected_level = '1' AND (h.country_uid = su.selected_uid OR (h.level = '1' AND h.org_unit_id = su.selected_uid)))
		        OR (su.selected_level = '6' AND (h.facility_uid = su.selected_uid OR (h.level = '6' AND h.org_unit_id = su.selected_uid)))
		        OR (su.selected_level = '2' AND h.region_uid = su.selected_uid)
		        OR (su.selected_level = '3' AND h.district_uid = su.selected_uid)
		        OR (su.selected_level = '5' AND (h.sub_county_uid = su.selected_uid OR (h.level = '5' AND h.org_unit_id = su.selected_uid)))
		      )
		     WHERE su.selected_level IN ('2', '3', '5', '6')
		       AND COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) IS NOT NULL
		       ` + orgUnitFilterClause + `
		   )
		   SELECT
		     x.org_unit_id,
		     x.org_unit_name,
		     x.data_element_id,
		     x."period",
		     x.value,
		     x."level",
		     x.level_of_care,
		     x.ownership,
		     x.dataelement
		   FROM (
		     SELECT
		       sf.selected_uid AS org_unit_id,
		       sf.selected_name AS org_unit_name,
		       hs.data_element_id,
		       hs."period",
		       SUM(
		         CASE
		           WHEN hs.value ~ '^\s*-?\d+(\.\d+)?\s*$' THEN TRIM(hs.value)::numeric
		           ELSE 0
		         END
		       )::bigint AS value,
		       sf.selected_level AS "level",
		       ` + levelOfCareExpr + ` AS level_of_care,
		       ` + ownershipExpr + ` AS ownership,
		       hs.dataelement
		     FROM selected_facilities sf
		     JOIN report.hmis_summary hs
		       ON hs.org_unit_id = sf.facility_uid
		     LEFT JOIN org_unit_attrs oa
		       ON oa.facility_uid = hs.org_unit_id
		     ` + whereClause + `
		     GROUP BY 1, 2, 3, 4, 6, 7, 8, 9
		    
		     UNION ALL

		     SELECT
		       su.selected_uid AS org_unit_id,
		       su.selected_name AS org_unit_name,
		       hs.data_element_id,
		       hs."period",
		       SUM(
		         CASE
		           WHEN hs.value ~ '^\s*-?\d+(\.\d+)?\s*$' THEN TRIM(hs.value)::numeric
		           ELSE 0
		         END
		       )::bigint AS value,
		       su.selected_level AS "level",
		       ` + levelOfCareExpr + ` AS level_of_care,
		       ` + ownershipExpr + ` AS ownership,
		       hs.dataelement
		     FROM selected_units su
		     JOIN hiv.organisation_unit h
		       ON (
		         h.country_uid = su.selected_uid
		         OR (h.level = '1' AND h.org_unit_id = su.selected_uid)
		       )
		     JOIN report.hmis_summary hs
		       ON hs.org_unit_id = COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id)
		     LEFT JOIN org_unit_attrs oa
		       ON oa.facility_uid = hs.org_unit_id
		     ` + countryWhereClause + `
		     GROUP BY 1, 2, 3, 4, 6, 7, 8, 9
		   ) x
		   ORDER BY
		     x."period",
		     x."level",
		     x.org_unit_name,
		     x.data_element_id
		`
	} else {
		aggregationLevel := r.resolveAggregationLevel(ctx, requestedLevel, req.OU)
		orgUnitExpr, facilityExpr, levelExpr, _, _, _ := aggregationExpressions(aggregationLevel)
		orgUnitFilterJoin := ""
		if orgUnitFilterClause != "" {
			orgUnitFilterJoin = `
	       JOIN (
	         SELECT DISTINCT COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) AS facility_uid
	         FROM hiv.organisation_unit h
	         WHERE COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) IS NOT NULL
	           ` + orgUnitFilterClause + `
	       ) ou_filter
	         ON ou_filter.facility_uid = hs.org_unit_id
	       `
		}

		query = `
		   SELECT 
		      ` + orgUnitExpr + ` AS org_unit_id,
		      ` + facilityExpr + ` AS org_unit_name,
		      hs.data_element_id,
		      hs."period",
		      SUM(
		        CASE
		          WHEN hs.value ~ '^\s*-?\d+(\.\d+)?\s*$' THEN TRIM(hs.value)::numeric
		          ELSE 0
		        END
		      )::bigint AS value,
		      ` + levelExpr + ` AS "level",
	          ` + levelOfCareExpr + ` AS level_of_care,
	          ` + ownershipExpr + ` AS ownership,
	          hs.dataelement
	       FROM report.hmis_summary hs
	       LEFT JOIN (
	         SELECT
	           COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) AS facility_uid,
	           MAX(NULLIF(h.level_of_care, '')) AS level_of_care,
	           MAX(NULLIF(h.ownership, '')) AS ownership
	         FROM hiv.organisation_unit h
	         WHERE COALESCE(NULLIF(h.facility_uid, ''), h.org_unit_id) IS NOT NULL
	         GROUP BY 1
	       ) oa
	         ON oa.facility_uid = hs.org_unit_id
	       ` + orgUnitFilterJoin + `
	       ` + whereClause + `
	       GROUP BY 1, 2, 3, 4, 6, 7, 8
	       ORDER BY 
	          4,
	          1,
	          3
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
			&row.OrgUnitName,
			&row.DataElementID,
			&row.Period,
			&row.Value,
			&row.Level,
			&row.LevelOfCare,
			&row.Ownership,
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
