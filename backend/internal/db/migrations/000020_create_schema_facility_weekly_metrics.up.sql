CREATE TABLE facility_weekly_metrics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_record_id TEXT,
  -- original record_id from source file
  facility_id UUID NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
  disease_id UUID NOT NULL REFERENCES diseases(id) ON DELETE CASCADE,
  indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
  epi_week_id UUID NOT NULL REFERENCES epi_weeks(id) ON DELETE CASCADE,
  metric_value NUMERIC(14, 2) NOT NULL DEFAULT 0,
  source_name TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT facility_weekly_metrics_one_subject_chk CHECK (
    (
      disease_id IS NOT NULL
      AND indicator_id IS NULL
    )
    OR (
      disease_id IS NULL
      AND indicator_id IS NOT NULL
    )
  ),
  UNIQUE(source_record_id),
  UNIQUE(
    facility_id,
    disease_id,
    indicator_id,
    epi_week_id
  )
);

CREATE INDEX idx_facility_metrics_week ON facility_weekly_metrics(epi_week_id);

CREATE INDEX idx_facility_metrics_disease ON facility_weekly_metrics(disease_id);

CREATE INDEX idx_facility_metrics_facility ON facility_weekly_metrics(facility_id);

CREATE UNIQUE INDEX uq_facility_weekly_metrics_disease ON facility_weekly_metrics (facility_id, disease_id, epi_week_id)
WHERE
  disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_facility_weekly_metrics_indicator ON facility_weekly_metrics (facility_id, indicator_id, epi_week_id)
WHERE
  indicator_id IS NOT NULL;