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


INSERT INTO diseases (
  name,
  short_name,
  code,
  category,
  is_active
)
VALUES
  ('Acute Flaccid Paralysis', 'AFP', 'AFP', 'Neurological', TRUE),
  ('Anthrax', 'Anthrax', 'ANT', 'Zoonotic', TRUE),
  ('Cholera', 'Cholera', 'CHOL', 'Water-borne', TRUE),
  ('Dengue Fever', 'Dengue', 'DEN', 'Vector-borne', TRUE),
  ('Guinea Worm', 'Guinea Worm', 'GW', NULL, TRUE),
  ('Influenza new subtype', 'Influenza New Subtype', 'INS', 'Respiratory', TRUE),
  ('Leprosy', 'Leprosy', 'LEP', NULL, TRUE),
  ('Lymphatic Filariasis', 'LF', 'LF', NULL, TRUE),
  ('Malaria', 'Malaria', 'MAL', 'Vector-borne', TRUE),
  ('Maternal Death', 'Maternal Death', 'MATD', NULL, TRUE),
  ('Measles', 'Measles', 'MEA', 'Vaccine-preventable', TRUE),
  ('Onchocerciasis', 'Onchocerciasis', 'ONCHO', NULL, TRUE),
  ('Plague', 'Plague', 'PLG', 'Bacterial', TRUE),
  ('Schistosomiasis', 'Schistosomiasis', 'SCH', NULL, TRUE),
  ('Trachoma', 'Trachoma', 'TRA', NULL, TRUE),
  ('Yellow Fever', 'Yellow Fever', 'YF', 'Vaccine-preventable', TRUE)
ON CONFLICT (name) DO NOTHING;