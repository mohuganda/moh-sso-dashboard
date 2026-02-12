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

import {
  useListClientRolesQuery,
  useDeleteClientRoleMutation,
} from "../../../store/api/clientRoles.api";
import { useToast } from "../../notifications/toast/useToast";

import { CreateClientRoleForm } from "./create-client-role-form";

type Props = {
  id: string;
  clientId: string;
};

const headers = [
  { key: "name", header: "Role" },
  { key: "description", header: "Description" },
  { key: "actions", header: "" },
];

export function ClientRolesPanel({ id, clientId }: Props) {
  const toast = useToast();

  const { data: roles = [], isLoading } = useListClientRolesQuery(id);

  const [deleteRole] = useDeleteClientRoleMutation();

  const handleDelete = async (roleId: string, roleName: string) => {
    if (!confirm(`Delete role "${roleName}"?`)) return;

    try {
      await deleteRole({ clientId, role: roleId }).unwrap();
      toast.success("Role deleted", `"${roleName}" was removed`);
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
          <CreateClientRoleForm clientId={clientId} />
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
            description: r.description ?? "—",
          }))}
          headers={headers}
        >
          {({ rows, headers, getHeaderProps, getRowProps }) => (
            <Table>
              <TableHead>
                <TableRow>
                  {headers.map((header) => (
                    <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                  ))}
                </TableRow>
              </TableHead>

              <TableBody>
                {rows.map((row) => (
                  <TableRow {...getRowProps({ row })}>
                    {row.cells.map((cell) => {
                      if (cell.info.header === "actions") {
                        return (
                          <TableCell key={cell.id}>
                            <OverflowMenu size="sm">
                              <OverflowMenuItem
                                itemText="Delete"
                                isDelete
                                onClick={() => handleDelete(row.id, row.cells[0].value)}
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
