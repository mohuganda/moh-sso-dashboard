CREATE TABLE national_weekly_status (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  disease_id UUID NOT NULL REFERENCES diseases(id) ON DELETE CASCADE,
  indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
  epi_week_id UUID NOT NULL REFERENCES epi_weeks(id) ON DELETE CASCADE,
  status risk_level NOT NULL,
  source_name TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(disease_id, epi_week_id)
);

CREATE INDEX idx_national_status_week ON national_weekly_status(epi_week_id);

CREATE INDEX idx_national_status_disease ON national_weekly_status(disease_id);

CREATE INDEX idx_national_status_status ON national_weekly_status(status);

CREATE UNIQUE INDEX uq_national_weekly_status_disease ON national_weekly_status (disease_id, epi_week_id)
WHERE
  disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_national_weekly_status_indicator ON national_weekly_status (indicator_id, epi_week_id)
WHERE
  indicator_id IS NOT NULL;