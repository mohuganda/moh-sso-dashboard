CREATE TABLE health_context_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    context_type TEXT NOT NULL CHECK (
        context_type IN (
            'NATIONAL',
            'REGION',
            'DISTRICT',
            'CITY',
            'DIVISION',
            'MUNICIPALITY',
            'COUNTY',
            'SUB_COUNTY',
            'PARISH',
            'FACILITY',
            'PROGRAM',
            'DEPARTMENT',
            'TEAM',
            'CUSTOM'
        )
    ),
    parent_id UUID REFERENCES health_context_nodes(id) ON DELETE RESTRICT,
    source TEXT NOT NULL DEFAULT 'PORTAL',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE INDEX health_context_nodes_parent_idx
    ON health_context_nodes(parent_id);
CREATE INDEX health_context_nodes_type_enabled_idx
    ON health_context_nodes(context_type, enabled);

CREATE TABLE health_context_closure (
    ancestor_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE CASCADE,
    depth INTEGER NOT NULL CHECK (depth >= 0),
    PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX health_context_closure_descendant_idx
    ON health_context_closure(descendant_id, depth);

CREATE TABLE health_context_aliases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    context_node_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE CASCADE,
    namespace TEXT NOT NULL,
    external_id TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (namespace, external_id),
    UNIQUE (context_node_id, namespace)
);

CREATE TABLE user_health_context_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    context_node_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE CASCADE,
    scope_mode TEXT NOT NULL DEFAULT 'NODE_AND_DESCENDANTS' CHECK (
        scope_mode IN ('NODE_ONLY', 'NODE_AND_DESCENDANTS')
    ),
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    source TEXT NOT NULL DEFAULT 'PORTAL',
    source_reference TEXT,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (valid_until IS NULL OR valid_from IS NULL OR valid_until > valid_from)
);

CREATE UNIQUE INDEX user_health_context_assignment_unique_idx
    ON user_health_context_assignments (
        user_id,
        context_node_id,
        source,
        COALESCE(source_reference, '')
    );
CREATE UNIQUE INDEX user_health_context_default_unique_idx
    ON user_health_context_assignments(user_id)
    WHERE is_default;
CREATE INDEX user_health_context_assignment_user_idx
    ON user_health_context_assignments(user_id);

CREATE TABLE group_health_context_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES ihp_rbac_groups(id) ON DELETE CASCADE,
    context_node_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE CASCADE,
    scope_mode TEXT NOT NULL DEFAULT 'NODE_AND_DESCENDANTS' CHECK (
        scope_mode IN ('NODE_ONLY', 'NODE_AND_DESCENDANTS')
    ),
    source TEXT NOT NULL DEFAULT 'PORTAL',
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, context_node_id)
);

CREATE INDEX group_health_context_assignment_group_idx
    ON group_health_context_assignments(group_id);

CREATE TABLE user_active_health_contexts (
    user_id TEXT PRIMARY KEY,
    context_node_id UUID NOT NULL REFERENCES health_context_nodes(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION health_context_insert_closure()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO health_context_closure (ancestor_id, descendant_id, depth)
    VALUES (NEW.id, NEW.id, 0);

    IF NEW.parent_id IS NOT NULL THEN
        INSERT INTO health_context_closure (ancestor_id, descendant_id, depth)
        SELECT ancestor_id, NEW.id, depth + 1
        FROM health_context_closure
        WHERE descendant_id = NEW.parent_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER health_context_nodes_insert_closure
AFTER INSERT ON health_context_nodes
FOR EACH ROW EXECUTE FUNCTION health_context_insert_closure();

INSERT INTO health_context_nodes (code, name, context_type, source)
VALUES ('UG', 'Uganda', 'NATIONAL', 'SYSTEM')
ON CONFLICT (code) DO NOTHING;
