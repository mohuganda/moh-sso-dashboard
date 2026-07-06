import { useState } from "react";
import { useSelector } from "react-redux";
import {
  Button,
  IconButton,
  InlineLoading,
  InlineNotification,
  Modal,
  Search,
  Tag,
} from "@carbon/react";
import {
  Add,
  Archive,
  ChevronDown,
  ChevronRight,
  Download,
  Edit,
  Renew,
  TaskComplete,
  TrashCan,
} from "@carbon/react/icons";
import ExcelJS from "exceljs";

import { useHeaderPanel } from "@moh-sso/ui";
import { UploadTemplateModal } from "./UploadTemplateModal";
import { EditTemplateModal } from "./EditTemplateModal";
import {
  useGetTemplatesQuery,
  useGetTemplateStructureQuery,
  useLazyGetTemplateStructureQuery,
  usePublishTemplateMutation,
  useArchiveTemplateMutation,
  useDeleteTemplateMutation,
} from "../api";
import { useGetUserQuery } from "@moh-sso/users/api";
import { selectUser } from "@moh-sso/auth";
import type { DocumentTemplate, TemplateStructure } from "../types";

// ── Helpers ───────────────────────────────────────────────────────────────────

function formatFileType(raw: string): string {
  switch (raw?.toLowerCase()) {
    case "application/pdf": return "pdf";
    case "text/csv": return "csv";
    case "application/vnd.ms-excel": return "xls";
    case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": return "xlsx";
  }
  // already a short extension like "xlsx"
  return raw || "—";
}

