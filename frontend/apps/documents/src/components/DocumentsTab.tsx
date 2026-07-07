import { useMemo, useState } from "react";
import {
  Button,
  DataTable,
  InlineNotification,
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

import { PermissionGuard, PERMISSIONS } from "@moh-sso/auth";
import { EmptyState, useHeaderPanel } from "@moh-sso/ui";

import { UploadDocumentModal } from "./UploadDocumentModal";
import { DocumentRow } from "./document-row.component";
import { useGetDocumentStatsQuery, useListDocumentsQuery } from "../api";
import type { DocumentResponse } from "../types";

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

function getDocumentSearchText(document: DocumentResponse) {
  return [document.original_filename, document.content_type, document.status, document.object_key]
    .filter(Boolean)
    .join(" ")
    .toLowerCase();
}

function getDocumentStatus(document: Partial<DocumentResponse>) {
  return String(document.status ?? "UNKNOWN").toUpperCase();
}

function DocumentsTabContent() {
  const { openPanel, closePanel } = useHeaderPanel();

  const [statusFilter, setStatusFilter] = useState<StatusFilter>("ALL");
  const [searchTerm, setSearchTerm] = useState("");

  const { data: documents = [], isLoading, isFetching } = useListDocumentsQuery();

  const { data: serverStats } = useGetDocumentStatsQuery();

  const filteredDocs = useMemo(() => {
    const query = searchTerm.trim().toLowerCase();

    return documents.filter((document) => {
      const matchesSearch = !query || getDocumentSearchText(document).includes(query);

      const matchesStatus = statusFilter === "ALL" || getDocumentStatus(document) === statusFilter;

      return matchesSearch && matchesStatus;
    });
  }, [documents, searchTerm, statusFilter]);

  const stats = {
    total: serverStats?.total ?? documents.length,

    pending:
      serverStats?.pending ??
      documents.filter((document) => getDocumentStatus(document) === "PENDING").length,

    processing:
      serverStats?.processing ??
      documents.filter((document) => getDocumentStatus(document) === "PROCESSING").length,

    completed:
      serverStats?.completed ??
      documents.filter((document) => getDocumentStatus(document) === "COMPLETED").length,

    failed:
      serverStats?.failed ??
      documents.filter((document) => getDocumentStatus(document) === "FAILED").length,

    totalSize:
      serverStats?.total_size ??
      documents.reduce((sum, document) => sum + (document.size_bytes ?? 0), 0),
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
    <div className="documents-tab">
      <div
        className="documents-tab__actions"
        style={{
          display: "flex",
          justifyContent: "flex-end",
          marginBottom: "1rem",
        }}
      >
        <PermissionGuard permission={PERMISSIONS.documentsWrite}>
          <Button
            renderIcon={Upload}
            onClick={handleOpenUpload}
            size="sm"
            kind="primary"
            style={{
              width: "auto",
              minWidth: 0,
            }}
          >
            Upload Document
          </Button>
        </PermissionGuard>
      </div>

      <div
        className="documents-tab__stats"
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(160px, 1fr))",
          gap: "1px",
          background: "#e0e0e0",
          border: "1px solid #e0e0e0",
          borderRadius: 4,
          overflow: "hidden",
          marginBottom: "1.5rem",
        }}
      >
        {[
          {
            label: "Total uploads",
            value: stats.total,
            accent: "#0f62fe",
          },
          {
            label: "Completed",
            value: stats.completed,
            accent: "#198038",
          },
          {
            label: "In progress",
            value: stats.pending + stats.processing,
            accent: "#f1c21b",
          },
          {
            label: "Failed",
            value: stats.failed,
            accent: "#da1e28",
          },
        ].map(({ label, value, accent }) => (
          <div
            key={label}
            style={{
              background: "#fff",
              padding: "1rem 1.25rem",
            }}
          >
            <div
              style={{
                fontSize: "0.75rem",
                color: "#6f6f6f",
                marginBottom: "0.35rem",
                textTransform: "uppercase",
                letterSpacing: "0.04em",
              }}
            >
              {label}
            </div>

            <div
              style={{
                fontSize: "1.5rem",
                fontWeight: 700,
                color: accent,
              }}
            >
              {value}
            </div>
          </div>
        ))}
      </div>

      {documents.length === 0 ? (
        <EmptyState
          title="No documents yet"
          description={"No uploaded documents are currently available."}
        />
      ) : (
        <DataTable rows={filteredDocs} headers={TABLE_HEADERS}>
          {({ headers, getTableProps, getHeaderProps }) => (
            <TableContainer
              className="documents-tab__table"
              title="Uploaded documents"
              description={isFetching ? "Refreshing..." : undefined}
            >
              <TableToolbar>
                <TableToolbarContent>
                  <TableToolbarSearch
                    persistent
                    value={searchTerm}
                    placeholder="Search by filename, type, status…"
                    onChange={(_, value) => setSearchTerm(value ?? "")}
                  />

                  <Select
                    id="doc-status-filter"
                    size="sm"
                    labelText="Filter by document status"
                    hideLabel
                    value={statusFilter}
                    onChange={(event) => setStatusFilter(event.target.value as StatusFilter)}
                    style={{ width: 180 }}
                  >
                    {STATUS_OPTIONS.map((status) => (
                      <SelectItem
                        key={status}
                        value={status}
                        text={status === "ALL" ? "All statuses" : status}
                      />
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

export function DocumentsTab() {
  return (
    <PermissionGuard
      permission={PERMISSIONS.documentsRead}
      fallback={
        <InlineNotification
          kind="warning"
          title="Access denied"
          subtitle="You do not have permission to view uploaded documents."
          lowContrast
        />
      }
    >
      <DocumentsTabContent />
    </PermissionGuard>
  );
}
