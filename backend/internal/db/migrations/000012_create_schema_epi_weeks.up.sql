CREATE TABLE epi_weeks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  epi_year INT NOT NULL,
  epi_week INT NOT NULL CHECK (epi_week BETWEEN 1 AND 53),
  week_start_date DATE,
  week_end_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(epi_year, epi_week)
);


-- Seed epi weeks for one year
-- Change 2025 to any year you want
WITH year_input AS (
  SELECT 2025 AS epi_year
),
weeks AS (
  SELECT
    yi.epi_year,
    gs AS epi_week,
    to_date(yi.epi_year::text || lpad(gs::text, 2, '0') || '1', 'IYYYIWID') AS week_start_date,
    to_date(yi.epi_year::text || lpad(gs::text, 2, '0') || '7', 'IYYYIWID') AS week_end_date
  FROM year_input yi
  CROSS JOIN generate_series(1, 53) gs
)
INSERT INTO epi_weeks (
  epi_year,
  epi_week,
  week_start_date,
  week_end_date
)
SELECT
  epi_year,
  epi_week,
  week_start_date,
  week_end_date
FROM weeks
WHERE EXTRACT(ISOYEAR FROM week_start_date) = epi_year
ON CONFLICT (epi_year, epi_week) DO NOTHING;