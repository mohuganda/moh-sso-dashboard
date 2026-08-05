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
import { useModal, useToast } from "@moh-sso/ui";
import type { DocumentProcess, DocumentResponse } from "../types";

// ── Helpers ───────────────────────────────────────────────────────────────────

function StatusTag({ status, reason }: { status: string; reason?: string | null }) {
  const colorMap: Record<string, "gray" | "blue" | "green" | "red" | "magenta"> = {
    PENDING: "gray",
    PROCESSING: "blue",
    COMPLETED: "green",
    FAILED: "red",
    CANCELLED: "magenta",
  };
  return (
    <Tag type={colorMap[status] || "gray"} title={status === "FAILED" ? (reason ?? undefined) : undefined}>
      {status}
    </Tag>
  );
}

function requiresProcessing(document: DocumentResponse) {
  const type = (document.content_type || "").toLowerCase();
  const filename = (document.original_filename || "").toLowerCase();
  if (
    type === "text/csv" ||
    type === "application/vnd.ms-excel" ||
    type === "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
  )
    return true;
  return filename.endsWith(".csv") || filename.endsWith(".xls") || filename.endsWith(".xlsx");
}

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex++;
  }
  return `${value < 10 && unitIndex > 0 ? value.toFixed(1) : value.toFixed(0)} ${units[unitIndex]}`;
}

function formatFileType(contentType: string, filename?: string): string {
  switch (contentType) {
    case "application/pdf":
      return "pdf";
    case "text/csv":
      return "csv";
    case "application/vnd.ms-excel":
      return "xls";
    case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
      return "xlsx";
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
  const { openModal, closeModal } = useModal();
  const toast = useToast();
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
  const templateCode = document.metadata?.template_code || "No template";
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
      ? [currentUser!.firstName, currentUser!.lastName].filter(Boolean).join(" ") ||
        currentUser!.username
      : isLoadingUploader
        ? "Loading…"
        : uploaderUser
          ? [uploaderUser.firstName, uploaderUser.lastName].filter(Boolean).join(" ") ||
            uploaderUser.username
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
      toast.success("Download started", document.original_filename);
    } catch {
      toast.error("Download failed", "Please try again.");
    }
  };

  const deleteCurrentDocument = async () => {
    try {
      await deleteDocument(document.id).unwrap();
      closeModal();
      toast.success("Document deleted", document.original_filename);
    } catch {
      toast.error("Delete failed", "Please try again.");
    }
  };

  const handleDelete = () => {
    openModal({
      title: "Delete document",
      onClose: closeModal,
      content: (
        <p>
          Permanently delete <strong>{document.original_filename}</strong>? This cannot be undone.
        </p>
      ),
      primaryAction: {
        label: isDeleting ? "Deleting..." : "Delete",
        kind: "danger",
        disabled: isDeleting,
        onClick: () => void deleteCurrentDocument(),
      },
      secondaryAction: {
        label: "Cancel",
        onClick: closeModal,
      },
    });
  };

  const handleReprocess = async () => {
    try {
      await reprocessDocument(document.id).unwrap();
      toast.success("Reprocess started", document.original_filename);
    } catch {
      toast.error("Reprocess failed", "Please try again.");
    }
  };

  return (
    <TableRow>
      <TableCell>
        <div className="documents-file-cell">
          <span className="documents-file-cell__name">{document.original_filename}</span>
        </div>
        <div className="documents-file-cell__meta">
          {formatFileType(document.content_type, document.original_filename).toUpperCase()}
          {" · "}
          {formatFileSize(document.size_bytes)}
        </div>
      </TableCell>
      <TableCell className={templateCode === "No template" ? "documents-muted-cell" : undefined}>
        {templateCode}
      </TableCell>
      <TableCell className={reportDate === "—" ? "documents-muted-cell" : undefined}>
        {reportDate}
      </TableCell>
      <TableCell className="documents-secondary-cell">{uploaderName}</TableCell>
      <TableCell className="documents-nowrap-cell">
        {formatDateTime(document.created_at)}
      </TableCell>
      <TableCell>
        <StatusTag status={status} reason={latest?.error ?? latest?.message} />
      </TableCell>
      <TableCell>
        <div className="documents-row-actions">
          <IconButton
            label="View Details"
            kind="ghost"
            size="sm"
            onClick={() => navigate(`./${document.id}`)}
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
            kind="ghost"
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
