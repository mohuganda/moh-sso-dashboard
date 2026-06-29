import { useMemo } from "react";
import { useNavigate, useParams } from "react-router-dom";
import {
  Breadcrumb,
  BreadcrumbItem,
  Button,
  DataTable,
  ProgressBar,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableHeader,
  TableRow,
  Tag,
  Tile,
} from "@carbon/react";

import {
  useDeleteDocumentMutation,
  useGetDocumentProcessesQuery,
  useGetDocumentQuery,
  useLazyDownloadDocumentQuery,
  useLazyViewDocumentQuery,
  useReprocessDocumentMutation,
} from "../../../api";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import type { DocumentProcess, DocumentResponse } from "../../../types";
import { useToast } from "@moh-sso/ui";

function StatusTag({ status }: { status: string }) {
  const colorMap: Record<string, "gray" | "blue" | "green" | "red" | "magenta"> = {
    PENDING: "gray",
    PROCESSING: "blue",
    COMPLETED: "green",
    FAILED: "red",
    CANCELLED: "magenta",
  };

  return <Tag type={colorMap[status] || "gray"}>{status}</Tag>;
}

function requiresProcessing(document?: Partial<DocumentResponse>) {
  if (!document) {
    return false;
  }

  const type = (document.content_type || "").toLowerCase();

  const filename = (document.original_filename || "").toLowerCase();

  if (
    type === "text/csv" ||
    type === "application/vnd.ms-excel" ||
    type === "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  ) {
    return true;
  }

  return filename.endsWith(".csv") || filename.endsWith(".xls") || filename.endsWith(".xlsx");
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB", "TB"];
  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  return `${
    value < 10 && unitIndex > 0 ? value.toFixed(1) : Math.round(value)
  } ${units[unitIndex]}`;
}

function formatDate(value?: string | null) {
  if (!value) {
    return "—";
  }

  const date = new Date(value);

  return Number.isNaN(date.getTime()) ? "—" : date.toLocaleString();
}

function getLatestProcess(processes?: DocumentProcess[]) {
  if (!processes || processes.length === 0) {
    return undefined;
  }

  return [...processes].sort(
    (first, second) =>
      new Date(second.created_at ?? 0).getTime() - new Date(first.created_at ?? 0).getTime(),
  )[0];
}

function getEffectiveStatus(document?: DocumentResponse, latestProcess?: DocumentProcess) {
  return String(latestProcess?.status || document?.status || "UNKNOWN").toUpperCase();
}

function getEffectiveProgress(document?: DocumentResponse, latestProcess?: DocumentProcess) {
  if (typeof latestProcess?.progress === "number") {
    return latestProcess.progress;
  }

  const status = getEffectiveStatus(document, latestProcess);

  return status === "COMPLETED" ? 100 : 0;
}

