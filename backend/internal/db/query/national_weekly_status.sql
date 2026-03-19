-- name: CreateNationalWeeklyDiseaseStatus :one
INSERT INTO national_weekly_status (
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, NULL, $2, $3, $4
)
RETURNING *;

-- name: UpsertNationalWeeklyDiseaseStatus :one
INSERT INTO national_weekly_status (
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, NULL, $2, $3, $4
)
ON CONFLICT (disease_id, epi_week_id)
WHERE disease_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, national_weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: CreateNationalWeeklyIndicatorStatus :one
INSERT INTO national_weekly_status (
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  NULL, $1, $2, $3, $4
)
RETURNING *;

-- name: UpsertNationalWeeklyIndicatorStatus :one
INSERT INTO national_weekly_status (
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  NULL, $1, $2, $3, $4
)
ON CONFLICT (indicator_id, epi_week_id)
WHERE indicator_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, national_weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: ListNationalWeeklyDiseaseStatusesByWeek :many
SELECT
  s.*,
  d.name AS disease_name
FROM national_weekly_status s
JOIN diseases d ON d.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.disease_id IS NOT NULL
ORDER BY
  CASE s.status
    WHEN 'MAROON' THEN 1
    WHEN 'RED' THEN 2
    WHEN 'YELLOW' THEN 3
    WHEN 'GREEN' THEN 4
  END,
  d.name ASC;

-- name: ListNationalWeeklyIndicatorStatusesByWeek :many
SELECT
  s.*,
  i.name AS indicator_name
FROM national_weekly_status s
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
  AND s.indicator_id IS NOT NULL
ORDER BY
  CASE s.status
    WHEN 'MAROON' THEN 1
    WHEN 'RED' THEN 2
    WHEN 'YELLOW' THEN 3
    WHEN 'GREEN' THEN 4
  END,
  i.name ASC;

-- name: DeleteNationalStatusesByWeek :exec
DELETE FROM national_weekly_status
WHERE epi_week_id = $1;