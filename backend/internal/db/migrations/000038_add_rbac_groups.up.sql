CREATE TABLE IF NOT EXISTS ihp_rbac_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    keycloak_group_id TEXT UNIQUE,
    path TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    display_name TEXT,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ihp_rbac_group_members (
    group_id UUID NOT NULL REFERENCES ihp_rbac_groups(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    username TEXT,
    email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS ihp_rbac_group_realm_roles (
    group_id UUID NOT NULL REFERENCES ihp_rbac_groups(id) ON DELETE CASCADE,
    realm_role TEXT NOT NULL,
    PRIMARY KEY (group_id, realm_role)
);

CREATE TABLE IF NOT EXISTS ihp_rbac_group_system_roles (
    group_id UUID NOT NULL REFERENCES ihp_rbac_groups(id) ON DELETE CASCADE,
    system_role_id UUID NOT NULL REFERENCES ihp_system_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, system_role_id)
);

CREATE TABLE IF NOT EXISTS ihp_rbac_group_permissions (
    group_id UUID NOT NULL REFERENCES ihp_rbac_groups(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES ihp_permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, permission_id)
);

CREATE INDEX IF NOT EXISTS ihp_rbac_groups_path_idx ON ihp_rbac_groups(path);
CREATE INDEX IF NOT EXISTS ihp_rbac_group_members_user_id_idx ON ihp_rbac_group_members(user_id);
CREATE INDEX IF NOT EXISTS ihp_rbac_group_members_username_idx ON ihp_rbac_group_members(username);
CREATE INDEX IF NOT EXISTS ihp_rbac_group_members_email_idx ON ihp_rbac_group_members(email);
CREATE INDEX IF NOT EXISTS ihp_rbac_group_realm_roles_role_idx ON ihp_rbac_group_realm_roles(realm_role);
CREATE INDEX IF NOT EXISTS ihp_rbac_group_system_roles_role_idx ON ihp_rbac_group_system_roles(system_role_id);
CREATE INDEX IF NOT EXISTS ihp_rbac_group_permissions_permission_idx ON ihp_rbac_group_permissions(permission_id);
