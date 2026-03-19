-- name: CreateIndicator :one
INSERT INTO indicators (
  name,
  short_name,
  code,
  indicator_type
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetIndicatorByID :one
SELECT *
FROM indicators
WHERE id = $1
LIMIT 1;

-- name: GetIndicatorByName :one
SELECT *
FROM indicators
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ListIndicators :many
SELECT *
FROM indicators
ORDER BY name ASC;

-- name: ListActiveIndicators :many
SELECT *
FROM indicators
WHERE is_active = TRUE
ORDER BY name ASC;

-- name: UpsertIndicator :one
INSERT INTO indicators (
  name,
  short_name,
  code,
  indicator_type
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (name)
DO UPDATE SET
  short_name = COALESCE(EXCLUDED.short_name, indicators.short_name),
  code = COALESCE(EXCLUDED.code, indicators.code),
  indicator_type = COALESCE(EXCLUDED.indicator_type, indicators.indicator_type),
  updated_at = now()
RETURNING *;

-- name: SetIndicatorActiveState :one
UPDATE indicators
SET
  is_active = $2,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteIndicator :exec
DELETE FROM indicators
WHERE id = $1;