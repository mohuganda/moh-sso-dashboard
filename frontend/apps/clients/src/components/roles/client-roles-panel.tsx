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
  Modal,
} from "@carbon/react";
import { useState } from "react";

import { useListClientRolesQuery, useDeleteClientRoleMutation } from "@moh-sso/api";
import { useToast } from "@moh-sso/ui";

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
  const [roleToDelete, setRoleToDelete] = useState<string | null>(null);

  const { data: roles = [], isLoading } = useListClientRolesQuery(clientUuid);

  const [deleteRole, { isLoading: deletingRole }] = useDeleteClientRoleMutation();

  const handleDelete = async () => {
    if (!roleToDelete) return;

    try {
      await deleteRole({ clientUuid, roleName: roleToDelete }).unwrap();
      toast.success("Role deleted", `"${roleToDelete}" was removed`);
      setRoleToDelete(null);
    } catch {
      toast.error("Failed to delete role", "Please try again");
    }
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
                                  if (typeof roleName === "string") setRoleToDelete(roleName);
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

      <Modal
        open={roleToDelete !== null}
        modalHeading="Delete client role"
        primaryButtonText="Delete role"
        secondaryButtonText="Cancel"
        danger
        primaryButtonDisabled={deletingRole}
        onRequestSubmit={handleDelete}
        onRequestClose={() => setRoleToDelete(null)}
      >
        <p>
          Are you sure you want to delete the role <strong>{roleToDelete}</strong>?
        </p>
      </Modal>
    </Stack>
  );
}