export default function DocumentDetailsPage() {
  const toast = useToast();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const {
    data: document,
    isLoading: docLoading,
    error: docError,
  } = useGetDocumentQuery(id!, {
    skip: !id,
  });

  const processable = requiresProcessing(document);

  const { data: processes = [], isFetching: processesRefreshing } = useGetDocumentProcessesQuery(
    id!,
    {
      skip: !id || !processable,
      pollingInterval: processable ? 3000 : 0,
      refetchOnFocus: true,
    },
  );

  const [deleteDocument, { isLoading: deleting }] = useDeleteDocumentMutation();

  const [triggerDownload] = useLazyDownloadDocumentQuery();

  const [triggerView] = useLazyViewDocumentQuery();

  const [reprocessDocument, { isLoading: reprocessing }] = useReprocessDocumentMutation();

  const latestProcess = useMemo(() => getLatestProcess(processes), [processes]);

  const effectiveStatus = getEffectiveStatus(document, latestProcess);

  const effectiveProgress = getEffectiveProgress(document, latestProcess);

  const handleDelete = async () => {
    if (!id) {
      return;
    }

    const confirmed = window.confirm(
      `Delete "${document?.original_filename ?? "this document"}"? This action cannot be undone.`,
    );

    if (!confirmed) {
      return;
    }

    try {
      await deleteDocument(id).unwrap();

      toast.success("Document deleted", "The document was deleted successfully");

      navigate(-1);
    } catch (error: any) {
      toast.error("Delete failed", error?.data?.error || "Failed to delete document");
    }
  };

  const handleDownload = async () => {
    if (!id || !document) {
      return;
    }

    try {
      const blob = await triggerDownload(id).unwrap();

      const url = window.URL.createObjectURL(blob);

      const link = window.document.createElement("a");

      link.href = url;
      link.download = document.original_filename;
      link.click();

      window.URL.revokeObjectURL(url);
    } catch (error: any) {
      toast.error("Download failed", error?.data?.error || "Failed to download document");
    }
  };

  const handlePreview = async () => {
    if (!id) {
      return;
    }

    try {
      const blob = await triggerView(id).unwrap();
      const url = window.URL.createObjectURL(blob);

      const previewWindow = window.open(url, "_blank", "noopener,noreferrer");

      if (!previewWindow) {
        window.URL.revokeObjectURL(url);

        toast.error("Preview blocked", "Allow pop-ups to preview this document.");

        return;
      }

      window.setTimeout(() => {
        window.URL.revokeObjectURL(url);
      }, 60_000);
    } catch (error: any) {
      toast.error("Preview failed", error?.data?.error || "Failed to preview document");
    }
  };

  const handleReprocess = async () => {
    if (!id) {
      return;
    }

    try {
      await reprocessDocument(id).unwrap();

      toast.success("Reprocess started", "Reprocessing started successfully");
    } catch (error: any) {
      toast.error("Reprocess failed", error?.data?.error || "Failed to reprocess document");
    }
  };

  if (docLoading) {
    return <div>Loading...</div>;
  }

  if (docError || !document) {
    return <div>Failed to load document</div>;
  }

  const historyRows = processes.map((process, index) => ({
    id: process.id,
    attempt: index + 1,
    status: process.status,
    progress: process.progress,
    started: formatDate(process.started_at || process.created_at),
    finished: formatDate(process.finished_at),
    message: process.message || "—",
  }));

  const headers = [
    {
      key: "attempt",
      header: "Attempt",
    },
    {
      key: "status",
      header: "Status",
    },
    {
      key: "progress",
      header: "Progress",
    },
    {
      key: "started",
      header: "Started",
    },
    {
      key: "finished",
      header: "Finished",
    },
    {
      key: "message",
      header: "Message",
    },
  ];

  const canDownload = effectiveStatus === "COMPLETED";

  const canPreview = effectiveStatus === "COMPLETED";

  const canReprocess = processable && effectiveStatus !== "PROCESSING";

  return (
    <PermissionGuard permission={PERMISSIONS.documentsRead}>
      <div>
        <Breadcrumb style={{ marginBottom: "1rem" }}>
          <BreadcrumbItem onClick={() => navigate(-1)}>Documents</BreadcrumbItem>

          <BreadcrumbItem isCurrentPage>{document.original_filename}</BreadcrumbItem>
        </Breadcrumb>

        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "flex-start",
            gap: "1rem",
            marginBottom: "2rem",
            flexWrap: "wrap",
          }}
        >
          <div>
            <h2
              style={{
                margin: 0,
                marginBottom: "0.5rem",
              }}
            >
              {document.original_filename}
            </h2>

            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.75rem",
                flexWrap: "wrap",
              }}
            >
              <StatusTag status={effectiveStatus} />

              <span
                style={{
                  fontSize: "0.875rem",
                  color: "#6f6f6f",
                }}
              >
                {processable ? `Attempts: ${processes.length}` : "No processing required"}
              </span>

              {processesRefreshing && (
                <span
                  style={{
                    fontSize: "0.875rem",
                    color: "#6f6f6f",
                  }}
                >
                  Refreshing…
                </span>
              )}
            </div>
          </div>
        </div>

        <Tile style={{ marginBottom: "2rem" }}>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))",
              gap: "1rem",
            }}
          >
            <InfoItem label="Status" value={effectiveStatus} />

            <InfoItem label="Content Type" value={document.content_type || "—"} />

            <InfoItem label="Size" value={formatFileSize(document.size_bytes)} />

            <InfoItem label="Object Key" value={document.object_key || "—"} />

            <InfoItem label="Created At" value={formatDate(document.created_at)} />

            <InfoItem label="Updated At" value={formatDate(document.updated_at)} />
          </div>
        </Tile>

        <Tile style={{ marginBottom: "2rem" }}>
          <h4 style={{ marginBottom: "1rem" }}>
            {processable ? "Latest Processing Attempt" : "Document Status"}
          </h4>

          <Stack gap={4}>
            <StatusTag status={effectiveStatus} />

            <div style={{ width: 320 }}>
              <ProgressBar value={effectiveProgress} max={100} label="" size="small" />
            </div>

            {latestProcess ? (
              <div>
                Last Update: {formatDate(latestProcess.updated_at || latestProcess.created_at)}
              </div>
            ) : (
              <div>No processing history for this document.</div>
            )}
          </Stack>
        </Tile>

        <Stack
          orientation="horizontal"
          gap={4}
          style={{
            marginBottom: "2rem",
            flexWrap: "wrap",
          }}
        >
          <PermissionGuard permission={PERMISSIONS.documentsRead}>
            <Button kind="primary" disabled={!canDownload} onClick={handleDownload}>
              Download
            </Button>
          </PermissionGuard>

          <PermissionGuard permission={PERMISSIONS.documentsRead}>
            <Button kind="secondary" disabled={!canPreview} onClick={handlePreview}>
              Preview
            </Button>
          </PermissionGuard>

          <PermissionGuard permission={PERMISSIONS.documentsProcess}>
            <Button
              kind="secondary"
              disabled={!canReprocess || reprocessing}
              onClick={handleReprocess}
            >
              {reprocessing ? "Reprocessing..." : "Reprocess"}
            </Button>
          </PermissionGuard>

          <PermissionGuard permission={PERMISSIONS.documentsWrite}>
            <Button kind="danger--tertiary" onClick={handleDelete} disabled={deleting}>
              {deleting ? "Deleting..." : "Delete"}
            </Button>
          </PermissionGuard>
        </Stack>

        <h3 style={{ marginBottom: "1rem" }}>Process History</h3>

        {!processable ? (
          <Tile>No processing history is available for this document type.</Tile>
        ) : historyRows.length === 0 ? (
          <Tile>No process attempts found.</Tile>
        ) : (
          <DataTable rows={historyRows} headers={headers}>
            {({ rows, headers: tableHeaders, getTableProps, getHeaderProps, getRowProps }) => (
              <TableContainer>
                <Table {...getTableProps()}>
                  <TableHead>
                    <TableRow>
                      {tableHeaders.map((header) => (
                        <TableHeader
                          {...getHeaderProps({
                            header,
                          })}
                        >
                          {header.header}
                        </TableHeader>
                      ))}
                    </TableRow>
                  </TableHead>

                  <TableBody>
                    {rows.map((row) => (
                      <TableRow {...getRowProps({ row })}>
                        {row.cells.map((cell) => {
                          if (cell.info.header === "status") {
                            return (
                              <TableCell key={cell.id}>
                                <StatusTag status={String(cell.value)} />
                              </TableCell>
                            );
                          }

                          if (cell.info.header === "progress") {
                            return (
                              <TableCell key={cell.id}>
                                <div
                                  style={{
                                    width: 150,
                                  }}
                                >
                                  <ProgressBar
                                    value={Number(cell.value) || 0}
                                    max={100}
                                    label=""
                                    size="small"
                                  />
                                </div>
                              </TableCell>
                            );
                          }

                          return <TableCell key={cell.id}>{String(cell.value ?? "—")}</TableCell>;
                        })}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </DataTable>
        )}
      </div>
    </PermissionGuard>
  );
}

function InfoItem({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div
        style={{
          fontSize: "0.75rem",
          color: "#6f6f6f",
          marginBottom: "0.25rem",
        }}
      >
        {label}
      </div>

      <div
        style={{
          fontWeight: 500,
          wordBreak: "break-word",
        }}
      >
        {value}
      </div>
    </div>
  );
}
