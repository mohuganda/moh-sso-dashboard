import { Button, TextInput } from "@carbon/react";
import { useMemo, useState } from "react";

import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { RbacPermission } from "../types";

type PermissionCatalogPanelProps = {
  permissions: RbacPermission[];
  selectedPermission: string;
  selectedPermissionData?: RbacPermission;
  permissionDescription: string;
  onSelectedPermissionChange: (value: string) => void;
  onPermissionDescriptionChange: (value: string) => void;
  onSave: () => void;
};

export function PermissionCatalogPanel({
  permissions,
  selectedPermission,
  selectedPermissionData,
  permissionDescription,
  onSelectedPermissionChange,
  onPermissionDescriptionChange,
  onSave,
}: PermissionCatalogPanelProps) {
  const [search, setSearch] = useState("");
  const groupedPermissions = useMemo(() => {
    const query = search.trim().toLowerCase();
    return permissions
      .filter((permission) => {
        if (!query) return true;
        return [permission.key, permission.displayName, permission.description, permission.category, permission.status]
          .filter(Boolean)
          .some((value) => value?.toLowerCase().includes(query));
      })
      .reduce<Record<string, RbacPermission[]>>((groups, permission) => {
        const category = permission.category || permission.key.split(":")[0] || "other";
        groups[category] = [...(groups[category] ?? []), permission];
        return groups;
      }, {});
  }, [permissions, search]);

  return (
    <section>
      <h3>Permission Catalog</h3>
      <TextInput
        id="rbac-permission-search"
        labelText="Search permissions"
        placeholder="Search by key, category, description, or status"
        value={search}
        onChange={(event) => setSearch(event.target.value)}
      />
      <TextInput
        id="rbac-permission-key"
        labelText="Permission key"
        placeholder="users:read"
        value={selectedPermission}
        onChange={(event) => onSelectedPermissionChange(event.target.value)}
      />
      <TextInput
        id="rbac-permission-description"
        labelText="Description"
        value={permissionDescription}
        onChange={(event) => onPermissionDescriptionChange(event.target.value)}
      />
      {selectedPermissionData && (
        <small>
          {selectedPermissionData.displayName || selectedPermissionData.key} ·{" "}
          {selectedPermissionData.status || "active"}
        </small>
      )}
      <PermissionGuard permission={PERMISSIONS.rbacPermissionsWrite}>
        <Button size="sm" onClick={onSave}>
          Save metadata
        </Button>
      </PermissionGuard>
      <small>{permissions.length} permissions registered</small>
      <div className="rbac-permission-catalog">
        {Object.entries(groupedPermissions).map(([category, items]) => (
          <div className="rbac-permission-catalog__group" key={category}>
            <h4>{category}</h4>
            {items.map((permission) => (
              <button
                className={`rbac-permission-catalog__row ${permission.key === selectedPermission ? "is-active" : ""}`}
                key={permission.key}
                onClick={() => {
                  onSelectedPermissionChange(permission.key);
                  onPermissionDescriptionChange(permission.description || "");
                }}
                type="button"
              >
                <strong>{permission.displayName || permission.key}</strong>
                <small>{permission.key}</small>
                <small>{permission.description || "No description"}</small>
                <small>
                  System roles: {permission.systemRoleUsageCount ?? 0} · Realm roles:{" "}
                  {permission.realmRoleUsageCount ?? 0} · {permission.status || "active"}
                </small>
              </button>
            ))}
          </div>
        ))}
      </div>
    </section>
  );
}
