import {
  DataTable,
  InlineLoading,
  Tag,
  Tile,
  Button,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableSelectAll,
  TableSelectRow,
} from "@carbon/react";
import { useEffect, useMemo, useState } from "react";

import { AuditLogPanel } from "../components/AuditLogDrawer";
import { AuditLogFilters } from "../components/AuditLogFilters";
import { AuditMetricsPanel } from "../components/AuditMetricsPanel";

import { EmptyState, ErrorState, useHeaderPanel, useToast } from "@moh-sso/ui";

import { useListAuditLogsQuery } from "@moh-sso/api";
import type { AuditFilters, AuditLog, Cursor } from "@moh-sso/types";
import { AuditLogBulkActions } from "../components/audit-log-bulk-actions.component";
import { AuditLogActionsMenu } from "../components/audit-log-actions-menu.component";

type SuccessFilter = "true" | "false";

const headers = [
  { key: "shortId", header: "ID" },
  { key: "time", header: "Time" },
  { key: "actor", header: "Actor" },
  { key: "action", header: "Action" },
  { key: "client", header: "Client" },
  { key: "ip", header: "IP Address" },
  { key: "result", header: "Result" },
  { key: "actions", header: "" },
  { key: "raw", header: "" }, // hidden
];

function toRFC3339(d: Date) {
  return d.toISOString();
}

function getAuditLogTime(log: AuditLog) {
  if (!log.created_at?.Valid) return "—";

  const date = new Date(log.created_at.Time);
  if (Number.isNaN(date.getTime())) return "—";

  return date.toLocaleString();
}

function getAuditSuccess(log: AuditLog): boolean | null {
  const value = log.metadata?.RawMessage?.success;

  if (typeof value === "boolean") return value;

  return null;
}

function getAuditResult(log: AuditLog) {
  const success = getAuditSuccess(log);

  if (success === true) return "success";
  if (success === false) return "failure";

  return "—";
}

function getAuditClient(log: AuditLog) {
  return log.metadata?.RawMessage?.client_id ?? "—";
}

function getAuditIp(log: AuditLog) {
  return log.metadata?.RawMessage?.ip ?? "—";
}

function downloadFile(filename: string, content: string, type: string) {
  const blob = new Blob([content], { type });
  const url = URL.createObjectURL(blob);

  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();

  URL.revokeObjectURL(url);
}

function escapeCsvValue(value: unknown) {
  const text = String(value ?? "");
  return `"${text.replaceAll('"', '""')}"`;
}

function exportAuditLogsCsv(logs: AuditLog[]) {
  const header = ["id", "created_at", "actor", "user_id", "action", "client_id", "ip", "success"];

  const rows = logs.map((log) => [
    log.id,
    getAuditLogTime(log),
    log.username ?? "System",
    log.user_id ?? "",
    log.action,
    getAuditClient(log),
    getAuditIp(log),
    getAuditResult(log),
  ]);

  const csv = [header, ...rows].map((row) => row.map(escapeCsvValue).join(",")).join("\n");

  downloadFile("audit-logs.csv", csv, "text/csv;charset=utf-8");
}

function exportAuditLogsJson(logs: AuditLog[]) {
  downloadFile("audit-logs.json", JSON.stringify(logs, null, 2), "application/json;charset=utf-8");
}

