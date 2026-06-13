import { useMemo, useState } from "react";
import {
  Button,
  Checkbox,
  InlineLoading,
  InlineNotification,
  Tag,
  TextInput,
  Toggle,
} from "@carbon/react";
import { Add, Save, TrashCan } from "@carbon/react/icons";

import {
  useAddSystemAccessRoleMutation,
  useAssignRealmRolePermissionMutation,
  useAssignSystemRolePermissionMutation,
  useCreateSystemRoleMutation,
  useDeleteSystemRoleMutation,
  useGetRbacSystemQuery,
  useListRbacPermissionsQuery,
  useListRbacSystemsQuery,
  useListRealmRolePermissionsQuery,
  usePreviewRbacChangeMutation,
  useRemoveRealmRolePermissionMutation,
  useRemoveSystemAccessRoleMutation,
  useRemoveSystemRolePermissionMutation,
  useUpdateRbacSystemMutation,
} from "@moh-sso/api";
import { PERMISSIONS, PermissionGuard, useAuthorization } from "@moh-sso/auth";
import type { RbacPermission, RbacSystem, RbacSystemRole } from "@moh-sso/types";

import { EffectiveAccessPanel } from "../components/EffectiveAccessPanel";
import { GovernanceToolsPanel } from "../components/GovernanceToolsPanel";
import { SyncDriftPanel } from "../components/SyncDriftPanel";
import "./rbac-management.page.scss";

type SystemDraft = {
  displayName: string;
  description: string;
  icon: string;
  launchUrl: string;
  category: string;
  ownerTeam: string;
  ownerName: string;
  ownerEmail: string;
  supportUrl: string;
  documentationUrl: string;
  environment: string;
  criticality: string;
  enabled: boolean;
  sortOrder: number;
};

function permissionLabel(permission: RbacPermission) {
  return permission.displayName || permission.key;
}

function permissionCategory(permission: RbacPermission) {
  return permission.category || permission.key.split(":")[0] || "other";
}

function toDraft(system?: RbacSystem): SystemDraft {
  return {
    displayName: system?.displayName ?? "",
    description: system?.description ?? "",
    icon: system?.icon ?? "",
    launchUrl: system?.launchUrl ?? "",
    category: system?.category ?? "",
    ownerTeam: system?.ownerTeam ?? "",
    ownerName: system?.ownerName ?? "",
    ownerEmail: system?.ownerEmail ?? "",
    supportUrl: system?.supportUrl ?? "",
    documentationUrl: system?.documentationUrl ?? "",
    environment: system?.environment ?? "",
    criticality: system?.criticality ?? "",
    enabled: system?.enabled ?? true,
    sortOrder: system?.sortOrder ?? 0,
  };
}

