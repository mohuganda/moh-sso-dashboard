import { useCallback, useMemo, useState, type SyntheticEvent } from "react";
import {
  Button,
  FileUploaderDropContainer,
  Form,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Stack,
  Tab,
  TabList,
  TabPanel,
  TabPanels,
  Tabs,
  Tag,
  TextArea,
  TextInput,
  Toggle,
} from "@carbon/react";
import { TrashCan } from "@carbon/react/icons";
import {
  useGetTemplateStructureQuery,
  useUpdateTemplateMutation,
  useUpdateColumnMutation,
  useReplaceStructureMutation,
  useScanDocumentStructureMutation,
} from "../api";
import type { DocumentTemplate, TemplateColumnStructure } from "../types";
import { PermissionGuard, PERMISSIONS } from "@moh-sso/auth";
import "./documents-components.scss";

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

type SheetDraft = {
  name: string;
  excluded: boolean;
  headerRow: number;
  startRow: number;
  columns: NewColumnDraft[];
};

type NewColumnDraft = {
  column_key: string;
  column_name: string;
  data_type: string;
  required: boolean;
  filterable: boolean;
  column_order: number;
};

export type EditTemplateModalProps = {
  template: DocumentTemplate;
  onClose: () => void;
};

// ── Helpers ───────────────────────────────────────────────────────────────────

function getFileExtension(fileName: string): string {
  const idx = fileName.lastIndexOf(".");
  return idx >= 0 ? fileName.slice(idx).toLowerCase() : "";
}

function normalizeKey(value: string): string {
  return value
    .trim()
    .replace(/([a-z])([A-Z])/g, "$1_$2")
    .toLowerCase()
    .replace(/[\s\-/]+/g, "_")
    .replace(/[^a-z0-9_]/g, "")
    .replace(/_+/g, "_")
    .replace(/^_|_$/g, "");
}

// ── Edit tab ──────────────────────────────────────────────────────────────────

