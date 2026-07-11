CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS epi_weeks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    epi_year INT NOT NULL,
    epi_week INT NOT NULL CHECK (epi_week BETWEEN 1 AND 53),
    week_start_date DATE NOT NULL,
    week_end_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (epi_year, epi_week),
    CHECK (week_end_date = week_start_date + 6)
);

WITH current_epi_year AS (
    SELECT EXTRACT(ISOYEAR FROM CURRENT_DATE)::INT AS epi_year
),
generated_weeks AS (
    SELECT
        current_epi_year.epi_year,
        week_start::DATE AS week_start_date,
        (week_start + INTERVAL '6 days')::DATE AS week_end_date
    FROM current_epi_year
    CROSS JOIN LATERAL generate_series(
        TO_DATE(current_epi_year.epi_year || '-01-1', 'IYYY-IW-ID'),
        TO_DATE(current_epi_year.epi_year || '-53-1', 'IYYY-IW-ID'),
        INTERVAL '1 week'
    ) AS week_start
)
INSERT INTO epi_weeks (
    epi_year,
    epi_week,
    week_start_date,
    week_end_date
)
SELECT
    EXTRACT(ISOYEAR FROM week_start_date)::INT,
    EXTRACT(WEEK FROM week_start_date)::INT,
    week_start_date,
    week_end_date
FROM generated_weeks
WHERE EXTRACT(ISOYEAR FROM week_start_date)::INT = epi_year
ON CONFLICT (epi_year, epi_week)
DO UPDATE SET
    week_start_date = EXCLUDED.week_start_date,
    week_end_date = EXCLUDED.week_end_date;