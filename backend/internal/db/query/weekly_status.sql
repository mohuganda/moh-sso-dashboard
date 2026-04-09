-- name: CreateWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;


-- name: GetWeeklyStatusByID :one
SELECT *
FROM weekly_status
WHERE id = $1
LIMIT 1;


-- name: DeleteWeeklyStatus :exec
DELETE FROM weekly_status
WHERE id = $1;


-- name: ListWeeklyStatuses :many
SELECT *
FROM weekly_status
WHERE
  ($1::uuid IS NULL OR epi_week_id = $1)
  AND ($2::uuid IS NULL OR region_id = $2)
  AND ($3::uuid IS NULL OR district_id = $3)
  AND ($4::uuid IS NULL OR sub_county_id = $4)
  AND ($5::uuid IS NULL OR disease_id = $5)
  AND ($6::uuid IS NULL OR indicator_id = $6)
  AND ($7::risk_level IS NULL OR status = $7)
ORDER BY created_at DESC;


-- name: ListWeeklyStatusesByWeek :many
SELECT *
FROM weekly_status
WHERE epi_week_id = $1
ORDER BY created_at DESC;


-- name: ListWeeklyStatusesByRegion :many
SELECT *
FROM weekly_status
WHERE region_id = $1
ORDER BY created_at DESC;


-- name: ListWeeklyStatusesByDistrict :many
SELECT *
FROM weekly_status
WHERE district_id = $1
ORDER BY created_at DESC;


-- name: ListWeeklyStatusesBySubCounty :many
SELECT *
FROM weekly_status
WHERE sub_county_id = $1
ORDER BY created_at DESC;


-- name: ListNationalWeeklyStatusesByWeek :many
SELECT *
FROM weekly_status
WHERE epi_week_id = $1
  AND region_id IS NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
ORDER BY created_at DESC;


-- name: ListRegionWeeklyStatusesByWeek :many
SELECT *
FROM weekly_status
WHERE epi_week_id = $1
  AND region_id IS NOT NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
ORDER BY created_at DESC;


-- name: ListDistrictWeeklyStatusesByWeek :many
SELECT *
FROM weekly_status
WHERE epi_week_id = $1
  AND region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NULL
ORDER BY created_at DESC;


-- name: ListSubCountyWeeklyStatusesByWeek :many
SELECT *
FROM weekly_status
WHERE epi_week_id = $1
  AND region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NOT NULL
ORDER BY created_at DESC;

-- name: UpsertNationalDiseaseWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  NULL,
  NULL,
  NULL,
  $1,
  NULL,
  $2,
  $3,
  $4
)
ON CONFLICT (disease_id, epi_week_id)
WHERE
  region_id IS NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND disease_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: UpsertNationalIndicatorWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  NULL,
  NULL,
  NULL,
  NULL,
  $1,
  $2,
  $3,
  $4
)
ON CONFLICT (indicator_id, epi_week_id)
WHERE
  region_id IS NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND indicator_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;


-- name: UpsertRegionDiseaseWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1,
  NULL,
  NULL,
  $2,
  NULL,
  $3,
  $4,
  $5
)
ON CONFLICT (region_id, disease_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND disease_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;


-- name: UpsertRegionIndicatorWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1,
  NULL,
  NULL,
  NULL,
  $2,
  $3,
  $4,
  $5
)
ON CONFLICT (region_id, indicator_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND indicator_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;


-- name: UpsertDistrictDiseaseWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1,
  $2,
  NULL,
  $3,
  NULL,
  $4,
  $5,
  $6
)
ON CONFLICT (district_id, disease_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NULL
  AND disease_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;


-- name: UpsertDistrictIndicatorWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1,
  $2,
  NULL,
  NULL,
  $3,
  $4,
  $5,
  $6
)
ON CONFLICT (district_id, indicator_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NULL
  AND indicator_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;


-- name: UpsertSubCountyDiseaseWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1,
  $2,
  $3,
  $4,
  NULL,
  $5,
  $6,
  $7
)
ON CONFLICT (sub_county_id, disease_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NOT NULL
  AND disease_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;


-- name: UpsertSubCountyIndicatorWeeklyStatus :one
INSERT INTO weekly_status (
  region_id,
  district_id,
  sub_county_id,
  disease_id,
  indicator_id,
  epi_week_id,
  status,
  source_name
) VALUES (
  $1,
  $2,
  $3,
  NULL,
  $4,
  $5,
  $6,
  $7
)
ON CONFLICT (sub_county_id, indicator_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NOT NULL
  AND indicator_id IS NOT NULL
DO UPDATE SET
  status = EXCLUDED.status,
  source_name = COALESCE(EXCLUDED.source_name, weekly_status.source_name),
  imported_at = now()
RETURNING *;

-- name: ListWeeklyStatusesDetailed :many
SELECT
  ws.id,
  ws.region_id,
  ws.district_id,
  ws.sub_county_id,
  ws.disease_id,
  ws.indicator_id,
  ws.epi_week_id,
  ws.status,
  ws.source_name,
  ws.imported_at,
  ws.created_at,
  r.name AS region_name,
  d.name AS district_name,
  sc.name AS sub_county_name,
  dis.name AS disease_name,
  i.name AS indicator_name
FROM weekly_status ws
LEFT JOIN regions r ON ws.region_id = r.id
LEFT JOIN districts d ON ws.district_id = d.id
LEFT JOIN sub_counties sc ON ws.sub_county_id = sc.id
LEFT JOIN diseases dis ON ws.disease_id = dis.id
LEFT JOIN indicators i ON ws.indicator_id = i.id
WHERE
  ($1::uuid IS NULL OR ws.epi_week_id = $1)
  AND ($2::uuid IS NULL OR ws.region_id = $2)
  AND ($3::uuid IS NULL OR ws.district_id = $3)
  AND ($4::uuid IS NULL OR ws.sub_county_id = $4)
  AND ($5::uuid IS NULL OR ws.disease_id = $5)
  AND ($6::uuid IS NULL OR ws.indicator_id = $6)
  AND ($7::risk_level IS NULL OR ws.status = $7)
ORDER BY ws.created_at DESC;