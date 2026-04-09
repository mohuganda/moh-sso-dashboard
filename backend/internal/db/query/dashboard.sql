-- name: GetNationalWeeklyStatusSummary :many
SELECT
  status,
  COUNT(*)::bigint AS total_items
FROM weekly_status
WHERE epi_week_id = $1
  AND region_id IS NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
GROUP BY status
ORDER BY
  CASE status
    WHEN 'MAROON' THEN 1
    WHEN 'RED' THEN 2
    WHEN 'YELLOW' THEN 3
    WHEN 'GREEN' THEN 4
  END;


-- name: ListNationalWeeklySubjectsByWeek :many
SELECT
  s.id,
  s.epi_week_id,
  s.status,
  s.source_name,
  s.imported_at,
  s.created_at,
  d.id AS subject_id,
  d.name AS subject_name,
  'DISEASE' AS subject_type
FROM weekly_status s
JOIN diseases d ON d.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.region_id IS NULL
  AND s.district_id IS NULL
  AND s.sub_county_id IS NULL
  AND s.disease_id IS NOT NULL

UNION ALL

SELECT
  s.id,
  s.epi_week_id,
  s.status,
  s.source_name,
  s.imported_at,
  s.created_at,
  i.id AS subject_id,
  i.name AS subject_name,
  'INDICATOR' AS subject_type
FROM weekly_status s
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
  AND s.region_id IS NULL
  AND s.district_id IS NULL
  AND s.sub_county_id IS NULL
  AND s.indicator_id IS NOT NULL

ORDER BY status, subject_name;


-- name: ListDistrictWeeklySubjectsByWeek :many
SELECT
  s.id,
  s.region_id,
  r.name AS region_name,
  s.district_id,
  d.name AS district_name,
  s.epi_week_id,
  s.status,
  s.source_name,
  s.imported_at,
  s.created_at,
  dis.id AS subject_id,
  dis.name AS subject_name,
  'DISEASE' AS subject_type
FROM weekly_status s
JOIN districts d ON d.id = s.district_id
JOIN regions r ON r.id = s.region_id
JOIN diseases dis ON dis.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.region_id IS NOT NULL
  AND s.district_id IS NOT NULL
  AND s.sub_county_id IS NULL
  AND s.disease_id IS NOT NULL

UNION ALL

SELECT
  s.id,
  s.region_id,
  r.name AS region_name,
  s.district_id,
  d.name AS district_name,
  s.epi_week_id,
  s.status,
  s.source_name,
  s.imported_at,
  s.created_at,
  i.id AS subject_id,
  i.name AS subject_name,
  'INDICATOR' AS subject_type
FROM weekly_status s
JOIN districts d ON d.id = s.district_id
JOIN regions r ON r.id = s.region_id
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
  AND s.region_id IS NOT NULL
  AND s.district_id IS NOT NULL
  AND s.sub_county_id IS NULL
  AND s.indicator_id IS NOT NULL

ORDER BY district_name ASC, status ASC, subject_name ASC;


-- name: ListRegionWeeklySubjectsByWeek :many
SELECT
  s.id,
  s.region_id,
  r.name AS region_name,
  s.epi_week_id,
  s.status,
  s.source_name,
  s.imported_at,
  s.created_at,
  dis.id AS subject_id,
  dis.name AS subject_name,
  'DISEASE' AS subject_type
FROM weekly_status s
JOIN regions r ON r.id = s.region_id
JOIN diseases dis ON dis.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.region_id IS NOT NULL
  AND s.district_id IS NULL
  AND s.sub_county_id IS NULL
  AND s.disease_id IS NOT NULL

UNION ALL

SELECT
  s.id,
  s.region_id,
  r.name AS region_name,
  s.epi_week_id,
  s.status,
  s.source_name,
  s.imported_at,
  s.created_at,
  i.id AS subject_id,
  i.name AS subject_name,
  'INDICATOR' AS subject_type
FROM weekly_status s
JOIN regions r ON r.id = s.region_id
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
  AND s.region_id IS NOT NULL
  AND s.district_id IS NULL
  AND s.sub_county_id IS NULL
  AND s.indicator_id IS NOT NULL

