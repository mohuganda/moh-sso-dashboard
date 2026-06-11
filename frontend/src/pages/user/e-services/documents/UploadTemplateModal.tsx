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
import * as XLSX from "xlsx";
import Papa from "papaparse";

import {
  useCreateDocumentMutation,
  useListStorageLocationsQuery,
} from "../../../../store/api/document.api";
import { useCreateTemplateStructureMutation } from "../../../../store/api/document_template.api";
import { formatFileSize } from "../../../../utils/utils";

// ── Types ─────────────────────────────────────────────────────────────────────

type ColumnDraft = {
  column_key: string;
  column_name: string;
  data_type: string;
  required: boolean;
  column_order: number;
};

type SheetDraft = {
  name: string;
  headerRow: number;
  columns: ColumnDraft[];
};

type ColumnDefaults = Partial<Pick<ColumnDraft, "column_key" | "data_type">>;

export type UploadTemplateModalProps = {
  onClose: () => void;
};

// ── Constants ─────────────────────────────────────────────────────────────────

const CSV_HEADER_READ_BYTES = 64 * 1024;

const DATA_TYPE_OPTIONS = [
  { value: "STRING", label: "Text" },
  { value: "INTEGER", label: "Integer" },
  { value: "DECIMAL", label: "Decimal" },
  { value: "DATE", label: "Date" },
  { value: "DATETIME", label: "Date & Time" },
  { value: "BOOLEAN", label: "Boolean" },
];

const IMPLEMENTED_TEMPLATES = [
  { code: "NMS_STOCK_REPORT", name: "NMS Stock Report" },
  { code: "JMS_STOCK_REPORT", name: "JMS Stock Report" },
  { code: "GF_PIPELINE", name: "GF Pipeline" },
  { code: "GDF_TB_ORDERS", name: "GDF TB Orders" },
  { code: "GHSC_PSM", name: "GHSC-PSM" },
  { code: "UNFPA", name: "UNFPA Pipeline" },
];

// ── Helpers ───────────────────────────────────────────────────────────────────

function getFileExtension(fileName: string): string {
  const idx = fileName.lastIndexOf(".");
  return idx >= 0 ? fileName.slice(idx).toLowerCase() : "";
}

function normalizeKey(value: string): string {
  return value
    .trim()
    .replace(/([a-z])([A-Z])/g, "$1_$2")   // CamelCase → snake_case
    .toLowerCase()
    .replace(/[\s\-\/]+/g, "_")             // spaces, hyphens, slashes → _
    .replace(/[^a-z0-9_]/g, "")            // strip remaining special chars
    .replace(/_+/g, "_")                   // collapse repeated underscores
    .replace(/^_|_$/g, "");               // trim leading/trailing underscores
}

function getGHSCPSMColumnDefaults(columnName: string): ColumnDefaults {
  const trimmed = columnName.trim();
  const key = normalizeKey(trimmed);

  if (trimmed === "#") return { column_key: "line_number", data_type: "INTEGER" };
  if (key === "commodity_category") return { column_key: "commodity_category" };
  if (key === "item_description") return { column_key: "item_description" };
  if (key === "uom" || key === "uo_m") return { column_key: "uom" };
  if (key === "stock_on_hand") return { column_key: "stock_on_hand", data_type: "DECIMAL" };
  if (key === "average_monthly_consumption_amc") return { column_key: "amc", data_type: "DECIMAL" };
  if (key.startsWith("mos_as_of_end_of_")) return { column_key: "mos", data_type: "DECIMAL" };
  if (key === "quantity_on_order") return { column_key: "quantity_on_order", data_type: "DECIMAL" };
  if (key === "requested_delivery_date") return { column_key: "requested_delivery_date", data_type: "DATE" };
  if (key === "estimated_delivery_date") return { column_key: "estimated_delivery_date", data_type: "DATE" };
  if (key === "status") return { column_key: "status" };
  if (key === "comment") return { column_key: "comment" };

  return {};
}

function applyTemplateDefaults(templateCode: string, sheets: SheetDraft[]): SheetDraft[] {
  if (templateCode !== "GHSC_PSM") return sheets;

  return sheets.map((sheet) => ({
    ...sheet,
    columns: sheet.columns.map((column) => ({
      ...column,
      ...getGHSCPSMColumnDefaults(column.column_name),
    })),
  }));
}

