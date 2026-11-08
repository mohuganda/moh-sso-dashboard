-- Reference rows may already be used by reports and health-context aliases.
-- Retain them when reverting this migration; never cascade-delete operational data.
SELECT 1;
