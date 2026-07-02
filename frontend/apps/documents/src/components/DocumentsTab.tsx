import { useMemo, useState } from "react";
import {
  Button,
  DataTable,
  Loading,
  Select,
  SelectItem,
  Table,
  TableBody,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  TableToolbar,
  TableToolbarContent,
  TableToolbarSearch,
} from "@carbon/react";
import { Upload } from "@carbon/react/icons";

import { EmptyState, useHeaderPanel } from "@moh-sso/ui";
import { UploadDocumentModal } from "./UploadDocumentModal";
import { useListDocumentsQuery, useGetDocumentStatsQuery } from "@moh-sso/api";
import type { DocumentResponse } from "@moh-sso/types";
import { DocumentRow } from "./document-row.component";

const TABLE_HEADERS = [
  { key: "filename", header: "Filename" },
  { key: "report_date", header: "Report Date" },
  { key: "uploaded_by", header: "Uploaded by" },
  { key: "uploaded", header: "Uploaded" },
  { key: "status", header: "Status" },
  { key: "actions", header: "" },
];

const STATUS_OPTIONS = ["ALL", "PENDING", "PROCESSING", "COMPLETED", "FAILED"] as const;
type StatusFilter = (typeof STATUS_OPTIONS)[number];

function getDocumentSearchText(doc: DocumentResponse) {
  return [doc.original_filename, doc.content_type, doc.status, doc.object_key]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
}

function getDocumentStatus(doc: Partial<DocumentResponse>) {
  return String(doc.status ?? "UNKNOWN").toUpperCase();
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value < 10 && unitIndex > 0 ? value.toFixed(1) : Math.round(value)} ${units[unitIndex]}`;
}

export function DocumentsTab() {
  const { openPanel, closePanel } = useHeaderPanel();
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("ALL");
  const [searchTerm, setSearchTerm] = useState("");

  const { data: documents = [], isLoading, isFetching } = useListDocumentsQuery();
  const { data: serverStats } = useGetDocumentStatsQuery();

  const filteredDocs = useMemo(() => {
    const q = searchTerm.trim().toLowerCase();
    return documents.filter((doc) => {
      const matchesSearch = !q || getDocumentSearchText(doc).includes(q);
      const matchesStatus =
        statusFilter === "ALL" || getDocumentStatus(doc) === statusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [documents, searchTerm, statusFilter]);

  const stats = {
    total:      serverStats?.total      ?? documents.length,
    pending:    serverStats?.pending    ?? documents.filter((d) => getDocumentStatus(d) === "PENDING").length,
    processing: serverStats?.processing ?? documents.filter((d) => getDocumentStatus(d) === "PROCESSING").length,
    completed:  serverStats?.completed  ?? documents.filter((d) => getDocumentStatus(d) === "COMPLETED").length,
    failed:     serverStats?.failed     ?? documents.filter((d) => getDocumentStatus(d) === "FAILED").length,
    totalSize:  serverStats?.total_size ?? documents.reduce((sum, d) => sum + (d.size_bytes ?? 0), 0),
  };

  function handleOpenUpload() {
    openPanel({
      title: "Upload Document",
      content: <UploadDocumentModal onClose={closePanel} />,
      size: "lg",
    });
  }

  if (isLoading) {
    return <Loading description="Loading documents..." withOverlay={false} />;
  }

  return (
    <div>
      {/* Upload button — above tiles, compact */}
      <div style={{ display: "flex", justifyContent: "flex-end", marginBottom: "1rem" }}>
        <Button renderIcon={Upload} onClick={handleOpenUpload} size="sm" kind="primary"
          style={{ width: "auto", minWidth: 0 }}>
          Upload Document
        </Button>
      </div>

      {/* Stats */}
      <div style={{
        display: "grid",
        gridTemplateColumns: "repeat(5, 1fr)",
        gap: "1px",
        background: "#e0e0e0",
        border: "1px solid #e0e0e0",
        borderRadius: 4,
        overflow: "hidden",
        marginBottom: "1.5rem",
      }}>
        {[
          { label: "Total uploads", value: stats.total,                      accent: "#0f62fe" },
          { label: "Completed",     value: stats.completed,                   accent: "#198038" },
          { label: "In progress",   value: stats.pending + stats.processing,  accent: "#f1c21b" },
          { label: "Failed",        value: stats.failed,                      accent: "#da1e28" },
        ].map(({ label, value, accent }) => (
          <div key={label} style={{ background: "#fff", padding: "1rem 1.25rem" }}>
            <div style={{ fontSize: "0.75rem", color: "#6f6f6f", marginBottom: "0.35rem", textTransform: "uppercase", letterSpacing: "0.04em" }}>
              {label}
            </div>
            <div style={{ fontSize: "1.5rem", fontWeight: 700, color: accent }}>
              {value}
            </div>
          </div>
        ))}
      </div>

      {/* Document list */}
      {documents.length === 0 ? (
        <EmptyState
          title="No documents yet"
          description="Upload a document to begin processing."
        />
      ) : (
        <DataTable rows={filteredDocs} headers={TABLE_HEADERS}>
          {({ headers, getTableProps, getHeaderProps }) => (
            <TableContainer
              title="Uploaded documents"
              description={isFetching ? "Refreshing..." : undefined}
            >
              <TableToolbar>
                <TableToolbarContent>
                  <TableToolbarSearch
                    persistent
                    value={searchTerm}
                    placeholder="Search by filename, type, status…"
                    onChange={(_, v) => setSearchTerm(v ?? "")}
                  />
                  <Select
                    id="doc-status-filter"
                    size="sm"
                    labelText=""
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
                    style={{ width: 180 }}
                  >
                    {STATUS_OPTIONS.map((s) => (
                      <SelectItem key={s} value={s} text={s === "ALL" ? "All statuses" : s} />
                    ))}
                  </Select>
                </TableToolbarContent>
              </TableToolbar>

              {filteredDocs.length === 0 ? (
                <div style={{ padding: "2rem 0" }}>
                  <EmptyState
                    title="No matching documents"
                    description="Try changing your search or status filter."
                  />
                </div>
              ) : (
                <Table {...getTableProps()} aria-label="Documents table">
                  <TableHead>
                    <TableRow>
                      {headers.map((header) => (
                        <TableHeader {...getHeaderProps({ header })} key={header.key}>
                          {header.header}
                        </TableHeader>
                      ))}
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {filteredDocs.map((doc) => (
                      <DocumentRow key={doc.id} document={doc} />
                    ))}
                  </TableBody>
                </Table>
              )}
            </TableContainer>
          )}
        </DataTable>
      )}
    </div>
  );
}
