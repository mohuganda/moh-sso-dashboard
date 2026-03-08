import {
  DataTable,
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  TableContainer,
  Button,
  Tag,
  InlineLoading,
} from "@carbon/react";
import { useGetSessionsQuery, useLogoutSessionMutation } from "../../store/api/sessions.api";

const headers = [
  { key: "ip", header: "IP Address" },
  { key: "started", header: "Started" },
  { key: "lastAccess", header: "Last Access" },
  { key: "clients", header: "Applications" },
  { key: "action", header: "" },
];

export default function UserSessionsTable() {
  const { data: sessions = [], isLoading } = useGetSessionsQuery();
  const [logoutSession] = useLogoutSessionMutation();

  const formatDate = (timestamp: number) => new Date(timestamp).toLocaleString();

  const rows = sessions.map((s) => ({
    id: s.id,
    ip: s.ipAddress,
    started: formatDate(s.start),
    lastAccess: formatDate(s.lastAccess),
    clients: Object.keys(s.clients),
  }));

  if (isLoading) return <InlineLoading description="Loading sessions..." />;

  return (
    <TableContainer title="Active Sessions">
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
              {rows.map((row) => (
                <TableRow {...getRowProps({ row })}>
                  <TableCell>{row.cells[0].value}</TableCell>

                  <TableCell>{row.cells[1].value}</TableCell>

                  <TableCell>{row.cells[2].value}</TableCell>

                  <TableCell>
                    {row.cells[3].value.map((c: string) => (
                      <Tag key={c} type="cyan" size="sm">
                        {c}
                      </Tag>
                    ))}
                  </TableCell>

                  <TableCell>
                    <Button size="sm" kind="danger--ghost" onClick={() => logoutSession(row.id)}>
                      Logout
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </DataTable>
    </TableContainer>
  );
}
