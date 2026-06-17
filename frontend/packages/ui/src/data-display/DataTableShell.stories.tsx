import {
  DataTable,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";

import { DataTableShell } from "./DataTableShell";
import { TableStatusTag } from "./TableStatusTag";

export default {
  title: "Data Display/DataTableShell",
};

const headers = [
  { key: "name", header: "Name" },
  { key: "status", header: "Status" },
];

type StoryRow = {
  id: string;
  name: string;
  status: string;
};

const rows: StoryRow[] = [
  { id: "1", name: "Client Registry", status: "Active" },
  { id: "2", name: "Audit Worker", status: "Disabled" },
];

const emptyRows: StoryRow[] = [];

export const Default = () => {
  return (
    <DataTableShell
      title="Applications"
      description="Reusable table frame for feature-owned data."
      rows={rows}
      headers={headers}
      getRowId={(row) => row.id}
    >
      {({ rows, headers }) => (
        <DataTable rows={rows} headers={headers}>
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
                    {row.cells.map((cell) => (
                      <TableCell key={cell.id}>
                        {cell.info.header === "status" ? (
                          <TableStatusTag status={String(cell.value)} />
                        ) : (
                          cell.value
                        )}
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </DataTable>
      )}
    </DataTableShell>
  );
};

export const Empty = () => {
  return (
    <DataTableShell
      title="Applications"
      rows={emptyRows}
      headers={headers}
      getRowId={(row) => row.id}
      emptyTitle="No applications found"
      emptyDescription="Adjust filters or add a new record."
    >
      {() => null}
    </DataTableShell>
  );
};
