-- name: CreateFacilityWeeklyDiseaseMetric :one
INSERT INTO facility_weekly_metrics (
  source_record_id,
  facility_id,
  disease_id,
  indicator_id,
  epi_week_id,
  metric_value,
  source_name
) VALUES (
  $1, $2, $3, NULL, $4, $5, $6
)
RETURNING *;

-- name: UpsertFacilityWeeklyDiseaseMetric :one
INSERT INTO facility_weekly_metrics (
  source_record_id,
  facility_id,
  disease_id,
  indicator_id,
  epi_week_id,
  metric_value,
  source_name
) VALUES (
  $1, $2, $3, NULL, $4, $5, $6
)
ON CONFLICT (facility_id, disease_id, epi_week_id)
WHERE disease_id IS NOT NULL
DO UPDATE SET
  source_record_id = COALESCE(EXCLUDED.source_record_id, facility_weekly_metrics.source_record_id),
  metric_value = EXCLUDED.metric_value,
  source_name = COALESCE(EXCLUDED.source_name, facility_weekly_metrics.source_name),
  imported_at = now()
RETURNING *;

-- name: CreateFacilityWeeklyIndicatorMetric :one
INSERT INTO facility_weekly_metrics (
  source_record_id,
  facility_id,
  disease_id,
  indicator_id,
  epi_week_id,
  metric_value,
  source_name
) VALUES (
  $1, $2, NULL, $3, $4, $5, $6
)
RETURNING *;

-- name: UpsertFacilityWeeklyIndicatorMetric :one
INSERT INTO facility_weekly_metrics (
  source_record_id,
  facility_id,
  disease_id,
  indicator_id,
  epi_week_id,
  metric_value,
  source_name
) VALUES (
  $1, $2, NULL, $3, $4, $5, $6
)
ON CONFLICT (facility_id, indicator_id, epi_week_id)
WHERE indicator_id IS NOT NULL
DO UPDATE SET
  source_record_id = COALESCE(EXCLUDED.source_record_id, facility_weekly_metrics.source_record_id),
  metric_value = EXCLUDED.metric_value,
  source_name = COALESCE(EXCLUDED.source_name, facility_weekly_metrics.source_name),
  imported_at = now()
RETURNING *;

-- name: GetFacilityWeeklyMetricByID :one
SELECT *
FROM facility_weekly_metrics
WHERE id = $1
LIMIT 1;

-- name: GetFacilityMetricBySourceRecordID :one
SELECT *
FROM facility_weekly_metrics
WHERE source_record_id = $1
LIMIT 1;

-- name: ListFacilityWeeklyDiseaseMetricsByWeek :many
SELECT
  m.*,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  d.name AS disease_name,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN diseases d ON d.id = m.disease_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.epi_week_id = $1
  AND m.disease_id IS NOT NULL
ORDER BY r.name ASC NULLS LAST,
         dct.name ASC NULLS LAST,
         sc.name ASC NULLS LAST,
         f.name ASC,
         d.name ASC;

-- name: ListFacilityWeeklyDiseaseMetricsByWeekInHealthContext :many
SELECT
  m.*,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  d.name AS disease_name,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN diseases d ON d.id = m.disease_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.epi_week_id = sqlc.arg('epi_week_id')
  AND m.disease_id IS NOT NULL
  AND health_context_alias_in_scope(
    'surveillance-facility',
    f.id::text,
    sqlc.arg('health_context_id')::uuid,
    sqlc.arg('include_descendants')::boolean
  )
ORDER BY r.name ASC NULLS LAST,
         dct.name ASC NULLS LAST,
         sc.name ASC NULLS LAST,
         f.name ASC,
         d.name ASC;

-- name: ListFacilityWeeklyIndicatorMetricsByWeek :many
SELECT
  m.*,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  i.name AS indicator_name,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN indicators i ON i.id = m.indicator_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.epi_week_id = $1
  AND m.indicator_id IS NOT NULL
ORDER BY r.name ASC NULLS LAST,
         dct.name ASC NULLS LAST,
         sc.name ASC NULLS LAST,
         f.name ASC,
         i.name ASC;

-- name: ListFacilityMetricsByFacility :many
SELECT
  m.id,
  m.source_record_id,
  m.facility_id,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  m.disease_id,
  m.indicator_id,
  m.epi_week_id,
  m.metric_value,
  m.source_name,
  m.imported_at,
  m.created_at,
  ew.epi_year,
  ew.epi_week,
  CASE
    WHEN m.disease_id IS NOT NULL THEN 'DISEASE'
    ELSE 'INDICATOR'
  END AS subject_type,
  COALESCE(d.name, i.name) AS subject_name
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
LEFT JOIN diseases d ON d.id = m.disease_id
LEFT JOIN indicators i ON i.id = m.indicator_id
WHERE m.facility_id = $1
ORDER BY ew.epi_year DESC, ew.epi_week DESC, subject_name ASC;

-- name: DeleteFacilityMetricsByWeek :exec
DELETE FROM facility_weekly_metrics
WHERE epi_week_id = $1;

