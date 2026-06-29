import { Stack, Tile, MultiSelect, InlineLoading, Button } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";

import { useGetUserAccessProfileQuery, useUpdateUserAccessMutation } from "@moh-sso/rbac/api";
import { useToast } from "@moh-sso/ui";

type Props = {
  userId: string;
};

export function UserClientRolesPanel({ userId }: Props) {
  const toast = useToast();

  const { data: profile, isLoading: loadingProfile } = useGetUserAccessProfileQuery(userId);

  const [updateAccess, { isLoading: saving }] = useUpdateUserAccessMutation();

  const [selectedClientId, setSelectedClientId] = useState<string | null>(null);

  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);

  const selectedClient = useMemo(
    () => profile?.assignable.systems.find((system) => system.clientId === selectedClientId),
    [profile?.assignable.systems, selectedClientId],
  );

  useEffect(() => {
    if (!selectedClientId || !profile) return;

    setSelectedRoles(profile.effectiveAccess.clientRoles[selectedClientId] ?? []);
  }, [selectedClientId, profile]);

  const roleItems = useMemo(
    () =>
      (selectedClient?.roles ?? []).map((role) => ({
        id: role.name,
        text: role.displayName || role.name,
      })),
    [selectedClient?.roles],
  );

  /* ------------------------------------------------
   * Save (FULL REPLACEMENT)
   * ------------------------------------------------ */
  const handleSave = async () => {
    if (!selectedClientId || !profile) return;

    try {
      await updateAccess({
        userId,
        data: {
          realmRoles: profile.effectiveAccess.realmRoles,
          clientRoles: {
            ...profile.effectiveAccess.clientRoles,
            [selectedClientId]: selectedRoles,
          },
        },
      }).unwrap();

      toast.success("Roles updated", "User client roles were updated successfully");
    } catch (err) {
      toast.error("Update failed", "Unable to update client roles");
    }
  };

  /* ------------------------------------------------
   * Loading
   * ------------------------------------------------ */
  if (loadingProfile) {
    return <InlineLoading description="Loading client roles…" />;
  }

  /* ------------------------------------------------
   * Render
   * ------------------------------------------------ */
  return (
    <Stack gap={5}>
      <Tile>
        <Stack gap={4}>
          <strong>Select application</strong>

          <MultiSelect
            id="client-selector"
            label=""
            hideLabel
            titleText="Application"
            items={(profile?.assignable.systems ?? []).map((system) => ({
              id: system.clientId,
              text: system.displayName || system.clientId,
            }))}
            itemToString={(item) => item?.text ?? ""}
            selectedItems={
              selectedClient
                ? [
                    {
                      id: selectedClient.clientId,
                      text: selectedClient.displayName || selectedClient.clientId,
                    },
                  ]
                : []
            }
            onChange={({ selectedItems }) => {
              setSelectedClientId((selectedItems ?? [])[0]?.id ?? null);
            }}
          />
        </Stack>
      </Tile>

      {/* --------------------------------
       * Role assignment
       * -------------------------------- */}
      {selectedClientId && (
        <Tile>
          <Stack gap={4}>
            <strong>Roles for {selectedClient?.displayName || selectedClient?.clientId}</strong>

            <MultiSelect
              id="client-roles"
              label=""
              hideLabel
              titleText="Client roles"
              items={roleItems}
              itemToString={(item) => item?.text ?? ""}
              selectedItems={roleItems.filter((r) => selectedRoles.includes(r.id))}
              onChange={({ selectedItems }) => {
                setSelectedRoles((selectedItems ?? []).map((r) => r.id));
              }}
            />

            <Button
              kind="primary"
              size="sm"
              disabled={saving || !selectedClientId}
              onClick={handleSave}
            >
              {saving ? "Saving…" : "Save roles"}
            </Button>
          </Stack>
        </Tile>
      )}
    </Stack>
  );
}
