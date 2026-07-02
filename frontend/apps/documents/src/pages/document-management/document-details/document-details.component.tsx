import { useMemo, useState } from "react";
import { useSelector } from "react-redux";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
  Button,
  InlineLoading,
  InlineNotification,
  Modal,
  FilterableMultiSelect,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  TableToolbarSearch,
  Tabs,
  Tag,
} from "@carbon/react";
import { ArrowLeft, Download, Renew, TrashCan } from "@carbon/icons-react";

import { useToast } from "@moh-sso/ui";
import {
  useDeleteDocumentMutation,
  useGetDocumentDataPreviewQuery,
  useGetDocumentProcessesQuery,
  useGetDocumentQuery,
  useLazyDownloadDocumentQuery,
  useReprocessDocumentMutation,
  useGetTemplateStructureQuery,
} from "../../../api";
import { useGetUserQuery } from "@moh-sso/users/api";
import { selectUser } from "@moh-sso/auth";
import type { DataPreviewSheet, DocumentProcess, DocumentResponse } from "../../../types";

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatDateTime(iso?: string | null) {
  if (!iso) return "—";
  const d = new Date(iso);
  const date = d.toLocaleDateString("en-GB", { day: "numeric", month: "long", year: "numeric" });
  const time = d.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
  return `${date} at ${time}`;
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let v = bytes; let i = 0;
  while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
  return `${v < 10 && i > 0 ? v.toFixed(1) : Math.round(v)} ${units[i]}`;
}

function formatFileType(contentType: string, filename?: string) {
  switch (contentType) {
    case "application/pdf": return "PDF";
    case "text/csv": return "CSV";
    case "application/vnd.ms-excel": return "XLS";
    case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": return "XLSX";
  }
  if (filename) {
    const ext = filename.split(".").pop()?.toUpperCase();
    if (ext && ext.length <= 5) return ext;
  }
  return contentType || "—";
}

function formatDuration(started?: string | null, finished?: string | null) {
  if (!started || !finished) return "—";
  const ms = new Date(finished).getTime() - new Date(started).getTime();
  if (ms < 0) return "—";
  if (ms < 1000) return `${ms}ms`;
  const s = Math.floor(ms / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  return m > 0 ? `${m}m ${s % 60}s` : `${s}s`;
}

function formatCell(v: unknown): string {
  if (v === null || v === undefined) return "";
  if (typeof v === "string") return v;
  return String(v);
}

function StatusTag({ status }: { status: string }) {
  const map: Record<string, "gray" | "blue" | "green" | "red" | "magenta"> = {
    PENDING: "gray", PROCESSING: "blue", COMPLETED: "green", FAILED: "red", CANCELLED: "magenta",
  };
  return <Tag type={map[status] ?? "gray"}>{status}</Tag>;
}

function getLatestProcess(processes: DocumentProcess[]) {
  if (!processes.length) return undefined;
  return [...processes].sort(
    (a, b) => new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime(),
  )[0];
}

function requiresProcessing(doc?: Partial<DocumentResponse>) {
  if (!doc) return false;
  const t = (doc.content_type || "").toLowerCase();
  const f = (doc.original_filename || "").toLowerCase();
  return ["text/csv", "application/vnd.ms-excel",
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"].includes(t) ||
    f.endsWith(".csv") || f.endsWith(".xls") || f.endsWith(".xlsx");
}

function InfoRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div style={{ display: "flex", gap: "0.5rem", padding: "0.5rem 0", borderBottom: "1px solid #f4f4f4" }}>
      <span style={{ width: 180, minWidth: 180, color: "#6f6f6f", fontSize: "0.875rem" }}>{label}</span>
      <span style={{ fontSize: "0.875rem", color: "#161616", wordBreak: "break-all" }}>{value || "—"}</span>
    </div>
  );
}

// ── CSV export ────────────────────────────────────────────────────────────────

