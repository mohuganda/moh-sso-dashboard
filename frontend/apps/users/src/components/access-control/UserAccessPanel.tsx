import { Button, InlineLoading, MultiSelect, Stack, Tag, Tile } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";

import {
  useGetUserAccessProfileQuery,
  useUpdateUserAccessMutation,
} from "@moh-sso/api";
import type { RbacAssignableSystemAccess, RbacPermission } from "@moh-sso/types";
import { FormInlineAlert, useToast } from "@moh-sso/ui";

type Props = {
  userId: string;
};

type SelectItem = {
  id: string;
  text: string;
};

export function UserAccessPanel({ userId }: Props) {
  const toast = useToast();
  const { data: profile, isLoading, isFetching, isError, error, refetch } =
    useGetUserAccessProfileQuery(userId);
  const [updateAccess, { isLoading: saving }] = useUpdateUserAccessMutation();

  const [realmRoles, setRealmRoles] = useState<string[]>([]);
  const [clientRoles, setClientRoles] = useState<Record<string, string[]>>({});

  useEffect(() => {
    if (!profile) return;

    setRealmRoles(profile.effectiveAccess.realmRoles ?? []);
    setClientRoles(profile.effectiveAccess.clientRoles ?? {});
  }, [profile]);

  const realmRoleItems = useMemo<SelectItem[]>(
    () =>
      (profile?.assignable.realmRoles ?? []).map((role) => ({
        id: role.name,
        text: role.displayName || role.name,
      })),
    [profile?.assignable.realmRoles],
  );

  const effectivePermissions = useMemo(
    () => groupPermissions(profile?.effectiveAccess.permissions ?? []),
    [profile?.effectiveAccess.permissions],
  );

  const handleSystemRolesChange = (system: RbacAssignableSystemAccess, roles: string[]) => {
    setClientRoles((current) => ({
      ...current,
      [system.clientId]: roles,
    }));
  };

  const handleSave = async () => {
    try {
      await updateAccess({
        userId,
        data: {
          realmRoles,
          clientRoles,
        },
      }).unwrap();

      toast.success("Access updated", "User roles and effective permissions were updated.");
    } catch (err) {
      toast.error("Access update failed", getApiErrorMessage(err, "Unable to update user access."));
    }
  };

  if (isLoading) {
    return <InlineLoading description="Loading user access..." />;
  }

  if (isError || !profile) {
    return (
      <Stack gap={4}>
        <FormInlineAlert
          title="Unable to load user access"
          subtitle={getApiErrorMessage(error, "Check RBAC configuration and try again.")}
        />
        <Button kind="secondary" size="sm" onClick={() => refetch()}>
          Retry
        </Button>
      </Stack>
    );
  }

  return (
    <Stack gap={5}>
      {isFetching && <InlineLoading description="Refreshing access profile..." />}

      <Tile>
        <Stack gap={4}>
          <strong>Realm roles</strong>
          <MultiSelect
            id={`user-${userId}-realm-roles`}
            titleText="Realm roles"
            label="Select realm roles"
            items={realmRoleItems}
            itemToString={(item) => item?.text ?? ""}
            selectedItems={realmRoleItems.filter((role) => realmRoles.includes(role.id))}
            onChange={({ selectedItems }) => {
              setRealmRoles((selectedItems ?? []).map((role) => role.id));
            }}
          />
        </Stack>
      </Tile>

      <Tile>
        <Stack gap={5}>
          <strong>System roles and app access</strong>
          {profile.assignable.systems.map((system) => {
            const roleItems = system.roles.map((role) => ({
              id: role.name,
              text: role.displayName || role.name,
            }));
            const selectedRoles = clientRoles[system.clientId] ?? [];

            return (
              <Stack gap={3} key={system.clientId}>
                <div>
                  <strong>{system.displayName || system.clientId}</strong>
                  <div>
                    <small>{system.clientId}</small>
                  </div>
                </div>

                <MultiSelect
                  id={`user-${userId}-${system.clientId}-roles`}
                  titleText="System roles"
                  label="Select system roles"
                  items={roleItems}
                  itemToString={(item) => item?.text ?? ""}
                  selectedItems={roleItems.filter((role) => selectedRoles.includes(role.id))}
                  onChange={({ selectedItems }) => {
                    handleSystemRolesChange(
                      system,
                      (selectedItems ?? []).map((role) => role.id),
                    );
                  }}
                />

                {system.accessRoles.length > 0 && (
                  <div>
                    <small>App access roles</small>
                    <TagList values={system.accessRoles} type="cyan" />
                  </div>
                )}
              </Stack>
            );
          })}
        </Stack>
      </Tile>

      <Tile>
        <Stack gap={4}>
          <strong>Effective permissions</strong>
          {Object.entries(effectivePermissions).map(([category, permissions]) => (
            <div key={category}>
              <small>{category}</small>
              <TagList values={permissions.map((permission) => permission.key)} type="green" />
            </div>
          ))}
          {profile.effectiveAccess.permissions.length === 0 && (
            <small>No permissions resolved for this user.</small>
          )}
        </Stack>
      </Tile>

      <Button kind="primary" disabled={saving} onClick={handleSave}>
        {saving ? "Saving access..." : "Save access"}
      </Button>
    </Stack>
  );
}

function groupPermissions(permissions: RbacPermission[]) {
  return permissions.reduce<Record<string, RbacPermission[]>>((groups, permission) => {
    const category = permission.category || permission.key.split(":")[0] || "other";
    groups[category] = [...(groups[category] ?? []), permission];
    return groups;
  }, {});
}

function TagList({ values, type }: { values: string[]; type: "cyan" | "green" }) {
  return (
    <div style={{ display: "flex", flexWrap: "wrap", gap: 4, marginTop: 8 }}>
      {values.map((value) => (
        <Tag key={value} size="sm" type={type}>
          {value}
        </Tag>
      ))}
    </div>
  );
}

function getApiErrorMessage(error: unknown, fallback: string): string {
  if (!error || typeof error !== "object") {
    return fallback;
  }

  const maybeError = error as {
    data?: {
      message?: unknown;
      error?: {
        message?: unknown;
      };
    };
    error?: unknown;
  };

  if (typeof maybeError.data?.message === "string") {
    return maybeError.data.message;
  }

  if (typeof maybeError.data?.error?.message === "string") {
    return maybeError.data.error.message;
  }

  if (typeof maybeError.error === "string") {
    return maybeError.error;
  }

  return fallback;
}
