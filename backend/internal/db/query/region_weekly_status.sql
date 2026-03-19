-- name: CreateRegionWeeklyDiseaseStatus :one
INSERT INTO region_weekly_status (
  region_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, $2, NULL, $3, $4, $5
)
RETURNING *;

-- name: UpsertRegionWeeklyDiseaseStatus :one
INSERT INTO region_weekly_status (
  region_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, $2, NULL, $3, $4, $5
)
ON CONFLICT (region_id, disease_id, epi_week_id)
WHERE disease_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, region_weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: CreateRegionWeeklyIndicatorStatus :one
INSERT INTO region_weekly_status (
  region_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, NULL, $2, $3, $4, $5
)
RETURNING *;

-- name: UpsertRegionWeeklyIndicatorStatus :one
INSERT INTO region_weekly_status (
  region_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, NULL, $2, $3, $4, $5
)
ON CONFLICT (region_id, indicator_id, epi_week_id)
WHERE indicator_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, region_weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: ListRegionWeeklyDiseaseStatusesByWeek :many
SELECT
  s.*,
  r.name AS region_name,
  d.name AS disease_name
FROM region_weekly_status s
JOIN regions r ON r.id = s.region_id
JOIN diseases d ON d.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.disease_id IS NOT NULL
ORDER BY r.name ASC, d.name ASC;

-- name: ListRegionWeeklyIndicatorStatusesByWeek :many
SELECT
  s.*,
  r.name AS region_name,
  i.name AS indicator_name
FROM region_weekly_status s
JOIN regions r ON r.id = s.region_id
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
  AND s.indicator_id IS NOT NULL
ORDER BY r.name ASC, i.name ASC;

-- name: DeleteRegionStatusesByWeek :exec
DELETE FROM region_weekly_status
WHERE epi_week_id = $1;