import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
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
  Tile,
  Tag,
} from "@carbon/react";
import { Upload } from "@carbon/react/icons";

import { EmptyState, useHeaderPanel } from "@moh-sso/ui";
import { UploadDocumentModal } from "../../components/UploadDocumentModal";
import { useListDocumentsQuery } from "@moh-sso/api";
import { DocumentRow } from "../../components/document-row.component";
import type { DocumentResponse } from "@moh-sso/types";

const headers = [
  { key: "filename", header: "Filename" },
  { key: "type", header: "Type" },
  { key: "size", header: "Size" },
  { key: "status", header: "Status" },
  { key: "progress", header: "Progress" },
  { key: "actions", header: "" },
];

const STATUS_OPTIONS = ["ALL", "PENDING", "PROCESSING", "COMPLETED", "FAILED"] as const;
type StatusFilter = (typeof STATUS_OPTIONS)[number];

function getDocumentSearchText(document: DocumentResponse) {
  return [document.original_filename, document.content_type, document.status, document.object_key]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
}

function getDocumentStatus(document: Partial<DocumentResponse>) {
  return String(document.status ?? "UNKNOWN").toUpperCase();
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

export default function DocumentPage() {
  const { openPanel } = useHeaderPanel();

  const [statusFilter, setStatusFilter] = useState<StatusFilter>("ALL");
  const [searchTerm, setSearchTerm] = useState("");

  const { data: documents = [], isLoading, isFetching } = useListDocumentsQuery();

  const filteredDocs = useMemo(() => {
    const normalizedSearch = searchTerm.trim().toLowerCase();

    return documents.filter((document) => {
      const matchesSearch =
        normalizedSearch === "" || getDocumentSearchText(document).includes(normalizedSearch);

      const documentStatus = getDocumentStatus(document);
      const matchesStatus = statusFilter === "ALL" || documentStatus === statusFilter;

      return matchesSearch && matchesStatus;
    });
  }, [documents, searchTerm, statusFilter]);

  const stats = useMemo(() => {
    const total = documents.length;
    const pending = documents.filter((doc) => getDocumentStatus(doc) === "PENDING").length;
    const processing = documents.filter((doc) => getDocumentStatus(doc) === "PROCESSING").length;
    const completed = documents.filter((doc) => getDocumentStatus(doc) === "COMPLETED").length;
    const failed = documents.filter((doc) => getDocumentStatus(doc) === "FAILED").length;
    const totalSize = documents.reduce((sum, doc) => sum + (doc.size_bytes ?? 0), 0);

    return {
      total,
      pending,
      processing,
      completed,
      failed,
      totalSize,
    };
  }, [documents]);

  const handleOpenUpload = () => {
    openPanel({
      title: "Upload Document",
      content: <UploadDocumentModal onClose={() => {}} />,
      size: "lg",
    });
  };

  if (isLoading) {
    return <Loading description="Loading documents..." withOverlay={false} />;
  }

  return (
    <div>
      <div style={{ marginBottom: "1rem" }}>
        <Breadcrumb noTrailingSlash>
          <BreadcrumbItem>
            <RouterLink to="/">Home</RouterLink>
          </BreadcrumbItem>
          <BreadcrumbItem isCurrentPage>Documents</BreadcrumbItem>
        </Breadcrumb>
      </div>

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "flex-start",
          gap: "1rem",
          marginBottom: "1.5rem",
          flexWrap: "wrap",
        }}
      >
        <div>
          <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Document Management</h2>
          <p style={{ margin: 0, color: "#6f6f6f" }}>
            Upload documents, preview PDFs, and monitor import progress.
          </p>
        </div>

        <Button renderIcon={Upload} onClick={handleOpenUpload}>
          Upload Document
        </Button>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))",
          gap: "1rem",
          marginBottom: "1.5rem",
        }}
      >
        <Tile>
          <div style={{ fontSize: "0.875rem", color: "#6f6f6f", marginBottom: "0.25rem" }}>
            Total documents
          </div>
          <div style={{ fontSize: "1.5rem", fontWeight: 600 }}>{stats.total}</div>
        </Tile>

        <Tile>
          <div style={{ fontSize: "0.875rem", color: "#6f6f6f", marginBottom: "0.25rem" }}>
            Completed
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
            <div style={{ fontSize: "1.5rem", fontWeight: 600 }}>{stats.completed}</div>
            <Tag type="green">Completed</Tag>
          </div>
        </Tile>

        <Tile>
          <div style={{ fontSize: "0.875rem", color: "#6f6f6f", marginBottom: "0.25rem" }}>
            In progress
          </div>
          <div style={{ fontSize: "1.5rem", fontWeight: 600 }}>
            {stats.pending + stats.processing}
          </div>
        </Tile>

        <Tile>
          <div style={{ fontSize: "0.875rem", color: "#6f6f6f", marginBottom: "0.25rem" }}>
            Failed
          </div>
          <div style={{ fontSize: "1.5rem", fontWeight: 600 }}>{stats.failed}</div>
        </Tile>

        <Tile>
          <div style={{ fontSize: "0.875rem", color: "#6f6f6f", marginBottom: "0.25rem" }}>
            Total size
          </div>
          <div style={{ fontSize: "1.5rem", fontWeight: 600 }}>
            {formatFileSize(stats.totalSize)}
          </div>
        </Tile>
      </div>

      {documents.length === 0 ? (
        <EmptyState
          title="No documents found"
          description="Upload a document to begin processing."
        />
      ) : (
        <DataTable rows={filteredDocs} headers={headers}>
          {({ headers, getTableProps, getHeaderProps }) => (
            <TableContainer
              title="Documents"
              description={isFetching ? "Refreshing documents..." : undefined}
            >
              <TableToolbar>
                <TableToolbarContent>
                  <TableToolbarSearch
                    persistent
                    value={searchTerm}
                    placeholder="Search by filename, type, status, or key"
                    onChange={(_, value) => setSearchTerm(value ?? "")}
                  />

                  <Select
                    id="status-filter"
                    size="sm"
                    labelText=""
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
                    style={{ width: 180 }}
                  >
                    <SelectItem value="ALL" text="All Statuses" />
                    <SelectItem value="PENDING" text="Pending" />
                    <SelectItem value="PROCESSING" text="Processing" />
                    <SelectItem value="COMPLETED" text="Completed" />
                    <SelectItem value="FAILED" text="Failed" />
                  </Select>
                </TableToolbarContent>
              </TableToolbar>

              {filteredDocs.length === 0 ? (
                <div style={{ padding: "2rem 0" }}>
                  <EmptyState
                    title="No matching documents"
                    description="Try changing your search term or status filter."
                  />
                </div>
              ) : (
                <Table {...getTableProps()} aria-label="Documents table">
                  <TableHead>
                    <TableRow>
                      {headers.map((header) => (
                        <TableHeader {...getHeaderProps({ header })}>{header.header}</TableHeader>
                      ))}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {filteredDocs.map((document) => (
                      <DocumentRow key={document.id} document={document} />
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
