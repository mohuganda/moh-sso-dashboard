import { useMemo, useState } from "react";
import {
  Button,
  DataTable,
  InlineLoading,
  InlineNotification,
  Tag,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  Tile,
} from "@carbon/react";

import { useGetSessionsQuery, useLogoutSessionMutation } from "@moh-sso/api";

import "./user-sessions-table.scss";

const headers = [
  { key: "ip", header: "IP Address" },
  { key: "started", header: "Started" },
  { key: "lastAccess", header: "Last Access" },
  { key: "clients", header: "Applications" },
  { key: "action", header: "" },
];

function formatDate(timestamp?: number): string {
  if (!timestamp) {
    return "—";
  }

  const value = timestamp < 10_000_000_000 ? timestamp * 1000 : timestamp;
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return "—";
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export default function UserSessionsTable() {
  const { data: sessions = [], isLoading, isFetching, isError, refetch } = useGetSessionsQuery();

  const [logoutSession, { isLoading: isLoggingOut }] = useLogoutSessionMutation();

  const [activeLogoutId, setActiveLogoutId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState("");

  const rows = useMemo(() => {
    return sessions.map((session) => ({
      id: session.id,
      ip: session.ipAddress || "—",
      started: formatDate(session.start),
      lastAccess: formatDate(session.lastAccess),
      clients: Object.keys(session.clients ?? {}),
    }));
  }, [sessions]);

  const handleLogoutSession = async (sessionId: string) => {
    setErrorMessage("");
    setActiveLogoutId(sessionId);

    try {
      await logoutSession(sessionId).unwrap();
      await refetch();
    } catch (error) {
      console.error(error);
      setErrorMessage("Unable to end this session. Please try again.");
    } finally {
      setActiveLogoutId(null);
    }
  };

  if (isLoading) {
    return <InlineLoading description="Loading sessions..." />;
  }

  if (isError) {
    return (
      <Tile className="sessions-state">
        <h4>Unable to load active sessions</h4>
        <p>Please check your connection and try again.</p>

        <Button size="sm" kind="ghost" onClick={() => refetch()}>
          Retry
        </Button>
      </Tile>
    );
  }

  return (
    <div className="sessions-table">
      {errorMessage && (
        <InlineNotification
          kind="error"
          title="Session logout failed"
          subtitle={errorMessage}
          lowContrast
          onClose={() => setErrorMessage("")}
        />
      )}

      {isFetching && !isLoading && <InlineLoading description="Refreshing sessions..." />}

      <TableContainer
        title="Active Sessions"
        description="Review and end active login sessions connected to your account."
      >
        <DataTable rows={rows} headers={headers}>
          {({ rows, headers, getHeaderProps, getRowProps }) => (
            <Table size="sm">
              <TableHead>
                <TableRow>
                  {headers.map((header) => (
                    <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                  ))}
                </TableRow>
              </TableHead>

              <TableBody>
                {rows.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={headers.length}>No active sessions found.</TableCell>
                  </TableRow>
                ) : (
                  rows.map((row) => {
                    const clients = row.cells.find((cell) => cell.info.header === "clients")
                      ?.value as string[];

                    return (
                      <TableRow {...getRowProps({ row })}>
                        <TableCell>
                          {row.cells.find((cell) => cell.info.header === "ip")?.value}
                        </TableCell>

                        <TableCell>
                          {row.cells.find((cell) => cell.info.header === "started")?.value}
                        </TableCell>

                        <TableCell>
                          {row.cells.find((cell) => cell.info.header === "lastAccess")?.value}
                        </TableCell>

                        <TableCell>
                          <div className="sessions-client-tags">
                            {clients.length > 0 ? (
                              clients.map((client) => (
                                <Tag key={client} type="cyan" size="sm">
                                  {client}
                                </Tag>
                              ))
                            ) : (
                              <span className="sessions-muted">—</span>
                            )}
                          </div>
                        </TableCell>

                        <TableCell>
                          <Button
                            size="sm"
                            kind="danger--ghost"
                            disabled={isLoggingOut}
                            onClick={() => handleLogoutSession(row.id)}
                          >
                            {activeLogoutId === row.id ? "Ending..." : "End session"}
                          </Button>
                        </TableCell>
                      </TableRow>
                    );
                  })
                )}
              </TableBody>
            </Table>
          )}
        </DataTable>
      </TableContainer>
    </div>
  );
}
