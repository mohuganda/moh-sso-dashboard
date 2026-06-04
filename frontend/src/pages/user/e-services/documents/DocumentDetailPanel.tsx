import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Button,
  InlineLoading,
  Modal,
  ProgressBar,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Tag,
} from "@carbon/react";

import { ChartBar, Download, Renew, TrashCan } from "@carbon/icons-react";

import {
  useGetDocumentProcessesQuery,
  useLazyDownloadDocumentQuery,
  useDeleteDocumentMutation,
  useReprocessDocumentMutation,
} from "../../../../store/api/document.api";
import type { DocumentProcess, DocumentResponse } from "../../../../store/types/documents.types";

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatFileSize(bytes?: number) {
  if (!bytes || bytes <= 0) return "0 B";
  const units = ["B", "KB", "MB", "GB"];
  let value = bytes;
  let i = 0;
  while (value >= 1024 && i < units.length - 1) { value /= 1024; i++; }
  return `${value < 10 && i > 0 ? value.toFixed(1) : Math.round(value)} ${units[i]}`;
}

function formatDateTime(iso?: string | null) {
  if (!iso) return "—";
  return new Date(iso).toLocaleString("en-GB", {
    day: "numeric", month: "long", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

function formatDuration(started?: string | null, finished?: string | null) {
  if (!started || !finished) return "—";
  const ms = new Date(finished).getTime() - new Date(started).getTime();
  if (ms < 0) return "—";
  if (ms < 1000) return `${ms}ms`;
  const secs = Math.floor(ms / 1000);
  if (secs < 60) return `${secs}s`;
  const mins = Math.floor(secs / 60);
  const rem = secs % 60;
  return rem > 0 ? `${mins}m ${rem}s` : `${mins}m`;
}

function StatusTag({ status }: { status: string }) {
  const map: Record<string, "gray" | "blue" | "green" | "red" | "magenta"> = {
    PENDING: "gray", PROCESSING: "blue", COMPLETED: "green", FAILED: "red", CANCELLED: "magenta",
  };
  return <Tag type={map[status] ?? "gray"}>{status}</Tag>;
}

function getLatestProcess(processes: DocumentProcess[]) {
  if (!processes.length) return undefined;
  return [...processes].sort(
    (a, b) => new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime(),
  )[0];
}

function requiresProcessing(doc: DocumentResponse) {
  const t = (doc.content_type || "").toLowerCase();
  const f = (doc.original_filename || "").toLowerCase();
  return (
    ["text/csv", "application/vnd.ms-excel",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"].includes(t) ||
    f.endsWith(".csv") || f.endsWith(".xls") || f.endsWith(".xlsx")
  );
}

// ── Info row helper ───────────────────────────────────────────────────────────

function InfoRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div style={{ display: "flex", gap: "0.5rem", padding: "0.45rem 0", borderBottom: "1px solid #f4f4f4" }}>
      <span style={{ width: 160, minWidth: 160, color: "#6f6f6f", fontSize: "0.875rem" }}>{label}</span>
      <span style={{ fontSize: "0.875rem", color: "#161616", wordBreak: "break-all" }}>{value}</span>
    </div>
  );
}

// ── Data preview table ────────────────────────────────────────────────────────

function PreviewTable({ columns, rows }: { columns: string[]; rows: Record<string, unknown>[] }) {
  if (rows.length === 0) {
    return <p style={{ color: "#6f6f6f", padding: "1rem 0", margin: 0 }}>No rows found for this document.</p>;
  }

  const formatCell = (v: unknown): string => {
    if (v === null || v === undefined) return "—";
    if (typeof v === "string") return v || "—";
    return String(v);
  };

  return (
    <div style={{ overflowX: "auto", marginTop: "0.75rem" }}>
      <table style={{ borderCollapse: "collapse", fontSize: "0.8125rem", width: "100%", minWidth: 600 }}>
        <thead>
          <tr style={{ background: "#f4f4f4", borderBottom: "2px solid #e0e0e0" }}>
            {columns.map((col) => (
              <th
                key={col}
                style={{
                  textAlign: "left", padding: "0.4rem 0.75rem",
                  fontWeight: 600, color: "#525252", whiteSpace: "nowrap",
                  textTransform: col.includes("_") ? "none" : undefined,
                }}
              >
                {col.replace(/_/g, " ")}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((row, idx) => (
            <tr key={idx} style={{ borderBottom: "1px solid #e8e8e8", background: idx % 2 ? "#fafafa" : "#fff" }}>
              {columns.map((col) => (
                <td
                  key={col}
                  style={{ padding: "0.4rem 0.75rem", whiteSpace: "nowrap", color: "#161616", maxWidth: 240, overflow: "hidden", textOverflow: "ellipsis" }}
                  title={formatCell(row[col])}
                >
                  {formatCell(row[col])}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

// ── Main component ────────────────────────────────────────────────────────────

type Props = {
  document: DocumentResponse;
  onClose: () => void;
};

export function DocumentDetailPanel({ document, onClose }: Props) {
  const navigate = useNavigate();
  const [confirmDelete, setConfirmDelete] = useState(false);

  const [triggerDownload] = useLazyDownloadDocumentQuery();
  const [deleteDocument, { isLoading: isDeleting }] = useDeleteDocumentMutation();
  const [reprocessDocument, { isLoading: isReprocessing }] = useReprocessDocumentMutation();

  const processable = requiresProcessing(document);
  const templateCode = document.metadata?.template_code ?? "";
  const reportDate = document.metadata?.report_date ?? "—";

  const { data: processes = [], isLoading: isLoadingProcesses } = useGetDocumentProcessesQuery(
    document.id,
    { skip: !processable, pollingInterval: processable ? 4000 : 0 },
  );

  const latest = getLatestProcess(processes);
  const status = (latest?.status || document.status || "UNKNOWN").toUpperCase();
  const progress = latest?.progress ?? (status === "COMPLETED" ? 100 : 0);
  const isBusy = isDeleting || isReprocessing;

  const handleDownload = async () => {
    try {
      const blob = await triggerDownload(document.id).unwrap();
      const url = window.URL.createObjectURL(blob);
      const link = window.document.createElement("a");
      link.href = url;
      link.download = document.original_filename;
      link.click();
      window.URL.revokeObjectURL(url);
    } catch (e) { console.error("Download failed", e); }
  };

  const handleReprocess = async () => {
    try { await reprocessDocument(document.id).unwrap(); }
    catch (e) { console.error("Reprocess failed", e); }
  };

  const handleDeleteConfirm = async () => {
    try {
      await deleteDocument(document.id).unwrap();
      setConfirmDelete(false);
      onClose();
    } catch (e) { console.error("Delete failed", e); }
  };

  return (
    <>
      {/* Delete confirmation modal */}
      <Modal
        open={confirmDelete}
        danger
        modalHeading="Delete document"
        primaryButtonText={isDeleting ? "Deleting…" : "Delete"}
        secondaryButtonText="Cancel"
        primaryButtonDisabled={isDeleting}
        onRequestClose={() => setConfirmDelete(false)}
        onRequestSubmit={() => void handleDeleteConfirm()}
      >
        <p>
          Permanently delete <strong>{document.original_filename}</strong>?
          This also removes all imported rows from the database and cannot be undone.
        </p>
      </Modal>

      <div style={{ display: "flex", flexDirection: "column", height: "100%", gap: "1.25rem" }}>

        {/* ── Status bar ── */}
        <div style={{
          display: "flex", alignItems: "center", gap: "0.75rem", flexWrap: "wrap",
          padding: "0.75rem 1rem", background: "#f4f4f4", borderRadius: 4,
        }}>
          <StatusTag status={status} />
          {templateCode && <Tag type="cyan" style={{ fontFamily: "monospace" }}>{templateCode}</Tag>}
          {reportDate !== "—" && (
            <Tag type="teal">Report: {reportDate}</Tag>
          )}
          <span style={{ color: "#6f6f6f", fontSize: "0.8rem", marginLeft: "auto" }}>
            {formatFileSize(document.size_bytes)}
          </span>
        </div>

        {/* ── Processing progress ── */}
        {processable && (
          <div>
            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "0.25rem" }}>
              <span style={{ fontSize: "0.8rem", color: "#525252" }}>Processing progress</span>
              <span style={{ fontSize: "0.8rem", color: "#525252" }}>{progress}%</span>
            </div>
            <ProgressBar value={progress} max={100} label="" size="md" />
          </div>
        )}

        {/* ── Action buttons ── */}
        <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
          <Button
            renderIcon={Download}
            kind="primary"
            size="sm"
            disabled={status !== "COMPLETED" || isBusy}
            onClick={() => void handleDownload()}
          >
            Download
          </Button>
          {processable && (
            <Button
              renderIcon={Renew}
              kind="secondary"
              size="sm"
              disabled={status !== "FAILED" || isBusy}
              onClick={() => void handleReprocess()}
            >
              {isReprocessing ? "Reprocessing…" : "Reprocess"}
            </Button>
          )}
          {templateCode && status === "COMPLETED" && (
            <Button
              renderIcon={ChartBar}
              kind="tertiary"
              size="sm"
              onClick={() => {
                onClose();
                navigate(`/apps/utilities/self-service/eservice/document-upload/${document.id}/preview`);
              }}
            >
              View Data Preview
            </Button>
          )}
          <Button
            renderIcon={TrashCan}
            kind="danger--ghost"
            size="sm"
            disabled={isBusy}
            onClick={() => setConfirmDelete(true)}
            style={{ marginLeft: "auto" }}
          >
            Delete
          </Button>
        </div>

        {/* ── Tabs ── */}
        <Tabs>
          <TabList aria-label="Document detail tabs" contained>
            <Tab>Document Info</Tab>
            <Tab>Process History</Tab>
          </TabList>

          <TabPanels>
            {/* ── Info ── */}
            <TabPanel style={{ paddingInline: 0, paddingTop: "1rem" }}>
              <div>
                <InfoRow label="Filename" value={document.original_filename} />
                <InfoRow label="Content type" value={document.content_type || "—"} />
                <InfoRow label="File size" value={formatFileSize(document.size_bytes)} />
                <InfoRow label="Uploaded" value={formatDateTime(document.created_at)} />
                <InfoRow label="Last updated" value={formatDateTime(document.updated_at)} />
                <InfoRow label="Report date" value={reportDate} />
                {templateCode && <InfoRow label="Template" value={<code>{templateCode}</code>} />}
                <InfoRow label="Storage location" value={document.storage_location} />
                <InfoRow label="Object key" value={<code style={{ fontSize: "0.8rem" }}>{document.object_key}</code>} />
                {document.checksum_sha256 && (
                  <InfoRow label="SHA-256" value={<code style={{ fontSize: "0.75rem" }}>{document.checksum_sha256}</code>} />
                )}
              </div>
            </TabPanel>

            {/* ── Process History ── */}
            <TabPanel style={{ paddingInline: 0, paddingTop: "1rem" }}>
              {isLoadingProcesses ? (
                <InlineLoading description="Loading processes…" />
              ) : processes.length === 0 ? (
                <p style={{ color: "#6f6f6f", margin: 0 }}>No processing jobs found.</p>
              ) : (
                <table style={{ borderCollapse: "collapse", fontSize: "0.875rem", width: "100%" }}>
                  <thead>
                    <tr style={{ background: "#f4f4f4", borderBottom: "2px solid #e0e0e0" }}>
                      <th style={{ textAlign: "left", padding: "0.45rem 0.75rem", fontWeight: 600, color: "#525252" }}>Type</th>
                      <th style={{ textAlign: "left", padding: "0.45rem 0.75rem", fontWeight: 600, color: "#525252" }}>Status</th>
                      <th style={{ textAlign: "left", padding: "0.45rem 0.75rem", fontWeight: 600, color: "#525252" }}>Progress</th>
                      <th style={{ textAlign: "left", padding: "0.45rem 0.75rem", fontWeight: 600, color: "#525252" }}>Started</th>
                      <th style={{ textAlign: "left", padding: "0.45rem 0.75rem", fontWeight: 600, color: "#525252" }}>Duration</th>
                      <th style={{ textAlign: "left", padding: "0.45rem 0.75rem", fontWeight: 600, color: "#525252" }}>Message</th>
                    </tr>
                  </thead>
                  <tbody>
                    {[...processes]
                      .sort((a, b) => new Date(b.created_at ?? 0).getTime() - new Date(a.created_at ?? 0).getTime())
                      .map((p) => (
                        <tr key={p.id} style={{ borderBottom: "1px solid #e8e8e8" }}>
                          <td style={{ padding: "0.45rem 0.75rem", fontFamily: "monospace", fontSize: "0.8rem" }}>{p.status === "FAILED" ? p.status : p.status}</td>
                          <td style={{ padding: "0.45rem 0.75rem" }}><StatusTag status={p.status.toUpperCase()} /></td>
                          <td style={{ padding: "0.45rem 0.75rem" }}>{p.progress}%</td>
                          <td style={{ padding: "0.45rem 0.75rem", fontSize: "0.8rem", whiteSpace: "nowrap" }}>{formatDateTime(p.started_at)}</td>
                          <td style={{ padding: "0.45rem 0.75rem", fontSize: "0.8rem" }}>{formatDuration(p.started_at, p.finished_at)}</td>
                          <td style={{ padding: "0.45rem 0.75rem", fontSize: "0.8rem", color: p.status === "FAILED" ? "#da1e28" : "#525252" }}>
                            {p.message || "—"}
                          </td>
                        </tr>
                      ))}
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
