import {
  Button,
  CodeSnippet,
  DataTable,
  InlineNotification,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  Tag,
  TextInput,
} from "@carbon/react";
import { View } from "@carbon/react/icons";
import { useMemo, useState } from "react";

import { useListRbacAuditEventsQuery } from "../api";
import type { RbacAuditEvent } from "../types";
import { DataTableShell } from "@moh-sso/ui";

const headers = [
  { key: "createdAt", header: "Date" },
  { key: "action", header: "Action" },
  { key: "resource", header: "Resource" },
  { key: "systemClientId", header: "System" },
  { key: "roleName", header: "Role" },
  { key: "permissionKey", header: "Permission" },
  { key: "actorUserId", header: "Actor" },
  { key: "actions", header: "Actions" },
  { key: "raw", header: "Raw" },
];

export function AuditTrailPanel() {
  const [actor, setActor] = useState("");
  const [systemClientId, setSystemClientId] = useState("");
  const [roleName, setRoleName] = useState("");
  const [permissionKey, setPermissionKey] = useState("");
  const [action, setAction] = useState("");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [selectedEvent, setSelectedEvent] = useState<RbacAuditEvent | null>(null);

  const { data, isFetching, isError, refetch } = useListRbacAuditEventsQuery({
    actor,
    systemClientId,
    roleName,
    permissionKey,
    action,
    from,
    to,
    limit: 100,
  });

  const auditEvents = useMemo(() => data ?? [], [data]);
  const rows = useMemo(
    () =>
      auditEvents.map((event) => ({
        ...event,
        resource: [event.resourceType, event.resourceId].filter(Boolean).join(": "),
        actions: "View",
        raw: event,
      })),
    [auditEvents],
  );

  const clearFilters = () => {
    setActor("");
    setSystemClientId("");
    setRoleName("");
    setPermissionKey("");
    setAction("");
    setFrom("");
    setTo("");
  };

  return (
    <section>
      <DataTableShell
        title="Audit Trail"
        description="Review RBAC governance changes, sync activity, role mappings, and permission changes."
        rows={rows}
        headers={headers}
        getRowId={(row) => row.id}
        isLoading={isFetching}
        loadingDescription="Loading RBAC audit events..."
        emptyTitle="No RBAC audit events found"
        emptyDescription="Try changing the filters or perform an RBAC governance action."
        filters={
          <div className="rbac-audit-filters">
            <TextInput id="rbac-audit-actor" labelText="Actor user ID" value={actor} onChange={(event) => setActor(event.target.value)} />
            <TextInput id="rbac-audit-system" labelText="System client ID" value={systemClientId} onChange={(event) => setSystemClientId(event.target.value)} />
            <TextInput id="rbac-audit-role" labelText="Role" value={roleName} onChange={(event) => setRoleName(event.target.value)} />
            <TextInput id="rbac-audit-permission" labelText="Permission" value={permissionKey} onChange={(event) => setPermissionKey(event.target.value)} />
            <TextInput id="rbac-audit-action" labelText="Action" value={action} onChange={(event) => setAction(event.target.value)} />
            <TextInput id="rbac-audit-from" labelText="From" placeholder="2026-06-01T00:00:00Z" value={from} onChange={(event) => setFrom(event.target.value)} />
            <TextInput id="rbac-audit-to" labelText="To" placeholder="2026-06-30T23:59:59Z" value={to} onChange={(event) => setTo(event.target.value)} />
            <div className="rbac-audit-actions">
              <Button kind="secondary" size="sm" onClick={clearFilters}>
                Clear
              </Button>
              <Button kind="ghost" size="sm" onClick={() => refetch()}>
                Refresh
              </Button>
            </div>
          </div>
        }
        topContent={
          <>
            {isError && (
              <InlineNotification
                kind="error"
                lowContrast
                title="Unable to load RBAC audit trail"
                subtitle="Check your connection and try again."
              />
            )}
            {selectedEvent && (
              <AuditEventDetail event={selectedEvent} onClose={() => setSelectedEvent(null)} />
            )}
          </>
        }
      >
        {({ rows, headers }) => (
          <DataTable rows={rows} headers={headers}>
            {({ rows, headers, getHeaderProps, getRowProps }) => (
              <Table size="lg">
                <TableHead>
                  <TableRow>
                    {headers
                      .filter((header) => header.key !== "raw")
                      .map((header) => {
                        const { key, ...headerProps } = getHeaderProps({ header });
                        return (
                          <TableHeader key={key} {...headerProps}>
                            {header.header}
                          </TableHeader>
                        );
                      })}
                  </TableRow>
                </TableHead>
                <TableBody>
                  {rows.map((row) => {
                    const raw = row.cells.find((cell) => cell.info.header === "raw")?.value as RbacAuditEvent;
                    const { key, ...rowProps } = getRowProps({ row });

                    return (
                      <TableRow key={key} {...rowProps}>
                        {row.cells.map((cell) => {
                          if (cell.info.header === "raw") return null;
                          if (cell.info.header === "action") {
                            return (
                              <TableCell key={cell.id}>
                                <Tag type={tagForAction(String(cell.value))}>{cell.value || "unknown"}</Tag>
                              </TableCell>
                            );
                          }
                          if (cell.info.header === "actions") {
                            return (
                              <TableCell key={cell.id}>
                                <Button
                                  hasIconOnly
                                  iconDescription="View RBAC audit event"
                                  kind="ghost"
                                  renderIcon={View}
                                  size="sm"
                                  onClick={() => setSelectedEvent(raw)}
                                />
                              </TableCell>
                            );
                          }
                          return <TableCell key={cell.id}>{formatCellValue(cell.value)}</TableCell>;
                        })}
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            )}
          </DataTable>
        )}
      </DataTableShell>
    </section>
  );
}

function AuditEventDetail({ event, onClose }: { event: RbacAuditEvent; onClose: () => void }) {
  return (
    <div className="rbac-audit-detail">
      <div className="rbac-audit-detail__header">
        <div>
          <h4>{event.action}</h4>
          <small>{event.createdAt}</small>
        </div>
        <Button kind="ghost" size="sm" onClick={onClose}>
          Close
        </Button>
      </div>
      <dl className="rbac-audit-detail__meta">
        <div>
          <dt>Resource</dt>
          <dd>{[event.resourceType, event.resourceId].filter(Boolean).join(": ") || "—"}</dd>
        </div>
        <div>
          <dt>System</dt>
          <dd>{event.systemClientId || "—"}</dd>
        </div>
        <div>
          <dt>Role</dt>
          <dd>{event.roleName || "—"}</dd>
        </div>
        <div>
          <dt>Permission</dt>
          <dd>{event.permissionKey || "—"}</dd>
        </div>
        <div>
          <dt>Actor</dt>
          <dd>{event.actorUserId || "system"}</dd>
        </div>
      </dl>
      <CodeSnippet type="multi" feedback="Copied audit details">
        {JSON.stringify(event.details ?? {}, null, 2)}
      </CodeSnippet>
    </div>
  );
}

function formatCellValue(value: unknown) {
  if (typeof value !== "string") return "—";
  return value.trim() || "—";
}

function tagForAction(action: string) {
  if (action.includes("removed") || action.includes("deleted")) return "red";
  if (action.includes("assigned") || action.includes("created") || action.includes("added")) return "green";
  if (action.includes("updated") || action.includes("sync")) return "blue";
  return "gray";
}
