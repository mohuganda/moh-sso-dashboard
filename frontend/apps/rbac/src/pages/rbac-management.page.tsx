import { useMemo, useState } from "react";
import {
  Button,
  Checkbox,
  DataTable,
  Dropdown,
  InlineLoading,
  InlineNotification,
  Tag,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  TextInput,
  Toggle,
} from "@carbon/react";
import { Add, Renew, Save, TrashCan } from "@carbon/react/icons";

import {
  useAddRbacGroupMemberMutation,
  useAddSystemAccessRoleMutation,
  useAssignGroupPermissionMutation,
  useAssignGroupRealmRoleMutation,
  useAssignGroupSystemRoleMutation,
  useAssignRealmRolePermissionMutation,
  useAssignSystemRolePermissionMutation,
  useCreateSystemRoleMutation,
  useDeleteSystemRoleMutation,
  useGetRbacSystemQuery,
  useListRbacGroupMembersQuery,
  useListRbacGroupsQuery,
  useListRbacPermissionsQuery,
  useListRbacSystemsQuery,
  useListRealmRolePermissionsQuery,
  usePreviewRbacChangeMutation,
  useRemoveGroupPermissionMutation,
  useRemoveRbacGroupMemberMutation,
  useRemoveGroupRealmRoleMutation,
  useRemoveGroupSystemRoleMutation,
  useRemoveRealmRolePermissionMutation,
  useRemoveSystemAccessRoleMutation,
  useRemoveSystemRolePermissionMutation,
  useSyncRbacGroupMembersMutation,
  useUpdateRbacSystemMutation,
  useUpsertRbacGroupMutation,
} from "../api";
import { PERMISSIONS, PermissionGuard, useAuthorization } from "@moh-sso/auth";
import { useListUsersQuery, type User } from "@moh-sso/users";
import { useModal, useToast } from "@moh-sso/ui";
import type { RbacGroup, RbacGroupMember, RbacPermission, RbacSystem, RbacSystemRole } from "../types";

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