-- name: ListFacilityDiseaseMetricsTrend :many
SELECT
  m.id,
  m.source_record_id,
  m.facility_id,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  m.disease_id,
  d.name AS disease_name,
  m.epi_week_id,
  m.metric_value,
  m.source_name,
  m.imported_at,
  m.created_at,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN diseases d ON d.id = m.disease_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.facility_id = $1
  AND m.disease_id = $2
ORDER BY ew.epi_year ASC, ew.epi_week ASC;

-- name: ListFacilityIndicatorMetricsTrend :many
SELECT
  m.id,
  m.source_record_id,
  m.facility_id,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  m.indicator_id,
  i.name AS indicator_name,
  m.epi_week_id,
  m.metric_value,
  m.source_name,
  m.imported_at,
  m.created_at,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN indicators i ON i.id = m.indicator_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.facility_id = $1
  AND m.indicator_id = $2
ORDER BY ew.epi_year ASC, ew.epi_week ASC;


-- name: ListDiseaseWeeklyTrendAggregated :many
SELECT
  ew.id AS epi_week_id,
  ew.epi_year,
  ew.epi_week,
  COALESCE(SUM(m.metric_value), 0)::bigint AS total_cases
FROM epi_weeks ew
LEFT JOIN facility_weekly_metrics m
  ON m.epi_week_id = ew.id
  AND m.disease_id = sqlc.narg('disease_id')
LEFT JOIN facilities f
  ON f.id = m.facility_id
LEFT JOIN sub_counties sc
  ON sc.id = f.sub_county_id
LEFT JOIN districts dct
  ON dct.id = sc.district_id
LEFT JOIN regions r
  ON r.id = dct.region_id
WHERE ew.epi_year = sqlc.arg('epi_year')
  AND (
    sqlc.narg('region_id')::uuid IS NULL
    OR r.id = sqlc.narg('region_id')::uuid
  )
  AND (
    sqlc.narg('district_id')::uuid IS NULL
    OR dct.id = sqlc.narg('district_id')::uuid
  )
GROUP BY ew.id, ew.epi_year, ew.epi_week
ORDER BY ew.epi_year ASC, ew.epi_week ASC;

-- name: ListDiseaseWeeklyTrendAggregatedInHealthContext :many
SELECT
  ew.id AS epi_week_id,
  ew.epi_year,
  ew.epi_week,
  COALESCE(SUM(m.metric_value), 0)::bigint AS total_cases
FROM epi_weeks ew
LEFT JOIN facility_weekly_metrics m
  ON m.epi_week_id = ew.id
  AND m.disease_id = sqlc.narg('disease_id')
  AND health_context_alias_in_scope(
    'surveillance-facility',
    m.facility_id::text,
    sqlc.arg('health_context_id')::uuid,
    sqlc.arg('include_descendants')::boolean
  )
LEFT JOIN facilities f
  ON f.id = m.facility_id
LEFT JOIN sub_counties sc
  ON sc.id = f.sub_county_id
LEFT JOIN districts dct
  ON dct.id = sc.district_id
LEFT JOIN regions r
  ON r.id = dct.region_id
WHERE ew.epi_year = sqlc.arg('epi_year')
  AND (
    sqlc.narg('region_id')::uuid IS NULL
    OR r.id = sqlc.narg('region_id')::uuid
  )
  AND (
    sqlc.narg('district_id')::uuid IS NULL
    OR dct.id = sqlc.narg('district_id')::uuid
  )
GROUP BY ew.id, ew.epi_year, ew.epi_week
ORDER BY ew.epi_year ASC, ew.epi_week ASC;

-- name: ListFacilityWeeklyDiseaseMetricsByWeekAndDisease :many
SELECT
  m.*,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  d.name AS disease_name,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN diseases d ON d.id = m.disease_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.epi_week_id = sqlc.arg('epi_week_id')
  AND m.disease_id = sqlc.arg('disease_id')
ORDER BY r.name ASC NULLS LAST,
         dct.name ASC NULLS LAST,
         sc.name ASC NULLS LAST,
         f.name ASC,
         d.name ASC;

-- name: ListFacilityWeeklyDiseaseMetricsByWeekAndDiseaseInHealthContext :many
SELECT
  m.*,
  f.name AS facility_name,
  sc.id AS sub_county_id,
  sc.name AS sub_county_name,
  dct.id AS district_id,
  dct.name AS district_name,
  r.id AS region_id,
  r.name AS region_name,
  d.name AS disease_name,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN districts dct ON dct.id = sc.district_id
LEFT JOIN regions r ON r.id = dct.region_id
JOIN diseases d ON d.id = m.disease_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.epi_week_id = sqlc.arg('epi_week_id')
  AND m.disease_id = sqlc.arg('disease_id')
  AND health_context_alias_in_scope(
    'surveillance-facility',
    f.id::text,
    sqlc.arg('health_context_id')::uuid,
    sqlc.arg('include_descendants')::boolean
  )
ORDER BY r.name ASC NULLS LAST,
         dct.name ASC NULLS LAST,
         sc.name ASC NULLS LAST,
         f.name ASC,
         d.name ASC;
