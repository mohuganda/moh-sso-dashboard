import { InlineLoading, Search, Tile } from "@carbon/react";
import type { ReactNode } from "react";

import { PageShell } from "../layout/PageShell/PageShell";
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
  searchValue?: string;
  resultsDescription?: string;
  searchPlaceholder?: string;
  onSearch?: (value: string) => void;
  toolbarActions?: ReactNode;
  topContent?: ReactNode;
  filters?: ReactNode;
  bulkActions?: ReactNode;
  children: (args: { rows: Array<T & { id: string }>; headers: DataTableHeader[] }) => ReactNode;
};

export function DataTableShell<T extends object>({
  title,
  description,
  rows,
  headers,
  getRowId,
  isLoading = false,
  isFetching = false,
  searchValue,
  searchPlaceholder,
  onSearch,
  resultsDescription,
  toolbarActions,
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
      <PageShell title={title} subtitle={description} actions={toolbarActions}>
        {topContent}

        {(filters || onSearch) && (
          <Tile className="moh-data-table-shell__filters" role="group" aria-label={`Filter ${title.toLowerCase()}`}>
            {onSearch && (
              <Search
                id={`table-search-${title.toLowerCase().replace(/\s+/g, "-")}`}
                labelText={`Search ${title.toLowerCase()}`}
                placeholder={searchPlaceholder}
                value={searchValue}
                onChange={(event) => onSearch(event.target.value)}
              />
            )}
            {filters}
          </Tile>
        )}
        {resultsDescription && (
          <p className="moh-data-table-shell__results" role="status">
            {resultsDescription}
          </p>
        )}
        {isFetching && !isLoading && <InlineLoading description="Refreshing results…" />}

        <Tile className="moh-data-table-shell__table">
          {tableState ? (
            tableState
          ) : isLoading ? (
            <InlineLoading description={loadingDescription} />
          ) : rows.length === 0 && emptyTitle ? (
            <TableEmptyState
              title={emptyTitle}
              description={emptyDescription}
              actions={emptyActions}
            />
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
      </PageShell>
    </div>
  );
}
