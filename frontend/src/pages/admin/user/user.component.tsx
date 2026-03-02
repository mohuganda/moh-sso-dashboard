import {
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableSelectRow,
  TableSelectAll,
  InlineLoading,
  Tile,
  Tag,
  Pagination,
} from "@carbon/react";
import { useEffect, useMemo, useState } from "react";

import { ErrorState } from "../../../components/errorstate/ErrorState";
import { useHeaderPanel } from "../../../components/header-panel/header-panel.context";
import { useToast } from "../../../components/notifications/toast/useToast";
import { UserFormPanel } from "../../../components/panels/create-user-panel";
import { UserClientRolesPanel } from "../../../components/panels/user-client-roles-panel";
import { useEnableUserModal } from "../../../components/user/useEnableUserModal";
import { useResetPasswordModal } from "../../../components/user/useResetPasswordModal";
import { UserFilters } from "../../../components/user/UserFilters";
import { useListUsersQuery, useToggleUserMutation } from "../../../store/api/users.api";
import type { User } from "../../../store/types/user.types";

import { UserActionsMenu } from "./user-actions-menu.component";
import { UserBulkActions } from "./user-bulk-actions.component";

/* -----------------------------
 * Table headers
 * ----------------------------- */
const headers = [
  { key: "username", header: "Username" },
  { key: "email", header: "Email" },
  { key: "status", header: "Status" },
  { key: "verified", header: "Email verified" },
  { key: "lastLogin", header: "Last login" },
  { key: "actions", header: "" },
  { key: "raw", header: "" },
];

type BulkAction = "enable" | "disable" | "roles" | null;

