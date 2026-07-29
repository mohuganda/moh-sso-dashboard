-- name: CreateSubCounty :one
INSERT INTO sub_counties (
  name,
  district_id,
  code
) VALUES (
  $1,
  $2,
  $3
)
RETURNING *;

-- name: GetSubCountyByID :one
SELECT *
FROM sub_counties
WHERE id = $1
LIMIT 1;

-- name: GetSubcountyByName :one
SELECT *
FROM sub_counties
WHERE LOWER(name) = LOWER(sqlc.arg(name))
LIMIT 1;

-- name: GetSubCountyByNameAndDistrict :one
SELECT *
FROM sub_counties
WHERE LOWER(name) = LOWER($1)
  AND district_id = $2
LIMIT 1;

-- name: ListSubCountiesByDistrict :many
SELECT *
FROM sub_counties
WHERE district_id = $1
ORDER BY name ASC;

-- name: ListSubCountiesByDistrictInHealthContext :many
SELECT *
FROM sub_counties
WHERE district_id = sqlc.arg('district_id')
  AND health_context_alias_related_to_scope(
    'surveillance-sub-county',
    id::text,
    sqlc.arg('health_context_id')::uuid,
    sqlc.arg('include_descendants')::boolean
  )
ORDER BY name ASC;

-- name: UpsertSubCounty :one
INSERT INTO sub_counties (
  name,
  district_id,
  code
) VALUES (
  $1,
  $2,
  $3
)
ON CONFLICT (name, district_id)
DO UPDATE SET
  code = COALESCE(EXCLUDED.code, sub_counties.code),
  updated_at = now()
RETURNING *;

-- name: DeleteSubCounty :exec
DELETE FROM sub_counties
WHERE id = $1;
