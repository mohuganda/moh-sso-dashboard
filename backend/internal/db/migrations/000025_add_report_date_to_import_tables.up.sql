-- Adds report_date to the four import tables that live in remote-postgres.
-- Wrapped in a schema-existence check so this migration is safe to run
-- against the main sso_dashboard database (which has no import schema).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'import') THEN
    ALTER TABLE import.nms_stock_issues  ADD COLUMN IF NOT EXISTS report_date DATE;
    ALTER TABLE import.nms_stock_on_hand ADD COLUMN IF NOT EXISTS report_date DATE;
    ALTER TABLE import.jms_stock_issues  ADD COLUMN IF NOT EXISTS report_date DATE;
    ALTER TABLE import.jms_stock_on_hand ADD COLUMN IF NOT EXISTS report_date DATE;
  END IF;
END;
$$;
