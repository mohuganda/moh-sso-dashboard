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

import { ClientFilters } from "../components/ClientFilters";
import { useEnableClientModal } from "../components/useEnableClientModal";
import { ErrorState , useHeaderPanel , useToast } from "@moh-sso/ui";
import { ClientFormPanel } from "../components/client-form-panel";
import { useListClientsQuery, useToggleClientMutation } from "@moh-sso/api";
import type { Client } from "@moh-sso/types";

import { ClientActionsMenu } from "./client-actions-menu.component";
import { ClientBulkActions } from "./client-bulk-actions.component";

/* -----------------------------
 * Filters
 * ----------------------------- */
type StatusFilter = "all" | "enabled" | "disabled";
type TypeFilter = "all" | "public" | "confidential";

const STATUS_OPTIONS = [
  { id: "all", label: "All" },
  { id: "enabled", label: "Enabled" },
  { id: "disabled", label: "Disabled" },
] as const;

const TYPE_OPTIONS = [
  { id: "all", label: "All" },
  { id: "public", label: "Public" },
  { id: "confidential", label: "Confidential" },
] as const;

/* -----------------------------
 * Table headers
 * ----------------------------- */
const headers = [
  { key: "name", header: "Name" },
  { key: "clientId", header: "Client ID" },
  { key: "type", header: "Type" },
  { key: "status", header: "Status" },
  { key: "actions", header: "" },
  { key: "raw", header: "" }, // hidden
];

