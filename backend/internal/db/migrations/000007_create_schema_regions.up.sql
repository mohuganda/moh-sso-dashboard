CREATE TABLE regions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT NOT NULL UNIQUE,
  code TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO
  regions (name, code)
VALUES
  ('Acholi', NULL),
  ('Ankole', NULL),
  ('Bugisu', NULL),
  ('Bukedi', NULL),
  ('Bunyoro', NULL),
  ('Busoga', NULL),
  ('Kampala', NULL),
  ('Karamoja', NULL),
  ('Kigezi', NULL),
  ('Lango', NULL),
  ('North Buganda', NULL),
  ('North Central', NULL),
  ('South Buganda', NULL),
  ('South Central', NULL),
  ('Teso', NULL),
  ('Tooro', NULL),
  ('West Nile', NULL) ON CONFLICT (name) DO
UPDATE
SET
  updated_at = now();