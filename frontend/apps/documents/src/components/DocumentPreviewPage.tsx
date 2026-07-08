import { useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
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
import { ArrowLeft, Download } from "@carbon/react/icons";

import { useToast } from "@moh-sso/ui";

import {
  useGetDocumentQuery,
  useGetDocumentDataPreviewQuery,
  useLazyDownloadDocumentQuery,
  useGetTemplateStructureQuery,
} from "../api";
import type { DataPreviewSheet } from "../types";
import "./documents-components.scss";

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

function SheetView({
  sheet,
  filterableKeys,
}: {
  sheet: DataPreviewSheet;
  filterableKeys: string[];
}) {
  const [search, setSearch] = useState("");
  const [selections, setSelections] = useState<Record<string, string[]>>({});
  const [sortCol, setSortCol] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDir>(null);
  const [page, setPage] = useState(0);

  const activeFilterKeys = useMemo(
    () => filterableKeys.filter((k) => sheet.columns.includes(k)),
    [filterableKeys, sheet.columns],
  );

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
  }, [sheet.rows, sheet.columns, search, selections, activeFilterKeys, sortCol, sortDir]);

  const totalPages = Math.ceil(filtered.length / ROWS_PER_PAGE);
  const pageRows = filtered.slice(page * ROWS_PER_PAGE, (page + 1) * ROWS_PER_PAGE);

  function handleSort(col: string) {
    if (sortCol !== col) {
      setSortCol(col);
      setSortDir("asc");
    } else if (sortDir === "asc") setSortDir("desc");
    else {
      setSortCol(null);
      setSortDir(null);
    }
    setPage(0);
  }

  function clearFilters() {
    setSearch("");
    setSelections({});
    setSortCol(null);
    setSortDir(null);
    setPage(0);
  }

  const hasActiveFilters = !!(
    search ||
    Object.values(selections).some((s) => s.length > 0) ||
    sortCol
  );

  return (
    <div>
      {/* ── Toolbar ── */}
      <div className="document-preview__toolbar">
        <div className="document-preview__search">
          <TableToolbarSearch
            persistent
            value={search}
            placeholder="Search all columns…"
            onChange={(_, v) => {
              setSearch(v ?? "");
              setPage(0);
            }}
            labelText="Search"
          />
        </div>

        {activeFilterKeys.map((key) => {
          const opts = filterOptions[key] ?? [];
          if (opts.length === 0) return null;
          const label = key.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
          const selected = (selections[key] ?? []).map((v) => ({ id: v, label: v }));
          return (
            <div key={key} className="document-preview__filter">
              <FilterableMultiSelect
                id={`filter-${sheet.name}-${key}`}
                titleText={label}
                placeholder="Type to search…"
                items={opts}
                itemToString={(item) => item?.label ?? ""}
                selectedItems={selected}
                onChange={({ selectedItems }) => {
                  setSelections((prev) => ({
                    ...prev,
                    [key]: (selectedItems ?? []).map((i) => i.id),
                  }));
                  setPage(0);
                }}
                size="sm"
              />
            </div>
          );
        })}

        {hasActiveFilters && (
          <Button kind="ghost" size="sm" onClick={clearFilters}>
            Clear filters
          </Button>
        )}

        <div className="document-preview__row-count">
          {filtered.length.toLocaleString()} / {sheet.row_count.toLocaleString()} rows
          {sheet.rows.length < sheet.row_count && (
            <span className="document-preview__row-count-note">
              (preview: first {sheet.rows.length} loaded)
            </span>
          )}
        </div>
      </div>

      {/* ── Table ── */}
      <div className="document-preview__table-scroll">
        <table className="document-preview__table">
          <thead>
            <tr>
              {sheet.columns.map((col) => {
                const isActive = sortCol === col;
                const arrow = isActive ? (sortDir === "asc" ? " ▲" : " ▼") : "";
                return (
                  <th
                    key={col}
                    onClick={() => handleSort(col)}
                    className={isActive ? "document-preview__sort-header--active" : undefined}
                    title={`Sort by ${col}`}
                  >
                    {col.replace(/_/g, " ")}
                    {arrow}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {pageRows.length === 0 ? (
              <tr>
                <td colSpan={sheet.columns.length} className="document-preview__no-rows">
                  No rows match the current filters.
                </td>
              </tr>
            ) : (
              pageRows.map((row, idx) => (
                <tr key={idx}>
                  {sheet.columns.map((col) => {
                    const val = formatCell(row[col]);
                    return (
                      <td
                        key={col}
                        className={val ? undefined : "document-preview__empty-cell"}
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
        <div className="document-preview__pagination">
          <Button kind="ghost" size="sm" disabled={page === 0} onClick={() => setPage(0)}>
            «
          </Button>
          <Button
            kind="ghost"
            size="sm"
            disabled={page === 0}
            onClick={() => setPage((p) => p - 1)}
          >
            ‹ Prev
          </Button>
          <span className="document-preview__page-label">
            Page {page + 1} of {totalPages}
          </span>
          <Button
            kind="ghost"
            size="sm"
            disabled={page >= totalPages - 1}
            onClick={() => setPage((p) => p + 1)}
          >
            Next ›
          </Button>
          <Button
            kind="ghost"
            size="sm"
            disabled={page >= totalPages - 1}
            onClick={() => setPage(totalPages - 1)}
          >
            »
          </Button>
        </div>
      )}
    </div>
  );
}

// ── Page ──────────────────────────────────────────────────────────────────────

export default function DocumentPreviewPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const toast = useToast();
  const [triggerDownload] = useLazyDownloadDocumentQuery();

  const { data: document, isLoading: isLoadingDoc } = useGetDocumentQuery(id!, { skip: !id });
  const {
    data: preview,
    isLoading: isLoadingPreview,
    isError,
  } = useGetDocumentDataPreviewQuery(id!, { skip: !id });

  const templateCode = (document?.metadata?.template_code as string | undefined) ?? "";
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
      toast.success("Download started", document.original_filename);
    } catch {
      toast.error("Download failed", "Please try again.");
    }
  };

  const isLoading = isLoadingDoc || isLoadingPreview;

  return (
    <div className="document-preview">
      {/* Header */}
      <div className="document-preview__header">
        <div className="document-preview__header-content">
          <div className="document-preview__title-row">
            <Button
              kind="ghost"
              size="sm"
              renderIcon={ArrowLeft}
              onClick={() => navigate(-1)}
              className="document-preview__back"
            >
              Back
            </Button>
            <h2 className="document-preview__title">
              {document?.original_filename ?? "Data Preview"}
            </h2>
          </div>
          {document && (
            <div className="document-preview__metadata">
              {preview?.template_code && (
                <Tag type="cyan" className="document-detail-panel__mono-tag">
                  {preview.template_code}
                </Tag>
              )}
              {preview?.report_date && <Tag type="teal">Report: {preview.report_date}</Tag>}
              <span className="document-preview__uploaded">
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
        <InlineLoading description="Loading data…" className="document-preview__loading" />
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
                <Tag type="gray" className="document-preview__tab-tag">
                  {sheet.row_count.toLocaleString()}
                </Tag>
              </Tab>
            ))}
          </TabList>
          <TabPanels>
            {preview.sheets.map((sheet) => (
              <TabPanel key={sheet.name} className="document-preview__tab-panel">
                <SheetView sheet={sheet} filterableKeys={filterableKeysBySheet[sheet.name] ?? []} />
              </TabPanel>
            ))}
          </TabPanels>
        </Tabs>
      )}
    </div>
  );
}
