DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'import') THEN
    RETURN;
  END IF;
  DROP TABLE IF EXISTS import.ghsc_psm_malaria;
  DROP TABLE IF EXISTS import.ghsc_psm_pharma;
  DROP TABLE IF EXISTS import.ghsc_psm_lab;
  DROP TABLE IF EXISTS import.gdf_tb_orders;
  DROP TABLE IF EXISTS import.gf_pipeline;
END;
$$;
