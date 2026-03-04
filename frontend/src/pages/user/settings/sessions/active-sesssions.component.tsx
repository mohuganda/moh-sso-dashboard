import {
  Table,
  TableHead,
  TableRow,
  TableHeader,
  TableBody,
  TableCell,
  Button,
  Tile,
} from "@carbon/react";

export default function ActiveSessionsPage() {
  const { data: sessions = [] } = useGetSessionsQuery();
  const [terminateSession] = useTerminateSessionMutation();

  return (
    <Tile>
      <h3>Active Sessions</h3>

      <Table>
        <TableHead>
          <TableRow>
            <TableHeader>IP Address</TableHeader>
            <TableHeader>Client</TableHeader>
            <TableHeader>Last Access</TableHeader>
            <TableHeader>Action</TableHeader>
          </TableRow>
        </TableHead>

        <TableBody>
          {sessions.map((session) => (
            <TableRow key={session.id}>
              <TableCell>{session.ipAddress}</TableCell>
              <TableCell>{session.clientId}</TableCell>
              <TableCell>{session.lastAccess}</TableCell>
              <TableCell>
                <Button kind="danger--ghost" size="sm" onClick={() => terminateSession(session.id)}>
                  Terminate
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </Tile>
  );
}
