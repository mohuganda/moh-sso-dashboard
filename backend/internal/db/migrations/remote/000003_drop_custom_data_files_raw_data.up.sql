-- raw_data duplicated file_data for every write path (CSV wrote the exact
-- same map into both columns; Excel's row_hash is computed from the raw
-- extracted values in memory before either column is populated, so dropping
-- the stored column doesn't affect hash stability). Keep only file_data.
ALTER TABLE import.custom_data_files
    DROP COLUMN IF EXISTS raw_data;