function EditTab({ template, onClose }: EditTemplateModalProps) {
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
        .map((col) => ({ ...col, sheetId: sheet.id, templateId: template.id })),
    );
  }, [structure, template.id]);

  function getDraft(col: ColumnDraft): ColumnDraft {
    return columnDrafts[col.id] ?? col;
  }

  const updateDraft = useCallback(
    (colId: string, patch: Partial<ColumnDraft>) => {
      setColumnDrafts((prev) => ({
        ...prev,
        [colId]: { ...(prev[colId] ?? allColumns.find((c) => c.id === colId)!), ...patch },
      }));
    },
    [allColumns],
  );

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

      <div>
        <p className="document-edit-template__section-title">Column structure</p>

        {isLoadingStructure && <InlineLoading description="Loading columns..." />}

        {!isLoadingStructure && !hasSheets && (
          <p className="document-edit-template__empty">No columns defined for this template.</p>
        )}

        {!isLoadingStructure &&
          hasSheets &&
          structure.sheets.map((sheet) => (
            <div key={sheet.id} className="document-edit-template__sheet">
              {multiSheet && (
                <p className="document-upload-modal__sheet-title">Sheet: {sheet.name}</p>
              )}
              {sheet.columns.length === 0 ? (
                <p className="document-edit-template__empty">No columns.</p>
              ) : (
                <table className="document-upload-modal__columns-table">
                  <thead>
                    <tr>
                      <th>Column name</th>
                      <th>
                        Column key{" "}
                        <span className="document-upload-modal__header-note">(editable)</span>
                      </th>
                      <th>Data type</th>
                      <th>Required</th>
                      <th>Filterable</th>
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
                            className={
                              draft.required ? "document-upload-modal__required-row" : undefined
                            }
                          >
                            <td
                              className={
                                draft.required ? "document-upload-modal__required-name" : undefined
                              }
                            >
                              {col.column_name}
                            </td>
                            <td className="document-upload-modal__compact-cell">
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
                                className="document-upload-modal__mono-input"
                              />
                            </td>
                            <td className="document-upload-modal__compact-cell">
                              <Select
                                id={`dtype-${col.id}`}
                                labelText=""
                                hideLabel
                                size="sm"
                                value={draft.data_type}
                                onChange={(e) => updateDraft(col.id, { data_type: e.target.value })}
                                disabled={isSaving}
                              >
                                {DATA_TYPE_OPTIONS.map((opt) => (
                                  <SelectItem key={opt.value} value={opt.value} text={opt.label} />
                                ))}
                              </Select>
                            </td>
                            <td>
                              <Toggle
                                id={`req-${col.id}`}
                                labelText="Required"
                                hideLabel
                                size="sm"
                                toggled={draft.required}
                                onToggle={(checked) => updateDraft(col.id, { required: checked })}
                                disabled={isSaving}
                              />
                            </td>
                            <td>
                              <Toggle
                                id={`filter-${col.id}`}
                                labelText="Filterable"
                                hideLabel
                                size="sm"
                                toggled={!!draft.configuration?.filterable}
                                onToggle={(checked) =>
                                  updateDraft(col.id, {
                                    configuration: {
                                      ...(draft.configuration ?? {}),
                                      filterable: checked,
                                    },
                                  })
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

      <div className="document-upload-modal__actions">
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
  );
}

// ── Replace tab ───────────────────────────────────────────────────────────────

function ReplaceTab({ template, onClose }: EditTemplateModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const [sheets, setSheets] = useState<SheetDraft[]>([]);
  const [error, setError] = useState<string | null>(null);

  const [scanStructure, { isLoading: isScanning }] = useScanDocumentStructureMutation();
  const [replaceStructure, { isLoading: isReplacing }] = useReplaceStructureMutation();

  const isProcessing = isScanning || isReplacing;
  const includedSheets = sheets.filter((s) => !s.excluded);
  const totalColumns = includedSheets.reduce((sum, s) => sum + s.columns.length, 0);
  const requiredCount = includedSheets.reduce(
    (sum, s) => sum + s.columns.filter((c) => c.required).length,
    0,
  );

  const handleFileChange = useCallback(
    async (_event: SyntheticEvent<HTMLElement, Event>, { addedFiles }: { addedFiles: File[] }) => {
      const selected = addedFiles[0];
      if (!selected) return;
      const ext = getFileExtension(selected.name);
      if (![".xlsx", ".xls"].includes(ext)) {
        setError("File must be an Excel file (.xlsx or .xls).");
        return;
      }
      setError(null);
      setFile(selected);
      setSheets([]);
      try {
        const result = await scanStructure(selected).unwrap();
        if (result.sheets.length === 0) {
          setError(
            "No usable sheets found. Make sure visible sheets have at least 4 column headers.",
          );
          setFile(null);
          return;
        }
        setSheets(
          result.sheets.map((s) => ({
            name: s.name,
            excluded: false,
            headerRow: s.header_row,
            startRow: s.start_row,
            columns: s.columns.map((col, i) => ({
              column_key: col.column_key,
              column_name: col.column_name,
              data_type: "STRING",
              required: false,
              filterable: false,
              column_order: i + 1,
            })),
          })),
        );
      } catch {
        setError("Failed to read file. Please check it is a valid Excel file and try again.");
        setFile(null);
      }
    },
    [scanStructure],
  );

  const toggleSheetExcluded = useCallback((si: number) => {
    setSheets((prev) => prev.map((s, i) => (i !== si ? s : { ...s, excluded: !s.excluded })));
  }, []);

  const updateColumn = useCallback((si: number, ci: number, patch: Partial<NewColumnDraft>) => {
    setSheets((prev) =>
      prev.map((sheet, s) =>
        s !== si
          ? sheet
          : {
              ...sheet,
              columns: sheet.columns.map((col, c) => (c !== ci ? col : { ...col, ...patch })),
            },
      ),
    );
  }, []);

  async function handleReplace() {
    if (!file || includedSheets.length === 0) return;
    setError(null);
    try {
      await replaceStructure({
        templateId: template.id,
        sheets: includedSheets.map((sheet, si) => ({
          sheet: {
            code: normalizeKey(sheet.name),
            name: sheet.name,
            display_name: sheet.name,
            required: true,
            sheet_order: si + 1,
            header_row: sheet.headerRow,
            start_row: sheet.startRow,
            allow_extra_columns: true,
            allow_duplicate_headers: false,
            configuration: {},
          },
          columns: sheet.columns.map((col) => ({
            column_key: col.column_key,
            column_name: col.column_name,
            display_name: col.column_name,
            data_type: col.data_type,
            required: col.required,
            is_unique: false,
            column_order: col.column_order,
            default_value: null,
            allowed_values: [],
            aliases: [],
            configuration: { filterable: col.filterable },
          })),
        })),
      }).unwrap();
      onClose();
    } catch {
      setError("Failed to replace columns. Please try again.");
    }
  }

  return (
    <Stack gap={6}>
      <InlineNotification
        kind="warning"
        title="This replaces all existing columns"
        subtitle="All current sheets and columns for this template will be deleted and replaced with the columns from the new file. Uploaded data rows already stored in the database are not affected."
        lowContrast
        hideCloseButton
      />

      {error && (
        <InlineNotification
          kind="error"
          title="Error"
          subtitle={error}
          lowContrast
          onCloseButtonClick={() => setError(null)}
        />
      )}

      <Form>
        <Stack gap={5}>
          <div>
            <p className="document-upload-modal__section-title">Upload new template file</p>
            <FileUploaderDropContainer
              labelText="Drag and drop an Excel file (.xlsx / .xls) here, or click to browse"
              accept={[".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isProcessing}
            />
          </div>

          {file && !isScanning && (
            <div className="document-upload-modal__file">
              <Tag type="cyan">New file</Tag>
              <span className="document-upload-modal__file-name">{file.name}</span>
              <Button
                kind="ghost"
                size="sm"
                renderIcon={TrashCan}
                iconDescription="Remove"
                onClick={() => {
                  setFile(null);
                  setSheets([]);
                }}
                disabled={isProcessing}
              >
                Remove
              </Button>
            </div>
          )}

          {isScanning && <InlineLoading description="Detecting columns from file..." />}

          {!isScanning && sheets.length > 0 && (
            <div>
              <div className="document-upload-modal__detected-header">
                <p className="document-upload-modal__detected-title">
                  New columns —{" "}
                  <span className="document-upload-modal__detected-note">
                    review before replacing
                  </span>
                </p>
                <span className="document-upload-modal__detected-count">
                  {requiredCount} of {totalColumns} required
                </span>
              </div>

              {sheets.map((sheet, si) => (
                <div
                  key={sheet.name}
                  className={[
                    "document-upload-modal__sheet",
                    sheet.excluded ? "document-upload-modal__sheet--excluded" : "",
                  ]
                    .filter(Boolean)
                    .join(" ")}
                >
                  {sheets.length > 1 && (
                    <div className="document-upload-modal__sheet-header">
                      <p
                        className={[
                          "document-upload-modal__sheet-title",
                          sheet.excluded ? "document-upload-modal__sheet-title--excluded" : "",
                        ]
                          .filter(Boolean)
                          .join(" ")}
                      >
                        Sheet: {sheet.name}
                      </p>
                      <Toggle
                        id={`rsheet-include-${si}`}
                        labelText="Include sheet"
                        labelA="Excluded"
                        labelB="Included"
                        size="sm"
                        toggled={!sheet.excluded}
                        onToggle={() => toggleSheetExcluded(si)}
                        disabled={isProcessing}
                      />
                    </div>
                  )}
                  {!sheet.excluded && (
                    <table className="document-upload-modal__columns-table">
                      <thead>
                        <tr>
                          <th>Column name</th>
                          <th>
                            Column key{" "}
                            <span className="document-upload-modal__header-note">(editable)</span>
                          </th>
                          <th>Data type</th>
                          <th>Required</th>
                          <th>Filterable</th>
                        </tr>
                      </thead>
                      <tbody>
                        {sheet.columns.map((col, ci) => (
                          <tr
                            key={ci}
                            className={
                              col.required ? "document-upload-modal__required-row" : undefined
                            }
                          >
                            <td
                              className={
                                col.required ? "document-upload-modal__required-name" : undefined
                              }
                            >
                              {col.column_name}
                            </td>
                            <td className="document-upload-modal__compact-cell">
                              <TextInput
                                id={`nkey-${si}-${ci}`}
                                labelText=""
                                hideLabel
                                size="sm"
                                value={col.column_key}
                                onChange={(e) =>
                                  updateColumn(si, ci, {
                                    column_key: e.target.value
                                      .trim()
                                      .toLowerCase()
                                      .replace(/[^a-z0-9_]/g, ""),
                                  })
                                }
                                disabled={isProcessing}
                                className="document-upload-modal__mono-input"
                              />
                            </td>
                            <td className="document-upload-modal__compact-cell">
                              <Select
                                id={`ndtype-${si}-${ci}`}
                                labelText=""
                                hideLabel
                                size="sm"
                                value={col.data_type}
                                onChange={(e) =>
                                  updateColumn(si, ci, { data_type: e.target.value })
                                }
                                disabled={isProcessing}
                              >
                                {DATA_TYPE_OPTIONS.map((opt) => (
                                  <SelectItem key={opt.value} value={opt.value} text={opt.label} />
                                ))}
                              </Select>
                            </td>
                            <td>
                              <Toggle
                                id={`nreq-${si}-${ci}`}
                                labelText="Required"
                                hideLabel
                                size="sm"
                                toggled={col.required}
                                onToggle={(checked) => updateColumn(si, ci, { required: checked })}
                                disabled={isProcessing}
                              />
                            </td>
                            <td>
                              <Toggle
                                id={`nfilter-${si}-${ci}`}
                                labelText="Filterable"
                                hideLabel
                                size="sm"
                                toggled={col.filterable}
                                onToggle={(checked) =>
                                  updateColumn(si, ci, { filterable: checked })
                                }
                                disabled={isProcessing}
                              />
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  )}
                </div>
              ))}
            </div>
          )}

          <div className="document-upload-modal__actions">
            {isReplacing && <InlineLoading description="Replacing columns..." />}
            <Button kind="secondary" onClick={onClose} disabled={isProcessing}>
              Cancel
            </Button>
            <Button
              kind="danger"
              onClick={() => void handleReplace()}
              disabled={isProcessing || includedSheets.length === 0}
            >
              Replace Columns
            </Button>
          </div>
        </Stack>
      </Form>
    </Stack>
  );
}

// ── Component ─────────────────────────────────────────────────────────────────

function EditTemplateModalContent({
  template,

  onClose,
}: EditTemplateModalProps) {
  return (
    <div className="document-edit-template">
      <div className="document-edit-template__header">
        <h2>Edit Template</h2>

        <div className="document-edit-template__metadata">
          <Tag type="cool-gray" className="document-edit-template__tag">
            {template.code}
          </Tag>

          <span className="document-edit-template__meta-text">
            {template.file_type.toUpperCase()} · v{template.version}
          </span>
        </div>
      </div>

      <Tabs>
        <TabList aria-label="Edit template options" contained>
          <Tab>Edit columns &amp; metadata</Tab>

          <Tab>Replace columns from file</Tab>
        </TabList>

        <TabPanels>
          <TabPanel className="document-edit-template__tab-panel">
            <EditTab template={template} onClose={onClose} />
          </TabPanel>

          <TabPanel className="document-edit-template__tab-panel">
            <ReplaceTab template={template} onClose={onClose} />
          </TabPanel>
        </TabPanels>
      </Tabs>
    </div>
  );
}

export function EditTemplateModal({ template, onClose }: EditTemplateModalProps) {
  return (
    <PermissionGuard
      permission={PERMISSIONS.documentTemplatesWrite}
      fallback={
        <div className="document-edit-template">
          <InlineNotification
            kind="warning"
            title="Access denied"
            subtitle="You do not have permission to edit document templates."
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
      <EditTemplateModalContent template={template} onClose={onClose} />
    </PermissionGuard>
  );
}
