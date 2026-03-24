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
SELECT id, name, code, district_id, created_at, updated_at
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