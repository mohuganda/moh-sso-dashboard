import { Fragment, useState } from "react";
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
import * as XLSX from "xlsx";

import { PermissionGuard, PERMISSIONS, selectUser } from "@moh-sso/auth";
import { useHeaderPanel } from "@moh-sso/ui";
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
    <div
      style={{
        padding: "1rem 1.5rem 1.25rem",
        background: "#f4f4f4",
      }}
    >
      {structure.sheets.map((sheet, sheetIndex) => (
        <div
          key={sheet.id}
          style={{
            marginBottom: sheetIndex === structure.sheets.length - 1 ? 0 : "1.25rem",
          }}
        >
          {structure.sheets.length > 1 && (
            <p
              style={{
                margin: "0 0 0.5rem",
                fontWeight: 600,
                fontSize: "0.875rem",
              }}
            >
              Sheet: {sheet.name}
            </p>
          )}

          {sheet.columns.length === 0 ? (
            <p
              style={{
                margin: 0,
                color: "#6f6f6f",
                fontSize: "0.875rem",
              }}
            >
              No columns defined.
            </p>
          ) : (
            <table
              style={{
                borderCollapse: "collapse",
                fontSize: "0.8125rem",
                width: "100%",
                maxWidth: 640,
              }}
            >
              <thead>
                <tr
                  style={{
                    borderBottom: "1px solid #c6c6c6",
                  }}
                >
                  <th
                    style={{
                      textAlign: "left",
                      padding: "0.35rem 0.75rem",
                      fontWeight: 600,
                      color: "#525252",
                    }}
                  >
                    Column
                  </th>

                  <th
                    style={{
                      textAlign: "left",
                      padding: "0.35rem 0.75rem",
                      fontWeight: 600,
                      color: "#525252",
                    }}
                  >
                    Type
                  </th>

                  <th
                    style={{
                      textAlign: "center",
                      padding: "0.35rem 0.75rem",
                      fontWeight: 600,
                      color: "#525252",
                    }}
                  >
                    Required
                  </th>
                </tr>
              </thead>

              <tbody>
                {sheet.columns
                  .slice()
                  .sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0))
                  .map((column) => (
                    <tr
                      key={column.id}
                      style={{
                        borderBottom: "1px solid #e8e8e8",
                        background: column.required ? "#fff" : undefined,
                      }}
                    >
                      <td
                        style={{
                          padding: "0.35rem 0.75rem",
                          fontFamily: "monospace",
                          fontWeight: column.required ? 600 : 400,
                          color: column.required ? "#161616" : "#525252",
                        }}
                      >
                        {column.column_name}
                      </td>

                      <td
                        style={{
                          padding: "0.35rem 0.75rem",
                          color: "#525252",
                        }}
                      >
                        {column.data_type}
                      </td>

                      <td
                        style={{
                          padding: "0.35rem 0.75rem",
                          textAlign: "center",
                        }}
                      >
                        {column.required ? (
                          <span
                            style={{
                              color: "#198038",
                              fontWeight: 700,
                            }}
                          >
                            ✓
                          </span>
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
          Are you sure you want to delete <strong>{template.name}</strong> ({template.code})? This
          cannot be undone.
        </p>
      )}
    </Modal>
  );
}

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
    const result = await getStructure(template.code);

    if (result.data) {
      downloadTemplateFile(template, result.data);
    }
  }

  async function handlePublish(template: DocumentTemplate) {
    try {
      await publishTemplate(template.id).unwrap();
    } catch {
      // Error is exposed through RTK Query state.
    }
  }

  async function handleArchive(template: DocumentTemplate) {
    try {
      await archiveTemplate(template.id).unwrap();
    } catch {
      // Error is exposed through RTK Query state.
    }
  }

  async function handleDeleteConfirm() {
    if (!deleteTarget) {
      return;
    }

    try {
      await deleteTemplate(deleteTarget.id).unwrap();
    } finally {
      setDeleteTarget(null);
    }
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
        <PermissionGuard permission={PERMISSIONS.documentTemplatesWrite}>
          <DeleteConfirmModal
            template={deleteTarget}
            onConfirm={() => void handleDeleteConfirm()}
            onCancel={() => setDeleteTarget(null)}
            isLoading={isDeleting}
          />
        </PermissionGuard>

        <div
          style={{
            marginBottom: "1rem",
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <p style={{ margin: 0, color: "#6f6f6f" }}>
            {templates.length} template
            {templates.length !== 1 ? "s" : ""} total —{" "}
            {templates.filter((template) => template.is_active && !template.archived_at).length}{" "}
            active
          </p>

          <PermissionGuard permission={PERMISSIONS.documentTemplatesWrite}>
            <Button renderIcon={Add} onClick={handleNewTemplate} size="sm">
              New Template
            </Button>
          </PermissionGuard>
        </div>

        <div style={{ marginBottom: "0.5rem" }}>
          <Search
            placeholder="Search by name, code, or file type"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
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
              ? "No templates are currently available."
              : "No templates match your search."}
          </div>
        ) : (
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: "0.875rem",
            }}
          >
            <thead>
              <tr
                style={{
                  borderBottom: "2px solid #e0e0e0",
                  background: "#f4f4f4",
                }}
              >
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
                              onClick={() => setDeleteTarget(template)}
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
        )}
      </>
    </PermissionGuard>
  );
}
