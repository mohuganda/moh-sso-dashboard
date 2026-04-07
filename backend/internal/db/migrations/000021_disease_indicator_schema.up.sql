CREATE TABLE disease_indicators (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  disease_id UUID NOT NULL REFERENCES diseases(id) ON DELETE CASCADE,
  indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (disease_id, indicator_id)
);



INSERT INTO disease_indicators (
  disease_id,
  indicator_id,
  is_active
)
SELECT d.id, i.id, TRUE
FROM diseases d
JOIN indicators i ON i.name = 'Cases'
WHERE d.name IN (
  'Acute Flaccid Paralysis',
  'Anthrax',
  'Cholera',
  'Dengue Fever',
  'Guinea Worm',
  'Influenza new subtype',
  'Leprosy',
  'Lymphatic Filariasis',
  'Malaria',
  'Measles',
  'Onchocerciasis',
  'Plague',
  'Schistosomiasis',
  'Trachoma',
  'Yellow Fever'
)
ON CONFLICT (disease_id, indicator_id) DO NOTHING;

-- Malaria-specific lab/testing indicators from data.csv
INSERT INTO disease_indicators (
  disease_id,
  indicator_id,
  is_active
)
SELECT d.id, i.id, TRUE
FROM diseases d
JOIN indicators i
  ON i.name IN (
    'Cases Tested with Microscopy',
    'Cases Tested with RDT',
    'Microscopy Positive Cases',
    'RDT Positive Cases'
  )
WHERE d.name = 'Malaria'
ON CONFLICT (disease_id, indicator_id) DO NOTHING;

-- Maternal death-related indicators
INSERT INTO disease_indicators (
  disease_id,
  indicator_id,
  is_active
)
SELECT d.id, i.id, TRUE
FROM diseases d
JOIN indicators i
  ON i.name IN (
    'Maternal Deaths',
    'Early Neonatal Deaths 0-7 days',
    'Fresh Still Birth Deaths',
    'Macerated Still Births'
  )
WHERE d.name = 'Maternal Death'
ON CONFLICT (disease_id, indicator_id) DO NOTHING;