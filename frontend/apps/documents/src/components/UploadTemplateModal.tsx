import { useCallback, useMemo, useState, type SyntheticEvent } from "react";
import {
  Button,
  FileUploaderDropContainer,
  Form,
  FormGroup,
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
import { TrashCan } from "@carbon/icons-react";

import { useSelector } from "react-redux";

import {
  useCreateDocumentMutation,
  useListStorageLocationsQuery,
  useLazyParseDocumentStructureQuery,
  useCreateTemplateStructureMutation,
} from "../api";
import { selectUser } from "@moh-sso/auth";
import { formatFileSize } from "@moh-sso/utils";

// ── Types ─────────────────────────────────────────────────────────────────────

type ColumnDraft = {
  column_key: string;
  column_name: string;
  data_type: string;
  required: boolean;
  filterable: boolean;
  column_order: number;
};

type SheetDraft = {
  name: string;
  excluded: boolean;
  headerRow: number;
  startRow: number;
  columns: ColumnDraft[];
};

export type UploadTemplateModalProps = {
  onClose: () => void;
};

// ── Constants ─────────────────────────────────────────────────────────────────

const DATA_TYPE_OPTIONS = [
  { value: "STRING", label: "Text" },
  { value: "INTEGER", label: "Integer" },
  { value: "DECIMAL", label: "Decimal" },
  { value: "DATE", label: "Date" },
  { value: "DATETIME", label: "Date & Time" },
  { value: "BOOLEAN", label: "Boolean" },
];

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

// ── Component ─────────────────────────────────────────────────────────────────

export const UploadTemplateModal: React.FC<UploadTemplateModalProps> = ({ onClose }) => {
  const currentUser = useSelector(selectUser);

  const [file, setFile] = useState<File | null>(null);
  const [sheets, setSheets] = useState<SheetDraft[]>([]);
  const [templateName, setTemplateName] = useState("");
  const [templateDescription, setTemplateDescription] = useState("");
  const [storageLocation, setStorageLocation] = useState("");
  const [preUploadedDocId, setPreUploadedDocId] = useState<string | null>(null);
  const [isDetecting, setIsDetecting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [createDocument, { isLoading: isUploadingFile }] = useCreateDocumentMutation();
  const [createTemplateStructure, { isLoading: isCreatingStructure }] =
    useCreateTemplateStructureMutation();
  const [triggerParseStructure] = useLazyParseDocumentStructureQuery();

  const {
    data: locations = [],
    isLoading: isLocationsLoading,
  } = useListStorageLocationsQuery();

  const activeLocations = useMemo(() => locations.filter((l) => l.is_active), [locations]);
  const localLocation = useMemo(
    () => activeLocations.find((l) => l.provider === "local"),
    [activeLocations],
  );
  if (localLocation && !storageLocation) {
    setStorageLocation(localLocation.id);
  }

  const templateCode = useMemo(() => normalizeKey(templateName).toUpperCase(), [templateName]);
  const isSubmitting = isUploadingFile || isCreatingStructure || isDetecting;

  const includedSheets = sheets.filter((s) => !s.excluded);
  const totalColumns = includedSheets.reduce((sum, s) => sum + s.columns.length, 0);
  const requiredCount = includedSheets.reduce(
    (sum, s) => sum + s.columns.filter((c) => c.required).length,
    0,
  );

  // ── File selection → upload + detect structure ────────────────────────────

  const handleFileChange = useCallback(
    async (_event: SyntheticEvent<HTMLElement, Event>, { addedFiles }: { addedFiles: File[] }) => {
      const selected = addedFiles[0];
      if (!selected) return;

      const ext = getFileExtension(selected.name);
      if (![".xlsx", ".xls"].includes(ext)) {
        setError("Templates must be Excel files (.xlsx or .xls).");
        return;
      }
      if (!storageLocation) {
        setError("Storage location unavailable.");
        return;
      }

      setError(null);
      setFile(selected);
      setSheets([]);
      setTemplateName("");
      setTemplateDescription("");
      setPreUploadedDocId(null);
      setIsDetecting(true);

      try {
        // Upload first — server detects structure; avoids parsing large files in browser
        const uploaded = await createDocument({
          file: selected,
          storageLocation,
          isTemplate: true,
        }).unwrap();

        setPreUploadedDocId(uploaded.id);

        const result = await triggerParseStructure(uploaded.id).unwrap();

        if (result.sheets.length === 0) {
          setError("No usable sheets found. Make sure visible sheets have at least 4 column headers.");
          setFile(null);
          setPreUploadedDocId(null);
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
        setError("Failed to read the file. Please check it is a valid Excel file and try again.");
        setFile(null);
        setPreUploadedDocId(null);
      } finally {
        setIsDetecting(false);
      }
    },
    [storageLocation, createDocument, triggerParseStructure],
  );

  const handleClearFile = useCallback(() => {
    setFile(null);
    setSheets([]);
    setPreUploadedDocId(null);
    setError(null);
  }, []);

  // ── Column editor ─────────────────────────────────────────────────────────

  const toggleSheetExcluded = useCallback((sheetIndex: number) => {
    setSheets((prev) =>
      prev.map((s, i) => (i !== sheetIndex ? s : { ...s, excluded: !s.excluded })),
    );
  }, []);

  const updateColumn = useCallback(
    (sheetIndex: number, colIndex: number, patch: Partial<ColumnDraft>) => {
      setSheets((prev) =>
        prev.map((sheet, si) =>
          si !== sheetIndex
            ? sheet
            : {
                ...sheet,
                columns: sheet.columns.map((col, ci) =>
                  ci !== colIndex ? col : { ...col, ...patch },
                ),
              },
        ),
      );
    },
    [],
  );

  // ── Submit ────────────────────────────────────────────────────────────────

  const handleUpload = async () => {
    if (!file || !preUploadedDocId) return setError("Please select a file.");
    if (!storageLocation) return setError("Storage location unavailable.");
    if (!templateName.trim()) return setError("Please enter a template name.");
    const includedForSubmit = sheets.filter((s) => !s.excluded);
    if (includedForSubmit.length === 0) return setError("No sheets are included. Toggle at least one sheet on.");

    setError(null);

    try {
      await createTemplateStructure({
        template: {
          document_id: preUploadedDocId,
          code: templateCode,
          name: templateName.trim(),
          description: templateDescription.trim(),
          file_type: getFileExtension(file.name).replace(".", ""),
          created_by: currentUser?.id,
          configuration: {
            allowed_extensions: [getFileExtension(file.name)],
          },
        },
        sheets: includedForSubmit.map((sheet, si) => ({
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
    } catch (err: unknown) {
      const errData =
        typeof err === "object" && err !== null && "data" in err
          ? (err as { data?: { message?: string; code?: string } }).data
          : null;

      let message = errData?.message ?? "Upload failed. Please try again.";

      if (errData?.code === "TEMPLATE_CODE_EXISTS") {
        message = `A template with code "${templateCode}" already exists. Try a different name, or edit the existing template from the Templates list.`;
      }

      setError(message);
    }
  };

  const isSubmitDisabled =
    isSubmitting ||
    !file ||
    !preUploadedDocId ||
    !storageLocation ||
    !templateName.trim() ||
    includedSheets.length === 0;

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div style={{ maxWidth: 760 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Template</h2>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload an Excel file to define a reusable template. Set which columns are required —
          uploads against this template will be rejected if required columns are missing.
        </p>
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

          {/* ── File drop ── */}
          <FormGroup legendText="Template file">
            <FileUploaderDropContainer
              labelText="Drag and drop an Excel file (.xlsx / .xls) here, or click to browse"
              accept={[".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isSubmitting}
            />
          </FormGroup>

          {/* ── Template name & description (shown before columns so user always sees them) ── */}
          <div>
            <TextInput
              id="template-name"
              labelText="Template name"
              helperText='Give this template a unique name, e.g. "GHSC PSM" or "NMS Stock Report".'
              placeholder="e.g. GHSC PSM"
              value={templateName}
              onChange={(e) => {
                setTemplateName(e.target.value);
                setError(null);
              }}
              disabled={isSubmitting}
            />
            {templateName.trim() && (
              <div style={{ marginTop: "0.4rem", display: "flex", alignItems: "center", gap: "0.5rem" }}>
                <span style={{ fontSize: "0.75rem", color: "#6f6f6f" }}>Code:</span>
                <code
                  style={{
                    fontSize: "0.8125rem",
                    fontFamily: "monospace",
                    backgroundColor: "#f4f4f4",
                    padding: "0.1rem 0.5rem",
                    borderRadius: "3px",
                    color: "#161616",
                  }}
                >
                  {templateCode}
                </code>
              </div>
            )}
          </div>

          <TextArea
            id="template-description"
            labelText="Description (optional)"
            placeholder="Describe what this template is used for"
            value={templateDescription}
            onChange={(e) => setTemplateDescription(e.target.value)}
            disabled={isSubmitting}
          />

          {file && (
            <div style={{ display: "flex", alignItems: "center", gap: "0.75rem", flexWrap: "wrap" }}>
              <Tag type="cyan">Template file</Tag>
              <span style={{ fontWeight: 500 }}>{file.name}</span>
              <span style={{ color: "#6f6f6f" }}>{formatFileSize(file.size)}</span>
              {!isDetecting && (
                <Button
                  kind="ghost"
                  size="sm"
                  renderIcon={TrashCan}
                  iconDescription="Remove file"
                  onClick={handleClearFile}
                  disabled={isSubmitting}
                >
                  Remove
                </Button>
              )}
            </div>
          )}

          {isDetecting && (
            <InlineLoading
              description={
                isUploadingFile
                  ? "Uploading file…"
                  : "Detecting columns from file…"
              }
            />
          )}

          {/* ── Column editor ── */}
          {!isDetecting && sheets.length > 0 && (
            <div>
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "baseline",
                  marginBottom: "0.75rem",
                }}
              >
                <p style={{ margin: 0, fontWeight: 600 }}>
                  Columns detected —{" "}
                  <span style={{ color: "#6f6f6f", fontWeight: 400 }}>
                    toggle <strong>Required</strong> for columns that must not be empty when data is
                    uploaded
                  </span>
                </p>
                <span style={{ color: "#6f6f6f", fontSize: "0.875rem", whiteSpace: "nowrap" }}>
                  {requiredCount} of {totalColumns} required
                </span>
              </div>

              {sheets.map((sheet, si) => (
                <div key={sheet.name} style={{ marginBottom: "1.5rem", opacity: sheet.excluded ? 0.6 : 1 }}>
                  {sheets.length > 1 && (
                    <div style={{ display: "flex", alignItems: "center", gap: "0.75rem", marginBottom: "0.5rem", flexWrap: "wrap" }}>
                      <p style={{ margin: 0, fontSize: "0.875rem", color: sheet.excluded ? "#a8a8a8" : "#6f6f6f", fontWeight: 600 }}>
                        Sheet: {sheet.name}
                      </p>
                      <Toggle
                        id={`sheet-include-${si}`}
                        labelText="Include sheet"
                        labelA="Excluded"
                        labelB="Included"
                        size="sm"
                        toggled={!sheet.excluded}
                        onToggle={() => toggleSheetExcluded(si)}
                        disabled={isSubmitting}
                      />
                    </div>
                  )}

                  {!sheet.excluded && (
                    <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "0.875rem" }}>
                      <thead>
                        <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
                          <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "25%", verticalAlign: "middle" }}>
                            Column name
                          </th>
                          <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "30%", verticalAlign: "middle" }}>
                            Column key
                            <span style={{ fontWeight: 400, color: "#6f6f6f", fontSize: "0.75rem", marginLeft: "0.35rem" }}>
                              (editable)
                            </span>
                          </th>
                          <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "20%", verticalAlign: "middle" }}>
                            Data type
                          </th>
                          <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "12%", verticalAlign: "middle" }}>
                            Required
                          </th>
                          <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "13%", verticalAlign: "middle" }}>
                            Filterable
                          </th>
                        </tr>
                      </thead>
                      <tbody>
                        {sheet.columns.map((col, ci) => (
                          <tr
                            key={`${si}-${ci}`}
                            style={{
                              borderBottom: "1px solid #f4f4f4",
                              backgroundColor: col.required ? "#f0f7ff" : undefined,
                            }}
                          >
                            <td style={{ padding: "0.5rem 0.75rem", fontWeight: col.required ? 600 : 400, verticalAlign: "middle" }}>
                              {col.column_name}
                            </td>
                            <td style={{ padding: "0.25rem 0.75rem", verticalAlign: "middle" }}>
                              <TextInput
                                id={`key-${si}-${ci}`}
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
                                disabled={isSubmitting}
                                style={{ fontFamily: "monospace" }}
                              />
                            </td>
                            <td style={{ padding: "0.25rem 0.75rem", verticalAlign: "middle" }}>
                              <Select
                                id={`dtype-${si}-${ci}`}
                                labelText=""
                                hideLabel
                                size="sm"
                                value={col.data_type}
                                onChange={(e) =>
                                  updateColumn(si, ci, { data_type: e.target.value })
                                }
                                disabled={isSubmitting}
                              >
                                {DATA_TYPE_OPTIONS.map((opt) => (
                                  <SelectItem key={opt.value} value={opt.value} text={opt.label} />
                                ))}
                              </Select>
                            </td>
                            <td style={{ padding: "0.5rem 0.75rem", verticalAlign: "middle" }}>
                              <Toggle
                                id={`req-${si}-${ci}`}
                                labelText="Required"
                                hideLabel
                                size="sm"
                                toggled={col.required}
                                onToggle={(checked) => updateColumn(si, ci, { required: checked })}
                                disabled={isSubmitting}
                              />
                            </td>
                            <td style={{ padding: "0.5rem 0.75rem", verticalAlign: "middle" }}>
                              <Toggle
                                id={`filter-${si}-${ci}`}
                                labelText="Filterable"
                                hideLabel
                                size="sm"
                                toggled={col.filterable}
                                onToggle={(checked) => updateColumn(si, ci, { filterable: checked })}
                                disabled={isSubmitting}
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

          {/* ── Storage location (locked to local) ── */}
          <Select
            id="storage-location"
            labelText="Storage location"
            value={storageLocation}
            onChange={() => {}}
            disabled
          >
            {localLocation ? (
              <SelectItem value={localLocation.id} text={localLocation.name} />
            ) : (
              <SelectItem value="" text={isLocationsLoading ? "Loading…" : "No local storage"} />
            )}
          </Select>

          {/* ── Actions ── */}
          <div
            style={{
              display: "flex",
              justifyContent: "flex-end",
              alignItems: "center",
              gap: "1rem",
              marginTop: "0.5rem",
            }}
          >
            {isCreatingStructure && (
              <InlineLoading description="Saving template structure…" />
            )}
            <Button kind="secondary" onClick={onClose} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button onClick={() => void handleUpload()} disabled={isSubmitDisabled}>
              Save Template
            </Button>
          </div>
        </Stack>
      </Form>
    </div>
  );
};
