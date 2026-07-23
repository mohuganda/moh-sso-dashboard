import { useMemo, useState, type ReactNode } from "react";
import { useSelector } from "react-redux";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import {
  Button,
  Checkbox,
  FilterableMultiSelect,
  InlineLoading,
  InlineNotification,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Tag,
} from "@carbon/react";
import { Download, Renew, TrashCan } from "@carbon/react/icons";

import { PermissionGuard, PERMISSIONS, selectUser } from "@moh-sso/auth";
import { useModal, useToast } from "@moh-sso/ui";
import { useGetUserQuery } from "@moh-sso/users/api";

import {
  useDeleteDocumentMutation,
  useExportDataPreviewMutation,
  useGetDocumentDataPreviewQuery,
  useGetDocumentProcessesQuery,
  useGetDocumentQuery,
  useGetTemplateStructureQuery,
  useLazyDownloadDocumentQuery,
  useReprocessDocumentMutation,
} from "../../../api";
import type { DataPreviewSheet, DocumentProcess, DocumentResponse } from "../../../types";
import "../../../components/documents-components.scss";

function formatDateTime(iso?: string | null) {
  if (!iso) {
    return "—";
  }

  const dateValue = new Date(iso);

  if (Number.isNaN(dateValue.getTime())) {
    return "—";
  }

  const date = dateValue.toLocaleDateString("en-GB", {
    day: "numeric",
    month: "long",
    year: "numeric",
  });

  const time = dateValue.toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
  });

  return `${date} at ${time}`;
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  const formattedValue = value < 10 && unitIndex > 0 ? value.toFixed(1) : Math.round(value);

  return `${formattedValue} ${units[unitIndex]}`;
}

function formatFileType(contentType: string, filename?: string) {
  switch (contentType?.toLowerCase()) {
    case "application/pdf":
      return "PDF";

    case "text/csv":
      return "CSV";

    case "application/vnd.ms-excel":
      return "XLS";

    case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
      return "XLSX";

    default:
      break;
  }

  if (filename) {
    const extension = filename.split(".").pop()?.toUpperCase();

    if (extension && extension.length <= 5) {
      return extension;
    }
  }

  return contentType || "—";
}

function formatDuration(started?: string | null, finished?: string | null) {
  if (!started || !finished) {
    return "—";
  }

  const milliseconds = new Date(finished).getTime() - new Date(started).getTime();

  if (milliseconds < 0) {
    return "—";
  }

  if (milliseconds < 1000) {
    return `${milliseconds}ms`;
  }

  const seconds = Math.floor(milliseconds / 1000);

  if (seconds < 60) {
    return `${seconds}s`;
  }

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  return remainingSeconds > 0 ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`;
}

function formatCell(value: unknown): string {
  if (value === null || value === undefined) {
    return "";
  }

  if (typeof value === "string") {
    return value;
  }

  return String(value);
}

function StatusTag({ status }: { status: string }) {
  const statusTypeMap: Record<string, "gray" | "blue" | "green" | "red" | "magenta"> = {
    PENDING: "gray",
    PROCESSING: "blue",
    COMPLETED: "green",
    FAILED: "red",
    CANCELLED: "magenta",
  };

  return <Tag type={statusTypeMap[status] ?? "gray"}>{status}</Tag>;
}

function getLatestProcess(processes: DocumentProcess[]) {
  if (processes.length === 0) {
    return undefined;
  }

  return [...processes].sort(
    (a, b) => new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime(),
  )[0];
}

function requiresProcessing(document?: Partial<DocumentResponse>) {
  if (!document) {
    return false;
  }

  const contentType = (document.content_type || "").toLowerCase();

  const filename = (document.original_filename || "").toLowerCase();

  return (
    [
      "text/csv",
      "application/vnd.ms-excel",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ].includes(contentType) ||
    filename.endsWith(".csv") ||
    filename.endsWith(".xls") ||
    filename.endsWith(".xlsx")
  );
}

function exportCSV(columns: string[], rows: Record<string, unknown>[], filename: string) {
  const escapeValue = (value: unknown) => {
    const formatted = formatCell(value);

    return formatted.includes(",") || formatted.includes('"') || formatted.includes("\n")
      ? `"${formatted.replace(/"/g, '""')}"`
      : formatted;
  };

  const header = columns.map(escapeValue).join(",");

  const body = rows
    .map((row) => columns.map((column) => escapeValue(row[column])).join(","))
    .join("\n");

  const blob = new Blob([`${header}\n${body}`], {
    type: "text/csv;charset=utf-8;",
  });

  const url = window.URL.createObjectURL(blob);

  const anchor = window.document.createElement("a");

  anchor.href = url;
  anchor.download = filename;

  window.document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();

  window.URL.revokeObjectURL(url);
}

