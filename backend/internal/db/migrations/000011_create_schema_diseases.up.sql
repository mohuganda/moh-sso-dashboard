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
  ('Malaria', 'Malaria', 'MAL', 'Vector-borne', TRUE),
  ('Measles', 'Measles', 'MEA', 'Vaccine-preventable', TRUE),
  ('Mpox', 'Mpox', 'MPOX', 'Viral', TRUE),
  ('Cholera', 'Cholera', 'CHOL', 'Water-borne', TRUE),
  ('Typhoid Fever', 'Typhoid', 'TYP', 'Water-borne', TRUE),
  ('Dengue Fever', 'Dengue', 'DEN', 'Vector-borne', TRUE),
  ('Yellow Fever', 'YellowFever', 'YF', 'Vaccine-preventable', TRUE),
  ('Ebola Virus Disease', 'EVD', 'EVD', 'Viral Hemorrhagic Fever', TRUE),
  ('Marburg Virus Disease', 'MVD', 'MVD', 'Viral Hemorrhagic Fever', TRUE),
  ('COVID-19', 'COVID-19', 'COVID', 'Respiratory', TRUE),
  ('Tuberculosis', 'TB', 'TB', 'Respiratory', TRUE),
  ('Acute Flaccid Paralysis', 'AFP', 'AFP', 'Neurological', TRUE),
  ('Meningitis', 'Meningitis', 'MEN', 'Bacterial', TRUE),
  ('Plague', 'Plague', 'PLG', 'Bacterial', TRUE),
  ('Rabies', 'Rabies', 'RAB', 'Zoonotic', TRUE),
  ('Anthrax', 'Anthrax', 'ANT', 'Zoonotic', TRUE),
  ('Hepatitis E', 'HepE', 'HEPE', 'Water-borne', TRUE),
  ('Dysentery', 'Dysentery', 'DYS', 'Gastrointestinal', TRUE),
  ('Pertussis', 'Pertussis', 'PERT', 'Vaccine-preventable', TRUE),
  ('Neonatal Tetanus', 'NNT', 'NNT', 'Vaccine-preventable', TRUE)
ON CONFLICT (name) DO NOTHING;