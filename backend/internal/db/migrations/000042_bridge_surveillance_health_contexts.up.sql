-- Materialize the surveillance geography as health-context nodes. Domain IDs
-- remain owned by surveillance and are linked through explicit aliases.
CREATE OR REPLACE FUNCTION sync_surveillance_health_context(
    p_namespace TEXT,
    p_external_id UUID,
    p_code TEXT,
    p_name TEXT,
    p_context_type TEXT,
    p_parent_namespace TEXT DEFAULT NULL,
    p_parent_external_id UUID DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
    context_id UUID;
    parent_context_id UUID;
    current_parent_id UUID;
BEGIN
    IF p_parent_namespace IS NOT NULL AND p_parent_external_id IS NOT NULL THEN
        SELECT context_node_id
        INTO parent_context_id
        FROM health_context_aliases
        WHERE namespace = p_parent_namespace
          AND external_id = p_parent_external_id::text;
    ELSE
        SELECT id
        INTO parent_context_id
        FROM health_context_nodes
        WHERE code = 'UG';
    END IF;

    SELECT context_node_id
    INTO context_id
    FROM health_context_aliases
    WHERE namespace = p_namespace
      AND external_id = p_external_id::text;

    IF context_id IS NULL THEN
        INSERT INTO health_context_nodes (
            code,
            name,
            context_type,
            parent_id,
            source,
            metadata,
            enabled
        )
        VALUES (
            p_code,
            p_name,
            p_context_type,
            parent_context_id,
            'SURVEILLANCE',
            jsonb_build_object(
                'namespace', p_namespace,
                'externalId', p_external_id::text
            ),
            TRUE
        )
        ON CONFLICT (code) DO UPDATE
        SET name = EXCLUDED.name,
            context_type = EXCLUDED.context_type,
            parent_id = EXCLUDED.parent_id,
            metadata = health_context_nodes.metadata || EXCLUDED.metadata,
            enabled = TRUE,
            updated_at = NOW()
        RETURNING id INTO context_id;

        INSERT INTO health_context_aliases (
            context_node_id,
            namespace,
            external_id
        )
        VALUES (context_id, p_namespace, p_external_id::text)
        ON CONFLICT (namespace, external_id) DO UPDATE
        SET context_node_id = EXCLUDED.context_node_id;
    ELSE
        SELECT parent_id
        INTO current_parent_id
        FROM health_context_nodes
        WHERE id = context_id;

        IF current_parent_id IS DISTINCT FROM parent_context_id THEN
            DELETE FROM health_context_closure
            WHERE descendant_id IN (
                SELECT descendant_id
                FROM health_context_closure
                WHERE ancestor_id = context_id
            )
            AND ancestor_id IN (
                SELECT ancestor_id
                FROM health_context_closure
                WHERE descendant_id = context_id
                  AND ancestor_id <> context_id
            );

            IF parent_context_id IS NOT NULL THEN
                INSERT INTO health_context_closure (
                    ancestor_id,
                    descendant_id,
                    depth
                )
                SELECT parent_tree.ancestor_id,
                       child_tree.descendant_id,
                       parent_tree.depth + child_tree.depth + 1
                FROM health_context_closure parent_tree
                CROSS JOIN health_context_closure child_tree
                WHERE parent_tree.descendant_id = parent_context_id
                  AND child_tree.ancestor_id = context_id
                ON CONFLICT (ancestor_id, descendant_id) DO UPDATE
                SET depth = EXCLUDED.depth;
            END IF;
        END IF;

        UPDATE health_context_nodes
        SET name = p_name,
            context_type = p_context_type,
            parent_id = parent_context_id,
            enabled = TRUE,
            updated_at = NOW()
        WHERE id = context_id
          AND (
              name IS DISTINCT FROM p_name
              OR context_type IS DISTINCT FROM p_context_type
              OR parent_id IS DISTINCT FROM parent_context_id
              OR enabled IS DISTINCT FROM TRUE
          );
    END IF;

    RETURN context_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION health_context_alias_in_scope(
    p_namespace TEXT,
    p_external_id TEXT,
    p_context_id UUID,
    p_include_descendants BOOLEAN
)
RETURNS BOOLEAN AS $$
    SELECT EXISTS (
        SELECT 1
        FROM health_context_aliases alias
        WHERE alias.namespace = p_namespace
          AND alias.external_id = p_external_id
          AND (
              alias.context_node_id = p_context_id
              OR (
                  p_include_descendants
                  AND EXISTS (
                      SELECT 1
                      FROM health_context_closure closure
                      WHERE closure.ancestor_id = p_context_id
                        AND closure.descendant_id = alias.context_node_id
                  )
              )
          )
    );
$$ LANGUAGE sql STABLE;

-- Hierarchy selectors may expose the selected node's ancestors so a
-- facility-scoped user can still navigate through its region and district.
CREATE OR REPLACE FUNCTION health_context_alias_related_to_scope(
    p_namespace TEXT,
    p_external_id TEXT,
    p_context_id UUID,
    p_include_descendants BOOLEAN
)
RETURNS BOOLEAN AS $$
    SELECT EXISTS (
        SELECT 1
        FROM health_context_aliases alias
        WHERE alias.namespace = p_namespace
          AND alias.external_id = p_external_id
          AND (
              health_context_alias_in_scope(
                  p_namespace,
                  p_external_id,
                  p_context_id,
                  p_include_descendants
              )
              OR EXISTS (
                  SELECT 1
                  FROM health_context_closure closure
                  WHERE closure.ancestor_id = alias.context_node_id
                    AND closure.descendant_id = p_context_id
              )
          )
    );
$$ LANGUAGE sql STABLE;

CREATE OR REPLACE FUNCTION sync_surveillance_region_health_context()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM sync_surveillance_health_context(
        'surveillance-region',
        NEW.id,
        'SURV-REGION-' || NEW.id::text,
        NEW.name,
        'REGION'
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sync_surveillance_district_health_context()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM sync_surveillance_health_context(
        'surveillance-district',
        NEW.id,
        'SURV-DISTRICT-' || NEW.id::text,
        NEW.name,
        'DISTRICT',
        'surveillance-region',
        NEW.region_id
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sync_surveillance_sub_county_health_context()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM sync_surveillance_health_context(
        'surveillance-sub-county',
        NEW.id,
        'SURV-SUB-COUNTY-' || NEW.id::text,
        NEW.name,
        'SUB_COUNTY',
        'surveillance-district',
        NEW.district_id
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION sync_surveillance_facility_health_context()
RETURNS TRIGGER AS $$
DECLARE
    parent_namespace TEXT;
    parent_external_id UUID;
BEGIN
    IF NEW.sub_county_id IS NOT NULL THEN
        parent_namespace := 'surveillance-sub-county';
        parent_external_id := NEW.sub_county_id;
    ELSIF NEW.district_id IS NOT NULL THEN
        parent_namespace := 'surveillance-district';
        parent_external_id := NEW.district_id;
    ELSIF NEW.region_id IS NOT NULL THEN
        parent_namespace := 'surveillance-region';
        parent_external_id := NEW.region_id;
    END IF;

    PERFORM sync_surveillance_health_context(
        'surveillance-facility',
        NEW.id,
        'SURV-FACILITY-' || NEW.id::text,
        NEW.name,
        'FACILITY',
        parent_namespace,
        parent_external_id
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER regions_sync_health_context
AFTER INSERT OR UPDATE OF name, code ON regions
FOR EACH ROW EXECUTE FUNCTION sync_surveillance_region_health_context();

CREATE TRIGGER districts_sync_health_context
AFTER INSERT OR UPDATE OF name, code, region_id ON districts
FOR EACH ROW EXECUTE FUNCTION sync_surveillance_district_health_context();

CREATE TRIGGER sub_counties_sync_health_context
AFTER INSERT OR UPDATE OF name, code, district_id ON sub_counties
FOR EACH ROW EXECUTE FUNCTION sync_surveillance_sub_county_health_context();

CREATE TRIGGER facilities_sync_health_context
AFTER INSERT OR UPDATE OF name, district_id, sub_county_id, region_id ON facilities
FOR EACH ROW EXECUTE FUNCTION sync_surveillance_facility_health_context();

DO $$
DECLARE
    item RECORD;
BEGIN
    FOR item IN SELECT id, name FROM regions ORDER BY name LOOP
        PERFORM sync_surveillance_health_context(
            'surveillance-region',
            item.id,
            'SURV-REGION-' || item.id::text,
            item.name,
            'REGION'
        );
    END LOOP;

    FOR item IN SELECT id, name, region_id FROM districts ORDER BY name LOOP
        PERFORM sync_surveillance_health_context(
            'surveillance-district',
            item.id,
            'SURV-DISTRICT-' || item.id::text,
            item.name,
            'DISTRICT',
            'surveillance-region',
            item.region_id
        );
    END LOOP;

    FOR item IN SELECT id, name, district_id FROM sub_counties ORDER BY name LOOP
        PERFORM sync_surveillance_health_context(
            'surveillance-sub-county',
            item.id,
            'SURV-SUB-COUNTY-' || item.id::text,
            item.name,
            'SUB_COUNTY',
            'surveillance-district',
            item.district_id
        );
    END LOOP;

    FOR item IN
        SELECT id, name, region_id, district_id, sub_county_id
        FROM facilities
        ORDER BY name
    LOOP
        IF item.sub_county_id IS NOT NULL THEN
            PERFORM sync_surveillance_health_context(
                'surveillance-facility',
                item.id,
                'SURV-FACILITY-' || item.id::text,
                item.name,
                'FACILITY',
                'surveillance-sub-county',
                item.sub_county_id
            );
        ELSIF item.district_id IS NOT NULL THEN
            PERFORM sync_surveillance_health_context(
                'surveillance-facility',
                item.id,
                'SURV-FACILITY-' || item.id::text,
                item.name,
                'FACILITY',
                'surveillance-district',
                item.district_id
            );
        ELSIF item.region_id IS NOT NULL THEN
            PERFORM sync_surveillance_health_context(
                'surveillance-facility',
                item.id,
                'SURV-FACILITY-' || item.id::text,
                item.name,
                'FACILITY',
                'surveillance-region',
                item.region_id
            );
        ELSE
            PERFORM sync_surveillance_health_context(
                'surveillance-facility',
                item.id,
                'SURV-FACILITY-' || item.id::text,
                item.name,
                'FACILITY'
            );
        END IF;
    END LOOP;
END;
$$;
