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