export default function RbacManagementPage() {
  const { can } = useAuthorization();
  const { data: systems = [], isLoading: systemsLoading } = useListRbacSystemsQuery();
  const { data: permissions = [], isLoading: permissionsLoading } = useListRbacPermissionsQuery();
  const { data: realmRoles = [] } = useListRealmRolePermissionsQuery();
  const [selectedClientId, setSelectedClientId] = useState<string>("");
  const activeClientId = selectedClientId || systems[0]?.clientId || "";
  const { data: systemDetail, isFetching: systemLoading } = useGetRbacSystemQuery(activeClientId, {
    skip: !activeClientId,
  });

  const [draft, setDraft] = useState<SystemDraft>(() => toDraft());
  const [draftClientId, setDraftClientId] = useState<string>("");
  const [newRoleName, setNewRoleName] = useState("");
  const [newRoleDisplayName, setNewRoleDisplayName] = useState("");
  const [newAccessRole, setNewAccessRole] = useState("");
  const [realmRoleName, setRealmRoleName] = useState("");
  const [selectedRoleId, setSelectedRoleId] = useState<string>("");
  const [selectedRealmRole, setSelectedRealmRole] = useState<string>("");
  const [error, setError] = useState<string | null>(null);

  const [updateSystem, { isLoading: savingSystem }] = useUpdateRbacSystemMutation();
  const [createRole, { isLoading: creatingRole }] = useCreateSystemRoleMutation();
  const [deleteRole] = useDeleteSystemRoleMutation();
  const [assignRolePermission] = useAssignSystemRolePermissionMutation();
  const [removeRolePermission] = useRemoveSystemRolePermissionMutation();
  const [assignRealmPermission] = useAssignRealmRolePermissionMutation();
  const [removeRealmPermission] = useRemoveRealmRolePermissionMutation();
  const [addAccessRole] = useAddSystemAccessRoleMutation();
  const [removeAccessRole] = useRemoveSystemAccessRoleMutation();
  const [previewChange] = usePreviewRbacChangeMutation();

  const canWriteSystems = can(PERMISSIONS.rbacWrite);
  const canWriteRoles = can(PERMISSIONS.rbacRolesWrite);
  const canWritePermissions = can(PERMISSIONS.rbacPermissionsWrite);

  const selectedRole = useMemo(
    () => systemDetail?.roles.find((role) => role.id === selectedRoleId) ?? systemDetail?.roles[0],
    [selectedRoleId, systemDetail?.roles],
  );

  const selectedRealmRoleGroup = useMemo(
    () =>
      realmRoles.find((group) => group.realmRole === selectedRealmRole) ??
      realmRoles.find((group) => group.realmRole === realmRoleName),
    [realmRoleName, realmRoles, selectedRealmRole],
  );

  const groupedPermissions = useMemo(() => {
    return permissions.reduce<Record<string, RbacPermission[]>>((groups, permission) => {
      const category = permissionCategory(permission);
      groups[category] = [...(groups[category] ?? []), permission];
      return groups;
    }, {});
  }, [permissions]);

  const rolePermissionKeys = new Set(selectedRole?.permissions.map((permission) => permission.key) ?? []);
  const realmPermissionKeys = new Set(
    selectedRealmRoleGroup?.permissions.map((permission) => permission.key) ?? [],
  );

  const handleSelectSystem = (clientId: string) => {
    const system = systems.find((item) => item.clientId === clientId);
    setSelectedClientId(clientId);
    setSelectedRoleId("");
    setDraftClientId(clientId);
    setDraft(toDraft(system));
  };

  const handleSaveSystem = async () => {
    const clientId = draftClientId || activeClientId;
    if (!clientId) return;
    try {
      setError(null);
      await updateSystem({ clientId, data: draft }).unwrap();
    } catch {
      setError("Unable to save system metadata.");
    }
  };

  const handleCreateRole = async () => {
    if (!activeClientId || !newRoleName.trim()) return;
    try {
      setError(null);
      await createRole({
        clientId: activeClientId,
        data: { name: newRoleName, displayName: newRoleDisplayName, enabled: true },
      }).unwrap();
      setNewRoleName("");
      setNewRoleDisplayName("");
    } catch {
      setError("Unable to create system role.");
    }
  };

  const handleDeleteRole = async (role: RbacSystemRole) => {
    const preview = await previewChange({
      action: "delete-role",
      resourceType: "system-role",
      resourceId: role.id,
      systemClientId: role.clientId,
      roleName: role.name,
    }).unwrap();
    const warning = preview.warnings.length ? `\n\n${preview.warnings.join("\n")}` : "";
    if (!confirm(`Delete role "${role.name}"? Risk: ${preview.riskLevel}.${warning}`)) return;
    try {
      setError(null);
      await deleteRole({ roleId: role.id, clientId: role.clientId }).unwrap();
    } catch {
      setError("Unable to delete system role.");
    }
  };

  const toggleRolePermission = async (permission: RbacPermission, checked: boolean) => {
    if (!selectedRole) return;
    try {
      setError(null);
      if (checked) {
        await assignRolePermission({
          roleId: selectedRole.id,
          clientId: selectedRole.clientId,
          data: { permissionKey: permission.key },
        }).unwrap();
      } else {
        await removeRolePermission({
          roleId: selectedRole.id,
          clientId: selectedRole.clientId,
          permissionKey: permission.key,
        }).unwrap();
      }
    } catch {
      setError("Unable to update role permissions.");
    }
  };

  const toggleRealmPermission = async (permission: RbacPermission, checked: boolean) => {
    const realmRole = selectedRealmRole || realmRoleName.trim();
    if (!realmRole) return;
    try {
      setError(null);
      if (checked) {
        await assignRealmPermission({ realmRole, data: { permissionKey: permission.key } }).unwrap();
      } else {
        await removeRealmPermission({ realmRole, permissionKey: permission.key }).unwrap();
      }
      setSelectedRealmRole(realmRole);
    } catch {
      setError("Unable to update realm role permissions.");
    }
  };

  const handleAddAccessRole = async () => {
    if (!activeClientId || !newAccessRole.trim()) return;
    try {
      setError(null);
      await addAccessRole({ clientId: activeClientId, data: { roleName: newAccessRole } }).unwrap();
      setNewAccessRole("");
    } catch {
      setError("Unable to add access role.");
    }
  };

  const handleRemoveAccessRole = async (roleName: string) => {
    if (!activeClientId) return;
    try {
      setError(null);
      await removeAccessRole({ clientId: activeClientId, roleName }).unwrap();
    } catch {
      setError("Unable to remove access role. Use backend force mode only after review.");
    }
  };

  if (systemsLoading || permissionsLoading) {
    return <InlineLoading description="Loading RBAC mappings..." />;
  }

  return (
    <div className="rbac-page">
      <header className="rbac-page__header">
        <div>
          <h1>RBAC Management</h1>
          <p>Manage portal systems, system roles, access roles, and permission mappings.</p>
        </div>
      </header>

      {error && (
        <InlineNotification
          kind="error"
          lowContrast
          title="RBAC update failed"
          subtitle={error}
          onClose={() => setError(null)}
        />
      )}

      <SyncDriftPanel />
      <EffectiveAccessPanel />
      <GovernanceToolsPanel activeClientId={activeClientId} />

      <section className="rbac-page__layout">
        <aside className="rbac-page__systems">
          <div className="rbac-page__panel-header">
            <h2>Systems</h2>
            <Tag type="blue">{systems.length}</Tag>
          </div>
          <div className="rbac-system-list">
            {systems.map((system) => (
              <button
                className={`rbac-system-list__item ${system.clientId === activeClientId ? "is-active" : ""}`}
                key={system.clientId}
                onClick={() => handleSelectSystem(system.clientId)}
                type="button"
              >
                <span>{system.displayName}</span>
                <small>{system.clientId}</small>
              </button>
            ))}
          </div>
        </aside>

        <main className="rbac-page__main">
          {systemLoading && <InlineLoading description="Loading system details..." />}

          <section className="rbac-card">
            <div className="rbac-page__panel-header">
              <h2>System Metadata</h2>
              <PermissionGuard permission={PERMISSIONS.rbacWrite}>
                <Button
                  renderIcon={Save}
                  size="sm"
                  disabled={!activeClientId || savingSystem}
                  onClick={handleSaveSystem}
                >
                  Save system
                </Button>
              </PermissionGuard>
            </div>
            <div className="rbac-form-grid">
              <TextInput
                id="rbac-client-id"
                labelText="Client ID"
                value={draftClientId || activeClientId}
                disabled={!canWriteSystems}
                onChange={(event) => setDraftClientId(event.target.value)}
              />
              <TextInput
                id="rbac-display-name"
                labelText="Display name"
                value={draft.displayName || systemDetail?.displayName || ""}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, displayName: event.target.value })}
              />
              <TextInput
                id="rbac-category"
                labelText="Category"
                value={draft.category}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, category: event.target.value })}
              />
              <TextInput
                id="rbac-launch-url"
                labelText="Launch URL"
                value={draft.launchUrl}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, launchUrl: event.target.value })}
              />
              <TextInput
                id="rbac-icon"
                labelText="Icon"
                value={draft.icon}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, icon: event.target.value })}
              />
              <TextInput
                id="rbac-sort-order"
                labelText="Sort order"
                type="number"
                value={String(draft.sortOrder)}
                disabled={!canWriteSystems}
                onChange={(event) =>
                  setDraft({ ...draft, sortOrder: Number(event.target.value) || 0 })
                }
              />
              <TextInput
                id="rbac-description"
                labelText="Description"
                value={draft.description}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, description: event.target.value })}
              />
              <TextInput
                id="rbac-owner-team"
                labelText="Owner team"
                value={draft.ownerTeam}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, ownerTeam: event.target.value })}
              />
              <TextInput
                id="rbac-owner-name"
                labelText="Owner name"
                value={draft.ownerName}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, ownerName: event.target.value })}
              />
              <TextInput
                id="rbac-owner-email"
                labelText="Owner email"
                value={draft.ownerEmail}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, ownerEmail: event.target.value })}
              />
              <TextInput
                id="rbac-support-url"
                labelText="Support URL"
                value={draft.supportUrl}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, supportUrl: event.target.value })}
              />
              <TextInput
                id="rbac-documentation-url"
                labelText="Documentation URL"
                value={draft.documentationUrl}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, documentationUrl: event.target.value })}
              />
              <TextInput
                id="rbac-environment"
                labelText="Environment"
                value={draft.environment}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, environment: event.target.value })}
              />
              <TextInput
                id="rbac-criticality"
                labelText="Criticality"
                value={draft.criticality}
                disabled={!canWriteSystems}
                onChange={(event) => setDraft({ ...draft, criticality: event.target.value })}
              />
              <Toggle
                id="rbac-enabled"
                labelText="Enabled"
                toggled={draft.enabled}
                disabled={!canWriteSystems}
                onToggle={(enabled) => setDraft({ ...draft, enabled })}
              />
            </div>
          </section>

          <section className="rbac-card">
            <div className="rbac-page__panel-header">
              <h2>Access Roles</h2>
              <div className="rbac-inline-form">
                <TextInput
                  id="rbac-new-access-role"
                  hideLabel
                  labelText="Access role"
                  placeholder="viewer"
                  value={newAccessRole}
                  disabled={!canWriteRoles}
                  onChange={(event) => setNewAccessRole(event.target.value)}
                />
                <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                  <Button renderIcon={Add} size="sm" onClick={handleAddAccessRole}>
                    Add
                  </Button>
                </PermissionGuard>
              </div>
            </div>
            <div className="rbac-tag-list">
              {(systemDetail?.accessRoles ?? []).map((role) => (
                <Tag
                  key={role}
                  type="cyan"
                  filter={canWriteRoles}
                  onClose={
                    canWriteRoles
                      ? (event) => {
                          event.preventDefault();
                          void handleRemoveAccessRole(role);
                        }
                      : undefined
                  }
                >
                  {role}
                </Tag>
              ))}
            </div>
          </section>

          <section className="rbac-card">
            <div className="rbac-page__panel-header">
              <h2>System Roles</h2>
              <div className="rbac-inline-form">
                <TextInput
                  id="rbac-new-role-name"
                  hideLabel
                  labelText="Role name"
                  placeholder="role_name"
                  value={newRoleName}
                  disabled={!canWriteRoles}
                  onChange={(event) => setNewRoleName(event.target.value)}
                />
                <TextInput
                  id="rbac-new-role-display"
                  hideLabel
                  labelText="Role display name"
                  placeholder="Display name"
                  value={newRoleDisplayName}
                  disabled={!canWriteRoles}
                  onChange={(event) => setNewRoleDisplayName(event.target.value)}
                />
                <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                  <Button
                    renderIcon={Add}
                    size="sm"
                    disabled={creatingRole}
                    onClick={handleCreateRole}
                  >
                    Create
                  </Button>
                </PermissionGuard>
              </div>
            </div>

            <div className="rbac-role-grid">
              <div className="rbac-role-list">
                {(systemDetail?.roles ?? []).map((role) => (
                  <button
                    className={`rbac-role-list__item ${role.id === selectedRole?.id ? "is-active" : ""}`}
                    key={role.id}
                    onClick={() => setSelectedRoleId(role.id)}
                    type="button"
                  >
                    <span>{role.displayName || role.name}</span>
                    <small>{role.permissions.length} permissions</small>
                    <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                      <Button
                        hasIconOnly
                        iconDescription="Delete role"
                        kind="ghost"
                        renderIcon={TrashCan}
                        size="sm"
                        onClick={(event) => {
                          event.stopPropagation();
                          void handleDeleteRole(role);
                        }}
                      />
                    </PermissionGuard>
                  </button>
                ))}
              </div>
              <PermissionMatrix
                groupedPermissions={groupedPermissions}
                selectedKeys={rolePermissionKeys}
                disabled={!selectedRole || !canWritePermissions}
                onToggle={toggleRolePermission}
              />
            </div>
          </section>

          <section className="rbac-card">
            <div className="rbac-page__panel-header">
              <h2>Realm Role Permissions</h2>
              <div className="rbac-inline-form">
                <TextInput
                  id="rbac-realm-role-name"
                  hideLabel
                  labelText="Realm role"
                  placeholder="admin"
                  value={realmRoleName}
                  disabled={!canWritePermissions}
                  onChange={(event) => setRealmRoleName(event.target.value)}
                />
              </div>
            </div>
            <div className="rbac-realm-layout">
              <div className="rbac-role-list">
                {realmRoles.map((group) => (
                  <button
                    className={`rbac-role-list__item ${group.realmRole === selectedRealmRole ? "is-active" : ""}`}
                    key={group.realmRole}
                    onClick={() => {
                      setSelectedRealmRole(group.realmRole);
                      setRealmRoleName(group.realmRole);
                    }}
                    type="button"
                  >
                    <span>{group.realmRole}</span>
                    <small>{group.permissions.length} permissions</small>
                  </button>
                ))}
              </div>
              <PermissionMatrix
                groupedPermissions={groupedPermissions}
                selectedKeys={realmPermissionKeys}
                disabled={(!selectedRealmRole && !realmRoleName.trim()) || !canWritePermissions}
                onToggle={toggleRealmPermission}
              />
            </div>
          </section>
        </main>
      </section>
    </div>
  );
}

type PermissionMatrixProps = {
  groupedPermissions: Record<string, RbacPermission[]>;
  selectedKeys: Set<string>;
  disabled?: boolean;
  onToggle: (permission: RbacPermission, checked: boolean) => Promise<void>;
};

function PermissionMatrix({
  groupedPermissions,
  selectedKeys,
  disabled = false,
  onToggle,
}: PermissionMatrixProps) {
  return (
    <div className="rbac-permission-matrix">
      {Object.entries(groupedPermissions).map(([category, permissions]) => (
        <div className="rbac-permission-group" key={category}>
          <h3>{category}</h3>
          {permissions.map((permission) => (
            <Checkbox
              id={`${category}-${permission.key}`}
              key={permission.key}
              labelText={permissionLabel(permission)}
              checked={selectedKeys.has(permission.key)}
              disabled={disabled}
              onChange={(_, { checked }) => {
                void onToggle(permission, checked);
              }}
            />
          ))}
        </div>
      ))}
    </div>
  );
}
