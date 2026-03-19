-- name: GetNationalWeeklyStatusSummary :many
SELECT
  status,
  COUNT(*)::bigint AS total_items
FROM (
  SELECT status
  FROM national_weekly_status
  WHERE epi_week_id = $1
) x
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
FROM national_weekly_status s
JOIN diseases d ON d.id = s.disease_id
WHERE s.epi_week_id = $1
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
FROM national_weekly_status s
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
  AND s.indicator_id IS NOT NULL

ORDER BY status, subject_name;


-- name: ListDistrictWeeklySubjectsByWeek :many
SELECT
  s.id,
  s.district_id,
  d.name AS district_name,
  s.epi_week_id,
  s.status,
  s.source_name,
  dis.id AS subject_id,
  dis.name AS subject_name,
  'DISEASE' AS subject_type
FROM district_weekly_status s
JOIN districts d ON d.id = s.district_id
JOIN diseases dis ON dis.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.disease_id IS NOT NULL

UNION ALL

SELECT
  s.id,
  s.district_id,
  d.name AS district_name,
  s.epi_week_id,
  s.status,
  s.source_name,
  i.id AS subject_id,
  i.name AS subject_name,
  'INDICATOR' AS subject_type
FROM district_weekly_status s
JOIN districts d ON d.id = s.district_id
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
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
  dis.id AS subject_id,
  dis.name AS subject_name,
  'DISEASE' AS subject_type
FROM region_weekly_status s
JOIN regions r ON r.id = s.region_id
JOIN diseases dis ON dis.id = s.disease_id
WHERE s.epi_week_id = $1
  AND s.disease_id IS NOT NULL

UNION ALL

SELECT
  s.id,
  s.region_id,
  r.name AS region_name,
  s.epi_week_id,
  s.status,
  s.source_name,
  i.id AS subject_id,
  i.name AS subject_name,
  'INDICATOR' AS subject_type
FROM region_weekly_status s
JOIN regions r ON r.id = s.region_id
JOIN indicators i ON i.id = s.indicator_id
WHERE s.epi_week_id = $1
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
    FROM district_weekly_status ds
    WHERE ds.disease_id = d.id
      AND ds.epi_week_id = $2
      AND ds.status = 'MAROON'
  ), 0) AS maroon_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM district_weekly_status ds
    WHERE ds.disease_id = d.id
      AND ds.epi_week_id = $2
      AND ds.status = 'RED'
  ), 0) AS red_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM district_weekly_status ds
    WHERE ds.disease_id = d.id
      AND ds.epi_week_id = $2
      AND ds.status = 'YELLOW'
  ), 0) AS yellow_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM district_weekly_status ds
    WHERE ds.disease_id = d.id
      AND ds.epi_week_id = $2
      AND ds.status = 'GREEN'
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
    FROM district_weekly_status ds
    WHERE ds.indicator_id = i.id
      AND ds.epi_week_id = $2
      AND ds.status = 'MAROON'
  ), 0) AS maroon_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM district_weekly_status ds
    WHERE ds.indicator_id = i.id
      AND ds.epi_week_id = $2
      AND ds.status = 'RED'
  ), 0) AS red_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM district_weekly_status ds
    WHERE ds.indicator_id = i.id
      AND ds.epi_week_id = $2
      AND ds.status = 'YELLOW'
  ), 0) AS yellow_districts,
  COALESCE((
    SELECT COUNT(*)::bigint
    FROM district_weekly_status ds
    WHERE ds.indicator_id = i.id
      AND ds.epi_week_id = $2
      AND ds.status = 'GREEN'
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