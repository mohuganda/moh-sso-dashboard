CREATE TABLE ihp_realm_role_system_roles (
    realm_role TEXT NOT NULL,
    system_role_id UUID NOT NULL REFERENCES ihp_system_roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (realm_role, system_role_id)
);

CREATE INDEX ihp_realm_role_system_roles_system_role_id_idx
    ON ihp_realm_role_system_roles(system_role_id);
