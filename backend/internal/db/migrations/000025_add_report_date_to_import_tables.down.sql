DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'import') THEN
    ALTER TABLE import.nms_stock_issues  DROP COLUMN IF EXISTS report_date;
    ALTER TABLE import.nms_stock_on_hand DROP COLUMN IF EXISTS report_date;
    ALTER TABLE import.jms_stock_issues  DROP COLUMN IF EXISTS report_date;
    ALTER TABLE import.jms_stock_on_hand DROP COLUMN IF EXISTS report_date;
  END IF;
END;
$$;
