import { InlineLoading, Tile } from "@carbon/react";
import type { ReactNode } from "react";

import { TableEmptyState } from "./TableEmptyState";
import "./DataTableShell.scss";

export type DataTableHeader = {
  key: string;
  header: string;
};

type DataTableShellProps<T> = {
  title: string;
  description?: string;
  rows: T[];
  headers: DataTableHeader[];
  getRowId: (row: T) => string;
  isLoading?: boolean;
  isFetching?: boolean;
  loadingDescription?: string;
  tableState?: ReactNode;
  emptyTitle?: string;
  emptyDescription?: string;
  emptyActions?: ReactNode;
  searchPlaceholder?: string;
  onSearch?: (value: string) => void;
  toolbarActions?: ReactNode;
  topContent?: ReactNode;
  filters?: ReactNode;
  bulkActions?: ReactNode;
  children: (args: {
    rows: Array<T & { id: string }>;
    headers: DataTableHeader[];
  }) => ReactNode;
};

export function DataTableShell<T extends object>({
  title,
  description,
  rows,
  headers,
  getRowId,
  isLoading = false,
  loadingDescription = "Loading...",
  tableState,
  emptyTitle,
  emptyDescription,
  emptyActions,
  filters,
  topContent,
  bulkActions,
  children,
}: DataTableShellProps<T>) {
  return (
    <div className="moh-data-table-shell">
      <div className="moh-data-table-shell__header">
        <h3 className="moh-data-table-shell__title">{title}</h3>
        {description && <p className="moh-data-table-shell__description">{description}</p>}
      </div>

      {topContent}

      {filters && <Tile className="moh-data-table-shell__filters">{filters}</Tile>}

      <Tile className="moh-data-table-shell__table">
        {tableState ? (
          tableState
        ) : isLoading ? (
          <InlineLoading description={loadingDescription} />
        ) : rows.length === 0 && emptyTitle ? (
          <TableEmptyState title={emptyTitle} description={emptyDescription} actions={emptyActions} />
        ) : (
          <>
            {bulkActions}
            <div className="moh-data-table-shell__table-scroll">
              {children({
                rows: rows.map((row) => ({ ...row, id: getRowId(row) })),
                headers,
              })}
            </div>
          </>
        )}
      </Tile>
    </div>
  );
}