const ROWS_PER_PAGE = 100;

type SheetViewProps = {
  documentId: string;
  sheet: DataPreviewSheet;
  reportDate?: string;
  filterableKeys: string[];
};

function getRowHash(row: Record<string, unknown>): string | undefined {
  const value = row._row_hash;
  return typeof value === "string" ? value : undefined;
}

function downloadBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob);
  const anchor = window.document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  window.document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  window.URL.revokeObjectURL(url);
}

function SheetView({ documentId, sheet, reportDate, filterableKeys }: SheetViewProps) {
  const toast = useToast();
  const [exportDataPreview, { isLoading: isExporting }] = useExportDataPreviewMutation();

  const [searchValues, setSearchValues] = useState<string[]>([]);
  const [selections, setSelections] = useState<Record<string, string[]>>({});
  const [sortColumn, setSortColumn] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc" | null>(null);
  const [page, setPage] = useState(0);
  const [selectedHashes, setSelectedHashes] = useState<Set<string>>(new Set());

  const activeFilterKeys = useMemo(
    () => filterableKeys.filter((key) => sheet.columns.includes(key)),
    [filterableKeys, sheet.columns],
  );

  const filterOptions = useMemo(() => {
    const result: Record<
      string,
      {
        id: string;
        label: string;
      }[]
    > = {};

    for (const key of activeFilterKeys) {
      const uniqueValues = [
        ...new Set(sheet.rows.map((row) => formatCell(row[key])).filter(Boolean)),
      ].sort();

      result[key] = uniqueValues.map((value) => ({
        id: value,
        label: value,
      }));
    }

    return result;
  }, [sheet.rows, activeFilterKeys]);

  const searchOptions = useMemo(() => {
    const unique = new Set<string>();
    for (const row of sheet.rows) {
      for (const column of sheet.columns) {
        const value = formatCell(row[column]);
        if (value) unique.add(value);
      }
    }
    return [...unique].sort().map((value) => ({ id: value, label: value }));
  }, [sheet.rows, sheet.columns]);

  const filteredRows = useMemo(() => {
    let rows = sheet.rows;

    for (const key of activeFilterKeys) {
      const selectedValues = selections[key] ?? [];

      if (selectedValues.length === 0) {
        continue;
      }

      const selectedSet = new Set(selectedValues);

      rows = rows.filter((row) => selectedSet.has(formatCell(row[key])));
    }

    if (searchValues.length > 0) {
      const searchSet = new Set(searchValues);
      rows = rows.filter((row) =>
        sheet.columns.some((column) => searchSet.has(formatCell(row[column]))),
      );
    }

    if (sortColumn && sortDirection) {
      rows = [...rows].sort((a, b) => {
        const firstValue = formatCell(a[sortColumn]);

        const secondValue = formatCell(b[sortColumn]);

        const bothNumeric =
          firstValue !== "" &&
          secondValue !== "" &&
          !Number.isNaN(Number(firstValue)) &&
          !Number.isNaN(Number(secondValue));

        const comparison = bothNumeric
          ? Number(firstValue) - Number(secondValue)
          : firstValue.localeCompare(secondValue, undefined, {
              sensitivity: "base",
            });

        return sortDirection === "asc" ? comparison : -comparison;
      });
    }

    return rows;
  }, [
    sheet.rows,
    sheet.columns,
    searchValues,
    selections,
    activeFilterKeys,
    sortColumn,
    sortDirection,
  ]);

  const totalPages = Math.max(1, Math.ceil(filteredRows.length / ROWS_PER_PAGE));

  const safePageIndex = Math.min(page, totalPages - 1);

  const pageRows = filteredRows.slice(
    safePageIndex * ROWS_PER_PAGE,
    (safePageIndex + 1) * ROWS_PER_PAGE,
  );

  const pageHashes = useMemo(
    () => pageRows.map(getRowHash).filter((hash): hash is string => !!hash),
    [pageRows],
  );
  const allPageSelected = pageHashes.length > 0 && pageHashes.every((h) => selectedHashes.has(h));
  const somePageSelected = pageHashes.some((h) => selectedHashes.has(h));

  function toggleRow(hash: string | undefined) {
    if (!hash) return;
    setSelectedHashes((prev) => {
      const next = new Set(prev);
      if (next.has(hash)) next.delete(hash);
      else next.add(hash);
      return next;
    });
  }

  function togglePage() {
    setSelectedHashes((prev) => {
      const next = new Set(prev);
      if (allPageSelected) {
        for (const h of pageHashes) next.delete(h);
      } else {
        for (const h of pageHashes) next.add(h);
      }
      return next;
    });
  }

  function handleSort(column: string) {
    if (sortColumn !== column) {
      setSortColumn(column);
      setSortDirection("asc");
    } else if (sortDirection === "asc") {
      setSortDirection("desc");
    } else {
      setSortColumn(null);
      setSortDirection(null);
    }

    setPage(0);
  }

  function clearFilters() {
    setSearchValues([]);
    setSelections({});
    setSortColumn(null);
    setSortDirection(null);
    setPage(0);
  }

  const hasActiveFilters =
    searchValues.length > 0 ||
    Object.values(selections).some((selected) => selected.length > 0) ||
    Boolean(sortColumn);

  const baseFilename = `${sheet.name.replace(/\s+/g, "_")}${reportDate ? `_${reportDate}` : ""}`;

  function downloadSelected() {
    const selectedRows = sheet.rows.filter((row) => {
      const hash = getRowHash(row);
      return hash && selectedHashes.has(hash);
    });
    if (selectedRows.length === 0) return;

    exportCSV(sheet.columns, selectedRows, `${baseFilename}_selected.csv`);
    toast.success("Download started", `${selectedRows.length} selected rows`);
  }

  async function downloadAllMatching() {
    const activeFilters = Object.fromEntries(
      Object.entries(selections).filter(([, values]) => values.length > 0),
    );
    try {
      const blob = await exportDataPreview({
        id: documentId,
        sheetCode: sheet.code || sheet.name,
        columns: sheet.columns,
        search: searchValues.length > 0 ? searchValues : undefined,
        filters: Object.keys(activeFilters).length > 0 ? activeFilters : undefined,
      }).unwrap();
      downloadBlob(blob, `${baseFilename}_export.csv`);
      toast.success("Export started", "All matching rows from the server");
    } catch {
      toast.error("Export failed", "Please try again.");
    }
  }

  return (
    <div className="document-sheet-view">
      <div className="document-sheet-view__toolbar">
        <div className="document-sheet-view__search">
          <FilterableMultiSelect
            id={`search-${sheet.name}`}
            titleText="Search values"
            placeholder="Type to search…"
            items={searchOptions}
            itemToString={(item) => item?.label ?? ""}
            selectedItems={searchValues.map((value) => ({ id: value, label: value }))}
            onChange={({ selectedItems }) => {
              setSearchValues((selectedItems ?? []).map((item) => item.id));
              setPage(0);
            }}
          />
        </div>

        {activeFilterKeys.map((key) => {
          const options = filterOptions[key] ?? [];

          if (options.length === 0) {
            return null;
          }

          const label = key
            .replace(/_/g, " ")
            .replace(/\b\w/g, (character) => character.toUpperCase());

          const selectedItems = (selections[key] ?? []).map((value) => ({
            id: value,
            label: value,
          }));

          return (
            <div key={key} className="document-sheet-view__filter">
              <FilterableMultiSelect
                id={`filter-${sheet.name}-${key}`}
                titleText={label}
                placeholder="Type to search…"
                items={options}
                itemToString={(item) => item?.label ?? ""}
                selectedItems={selectedItems}
                onChange={({ selectedItems: changedItems }) => {
                  setSelections((previous) => ({
                    ...previous,
                    [key]: (changedItems ?? []).map((item) => item.id),
                  }));

                  setPage(0);
                }}
                size="sm"
              />
            </div>
          );
        })}

        {hasActiveFilters && (
          <button type="button" onClick={clearFilters} className="document-sheet-view__clear">
            Clear
          </button>
        )}

        <div className="document-sheet-view__actions">
          <span className="document-sheet-view__row-count">
            {filteredRows.length.toLocaleString()} / {sheet.row_count.toLocaleString()} rows
          </span>
        </div>
      </div>

      <div className="document-preview__selection-bar">
        <span className="document-preview__selection-count">
          {selectedHashes.size > 0
            ? `${selectedHashes.size.toLocaleString()} row${selectedHashes.size === 1 ? "" : "s"} selected`
            : "Select rows to download only those"}
        </span>

        <PermissionGuard permission={PERMISSIONS.documentsRead}>
          <Button
            kind="secondary"
            size="sm"
            renderIcon={Download}
            disabled={selectedHashes.size === 0}
            onClick={downloadSelected}
          >
            Download selected
          </Button>

          <Button
            kind="ghost"
            size="sm"
            renderIcon={Download}
            disabled={isExporting}
            onClick={() => void downloadAllMatching()}
          >
            {isExporting
              ? "Exporting…"
              : hasActiveFilters
                ? "Export all rows matching filters"
                : "Export all rows"}
          </Button>
        </PermissionGuard>
      </div>

      <div className="document-sheet-view__table-scroll">
        <table className="document-sheet-view__table">
          <thead>
            <tr>
              <th className="document-preview__checkbox-cell">
                <Checkbox
                  id={`select-page-${sheet.name}`}
                  labelText="Select all rows on this page"
                  hideLabel
                  checked={allPageSelected}
                  indeterminate={!allPageSelected && somePageSelected}
                  onChange={togglePage}
                  disabled={pageHashes.length === 0}
                />
              </th>
              {sheet.columns.map((column) => {
                const active = sortColumn === column;

                const arrow = active ? (sortDirection === "asc" ? " ▲" : " ▼") : "";

                const label = column
                  .replace(/_/g, " ")
                  .replace(/\b\w/g, (character) => character.toUpperCase());

                return (
                  <th
                    key={column}
                    onClick={() => handleSort(column)}
                    title={`Sort by ${label}`}
                    className={active ? "document-sheet-view__sort-header--active" : undefined}
                  >
                    {label}
                    {arrow}
                  </th>
                );
              })}
            </tr>
          </thead>

          <tbody>
            {pageRows.length === 0 ? (
              <tr>
                <td colSpan={sheet.columns.length + 1} className="document-sheet-view__no-rows">
                  No rows match the current filters.
                </td>
              </tr>
            ) : (
              pageRows.map((row, rowIndex) => {
                const hash = getRowHash(row);
                const isSelected = !!hash && selectedHashes.has(hash);
                return (
                  <tr
                    key={hash ?? rowIndex}
                    className={isSelected ? "document-preview__row--selected" : undefined}
                  >
                    <td className="document-preview__checkbox-cell">
                      <Checkbox
                        id={`select-row-${sheet.name}-${hash ?? rowIndex}`}
                        labelText="Select row"
                        hideLabel
                        checked={isSelected}
                        onChange={() => toggleRow(hash)}
                        disabled={!hash}
                      />
                    </td>
                    {sheet.columns.map((column) => {
                      const value = formatCell(row[column]);

                      return (
                        <td
                          key={column}
                          className={!value ? "document-sheet-view__empty-cell" : undefined}
                        >
                          {value || "—"}
                        </td>
                      );
                    })}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div className="document-sheet-view__pagination">
          <Button kind="ghost" size="sm" disabled={safePageIndex === 0} onClick={() => setPage(0)}>
            «
          </Button>

          <Button
            kind="ghost"
            size="sm"
            disabled={safePageIndex === 0}
            onClick={() => setPage((current) => Math.max(current - 1, 0))}
          >
            ‹
          </Button>

          <span className="document-sheet-view__page-label">
            Page {safePageIndex + 1} of {totalPages}
            <span className="document-sheet-view__page-range">
              ({(safePageIndex * ROWS_PER_PAGE + 1).toLocaleString()}–
              {Math.min((safePageIndex + 1) * ROWS_PER_PAGE, filteredRows.length).toLocaleString()})
            </span>
          </span>

          <Button
            kind="ghost"
            size="sm"
            disabled={safePageIndex >= totalPages - 1}
            onClick={() => setPage((current) => Math.min(current + 1, totalPages - 1))}
          >
            ›
          </Button>

          <Button
            kind="ghost"
            size="sm"
            disabled={safePageIndex >= totalPages - 1}
            onClick={() => setPage(totalPages - 1)}
          >
            »
          </Button>
        </div>
      )}
    </div>
  );
}

function InfoItem({
  label,
  value,
  isLast,
  index,
}: {
  label: string;
  value: ReactNode;
  isLast: boolean;
  index: number;
}) {
  void index;

  return (
    <div
      className={["document-details-info-item", isLast ? "document-details-info-item--last" : ""]
        .filter(Boolean)
        .join(" ")}
    >
      <span className="document-details-info-item__label">{label}</span>

      <span className="document-details-info-item__value">{value}</span>
    </div>
  );
}

function DocumentDetailsPageContent() {
  const toast = useToast();
  const { openModal, closeModal } = useModal();

  const { id } = useParams<{ id: string }>();

  const navigate = useNavigate();

  const {
    data: document,
    isLoading: documentLoading,
    error: documentError,
  } = useGetDocumentQuery(id ?? "", {
    skip: !id,
  });

  const processable = requiresProcessing(document);

  const templateCode = document?.metadata?.template_code ?? "";

  const reportDate = document?.metadata?.report_date ?? "";

  const { data: processes = [], isFetching: processesRefreshing } = useGetDocumentProcessesQuery(
    id ?? "",
    {
      skip: !id || !processable,
      pollingInterval: processable ? 3000 : 0,
    },
  );

  const latest = useMemo(() => getLatestProcess(processes), [processes]);

  const status = (latest?.status || document?.status || "UNKNOWN").toUpperCase();
  const isProcessing = status === "PENDING" || status === "PROCESSING";

  const {
    data: preview,
    isLoading: previewLoading,
    isError: previewError,
  } = useGetDocumentDataPreviewQuery(id ?? "", {
    skip: !id || !processable,
    pollingInterval: isProcessing ? 3000 : 0,
  });

  const { data: structure } = useGetTemplateStructureQuery(templateCode, {
    skip: !templateCode,
  });

  const filterableKeysBySheet = useMemo(() => {
    const map: Record<string, string[]> = {};

    if (!structure) {
      return map;
    }

    for (const sheet of structure.sheets) {
      const keys = sheet.columns
        .filter((column) => column.configuration?.filterable === true)
        .map((column) => column.column_key);

      map[sheet.name] = keys;

      if (sheet.display_name && sheet.display_name !== sheet.name) {
        map[sheet.display_name] = keys;
      }
    }

    return map;
  }, [structure]);

  const currentUser = useSelector(selectUser);

  const isOwnUpload = Boolean(document) && currentUser?.id === document?.uploaded_by;

  const { data: uploaderUser, isLoading: isLoadingUploader } = useGetUserQuery(
    document?.uploaded_by ?? "",
    {
      skip: !document?.uploaded_by || isOwnUpload,
    },
  );

  let uploaderName = "—";

  if (isOwnUpload && currentUser) {
    uploaderName =
      [currentUser.firstName, currentUser.lastName].filter(Boolean).join(" ") ||
      currentUser.username ||
      "—";
  } else if (isLoadingUploader) {
    uploaderName = "Loading…";
  } else if (uploaderUser) {
    uploaderName =
      [uploaderUser.firstName, uploaderUser.lastName].filter(Boolean).join(" ") ||
      uploaderUser.username ||
      "—";
  }

  const [deleteDocument, { isLoading: deleting }] = useDeleteDocumentMutation();

  const [triggerDownload] = useLazyDownloadDocumentQuery();

  const [reprocessDocument, { isLoading: reprocessing }] = useReprocessDocumentMutation();

  async function handleDownloadFile() {
    if (!id || !document) {
      return;
    }

    try {
      const blob = await triggerDownload(id).unwrap();

      const url = window.URL.createObjectURL(blob);

      const anchor = window.document.createElement("a");

      anchor.href = url;
      anchor.download = document.original_filename;

      window.document.body.appendChild(anchor);

      anchor.click();
      anchor.remove();

      window.URL.revokeObjectURL(url);
      toast.success("Download started", document.original_filename);
    } catch (error: unknown) {
      const message =
        typeof error === "object" && error !== null && "data" in error
          ? String(
              (
                error as {
                  data?: {
                    error?: unknown;
                  };
                }
              ).data?.error ?? "Failed to download",
            )
          : "Failed to download";

      toast.error("Download failed", message);
    }
  }

  async function handleReprocess() {
    if (!id) {
      return;
    }

    try {
      await reprocessDocument(id).unwrap();

      toast.success("Reprocess started", "The document is being reprocessed.");
    } catch (error: unknown) {
      const message =
        typeof error === "object" && error !== null && "data" in error
          ? String(
              (
                error as {
                  data?: {
                    error?: unknown;
                  };
                }
              ).data?.error ?? "Failed to reprocess",
            )
          : "Failed to reprocess";

      toast.error("Reprocess failed", message);
    }
  }

  async function handleDeleteConfirm() {
    if (!id) {
      return;
    }

    try {
      await deleteDocument(id).unwrap();

      closeModal();
      toast.success("Deleted", "Document deleted successfully.");

      navigate("..");
    } catch (error: unknown) {
      const message =
        typeof error === "object" && error !== null && "data" in error
          ? String(
              (
                error as {
                  data?: {
                    error?: unknown;
                  };
                }
              ).data?.error ?? "Failed to delete",
            )
          : "Failed to delete";

      toast.error("Delete failed", message);
    }
  }

  function handleDeleteRequest() {
    if (!document) {
      return;
    }

    openModal({
      title: "Delete document",
      onClose: closeModal,
      content: (
        <p>
          Permanently delete <strong>{document.original_filename}</strong>? All imported rows will
          also be removed. This cannot be undone.
        </p>
      ),
      primaryAction: {
        label: deleting ? "Deleting..." : "Delete",
        kind: "danger",
        disabled: deleting,
        onClick: () => void handleDeleteConfirm(),
      },
      secondaryAction: {
        label: "Cancel",
        onClick: closeModal,
      },
    });
  }

  if (documentLoading) {
    return <InlineLoading description="Loading document…" className="documents-loading-block" />;
  }

  if (documentError || !document) {
    return (
      <InlineNotification
        kind="error"
        title="Not found"
        subtitle="Document could not be loaded."
        lowContrast
      />
    );
  }

  const isBusy = deleting || reprocessing;

  const informationItems: {
    label: string;
    value: ReactNode;
  }[] = [
    {
      label: "Status",
      value: <StatusTag status={status} />,
    },
    {
      label: "Report Date",
      value: reportDate || "—",
    },
    {
      label: "Template",
      value: templateCode || "—",
    },
    {
      label: "File Type",
      value: formatFileType(document.content_type, document.original_filename),
    },
    {
      label: "File Size",
      value: formatFileSize(document.size_bytes),
    },
    {
      label: "Uploaded",
      value: formatDateTime(document.created_at),
    },
    {
      label: "Uploaded by",
      value: uploaderName,
    },
    {
      label: "Processing Time",
      value: formatDuration(latest?.started_at, latest?.finished_at),
    },
  ];

  if (status === "FAILED" && (latest?.error || latest?.message)) {
    informationItems.push({
      label: "Failure reason",
      value: (
        <span className="document-details-page__failure">{latest.error ?? latest.message}</span>
      ),
    });
  }

  return (
    <>
      <div className="document-details-page">
        <div className="document-details-page__title-row">
          <h2 className="document-details-page__title">{document.original_filename}</h2>

          {processesRefreshing && (
            <span className="document-details-page__refreshing">Refreshing…</span>
          )}
        </div>

        <Tabs>
          <TabList aria-label="Document detail tabs" contained>
            {processable && <Tab>Data Preview</Tab>}

            <Tab>Document Info</Tab>
          </TabList>

          <TabPanels>
            {processable && (
              <TabPanel className="document-details-page__tab-panel">
                {previewLoading ? (
                  <InlineLoading description="Loading imported data…" />
                ) : previewError ? (
                  <InlineNotification
                    kind="error"
                    title="Preview failed"
                    subtitle="Could not load imported data."
                    lowContrast
                  />
                ) : !preview || preview.sheets.length === 0 ? (
                  <InlineNotification
                    kind="info"
                    title="No data available"
                    subtitle="No imported rows found. The file may still be processing."
                    lowContrast
                  />
                ) : (
                  <Tabs>
                    <TabList aria-label="Sheet tabs">
                      {preview.sheets.map((sheet) => (
                        <Tab key={sheet.name}>
                          {sheet.name}

                          <Tag type="gray" className="document-details-page__tab-tag">
                            {sheet.row_count.toLocaleString()}
                          </Tag>
                        </Tab>
                      ))}
                    </TabList>

                    <TabPanels>
                      {preview.sheets.map((sheet) => (
                        <TabPanel key={sheet.name} className="document-preview__tab-panel">
                          <SheetView
                            documentId={id ?? ""}
                            sheet={sheet}
                            reportDate={reportDate}
                            filterableKeys={filterableKeysBySheet[sheet.name] ?? []}
                          />
                        </TabPanel>
                      ))}
                    </TabPanels>
                  </Tabs>
                )}
              </TabPanel>
            )}

            <TabPanel className="document-details-page__tab-panel document-details-page__tab-panel--info">
              <div className="document-details-page__info-list">
                {informationItems.map((item, index) => (
                  <InfoItem
                    key={item.label}
                    label={item.label}
                    value={item.value}
                    index={index}
                    isLast={index === informationItems.length - 1}
                  />
                ))}
              </div>

              <div className="document-details-page__actions">
                <PermissionGuard permission={PERMISSIONS.documentsRead}>
                  <button
                    type="button"
                    onClick={() => void handleDownloadFile()}
                    disabled={status !== "COMPLETED" || isBusy}
                    className="document-details-action document-details-action--primary"
                  >
                    <Download size={16} />

                    <span className="document-details-action__label">Download</span>
                  </button>
                </PermissionGuard>

                {processable && (
                  <PermissionGuard permission={PERMISSIONS.documentsProcess}>
                    <button
                      type="button"
                      onClick={() => void handleReprocess()}
                      disabled={status !== "FAILED" || isBusy}
                      className="document-details-action document-details-action--neutral"
                    >
                      <Renew size={16} />

                      <span className="document-details-action__label">
                        {reprocessing ? "Reprocessing…" : "Reprocess"}
                      </span>
                    </button>
                  </PermissionGuard>
                )}

                <PermissionGuard permission={PERMISSIONS.documentsWrite}>
                  <button
                    type="button"
                    onClick={handleDeleteRequest}
                    disabled={isBusy}
                    className="document-details-action document-details-action--danger"
                  >
                    <TrashCan size={16} />

                    <span className="document-details-action__label">Delete</span>
                  </button>
                </PermissionGuard>
              </div>
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </>
  );
}

export default function DocumentDetailsPage() {
  return (
    <PermissionGuard
      permission={PERMISSIONS.documentsRead}
      fallback={
        <div>
          <InlineNotification
            kind="warning"
            title="Access denied"
            subtitle="You do not have permission to view this document."
            lowContrast
            hideCloseButton
          />

          <div className="document-permission-fallback__actions">
            <Button kind="secondary" as={RouterLink} to="..">
              Back to Documents
            </Button>
          </div>
        </div>
      }
    >
      <DocumentDetailsPageContent />
    </PermissionGuard>
  );
}
