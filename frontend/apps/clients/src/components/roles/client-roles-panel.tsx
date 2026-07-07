import {
  Stack,
  Tile,
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  InlineLoading,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";

import { useListClientRolesQuery, useDeleteClientRoleMutation } from "../../api";
import { useModal, useToast } from "@moh-sso/ui";

import { CreateClientRoleForm } from "./create-client-role-form";

type Props = {
  clientUuid: string;
  clientId: string;
};

const headers = [
  { key: "name", header: "Role" },
  { key: "description", header: "Description" },
  { key: "actions", header: "" },
  { key: "roleName", header: "" },
];

export function ClientRolesPanel({ clientUuid, clientId }: Props) {
  const toast = useToast();
  const { openModal, closeModal } = useModal();

  const { data: roles = [], isLoading } = useListClientRolesQuery(clientUuid);

  const [deleteRole, { isLoading: deletingRole }] = useDeleteClientRoleMutation();

  const handleDelete = async (roleName: string) => {
    try {
      await deleteRole({ clientUuid, roleName }).unwrap();
      closeModal();
      toast.success("Role deleted", `"${roleName}" was removed`);
    } catch {
      toast.error("Failed to delete role", "Please try again");
    }
  };

  const handleDeleteRequest = (roleName: string) => {
    openModal({
      title: "Delete client role",
      onClose: closeModal,
      content: (
        <p>
          Are you sure you want to delete the role <strong>{roleName}</strong>?
        </p>
      ),
      primaryAction: {
        label: deletingRole ? "Deleting..." : "Delete role",
        kind: "danger",
        disabled: deletingRole,
        onClick: () => void handleDelete(roleName),
      },
      secondaryAction: {
        label: "Cancel",
        onClick: closeModal,
      },
    });
  };

  if (isLoading) {
    return <InlineLoading description="Loading roles…" />;
  }

  return (
    <Stack gap={6}>
      {/* -----------------------------
       * Create role
       * ----------------------------- */}
      <Tile>
        <Stack gap={4}>
          <strong>Add role</strong>
          <CreateClientRoleForm clientUuid={clientUuid} />
        </Stack>
      </Tile>

      {/* -----------------------------
       * Roles list
       * ----------------------------- */}
      <Tile>
        <DataTable
          rows={roles.map((r) => ({
            id: r.id,
            name: r.name.replace(`${clientId}:`, ""),
            roleName: r.name,
            description: r.description ?? "—",
          }))}
          headers={headers}
        >
          {({ rows, headers, getHeaderProps, getRowProps }) => (
            <Table>
              <TableHead>
                <TableRow>
                  {headers
                    .filter((header) => header.key !== "roleName")
                    .map((header) => (
                      <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                    ))}
                </TableRow>
              </TableHead>

              <TableBody>
                {rows.map((row) => (
                  <TableRow {...getRowProps({ row })}>
                    {row.cells.map((cell) => {
                      if (cell.info.header === "roleName") {
                        return null;
                      }

                      if (cell.info.header === "actions") {
                        return (
                          <TableCell key={cell.id}>
                            <OverflowMenu size="sm">
                              <OverflowMenuItem
                                itemText="Delete"
                                isDelete
                                onClick={() => {
                                  const roleName = row.cells.find(
                                    (cell) => cell.info.header === "roleName",
                                  )?.value;
                                  if (typeof roleName === "string") {
                                    handleDeleteRequest(roleName);
                                  }
                                }}
                              />
                            </OverflowMenu>
                          </TableCell>
                        );
                      }

                      return <TableCell key={cell.id}>{cell.value}</TableCell>;
                    })}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </DataTable>
      </Tile>
    </Stack>
  );
}
