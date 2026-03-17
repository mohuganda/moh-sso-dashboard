-- 20260309_drop_announcements.sql

-- -------------------------------------
-- Drop trigger first
-- -------------------------------------
DROP TRIGGER IF EXISTS trg_announcements_set_updated_fields ON announcements;

-- -------------------------------------
-- Drop trigger function
-- -------------------------------------
DROP FUNCTION IF EXISTS set_announcements_updated_fields();

-- -------------------------------------
-- Drop child tables first
-- -------------------------------------
DROP TABLE IF EXISTS announcement_users;
DROP TABLE IF EXISTS announcement_roles;
DROP TABLE IF EXISTS announcement_clients;

-- -------------------------------------
-- Drop main table
-- -------------------------------------
DROP TABLE IF EXISTS announcements;

-- -------------------------------------
-- Drop enum types
-- -------------------------------------
DROP TYPE IF EXISTS announcement_audience_type;
DROP TYPE IF EXISTS announcement_level;
DROP TYPE IF EXISTS announcement_status;