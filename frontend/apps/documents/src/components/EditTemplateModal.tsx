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
import { TrashCan } from "@carbon/icons-react";
import {
  useGetTemplateStructureQuery,
  useUpdateTemplateMutation,
  useUpdateColumnMutation,
  useReplaceStructureMutation,
  useScanDocumentStructureMutation,
} from "../api";
import type {
  DocumentTemplate,
  TemplateColumnStructure,
} from "../types";

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
    .replace(/[\s\-\/]+/g, "_")
    .replace(/[^a-z0-9_]/g, "")
    .replace(/_+/g, "_")
    .replace(/^_|_$/g, "");
}

// ── Edit tab ──────────────────────────────────────────────────────────────────

function EditTab({ template, onClose }: EditTemplateModalProps) {
  const [name, setName] = useState(template.name);
  const [description, setDescription] = useState(template.description ?? "");
  const [error, setError] = useState<string | null>(null);

  const { data: structure, isLoading: isLoadingStructure } = useGetTemplateStructureQuery(template.code);
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
    if (!name.trim()) { setError("Template name is required."); return; }
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
        <InlineNotification kind="error" title="Error" subtitle={error} lowContrast onCloseButtonClick={() => setError(null)} />
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
        <p style={{ margin: "0 0 0.75rem", fontWeight: 600, fontSize: "0.9375rem" }}>Column structure</p>

        {isLoadingStructure && <InlineLoading description="Loading columns..." />}

        {!isLoadingStructure && !hasSheets && (
          <p style={{ color: "#6f6f6f", fontSize: "0.875rem", margin: 0 }}>No columns defined for this template.</p>
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
                    <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "25%", verticalAlign: "middle" }}>Column name</th>
                    <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "27%", verticalAlign: "middle" }}>
                      Column key <span style={{ fontWeight: 400, color: "#6f6f6f", fontSize: "0.75rem" }}>(editable)</span>
                    </th>
                    <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "22%", verticalAlign: "middle" }}>Data type</th>
                    <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "13%", verticalAlign: "middle" }}>Required</th>
                    <th style={{ textAlign: "left", padding: "0.4rem 0.75rem", fontWeight: 600, width: "13%", verticalAlign: "middle" }}>Filterable</th>
                  </tr>
                </thead>
                <tbody>
                  {sheet.columns.slice().sort((a, b) => (a.column_order ?? 0) - (b.column_order ?? 0)).map((col) => {
                    const base = allColumns.find((c) => c.id === col.id)!;
                    const draft = getDraft(base);
                    return (
                      <tr key={col.id} style={{ borderBottom: "1px solid #f4f4f4", background: draft.required ? "#f0f7ff" : undefined }}>
                        <td style={{ padding: "0.4rem 0.75rem", fontWeight: draft.required ? 600 : 400, verticalAlign: "middle" }}>{col.column_name}</td>
                        <td style={{ padding: "0.25rem 0.75rem", verticalAlign: "middle" }}>
                          <TextInput id={`key-${col.id}`} labelText="" hideLabel size="sm"
                            value={draft.column_key ?? ""}
                            onChange={(e) => updateDraft(col.id, { column_key: e.target.value.trim().toLowerCase().replace(/[^a-z0-9_]/g, "") })}
                            disabled={isSaving} style={{ fontFamily: "monospace" }} />
                        </td>
                        <td style={{ padding: "0.25rem 0.75rem", verticalAlign: "middle" }}>
                          <Select id={`dtype-${col.id}`} labelText="" hideLabel size="sm"
                            value={draft.data_type} onChange={(e) => updateDraft(col.id, { data_type: e.target.value })} disabled={isSaving}>
                            {DATA_TYPE_OPTIONS.map((opt) => <SelectItem key={opt.value} value={opt.value} text={opt.label} />)}
                          </Select>
                        </td>
                        <td style={{ padding: "0.4rem 0.75rem", verticalAlign: "middle" }}>
                          <Toggle id={`req-${col.id}`} labelText="Required" hideLabel size="sm"
                            toggled={draft.required} onToggle={(checked) => updateDraft(col.id, { required: checked })} disabled={isSaving} />
                        </td>
                        <td style={{ padding: "0.4rem 0.75rem", verticalAlign: "middle" }}>
                          <Toggle id={`filter-${col.id}`} labelText="Filterable" hideLabel size="sm"
                            toggled={!!(draft.configuration?.filterable)}
                            onToggle={(checked) => updateDraft(col.id, { configuration: { ...(draft.configuration ?? {}), filterable: checked } })}
                            disabled={isSaving} />
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

      <div style={{ display: "flex", justifyContent: "flex-end", alignItems: "center", gap: "1rem" }}>
        {isSaving && <InlineLoading description="Saving changes..." />}
        <Button kind="secondary" onClick={onClose} disabled={isSaving}>Cancel</Button>
        <Button onClick={() => void handleSave()} disabled={isSaving || !name.trim() || isLoadingStructure}>
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
  const requiredCount = includedSheets.reduce((sum, s) => sum + s.columns.filter((c) => c.required).length, 0);

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
          setError("No usable sheets found. Make sure visible sheets have at least 4 column headers.");
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
    setSheets((prev) => prev.map((s, i) => i !== si ? s : { ...s, excluded: !s.excluded }));
  }, []);

  const updateColumn = useCallback(
    (si: number, ci: number, patch: Partial<NewColumnDraft>) => {
      setSheets((prev) =>
        prev.map((sheet, s) =>
          s !== si ? sheet : {
            ...sheet,
            columns: sheet.columns.map((col, c) => c !== ci ? col : { ...col, ...patch }),
          },
        ),
      );
    },
    [],
  );

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
        <InlineNotification kind="error" title="Error" subtitle={error} lowContrast onCloseButtonClick={() => setError(null)} />
      )}

      <Form>
        <Stack gap={5}>
          <div>
            <p style={{ margin: "0 0 0.5rem", fontWeight: 600 }}>Upload new template file</p>
            <FileUploaderDropContainer
              labelText="Drag and drop an Excel file (.xlsx / .xls) here, or click to browse"
              accept={[".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isProcessing}
            />
          </div>

          {file && !isScanning && (
            <div style={{ display: "flex", alignItems: "center", gap: "0.75rem", flexWrap: "wrap" }}>
              <Tag type="cyan">New file</Tag>
              <span style={{ fontWeight: 500 }}>{file.name}</span>
              <Button kind="ghost" size="sm" renderIcon={TrashCan} iconDescription="Remove" onClick={() => { setFile(null); setSheets([]); }} disabled={isProcessing}>
                Remove
              </Button>
            </div>
          )}

          {isScanning && <InlineLoading description="Detecting columns from file..." />}

          {!isScanning && sheets.length > 0 && (
            <div>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "baseline", marginBottom: "0.75rem" }}>
                <p style={{ margin: 0, fontWeight: 600 }}>
                  New columns —{" "}
                  <span style={{ color: "#6f6f6f", fontWeight: 400 }}>review before replacing</span>
                </p>
                <span style={{ color: "#6f6f6f", fontSize: "0.875rem" }}>{requiredCount} of {totalColumns} required</span>
              </div>

              {sheets.map((sheet, si) => (
                <div key={sheet.name} style={{ marginBottom: "1.5rem", opacity: sheet.excluded ? 0.6 : 1 }}>
                  {sheets.length > 1 && (
                    <div style={{ display: "flex", alignItems: "center", gap: "0.75rem", marginBottom: "0.5rem", flexWrap: "wrap" }}>
                      <p style={{ margin: 0, fontSize: "0.875rem", color: sheet.excluded ? "#a8a8a8" : "#6f6f6f", fontWeight: 600 }}>
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
                  <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.875rem" }}>
                    <thead>
                      <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "25%", verticalAlign: "middle" }}>Column name</th>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "27%", verticalAlign: "middle" }}>
                          Column key <span style={{ fontWeight: 400, color: "#6f6f6f", fontSize: "0.75rem" }}>(editable)</span>
                        </th>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "20%", verticalAlign: "middle" }}>Data type</th>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "14%", verticalAlign: "middle" }}>Required</th>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "14%", verticalAlign: "middle" }}>Filterable</th>
                      </tr>
                    </thead>
                    <tbody>
                      {sheet.columns.map((col, ci) => (
                        <tr key={ci} style={{ borderBottom: "1px solid #f4f4f4", backgroundColor: col.required ? "#f0f7ff" : undefined }}>
                          <td style={{ padding: "0.5rem 0.75rem", fontWeight: col.required ? 600 : 400, verticalAlign: "middle" }}>{col.column_name}</td>
                          <td style={{ padding: "0.25rem 0.75rem", verticalAlign: "middle" }}>
                            <TextInput id={`nkey-${si}-${ci}`} labelText="" hideLabel size="sm"
                              value={col.column_key}
                              onChange={(e) => updateColumn(si, ci, { column_key: e.target.value.trim().toLowerCase().replace(/[^a-z0-9_]/g, "") })}
                              disabled={isProcessing} style={{ fontFamily: "monospace" }} />
                          </td>
                          <td style={{ padding: "0.25rem 0.75rem", verticalAlign: "middle" }}>
                            <Select id={`ndtype-${si}-${ci}`} labelText="" hideLabel size="sm"
                              value={col.data_type} onChange={(e) => updateColumn(si, ci, { data_type: e.target.value })} disabled={isProcessing}>
                              {DATA_TYPE_OPTIONS.map((opt) => <SelectItem key={opt.value} value={opt.value} text={opt.label} />)}
                            </Select>
                          </td>
                          <td style={{ padding: "0.5rem 0.75rem", verticalAlign: "middle" }}>
                            <Toggle id={`nreq-${si}-${ci}`} labelText="Required" hideLabel size="sm"
                              toggled={col.required} onToggle={(checked) => updateColumn(si, ci, { required: checked })} disabled={isProcessing} />
                          </td>
                          <td style={{ padding: "0.5rem 0.75rem", verticalAlign: "middle" }}>
                            <Toggle id={`nfilter-${si}-${ci}`} labelText="Filterable" hideLabel size="sm"
                              toggled={col.filterable} onToggle={(checked) => updateColumn(si, ci, { filterable: checked })} disabled={isProcessing} />
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

          <div style={{ display: "flex", justifyContent: "flex-end", alignItems: "center", gap: "1rem" }}>
            {isReplacing && <InlineLoading description="Replacing columns..." />}
            <Button kind="secondary" onClick={onClose} disabled={isProcessing}>Cancel</Button>
            <Button kind="danger" onClick={() => void handleReplace()} disabled={isProcessing || includedSheets.length === 0}>
              Replace Columns
            </Button>
          </div>
        </Stack>
      </Form>
    </Stack>
  );
}

// ── Component ─────────────────────────────────────────────────────────────────

export function EditTemplateModal({ template, onClose }: EditTemplateModalProps) {
  return (
    <div style={{ maxWidth: 760 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Edit Template</h2>
        <div style={{ display: "flex", gap: "0.5rem", alignItems: "center", flexWrap: "wrap" }}>
          <Tag type="cool-gray" style={{ margin: 0 }}>{template.code}</Tag>
          <span style={{ color: "#6f6f6f", fontSize: "0.875rem" }}>
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
          <TabPanel style={{ paddingInline: 0, paddingTop: "1.5rem" }}>
            <EditTab template={template} onClose={onClose} />
          </TabPanel>
          <TabPanel style={{ paddingInline: 0, paddingTop: "1.5rem" }}>
            <ReplaceTab template={template} onClose={onClose} />
          </TabPanel>
        </TabPanels>
      </Tabs>
    </div>
  );
}
