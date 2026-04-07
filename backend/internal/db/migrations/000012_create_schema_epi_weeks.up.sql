CREATE TABLE epi_weeks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  epi_year INT NOT NULL,
  epi_week INT NOT NULL CHECK (epi_week BETWEEN 1 AND 53),
  week_start_date DATE,
  week_end_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(epi_year, epi_week)
);
