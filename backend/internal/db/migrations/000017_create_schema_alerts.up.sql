CREATE TABLE alerts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  external_id TEXT UNIQUE,
  disease_id UUID NOT NULL REFERENCES diseases(id) ON DELETE RESTRICT,
  district_id UUID REFERENCES districts(id) ON DELETE SET NULL,
  epi_week_id UUID REFERENCES epi_weeks(id) ON DELETE SET NULL,
  occurred_on DATE,
  created_on DATE,
  narrative TEXT NOT NULL,
  submitted_by TEXT,
  status alert_status NOT NULL DEFAULT 'NEW',
  source_name TEXT,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_alerts_disease ON alerts(disease_id);
CREATE INDEX idx_alerts_district ON alerts(district_id);
CREATE INDEX idx_alerts_week ON alerts(epi_week_id);
CREATE INDEX idx_alerts_created_on ON alerts(created_on);
CREATE INDEX idx_alerts_status ON alerts(status);