-- Surveillance-owned schema. Document IDs are opaque references to the primary database.

-- Source: 000007_create_schema_regions.up.sql
CREATE TABLE regions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  code TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);



-- Source: 000008_create_schema_districts.up.sql
CREATE TABLE districts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  region_id UUID REFERENCES regions(id) ON DELETE
  SET
    NULL,
    code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);



-- Source: 000009_create_schema_sub_counties.up.sql
CREATE TABLE sub_counties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  district_id UUID NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
  code TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(name, district_id)
);



-- Source: 000010_create_schema_facilities.up.sql
CREATE TABLE facilities (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id TEXT UNIQUE, -- from source, e.g. record systems if available
  name TEXT NOT NULL,
  district_id UUID REFERENCES districts(id) ON DELETE SET NULL,
  sub_county_id UUID REFERENCES sub_counties(id) ON DELETE SET NULL,
  region_id UUID REFERENCES regions(id) ON DELETE SET NULL,
  facility_level TEXT,
  facility_type TEXT,
  dhis2_org_unit_id TEXT,
  latitude NUMERIC(10, 7),
  longitude NUMERIC(10, 7),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(name, district_id)
);




-- Source: 000011_create_schema_diseases.up.sql
CREATE TABLE diseases (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  short_name TEXT,
  code TEXT UNIQUE,
  category TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO
  diseases (name, short_name, code, category, is_active)
VALUES
  (
    'Acute Flaccid Paralysis',
    NULL,
    NULL,
    NULL,
    TRUE
  ),
  ('Anthrax', NULL, NULL, NULL, TRUE),
  ('Cholera', NULL, NULL, NULL, TRUE),
  ('Dengue Fever', NULL, NULL, NULL, TRUE),
  ('Guinea Worm', NULL, NULL, NULL, TRUE),
  ('Influenza new subtype', NULL, NULL, NULL, TRUE),
  ('Leprosy', NULL, NULL, NULL, TRUE),
  ('Lymphatic Filariasis', NULL, NULL, NULL, TRUE),
  ('Measles', NULL, NULL, NULL, TRUE),
  ('Onchocerciasis', NULL, NULL, NULL, TRUE),
  ('Plague', NULL, NULL, NULL, TRUE),
  ('Schistosomiasis', NULL, NULL, NULL, TRUE),
  ('Trachoma', NULL, NULL, NULL, TRUE),
  ('Yellow Fever', NULL, NULL, NULL, TRUE) ON CONFLICT (name) DO
UPDATE
SET
  is_active = TRUE,
  updated_at = now();

-- Source: 000012_create_schema_epi_weeks.up.sql
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

-- Source: 000013_create_schema_indicators.up.sql
CREATE TABLE indicators (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  short_name TEXT,
  code TEXT UNIQUE,
  indicator_type TEXT,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO
  indicators (
    name,
    short_name,
    code,
    indicator_type,
    is_active
  )
VALUES
  ('Cases', NULL, NULL, NULL, TRUE),
  (
    'Cases Tested with Microscopy',
    NULL,
    NULL,
    NULL,
    TRUE
  ),
  ('Cases Tested with RDT', NULL, NULL, NULL, TRUE),
  (
    'Early Neonatal Deaths 0-7 days',
    NULL,
    NULL,
    NULL,
    TRUE
  ),
  (
    'Fresh Still Birth Deaths',
    NULL,
    NULL,
    NULL,
    TRUE
  ),
  ('Macerated Still Births', NULL, NULL, NULL, TRUE),
  ('Maternal Deaths', NULL, NULL, NULL, TRUE),
  (
    'Microscopy Positive Cases',
    NULL,
    NULL,
    NULL,
    TRUE
  ),
  ('RDT Positive Cases', NULL, NULL, NULL, TRUE) ON CONFLICT (name) DO
UPDATE
SET
  is_active = TRUE,
  updated_at = now();

-- Source: 000014_weekly_status_schema.up.sql

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

-- Source: 000015_create_schema_alerts.up.sql
CREATE TYPE alert_status AS ENUM (
    'NEW',
    'ACKNOWLEDGE',
    'INVESTIGATING',
    'RESOLVED',
    'CLOSED'
);

CREATE TABLE alerts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id TEXT UNIQUE,
  disease_id UUID NOT NULL REFERENCES diseases(id) ON DELETE RESTRICT,
  district_id UUID REFERENCES districts(id) ON DELETE SET NULL,
  epi_week_id UUID REFERENCES epi_weeks(id) ON DELETE SET NULL,
  occurred_on DATE,
  created_on DATE,
  narrative TEXT NOT NULL,
  submitted_by TEXT,
  status alert_status NOT NULL DEFAULT 'NEW',
  source_name TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_alerts_disease ON alerts(disease_id);
CREATE INDEX idx_alerts_district ON alerts(district_id);
CREATE INDEX idx_alerts_week ON alerts(epi_week_id);
CREATE INDEX idx_alerts_created_on ON alerts(created_on);
CREATE INDEX idx_alerts_status ON alerts(status);

-- Source: 000016_create_schema_surveillance_import_batches.up.sql
CREATE TABLE surveillance_import_batches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_name TEXT NOT NULL,
  file_name TEXT,
  dataset_type TEXT NOT NULL,
  imported_by TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  status TEXT NOT NULL DEFAULT 'PENDING',
  notes TEXT,
  total_rows INT NOT NULL DEFAULT 0,
  success_rows INT NOT NULL DEFAULT 0,
  failed_rows INT NOT NULL DEFAULT 0,
  completed_at TIMESTAMPTZ,
  document_id UUID
);

ALTER TABLE surveillance_import_batches
ADD CONSTRAINT surveillance_import_batches_status_chk
CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED'));

-- Source: 000017_create_schema_surveillance_import_raw_rows.up.sql
CREATE TABLE surveillance_import_raw_rows (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  batch_id UUID NOT NULL REFERENCES surveillance_import_batches(id) ON DELETE CASCADE,
  row_number INT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_surveillance_import_raw_rows_batch ON surveillance_import_raw_rows(batch_id);
CREATE INDEX idx_surveillance_import_raw_rows_payload ON surveillance_import_raw_rows USING GIN(payload);


ALTER TABLE surveillance_import_raw_rows
ADD COLUMN status TEXT NOT NULL DEFAULT 'PENDING',
ADD COLUMN error_message TEXT;

ALTER TABLE surveillance_import_raw_rows
ADD CONSTRAINT surveillance_import_raw_rows_status_chk
CHECK (status IN ('PENDING', 'PROCESSED', 'FAILED'));

-- Source: 000018_create_schema_facility_weekly_metrics.up.sql
CREATE TABLE facility_weekly_metrics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  source_record_id TEXT,
  -- original record_id from source file
  facility_id UUID NOT NULL REFERENCES facilities(id) ON DELETE CASCADE,
  disease_id UUID REFERENCES diseases(id) ON DELETE CASCADE,
  indicator_id UUID REFERENCES indicators(id) ON DELETE CASCADE,
  epi_week_id UUID NOT NULL REFERENCES epi_weeks(id) ON DELETE CASCADE,
  metric_value NUMERIC(14, 2) NOT NULL DEFAULT 0,
  source_name TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT facility_weekly_metrics_one_subject_chk CHECK (
    (disease_id IS NOT NULL AND indicator_id IS NULL)
    OR
    (disease_id IS NULL AND indicator_id IS NOT NULL)
  ),
  UNIQUE(source_record_id)
);

CREATE INDEX idx_facility_metrics_week
ON facility_weekly_metrics(epi_week_id);

CREATE INDEX idx_facility_metrics_disease
ON facility_weekly_metrics(disease_id);

CREATE INDEX idx_facility_metrics_indicator
ON facility_weekly_metrics(indicator_id);

CREATE INDEX idx_facility_metrics_facility
ON facility_weekly_metrics(facility_id);

CREATE UNIQUE INDEX uq_facility_weekly_metrics_disease
ON facility_weekly_metrics (facility_id, disease_id, epi_week_id)
WHERE disease_id IS NOT NULL;

CREATE UNIQUE INDEX uq_facility_weekly_metrics_indicator
ON facility_weekly_metrics (facility_id, indicator_id, epi_week_id)
WHERE indicator_id IS NOT NULL;

-- Source: 000019_disease_indicator_schema.up.sql
CREATE TABLE disease_indicators (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  disease_id UUID NOT NULL REFERENCES diseases(id) ON DELETE CASCADE,
  indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (disease_id, indicator_id)
);

INSERT INTO
  disease_indicators (disease_id, indicator_id, is_active)
VALUES
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Acute Flaccid Paralysis'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Anthrax'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Cholera'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Dengue Fever'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Guinea Worm'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Influenza new subtype'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Leprosy'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Lymphatic Filariasis'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Measles'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Onchocerciasis'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Plague'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Schistosomiasis'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Trachoma'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ),
  (
    (
      SELECT
        id
      FROM
        diseases
      WHERE
        name = 'Yellow Fever'
    ),
    (
      SELECT
        id
      FROM
        indicators
      WHERE
        name = 'Cases'
    ),
    TRUE
  ) ON CONFLICT (disease_id, indicator_id) DO
UPDATE
SET
  is_active = TRUE,
  updated_at = now();
