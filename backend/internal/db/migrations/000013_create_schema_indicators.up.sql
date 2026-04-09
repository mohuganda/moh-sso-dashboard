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