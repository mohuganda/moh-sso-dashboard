-- name: CreateDistrict :one
INSERT INTO districts (
  name,
  region_id,
  code
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: GetDistrictByID :one
SELECT *
FROM districts
WHERE id = $1
LIMIT 1;

-- name: GetDistrictByName :one
SELECT *
FROM districts
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ListDistricts :many
SELECT d.*, r.name AS region_name
FROM districts d
LEFT JOIN regions r ON r.id = d.region_id
ORDER BY d.name ASC;

-- name: ListDistrictsInHealthContext :many
SELECT d.*, r.name AS region_name
FROM districts d
LEFT JOIN regions r ON r.id = d.region_id
WHERE health_context_alias_related_to_scope(
  'surveillance-district',
  d.id::text,
  sqlc.arg('health_context_id')::uuid,
  sqlc.arg('include_descendants')::boolean
)
ORDER BY d.name ASC;

-- name: ListDistrictsByRegion :many
SELECT *
FROM districts
WHERE region_id = $1
ORDER BY name ASC;

-- name: ListDistrictsByRegionInHealthContext :many
SELECT *
FROM districts
WHERE region_id = sqlc.arg('region_id')
  AND health_context_alias_related_to_scope(
    'surveillance-district',
    id::text,
    sqlc.arg('health_context_id')::uuid,
    sqlc.arg('include_descendants')::boolean
  )
ORDER BY name ASC;

-- name: UpsertDistrict :one
INSERT INTO districts (
  name,
  region_id,
  code
) VALUES (
  $1,
  $2,
  $3
)
ON CONFLICT (name)
DO UPDATE SET
  region_id = COALESCE(EXCLUDED.region_id, districts.region_id),
  code = COALESCE(EXCLUDED.code, districts.code),
  updated_at = now()
RETURNING *;

-- name: DeleteDistrict :exec
DELETE FROM districts
WHERE id = $1;
