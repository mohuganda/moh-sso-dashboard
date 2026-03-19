CREATE TABLE sub_counties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL,
  district_id UUID NOT NULL REFERENCES districts(id) ON DELETE CASCADE,
  code TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(name, district_id)
);