async function readSheetsFromFile(file: File): Promise<SheetDraft[]> {
  const ext = getFileExtension(file.name);

  if (ext === ".csv") {
    const text = await file.slice(0, CSV_HEADER_READ_BYTES).text();
    const headers = await new Promise<string[]>((resolve, reject) => {
      Papa.parse<Record<string, unknown>>(text, {
        header: true,
        skipEmptyLines: true,
        preview: 1,
        complete: (results) => {
          resolve((results.meta.fields ?? []).map((h) => String(h ?? "").trim()).filter(Boolean));
        },
        error: reject,
      });
    });

    return [
      {
        name: "CSV Data",
        headerRow: 1,
        columns: headers.map((name, i) => ({
          column_key: normalizeKey(name) || `col_${i + 1}`,
          column_name: name,
          data_type: "STRING",
          required: false,
          column_order: i + 1,
        })),
      },
    ];
  }

  if (ext === ".xlsx" || ext === ".xls") {
    const buffer = await file.arrayBuffer();
    const workbook = XLSX.read(buffer, { type: "array", sheetRows: 10 });

    const visibleNames = workbook.SheetNames.filter((_, idx) => {
      const meta = workbook.Workbook?.Sheets?.[idx];
      return !meta?.Hidden;
    });

    return visibleNames
      .map((sheetName) => {
        const ws = workbook.Sheets[sheetName];
        const rows = XLSX.utils.sheet_to_json<(string | number | null)[]>(ws, {
          header: 1,
          defval: "",
          blankrows: true,
        });

        // Find the first row with at least 2 non-empty cells — that is the header row.
        let headerRowIdx = 0;
        for (let i = 0; i < rows.length; i++) {
          const nonEmpty = rows[i].filter((cell) => String(cell ?? "").trim() !== "").length;
          if (nonEmpty >= 4) {
            headerRowIdx = i;
            break;
          }
        }

        const headerRow = Array.isArray(rows[headerRowIdx]) ? rows[headerRowIdx] : [];
        const rawHeaders = headerRow.map((h) => String(h ?? "").trim()).filter(Boolean);

        // Deduplicate column names and keys so the DB unique constraints aren't violated.
        // When a header appears more than once, suffix duplicates with _2, _3, …
        const nameCounts: Record<string, number> = {};
        const keyCounts: Record<string, number> = {};
        const headers = rawHeaders.map((name) => {
          const baseKey = normalizeKey(name) || `col`;
          nameCounts[name] = (nameCounts[name] ?? 0) + 1;
          keyCounts[baseKey] = (keyCounts[baseKey] ?? 0) + 1;
          const nc = nameCounts[name];
          const kc = keyCounts[baseKey];
          return {
            name: nc > 1 ? `${name}_${nc}` : name,
            key: kc > 1 ? `${baseKey}_${kc}` : baseKey,
          };
        });

        return {
          name: sheetName,
          headerRow: headerRowIdx + 1,
          columns: headers.map(({ name, key }, i) => ({
            column_key: key,
            column_name: name,
            data_type: "STRING",
            required: false,
            column_order: i + 1,
          })),
        };
      })
      .filter((s) => s.columns.length > 0);
  }

  return [];
}

// ── Component ─────────────────────────────────────────────────────────────────