function exportCSV(columns: string[], rows: Record<string, unknown>[], filename: string) {
  const escape = (v: unknown) => {
    const s = formatCell(v);
    return s.includes(",") || s.includes('"') || s.includes("\n")
      ? `"${s.replace(/"/g, '""')}"` : s;
  };
  const header = columns.map(escape).join(",");
  const body = rows.map((r) => columns.map((c) => escape(r[c])).join(",")).join("\n");
  const blob = new Blob([`${header}\n${body}`], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = window.document.createElement("a");
  a.href = url; a.download = filename; a.click();
  URL.revokeObjectURL(url);
}

// ── Sheet view ────────────────────────────────────────────────────────────────

const ROWS_PER_PAGE = 100;

function SheetView({ sheet, reportDate: docReportDate, filterableKeys }: {
  sheet: DataPreviewSheet;
  reportDate?: string;
  filterableKeys: string[];
}) {
  const [search, setSearch] = useState("");
  const [selections, setSelections] = useState<Record<string, string[]>>({});
  const [sortCol, setSortCol] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<"asc" | "desc" | null>(null);
  const [page, setPage] = useState(0);

  // Keys that are both marked filterable and present in this sheet
  const activeFilterKeys = useMemo(
    () => filterableKeys.filter((k) => sheet.columns.includes(k)),
    [filterableKeys, sheet.columns],
  );

  // Unique values per filterable key for building the dropdowns
  const filterOptions = useMemo(() => {
    const result: Record<string, { id: string; label: string }[]> = {};
    for (const key of activeFilterKeys) {
      const unique = [...new Set(sheet.rows.map((r) => formatCell(r[key])).filter(Boolean))].sort();
      result[key] = unique.map((v) => ({ id: v, label: v }));
    }
    return result;
  }, [sheet.rows, activeFilterKeys]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    let rows = sheet.rows;

    for (const key of activeFilterKeys) {
      const sel = selections[key] ?? [];
      if (sel.length === 0) continue;
      const set = new Set(sel);
      rows = rows.filter((r) => set.has(formatCell(r[key])));
    }

    if (q) {
      rows = rows.filter((r) =>
        sheet.columns.some((col) => formatCell(r[col]).toLowerCase().includes(q)),
      );
    }

    if (sortCol && sortDir) {
      rows = [...rows].sort((a, b) => {
        const av = formatCell(a[sortCol]); const bv = formatCell(b[sortCol]);
        const isNum = !isNaN(Number(av)) && !isNaN(Number(bv)) && av !== "" && bv !== "";
        const cmp = isNum ? Number(av) - Number(bv) : av.localeCompare(bv, undefined, { sensitivity: "base" });
        return sortDir === "asc" ? cmp : -cmp;
      });
    }

    return rows;
  }, [sheet.rows, search, selections, activeFilterKeys, sortCol, sortDir]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / ROWS_PER_PAGE));
  const safePageIdx = Math.min(page, totalPages - 1);
  const pageRows = filtered.slice(safePageIdx * ROWS_PER_PAGE, (safePageIdx + 1) * ROWS_PER_PAGE);

  function handleSort(col: string) {
    if (sortCol !== col) { setSortCol(col); setSortDir("asc"); }
    else if (sortDir === "asc") setSortDir("desc");
    else { setSortCol(null); setSortDir(null); }
    setPage(0);
  }

  function clearFilters() {
    setSearch(""); setSelections({});
    setSortCol(null); setSortDir(null); setPage(0);
  }

  const hasActiveFilters = !!(search || Object.values(selections).some((s) => s.length > 0) || sortCol);

  const csvFilename = `${sheet.name.replace(/\s+/g, "_")}${docReportDate ? `_${docReportDate}` : ""}${hasActiveFilters ? "_filtered" : ""}.csv`;

  return (
    <div style={{ minWidth: 0, width: "100%" }}>
      {/* Toolbar */}
      <div style={{
        display: "flex", alignItems: "flex-end", gap: "0.75rem",
        flexWrap: "wrap", padding: "0.75rem",
        background: "#f4f4f4", borderRadius: "4px 4px 0 0",
        borderBottom: "1px solid #e0e0e0",
      }}>
        {/* Search */}
        <div style={{ flex: "1 1 180px", minWidth: 160 }}>
          <TableToolbarSearch
            persistent
            value={search}
            placeholder="Search…"
            onChange={(_, v) => { setSearch(v ?? ""); setPage(0); }}
            labelText="Search"
          />
        </div>

        {/* Dynamic filterable-column dropdowns */}
        {activeFilterKeys.map((key) => {
          const opts = filterOptions[key] ?? [];
          if (opts.length === 0) return null;
          const label = key.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
          const selected = (selections[key] ?? []).map((v) => ({ id: v, label: v }));
          return (
            <div key={key} style={{ flex: "0 1 280px", minWidth: 200 }}>
              <FilterableMultiSelect
                id={`filter-${sheet.name}-${key}`}
                titleText={label}
                placeholder="Type to search…"
                label={selected.length > 0 ? `${selected.length} selected` : `All ${label.toLowerCase()}`}
                items={opts}
                itemToString={(item) => item?.label ?? ""}
                selectedItems={selected}
                onChange={({ selectedItems }) => {
                  setSelections((prev) => ({ ...prev, [key]: (selectedItems ?? []).map((i) => i.id) }));
                  setPage(0);
                }}
                size="sm"
              />
            </div>
          );
        })}

        {hasActiveFilters && (
          <button
            onClick={clearFilters}
            style={{
              fontSize: "0.8rem", padding: "0.35rem 0.6rem",
              border: "1px solid #c6c6c6", borderRadius: 3,
              background: "#fff", cursor: "pointer", color: "#525252",
              whiteSpace: "nowrap", alignSelf: "flex-end",
            }}
          >
            Clear
          </button>
        )}

        {/* Row count + Download — pushed right */}
        <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: "0.75rem" }}>
          <span style={{ fontSize: "0.8rem", color: "#6f6f6f", whiteSpace: "nowrap" }}>
            {filtered.length.toLocaleString()} / {sheet.row_count.toLocaleString()} rows
          </span>
          <Button
            renderIcon={Download}
            kind="primary"
            size="sm"
            onClick={() => exportCSV(sheet.columns, filtered, csvFilename)}
            disabled={filtered.length === 0}
          >
            {hasActiveFilters ? "Download filtered" : "Download all"}
          </Button>
        </div>
      </div>

      {/* Table — scrolls horizontally so every column is fully visible */}
      <div style={{ overflowX: "auto", border: "1px solid #e0e0e0", borderTop: "none" }}>
        <table style={{ borderCollapse: "collapse", fontSize: "0.8125rem", width: "max-content", minWidth: "100%" }}>
          <thead>
            <tr style={{ background: "#e0e0e0" }}>
              {sheet.columns.map((col) => {
                const active = sortCol === col;
                const arrow = active ? (sortDir === "asc" ? " ▲" : " ▼") : "";
                const label = col.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
                return (
                  <th
                    key={col}
                    onClick={() => handleSort(col)}
                    title={`Sort by ${label}`}
                    style={{
                      textAlign: "left", padding: "0.5rem 0.875rem", fontWeight: 600,
                      color: "#161616", whiteSpace: "nowrap", cursor: "pointer",
                      userSelect: "none", position: "sticky", top: 0,
                      background: active ? "#d0e2ff" : "#e0e0e0",
                      borderBottom: active ? "2px solid #0f62fe" : "2px solid #c6c6c6",
                      borderRight: "1px solid #c6c6c6",
                    }}
                  >
                    {label}{arrow}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {pageRows.length === 0 ? (
              <tr>
                <td colSpan={sheet.columns.length}
                  style={{ padding: "2rem", textAlign: "center", color: "#6f6f6f" }}>
                  No rows match the current filters.
                </td>
              </tr>
            ) : (
              pageRows.map((row, idx) => (
                <tr key={idx} style={{ borderBottom: "1px solid #e8e8e8", background: idx % 2 ? "#fafafa" : "#fff" }}>
                  {sheet.columns.map((col) => {
                    const val = formatCell(row[col]);
                    return (
                      <td key={col}
                        style={{
                          padding: "0.4rem 0.875rem",
                          whiteSpace: "nowrap",
                          color: val ? "#161616" : "#c6c6c6",
                          borderRight: "1px solid #f0f0f0",
                        }}
                      >
                        {val || "—"}
                      </td>
                    );
                  })}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div style={{
          display: "flex", gap: "0.5rem", alignItems: "center",
          padding: "0.75rem", justifyContent: "center", background: "#f4f4f4",
          border: "1px solid #e0e0e0", borderTop: "none",
        }}>
          <Button kind="ghost" size="sm" disabled={safePageIdx === 0} onClick={() => setPage(0)}>«</Button>
          <Button kind="ghost" size="sm" disabled={safePageIdx === 0} onClick={() => setPage((p) => p - 1)}>‹</Button>
          <span style={{ fontSize: "0.875rem", color: "#525252" }}>
            Page {safePageIdx + 1} of {totalPages}
            <span style={{ color: "#8d8d8d", marginLeft: "0.5rem" }}>
              ({(safePageIdx * ROWS_PER_PAGE + 1).toLocaleString()}–{Math.min((safePageIdx + 1) * ROWS_PER_PAGE, filtered.length).toLocaleString()})
            </span>
          </span>
          <Button kind="ghost" size="sm" disabled={safePageIdx >= totalPages - 1} onClick={() => setPage((p) => p + 1)}>›</Button>
          <Button kind="ghost" size="sm" disabled={safePageIdx >= totalPages - 1} onClick={() => setPage(totalPages - 1)}>»</Button>
        </div>
      )}
    </div>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function DocumentDetailsPage() {
  const toast = useToast();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [confirmDelete, setConfirmDelete] = useState(false);

  const { data: document, isLoading: docLoading, error: docError } = useGetDocumentQuery(id!, { skip: !id });
  const processable = requiresProcessing(document);
  const templateCode = document?.metadata?.template_code ?? "";
  const reportDate = document?.metadata?.report_date ?? "";

  const { data: processes = [], isFetching: processesRefreshing } = useGetDocumentProcessesQuery(id!, {
    skip: !id || !processable,
    pollingInterval: processable ? 3000 : 0,
  });

  const { data: preview, isLoading: previewLoading, isError: previewError } =
    useGetDocumentDataPreviewQuery(id!, { skip: !id || !templateCode });

  const { data: structure } = useGetTemplateStructureQuery(templateCode, { skip: !templateCode });

  const filterableKeysBySheet = useMemo(() => {
    if (!structure) return {} as Record<string, string[]>;
    const map: Record<string, string[]> = {};
    for (const s of structure.sheets) {
      const keys = s.columns
        .filter((c) => c.configuration?.filterable === true)
        .map((c) => c.column_key);
      map[s.name] = keys;
      if (s.display_name && s.display_name !== s.name) map[s.display_name] = keys;
    }
    return map;
  }, [structure]);

  const currentUser = useSelector(selectUser);
  const isOwnUpload = !!document && currentUser?.id === document.uploaded_by;

  const { data: uploaderUser, isLoading: isLoadingUploader } = useGetUserQuery(
    document?.uploaded_by ?? "",
    { skip: !document?.uploaded_by || isOwnUpload },
  );

  const uploaderName = isOwnUpload && currentUser
    ? [currentUser.firstName, currentUser.lastName].filter(Boolean).join(" ") || currentUser.username
    : isLoadingUploader
      ? "Loading…"
      : uploaderUser
        ? [uploaderUser.firstName, uploaderUser.lastName].filter(Boolean).join(" ") || uploaderUser.username
        : "—";

  const [deleteDocument, { isLoading: deleting }] = useDeleteDocumentMutation();
  const [triggerDownload] = useLazyDownloadDocumentQuery();
  const [reprocessDocument, { isLoading: reprocessing }] = useReprocessDocumentMutation();

  const latest = useMemo(() => getLatestProcess(processes), [processes]);
  const status = (latest?.status || document?.status || "UNKNOWN").toUpperCase();

  const handleDownloadFile = async () => {
    if (!id || !document) return;
    try {
      const blob = await triggerDownload(id).unwrap();
      const url = window.URL.createObjectURL(blob);
      const a = window.document.createElement("a");
      a.href = url; a.download = document.original_filename; a.click();
      URL.revokeObjectURL(url);
    } catch (err: unknown) {
      toast.error("Download failed", String((err as { data?: { error?: string } })?.data?.error ?? "Failed to download"));
    }
  };

  const handleReprocess = async () => {
    if (!id) return;
    try {
      await reprocessDocument(id).unwrap();
      toast.success("Reprocess started", "The document is being reprocessed.");
    } catch (err: unknown) {
      toast.error("Reprocess failed", String((err as { data?: { error?: string } })?.data?.error ?? "Failed to reprocess"));
    }
  };

  const handleDeleteConfirm = async () => {
    if (!id) return;
    try {
      await deleteDocument(id).unwrap();
      toast.success("Deleted", "Document deleted successfully.");
      navigate("/apps/utilities/self-service/eservice/document-upload");
    } catch (err: unknown) {
      toast.error("Delete failed", String((err as { data?: { error?: string } })?.data?.error ?? "Failed to delete"));
    }
  };

  if (docLoading) return <InlineLoading description="Loading document…" style={{ padding: "2rem" }} />;
  if (docError || !document) return (
    <InlineNotification kind="error" title="Not found" subtitle="Document could not be loaded." lowContrast />
  );

  const isBusy = deleting || reprocessing;

  return (
    <>
      {/* Delete confirmation */}
      <Modal
        open={confirmDelete}
        danger
        modalHeading="Delete document"
        primaryButtonText={deleting ? "Deleting…" : "Delete"}
        secondaryButtonText="Cancel"
        primaryButtonDisabled={deleting}
        onRequestClose={() => setConfirmDelete(false)}
        onRequestSubmit={() => void handleDeleteConfirm()}
      >
        <p>
          Permanently delete <strong>{document.original_filename}</strong>?
          All imported rows will also be removed. This cannot be undone.
        </p>
      </Modal>

      <div style={{ display: "flex", flexDirection: "column", gap: "1.25rem" }}>
        {/* Breadcrumb */}
        <Breadcrumb noTrailingSlash>
          <BreadcrumbItem><RouterLink to="/">Home</RouterLink></BreadcrumbItem>
          <BreadcrumbItem>
            <RouterLink to="/apps/utilities/self-service/eservice/document-upload">Documents</RouterLink>
          </BreadcrumbItem>
          <BreadcrumbItem isCurrentPage>{document.original_filename}</BreadcrumbItem>
        </Breadcrumb>

        {/* Header — back + title only */}
        <div style={{ display: "flex", alignItems: "center", gap: "0.75rem" }}>
          <Button kind="ghost" size="sm" renderIcon={ArrowLeft}
            onClick={() => navigate(-1)} style={{ paddingLeft: 0 }}>
            Back
          </Button>
          <h2 style={{ margin: 0, fontSize: "1.25rem", fontWeight: 600 }}>
            {document.original_filename}
          </h2>
          {processesRefreshing && (
            <span style={{ fontSize: "0.8rem", color: "#6f6f6f" }}>Refreshing…</span>
          )}
        </div>

        {/* Tabs */}
        <Tabs>
          <TabList aria-label="Document detail tabs" contained>
            {templateCode && <Tab>Data Preview</Tab>}
            <Tab>Document Info</Tab>
          </TabList>

          <TabPanels>
            {/* ── Data Preview ── */}
            {templateCode && (
              <TabPanel style={{ paddingInline: 0, paddingTop: "1.25rem", minWidth: 0, overflow: "hidden" }}>
                {previewLoading ? (
                  <InlineLoading description="Loading imported data…" />
                ) : previewError ? (
                  <InlineNotification kind="error" title="Preview failed"
                    subtitle="Could not load imported data." lowContrast />
                ) : !preview || preview.sheets.length === 0 ? (
                  <InlineNotification kind="info" title="No data available"
                    subtitle="No imported rows found. The file may still be processing."
                    lowContrast />
                ) : (
                  <Tabs>
                    <TabList aria-label="Sheet tabs">
                      {preview.sheets.map((sheet) => (
                        <Tab key={sheet.name}>
                          {sheet.name}
                          <Tag type="gray" style={{ marginLeft: "0.5rem" }}>
                            {sheet.row_count.toLocaleString()}
                          </Tag>
                        </Tab>
                      ))}
                    </TabList>
                    <TabPanels>
                      {preview.sheets.map((sheet) => (
                        <TabPanel key={sheet.name} style={{ paddingInline: 0, paddingTop: "0.75rem" }}>
                          <SheetView sheet={sheet} reportDate={reportDate} filterableKeys={filterableKeysBySheet[sheet.name] ?? []} />
                        </TabPanel>
                      ))}
                    </TabPanels>
                  </Tabs>
                )}
              </TabPanel>
            )}

            {/* ── Document Info ── */}
            <TabPanel style={{ paddingInline: 0, paddingTop: "1.5rem" }}>

              {/* Key-value list */}
              <div style={{ border: "1px solid #e0e0e0", borderRadius: 4, overflow: "hidden", marginBottom: "1.5rem" }}>
                {[
                  { label: "Status",          value: <StatusTag status={status} /> },
                  { label: "Report Date",     value: reportDate || "—" },
                  { label: "Template",        value: templateCode || "—" },
                  { label: "File Type",       value: formatFileType(document.content_type, document.original_filename) },
                  { label: "File Size",       value: formatFileSize(document.size_bytes) },
                  { label: "Uploaded",        value: formatDateTime(document.created_at) },
                  { label: "Uploaded by",     value: uploaderName },
                  { label: "Processing Time", value: formatDuration(latest?.started_at, latest?.finished_at) },
                  ...(status === "FAILED" && (latest?.error || latest?.message) ? [{
                    label: "Failure reason",
                    value: <span style={{ color: "#da1e28" }}>{latest?.error ?? latest?.message}</span>,
                  }] : []),
                ].map(({ label, value }, idx, arr) => (
                  <div key={label} style={{
                    display: "flex",
                    alignItems: "center",
                    padding: "0.75rem 1.25rem",
                    background: idx % 2 === 0 ? "#fff" : "#f9f9f9",
                    borderBottom: idx < arr.length - 1 ? "1px solid #e8e8e8" : "none",
                  }}>
                    <span style={{ width: 160, minWidth: 160, fontSize: "0.8125rem", color: "#6f6f6f", fontWeight: 500 }}>
                      {label}
                    </span>
                    <span style={{ fontSize: "0.875rem", color: "#161616", fontWeight: 400 }}>
                      {value}
                    </span>
                  </div>
                ))}
              </div>

              {/* Actions */}
              <div style={{ display: "flex", gap: "0.75rem" }}>
                {/* Download */}
                <button
                  onClick={() => void handleDownloadFile()}
                  disabled={status !== "COMPLETED" || isBusy}
                  style={{
                    display: "flex", flexDirection: "column", alignItems: "center",
                    gap: "0.3rem", padding: "0.6rem 1rem",
                    border: "1px solid #0f62fe", borderRadius: 4, background: "#fff",
                    cursor: status !== "COMPLETED" || isBusy ? "not-allowed" : "pointer",
                    opacity: status !== "COMPLETED" || isBusy ? 0.4 : 1,
                    transition: "background 0.1s",
                  }}
                  onMouseEnter={(e) => { if (status === "COMPLETED" && !isBusy) (e.currentTarget as HTMLButtonElement).style.background = "#edf5ff"; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.background = "#fff"; }}
                >
                  <Download size={16} style={{ color: "#0f62fe" }} />
                  <span style={{ fontSize: "0.75rem", fontWeight: 600, color: "#0f62fe" }}>Download</span>
                </button>

                {/* Reprocess */}
                <button
                  onClick={() => void handleReprocess()}
                  disabled={status !== "FAILED" || isBusy || !processable}
                  style={{
                    display: "flex", flexDirection: "column", alignItems: "center",
                    gap: "0.3rem", padding: "0.6rem 1rem",
                    border: "1px solid #393939", borderRadius: 4, background: "#fff",
                    cursor: status !== "FAILED" || isBusy || !processable ? "not-allowed" : "pointer",
                    opacity: status !== "FAILED" || isBusy || !processable ? 0.4 : 1,
                    transition: "background 0.1s",
                  }}
                  onMouseEnter={(e) => { if (status === "FAILED" && !isBusy && processable) (e.currentTarget as HTMLButtonElement).style.background = "#f4f4f4"; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.background = "#fff"; }}
                >
                  <Renew size={16} style={{ color: "#393939" }} />
                  <span style={{ fontSize: "0.75rem", fontWeight: 600, color: "#393939" }}>
                    {reprocessing ? "Reprocessing…" : "Reprocess"}
                  </span>
                </button>

                {/* Delete */}
                <button
                  onClick={() => setConfirmDelete(true)}
                  disabled={isBusy}
                  style={{
                    display: "flex", flexDirection: "column", alignItems: "center",
                    gap: "0.3rem", padding: "0.6rem 1rem",
                    border: "1px solid #da1e28", borderRadius: 4, background: "#fff",
                    cursor: isBusy ? "not-allowed" : "pointer",
                    opacity: isBusy ? 0.4 : 1,
                    transition: "background 0.1s",
                  }}
                  onMouseEnter={(e) => { if (!isBusy) (e.currentTarget as HTMLButtonElement).style.background = "#fff1f1"; }}
                  onMouseLeave={(e) => { (e.currentTarget as HTMLButtonElement).style.background = "#fff"; }}
                >
                  <TrashCan size={16} style={{ color: "#da1e28" }} />
                  <span style={{ fontSize: "0.75rem", fontWeight: 600, color: "#da1e28" }}>Delete</span>
                </button>
              </div>
            </TabPanel>

          </TabPanels>
        </Tabs>
      </div>
    </>
  );
}
