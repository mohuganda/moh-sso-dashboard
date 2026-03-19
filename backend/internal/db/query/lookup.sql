-- name: ResolveDiseaseByName :one
SELECT *
FROM diseases
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ResolveIndicatorByName :one
SELECT *
FROM indicators
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ResolveWeekByYearWeek :one
SELECT *
FROM epi_weeks
WHERE epi_year = $1
  AND epi_week = $2
LIMIT 1;

-- name: ResolveDistrictByName :one
SELECT *
FROM districts
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ResolveRegionByName :one
SELECT *
FROM regions
WHERE LOWER(name) = LOWER($1)
LIMIT 1;

-- name: ResolveSubCountyByNameAndDistrict :one
SELECT *
FROM sub_counties
WHERE LOWER(name) = LOWER($1)
  AND district_id = $2
LIMIT 1;

-- name: ResolveFacilityByExternalID :one
SELECT *
FROM facilities
WHERE external_id = $1
LIMIT 1;

-- name: ResolveFacilityByNameAndDistrict :one
SELECT *
FROM facilities
WHERE LOWER(name) = LOWER($1)
  AND district_id = $2
LIMIT 1;