export const UploadTemplateModal: React.FC<UploadTemplateModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [sheets, setSheets] = useState<SheetDraft[]>([]);
  const [selectedTemplateCode, setSelectedTemplateCode] = useState("");
  const [templateDescription, setTemplateDescription] = useState("");
  const [storageLocation, setStorageLocation] = useState("");
  const [isReadingFile, setIsReadingFile] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [createDocument, { isLoading: isUploadingFile }] = useCreateDocumentMutation();
  const [createTemplateStructure, { isLoading: isCreatingStructure }] =
    useCreateTemplateStructureMutation();

  const {
    data: locations = [],
    isLoading: isLocationsLoading,
  } = useListStorageLocationsQuery();

  const activeLocations = useMemo(() => locations.filter((l) => l.is_active), [locations]);

  // Auto-select local storage as soon as locations load
  const localLocation = useMemo(
    () => activeLocations.find((l) => l.provider === "local"),
    [activeLocations],
  );
  if (localLocation && !storageLocation) {
    setStorageLocation(localLocation.id);
  }

  const selectedTemplate = useMemo(
    () => IMPLEMENTED_TEMPLATES.find((t) => t.code === selectedTemplateCode) ?? null,
    [selectedTemplateCode],
  );

  const isSubmitting = isUploadingFile || isCreatingStructure;

  const totalColumns = sheets.reduce((sum, s) => sum + s.columns.length, 0);
  const requiredCount = sheets.reduce(
    (sum, s) => sum + s.columns.filter((c) => c.required).length,
    0,
  );

  // ── File selection ────────────────────────────────────────────────────────

  const handleFileChange = useCallback(
    async (_event: SyntheticEvent<HTMLElement, Event>, { addedFiles }: { addedFiles: File[] }) => {
      const selected = addedFiles[0];
      if (!selected) return;

      const ext = getFileExtension(selected.name);
      if (![".csv", ".xlsx", ".xls"].includes(ext)) {
        setError("Templates must be CSV or Excel files (.csv, .xlsx, .xls).");
        return;
      }

      setError(null);
      setFile(selected);
      setSheets([]);
      setIsReadingFile(true);

      try {
        const detected = await readSheetsFromFile(selected);
        if (detected.length === 0 || detected.every((s) => s.columns.length === 0)) {
          setError("No column headers found in the file. Make sure row 1 contains headers.");
          setFile(null);
          return;
        }
        setSheets(applyTemplateDefaults(selectedTemplateCode, detected));
      } catch {
        setError("Failed to read file headers. Please check the file and try again.");
        setFile(null);
      } finally {
        setIsReadingFile(false);
      }
    },
    [selectedTemplateCode],
  );

  const handleClearFile = useCallback(() => {
    setFile(null);
    setSheets([]);
    setError(null);
  }, []);

  // ── Column editor ─────────────────────────────────────────────────────────

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
    if (!file) return setError("Please select a file.");
    if (!storageLocation) return setError("Storage location unavailable.");
    if (!selectedTemplate) return setError("Please select a template.");
    if (sheets.length === 0) return setError("No columns detected in the file.");

    setError(null);

    try {
      const uploadedDoc = await createDocument({
        file,
        storageLocation,
        isTemplate: true,
      }).unwrap();

      await createTemplateStructure({
        template: {
          document_id: uploadedDoc.id,
          code: selectedTemplate.code,
          name: selectedTemplate.name,
          description: templateDescription.trim(),
          file_type: getFileExtension(file.name).replace(".", ""),
          configuration: {
            allowed_extensions: [getFileExtension(file.name)],
          },
        },
        sheets: sheets.map((sheet, si) => ({
          sheet: {
            code: normalizeKey(sheet.name),
            name: sheet.name,
            display_name: sheet.name,
            required: true,
            sheet_order: si + 1,
            header_row: sheet.headerRow,
            start_row: sheet.headerRow + 1,
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
            configuration: {},
          })),
        })),
      }).unwrap();

      onClose();
    } catch (err: unknown) {
      const message =
        typeof err === "object" &&
        err !== null &&
        "data" in err &&
        typeof (err as { data?: { message?: unknown } }).data?.message === "string"
          ? (err as { data?: { message?: string } }).data!.message
          : "Upload failed. Please try again.";
      setError(message!);
    }
  };

  const isSubmitDisabled =
    isSubmitting ||
    isReadingFile ||
    !file ||
    !storageLocation ||
    !selectedTemplate ||
    sheets.length === 0;

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div style={{ maxWidth: 760 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Template</h2>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload a CSV or Excel file to define a reusable template. Set which columns are required
          — uploads against this template will be rejected if required columns are missing.
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
              labelText="Drag and drop a CSV or Excel file here, or click to browse"
              accept={[".csv", ".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isSubmitting}
            />
          </FormGroup>

          {file && (
            <div style={{ display: "flex", alignItems: "center", gap: "0.75rem", flexWrap: "wrap" }}>
              <Tag type="cyan">Template file</Tag>
              <span style={{ fontWeight: 500 }}>{file.name}</span>
              <span style={{ color: "#6f6f6f" }}>{formatFileSize(file.size)}</span>
              {!isReadingFile && (
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

          {isReadingFile && <InlineLoading description="Reading columns from file..." />}

          {/* ── Column editor ── */}
          {!isReadingFile && sheets.length > 0 && (
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
                <div key={sheet.name} style={{ marginBottom: "1.5rem" }}>
                  {sheets.length > 1 && (
                    <p
                      style={{
                        margin: "0 0 0.5rem",
                        fontSize: "0.875rem",
                        color: "#6f6f6f",
                        fontWeight: 600,
                      }}
                    >
                      Sheet: {sheet.name}
                    </p>
                  )}

                  <table
                    style={{
                      width: "100%",
                      borderCollapse: "collapse",
                      fontSize: "0.875rem",
                    }}
                  >
                    <thead>
                      <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "30%" }}>
                          Column name
                        </th>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "25%" }}>
                          Column key
                          <span style={{ fontWeight: 400, color: "#6f6f6f", fontSize: "0.75rem", marginLeft: "0.35rem" }}>
                            (editable)
                          </span>
                        </th>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "25%" }}>
                          Data type
                        </th>
                        <th style={{ textAlign: "left", padding: "0.5rem 0.75rem", fontWeight: 600, width: "20%" }}>
                          Required
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
                          <td style={{ padding: "0.5rem 0.75rem", fontWeight: col.required ? 600 : 400 }}>
                            {col.column_name}
                          </td>
                          <td style={{ padding: "0.3rem 0.75rem" }}>
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
                          <td style={{ padding: "0.4rem 0.75rem" }}>
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
                          <td style={{ padding: "0.4rem 0.75rem" }}>
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
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ))}
            </div>
          )}

          {/* ── Template selector ── */}
          <div>
            <Select
              id="template-name"
              labelText="Template name"
              helperText="Select the template this file defines."
              value={selectedTemplateCode}
              onChange={(e) => {
                const nextTemplateCode = e.target.value;
                setSelectedTemplateCode(nextTemplateCode);
                setSheets((current) => applyTemplateDefaults(nextTemplateCode, current));
                setError(null);
              }}
              disabled={isSubmitting}
            >
              <SelectItem value="" text="Select a template" />
              {IMPLEMENTED_TEMPLATES.map((t) => (
                <SelectItem key={t.code} value={t.code} text={t.name} />
              ))}
            </Select>

            {selectedTemplate && (
              <div style={{ marginTop: "0.5rem", display: "flex", alignItems: "center", gap: "0.5rem" }}>
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
                  {selectedTemplate.code}
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
            {isSubmitting && (
              <InlineLoading
                description={isUploadingFile ? "Uploading file..." : "Saving template structure..."}
              />
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
