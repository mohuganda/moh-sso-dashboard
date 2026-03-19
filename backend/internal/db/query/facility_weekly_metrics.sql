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
  d.name AS disease_name,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
JOIN diseases d ON d.id = m.disease_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.epi_week_id = $1
  AND m.disease_id IS NOT NULL
ORDER BY f.name ASC, d.name ASC;

-- name: ListFacilityWeeklyIndicatorMetricsByWeek :many
SELECT
  m.*,
  f.name AS facility_name,
  i.name AS indicator_name,
  ew.epi_year,
  ew.epi_week
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
JOIN indicators i ON i.id = m.indicator_id
JOIN epi_weeks ew ON ew.id = m.epi_week_id
WHERE m.epi_week_id = $1
  AND m.indicator_id IS NOT NULL
ORDER BY f.name ASC, i.name ASC;

-- name: ListFacilityMetricsByFacility :many
SELECT
  m.id,
  m.source_record_id,
  m.facility_id,
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
JOIN epi_weeks ew ON ew.id = m.epi_week_id
LEFT JOIN diseases d ON d.id = m.disease_id
LEFT JOIN indicators i ON i.id = m.indicator_id
WHERE m.facility_id = $1
ORDER BY ew.epi_year DESC, ew.epi_week DESC, subject_name ASC;

-- name: DeleteFacilityMetricsByWeek :exec
DELETE FROM facility_weekly_metrics
WHERE epi_week_id = $1;