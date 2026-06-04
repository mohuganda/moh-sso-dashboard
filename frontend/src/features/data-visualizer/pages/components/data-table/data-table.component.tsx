import {
  DataTable,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@carbon/react";

interface ListProps {
  columns: any;
  data: any;
  handleIssueClick: (row: any) => void;
}

const DataList: React.FC<ListProps> = ({ columns, data, handleIssueClick }) => {
  return (
    <DataTable rows={data} headers={columns}>
      {({ rows, headers, getTableProps, getHeaderProps, getRowProps, getCellProps }) => (
        <Table {...getTableProps()}>
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
                  if (cell.info.header === "issue") {
                    return (
                      <TableCell key={cell.id}>
                        <button
                          type="button"
                          className="issue-clickable-cell"
                          onClick={() => handleIssueClick(row)}
                        >
                          {cell.value}
                        </button>
                      </TableCell>
                    );
                  }
                  return <TableCell {...getCellProps({ cell })}>{cell.value}</TableCell>;
                })}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </DataTable>
  );
};

export default DataList;
