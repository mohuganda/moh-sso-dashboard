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

INSERT INTO indicators (
  name,
  short_name,
  code,
  indicator_type,
  is_active
)
VALUES
  ('Cases', 'Cases', 'CASES', 'count', TRUE),
  ('Cases Tested with Microscopy', 'Microscopy Tested', 'TEST_MIC', 'laboratory', TRUE),
  ('Cases Tested with RDT', 'RDT Tested', 'TEST_RDT', 'laboratory', TRUE),
  ('Microscopy Positive Cases', 'Microscopy Positive', 'POS_MIC', 'laboratory', TRUE),
  ('RDT Positive Cases', 'RDT Positive', 'POS_RDT', 'laboratory', TRUE),
  ('Maternal Deaths', 'Maternal Deaths', 'MAT_DEATH', 'mortality', TRUE),
  ('Early Neonatal Deaths 0-7 days', 'Early Neonatal Deaths', 'ENND_0_7', 'mortality', TRUE),
  ('Fresh Still Birth Deaths', 'Fresh Still Birth Deaths', 'FSB_DEATH', 'mortality', TRUE),
  ('Macerated Still Births', 'Macerated Still Births', 'MSB', 'mortality', TRUE)
ON CONFLICT (name) DO NOTHING;