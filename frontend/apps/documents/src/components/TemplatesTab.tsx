import { Fragment, useState } from "react";
import { useSelector } from "react-redux";
import {
  Button,
  IconButton,
  InlineLoading,
  InlineNotification,
  Search,
  Tag,
  Tile,
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
import * as XLSX from "xlsx";

import { PermissionGuard, PERMISSIONS, selectUser } from "@moh-sso/auth";
import { useHeaderPanel, useModal, useToast } from "@moh-sso/ui";
import { useGetUserQuery } from "@moh-sso/users/api";

import { UploadTemplateModal } from "./UploadTemplateModal";
import { EditTemplateModal } from "./EditTemplateModal";
import {
  useArchiveTemplateMutation,
  useDeleteTemplateMutation,
  useGetTemplatesQuery,
  useGetTemplateStructureQuery,
  useLazyGetTemplateStructureQuery,
  usePublishTemplateMutation,
} from "../api";
import type { DocumentTemplate, TemplateStructure } from "../types";

function formatFileType(raw: string): string {
  switch (raw?.toLowerCase()) {
    case "application/pdf":
      return "pdf";
    case "text/csv":
      return "csv";
    case "application/vnd.ms-excel":
      return "xls";
    case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
      return "xlsx";
    default:
      return raw || "—";
  }
}

function formatDateTime(iso?: string): string {
  if (!iso) {
    return "—";
  }

  return new Date(iso).toLocaleString("en-GB", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function downloadTemplateFile(template: DocumentTemplate, structure: TemplateStructure) {
  const workbook = XLSX.utils.book_new();

  for (const sheet of structure.sheets) {
    const ordered = sheet.columns
      .slice()
      .sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0));

    const headers = ordered.map((column) => column.column_name);
    const worksheet = XLSX.utils.aoa_to_sheet([headers]);

    worksheet["!cols"] = ordered.map((column) => ({
      wch: Math.max(column.column_name.length + 4, 18),
    }));

    XLSX.utils.book_append_sheet(workbook, worksheet, sheet.name);
  }

  const buffer = XLSX.write(workbook, {
    type: "array",
    bookType: "xlsx",
  }) as ArrayBuffer;

  const blob = new Blob([buffer], {
    type: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  });

  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");

  anchor.href = url;
  anchor.download = `${template.code}_template.xlsx`;
  anchor.click();

  URL.revokeObjectURL(url);
}

function getTemplateStatus(template: DocumentTemplate): "active" | "archived" | "draft" {
  if (template.archived_at) {
    return "archived";
  }

  if (template.is_active) {
    return "active";
  }

  return "draft";
}

function StatusTag({ template }: { template: DocumentTemplate }) {
  const status = getTemplateStatus(template);

  const statusMap = {
    active: {
      type: "green" as const,
      label: "Active",
    },
    draft: {
      type: "gray" as const,
      label: "Draft",
    },
    archived: {
      type: "cool-gray" as const,
      label: "Archived",
    },
  };

  const { type, label } = statusMap[status];

  return <Tag type={type}>{label}</Tag>;
}

function UploaderName({ userId }: { userId?: string }) {
  const currentUser = useSelector(selectUser);
  const isOwn = Boolean(userId) && currentUser?.id === userId;

  const { data: user, isLoading } = useGetUserQuery(userId!, {
    skip: !userId || isOwn,
  });

  if (!userId) {
    return <>—</>;
  }

  if (isOwn && currentUser) {
    const name = [currentUser.firstName, currentUser.lastName].filter(Boolean).join(" ");

    return <>{name || currentUser.username}</>;
  }

  if (isLoading) {
    return <span style={{ color: "#6f6f6f" }}>Loading…</span>;
  }

  if (!user) {
    return <>—</>;
  }

  const name = [user.firstName, user.lastName].filter(Boolean).join(" ");

  return <>{name || user.username}</>;
}

function TemplateStructureView({ code }: { code: string }) {
  const { data: structure, isLoading, isError } = useGetTemplateStructureQuery(code);

  if (isLoading) {
    return <InlineLoading description="Loading structure..." style={{ padding: "1rem 1.5rem" }} />;
  }

  if (isError || !structure) {
    return (
      <p
        style={{
          padding: "1rem 1.5rem",
          color: "#da1e28",
          margin: 0,
        }}
      >
        Failed to load template structure.
      </p>
    );
  }

  if (structure.sheets.length === 0) {
    return (
      <p
        style={{
          padding: "1rem 1.5rem",
          color: "#6f6f6f",
          margin: 0,
        }}
      >
        No sheets defined.
      </p>
    );
  }

  return (
    <div className="documents-template-structure">
      {structure.sheets.map((sheet, sheetIndex) => (
        <div
          key={sheet.id}
          className="documents-template-structure__sheet"
          data-last={sheetIndex === structure.sheets.length - 1 ? "true" : undefined}
        >
          {structure.sheets.length > 1 && (
            <p className="documents-template-structure__sheet-title">Sheet: {sheet.name}</p>
          )}

          {sheet.columns.length === 0 ? (
            <p className="documents-template-structure__empty">No columns defined.</p>
          ) : (
            <table className="documents-template-structure__table">
              <thead>
                <tr>
                  <th>Column</th>

                  <th>Type</th>

                  <th>Required</th>
                </tr>
              </thead>

              <tbody>
                {sheet.columns
                  .slice()
                  .sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0))
                  .map((column) => (
                    <tr
                      key={column.id}
                      className={column.required ? "is-required" : undefined}
                    >
                      <td className="documents-template-structure__column">
                        {column.column_name}
                      </td>

                      <td>{column.data_type}</td>

                      <td>
                        {column.required ? (
                          <Tag type="green" size="sm">Required</Tag>
                        ) : (
                          <span className="documents-muted-cell">—</span>
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

export function TemplatesTab() {
  const { openPanel, closePanel } = useHeaderPanel();
  const { openModal, closeModal } = useModal();
  const toast = useToast();

  const [search, setSearch] = useState("");
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  const { data: templates = [], isLoading, isError } = useGetTemplatesQuery();

  const [publishTemplate, { isLoading: isPublishing }] = usePublishTemplateMutation();

  const [archiveTemplate, { isLoading: isArchiving }] = useArchiveTemplateMutation();

  const [deleteTemplate, { isLoading: isDeleting }] = useDeleteTemplateMutation();

  const [getStructure] = useLazyGetTemplateStructureQuery();

  const isMutating = isPublishing || isArchiving || isDeleting;

  const filtered = templates.filter((template) => {
    const query = search.trim().toLowerCase();

    if (!query) {
      return true;
    }

    return (
      template.name.toLowerCase().includes(query) ||
      template.code.toLowerCase().includes(query) ||
      template.file_type.toLowerCase().includes(query)
    );
  });

  function toggleExpand(id: string) {
    setExpanded((previous) => {
      const next = new Set(previous);

      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }

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

  function handleEdit(template: DocumentTemplate) {
    openPanel({
      title: "Edit Template",
      content: <EditTemplateModal template={template} onClose={closePanel} />,
      size: "lg",
    });
  }

  async function handleDownload(template: DocumentTemplate) {
    try {
      const result = await getStructure(template.code);

      if (result.data) {
        downloadTemplateFile(template, result.data);
        toast.success("Template downloaded", template.name);
        return;
      }

      toast.error("Download failed", "Template structure is unavailable.");
    } catch {
      toast.error("Download failed", "Please try again.");
    }
  }

  async function handlePublish(template: DocumentTemplate) {
    try {
      await publishTemplate(template.id).unwrap();
      toast.success("Template published", template.name);
    } catch {
      toast.error("Publish failed", "Please try again.");
    }
  }

  async function handleArchive(template: DocumentTemplate) {
    try {
      await archiveTemplate(template.id).unwrap();
      toast.success("Template archived", template.name);
    } catch {
      toast.error("Archive failed", "Please try again.");
    }
  }

  async function handleDeleteConfirm(template: DocumentTemplate) {
    try {
      await deleteTemplate(template.id).unwrap();
      closeModal();
      toast.success("Template deleted", template.name);
    } catch {
      toast.error("Delete failed", "Please try again.");
    }
  }

  function handleDeleteRequest(template: DocumentTemplate) {
    openModal({
      title: "Delete template",
      onClose: closeModal,
      content: (
        <p>
          Are you sure you want to delete <strong>{template.name}</strong> ({template.code})? This
          cannot be undone.
        </p>
      ),
      primaryAction: {
        label: isDeleting ? "Deleting..." : "Delete",
        kind: "danger",
        disabled: isDeleting,
        onClick: () => void handleDeleteConfirm(template),
      },
      secondaryAction: {
        label: "Cancel",
        onClick: closeModal,
      },
    });
  }

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
    <PermissionGuard
      permission={PERMISSIONS.documentTemplatesRead}
      fallback={
        <InlineNotification
          kind="warning"
          title="Access denied"
          subtitle="You do not have permission to view document templates."
          lowContrast
        />
      }
    >
      <>
        <div className="documents-templates__toolbar">
          <div>
            <h3>Templates</h3>
            <p>
              {templates.length} template
              {templates.length !== 1 ? "s" : ""} total ·{" "}
              {templates.filter((template) => template.is_active && !template.archived_at).length}{" "}
              active
            </p>
          </div>

          <PermissionGuard permission={PERMISSIONS.documentTemplatesWrite}>
            <Button renderIcon={Add} onClick={handleNewTemplate} size="sm">
              New Template
            </Button>
          </PermissionGuard>
        </div>

        <Tile className="documents-templates__filters">
          <Search
            placeholder="Search by name, code, or file type"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            labelText="Search templates"
            size="lg"
          />
        </Tile>

        {filtered.length === 0 ? (
          <div className="documents-templates__empty">
            {templates.length === 0
              ? "No templates are currently available."
              : "No templates match your search."}
          </div>
        ) : (
          <div className="documents-templates__table-scroll">
            <table className="documents-templates__table">
            <thead>
              <tr>
                <th style={{ width: 32 }} />

                <th
                  style={{
                    textAlign: "left",
                    padding: "0.6rem 1rem",
                    fontWeight: 600,
                  }}
                >
                  Name
                </th>

                <th
                  style={{
                    textAlign: "left",
                    padding: "0.6rem 1rem",
                    fontWeight: 600,
                  }}
                >
                  Code
                </th>

                <th
                  style={{
                    textAlign: "left",
                    padding: "0.6rem 1rem",
                    fontWeight: 600,
                  }}
                >
                  File type
                </th>

                <th
                  style={{
                    textAlign: "left",
                    padding: "0.6rem 1rem",
                    fontWeight: 600,
                  }}
                >
                  Uploaded by
                </th>

                <th
                  style={{
                    textAlign: "left",
                    padding: "0.6rem 1rem",
                    fontWeight: 600,
                  }}
                >
                  Uploaded
                </th>

                <th
                  style={{
                    textAlign: "left",
                    padding: "0.6rem 1rem",
                    fontWeight: 600,
                  }}
                >
                  Status
                </th>

                <th
                  style={{
                    textAlign: "left",
                    padding: "0.6rem 1rem",
                    fontWeight: 600,
                  }}
                >
                  Version
                </th>

                <th
                  style={{
                    textAlign: "right",
                    padding: "0.6rem 0.75rem",
                    fontWeight: 600,
                    width: 150,
                  }}
                >
                  Actions
                </th>
              </tr>
            </thead>

            <tbody>
              {filtered.map((template) => {
                const isExpanded = expanded.has(template.id);
                const status = getTemplateStatus(template);

                return (
                  <Fragment key={template.id}>
                    <tr
                      style={{
                        borderBottom: isExpanded ? "none" : "1px solid #e0e0e0",
                        background: isExpanded ? "#f9f9f9" : "#fff",
                      }}
                    >
                      <td
                        style={{
                          padding: "0 0 0 0.75rem",
                          width: 32,
                        }}
                      >
                        <button
                          type="button"
                          onClick={() => toggleExpand(template.id)}
                          style={{
                            background: "none",
                            border: "none",
                            cursor: "pointer",
                            padding: "0.5rem 0.25rem",
                            color: "#525252",
                            display: "flex",
                            alignItems: "center",
                          }}
                          aria-expanded={isExpanded}
                          aria-label={
                            isExpanded ? `Collapse ${template.name}` : `Expand ${template.name}`
                          }
                        >
                          {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
                        </button>
                      </td>

                      <td
                        style={{
                          padding: "0.6rem 1rem",
                          fontWeight: 500,
                        }}
                      >
                        {template.name}

                        {template.description && (
                          <div
                            style={{
                              fontSize: "0.8rem",
                              color: "#6f6f6f",
                              fontWeight: 400,
                              marginTop: 2,
                            }}
                          >
                            {template.description}
                          </div>
                        )}
                      </td>

                      <td
                        style={{
                          padding: "0.6rem 1rem",
                          fontFamily: "monospace",
                          color: "#525252",
                        }}
                      >
                        {template.code}
                      </td>

                      <td
                        style={{
                          padding: "0.6rem 1rem",
                          textTransform: "uppercase",
                          fontSize: "0.8rem",
                          color: "#525252",
                        }}
                      >
                        {formatFileType(template.file_type)}
                      </td>

                      <td
                        style={{
                          padding: "0.6rem 1rem",
                          fontSize: "0.8rem",
                          color: "#525252",
                        }}
                      >
                        <UploaderName userId={template.created_by} />
                      </td>

                      <td
                        style={{
                          padding: "0.6rem 1rem",
                          fontSize: "0.8rem",
                          color: "#525252",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {formatDateTime(template.created_at)}
                      </td>

                      <td
                        style={{
                          padding: "0.6rem 1rem",
                        }}
                      >
                        <StatusTag template={template} />
                      </td>

                      <td
                        style={{
                          padding: "0.6rem 1rem",
                          color: "#525252",
                        }}
                      >
                        v{template.version}
                      </td>

                      <td
                        style={{
                          padding: "0.4rem 0.75rem",
                          textAlign: "right",
                        }}
                      >
                        <div
                          style={{
                            display: "flex",
                            gap: "0.125rem",
                            justifyContent: "flex-end",
                            alignItems: "center",
                          }}
                        >
                          {status === "draft" && (
                            <PermissionGuard permission={PERMISSIONS.documentTemplatesPublish}>
                              <IconButton
                                label="Publish"
                                kind="ghost"
                                size="sm"
                                onClick={() => void handlePublish(template)}
                                disabled={isMutating}
                              >
                                <TaskComplete />
                              </IconButton>
                            </PermissionGuard>
                          )}

                          {status === "active" && (
                            <PermissionGuard permission={PERMISSIONS.documentTemplatesWrite}>
                              <IconButton
                                label="Archive"
                                kind="ghost"
                                size="sm"
                                onClick={() => void handleArchive(template)}
                                disabled={isMutating}
                              >
                                <Archive />
                              </IconButton>
                            </PermissionGuard>
                          )}

                          {status === "archived" && (
                            <PermissionGuard permission={PERMISSIONS.documentTemplatesPublish}>
                              <IconButton
                                label="Re-publish"
                                kind="ghost"
                                size="sm"
                                onClick={() => void handlePublish(template)}
                                disabled={isMutating}
                              >
                                <Renew />
                              </IconButton>
                            </PermissionGuard>
                          )}

                          <PermissionGuard permission={PERMISSIONS.documentTemplatesRead}>
                            <IconButton
                              label="Download empty template"
                              kind="ghost"
                              size="sm"
                              onClick={() => void handleDownload(template)}
                              disabled={isMutating}
                            >
                              <Download />
                            </IconButton>
                          </PermissionGuard>

                          <PermissionGuard permission={PERMISSIONS.documentTemplatesWrite}>
                            <IconButton
                              label="Edit"
                              kind="ghost"
                              size="sm"
                              onClick={() => handleEdit(template)}
                              disabled={isMutating}
                            >
                              <Edit />
                            </IconButton>
                          </PermissionGuard>

                            <PermissionGuard permission={PERMISSIONS.documentTemplatesWrite}>
                              <IconButton
                                label="Delete"
                                kind="ghost"
                                size="sm"
                                onClick={() => handleDeleteRequest(template)}
                                disabled={isMutating}
                              >
                              <TrashCan />
                            </IconButton>
                          </PermissionGuard>
                        </div>
                      </td>
                    </tr>

                    {isExpanded && (
                      <tr
                        style={{
                          borderBottom: "1px solid #e0e0e0",
                        }}
                      >
                        <td />

                        <td colSpan={8} style={{ padding: 0 }}>
                          <TemplateStructureView code={template.code} />
                        </td>
                      </tr>
                    )}
                  </Fragment>
                );
              })}
            </tbody>
            </table>
          </div>
        )}
      </>
    </PermissionGuard>
  );
}
