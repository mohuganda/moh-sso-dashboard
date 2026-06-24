CREATE TABLE ihp_systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    launch_url TEXT,
    category TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ihp_system_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL REFERENCES ihp_systems(id) ON DELETE CASCADE,
    role_name TEXT NOT NULL,
    display_name TEXT,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (system_id, role_name)
);

CREATE TABLE ihp_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    permission_key TEXT UNIQUE NOT NULL,
    display_name TEXT,
    description TEXT,
    category TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ihp_system_role_permissions (
    system_role_id UUID NOT NULL REFERENCES ihp_system_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES ihp_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (system_role_id, permission_id)
);

CREATE TABLE ihp_realm_role_permissions (
    realm_role TEXT NOT NULL,
    permission_id UUID NOT NULL REFERENCES ihp_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (realm_role, permission_id)
);

CREATE TABLE ihp_system_access_roles (
    system_id UUID NOT NULL REFERENCES ihp_systems(id) ON DELETE CASCADE,
    role_name TEXT NOT NULL,
    PRIMARY KEY (system_id, role_name)
);

CREATE INDEX ihp_system_roles_system_id_idx ON ihp_system_roles(system_id);
CREATE INDEX ihp_system_role_permissions_permission_id_idx ON ihp_system_role_permissions(permission_id);
CREATE INDEX ihp_realm_role_permissions_permission_id_idx ON ihp_realm_role_permissions(permission_id);

