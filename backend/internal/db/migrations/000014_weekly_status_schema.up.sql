
CREATE TYPE risk_level AS ENUM ('MAROON', 'RED', 'YELLOW', 'GREEN');

CREATE TABLE weekly_status (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  region_id UUID REFERENCES regions(id) ON DELETE CASCADE,
  district_id UUID REFERENCES districts(id) ON DELETE CASCADE,
  sub_county_id UUID REFERENCES sub_counties(id) ON DELETE CASCADE,

  disease_id UUID REFERENCES diseases(id) ON DELETE CASCADE,
  indicator_id UUID REFERENCES indicators(id) ON DELETE CASCADE,

  epi_week_id UUID NOT NULL REFERENCES epi_weeks(id) ON DELETE CASCADE,
  status risk_level NOT NULL,
  source_name TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT weekly_status_subject_chk CHECK (
    (
      disease_id IS NOT NULL
      AND indicator_id IS NULL
    )
    OR (
      disease_id IS NULL
      AND indicator_id IS NOT NULL
    )
  ),

  CONSTRAINT weekly_status_scope_chk CHECK (
    (
      region_id IS NULL
      AND district_id IS NULL
      AND sub_county_id IS NULL
    )
    OR (
      region_id IS NOT NULL
      AND district_id IS NULL
      AND sub_county_id IS NULL
    )
    OR (
      region_id IS NOT NULL
      AND district_id IS NOT NULL
      AND sub_county_id IS NULL
    )
    OR (
      region_id IS NOT NULL
      AND district_id IS NOT NULL
      AND sub_county_id IS NOT NULL
    )
  )
);

CREATE UNIQUE INDEX uq_weekly_status_national_disease
ON weekly_status (disease_id, epi_week_id)
WHERE
  region_id IS NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_weekly_status_national_indicator
ON weekly_status (indicator_id, epi_week_id)
WHERE
  region_id IS NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND indicator_id IS NOT NULL;

CREATE UNIQUE INDEX uq_weekly_status_region_disease
ON weekly_status (region_id, disease_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_weekly_status_region_indicator
ON weekly_status (region_id, indicator_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NULL
  AND sub_county_id IS NULL
  AND indicator_id IS NOT NULL;

CREATE UNIQUE INDEX uq_weekly_status_district_disease
ON weekly_status (district_id, disease_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NULL
  AND disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_weekly_status_district_indicator
ON weekly_status (district_id, indicator_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NULL
  AND indicator_id IS NOT NULL;

CREATE UNIQUE INDEX uq_weekly_status_sub_county_disease
ON weekly_status (sub_county_id, disease_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NOT NULL
  AND disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_weekly_status_sub_county_indicator
ON weekly_status (sub_county_id, indicator_id, epi_week_id)
WHERE
  region_id IS NOT NULL
  AND district_id IS NOT NULL
  AND sub_county_id IS NOT NULL
  AND indicator_id IS NOT NULL;

CREATE INDEX idx_weekly_status_week ON weekly_status(epi_week_id);
CREATE INDEX idx_weekly_status_region ON weekly_status(region_id);
CREATE INDEX idx_weekly_status_district ON weekly_status(district_id);
CREATE INDEX idx_weekly_status_sub_county ON weekly_status(sub_county_id);
CREATE INDEX idx_weekly_status_disease ON weekly_status(disease_id);
CREATE INDEX idx_weekly_status_indicator ON weekly_status(indicator_id);
CREATE INDEX idx_weekly_status_status ON weekly_status(status);