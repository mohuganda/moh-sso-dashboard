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