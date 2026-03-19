-- name: CreateDistrictWeeklyDiseaseStatus :one
INSERT INTO district_weekly_status (
  district_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, $2, NULL, $3, $4, $5
)
RETURNING *;

-- name: UpsertDistrictWeeklyDiseaseStatus :one
INSERT INTO district_weekly_status (
  district_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, $2, NULL, $3, $4, $5
)
ON CONFLICT (district_id, disease_id, epi_week_id)
WHERE disease_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, district_weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: CreateDistrictWeeklyIndicatorStatus :one
INSERT INTO district_weekly_status (
  district_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, NULL, $2, $3, $4, $5
)
RETURNING *;

-- name: UpsertDistrictWeeklyIndicatorStatus :one
INSERT INTO district_weekly_status (
  district_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, NULL, $2, $3, $4, $5
)
ON CONFLICT (district_id, indicator_id, epi_week_id)
WHERE indicator_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, district_weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: ListDistrictWeeklyDiseaseStatusesByWeek :many
SELECT
  s.*,
  d.name AS district_name,
  dis.name AS disease_name
FROM district_weekly_status s
JOIN districts d ON d.id = s.district_id
JOIN diseases dis ON dis.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.disease_id IS NOT NULL
ORDER BY d.name ASC, dis.name ASC;

-- name: ListDistrictWeeklyIndicatorStatusesByWeek :many
SELECT
  s.*,
  d.name AS district_name,
  i.name AS indicator_name
FROM district_weekly_status s
JOIN districts d ON d.id = s.district_id
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
  AND s.indicator_id IS NOT NULL
ORDER BY d.name ASC, i.name ASC;

-- name: ListDistrictStatusesByDistrictAndWeek :many
SELECT
  s.id,
  s.district_id,
  s.disease_id,
  s.indicator_id,
  s.epi_week_id,
  s.status,
  s.source_name,
  s.imported_at,
  s.created_at,
  CASE
    WHEN s.disease_id IS NOT NULL THEN 'DISEASE'
    ELSE 'INDICATOR'
  END AS subject_type,
  COALESCE(dis.name, i.name) AS subject_name
FROM district_weekly_status s
LEFT JOIN diseases dis ON dis.id = s.disease_id
LEFT JOIN indicators i ON i.id = s.indicator_id
WHERE s.district_id = $1
  AND s.epi_week_id = $2
ORDER BY
  CASE s.status
    WHEN 'MAROON' THEN 1
    WHEN 'RED' THEN 2
    WHEN 'YELLOW' THEN 3
    WHEN 'GREEN' THEN 4
  END,
  subject_name ASC;

-- name: DeleteDistrictStatusesByWeek :exec
DELETE FROM district_weekly_status
WHERE epi_week_id = $1;