-- name: CreateStorageLocation :one
INSERT INTO storage_locations (
    id,
    code,
    name,
    provider,
    base_uri,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;


-- name: GetStorageLocationByID :one
SELECT *
FROM storage_locations
WHERE id = $1;


-- name: GetStorageLocationByCode :one
SELECT *
FROM storage_locations
WHERE code = $1
  AND is_active = TRUE;


-- name: ListActiveStorageLocations :many
SELECT *
FROM storage_locations
WHERE is_active = TRUE
ORDER BY created_at DESC;


-- name: UpdateStorageLocation :one
UPDATE storage_locations
SET name = $2,
    provider = $3,
    base_uri = $4,
    is_active = $5
WHERE id = $1
RETURNING *;


-- name: DeleteStorageLocation :exec
DELETE FROM storage_locations
WHERE id = $1;