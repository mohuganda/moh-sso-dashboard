import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import Papa from "papaparse";
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
import { ArrowLeft, Download } from "@carbon/react/icons";

import { useToast } from "@moh-sso/ui";

import {
  useGetDocumentQuery,
  useGetDocumentDataPreviewQuery,
  useLazyDownloadDocumentQuery,
  useExportDataPreviewMutation,
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

function getRowHash(row: Record<string, unknown>): string | undefined {
  const value = row._row_hash;
  return typeof value === "string" ? value : undefined;
}

function downloadBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob);
  const link = window.document.createElement("a");
  link.href = url;
  link.download = filename;
  link.click();
  window.URL.revokeObjectURL(url);
}

type SortDir = "asc" | "desc" | null;

// ── Sheet view ────────────────────────────────────────────────────────────────

function SheetView({
  documentId,
  sheet,
  filterableKeys,
}: {
  documentId: string;
  sheet: DataPreviewSheet;
  filterableKeys: string[];
}) {
  const toast = useToast();
  const [exportDataPreview, { isLoading: isExporting }] = useExportDataPreviewMutation();

  const [searchValues, setSearchValues] = useState<string[]>([]);
  const [selections, setSelections] = useState<Record<string, string[]>>({});
  const [sortCol, setSortCol] = useState<string | null>(null);
  const [sortDir, setSortDir] = useState<SortDir>(null);
  const [page, setPage] = useState(0);
  const [selectedHashes, setSelectedHashes] = useState<Set<string>>(new Set());

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

  const searchOptions = useMemo(() => {
    const unique = new Set<string>();
    for (const row of sheet.rows) {
      for (const col of sheet.columns) {
        const value = formatCell(row[col]);
        if (value) unique.add(value);
      }
    }
    return [...unique].sort().map((v) => ({ id: v, label: v }));
  }, [sheet.rows, sheet.columns]);

  const filtered = useMemo(() => {
    let rows = sheet.rows;

    for (const key of activeFilterKeys) {
      const sel = selections[key] ?? [];
      if (sel.length === 0) continue;
      const set = new Set(sel);
      rows = rows.filter((r) => set.has(formatCell(r[key])));
    }

    if (searchValues.length > 0) {
      const set = new Set(searchValues);
      rows = rows.filter((r) => sheet.columns.some((col) => set.has(formatCell(r[col]))));
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
  }, [sheet.rows, sheet.columns, searchValues, selections, activeFilterKeys, sortCol, sortDir]);

  const totalPages = Math.ceil(filtered.length / ROWS_PER_PAGE);
  const pageRows = filtered.slice(page * ROWS_PER_PAGE, (page + 1) * ROWS_PER_PAGE);

  const pageHashes = useMemo(
    () => pageRows.map(getRowHash).filter((h): h is string => !!h),
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

  function downloadSelected() {
    const selectedRows = sheet.rows.filter((r) => {
      const hash = getRowHash(r);
      return hash && selectedHashes.has(hash);
    });
    if (selectedRows.length === 0) return;

    const csv = Papa.unparse({
      fields: sheet.columns,
      data: selectedRows.map((row) => sheet.columns.map((col) => formatCell(row[col]))),
    });
    downloadBlob(
      new Blob([csv], { type: "text/csv;charset=utf-8;" }),
      `${sheet.code || sheet.name}-selected.csv`,
    );
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
      downloadBlob(blob, `${sheet.code || sheet.name}-export.csv`);
      toast.success("Export started", "All matching rows from the server");
    } catch {
      toast.error("Export failed", "Please try again.");
    }
  }

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
    setSearchValues([]);
    setSelections({});
    setSortCol(null);
    setSortDir(null);
    setPage(0);
  }

  const hasActiveFilters = !!(
    searchValues.length > 0 ||
    Object.values(selections).some((s) => s.length > 0) ||
    sortCol
  );

  return (
    <div>
      {/* ── Toolbar ── */}
      <div className="document-preview__toolbar">
        <div className="document-preview__search">
          <FilterableMultiSelect
            id={`search-${sheet.name}`}
            titleText="Search values"
            placeholder="Type to search…"
            items={searchOptions}
            itemToString={(item) => item?.label ?? ""}
            selectedItems={searchValues.map((v) => ({ id: v, label: v }))}
            onChange={({ selectedItems }) => {
              setSearchValues((selectedItems ?? []).map((i) => i.id));
              setPage(0);
            }}
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

      {/* ── Selection actions ── */}
      <div className="document-preview__selection-bar">
        <span className="document-preview__selection-count">
          {selectedHashes.size > 0
            ? `${selectedHashes.size.toLocaleString()} row${selectedHashes.size === 1 ? "" : "s"} selected`
            : "Select rows to download only those"}
        </span>

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
      </div>

      {/* ── Table ── */}
      <div className="document-preview__table-scroll">
        <table className="document-preview__table">
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
                <td colSpan={sheet.columns.length + 1} className="document-preview__no-rows">
                  No rows match the current filters.
                </td>
              </tr>
            ) : (
              pageRows.map((row, idx) => {
                const hash = getRowHash(row);
                const isSelected = !!hash && selectedHashes.has(hash);
                return (
                  <tr
                    key={hash ?? idx}
                    className={isSelected ? "document-preview__row--selected" : undefined}
                  >
                    <td className="document-preview__checkbox-cell">
                      <Checkbox
                        id={`select-row-${sheet.name}-${hash ?? idx}`}
                        labelText="Select row"
                        hideLabel
                        checked={isSelected}
                        onChange={() => toggleRow(hash)}
                        disabled={!hash}
                      />
                    </td>
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
                );
              })
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

  // While the document is still PENDING/PROCESSING, its background import
  // job hasn't necessarily written rows yet — poll both queries so the page
  // picks up the real data (and status) once processing finishes, instead of
  // holding on to a "no data yet" response fetched too early.
  const [isProcessing, setIsProcessing] = useState(true);

  const { data: document, isLoading: isLoadingDoc } = useGetDocumentQuery(id!, {
    skip: !id,
    pollingInterval: isProcessing ? 3000 : 0,
  });
  const {
    data: preview,
    isLoading: isLoadingPreview,
    isError,
  } = useGetDocumentDataPreviewQuery(id!, {
    skip: !id,
    pollingInterval: isProcessing ? 3000 : 0,
  });

  useEffect(() => {
    const status = document?.status?.toUpperCase();
    setIsProcessing(status === "PENDING" || status === "PROCESSING");
  }, [document?.status]);

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
      downloadBlob(blob, document.original_filename);
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
                <SheetView
                  documentId={id!}
                  sheet={sheet}
                  filterableKeys={filterableKeysBySheet[sheet.name] ?? []}
                />
              </TabPanel>
            ))}
          </TabPanels>
        </Tabs>
      )}
    </div>
  );
}
