import {
  DataTable,
  InlineLoading,
  Tag,
  Tile,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";
import { useEffect, useMemo, useState } from "react";

import { AuditLogPanel } from "../components/AuditLogDrawer";
import { AuditLogFilters } from "../components/AuditLogFilters";
import { AuditMetricsPanel } from "../components/AuditMetricsPanel";
import { EmptyState , ErrorState , useHeaderPanel } from "@moh-sso/ui";
import { useListAuditLogsQuery } from "@moh-sso/api";
import type { AuditFilters, AuditLog, Cursor } from "@moh-sso/types";

type SuccessFilter = "true" | "false";

function toRFC3339(d: Date) {
  return d.toISOString();
}

export default function AuditLogs() {
  const now = new Date();
  const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

  const [from, setFrom] = useState(toRFC3339(start));
  const [to, setTo] = useState(toRFC3339(now));
  const [action, setAction] = useState<string>();
  const [clientId, setClientId] = useState<string>();
  const [success, setSuccess] = useState<SuccessFilter>();

  const [cursor, setCursor] = useState<Cursor | null>(null);
  const [hasMore, setHasMore] = useState(true);
  const [items, setItems] = useState<AuditLog[]>([]);

  const { openPanel, closePanel } = useHeaderPanel();

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
    { skip: !from || !to },
  );

  useEffect(() => {
    if (!data) return;

    setItems((prev) => (cursor ? [...prev, ...data.items] : data.items));
    setHasMore(data.has_more);

    if (data.next_cursor) {
      setCursor(data.next_cursor);
    }
  }, [data, setItems, setCursor, cursor]);

  useEffect(() => {
    setCursor(null);
    setHasMore(true);
    setItems([]);
    refetch();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [from, to, action, clientId, success]);

  const rows = useMemo(
    () =>
      items.map((log) => ({
        id: log.id,
        shortId: log.id.slice(0, 8),
        time: log.created_at.Valid ? new Date(log.created_at.Time).toLocaleString() : "—",
        actor: log.username ?? "System",
        action: log.action,
        client: log.metadata?.RawMessage?.client_id ?? "—",
        result:
          log.metadata?.RawMessage?.success === true
            ? "success"
            : log.metadata?.RawMessage?.success === false
              ? "failure"
              : "—",
        raw: log,
      })),
    [items],
  );

  const headers = [
    { key: "shortId", header: "ID" },
    { key: "time", header: "Time" },
    { key: "actor", header: "Actor" },
    { key: "action", header: "Action" },
    { key: "client", header: "Client" },
    { key: "result", header: "Result" },
    { key: "raw", header: "" },
  ];

  function clearFilters() {
    setAction(undefined);
    setClientId(undefined);
    setSuccess(undefined);
  }

  return (
    <div style={{ padding: 16, display: "grid", gap: 16 }}>
      <div>
        <h3 style={{ margin: 0 }}>Audit Logs</h3>
        <p style={{ marginTop: 6, opacity: 0.8 }}>
          Security and administrative activity across the platform.
        </p>
      </div>

      <AuditMetricsPanel from={from} to={to} />

      <Tile>
        <AuditLogFilters
          onFromChange={setFrom}
          onToChange={setTo}
          onClientChange={setClientId}
          onSuccessChange={setSuccess}
          onClear={clearFilters}
          onExportCsv={() => {}}
          onExportJson={() => {}}
        />
      </Tile>

      <Tile>
        {isLoading && <InlineLoading description="Loading audit logs…" />}

        {isError && (
          <ErrorState
            title="Failed to load audit logs"
            description={(error as any)?.data?.message}
            primaryAction={{ label: "Retry", onClick: refetch }}
          />
        )}

        {!isLoading && rows.length === 0 && (
          <EmptyState
            title="No audit logs found"
            description="No audit events match the selected filters."
            secondaryAction={{ label: "Clear filters", onClick: clearFilters }}
          />
        )}

        {rows.length > 0 && (
          <>
            <DataTable rows={rows} headers={headers}>
              {({ rows, headers, getHeaderProps, getRowProps }) => (
                <TableContainer>
                  <Table size="lg">
                    <TableHead>
                      <TableRow>
                        {headers.map(
                          (header) =>
                            header.key !== "raw" && (
                              <TableHeader {...getHeaderProps({ header })}>
                                {header.header}
                              </TableHeader>
                            ),
                        )}
                        <TableHeader />
                      </TableRow>
                    </TableHead>

                    <TableBody>
                      {rows.map((row) => {
                        const raw = row.cells.find((c) => c.info.header === "raw")
                          ?.value as AuditLog;

                        return (
                          <TableRow {...getRowProps({ row })}>
                            {row.cells.map((cell) => {
                              if (cell.info.header === "raw") return null;

                              return (
                                <TableCell key={cell.id}>
                                  {cell.info.header === "result" ? (
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
                                  ) : (
                                    cell.value
                                  )}
                                </TableCell>
                              );
                            })}

                            <TableCell style={{ textAlign: "right" }}>
                              <OverflowMenu size="sm" flipped>
                                <OverflowMenuItem
                                  itemText="View"
                                  onClick={() => {
                                    openPanel({
                                      title: "Audit Log",
                                      content: <AuditLogPanel log={raw} onClose={closePanel} />,
                                    });
                                  }}
                                />
                              </OverflowMenu>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </TableContainer>
              )}
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
