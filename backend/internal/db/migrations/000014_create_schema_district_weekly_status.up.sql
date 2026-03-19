CREATE TABLE district_weekly_status (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  district_id UUID NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
  disease_id UUID NOT NULL REFERENCES diseases(id) ON DELETE CASCADE,
  indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
  epi_week_id UUID NOT NULL REFERENCES epi_weeks(id) ON DELETE CASCADE,
  status risk_level NOT NULL,
  source_name TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT district_weekly_status_one_subject_chk CHECK (
    (
      disease_id IS NOT NULL
      AND indicator_id IS NULL
    )
    OR (
      disease_id IS NULL
      AND indicator_id IS NOT NULL
    )
  ),
  UNIQUE(
    district_id,
    disease_id,
    indicator_id,
    epi_week_id
  )
);

CREATE INDEX idx_district_status_week ON district_weekly_status(epi_week_id);

CREATE INDEX idx_district_status_district ON district_weekly_status(district_id);

CREATE INDEX idx_district_status_disease ON district_weekly_status(disease_id);

CREATE INDEX idx_district_status_status ON district_weekly_status(status);

CREATE UNIQUE INDEX uq_district_weekly_status_disease ON district_weekly_status (district_id, disease_id, epi_week_id)
WHERE
  disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_district_weekly_status_indicator ON district_weekly_status (district_id, indicator_id, epi_week_id)
WHERE
  indicator_id IS NOT NULL;