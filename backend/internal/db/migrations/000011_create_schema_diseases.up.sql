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