ORDER BY region_name ASC, status ASC, subject_name ASC;


-- name: GetWeeklySubjectTotals :many
SELECT
  d.id AS subject_id,
  d.name AS subject_name,
  'DISEASE' AS subject_type,
  SUM(m.metric_value)::numeric AS total_value
FROM facility_weekly_metrics m
JOIN diseases d ON d.id = m.disease_id
WHERE m.epi_week_id = $1
  AND m.disease_id IS NOT NULL
GROUP BY d.id, d.name

UNION ALL

SELECT
  i.id AS subject_id,
  i.name AS subject_name,
  'INDICATOR' AS subject_type,
  SUM(m.metric_value)::numeric AS total_value
FROM facility_weekly_metrics m
JOIN indicators i ON i.id = m.indicator_id
WHERE m.epi_week_id = $1
  AND m.indicator_id IS NOT NULL
GROUP BY i.id, i.name

ORDER BY total_value DESC, subject_name ASC;


-- name: GetDiseaseDashboardSummaryByWeek :one
SELECT
  d.id AS disease_id,
  d.name AS disease_name,
  COALESCE((
    SELECT SUM(m.metric_value)::numeric
    FROM facility_weekly_metrics m
    WHERE m.disease_id = d.id
      AND m.epi_week_id = $2
  ), 0) AS facility_total,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.disease_id = d.id
      AND ws.epi_week_id = $2
      AND ws.status = 'MAROON'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS maroon_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.disease_id = d.id
      AND ws.epi_week_id = $2
      AND ws.status = 'RED'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS red_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.disease_id = d.id
      AND ws.epi_week_id = $2
      AND ws.status = 'YELLOW'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS yellow_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.disease_id = d.id
      AND ws.epi_week_id = $2
      AND ws.status = 'GREEN'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS green_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM alerts a
    WHERE a.disease_id = d.id
      AND a.epi_week_id = $2
  ), 0) AS total_alerts
FROM diseases d
WHERE d.id = $1
LIMIT 1;


-- name: GetIndicatorDashboardSummaryByWeek :one
SELECT
  i.id AS indicator_id,
  i.name AS indicator_name,
  COALESCE((
    SELECT SUM(m.metric_value)::numeric
    FROM facility_weekly_metrics m
    WHERE m.indicator_id = i.id
      AND m.epi_week_id = $2
  ), 0) AS facility_total,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.indicator_id = i.id
      AND ws.epi_week_id = $2
      AND ws.status = 'MAROON'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS maroon_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.indicator_id = i.id
      AND ws.epi_week_id = $2
      AND ws.status = 'RED'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS red_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.indicator_id = i.id
      AND ws.epi_week_id = $2
      AND ws.status = 'YELLOW'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS yellow_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM weekly_status ws
    WHERE ws.indicator_id = i.id
      AND ws.epi_week_id = $2
      AND ws.status = 'GREEN'
      AND ws.region_id IS NOT NULL
      AND ws.district_id IS NOT NULL
      AND ws.sub_county_id IS NULL
  ), 0) AS green_districts
FROM indicators i
WHERE i.id = $1
LIMIT 1;


-- name: GetTopFacilitiesByDiseaseAndWeek :many
SELECT
  f.id AS facility_id,
  f.name AS facility_name,
  d.name AS district_name,
  r.name AS region_name,
  m.metric_value
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN districts d ON d.id = f.district_id
LEFT JOIN regions r ON r.id = f.region_id
WHERE m.disease_id = $1
  AND m.epi_week_id = $2
ORDER BY m.metric_value DESC, f.name ASC
LIMIT $3;


-- name: GetTopFacilitiesByIndicatorAndWeek :many
SELECT
  f.id AS facility_id,
  f.name AS facility_name,
  d.name AS district_name,
  r.name AS region_name,
  m.metric_value
FROM facility_weekly_metrics m
JOIN facilities f ON f.id = m.facility_id
LEFT JOIN districts d ON d.id = f.district_id
LEFT JOIN regions r ON r.id = f.region_id
WHERE m.indicator_id = $1
  AND m.epi_week_id = $2
ORDER BY m.metric_value DESC, f.name ASC
LIMIT $3;
