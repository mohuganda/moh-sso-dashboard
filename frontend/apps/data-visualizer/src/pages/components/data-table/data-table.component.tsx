import React from 'react';
import {
  DataTable,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
  TableExpandHeader,
  TableExpandRow,
  TableExpandedRow,
} from "@carbon/react";
import IssueDetail from "../../../../../issue-tracker/src/pages/issue-detail/issue-detail.component.tsx";


interface ListProps {
  columns: any;
  data: any;
  selectedIssue?: any;
  closeView: () => void;
  handleIssueClick: (row: any) => void;
}

const DataList: React.FC<ListProps> = ({ columns, data, handleIssueClick, closeView }) => {
  return (
      <DataTable rows={data} headers={columns}>
        {({
            rows,
            headers,
            getTableProps,
            getHeaderProps,
            getRowProps,
            getExpandHeaderProps,
          }) => (
            <Table {...getTableProps()}>
              <TableHead>
                <TableRow>
                  <TableExpandHeader {...getExpandHeaderProps()} />
                  {headers.map((header) => {
                    const { key: headerKey, ...headerProps } = getHeaderProps({ header });
                    return (
                      <TableHeader key={headerKey} {...headerProps}>
                        {header.header}
                      </TableHeader>
                    );
                  })}
                </TableRow>
              </TableHead>
              <TableBody>
                {rows.map((row) => {
                  const issue = data.find((item: any) => item.id === row.id);
                  const { key: rowKey, ...rowProps } = getRowProps({ row });
                  return (
                      <React.Fragment key={row.id}>
                        <TableExpandRow key={rowKey} {...rowProps}>
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
                            return <TableCell key={cell.id}>{cell.value}</TableCell>;
                          })}
                        </TableExpandRow>
                        <TableExpandedRow colSpan={headers.length + 1}>
                          {row.isExpanded && (
                              <section>
                                <IssueDetail selectedIssue={issue} goToBack={closeView} showBack={false}/>
                              </section>
                          )}
                        </TableExpandedRow>
                      </React.Fragment>
                  )
                })}
              </TableBody>
            </Table>
        )}
      </DataTable>
  );
};

export default DataList;