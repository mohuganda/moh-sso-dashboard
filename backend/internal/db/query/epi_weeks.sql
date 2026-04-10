-- name: CreateEpiWeek :one
INSERT INTO epi_weeks (
  epi_year,
  epi_week,
  week_start_date,
  week_end_date
) VALUES (
  $1,
  $2,
  $3,
  $4
)
RETURNING *;

-- name: GetEpiWeekByID :one
SELECT *
FROM epi_weeks
WHERE id = $1
LIMIT 1;

-- name: GetEpiWeekByYearWeek :one
SELECT *
FROM epi_weeks
WHERE epi_year = $1
  AND epi_week = $2
LIMIT 1;

-- name: ListEpiWeeksByYear :many
SELECT *
FROM epi_weeks
WHERE epi_year = $1
ORDER BY epi_week DESC;

-- name: UpsertEpiWeek :one
INSERT INTO epi_weeks (
  epi_year,
  epi_week,
  week_start_date,
  week_end_date
) VALUES (
  $1,
  $2,
  $3,
  $4
)
ON CONFLICT (epi_year, epi_week)
DO UPDATE SET
  week_start_date = COALESCE(EXCLUDED.week_start_date, epi_weeks.week_start_date),
  week_end_date = COALESCE(EXCLUDED.week_end_date, epi_weeks.week_end_date)
RETURNING *;  