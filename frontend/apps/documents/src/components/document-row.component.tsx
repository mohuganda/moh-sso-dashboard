import { TableRow, TableCell, Tag, IconButton } from "@carbon/react";
import { DataVis_1, Download, TrashCan, Renew } from "@carbon/react/icons";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import {
  useGetDocumentProcessesQuery,
  useLazyDownloadDocumentQuery,
  useDeleteDocumentMutation,
  useReprocessDocumentMutation,
} from "../api";
import { useGetUserQuery } from "@moh-sso/users/api";
import { selectUser } from "@moh-sso/auth";
import type { DocumentProcess, DocumentResponse } from "../types";

// ── Helpers ───────────────────────────────────────────────────────────────────

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

function requiresProcessing(document: DocumentResponse) {
  const type = (document.content_type || "").toLowerCase();
  const filename = (document.original_filename || "").toLowerCase();
  if (
    type === "text/csv" ||
    type === "application/vnd.ms-excel" ||
    type === "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  ) return true;
  return filename.endsWith(".csv") || filename.endsWith(".xls") || filename.endsWith(".xlsx");
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) { value /= 1024; unitIndex++; }
  return `${value < 10 && unitIndex > 0 ? value.toFixed(1) : value.toFixed(0)} ${units[unitIndex]}`;
}

function formatFileType(contentType: string, filename?: string): string {
  switch (contentType) {
    case "application/pdf": return "pdf";
    case "text/csv": return "csv";
    case "application/vnd.ms-excel": return "xls";
    case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": return "xlsx";
  }
  if (filename) {
    const ext = filename.split(".").pop()?.toLowerCase();
    if (ext && ext.length <= 5) return ext;
  }
  return contentType || "—";
}

function formatDateTime(iso?: string): string {
  if (!iso) return "—";
  const d = new Date(iso);
  const date = d.toLocaleDateString("en-GB", { day: "numeric", month: "long", year: "numeric" });
  const time = d.toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" });
  return `${date} at ${time}`;
}


function getLatestProcess(processes: DocumentProcess[]): DocumentProcess | undefined {
  if (!processes.length) return undefined;
  return [...processes].sort(
    (a, b) => new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime(),
  )[0];
}

// ── Component ─────────────────────────────────────────────────────────────────

type Props = {
  document: DocumentResponse;
};

export function DocumentRow({ document }: Props) {
  const navigate = useNavigate();
  const [triggerDownload] = useLazyDownloadDocumentQuery();
  const [deleteDocument, { isLoading: isDeleting }] = useDeleteDocumentMutation();
  const [reprocessDocument, { isLoading: isReprocessing }] = useReprocessDocumentMutation();

  const processable = requiresProcessing(document);

  const { data: processes = [] } = useGetDocumentProcessesQuery(document.id, {
    skip: !processable,
    pollingInterval: processable ? 3000 : 0,
  });

  const latest = getLatestProcess(processes);
  const status = (latest?.status || document.status || "UNKNOWN").toUpperCase();

  const reportDate = document.metadata?.report_date ?? "—";
  const isBusy = isDeleting || isReprocessing;

  // ── Uploader name ──
  const currentUser = useSelector(selectUser);
  const isOwnUpload = !!document.uploaded_by && currentUser?.id === document.uploaded_by;
  const { data: uploaderUser, isLoading: isLoadingUploader } = useGetUserQuery(
    document.uploaded_by,
    { skip: !document.uploaded_by || isOwnUpload },
  );
  const uploaderName = !document.uploaded_by
    ? "—"
    : isOwnUpload
      ? [currentUser!.firstName, currentUser!.lastName].filter(Boolean).join(" ") || currentUser!.username
      : isLoadingUploader
        ? "Loading…"
        : uploaderUser
          ? [uploaderUser.firstName, uploaderUser.lastName].filter(Boolean).join(" ") || uploaderUser.username
          : "—";

  const handleDownload = async () => {
    try {
      const blob = await triggerDownload(document.id).unwrap();
      const url = window.URL.createObjectURL(blob);
      const link = window.document.createElement("a");
      link.href = url;
      link.download = document.original_filename;
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (e) {
      console.error("Download failed", e);
    }
  };

  const handleDelete = async () => {
    if (!window.confirm(`Delete "${document.original_filename}"? This cannot be undone.`)) return;
    try {
      await deleteDocument(document.id).unwrap();
    } catch (e) {
      console.error("Delete failed", e);
    }
  };

  const handleReprocess = async () => {
    try {
      await reprocessDocument(document.id).unwrap();
    } catch (e) {
      console.error("Reprocess failed", e);
    }
  };

  return (
    <TableRow>
      <TableCell>
        <div style={{ fontWeight: 500 }}>{document.original_filename}</div>
        <div style={{ fontSize: "0.75rem", color: "#6f6f6f", marginTop: 2 }}>
          {formatFileType(document.content_type, document.original_filename).toUpperCase()}
          {" · "}
          {formatFileSize(document.size_bytes)}
        </div>
      </TableCell>
      <TableCell style={{ fontSize: "0.875rem", color: reportDate === "—" ? "#8d8d8d" : "#161616" }}>
        {reportDate}
      </TableCell>
      <TableCell style={{ fontSize: "0.875rem", color: "#525252" }}>
        {uploaderName}
      </TableCell>
      <TableCell style={{ fontSize: "0.875rem", whiteSpace: "nowrap" }}>
        {formatDateTime(document.created_at)}
      </TableCell>
      <TableCell>
        <StatusTag status={status} />
      </TableCell>
      <TableCell>
        <div style={{ display: "flex", gap: "0.125rem", alignItems: "center" }}>
          <IconButton
            label="View Details"
            kind="ghost"
            size="sm"
            onClick={() => navigate(`/apps/utilities/self-service/eservice/document-upload/${document.id}`)}
          >
            <DataVis_1 />
          </IconButton>
          <IconButton
            label="Download"
            kind="ghost"
            size="sm"
            onClick={() => void handleDownload()}
            disabled={status !== "COMPLETED" || isBusy}
          >
            <Download />
          </IconButton>
          <IconButton
            label={isReprocessing ? "Reprocessing…" : "Reprocess"}
            kind="ghost"
            size="sm"
            onClick={() => void handleReprocess()}
            disabled={!["FAILED", "COMPLETED"].includes(status) || isBusy}
          >
            <Renew />
          </IconButton>
          <IconButton
            label="Delete"
            kind="danger--ghost"
            size="sm"
            onClick={() => void handleDelete()}
            disabled={isBusy}
          >
            <TrashCan />
          </IconButton>
        </div>
      </TableCell>
    </TableRow>
  );
}
