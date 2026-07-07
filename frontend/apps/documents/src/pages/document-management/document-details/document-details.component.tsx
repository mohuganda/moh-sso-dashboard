import { useMemo, useState, type ReactNode } from "react";
import { useSelector } from "react-redux";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import {
  Button,
  FilterableMultiSelect,
  InlineLoading,
  InlineNotification,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  TableToolbarSearch,
  Tabs,
  Tag,
} from "@carbon/react";
import { Download, Renew, TrashCan } from "@carbon/react/icons";

import { PermissionGuard, PERMISSIONS, selectUser } from "@moh-sso/auth";
import { useModal, useToast } from "@moh-sso/ui";
import { useGetUserQuery } from "@moh-sso/users/api";

import {
  useDeleteDocumentMutation,
  useGetDocumentDataPreviewQuery,
  useGetDocumentProcessesQuery,
  useGetDocumentQuery,
  useGetTemplateStructureQuery,
  useLazyDownloadDocumentQuery,
  useReprocessDocumentMutation,
} from "../../../api";
import type { DataPreviewSheet, DocumentProcess, DocumentResponse } from "../../../types";

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
  sheet: DataPreviewSheet;
  reportDate?: string;
  filterableKeys: string[];
};

function SheetView({ sheet, reportDate, filterableKeys }: SheetViewProps) {
  const [search, setSearch] = useState("");
  const [selections, setSelections] = useState<Record<string, string[]>>({});
  const [sortColumn, setSortColumn] = useState<string | null>(null);
  const [sortDirection, setSortDirection] = useState<"asc" | "desc" | null>(null);
  const [page, setPage] = useState(0);

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

  const filteredRows = useMemo(() => {
    const query = search.trim().toLowerCase();

    let rows = sheet.rows;

    for (const key of activeFilterKeys) {
      const selectedValues = selections[key] ?? [];

      if (selectedValues.length === 0) {
        continue;
      }

      const selectedSet = new Set(selectedValues);

      rows = rows.filter((row) => selectedSet.has(formatCell(row[key])));
    }

    if (query) {
      rows = rows.filter((row) =>
        sheet.columns.some((column) => formatCell(row[column]).toLowerCase().includes(query)),
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
  }, [sheet.rows, sheet.columns, search, selections, activeFilterKeys, sortColumn, sortDirection]);

  const totalPages = Math.max(1, Math.ceil(filteredRows.length / ROWS_PER_PAGE));

  const safePageIndex = Math.min(page, totalPages - 1);

  const pageRows = filteredRows.slice(
    safePageIndex * ROWS_PER_PAGE,
    (safePageIndex + 1) * ROWS_PER_PAGE,
  );

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
    setSearch("");
    setSelections({});
    setSortColumn(null);
    setSortDirection(null);
    setPage(0);
  }

  const hasActiveFilters =
    Boolean(search) ||
    Object.values(selections).some((selected) => selected.length > 0) ||
    Boolean(sortColumn);

  const csvFilename = `${sheet.name.replace(/\s+/g, "_")}${reportDate ? `_${reportDate}` : ""}${
    hasActiveFilters ? "_filtered" : ""
  }.csv`;

  return (
    <div
      style={{
        minWidth: 0,
        width: "100%",
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "flex-end",
          gap: "0.75rem",
          flexWrap: "wrap",
          padding: "0.75rem",
          background: "#f4f4f4",
          borderRadius: "4px 4px 0 0",
          borderBottom: "1px solid #e0e0e0",
        }}
      >
        <div
          style={{
            flex: "1 1 180px",
            minWidth: 160,
          }}
        >
          <TableToolbarSearch
            persistent
            value={search}
            placeholder="Search…"
            onChange={(_, value) => {
              setSearch(value ?? "");
              setPage(0);
            }}
            labelText="Search"
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
            <div
              key={key}
              style={{
                flex: "0 1 280px",
                minWidth: 200,
              }}
            >
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
          <button
            type="button"
            onClick={clearFilters}
            style={{
              fontSize: "0.8rem",
              padding: "0.35rem 0.6rem",
              border: "1px solid #c6c6c6",
              borderRadius: 3,
              background: "#fff",
              cursor: "pointer",
              color: "#525252",
              whiteSpace: "nowrap",
              alignSelf: "flex-end",
            }}
          >
            Clear
          </button>
        )}

        <div
          style={{
            marginLeft: "auto",
            display: "flex",
            alignItems: "center",
            gap: "0.75rem",
          }}
        >
          <span
            style={{
              fontSize: "0.8rem",
              color: "#6f6f6f",
              whiteSpace: "nowrap",
            }}
          >
            {filteredRows.length.toLocaleString()} / {sheet.row_count.toLocaleString()} rows
          </span>

          <PermissionGuard permission={PERMISSIONS.documentsRead}>
            <Button
              renderIcon={Download}
              kind="primary"
              size="sm"
              onClick={() => exportCSV(sheet.columns, filteredRows, csvFilename)}
              disabled={filteredRows.length === 0}
            >
              {hasActiveFilters ? "Download filtered" : "Download all"}
            </Button>
          </PermissionGuard>
        </div>
      </div>

      <div
        style={{
          overflowX: "auto",
          border: "1px solid #e0e0e0",
          borderTop: "none",
        }}
      >
        <table
          style={{
            borderCollapse: "collapse",
            fontSize: "0.8125rem",
            width: "max-content",
            minWidth: "100%",
          }}
        >
          <thead>
            <tr
              style={{
                background: "#e0e0e0",
              }}
            >
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
                    style={{
                      textAlign: "left",
                      padding: "0.5rem 0.875rem",
                      fontWeight: 600,
                      color: "#161616",
                      whiteSpace: "nowrap",
                      cursor: "pointer",
                      userSelect: "none",
                      position: "sticky",
                      top: 0,
                      background: active ? "#d0e2ff" : "#e0e0e0",
                      borderBottom: active ? "2px solid #0f62fe" : "2px solid #c6c6c6",
                      borderRight: "1px solid #c6c6c6",
                    }}
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
                <td
                  colSpan={sheet.columns.length}
                  style={{
                    padding: "2rem",
                    textAlign: "center",
                    color: "#6f6f6f",
                  }}
                >
                  No rows match the current filters.
                </td>
              </tr>
            ) : (
              pageRows.map((row, rowIndex) => (
                <tr
                  key={rowIndex}
                  style={{
                    borderBottom: "1px solid #e8e8e8",
                    background: rowIndex % 2 ? "#fafafa" : "#fff",
                  }}
                >
                  {sheet.columns.map((column) => {
                    const value = formatCell(row[column]);

                    return (
                      <td
                        key={column}
                        style={{
                          padding: "0.4rem 0.875rem",
                          whiteSpace: "nowrap",
                          color: value ? "#161616" : "#c6c6c6",
                          borderRight: "1px solid #f0f0f0",
                        }}
                      >
                        {value || "—"}
                      </td>
                    );
                  })}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {totalPages > 1 && (
        <div
          style={{
            display: "flex",
            gap: "0.5rem",
            alignItems: "center",
            padding: "0.75rem",
            justifyContent: "center",
            background: "#f4f4f4",
            border: "1px solid #e0e0e0",
            borderTop: "none",
          }}
        >
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

          <span
            style={{
              fontSize: "0.875rem",
              color: "#525252",
            }}
          >
            Page {safePageIndex + 1} of {totalPages}
            <span
              style={{
                color: "#8d8d8d",
                marginLeft: "0.5rem",
              }}
            >
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
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        padding: "0.75rem 1.25rem",
        background: index % 2 === 0 ? "#fff" : "#f9f9f9",
        borderBottom: isLast ? "none" : "1px solid #e8e8e8",
      }}
    >
      <span
        style={{
          width: 160,
          minWidth: 160,
          fontSize: "0.8125rem",
          color: "#6f6f6f",
          fontWeight: 500,
        }}
      >
        {label}
      </span>

      <span
        style={{
          fontSize: "0.875rem",
          color: "#161616",
          fontWeight: 400,
          overflowWrap: "anywhere",
        }}
      >
        {value}
      </span>
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

  const {
    data: preview,
    isLoading: previewLoading,
    isError: previewError,
  } = useGetDocumentDataPreviewQuery(id ?? "", {
    skip: !id || !templateCode,
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

  const latest = useMemo(() => getLatestProcess(processes), [processes]);

  const status = (latest?.status || document?.status || "UNKNOWN").toUpperCase();

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
    return <InlineLoading description="Loading document…" style={{ padding: "2rem" }} />;
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
        <span
          style={{
            color: "#da1e28",
          }}
        >
          {latest.error ?? latest.message}
        </span>
      ),
    });
  }

  return (
    <>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          gap: "1.25rem",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: "0.75rem",
            flexWrap: "wrap",
          }}
        >
          <h2
            style={{
              margin: 0,
              fontSize: "1.25rem",
              fontWeight: 600,
              overflowWrap: "anywhere",
            }}
          >
            {document.original_filename}
          </h2>

          {processesRefreshing && (
            <span
              style={{
                fontSize: "0.8rem",
                color: "#6f6f6f",
              }}
            >
              Refreshing…
            </span>
          )}
        </div>

        <Tabs>
          <TabList aria-label="Document detail tabs" contained>
            {templateCode && <Tab>Data Preview</Tab>}

            <Tab>Document Info</Tab>
          </TabList>

          <TabPanels>
            {templateCode && (
              <TabPanel
                style={{
                  paddingInline: 0,
                  paddingTop: "1.25rem",
                  minWidth: 0,
                  overflow: "hidden",
                }}
              >
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

                          <Tag
                            type="gray"
                            style={{
                              marginLeft: "0.5rem",
                            }}
                          >
                            {sheet.row_count.toLocaleString()}
                          </Tag>
                        </Tab>
                      ))}
                    </TabList>

                    <TabPanels>
                      {preview.sheets.map((sheet) => (
                        <TabPanel
                          key={sheet.name}
                          style={{
                            paddingInline: 0,
                            paddingTop: "0.75rem",
                          }}
                        >
                          <SheetView
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

            <TabPanel
              style={{
                paddingInline: 0,
                paddingTop: "1.5rem",
              }}
            >
              <div
                style={{
                  border: "1px solid #e0e0e0",
                  borderRadius: 4,
                  overflow: "hidden",
                  marginBottom: "1.5rem",
                }}
              >
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

              <div
                style={{
                  display: "flex",
                  gap: "0.75rem",
                  flexWrap: "wrap",
                }}
              >
                <PermissionGuard permission={PERMISSIONS.documentsRead}>
                  <button
                    type="button"
                    onClick={() => void handleDownloadFile()}
                    disabled={status !== "COMPLETED" || isBusy}
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      alignItems: "center",
                      gap: "0.3rem",
                      padding: "0.6rem 1rem",
                      border: "1px solid #0f62fe",
                      borderRadius: 4,
                      background: "#fff",
                      cursor: status !== "COMPLETED" || isBusy ? "not-allowed" : "pointer",
                      opacity: status !== "COMPLETED" || isBusy ? 0.4 : 1,
                    }}
                  >
                    <Download
                      size={16}
                      style={{
                        color: "#0f62fe",
                      }}
                    />

                    <span
                      style={{
                        fontSize: "0.75rem",
                        fontWeight: 600,
                        color: "#0f62fe",
                      }}
                    >
                      Download
                    </span>
                  </button>
                </PermissionGuard>

                {processable && (
                  <PermissionGuard permission={PERMISSIONS.documentsProcess}>
                    <button
                      type="button"
                      onClick={() => void handleReprocess()}
                      disabled={status !== "FAILED" || isBusy}
                      style={{
                        display: "flex",
                        flexDirection: "column",
                        alignItems: "center",
                        gap: "0.3rem",
                        padding: "0.6rem 1rem",
                        border: "1px solid #393939",
                        borderRadius: 4,
                        background: "#fff",
                        cursor: status !== "FAILED" || isBusy ? "not-allowed" : "pointer",
                        opacity: status !== "FAILED" || isBusy ? 0.4 : 1,
                      }}
                    >
                      <Renew
                        size={16}
                        style={{
                          color: "#393939",
                        }}
                      />

                      <span
                        style={{
                          fontSize: "0.75rem",
                          fontWeight: 600,
                          color: "#393939",
                        }}
                      >
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
                    style={{
                      display: "flex",
                      flexDirection: "column",
                      alignItems: "center",
                      gap: "0.3rem",
                      padding: "0.6rem 1rem",
                      border: "1px solid #da1e28",
                      borderRadius: 4,
                      background: "#fff",
                      cursor: isBusy ? "not-allowed" : "pointer",
                      opacity: isBusy ? 0.4 : 1,
                    }}
                  >
                    <TrashCan
                      size={16}
                      style={{
                        color: "#da1e28",
                      }}
                    />

                    <span
                      style={{
                        fontSize: "0.75rem",
                        fontWeight: 600,
                        color: "#da1e28",
                      }}
                    >
                      Delete
                    </span>
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

          <div
            style={{
              display: "flex",
              justifyContent: "flex-end",
              marginTop: "1.5rem",
            }}
          >
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
