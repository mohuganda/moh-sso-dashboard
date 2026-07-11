import { type ReactNode } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import {
  Button,
  InlineLoading,
  InlineNotification,
  ProgressBar,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Tag,
} from "@carbon/react";
import { ChartBar, Download, Renew, TrashCan } from "@carbon/react/icons";

import { PermissionGuard, PERMISSIONS, selectUser } from "@moh-sso/auth";
import { useModal, useToast } from "@moh-sso/ui";
import { useGetUserQuery } from "@moh-sso/users/api";

import {
  useDeleteDocumentMutation,
  useGetDocumentProcessesQuery,
  useLazyDownloadDocumentQuery,
  useReprocessDocumentMutation,
} from "../api";
import type { DocumentProcess, DocumentResponse } from "../types";
import "./documents-components.scss";

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) {
    return "0 B";
  }

  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let unitIndex = 0;

  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }

  const formattedValue = value < 10 && unitIndex > 0 ? value.toFixed(1) : Math.round(value);

  return `${formattedValue} ${units[unitIndex]}`;
}

function formatDateTime(iso?: string | null) {
  if (!iso) {
    return "—";
  }

  return new Date(iso).toLocaleString("en-GB", {
    day: "numeric",
    month: "long",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatDuration(started?: string | null, finished?: string | null) {
  if (!started || !finished) {
    return "—";
  }

  const milliseconds = new Date(finished).getTime() - new Date(started).getTime();

  if (milliseconds < 0) {
    return "—";
  }

  if (milliseconds < 1000) {
    return `${milliseconds}ms`;
  }

  const seconds = Math.floor(milliseconds / 1000);

  if (seconds < 60) {
    return `${seconds}s`;
  }

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  return remainingSeconds > 0 ? `${minutes}m ${remainingSeconds}s` : `${minutes}m`;
}

function StatusTag({ status }: { status: string }) {
  const statusTypeMap: Record<string, "gray" | "blue" | "green" | "red" | "magenta"> = {
    PENDING: "gray",
    PROCESSING: "blue",
    COMPLETED: "green",
    FAILED: "red",
    CANCELLED: "magenta",
  };

  return <Tag type={statusTypeMap[status] ?? "gray"}>{status}</Tag>;
}

function getLatestProcess(processes: DocumentProcess[]) {
  if (processes.length === 0) {
    return undefined;
  }

  return [...processes].sort(
    (a, b) => new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime(),
  )[0];
}

function requiresProcessing(document: DocumentResponse) {
  const contentType = (document.content_type || "").toLowerCase();

  const filename = (document.original_filename || "").toLowerCase();

  return (
    [
      "text/csv",
      "application/vnd.ms-excel",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ].includes(contentType) ||
    filename.endsWith(".csv") ||
    filename.endsWith(".xls") ||
    filename.endsWith(".xlsx")
  );
}

function InfoRow({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="document-info-row">
      <span className="document-info-row__label">{label}</span>

      <span className="document-info-row__value">{value}</span>
    </div>
  );
}

type Props = {
  document: DocumentResponse;
  onClose: () => void;
};

function DocumentDetailPanelContent({ document, onClose }: Props) {
  const navigate = useNavigate();
  const { openModal, closeModal } = useModal();
  const toast = useToast();

  const [triggerDownload] = useLazyDownloadDocumentQuery();

  const [deleteDocument, { isLoading: isDeleting }] = useDeleteDocumentMutation();

  const [reprocessDocument, { isLoading: isReprocessing }] = useReprocessDocumentMutation();

  const processable = requiresProcessing(document);

  const templateCode = document.metadata?.template_code ?? "";

  const reportDate = document.metadata?.report_date ?? "—";

  const currentUser = useSelector(selectUser);

  const isOwnUpload = currentUser?.id === document.uploaded_by;

  const { data: uploaderUser, isLoading: isLoadingUploader } = useGetUserQuery(
    document.uploaded_by,
    {
      skip: !document.uploaded_by || isOwnUpload,
    },
  );

  const uploaderName = isOwnUpload
    ? [currentUser?.firstName, currentUser?.lastName].filter(Boolean).join(" ") ||
      currentUser?.username ||
      "—"
    : isLoadingUploader
      ? "Loading…"
      : uploaderUser
        ? [uploaderUser.firstName, uploaderUser.lastName].filter(Boolean).join(" ") ||
          uploaderUser.username
        : "—";

  const { data: processes = [], isLoading: isLoadingProcesses } = useGetDocumentProcessesQuery(
    document.id,
    {
      skip: !processable,
      pollingInterval: processable ? 4000 : 0,
    },
  );

  const latest = getLatestProcess(processes);

  const status = (latest?.status || document.status || "UNKNOWN").toUpperCase();

  const progress = latest?.progress ?? (status === "COMPLETED" ? 100 : 0);

  const isBusy = isDeleting || isReprocessing;

  async function handleDownload() {
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
  }

  async function handleReprocess() {
    try {
      await reprocessDocument(document.id).unwrap();
      toast.success("Reprocess started", document.original_filename);
    } catch {
      toast.error("Reprocess failed", "Please try again.");
    }
  }

  async function handleDeleteConfirm() {
    try {
      await deleteDocument(document.id).unwrap();

      closeModal();
      toast.success("Document deleted", document.original_filename);
      onClose();
    } catch {
      toast.error("Delete failed", "Please try again.");
    }
  }

  function handleDeleteRequest() {
    openModal({
      title: "Delete document",
      onClose: closeModal,
      content: (
        <p>
          Permanently delete <strong>{document.original_filename}</strong>? This also removes all
          imported rows from the database and cannot be undone.
        </p>
      ),
      primaryAction: {
        label: isDeleting ? "Deleting..." : "Delete",
        kind: "danger",
        disabled: isDeleting,
        onClick: () => void handleDeleteConfirm(),
      },
      secondaryAction: {
        label: "Cancel",
        onClick: closeModal,
      },
    });
  }

  function handlePreview() {
    onClose();

    navigate(`./${document.id}/preview`);
  }

  return (
    <>
      <div className="document-detail-panel">
        <div className="document-detail-panel__summary">
          <StatusTag status={status} />

          {templateCode && (
            <Tag type="cyan" className="document-detail-panel__mono-tag">
              {templateCode}
            </Tag>
          )}

          {reportDate !== "—" && <Tag type="teal">Report: {reportDate}</Tag>}

          <span className="document-detail-panel__size">{formatFileSize(document.size_bytes)}</span>
        </div>

        {processable && (
          <div>
            <div className="document-detail-panel__progress-header">
              <span className="document-detail-panel__progress-label">Processing progress</span>

              <span className="document-detail-panel__progress-value">{progress}%</span>
            </div>

            <ProgressBar
              value={progress}
              max={100}
              label="Document processing progress"
              hideLabel
              size="small"
            />
          </div>
        )}

        <div className="document-detail-panel__actions">
          <PermissionGuard permission={PERMISSIONS.documentsRead}>
            <Button
              renderIcon={Download}
              kind="primary"
              size="sm"
              disabled={status !== "COMPLETED" || isBusy}
              onClick={() => void handleDownload()}
            >
              Download
            </Button>
          </PermissionGuard>

          {processable && (
            <PermissionGuard permission={PERMISSIONS.documentsProcess}>
              <Button
                renderIcon={Renew}
                kind="secondary"
                size="sm"
                disabled={!["FAILED", "COMPLETED"].includes(status) || isBusy}
                onClick={() => void handleReprocess()}
              >
                {isReprocessing ? "Reprocessing…" : "Reprocess"}
              </Button>
            </PermissionGuard>
          )}

          {templateCode && status === "COMPLETED" && (
            <PermissionGuard permission={PERMISSIONS.documentsRead}>
              <Button
                renderIcon={ChartBar}
                kind="tertiary"
                size="sm"
                disabled={isBusy}
                onClick={handlePreview}
              >
                View Data Preview
              </Button>
            </PermissionGuard>
          )}

          <PermissionGuard permission={PERMISSIONS.documentsWrite}>
            <Button
              renderIcon={TrashCan}
              kind="ghost"
              size="sm"
              disabled={isBusy}
              onClick={handleDeleteRequest}
              className="document-detail-panel__delete-action"
            >
              Delete
            </Button>
          </PermissionGuard>
        </div>

        <Tabs>
          <TabList aria-label="Document detail tabs" contained>
            <Tab>Document Info</Tab>
            <Tab>Process History</Tab>
          </TabList>

          <TabPanels>
            <TabPanel className="document-detail-panel__tab-panel">
              <div>
                <InfoRow label="Filename" value={document.original_filename} />

                <InfoRow label="Content type" value={document.content_type || "—"} />

                <InfoRow label="File size" value={formatFileSize(document.size_bytes)} />

                <InfoRow label="Uploaded" value={formatDateTime(document.created_at)} />

                <InfoRow label="Uploaded by" value={uploaderName} />

                <InfoRow label="Last updated" value={formatDateTime(document.updated_at)} />

                <InfoRow label="Report date" value={reportDate} />

                {templateCode && <InfoRow label="Template" value={<code>{templateCode}</code>} />}

                {status === "FAILED" && (latest?.error || latest?.message) && (
                  <InfoRow
                    label="Failure reason"
                    value={
                      <span className="document-detail-panel__failure">
                        {latest.error ?? latest.message}
                      </span>
                    }
                  />
                )}

                <InfoRow label="Storage location" value={document.storage_location} />

                <InfoRow
                  label="Object key"
                  value={
                    <code className="document-detail-panel__code-sm">{document.object_key}</code>
                  }
                />

                {document.checksum_sha256 && (
                  <InfoRow
                    label="SHA-256"
                    value={
                      <code className="document-detail-panel__code-xs">
                        {document.checksum_sha256}
                      </code>
                    }
                  />
                )}
              </div>
            </TabPanel>

            <TabPanel className="document-detail-panel__tab-panel">
              {isLoadingProcesses ? (
                <InlineLoading description="Loading processes…" />
              ) : processes.length === 0 ? (
                <p className="document-detail-panel__empty">No processing jobs found.</p>
              ) : (
                <table className="document-table">
                  <thead>
                    <tr>
                      <th>Type</th>
                      <th>Status</th>
                      <th>Progress</th>
                      <th>Started</th>
                      <th>Duration</th>
                      <th>Message</th>
                    </tr>
                  </thead>

                  <tbody>
                    {[...processes]
                      .sort(
                        (a, b) =>
                          new Date(b.created_at ?? 0).getTime() -
                          new Date(a.created_at ?? 0).getTime(),
                      )
                      .map((process) => {
                        const processStatus = process.status.toUpperCase();

                        const message =
                          processStatus === "FAILED"
                            ? (process.error ?? process.message)
                            : process.message;

                        return (
                          <tr key={process.id}>
                            <td className="document-table__mono">{processStatus}</td>

                            <td>
                              <StatusTag status={processStatus} />
                            </td>

                            <td>{process.progress}%</td>

                            <td className="document-table__small document-table__nowrap">
                              {formatDateTime(process.started_at)}
                            </td>

                            <td className="document-table__small">
                              {formatDuration(process.started_at, process.finished_at)}
                            </td>

                            <td
                              className={
                                processStatus === "FAILED"
                                  ? "document-table__small document-table__error"
                                  : "document-table__small document-table__muted"
                              }
                            >
                              {message || "—"}
                            </td>
                          </tr>
                        );
                      })}
                  </tbody>
                </table>
              )}
            </TabPanel>
          </TabPanels>
        </Tabs>
      </div>
    </>
  );
}

export function DocumentDetailPanel({ document, onClose }: Props) {
  return (
    <PermissionGuard
      permission={PERMISSIONS.documentsRead}
      fallback={
        <div>
          <InlineNotification
            kind="warning"
            title="Access denied"
            subtitle="You do not have permission to view this document."
            lowContrast
            hideCloseButton
          />

          <div className="document-permission-fallback__actions">
            <Button kind="secondary" onClick={onClose}>
              Close
            </Button>
          </div>
        </div>
      }
    >
      <DocumentDetailPanelContent document={document} onClose={onClose} />
    </PermissionGuard>
  );
}
