-- name: CreateFacility :one
INSERT INTO facilities (
  external_id,
  name,
  district_id,
  sub_county_id,
  region_id,
  facility_level,
  facility_type,
  dhis2_org_unit_id,
  latitude,
  longitude
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetFacilityByID :one
SELECT *
FROM facilities
WHERE id = $1
LIMIT 1;

-- name: GetFacilityByExternalID :one
SELECT *
FROM facilities
WHERE external_id = $1
LIMIT 1;

-- name: GetFacilityByNameAndDistrict :one
SELECT *
FROM facilities
WHERE LOWER(name) = LOWER($1)
  AND district_id = $2
LIMIT 1;

-- name: ListFacilities :many
SELECT
  f.*,
  d.name AS district_name,
  sc.name AS sub_county_name,
  r.name AS region_name
FROM facilities f
LEFT JOIN districts d ON d.id = f.district_id
LEFT JOIN sub_counties sc ON sc.id = f.sub_county_id
LEFT JOIN regions r ON r.id = f.region_id
ORDER BY f.name ASC;

-- name: ListFacilitiesByDistrict :many
SELECT *
FROM facilities
WHERE district_id = $1
ORDER BY name ASC;

-- name: UpsertFacilityByExternalID :one
INSERT INTO facilities (
  external_id,
  name,
  district_id,
  sub_county_id,
  region_id,
  facility_level,
  facility_type,
  dhis2_org_unit_id,
  latitude,
  longitude
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (external_id)
DO UPDATE SET
  name = EXCLUDED.name,
  district_id = COALESCE(EXCLUDED.district_id, facilities.district_id),
  sub_county_id = COALESCE(EXCLUDED.sub_county_id, facilities.sub_county_id),
  region_id = COALESCE(EXCLUDED.region_id, facilities.region_id),
  facility_level = COALESCE(EXCLUDED.facility_level, facilities.facility_level),
  facility_type = COALESCE(EXCLUDED.facility_type, facilities.facility_type),
  dhis2_org_unit_id = COALESCE(EXCLUDED.dhis2_org_unit_id, facilities.dhis2_org_unit_id),
  latitude = COALESCE(EXCLUDED.latitude, facilities.latitude),
  longitude = COALESCE(EXCLUDED.longitude, facilities.longitude),
  updated_at = now()
RETURNING *;

-- name: UpsertFacilityByNameDistrict :one
INSERT INTO facilities (
  external_id,
  name,
  district_id,
  sub_county_id,
  region_id,
  facility_level,
  facility_type,
  dhis2_org_unit_id,
  latitude,
  longitude
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (name, district_id)
DO UPDATE SET
  external_id = COALESCE(EXCLUDED.external_id, facilities.external_id),
  sub_county_id = COALESCE(EXCLUDED.sub_county_id, facilities.sub_county_id),
  region_id = COALESCE(EXCLUDED.region_id, facilities.region_id),
  facility_level = COALESCE(EXCLUDED.facility_level, facilities.facility_level),
  facility_type = COALESCE(EXCLUDED.facility_type, facilities.facility_type),
  dhis2_org_unit_id = COALESCE(EXCLUDED.dhis2_org_unit_id, facilities.dhis2_org_unit_id),
  latitude = COALESCE(EXCLUDED.latitude, facilities.latitude),
  longitude = COALESCE(EXCLUDED.longitude, facilities.longitude),
  updated_at = now()
RETURNING *;

-- name: DeleteFacility :exec
DELETE FROM facilities
WHERE id = $1;