function formatDateTime(iso?: string): string {
  if (!iso) return "—";
  return new Date(iso).toLocaleString("en-GB", {
    day: "2-digit", month: "short", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
}

async function downloadTemplateFile(template: DocumentTemplate, structure: TemplateStructure) {
  const workbook = new ExcelJS.Workbook();
  workbook.creator = "MOH SSO Dashboard";
  workbook.created = new Date();

  const YELLOW = { type: "pattern" as const, pattern: "solid" as const, fgColor: { argb: "FFFFF2CC" } };
  const BLUE   = { type: "pattern" as const, pattern: "solid" as const, fgColor: { argb: "FFD0E4FF" } };

  for (const sheet of structure.sheets) {
    const ws = workbook.addWorksheet(sheet.name);
    const ordered = sheet.columns
      .slice()
      .sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0));

    ws.columns = ordered.map((col) => ({
      header: col.column_name,
      key: col.column_key,
      width: Math.max(col.column_name.length + 4, 18),
    }));

    // Style header row
    const headerRow = ws.getRow(1);
    headerRow.eachCell((cell, colNumber) => {
      const col = ordered[colNumber - 1];
      cell.fill = col?.required ? YELLOW : BLUE;
      cell.font = { bold: true, size: 11 };
      cell.alignment = { vertical: "middle", horizontal: "left" };
      cell.border = {
        bottom: { style: "medium", color: { argb: "FF000000" } },
      };
    });
    headerRow.height = 20;
  }

  const buffer = await workbook.xlsx.writeBuffer();
  const blob = new Blob([buffer], { type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${template.code}_template.xlsx`;
  a.click();
  URL.revokeObjectURL(url);
}

// ── Status helpers ────────────────────────────────────────────────────────────

function getTemplateStatus(t: DocumentTemplate): "active" | "archived" | "draft" {
  if (t.archived_at) return "archived";
  if (t.is_active) return "active";
  return "draft";
}

function StatusTag({ template }: { template: DocumentTemplate }) {
  const status = getTemplateStatus(template);
  const map = {
    active: { type: "green" as const, label: "Active" },
    draft: { type: "gray" as const, label: "Draft" },
    archived: { type: "cool-gray" as const, label: "Archived" },
  };
  const { type, label } = map[status];
  return <Tag type={type}>{label}</Tag>;
}

// ── Uploader name resolver ────────────────────────────────────────────────────

function UploaderName({ userId }: { userId?: string }) {
  const currentUser = useSelector(selectUser);
  const isOwn = !!userId && currentUser?.id === userId;

  const { data: user, isLoading } = useGetUserQuery(userId!, {
    skip: !userId || isOwn,
  });

  if (!userId) return <>—</>;
  if (isOwn) {
    const name = [currentUser!.firstName, currentUser!.lastName].filter(Boolean).join(" ");
    return <>{name || currentUser!.username}</>;
  }
  if (isLoading) return <span style={{ color: "#6f6f6f" }}>Loading…</span>;
  if (!user) return <>—</>;
  const name = [user.firstName, user.lastName].filter(Boolean).join(" ");
  return <>{name || user.username}</>;
}

// ── Structure expanded view ───────────────────────────────────────────────────

function TemplateStructureView({ code }: { code: string }) {
  const { data: structure, isLoading, isError } = useGetTemplateStructureQuery(code);

  if (isLoading) return <InlineLoading description="Loading structure..." style={{ padding: "1rem 1.5rem" }} />;
  if (isError || !structure) {
    return (
      <p style={{ padding: "1rem 1.5rem", color: "#da1e28", margin: 0 }}>
        Failed to load template structure.
      </p>
    );
  }

  if (structure.sheets.length === 0) {
    return <p style={{ padding: "1rem 1.5rem", color: "#6f6f6f", margin: 0 }}>No sheets defined.</p>;
  }

  return (
    <div style={{ padding: "1rem 1.5rem 1.25rem", background: "#f4f4f4" }}>
      {structure.sheets.map((sheet) => (
        <div key={sheet.id} style={{ marginBottom: sheet === structure.sheets[structure.sheets.length - 1] ? 0 : "1.25rem" }}>
          {structure.sheets.length > 1 && (
            <p style={{ margin: "0 0 0.5rem", fontWeight: 600, fontSize: "0.875rem" }}>
              Sheet: {sheet.name}
            </p>
          )}

          {sheet.columns.length === 0 ? (
            <p style={{ margin: 0, color: "#6f6f6f", fontSize: "0.875rem" }}>No columns defined.</p>
          ) : (
            <table style={{ borderCollapse: "collapse", fontSize: "0.8125rem", width: "100%", maxWidth: 640 }}>
              <thead>
                <tr style={{ borderBottom: "1px solid #c6c6c6" }}>
                  <th style={{ textAlign: "left", padding: "0.35rem 0.75rem", fontWeight: 600, color: "#525252" }}>Column</th>
                  <th style={{ textAlign: "left", padding: "0.35rem 0.75rem", fontWeight: 600, color: "#525252" }}>Type</th>
                  <th style={{ textAlign: "center", padding: "0.35rem 0.75rem", fontWeight: 600, color: "#525252" }}>Required</th>
                </tr>
              </thead>
              <tbody>
                {sheet.columns
                  .slice()
                  .sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0))
                  .map((col) => (
                    <tr
                      key={col.id}
                      style={{
                        borderBottom: "1px solid #e8e8e8",
                        background: col.required ? "#fff" : undefined,
                      }}
                    >
                      <td
                        style={{
                          padding: "0.35rem 0.75rem",
                          fontFamily: "monospace",
                          fontWeight: col.required ? 600 : 400,
                          color: col.required ? "#161616" : "#525252",
                        }}
                      >
                        {col.column_name}
                      </td>
                      <td style={{ padding: "0.35rem 0.75rem", color: "#525252" }}>
                        {col.data_type}
                      </td>
                      <td style={{ padding: "0.35rem 0.75rem", textAlign: "center" }}>
                        {col.required ? (
                          <span style={{ color: "#198038", fontWeight: 700 }}>✓</span>
                        ) : (
                          <span style={{ color: "#8d8d8d" }}>—</span>
                        )}
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
          )}
        </div>
      ))}
    </div>
  );
}

// ── Delete confirmation modal ─────────────────────────────────────────────────

type DeleteModalProps = {
  template: DocumentTemplate | null;
  onConfirm: () => void;
  onCancel: () => void;
  isLoading: boolean;
};

function DeleteConfirmModal({ template, onConfirm, onCancel, isLoading }: DeleteModalProps) {
  return (
    <Modal
      open={template !== null}
      danger
      modalHeading="Delete template"
      primaryButtonText={isLoading ? "Deleting..." : "Delete"}
      secondaryButtonText="Cancel"
      onRequestClose={onCancel}
      onRequestSubmit={onConfirm}
      primaryButtonDisabled={isLoading}
    >
      {template && (
        <p>
          Are you sure you want to delete <strong>{template.name}</strong> ({template.code})?
          This cannot be undone.
        </p>
      )}
    </Modal>
  );
}

// ── Main component ────────────────────────────────────────────────────────────

export function TemplatesTab() {
  const { openPanel, closePanel } = useHeaderPanel();

  const [search, setSearch] = useState("");
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [deleteTarget, setDeleteTarget] = useState<DocumentTemplate | null>(null);

  const { data: templates = [], isLoading, isError } = useGetTemplatesQuery();
  const [publishTemplate, { isLoading: isPublishing }] = usePublishTemplateMutation();
  const [archiveTemplate, { isLoading: isArchiving }] = useArchiveTemplateMutation();
  const [deleteTemplate, { isLoading: isDeleting }] = useDeleteTemplateMutation();
  const [getStructure] = useLazyGetTemplateStructureQuery();

  const isMutating = isPublishing || isArchiving || isDeleting;

  const filtered = templates.filter((t) => {
    const q = search.trim().toLowerCase();
    if (!q) return true;
    return (
      t.name.toLowerCase().includes(q) ||
      t.code.toLowerCase().includes(q) ||
      t.file_type.toLowerCase().includes(q)
    );
  });

  function toggleExpand(id: string) {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) { next.delete(id); } else { next.add(id); }
      return next;
    });
  }

  function handleNewTemplate() {
    openPanel({
      title: "Upload Template",
      content: <UploadTemplateModal onClose={closePanel} />,
      size: "lg",
    });
  }

  function handleEdit(t: DocumentTemplate) {
    openPanel({
      title: "Edit Template",
      content: <EditTemplateModal template={t} onClose={closePanel} />,
      size: "lg",
    });
  }

  async function handleDownload(t: DocumentTemplate) {
    const result = await getStructure(t.code);
    if (result.data) {
      await downloadTemplateFile(t, result.data);
    }
  }

  async function handlePublish(t: DocumentTemplate) {
    try {
      await publishTemplate(t.id).unwrap();
    } catch {
      // error surfaced via RTK Query
    }
  }

  async function handleArchive(t: DocumentTemplate) {
    try {
      await archiveTemplate(t.id).unwrap();
    } catch {
      // error surfaced via RTK Query
    }
  }

  async function handleDeleteConfirm() {
    if (!deleteTarget) return;
    try {
      await deleteTemplate(deleteTarget.id).unwrap();
    } finally {
      setDeleteTarget(null);
    }
  }

  // ── Render ──────────────────────────────────────────────────────────────────

  if (isLoading) {
    return <InlineLoading description="Loading templates..." style={{ padding: "2rem 0" }} />;
  }

  if (isError) {
    return (
      <InlineNotification
        kind="error"
        title="Failed to load templates"
        subtitle="Please refresh and try again."
        lowContrast
      />
    );
  }

  return (
    <>
      <DeleteConfirmModal
        template={deleteTarget}
        onConfirm={() => void handleDeleteConfirm()}
        onCancel={() => setDeleteTarget(null)}
        isLoading={isDeleting}
      />

      <div style={{ marginBottom: "1rem", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          {templates.length} template{templates.length !== 1 ? "s" : ""} total —{" "}
          {templates.filter((t) => t.is_active && !t.archived_at).length} active
        </p>
        <Button renderIcon={Add} onClick={handleNewTemplate} size="sm">
          New Template
        </Button>
      </div>

      {/* Search */}
      <div style={{ marginBottom: "0.5rem" }}>
        <Search
          placeholder="Search by name, code, or file type"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          labelText="Search templates"
          size="lg"
        />
      </div>

      {filtered.length === 0 ? (
        <div
          style={{
            padding: "3rem 1rem",
            textAlign: "center",
            color: "#6f6f6f",
            border: "1px dashed #c6c6c6",
          }}
        >
          {templates.length === 0
            ? 'No templates yet. Click "New Template" to create one.'
            : "No templates match your search."}
        </div>
      ) : (
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.875rem" }}>
          <thead>
            <tr style={{ borderBottom: "2px solid #e0e0e0", background: "#f4f4f4" }}>
              <th style={{ width: 32 }} />
              <th style={{ textAlign: "left", padding: "0.6rem 1rem", fontWeight: 600 }}>Name</th>
              <th style={{ textAlign: "left", padding: "0.6rem 1rem", fontWeight: 600 }}>Code</th>
              <th style={{ textAlign: "left", padding: "0.6rem 1rem", fontWeight: 600 }}>File type</th>
              <th style={{ textAlign: "left", padding: "0.6rem 1rem", fontWeight: 600 }}>Uploaded by</th>
              <th style={{ textAlign: "left", padding: "0.6rem 1rem", fontWeight: 600 }}>Uploaded</th>
              <th style={{ textAlign: "left", padding: "0.6rem 1rem", fontWeight: 600 }}>Status</th>
              <th style={{ textAlign: "left", padding: "0.6rem 1rem", fontWeight: 600 }}>Version</th>
              <th style={{ textAlign: "right", padding: "0.6rem 0.75rem", fontWeight: 600, width: 150 }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((t) => {
              const isExpanded = expanded.has(t.id);
              const status = getTemplateStatus(t);

              return (
                <>
                  <tr
                    key={t.id}
                    style={{
                      borderBottom: isExpanded ? "none" : "1px solid #e0e0e0",
                      background: isExpanded ? "#f9f9f9" : "#fff",
                    }}
                  >
                    {/* Expand toggle */}
                    <td style={{ padding: "0 0 0 0.75rem", width: 32 }}>
                      <button
                        onClick={() => toggleExpand(t.id)}
                        style={{
                          background: "none",
                          border: "none",
                          cursor: "pointer",
                          padding: "0.5rem 0.25rem",
                          color: "#525252",
                          display: "flex",
                          alignItems: "center",
                        }}
                        aria-label={isExpanded ? "Collapse" : "Expand"}
                      >
                        {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                      </button>
                    </td>

                    <td style={{ padding: "0.6rem 1rem", fontWeight: 500 }}>
                      {t.name}
                      {t.description && (
                        <div style={{ fontSize: "0.8rem", color: "#6f6f6f", fontWeight: 400, marginTop: 2 }}>
                          {t.description}
                        </div>
                      )}
                    </td>

                    <td style={{ padding: "0.6rem 1rem", fontFamily: "monospace", color: "#525252" }}>
                      {t.code}
                    </td>

                    <td style={{ padding: "0.6rem 1rem", textTransform: "uppercase", fontSize: "0.8rem", color: "#525252" }}>
                      {formatFileType(t.file_type)}
                    </td>

                    <td style={{ padding: "0.6rem 1rem", fontSize: "0.8rem", color: "#525252" }}>
                      <UploaderName userId={t.created_by} />
                    </td>

                    <td style={{ padding: "0.6rem 1rem", fontSize: "0.8rem", color: "#525252", whiteSpace: "nowrap" }}>
                      {formatDateTime(t.created_at)}
                    </td>

                    <td style={{ padding: "0.6rem 1rem" }}>
                      <StatusTag template={t} />
                    </td>

                    <td style={{ padding: "0.6rem 1rem", color: "#525252" }}>v{t.version}</td>

                    <td style={{ padding: "0.4rem 0.75rem", textAlign: "right" }}>
                      <div style={{ display: "flex", gap: "0.125rem", justifyContent: "flex-end", alignItems: "center" }}>
                        {status === "draft" && (
                          <IconButton
                            label="Publish"
                            kind="ghost"
                            size="sm"
                            onClick={() => void handlePublish(t)}
                            disabled={isMutating}
                          >
                            <TaskComplete />
                          </IconButton>
                        )}
                        {status === "active" && (
                          <IconButton
                            label="Archive"
                            kind="ghost"
                            size="sm"
                            onClick={() => void handleArchive(t)}
                            disabled={isMutating}
                          >
                            <Archive />
                          </IconButton>
                        )}
                        {status === "archived" && (
                          <IconButton
                            label="Re-publish"
                            kind="ghost"
                            size="sm"
                            onClick={() => void handlePublish(t)}
                            disabled={isMutating}
                          >
                            <Renew />
                          </IconButton>
                        )}
                        <IconButton
                          label="Download empty template"
                          kind="ghost"
                          size="sm"
                          onClick={() => void handleDownload(t)}
                          disabled={isMutating}
                        >
                          <Download />
                        </IconButton>
                        <IconButton
                          label="Edit"
                          kind="ghost"
                          size="sm"
                          onClick={() => handleEdit(t)}
                          disabled={isMutating}
                        >
                          <Edit />
                        </IconButton>
                        <IconButton
                          label="Delete"
                          kind="danger--ghost"
                          size="sm"
                          onClick={() => setDeleteTarget(t)}
                          disabled={isMutating}
                        >
                          <TrashCan />
                        </IconButton>
                      </div>
                    </td>
                  </tr>

                  {/* Expanded structure row */}
                  {isExpanded && (
                    <tr key={`${t.id}-expanded`} style={{ borderBottom: "1px solid #e0e0e0" }}>
                      <td />
                      <td colSpan={8} style={{ padding: 0 }}>
                        <TemplateStructureView code={t.code} />
                      </td>
                    </tr>
                  )}
                </>
              );
            })}
          </tbody>
        </table>
      )}
    </>
  );
}
