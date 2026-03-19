-- name: CreateDisease :one
INSERT INTO diseases (
  name,
  short_name,
  code,
  category
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetDiseaseByID :one
SELECT *
FROM diseases
WHERE id = $1
LIMIT 1;

-- name: GetDiseaseByName :one
SELECT *
FROM diseases
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ListDiseases :many
SELECT *
FROM diseases
ORDER BY name ASC;

-- name: ListActiveDiseases :many
SELECT *
FROM diseases
WHERE is_active = TRUE
ORDER BY name ASC;

-- name: UpsertDisease :one
INSERT INTO diseases (
  name,
  short_name,
  code,
  category
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (name)
DO UPDATE SET
  short_name = COALESCE(EXCLUDED.short_name, diseases.short_name),
  code = COALESCE(EXCLUDED.code, diseases.code),
  category = COALESCE(EXCLUDED.category, diseases.category),
  updated_at = now()
RETURNING *;

-- name: SetDiseaseActiveState :one
UPDATE diseases
SET
  is_active = $2,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteDisease :exec
DELETE FROM diseases
WHERE id = $1;