-- name: CreateRegion :one
INSERT INTO regions (
  name,
  code
) VALUES (
  $1,
  $2
)
RETURNING *;

-- name: GetRegionByID :one
SELECT *
FROM regions
WHERE id = $1
LIMIT 1;

-- name: GetRegionByName :one
SELECT *
FROM regions
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ListRegions :many
SELECT *
FROM regions
ORDER BY name ASC;

-- name: ListRegionsInHealthContext :many
SELECT *
FROM regions
WHERE health_context_alias_related_to_scope(
  'surveillance-region',
  id::text,
  sqlc.arg('health_context_id')::uuid,
  sqlc.arg('include_descendants')::boolean
)
ORDER BY name ASC;

-- name: UpsertRegion :one
INSERT INTO regions (
  name,
  code
) VALUES (
  $1,
  $2
)
ON CONFLICT (name)
DO UPDATE SET
  code = COALESCE(EXCLUDED.code, regions.code),
  updated_at = now()
RETURNING *;

-- name: DeleteRegion :exec
DELETE FROM regions
WHERE id = $1;
