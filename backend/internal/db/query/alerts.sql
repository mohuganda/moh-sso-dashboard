-- name: CreateAlert :one
INSERT INTO alerts (
  external_id,
  disease_id,
  district_id,
  epi_week_id,
  occurred_on,
  created_on,
  narrative,
  submitted_by,
  status,
  source_name
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpsertAlertByExternalID :one
INSERT INTO alerts (
  external_id,
  disease_id,
  district_id,
  epi_week_id,
  occurred_on,
  created_on,
  narrative,
  submitted_by,
  status,
  source_name
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (external_id)
DO UPDATE SET
  disease_id = EXCLUDED.disease_id,
  district_id = EXCLUDED.district_id,
  epi_week_id = EXCLUDED.epi_week_id,
  occurred_on = EXCLUDED.occurred_on,
  created_on = EXCLUDED.created_on,
  narrative = EXCLUDED.narrative,
  submitted_by = EXCLUDED.submitted_by,
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, alerts.source_name),
  imported_at = now(),
  updated_at = now()
RETURNING *;

-- name: GetAlertByID :one
SELECT *
FROM alerts
WHERE id = $1
LIMIT 1;

-- name: ListAlerts :many
SELECT
  a.id,
  a.external_id,
  a.disease_id,
  ds.name AS disease_name,
  a.district_id,
  d.name AS district_name,
  d.region_id,
  a.epi_week_id,
  a.occurred_on,
  a.created_on,
  a.narrative,
  a.submitted_by,
  a.status,
  a.source_name,
  a.imported_at,
  a.created_at,
  a.updated_at
FROM alerts a
LEFT JOIN diseases ds ON ds.id = a.disease_id
LEFT JOIN districts d ON d.id = a.district_id
WHERE
  (sqlc.narg(epi_week_id)::uuid IS NULL OR a.epi_week_id = sqlc.narg(epi_week_id)::uuid)
  AND (sqlc.narg(disease_id)::uuid IS NULL OR a.disease_id = sqlc.narg(disease_id)::uuid)
  AND (sqlc.narg(district_id)::uuid IS NULL OR a.district_id = sqlc.narg(district_id)::uuid)
  AND (sqlc.narg(region_id)::uuid IS NULL OR d.region_id = sqlc.narg(region_id)::uuid)
ORDER BY a.created_at DESC;

-- name: ListAlertsInHealthContext :many
SELECT
  a.id,
  a.external_id,
  a.disease_id,
  ds.name AS disease_name,
  a.district_id,
  d.name AS district_name,
  d.region_id,
  a.epi_week_id,
  a.occurred_on,
  a.created_on,
  a.narrative,
  a.submitted_by,
  a.status,
  a.source_name,
  a.imported_at,
  a.created_at,
  a.updated_at
FROM alerts a
LEFT JOIN diseases ds ON ds.id = a.disease_id
LEFT JOIN districts d ON d.id = a.district_id
WHERE
  (sqlc.narg(epi_week_id)::uuid IS NULL OR a.epi_week_id = sqlc.narg(epi_week_id)::uuid)
  AND (sqlc.narg(disease_id)::uuid IS NULL OR a.disease_id = sqlc.narg(disease_id)::uuid)
  AND (sqlc.narg(district_id)::uuid IS NULL OR a.district_id = sqlc.narg(district_id)::uuid)
  AND (sqlc.narg(region_id)::uuid IS NULL OR d.region_id = sqlc.narg(region_id)::uuid)
  AND a.district_id IS NOT NULL
  AND health_context_alias_in_scope(
    'surveillance-district', a.district_id::text, sqlc.arg(health_context_id)::uuid,
    sqlc.arg(include_descendants)::boolean
  )
ORDER BY a.created_at DESC;

-- name: ListAlertsByDisease :many
SELECT
  a.*,
  dist.name AS district_name,
  ew.epi_year,
  ew.epi_week
FROM alerts a
LEFT JOIN districts dist ON dist.id = a.district_id
LEFT JOIN epi_weeks ew ON ew.id = a.epi_week_id
WHERE a.disease_id = $1
ORDER BY a.created_on DESC, a.created_at DESC;

-- name: ListAlertsByDistrict :many
SELECT
  a.*,
  d.name AS disease_name,
  ew.epi_year,
  ew.epi_week
FROM alerts a
JOIN diseases d ON d.id = a.disease_id
LEFT JOIN epi_weeks ew ON ew.id = a.epi_week_id
WHERE a.district_id = $1
ORDER BY a.created_on DESC, a.created_at DESC;

-- name: ListAlertsByWeek :many
SELECT
  a.*,
  d.name AS disease_name,
  dist.name AS district_name
FROM alerts a
JOIN diseases d ON d.id = a.disease_id
LEFT JOIN districts dist ON dist.id = a.district_id
WHERE a.epi_week_id = $1
ORDER BY a.created_on DESC, a.created_at DESC;

-- name: UpdateAlertStatus :one
UPDATE alerts
SET
  status = $2,
  updated_at = now()
WHERE id = $1
RETURNING *;