function getApiErrorMessage(error: unknown, fallback: string): string {
  if (
    typeof error === "object" &&
    error !== null &&
    "data" in error &&
    typeof (error as { data?: unknown }).data === "object" &&
    (error as { data?: unknown }).data !== null
  ) {
    const data = (error as { data: { error?: { message?: string }; message?: string } }).data;
    return data.error?.message || data.message || fallback;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return fallback;
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
  const { openModal, closeModal } = useModal();
  const toast = useToast();
  const { data: systems = [], isLoading: systemsLoading } = useListRbacSystemsQuery();
  const { data: permissions = [], isLoading: permissionsLoading } = useListRbacPermissionsQuery();
  const { data: realmRoles = [] } = useListRealmRolePermissionsQuery();
  const { data: groups = [] } = useListRbacGroupsQuery();
  const { data: users = [] } = useListUsersQuery();
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
  const [selectedGroupId, setSelectedGroupId] = useState<string>("");
  const [newGroupPath, setNewGroupPath] = useState("");
  const [newGroupDisplayName, setNewGroupDisplayName] = useState("");
  const [groupRealmRoleName, setGroupRealmRoleName] = useState("");
  const [groupSystemClientId, setGroupSystemClientId] = useState("");
  const [groupSystemRoleName, setGroupSystemRoleName] = useState("");
  const [selectedGroupMemberUserId, setSelectedGroupMemberUserId] = useState("");
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
  const [upsertGroup, { isLoading: creatingGroup }] = useUpsertRbacGroupMutation();
  const [assignGroupPermission] = useAssignGroupPermissionMutation();
  const [removeGroupPermission] = useRemoveGroupPermissionMutation();
  const [assignGroupRealmRole] = useAssignGroupRealmRoleMutation();
  const [removeGroupRealmRole] = useRemoveGroupRealmRoleMutation();
  const [assignGroupSystemRole] = useAssignGroupSystemRoleMutation();
  const [removeGroupSystemRole] = useRemoveGroupSystemRoleMutation();
  const [addGroupMember, { isLoading: addingGroupMember }] = useAddRbacGroupMemberMutation();
  const [removeGroupMember, { isLoading: removingGroupMember }] = useRemoveRbacGroupMemberMutation();
  const [syncGroupMembers, { isLoading: syncingGroupMembers }] = useSyncRbacGroupMembersMutation();

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

  const selectedGroup = useMemo(
    () => groups.find((group) => group.id === selectedGroupId) ?? groups[0],
    [groups, selectedGroupId],
  );
  const selectedGroupIdForMembers = selectedGroup?.id ?? "";
  const {
    data: groupMembers = [],
    isLoading: groupMembersLoading,
    isFetching: groupMembersFetching,
  } = useListRbacGroupMembersQuery(selectedGroupIdForMembers, {
    skip: !selectedGroupIdForMembers,
  });

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
  const groupPermissionKeys = new Set(
    selectedGroup?.permissions.map((permission) => permission.key) ?? [],
  );
  const groupMemberUserIds = useMemo(
    () => new Set(groupMembers.map((member) => member.userId)),
    [groupMembers],
  );
  const assignableGroupUsers = useMemo(
    () => users.filter((user) => !groupMemberUserIds.has(user.id)),
    [groupMemberUserIds, users],
  );
  const groupMemberRows = useMemo(
    () =>
      groupMembers.map((member) => ({
        id: member.userId,
        username: member.username || "—",
        email: member.email || "—",
        userId: member.userId,
      })),
    [groupMembers],
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

  const handleCreateGroup = async () => {
    const path = newGroupPath.trim();
    if (!path) return;
    try {
      setError(null);
      const group = await upsertGroup({
        path,
        displayName: newGroupDisplayName,
        enabled: true,
      }).unwrap();
      setSelectedGroupId(group.id);
      setNewGroupPath("");
      setNewGroupDisplayName("");
    } catch {
      setError("Unable to create or update group.");
    }
  };

  const toggleGroupPermission = async (permission: RbacPermission, checked: boolean) => {
    if (!selectedGroup) return;
    try {
      setError(null);
      if (checked) {
        await assignGroupPermission({
          groupId: selectedGroup.id,
          data: { permissionKey: permission.key },
        }).unwrap();
      } else {
        await removeGroupPermission({
          groupId: selectedGroup.id,
          permissionKey: permission.key,
        }).unwrap();
      }
    } catch {
      setError("Unable to update group permissions.");
    }
  };

  const handleAssignGroupRealmRole = async () => {
    if (!selectedGroup || !groupRealmRoleName.trim()) return;
    try {
      setError(null);
      await assignGroupRealmRole({
        groupId: selectedGroup.id,
        data: { realmRole: groupRealmRoleName },
      }).unwrap();
      setGroupRealmRoleName("");
    } catch {
      setError("Unable to assign group realm role.");
    }
  };

  const handleRemoveGroupRealmRole = async (realmRole: string) => {
    if (!selectedGroup) return;
    try {
      setError(null);
      await removeGroupRealmRole({ groupId: selectedGroup.id, realmRole }).unwrap();
    } catch {
      setError("Unable to remove group realm role.");
    }
  };

  const handleAssignGroupSystemRole = async () => {
    if (!selectedGroup || !groupSystemClientId.trim() || !groupSystemRoleName.trim()) return;
    try {
      setError(null);
      await assignGroupSystemRole({
        groupId: selectedGroup.id,
        data: { clientId: groupSystemClientId, roleName: groupSystemRoleName },
      }).unwrap();
      setGroupSystemClientId("");
      setGroupSystemRoleName("");
    } catch {
      setError("Unable to assign group system role.");
    }
  };

  const handleRemoveGroupSystemRole = async (group: RbacGroup, clientId: string, roleName: string) => {
    try {
      setError(null);
      await removeGroupSystemRole({ groupId: group.id, clientId, roleName }).unwrap();
    } catch {
      setError("Unable to remove group system role.");
    }
  };

  const notifyGroupMemberSync = (warnings?: string[]) => {
    if (warnings?.length) {
      toast.warning("Group updated with warnings", warnings.join(" "));
    }
  };

  const handleAddGroupMember = async () => {
    if (!selectedGroup || !selectedGroupMemberUserId) return;
    const user = users.find((item) => item.id === selectedGroupMemberUserId);

    try {
      setError(null);
      const result = await addGroupMember({
        groupId: selectedGroup.id,
        data: { userId: selectedGroupMemberUserId },
      }).unwrap();
      setSelectedGroupMemberUserId("");
      toast.success(
        "Group member added",
        `${user?.username ?? "User"} was added to ${selectedGroup.displayName || selectedGroup.name}.`,
      );
      notifyGroupMemberSync(result.warnings);
    } catch (err) {
      const message = getApiErrorMessage(err, "Unable to add this user to the group.");
      setError(message);
      toast.error("Unable to add group member", message);
    }
  };

  const handleRemoveGroupMember = (member: RbacGroupMember) => {
    if (!selectedGroup) return;

    openModal({
      title: "Remove group member",
      content: (
        <p>
          Remove <strong>{member.username || member.email || member.userId}</strong> from{" "}
          <strong>{selectedGroup.displayName || selectedGroup.name}</strong>?
        </p>
      ),
      primaryAction: {
        label: "Remove",
        kind: "danger",
        onClick: async () => {
          try {
            setError(null);
            const result = await removeGroupMember({
              groupId: selectedGroup.id,
              userId: member.userId,
            }).unwrap();
            closeModal();
            toast.success("Group member removed", "The user's inherited access was refreshed.");
            notifyGroupMemberSync(result.warnings);
          } catch (err) {
            const message = getApiErrorMessage(err, "Unable to remove this user from the group.");
            setError(message);
            toast.error("Unable to remove group member", message);
          }
        },
      },
      secondaryAction: {
        label: "Cancel",
        onClick: closeModal,
      },
      onClose: closeModal,
    });
  };

  const handleSyncGroupMembers = async () => {
    if (!selectedGroup) return;

    try {
      setError(null);
      const result = await syncGroupMembers(selectedGroup.id).unwrap();
      toast.success(
        "Group members synced",
        `${result.memberCount} member${result.memberCount === 1 ? "" : "s"} synced from Keycloak.`,
      );
      notifyGroupMemberSync(result.warnings);
    } catch (err) {
      const message = getApiErrorMessage(err, "Unable to sync group members from Keycloak.");
      setError(message);
      toast.error("Unable to sync group members", message);
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

          <section className="rbac-card">
            <div className="rbac-page__panel-header">
              <div>
                <h2>Groups</h2>
                <p>Map Keycloak groups to realm roles, system roles, and direct permissions.</p>
              </div>
              <div className="rbac-inline-form">
                <TextInput
                  id="rbac-new-group-path"
                  hideLabel
                  labelText="Group path"
                  placeholder="/data-statistics/users"
                  value={newGroupPath}
                  disabled={!canWriteRoles}
                  onChange={(event) => setNewGroupPath(event.target.value)}
                />
                <TextInput
                  id="rbac-new-group-display"
                  hideLabel
                  labelText="Group display name"
                  placeholder="Data statistics users"
                  value={newGroupDisplayName}
                  disabled={!canWriteRoles}
                  onChange={(event) => setNewGroupDisplayName(event.target.value)}
                />
                <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                  <Button
                    renderIcon={Add}
                    size="sm"
                    disabled={creatingGroup || !newGroupPath.trim()}
                    onClick={handleCreateGroup}
                  >
                    Save group
                  </Button>
                </PermissionGuard>
              </div>
            </div>

            <div className="rbac-realm-layout">
              <div className="rbac-role-list">
                {groups.map((group) => (
                  <button
                    className={`rbac-role-list__item ${group.id === selectedGroup?.id ? "is-active" : ""}`}
                    key={group.id}
                    onClick={() => setSelectedGroupId(group.id)}
                    type="button"
                  >
                    <span>{group.displayName || group.name}</span>
                    <small>
                      {group.path} · {group.memberCount} members
                    </small>
                  </button>
                ))}
              </div>

              <div className="rbac-group-detail">
                {selectedGroup ? (
                  <>
                    <div className="rbac-page__panel-header">
                      <div>
                        <h3>{selectedGroup.displayName || selectedGroup.name}</h3>
                        <p>{selectedGroup.path}</p>
                      </div>
                      <Tag type={selectedGroup.enabled ? "green" : "gray"}>
                        {selectedGroup.enabled ? "Enabled" : "Disabled"}
                      </Tag>
                    </div>

                    <div className="rbac-group-detail__section">
                      <div className="rbac-group-detail__section-header">
                        <div>
                          <h4>Members</h4>
                          <p>
                            Manage Keycloak group membership. Changes are written to Keycloak and
                            synced back into the portal RBAC cache.
                          </p>
                        </div>
                        <Button
                          kind="ghost"
                          renderIcon={Renew}
                          size="sm"
                          disabled={syncingGroupMembers || !canWriteRoles}
                          onClick={handleSyncGroupMembers}
                        >
                          Sync members
                        </Button>
                      </div>

                      {!selectedGroup.keycloakGroupId && (
                        <InlineNotification
                          kind="warning"
                          lowContrast
                          title="Keycloak link missing"
                          subtitle="This group does not have a cached Keycloak group ID. The backend will try to match it by path; if that fails, run live RBAC sync or create the group in Keycloak first."
                        />
                      )}

                      <div className="rbac-inline-form rbac-group-members__add">
                        <Dropdown<User>
                          id="rbac-group-member-user"
                          titleText="Add user"
                          label="Select user"
                          items={assignableGroupUsers}
                          selectedItem={
                            assignableGroupUsers.find(
                              (user) => user.id === selectedGroupMemberUserId,
                            ) ?? undefined
                          }
                          itemToString={(user) =>
                            user ? `${user.username}${user.email ? ` · ${user.email}` : ""}` : ""
                          }
                          disabled={!canWriteRoles || addingGroupMember}
                          onChange={({ selectedItem }) =>
                            setSelectedGroupMemberUserId(selectedItem?.id ?? "")
                          }
                        />
                        <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                          <Button
                            size="md"
                            renderIcon={Add}
                            disabled={!selectedGroupMemberUserId || addingGroupMember}
                            onClick={handleAddGroupMember}
                          >
                            Add member
                          </Button>
                        </PermissionGuard>
                      </div>

                      {(groupMembersLoading || groupMembersFetching) && (
                        <InlineLoading description="Loading group members..." />
                      )}

                      <DataTable
                        rows={groupMemberRows}
                        headers={[
                          { key: "username", header: "Username" },
                          { key: "email", header: "Email" },
                          { key: "actions", header: "" },
                        ]}
                      >
                        {({ rows, headers, getHeaderProps, getRowProps }) => (
                          <TableContainer className="rbac-group-members-table">
                            <Table size="lg">
                              <TableHead>
                                <TableRow>
                                  {headers.map((header) => {
                                    const { key, ...headerProps } = getHeaderProps({ header });
                                    return (
                                      <TableHeader key={key} {...headerProps}>
                                        {header.header}
                                      </TableHeader>
                                    );
                                  })}
                                </TableRow>
                              </TableHead>
                              <TableBody>
                                {rows.length === 0 ? (
                                  <TableRow>
                                    <TableCell colSpan={headers.length}>
                                      <div className="rbac-group-members-table__empty">
                                        No synced members for this group.
                                      </div>
                                    </TableCell>
                                  </TableRow>
                                ) : (
                                  rows.map((row) => {
                                    const member = groupMembers.find(
                                      (item) => item.userId === row.id,
                                    );
                                    const { key, ...rowProps } = getRowProps({ row });

                                    return (
                                      <TableRow key={key} {...rowProps}>
                                        {row.cells.map((cell) => {
                                          if (cell.info.header === "actions") {
                                            return (
                                              <TableCell key={cell.id}>
                                                <PermissionGuard
                                                  permission={PERMISSIONS.rbacRolesWrite}
                                                >
                                                  <Button
                                                    hasIconOnly
                                                    iconDescription="Remove group member"
                                                    kind="ghost"
                                                    renderIcon={TrashCan}
                                                    size="sm"
                                                    disabled={!member || removingGroupMember}
                                                    onClick={() => {
                                                      if (member) handleRemoveGroupMember(member);
                                                    }}
                                                  />
                                                </PermissionGuard>
                                              </TableCell>
                                            );
                                          }

                                          return <TableCell key={cell.id}>{cell.value}</TableCell>;
                                        })}
                                      </TableRow>
                                    );
                                  })
                                )}
                              </TableBody>
                            </Table>
                          </TableContainer>
                        )}
                      </DataTable>
                    </div>

                    <div className="rbac-group-detail__section">
                      <h4>Realm roles</h4>
                      <div className="rbac-inline-form">
                        <TextInput
                          id="rbac-group-realm-role"
                          hideLabel
                          labelText="Realm role"
                          placeholder="user"
                          value={groupRealmRoleName}
                          disabled={!canWriteRoles}
                          onChange={(event) => setGroupRealmRoleName(event.target.value)}
                        />
                        <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                          <Button size="sm" renderIcon={Add} onClick={handleAssignGroupRealmRole}>
                            Add
                          </Button>
                        </PermissionGuard>
                      </div>
                      <div className="rbac-tag-list">
                        {selectedGroup.realmRoles.map((role) => (
                          <Tag
                            key={role}
                            type="purple"
                            filter={canWriteRoles}
                            onClose={
                              canWriteRoles
                                ? (event) => {
                                    event.preventDefault();
                                    void handleRemoveGroupRealmRole(role);
                                  }
                                : undefined
                            }
                          >
                            {role}
                          </Tag>
                        ))}
                      </div>
                    </div>

                    <div className="rbac-group-detail__section">
                      <h4>System roles</h4>
                      <div className="rbac-inline-form">
                        <TextInput
                          id="rbac-group-system-client"
                          hideLabel
                          labelText="Client ID"
                          placeholder="data-statistics"
                          value={groupSystemClientId}
                          disabled={!canWriteRoles}
                          onChange={(event) => setGroupSystemClientId(event.target.value)}
                        />
                        <TextInput
                          id="rbac-group-system-role"
                          hideLabel
                          labelText="Role name"
                          placeholder="data-statistics_access"
                          value={groupSystemRoleName}
                          disabled={!canWriteRoles}
                          onChange={(event) => setGroupSystemRoleName(event.target.value)}
                        />
                        <PermissionGuard permission={PERMISSIONS.rbacRolesWrite}>
                          <Button size="sm" renderIcon={Add} onClick={handleAssignGroupSystemRole}>
                            Add
                          </Button>
                        </PermissionGuard>
                      </div>
                      <div className="rbac-tag-list">
                        {selectedGroup.systemRoles.map((role) => (
                          <Tag
                            key={`${role.clientId}:${role.roleName}`}
                            type="cyan"
                            filter={canWriteRoles}
                            onClose={
                              canWriteRoles
                                ? (event) => {
                                    event.preventDefault();
                                    void handleRemoveGroupSystemRole(
                                      selectedGroup,
                                      role.clientId,
                                      role.roleName,
                                    );
                                  }
                                : undefined
                            }
                          >
                            {role.clientId}:{role.roleName}
                          </Tag>
                        ))}
                      </div>
                    </div>

                    <div className="rbac-group-detail__section">
                      <h4>Direct permissions</h4>
                      <PermissionMatrix
                        groupedPermissions={groupedPermissions}
                        selectedKeys={groupPermissionKeys}
                        disabled={!canWritePermissions}
                        onToggle={toggleGroupPermission}
                      />
                    </div>
                  </>
                ) : (
                  <p>No groups have been synced or created yet.</p>
                )}
              </div>
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
