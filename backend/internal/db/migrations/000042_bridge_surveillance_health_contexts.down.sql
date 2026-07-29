DROP TRIGGER IF EXISTS facilities_sync_health_context ON facilities;
DROP TRIGGER IF EXISTS sub_counties_sync_health_context ON sub_counties;
DROP TRIGGER IF EXISTS districts_sync_health_context ON districts;
DROP TRIGGER IF EXISTS regions_sync_health_context ON regions;

DROP FUNCTION IF EXISTS sync_surveillance_facility_health_context();
DROP FUNCTION IF EXISTS sync_surveillance_sub_county_health_context();
DROP FUNCTION IF EXISTS sync_surveillance_district_health_context();
DROP FUNCTION IF EXISTS sync_surveillance_region_health_context();
DROP FUNCTION IF EXISTS health_context_alias_related_to_scope(TEXT, TEXT, UUID, BOOLEAN);
DROP FUNCTION IF EXISTS health_context_alias_in_scope(TEXT, TEXT, UUID, BOOLEAN);
DROP FUNCTION IF EXISTS sync_surveillance_health_context(TEXT, UUID, TEXT, TEXT, TEXT, TEXT, UUID);

DELETE FROM user_active_health_contexts
WHERE context_node_id IN (
    SELECT id FROM health_context_nodes
    WHERE source = 'SURVEILLANCE' AND code LIKE 'SURV-%'
);

DELETE FROM user_health_context_assignments
WHERE context_node_id IN (
    SELECT id FROM health_context_nodes
    WHERE source = 'SURVEILLANCE' AND code LIKE 'SURV-%'
);

DELETE FROM group_health_context_assignments
WHERE context_node_id IN (
    SELECT id FROM health_context_nodes
    WHERE source = 'SURVEILLANCE' AND code LIKE 'SURV-%'
);

DELETE FROM health_context_nodes
WHERE source = 'SURVEILLANCE'
  AND code LIKE 'SURV-%';
