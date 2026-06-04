import {
  TableRow,
  TableCell,
  Tag,
  ProgressBar,
  OverflowMenu,
  OverflowMenuItem,
} from "@carbon/react";
import { useNavigate } from "react-router-dom";
import {
  useGetDocumentProcessesQuery,
  useLazyDownloadDocumentQuery,
  useLazyViewDocumentQuery,
  useReprocessDocumentMutation,
} from "@/store/api/document.api";
import type { DocumentProcess, DocumentResponse } from "@/store/types/documents.types";

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

type Props = {
  document: DocumentResponse;
};

function requiresProcessing(document: DocumentResponse) {
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
  if (!bytes || bytes <= 0) return "0 B";

  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  return `${value < 10 && unitIndex > 0 ? value.toFixed(1) : value.toFixed(0)} ${units[unitIndex]}`;
}

function getLatestProcess(processes: DocumentProcess[]): DocumentProcess | undefined {
  if (!processes.length) return undefined;

  return [...processes].sort(
    (a, b) => new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime(),
  )[0];
}

export function DocumentRow({ document }: Props) {
  const navigate = useNavigate();
  const [reprocessDocument, { isLoading: isReprocessing }] = useReprocessDocumentMutation();
  const [triggerDownload] = useLazyDownloadDocumentQuery();
  const [triggerView] = useLazyViewDocumentQuery();

  const processable = requiresProcessing(document);

  const { data: processes = [] } = useGetDocumentProcessesQuery(document.id, {
    skip: !processable,
    pollingInterval: processable ? 3000 : 0,
  });

  const latest = getLatestProcess(processes);

  const status = (latest?.status || document.status || "UNKNOWN").toUpperCase();
  const progress = latest?.progress ?? (status === "COMPLETED" ? 100 : 0);

  const handleView = async () => {
    try {
      const blob = await triggerView(document.id).unwrap();
      const url = window.URL.createObjectURL(blob);
      window.open(url, "_blank", "noopener,noreferrer");
    } catch (error) {
      console.error("Failed to preview document", error);
      navigate(`/apps/utilities/self-service/eservice/document-upload/${document.id}`);
    }
  };

  const handleDownload = async () => {
    try {
      const blob = await triggerDownload(document.id).unwrap();
      const url = window.URL.createObjectURL(blob);
      const link = window.document.createElement("a");
      link.href = url;
      link.download = document.original_filename;
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      console.error("Failed to download document", error);
    }
  };

  const handleReprocess = async () => {
    try {
      await reprocessDocument(document.id).unwrap();
    } catch (error) {
      console.error("Failed to reprocess document", error);
    }
  };

  return (
    <TableRow>
      <TableCell>{document.original_filename}</TableCell>
      <TableCell>{document.content_type || "—"}</TableCell>
      <TableCell>{formatFileSize(document.size_bytes)}</TableCell>

      <TableCell>
        <StatusTag status={status} />
      </TableCell>

      <TableCell>
        <div style={{ width: 150 }}>
          <ProgressBar value={progress} max={100} label="" size="small" />
        </div>
      </TableCell>

      <TableCell>
        <OverflowMenu size="sm" ariaLabel="Document actions">
          <OverflowMenuItem
            itemText="View Details"
            onClick={() =>
              navigate(`/apps/utilities/self-service/eservice/document-upload/${document.id}`)
            }
          />
          <OverflowMenuItem itemText="Open / Preview" onClick={handleView} />
          <OverflowMenuItem
            itemText="Download"
            onClick={handleDownload}
            disabled={status !== "COMPLETED"}
          />
          <OverflowMenuItem
            itemText={isReprocessing ? "Reprocessing..." : "Reprocess"}
            onClick={handleReprocess}
            disabled={!processable || isReprocessing || status === "PROCESSING"}
          />
        </OverflowMenu>
      </TableCell>
    </TableRow>
  );
}
