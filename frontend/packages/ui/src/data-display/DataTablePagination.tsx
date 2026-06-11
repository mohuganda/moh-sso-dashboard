import { Pagination } from "@carbon/react";

type DataTablePaginationProps = {
  page: number;
  pageSize: number;
  totalItems: number;
  pageSizes?: number[];
  onChange: (pagination: { page: number; pageSize: number }) => void;
};

export function DataTablePagination({
  page,
  pageSize,
  totalItems,
  pageSizes = [10, 20, 30, 50],
  onChange,
}: DataTablePaginationProps) {
  return (
    <Pagination
      page={page}
      pageSize={pageSize}
      pageSizes={pageSizes}
      totalItems={totalItems}
      onChange={onChange}
    />
  );
}