export default function AuditLogs() {
  const now = new Date();
  const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const toast = useToast();
  const { openPanel, closePanel } = useHeaderPanel();

  /* -----------------------------
   * Filters
   * ----------------------------- */
  const [from, setFrom] = useState(toRFC3339(start));
  const [to, setTo] = useState(toRFC3339(now));
  const [action, setAction] = useState<string>();
  const [clientId, setClientId] = useState<string>();
  const [success, setSuccess] = useState<SuccessFilter>();

  /* -----------------------------
   * Cursor pagination
   * ----------------------------- */
  const [cursor, setCursor] = useState<Cursor | null>(null);
  const [hasMore, setHasMore] = useState(true);
  const [items, setItems] = useState<AuditLog[]>([]);

  /* -----------------------------
   * Query
   * ----------------------------- */
  const queryArgs = useMemo<AuditFilters>(() => {
    const q: AuditFilters = {
      from,
      to,
      action,
      client_id: clientId,
      success,
      limit: 50,
    };

    if (cursor) {
      q.cursor = {
        cursor_id: cursor.cursor_id,
        cursor_created_at: cursor.cursor_created_at,
      };
    }

    return q;
  }, [from, to, action, clientId, success, cursor]);

  const { data, isLoading, isFetching, isError, error, refetch } = useListAuditLogsQuery(
    queryArgs,
    {
      skip: !from || !to,
    },
  );

  /* -----------------------------
   * Append cursor data
   * ----------------------------- */
  useEffect(() => {
    if (!data) return;

    setItems((prev) => {
      if (!cursor) return data.items;

      const existing = new Set(prev.map((item) => item.id));
      const next = data.items.filter((item) => !existing.has(item.id));

      return [...prev, ...next];
    });

    setHasMore(data.has_more);

    if (data.next_cursor) {
      setCursor(data.next_cursor);
    }
  }, [data, cursor]);

  /* -----------------------------
   * Reset list on filter change
   * ----------------------------- */
  useEffect(() => {
    setCursor(null);
    setHasMore(true);
    setItems([]);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [from, to, action, clientId, success]);

  /* -----------------------------
   * Rows
   * ----------------------------- */
  const rows = useMemo(
    () =>
      items.map((log) => ({
        id: log.id,
        shortId: log.id.slice(0, 8),
        time: getAuditLogTime(log),
        actor: log.username ?? "System",
        action: log.action,
        client: getAuditClient(log),
        ip: getAuditIp(log),
        result: getAuditResult(log),
        actions: "",
        raw: log,
      })),
    [items],
  );

  function clearFilters() {
    setAction(undefined);
    setClientId(undefined);
    setSuccess(undefined);
  }

  function handleView(log: AuditLog) {
    openPanel({
      title: "Audit Log",
      content: <AuditLogPanel log={log} onClose={closePanel} />,
      size: "md",
    });
  }

  function handleExportCsv(logs: AuditLog[]) {
    exportAuditLogsCsv(logs);

    toast.success({
      title: "Audit logs exported",
      subtitle: `${logs.length} audit log(s) exported as CSV.`,
    });
  }

  function handleExportJson(logs: AuditLog[]) {
    exportAuditLogsJson(logs);

    toast.success({
      title: "Audit logs exported",
      subtitle: `${logs.length} audit log(s) exported as JSON.`,
    });
  }

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      {/* Header */}
      <div>
        <h3 style={{ margin: 0 }}>Audit Logs</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Security and administrative activity across the platform.
        </p>
      </div>

      {/* Metrics */}
      <AuditMetricsPanel from={from} to={to} />

      {/* Filters */}
      <Tile>
        <AuditLogFilters
          action={action}
          clientId={clientId}
          success={success}
          onFromChange={setFrom}
          onToChange={setTo}
          onActionChange={setAction}
          onClientChange={setClientId}
          onSuccessChange={setSuccess}
          onClear={clearFilters}
          onExportCsv={() => handleExportCsv(items)}
          onExportJson={() => handleExportJson(items)}
        />
      </Tile>

      {/* Table */}
      <Tile>
        {isLoading && <InlineLoading description="Loading audit logs…" />}

        {isError && (
          <ErrorState
            title="Failed to load audit logs"
            description={(error as any)?.data?.message ?? "Failed to load audit logs"}
            primaryAction={{
              label: "Retry",
              onClick: refetch,
            }}
          />
        )}

        {!isLoading && !isError && rows.length === 0 && (
          <EmptyState
            title="No audit logs found"
            description="No audit events match the selected filters."
            secondaryAction={{
              label: "Clear filters",
              onClick: clearFilters,
            }}
          />
        )}

        {!isLoading && !isError && rows.length > 0 && (
          <>
            <DataTable rows={rows} headers={headers}>
              {({
                rows,
                headers,
                getHeaderProps,
                getRowProps,
                getSelectionProps,
                selectedRows,
              }) => {
                const selectedLogs = selectedRows.map(
                  (row) => row.cells.find((cell) => cell.info.header === "raw")?.value as AuditLog,
                );

                return (
                  <>
                    <AuditLogBulkActions
                      logs={selectedLogs}
                      onExportCsv={() => handleExportCsv(selectedLogs)}
                      onExportJson={() => handleExportJson(selectedLogs)}
                    />

                    <Table size="lg">
                      <TableHead>
                        <TableRow>
                          <TableSelectAll {...getSelectionProps()} />

                          {headers
                            .filter((header) => header.key !== "raw")
                            .map((header) => (
                              <TableHeader {...getHeaderProps({ header })}>
                                {header.header}
                              </TableHeader>
                            ))}
                        </TableRow>
                      </TableHead>

                      <TableBody>
                        {rows.map((row) => {
                          const raw = row.cells.find((cell) => cell.info.header === "raw")
                            ?.value as AuditLog;

                          return (
                            <TableRow {...getRowProps({ row })}>
                              <TableSelectRow {...getSelectionProps({ row })} />

                              {row.cells.map((cell) => {
                                if (cell.info.header === "raw") return null;

                                if (cell.info.header === "result") {
                                  return (
                                    <TableCell key={cell.id}>
                                      <Tag
                                        type={
                                          cell.value === "success"
                                            ? "green"
                                            : cell.value === "failure"
                                              ? "red"
                                              : "gray"
                                        }
                                      >
                                        {cell.value}
                                      </Tag>
                                    </TableCell>
                                  );
                                }

                                if (cell.info.header === "actions") {
                                  return (
                                    <TableCell key={cell.id}>
                                      {row.isSelected && (
                                        <AuditLogActionsMenu
                                          log={raw}
                                          onView={() => handleView(raw)}
                                        />
                                      )}
                                    </TableCell>
                                  );
                                }

                                return <TableCell key={cell.id}>{cell.value || "—"}</TableCell>;
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

            {hasMore && (
              <div style={{ textAlign: "center", padding: 16 }}>
                <Button kind="secondary" disabled={isFetching} onClick={() => refetch()}>
                  {isFetching ? "Loading…" : "Load more"}
                </Button>
              </div>
            )}
          </>
        )}
      </Tile>
    </div>
  );
}
