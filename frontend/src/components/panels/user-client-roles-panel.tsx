import { Stack, Tile, MultiSelect, InlineLoading, Button } from "@carbon/react";
import { useEffect, useMemo, useState } from "react";

import { useListClientRolesQuery } from "../../store/api/clientRoles.api";
import { useListClientsQuery } from "../../store/api/clients.api";
import {
  useGetUserClientRolesQuery,
  useUpdateUserClientRolesMutation,
} from "../../store/api/users.api";
import { useToast } from "../notifications/toast/useToast";

type Props = {
  userId: string;
};

export function UserClientRolesPanel({ userId }: Props) {
  const toast = useToast();

  const { data: clients = [], isLoading: loadingClients } = useListClientsQuery();

  const { data: assignments = [], isLoading: loadingAssignments } =
    useGetUserClientRolesQuery(userId);

  const [updateRoles, { isLoading: saving }] = useUpdateUserClientRolesMutation();

  const [selectedClientId, setSelectedClientId] = useState<string | null>(null);

  const [selectedClientUuid, setSelectedClientUuid] = useState<string | null>(null);

  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);

  const selectedClient = useMemo(
    () => clients.find((c) => c.clientId === selectedClientId),
    [clients, selectedClientId],
  );
  const { data: clientRoles = [], isLoading: loadingRoles } = useListClientRolesQuery(
    selectedClientId!,
    {
      skip: !selectedClientId,
    },
  );

  useEffect(() => {
    if (!selectedClientId) return;

    const assignment = assignments.find((a) => a.clientId === selectedClientId);

    const client = clients.find((c) => c.clientId === selectedClientId);
    setSelectedClientUuid(assignment?.id ?? client?.id ?? null);
    setSelectedRoles(assignment?.roles ?? []);
  }, [selectedClientId, assignments, clients]);

  const roleItems = useMemo(
    () =>
      clientRoles.map((r) => ({
        id: r.name,
        text: r.name,
      })),
    [clientRoles],
  );

  /* ------------------------------------------------
   * Save (FULL REPLACEMENT)
   * ------------------------------------------------ */
  const handleSave = async () => {
    if (!selectedClientId || !selectedClientUuid) return;

    try {
      await updateRoles({
        userId,
        clientId: selectedClientId,
        clientUuid: selectedClientUuid,
        roles: selectedRoles,
      }).unwrap();

      toast.success("Roles updated", "User client roles were updated successfully");
    } catch (err) {
      toast.error("Update failed", "Unable to update client roles");
    }
  };

  /* ------------------------------------------------
   * Loading
   * ------------------------------------------------ */
  if (loadingClients || loadingAssignments) {
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
            items={clients.map((c) => ({
              id: c.clientId,
              text: c.name,
            }))}
            itemToString={(item) => item?.text ?? ""}
            selectedItems={
              selectedClient ? [{ id: selectedClient.clientId, text: selectedClient.name }] : []
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
            <strong>Roles for {selectedClient?.name}</strong>

            {loadingRoles ? (
              <InlineLoading description="Loading roles…" />
            ) : (
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
            )}

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
