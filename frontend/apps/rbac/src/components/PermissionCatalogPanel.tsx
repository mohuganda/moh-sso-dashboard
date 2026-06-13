import { Button, TextInput } from "@carbon/react";

import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { RbacPermission } from "@moh-sso/types";

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
  return (
    <section>
      <h3>Permission Catalog</h3>
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
    </section>
  );
}