export default function UsersPage() {
  const { openPanel, closePanel } = useHeaderPanel();
  const { openEnableUserModal } = useEnableUserModal();
  const { openResetPasswordModal } = useResetPasswordModal();
  const toast = useToast();

  const [bulkAction, setBulkAction] = useState<BulkAction>(null);

  /* -----------------------------
   * Filters
   * ----------------------------- */
  const [statusFilter, setStatusFilter] = useState("all");
  const [roleFilter, setRoleFilter] = useState("all");
  const [neverLoggedIn, setNeverLoggedIn] = useState(false);

  /* -----------------------------
   * Pagination
   * ----------------------------- */
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  /* -----------------------------
   * Data
   * ----------------------------- */
  const { data: users = [], isLoading, isError, error, refetch } = useListUsersQuery();

  const [toggleUser] = useToggleUserMutation();

  useEffect(() => {
    setPage(1);
  }, [statusFilter, roleFilter, neverLoggedIn]);

  /* -----------------------------
   * Roles
   * ----------------------------- */
  const roles = useMemo(() => {
    const set = new Set<string>();
    users.forEach((u) => {
      u.realmRoles?.forEach((r) => set.add(r));
    });
    return ["all", ...Array.from(set)];
  }, [users]);

  /* -----------------------------
   * Filtering
   * ----------------------------- */
  const filteredUsers = useMemo(() => {
    return users.filter((u) => {
      if (statusFilter === "active" && !u.isActive) return false;
      if (statusFilter === "disabled" && u.isActive) return false;
      if (neverLoggedIn && u.lastLoginAt) return false;
      if (roleFilter !== "all" && !u.realmRoles?.includes(roleFilter)) return false;
      return true;
    });
  }, [users, statusFilter, roleFilter, neverLoggedIn]);

  /* -----------------------------
   * Pagination
   * ----------------------------- */
  const paginatedUsers = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredUsers.slice(start, start + pageSize);
  }, [filteredUsers, page, pageSize]);

  /* -----------------------------
   * Rows
   * ----------------------------- */
  const rows = paginatedUsers.map((u) => ({
    id: u.id,
    username: u.username,
    email: u.email ?? "—",
    status: u.isActive ? "Active" : "Disabled",
    verified: u.emailVerified ? "Verified" : "Not verified",
    lastLogin: u.lastLoginAt ? new Date(u.lastLoginAt).toLocaleString() : "Never",
    actions: "",
    raw: u,
  }));

  if (isLoading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading users…" />
      </div>
    );
  }

  if (isError) {
    return (
      <ErrorState
        title="Failed to load users"
        description={(error as any)?.data?.message ?? "Failed to load users"}
        primaryAction={{ label: "Retry", onClick: refetch }}
      />
    );
  }

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      <div>
        <h3 style={{ margin: 0 }}>Users</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Manage users, roles, and access to applications.
        </p>
      </div>

      <Tile>
        <UserFilters
          status={statusFilter}
          roles={roles}
          selectedRole={roleFilter}
          neverLoggedIn={neverLoggedIn}
          onStatusChange={setStatusFilter}
          onRoleChange={setRoleFilter}
          onToggleNeverLoggedIn={() => {
            setNeverLoggedIn((v) => !v);
          }}
        />
      </Tile>

      <Tile>
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getSelectionProps, selectedRows }) => {
            const selectedUsers = selectedRows.map(
              (r) => r.cells.find((c) => c.info.header === "raw")?.value as User,
            );

            return (
              <>
                {/* ================= BULK ACTIONS ================= */}
                <UserBulkActions
                  users={selectedUsers}
                  loadingAction={bulkAction}
                  onEnable={async () => {
                    try {
                      setBulkAction("enable");

                      await Promise.all(
                        selectedUsers.map((u) => toggleUser({ id: u.id, enabled: true }).unwrap()),
                      );

                      toast.success({
                        title: "Users enabled",
                        subtitle: `${selectedUsers.length} user(s) enabled.`,
                      });
                    } catch {
                      toast.error({
                        title: "Enable failed",
                        subtitle: "Some users could not be enabled.",
                      });
                    } finally {
                      setBulkAction(null);
                    }
                  }}
                  onDisable={async () => {
                    try {
                      setBulkAction("disable");

                      await Promise.all(
                        selectedUsers.map((u) => toggleUser({ id: u.id, enabled: false }).unwrap()),
                      );

                      toast.warning({
                        title: "Users disabled",
                        subtitle: `${selectedUsers.length} user(s) disabled.`,
                      });
                    } catch {
                      toast.error({
                        title: "Disable failed",
                        subtitle: "Some users could not be disabled.",
                      });
                    } finally {
                      setBulkAction(null);
                    }
                  }}
                  onAssignRoles={async () => {
                    setBulkAction("roles");

                    openPanel({
                      title: `Assign roles (${selectedUsers.length})`,
                      size: "lg",
                      content: (
                        <UserClientRolesPanel userId={selectedUsers.map((u) => u.id).join(",")} />
                      ),
                    });

                    setBulkAction(null);
                  }}
                />

                {/* ================= TABLE ================= */}
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableSelectAll {...getSelectionProps()} />
                      {headers
                        .filter((h) => h.key !== "raw")
                        .map((h) => (
                          <TableHeader {...getHeaderProps({ header: h })}>{h.header}</TableHeader>
                        ))}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {rows.map((row) => {
                      const user = row.cells.find((c) => c.info.header === "raw")?.value as User;

                      return (
                        <TableRow {...getRowProps({ row })}>
                          <TableSelectRow {...getSelectionProps({ row })} />

                          {row.cells.map((cell) => {
                            if (cell.info.header === "raw") return null;

                            if (cell.info.header === "status") {
                              return (
                                <TableCell key={cell.id}>
                                  <Tag type={!user.isAdmin ? "green" : "red"}>{cell.value}</Tag>
                                </TableCell>
                              );
                            }

                            if (cell.info.header === "actions") {
                              return (
                                <TableCell key={cell.id}>
                                  {row.isSelected && (
                                    <UserActionsMenu
                                      user={user}
                                      onEdit={() => {
                                        openPanel({
                                          title: "Edit user",
                                          content: (
                                            <UserFormPanel
                                              mode="edit"
                                              initialUser={user}
                                              onSuccess={closePanel}
                                            />
                                          ),
                                          size: "md",
                                        });
                                      }}
                                      onManageRoles={() => {
                                        openPanel({
                                          title: `Roles: ${user.username}`,
                                          size: "lg",
                                          content: <UserClientRolesPanel userId={user.id} />,
                                        });
                                      }}
                                      onToggleStatus={() => {
                                        openEnableUserModal({
                                          username: user.username,
                                          enabled: user.isActive ?? false,
                                          onConfirm: async () => {
                                            try {
                                              await toggleUser({
                                                id: user.id,
                                                enabled: !user.isActive,
                                              }).unwrap();

                                              toast.success({
                                                title: "User updated",
                                                subtitle: `${user.username} ${
                                                  user.isActive ? "disabled" : "enabled"
                                                }.`,
                                              });
                                            } catch {
                                              toast.error({
                                                title: "Update failed",
                                                subtitle: `Failed to update ${user.username}.`,
                                              });
                                            }
                                          },
                                        });
                                      }}
                                      onResetPassword={() => {
                                        openResetPasswordModal({
                                          username: user.username,
                                          email: user.email,
                                          onConfirm: () => {
                                            toast.info({
                                              title: "Password reset",
                                              subtitle: `Reset email sent to ${user.email}`,
                                            });
                                          },
                                        });
                                      }}
                                    />
                                  )}
                                </TableCell>
                              );
                            }

                            return <TableCell key={cell.id}>{cell.value}</TableCell>;
                          })}
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>

                <Pagination
                  page={page}
                  pageSize={pageSize}
                  pageSizes={[10, 20, 30, 50]}
                  totalItems={filteredUsers.length}
                  onChange={({ page, pageSize }) => {
                    setPage(page);
                    setPageSize(pageSize);
                  }}
                />
              </>
            );
          }}
        </DataTable>
      </Tile>
    </div>
  );
}
