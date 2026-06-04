import { useCallback, useMemo, useState } from "react";
import {
  Button,
  Form,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Stack,
  Tag,
  TextArea,
  TextInput,
  Toggle,
} from "@carbon/react";

import {
  useGetTemplateStructureQuery,
  useUpdateTemplateMutation,
  useUpdateColumnMutation,
} from "../../../../store/api/document_template.api";
import type {
  DocumentTemplate,
  TemplateColumnStructure,
} from "../../../../store/types/document_template.types";

// ── Constants ─────────────────────────────────────────────────────────────────

const DATA_TYPE_OPTIONS = [
  { value: "STRING", label: "Text" },
  { value: "INTEGER", label: "Integer" },
  { value: "DECIMAL", label: "Decimal" },
  { value: "DATE", label: "Date" },
  { value: "DATETIME", label: "Date & Time" },
  { value: "BOOLEAN", label: "Boolean" },
];

// ── Types ─────────────────────────────────────────────────────────────────────

type ColumnDraft = TemplateColumnStructure & { sheetId: string; templateId: string };

export type EditTemplateModalProps = {
  template: DocumentTemplate;
  onClose: () => void;
};

// ── Component ─────────────────────────────────────────────────────────────────

export function EditTemplateModal({ template, onClose }: EditTemplateModalProps) {
  const [name, setName] = useState(template.name);
  const [description, setDescription] = useState(template.description ?? "");
  const [error, setError] = useState<string | null>(null);

  const { data: structure, isLoading: isLoadingStructure } = useGetTemplateStructureQuery(
    template.code,
  );

  const [columnDrafts, setColumnDrafts] = useState<Record<string, ColumnDraft>>({});

  const allColumns: ColumnDraft[] = useMemo(() => {
    if (!structure) return [];
    return structure.sheets.flatMap((sheet) =>
      sheet.columns
        .slice()
        .sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0))
        .map((col) => ({
          ...col,
          sheetId: sheet.id,
          templateId: template.id,
        })),
    );
  }, [structure, template.id]);

  function getDraft(col: ColumnDraft): ColumnDraft {
    return columnDrafts[col.id] ?? col;
  }

  const updateDraft = useCallback((colId: string, patch: Partial<ColumnDraft>) => {
    setColumnDrafts((prev) => ({
      ...prev,
      [colId]: { ...(prev[colId] ?? allColumns.find((c) => c.id === colId)!), ...patch },
    }));
  }, [allColumns]);

  const [updateTemplate, { isLoading: isSavingMeta }] = useUpdateTemplateMutation();
  const [updateColumn, { isLoading: isSavingColumn }] = useUpdateColumnMutation();
  const isSaving = isSavingMeta || isSavingColumn;

  async function handleSave() {
    if (!name.trim()) {
      setError("Template name is required.");
      return;
    }
    setError(null);

    try {
      await updateTemplate({
        id: template.id,
        name: name.trim(),
        description: description.trim(),
        file_type: template.file_type,
        is_active: template.is_active,
        configuration: template.configuration ?? {},
      }).unwrap();

      const changedCols = Object.values(columnDrafts);
      await Promise.all(
        changedCols.map((col) =>
          updateColumn({
            templateId: col.templateId,
            sheetId: col.sheetId,
            columnId: col.id,
            ...(col.column_key ? { column_key: col.column_key } : {}),
            column_name: col.column_name,
            display_name: col.display_name ?? col.column_name,
            data_type: col.data_type,
            required: col.required,
            is_unique: col.is_unique,
            allowed_values: col.allowed_values ?? [],
            aliases: col.aliases ?? [],
            configuration: col.configuration ?? {},
          }).unwrap(),
        ),
      );

      onClose();
    } catch {
      setError("Failed to save changes. Please try again.");
    }
  }

  const hasSheets = structure && structure.sheets.length > 0;
  const multiSheet = hasSheets && structure.sheets.length > 1;

  return (
    <div style={{ maxWidth: 720 }}>
      {/* Header */}
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Edit Template</h2>
        <div style={{ display: "flex", gap: "0.5rem", alignItems: "center", flexWrap: "wrap" }}>
          <Tag type="cool-gray" style={{ margin: 0 }}>{template.code}</Tag>
          <span style={{ color: "#6f6f6f", fontSize: "0.875rem" }}>
            {template.file_type.toUpperCase()} · v{template.version}
          </span>
        </div>
      </div>

      <Form>
        <Stack gap={6}>
          {error && (
            <InlineNotification
              kind="error"
              title="Error"
              subtitle={error}
              lowContrast
              onCloseButtonClick={() => setError(null)}
            />
          )}

          {/* ── Metadata ── */}
          <TextInput
            id="edit-template-name"
            labelText="Template name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={isSaving}
            invalid={!name.trim()}
            invalidText="Name is required"
          />

          <TextArea
            id="edit-template-description"
            labelText="Description (optional)"
            placeholder="Describe what this template is used for"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={isSaving}
            rows={3}
          />

          {/* ── Column structure ── */}
          <div>
            <p style={{ margin: "0 0 0.75rem", fontWeight: 600, fontSize: "0.9375rem" }}>
              Column structure
            </p>

            {isLoadingStructure && (
              <InlineLoading description="Loading columns..." />
            )}

            {!isLoadingStructure && !hasSheets && (
              <p style={{ color: "#6f6f6f", fontSize: "0.875rem", margin: 0 }}>
                No columns defined for this template.
              </p>
            )}

            {!isLoadingStructure && hasSheets && structure.sheets.map((sheet) => (
              <div key={sheet.id} style={{ marginBottom: "1.25rem" }}>
                {multiSheet && (
                  <p style={{ margin: "0 0 0.5rem", fontSize: "0.875rem", color: "#525252", fontWeight: 600 }}>
                    Sheet: {sheet.name}
                  </p>
                )}

                {sheet.columns.length === 0 ? (
                  <p style={{ color: "#6f6f6f", fontSize: "0.875rem", margin: 0 }}>No columns.</p>
                ) : (
                  <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.875rem" }}>
                    <thead>
                      <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
                        <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "28%" }}>
                          Column name
                        </th>
                        <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "27%" }}>
                          Column key
                          <span style={{ fontWeight: 400, color: "#6f6f6f", fontSize: "0.75rem", marginLeft: "0.35rem" }}>(editable)</span>
                        </th>
                        <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "25%" }}>
                          Data type
                        </th>
                        <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "20%" }}>
                          Required
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {sheet.columns
                        .slice()
                        .sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0))
                        .map((col) => {
                          const base = allColumns.find((c) => c.id === col.id)!;
                          const draft = getDraft(base);
                          return (
                            <tr
                              key={col.id}
                              style={{
                                borderBottom: "1px solid #f4f4f4",
                                background: draft.required ? "#f0f7ff" : undefined,
                              }}
                            >
                              <td style={{ padding: "0.4rem 0.75rem", fontWeight: draft.required ? 600 : 400 }}>
                                {col.column_name}
                              </td>
                              <td style={{ padding: "0.3rem 0.75rem" }}>
                                <TextInput
                                  id={`key-${col.id}`}
                                  labelText=""
                                  hideLabel
                                  size="sm"
                                  value={draft.column_key ?? ""}
                                  onChange={(e) =>
                                    updateDraft(col.id, {
                                      column_key: e.target.value
                                        .trim()
                                        .toLowerCase()
                                        .replace(/[^a-z0-9_]/g, ""),
                                    })
                                  }
                                  disabled={isSaving}
                                  style={{ fontFamily: "monospace" }}
                                />
                              </td>
                              <td style={{ padding: "0.3rem 0.75rem" }}>
                                <Select
                                  id={`dtype-${col.id}`}
                                  labelText=""
                                  hideLabel
                                  size="sm"
                                  value={draft.data_type}
                                  onChange={(e) =>
                                    updateDraft(col.id, { data_type: e.target.value })
                                  }
                                  disabled={isSaving}
                                >
                                  {DATA_TYPE_OPTIONS.map((opt) => (
                                    <SelectItem key={opt.value} value={opt.value} text={opt.label} />
                                  ))}
                                </Select>
                              </td>
                              <td style={{ padding: "0.3rem 0.75rem" }}>
                                <Toggle
                                  id={`req-${col.id}`}
                                  labelText="Required"
                                  hideLabel
                                  size="sm"
                                  toggled={draft.required}
                                  onToggle={(checked) =>
                                    updateDraft(col.id, { required: checked })
                                  }
                                  disabled={isSaving}
                                />
                              </td>
                            </tr>
                          );
                        })}
                    </tbody>
                  </table>
                )}
              </div>
            ))}
          </div>

          {/* ── Actions ── */}
          <div style={{ display: "flex", justifyContent: "flex-end", alignItems: "center", gap: "1rem" }}>
            {isSaving && <InlineLoading description="Saving changes..." />}
            <Button kind="secondary" onClick={onClose} disabled={isSaving}>
              Cancel
            </Button>
            <Button
              onClick={() => void handleSave()}
              disabled={isSaving || !name.trim() || isLoadingStructure}
            >
              Save Changes
            </Button>
          </div>
        </Stack>
      </Form>
    </div>
  );
}
