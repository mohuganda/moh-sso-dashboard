-- Import processing updates every row's status. The UI reads staging by batch ID;
-- a GIN index on all source JSON adds substantial write cost without serving a query.
DROP INDEX IF EXISTS idx_surveillance_import_raw_rows_payload;
