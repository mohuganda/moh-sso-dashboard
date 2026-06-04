import { useMemo, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
  Button,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  TableToolbarSearch,
  Tabs,
  Tag,
} from "@carbon/react";
import { ArrowLeft, Download } from "@carbon/icons-react";

import {
  useGetDocumentQuery,
  useGetDocumentDataPreviewQuery,
  useLazyDownloadDocumentQuery,
} from "../../../../store/api/document.api";
import type { DataPreviewSheet } from "../../../../store/types/documents.types";

// ── Helpers ───────────────────────────────────────────────────────────────────

const ROWS_PER_PAGE = 100;

function formatCell(v: unknown): string {
  if (v === null || v === undefined) return "";
  if (typeof v === "string") return v;
  return String(v);
}

function formatDateTime(iso?: string) {
  if (!iso) return "—";
  const d = new Date(iso);
  const date = d.toLocaleDateString("en-GB", { day: "numeric", month: "long", year: "numeric" });
  const time = d.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
  return `${date} at ${time}`;
}

type SortDir = "asc" | "desc" | null;

// ── Sheet view ────────────────────────────────────────────────────────────────

function SheetView({ sheet }: { sheet: DataPreviewSheet }) {
  const [search, setSearch] = useState("");
  const [filters, setFilters] = useState<Record<string, string>>({});
  const [sortCol, setSortCol] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDir>(null);
  const [page, setPage] = useState(0);

  // Filterable columns: date columns + a few categorical ones
  const filterableCols = useMemo(
    () => sheet.columns.filter((c) =>
      ["location_code", "location_name", "district", "item_status", "uom",
       "unit_of_measure_code", "ownership", "customer_name", "report_date",
       "funding_source", "item_inventory_status"].includes(c)
    ),
    [sheet.columns],
  );

  const uniqueValues = useMemo(() => {
    const map: Record<string, Set<string>> = {};
    for (const col of filterableCols) {
      map[col] = new Set(
        sheet.rows.map((r) => formatCell(r[col])).filter(Boolean),
      );
    }
    return map;
  }, [sheet.rows, filterableCols]);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();

    let rows = sheet.rows;

    // Column filters
    for (const [col, val] of Object.entries(filters)) {
      if (!val) continue;
      rows = rows.filter((r) => formatCell(r[col]) === val);
    }

    // Global search
    if (q) {
      rows = rows.filter((r) =>
        sheet.columns.some((col) => formatCell(r[col]).toLowerCase().includes(q)),
      );
    }

    // Sort
    if (sortCol && sortDir) {
      rows = [...rows].sort((a, b) => {
        const av = formatCell(a[sortCol]);
        const bv = formatCell(b[sortCol]);
        const num = !isNaN(Number(av)) && !isNaN(Number(bv));
        const cmp = num
          ? Number(av) - Number(bv)
          : av.localeCompare(bv, undefined, { sensitivity: "base" });
        return sortDir === "asc" ? cmp : -cmp;
      });
    }

    return rows;
  }, [sheet.rows, search, filters, sortCol, sortDir]);

  const totalPages = Math.ceil(filtered.length / ROWS_PER_PAGE);
  const pageRows = filtered.slice(page * ROWS_PER_PAGE, (page + 1) * ROWS_PER_PAGE);

  function handleSort(col: string) {
    if (sortCol !== col) { setSortCol(col); setSortDir("asc"); }
    else if (sortDir === "asc") setSortDir("desc");
    else { setSortCol(null); setSortDir(null); }
    setPage(0);
  }

  function handleFilter(col: string, val: string) {
    setFilters((prev) => ({ ...prev, [col]: val }));
    setPage(0);
  }

  function clearFilters() {
    setFilters({});
    setSearch("");
    setSortCol(null);
    setSortDir(null);
    setPage(0);
  }

  const hasActiveFilters = search || Object.values(filters).some(Boolean) || sortCol;

  return (
    <div>
      {/* ── Toolbar ── */}
      <div style={{
        display: "flex", gap: "0.75rem", flexWrap: "wrap",
        alignItems: "flex-end", marginBottom: "0.75rem",
        padding: "0.75rem", background: "#f4f4f4",
      }}>
        <div style={{ flex: "1 1 220px", minWidth: 200 }}>
          <TableToolbarSearch
            persistent
            value={search}
            placeholder="Search all columns…"
            onChange={(_, v) => { setSearch(v ?? ""); setPage(0); }}
            labelText="Search"
          />
        </div>

        {filterableCols.map((col) => (
          <div key={col} style={{ flex: "0 0 auto" }}>
            <Select
              id={`filter-${col}`}
              labelText={col.replace(/_/g, " ")}
              size="sm"
              value={filters[col] ?? ""}
              onChange={(e) => handleFilter(col, e.target.value)}
              style={{ minWidth: 160 }}
            >
              <SelectItem value="" text="All" />
              {[...uniqueValues[col]].sort().map((v) => (
                <SelectItem key={v} value={v} text={v} />
              ))}
            </Select>
          </div>
        ))}

        {hasActiveFilters && (
          <Button kind="ghost" size="sm" onClick={clearFilters}>
            Clear filters
          </Button>
        )}

        <div style={{ marginLeft: "auto", fontSize: "0.875rem", color: "#525252", alignSelf: "center" }}>
          {filtered.length.toLocaleString()} / {sheet.row_count.toLocaleString()} rows
          {sheet.rows.length < sheet.row_count && (
            <span style={{ color: "#e78c21", marginLeft: "0.5rem" }}>
              (preview: first {sheet.rows.length} loaded)
            </span>
          )}
        </div>
      </div>

      {/* ── Table ── */}
      <div style={{ overflowX: "auto" }}>
        <table style={{ borderCollapse: "collapse", fontSize: "0.8125rem", width: "100%", minWidth: 600 }}>
          <thead>
            <tr style={{ background: "#e8e8e8", borderBottom: "2px solid #c6c6c6", position: "sticky", top: 0 }}>
              {sheet.columns.map((col) => {
                const isActive = sortCol === col;
                const arrow = isActive ? (sortDir === "asc" ? " ▲" : " ▼") : "";
                return (
                  <th
                    key={col}
                    onClick={() => handleSort(col)}
                    style={{
                      textAlign: "left", padding: "0.5rem 0.75rem",
                      fontWeight: 600, color: "#161616", whiteSpace: "nowrap",
                      cursor: "pointer", userSelect: "none",
                      background: isActive ? "#d0e2ff" : undefined,
                      borderBottom: isActive ? "2px solid #0f62fe" : undefined,
                    }}
                    title={`Sort by ${col}`}
                  >
                    {col.replace(/_/g, " ")}{arrow}
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
                  style={{ padding: "2rem", textAlign: "center", color: "#6f6f6f" }}
                >
                  No rows match the current filters.
                </td>
              </tr>
            ) : (
              pageRows.map((row, idx) => (
                <tr
                  key={idx}
                  style={{ borderBottom: "1px solid #e8e8e8", background: idx % 2 ? "#fafafa" : "#fff" }}
                >
                  {sheet.columns.map((col) => {
                    const val = formatCell(row[col]);
                    return (
                      <td
                        key={col}
                        style={{
                          padding: "0.4rem 0.75rem", whiteSpace: "nowrap",
                          maxWidth: 260, overflow: "hidden", textOverflow: "ellipsis",
                          color: val ? "#161616" : "#a8a8a8",
                        }}
                        title={val || "—"}
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

      {/* ── Pagination ── */}
      {totalPages > 1 && (
        <div style={{
          display: "flex", gap: "0.5rem", alignItems: "center",
          padding: "0.75rem 0", justifyContent: "center", flexWrap: "wrap",
        }}>
          <Button kind="ghost" size="sm" disabled={page === 0} onClick={() => setPage(0)}>«</Button>
          <Button kind="ghost" size="sm" disabled={page === 0} onClick={() => setPage((p) => p - 1)}>‹ Prev</Button>
          <span style={{ fontSize: "0.875rem", color: "#525252" }}>
            Page {page + 1} of {totalPages}
          </span>
          <Button kind="ghost" size="sm" disabled={page >= totalPages - 1} onClick={() => setPage((p) => p + 1)}>Next ›</Button>
          <Button kind="ghost" size="sm" disabled={page >= totalPages - 1} onClick={() => setPage(totalPages - 1)}>»</Button>
        </div>
      )}
    </div>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function DocumentPreviewPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [triggerDownload] = useLazyDownloadDocumentQuery();

  const { data: document, isLoading: isLoadingDoc } = useGetDocumentQuery(id!, { skip: !id });
  const { data: preview, isLoading: isLoadingPreview, isError } = useGetDocumentDataPreviewQuery(
    id!, { skip: !id },
  );

  const handleDownload = async () => {
    if (!id || !document) return;
    try {
      const blob = await triggerDownload(id).unwrap();
      const url = window.URL.createObjectURL(blob);
      const link = window.document.createElement("a");
      link.href = url;
      link.download = document.original_filename;
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (e) { console.error("Download failed", e); }
  };

  const isLoading = isLoadingDoc || isLoadingPreview;

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: "1rem" }}>
      {/* Breadcrumb */}
      <Breadcrumb noTrailingSlash>
        <BreadcrumbItem><RouterLink to="/">Home</RouterLink></BreadcrumbItem>
        <BreadcrumbItem>
          <RouterLink to="/apps/utilities/self-service/eservice/document-upload">Documents</RouterLink>
        </BreadcrumbItem>
        <BreadcrumbItem isCurrentPage>Data Preview</BreadcrumbItem>
      </Breadcrumb>

      {/* Header */}
      <div style={{ display: "flex", alignItems: "flex-start", gap: "1rem", flexWrap: "wrap" }}>
        <div style={{ flex: 1 }}>
          <div style={{ display: "flex", alignItems: "center", gap: "0.75rem", flexWrap: "wrap" }}>
            <Button
              kind="ghost"
              size="sm"
              renderIcon={ArrowLeft}
              onClick={() => navigate(-1)}
              style={{ paddingLeft: 0 }}
            >
              Back
            </Button>
            <h2 style={{ margin: 0, fontSize: "1.25rem", fontWeight: 600 }}>
              {document?.original_filename ?? "Data Preview"}
            </h2>
          </div>
          {document && (
            <div style={{ display: "flex", gap: "0.5rem", marginTop: "0.5rem", flexWrap: "wrap", alignItems: "center" }}>
              {preview?.template_code && (
                <Tag type="cyan" style={{ fontFamily: "monospace" }}>{preview.template_code}</Tag>
              )}
              {preview?.report_date && (
                <Tag type="teal">Report: {preview.report_date}</Tag>
              )}
              <span style={{ fontSize: "0.8rem", color: "#6f6f6f" }}>
                Uploaded {formatDateTime(document.created_at)}
              </span>
            </div>
          )}
        </div>
        <Button
          renderIcon={Download}
          kind="primary"
          size="sm"
          disabled={!document || document.status !== "COMPLETED"}
          onClick={() => void handleDownload()}
        >
          Download file
        </Button>
      </div>

      {/* Content */}
      {isLoading ? (
        <InlineLoading description="Loading data…" style={{ padding: "2rem 0" }} />
      ) : isError ? (
        <InlineNotification
          kind="error"
          title="Failed to load preview"
          subtitle="Could not retrieve imported data. The file may still be processing."
          lowContrast
        />
      ) : !preview || preview.sheets.length === 0 ? (
        <InlineNotification
          kind="info"
          title="No data available"
          subtitle="No imported rows found for this document. It may still be processing or may not use a recognised template."
          lowContrast
        />
      ) : (
        <Tabs>
          <TabList aria-label="Sheet tabs" contained>
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
                <SheetView sheet={sheet} />
              </TabPanel>
            ))}
          </TabPanels>
        </Tabs>
      )}
    </div>
  );
}