export default function ClientsPage() {
  const { openPanel, closePanel } = useHeaderPanel();
  const { openEnableClientModal } = useEnableClientModal();
  const toast = useToast();

  const [bulkAction, setBulkAction] = useState<"enable" | "disable" | null>(null);

  /* -----------------------------
   * Filters
   * ----------------------------- */
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all");
  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all");

  /* -----------------------------
   * Pagination
   * ----------------------------- */
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  /* -----------------------------
   * Data
   * ----------------------------- */
  const { data: clients = [], isLoading, isError, error, refetch } = useListClientsQuery();

  const [toggleClient] = useToggleClientMutation();

  /* -----------------------------
   * Reset page on filter change
   * ----------------------------- */
  useEffect(() => {
    setPage(1);
  }, [statusFilter, typeFilter]);

  /* -----------------------------
   * Filtering
   * ----------------------------- */
  const filteredClients = useMemo(() => {
    return clients.filter((c) => {
      if (statusFilter === "enabled" && !c.enabled) return false;
      if (statusFilter === "disabled" && c.enabled) return false;
      if (typeFilter === "public" && !c.publicClient) return false;
      if (typeFilter === "confidential" && c.publicClient) return false;
      return true;
    });
  }, [clients, statusFilter, typeFilter]);

  /* -----------------------------
   * Pagination slice
   * ----------------------------- */
  const paginatedClients = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredClients.slice(start, start + pageSize);
  }, [filteredClients, page, pageSize]);

  /* -----------------------------
   * Rows
   * ----------------------------- */
  const rows = paginatedClients.map((c) => ({
    id: c.id,
    name: c.name,
    clientId: c.clientId,
    type: c.publicClient ? "Public" : "Confidential",
    status: c.enabled ? "Enabled" : "Disabled",
    actions: "",
    raw: c,
  }));

  /* -----------------------------
   * Loading / Error
   * ----------------------------- */
  if (isLoading) {
    return (
      <div style={{ padding: "2rem" }}>
        <InlineLoading description="Loading clients…" />
      </div>
    );
  }

  if (isError) {
    return (
      <ErrorState
        title="Failed to load clients"
        description={(error as any)?.data?.message ?? "Failed to load clients"}
        primaryAction={{ label: "Retry", onClick: refetch }}
      />
    );
  }

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* Header */}
      <div>
        <h3 style={{ margin: 0 }}>Clients</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Registered applications and services integrated with the platform.
        </p>
      </div>

      {/* Filters */}
      <Tile>
        <ClientFilters
          status={statusFilter}
          type={typeFilter}
          statusOptions={STATUS_OPTIONS}
          typeOptions={TYPE_OPTIONS}
          onStatusChange={setStatusFilter}
          onTypeChange={setTypeFilter}
        />
      </Tile>

      {/* Table */}
      <Tile>
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps, getSelectionProps, selectedRows }) => {
            const selectedClients = selectedRows.map(
              (r) => r.cells.find((c) => c.info.header === "raw")?.value as Client,
            );

            return (
              <>
                <ClientBulkActions
                  clients={selectedClients}
                  loadingAction={bulkAction}
                  onEnable={async () => {
                    try {
                      setBulkAction("enable");

                      await Promise.all(
                        selectedClients.map((c) =>
                          toggleClient({
                            id: c.clientId,
                            enabled: true,
                          }).unwrap(),
                        ),
                      );

                      toast.success({
                        title: "Clients enabled",
                        subtitle: `${selectedClients.length} client(s) enabled.`,
                      });
                    } catch {
                      toast.error({
                        title: "Enable failed",
                        subtitle: "Some clients could not be enabled.",
                      });
                    } finally {
                      setBulkAction(null);
                    }
                  }}
                  onDisable={async () => {
                    try {
                      setBulkAction("disable");

                      await Promise.all(
                        selectedClients.map((c) =>
                          toggleClient({
                            id: c.clientId,
                            enabled: false,
                          }).unwrap(),
                        ),
                      );

                      toast.warning({
                        title: "Clients disabled",
                        subtitle: `${selectedClients.length} client(s) disabled.`,
                      });
                    } catch {
                      toast.error({
                        title: "Disable failed",
                        subtitle: "Some clients could not be disabled.",
                      });
                    } finally {
                      setBulkAction(null);
                    }
                  }}
                />

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
                      const client = row.cells.find((c) => c.info.header === "raw")
                        ?.value as Client;

                      return (
                        <TableRow {...getRowProps({ row })}>
                          <TableSelectRow {...getSelectionProps({ row })} />

                          {row.cells.map((cell) => {
                            if (cell.info.header === "raw") return null;

                            if (cell.info.header === "status") {
                              return (
                                <TableCell key={cell.id}>
                                  <Tag type={client.enabled ? "green" : "red"}>{cell.value}</Tag>
                                </TableCell>
                              );
                            }

                            if (cell.info.header === "actions") {
                              return (
                                <TableCell key={cell.id}>
                                  {row.isSelected && (
                                    <ClientActionsMenu
                                      client={client}
                                      onEdit={() => {
                                        openPanel({
                                          title: "Edit client",
                                          content: (
                                            <ClientFormPanel
                                              mode="edit"
                                              initialClient={client}
                                              onSuccess={closePanel}
                                            />
                                          ),
                                          size: "md",
                                        });
                                      }}
                                      onToggleStatus={() => {
                                        openEnableClientModal({
                                          clientName: client.name,
                                          enabled: client.enabled,
                                          onConfirm: async () => {
                                            try {
                                              await toggleClient({
                                                id: client.clientId,
                                                enabled: !client.enabled,
                                              }).unwrap();

                                              toast.success({
                                                title: "Client updated",
                                                subtitle: `${client.name} updated successfully.`,
                                              });
                                            } catch {
                                              toast.error({
                                                title: "Update failed",
                                                subtitle: `Failed to update ${client.name}.`,
                                              });
                                            }
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
              </>
            );
          }}
        </DataTable>

        {/* ================= PAGINATION ================= */}
        <Pagination
          page={page}
          pageSize={pageSize}
          pageSizes={[10, 20, 30, 50]}
          totalItems={filteredClients.length}
          onChange={({ page, pageSize }) => {
            setPage(page);
            setPageSize(pageSize);
          }}
        />
      </Tile>
    </div>
  );
}
