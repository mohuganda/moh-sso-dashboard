import { useCallback, useEffect, useMemo, useState, type SyntheticEvent } from "react";
import { useSelector } from "react-redux";
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
import { TrashCan } from "@carbon/react/icons";

import { PermissionGuard, PERMISSIONS, selectUser } from "@moh-sso/auth";
import { formatFileSize } from "@moh-sso/utils";

import {
  useCreateDocumentMutation,
  useCreateTemplateStructureMutation,
  useLazyParseDocumentStructureQuery,
  useListStorageLocationsQuery,
} from "../api";

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

const DATA_TYPE_OPTIONS = [
  { value: "STRING", label: "Text" },
  { value: "INTEGER", label: "Integer" },
  { value: "DECIMAL", label: "Decimal" },
  { value: "DATE", label: "Date" },
  { value: "DATETIME", label: "Date & Time" },
  { value: "BOOLEAN", label: "Boolean" },
];

function getFileExtension(fileName: string): string {
  const index = fileName.lastIndexOf(".");

  return index >= 0 ? fileName.slice(index).toLowerCase() : "";
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

function UploadTemplateModalContent({ onClose }: UploadTemplateModalProps) {
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

  const { data: locations = [], isLoading: isLocationsLoading } = useListStorageLocationsQuery();

  const activeLocations = useMemo(
    () => locations.filter((location) => location.is_active),
    [locations],
  );

  const localLocation = useMemo(
    () => activeLocations.find((location) => location.provider === "local"),
    [activeLocations],
  );

  useEffect(() => {
    if (localLocation && !storageLocation) {
      setStorageLocation(localLocation.id);
    }
  }, [localLocation, storageLocation]);

  const templateCode = useMemo(() => normalizeKey(templateName).toUpperCase(), [templateName]);

  const isSubmitting = isUploadingFile || isCreatingStructure || isDetecting;

  const includedSheets = useMemo(() => sheets.filter((sheet) => !sheet.excluded), [sheets]);

  const totalColumns = useMemo(
    () => includedSheets.reduce((sum, sheet) => sum + sheet.columns.length, 0),
    [includedSheets],
  );

  const requiredCount = useMemo(
    () =>
      includedSheets.reduce(
        (sum, sheet) => sum + sheet.columns.filter((column) => column.required).length,
        0,
      ),
    [includedSheets],
  );

  const handleFileChange = useCallback(
    async (
      _event: SyntheticEvent<HTMLElement, Event>,
      {
        addedFiles,
      }: {
        addedFiles: File[];
      },
    ) => {
      const selectedFile = addedFiles[0];

      if (!selectedFile) {
        return;
      }

      const extension = getFileExtension(selectedFile.name);

      if (![".xlsx", ".xls"].includes(extension)) {
        setError("Templates must be Excel files (.xlsx or .xls).");
        return;
      }

      if (!storageLocation) {
        setError("Storage location unavailable.");
        return;
      }

      setError(null);
      setFile(selectedFile);
      setSheets([]);
      setTemplateName("");
      setTemplateDescription("");
      setPreUploadedDocId(null);
      setIsDetecting(true);

      try {
        const uploaded = await createDocument({
          file: selectedFile,
          storageLocation,
          isTemplate: true,
        }).unwrap();

        setPreUploadedDocId(uploaded.id);

        const result = await triggerParseStructure(uploaded.id).unwrap();

        if (result.sheets.length === 0) {
          setError(
            "No usable sheets found. Make sure visible sheets have at least 4 column headers.",
          );
          setFile(null);
          setPreUploadedDocId(null);
          return;
        }

        setSheets(
          result.sheets.map((sheet) => ({
            name: sheet.name,
            excluded: false,
            headerRow: sheet.header_row,
            startRow: sheet.start_row,
            columns: sheet.columns.map((column, index) => ({
              column_key: column.column_key,
              column_name: column.column_name,
              data_type: "STRING",
              required: false,
              filterable: false,
              column_order: index + 1,
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
    setTemplateName("");
    setTemplateDescription("");
    setError(null);
  }, []);

  const toggleSheetExcluded = useCallback((sheetIndex: number) => {
    setSheets((previous) =>
      previous.map((sheet, index) =>
        index !== sheetIndex
          ? sheet
          : {
              ...sheet,
              excluded: !sheet.excluded,
            },
      ),
    );
  }, []);

  const updateColumn = useCallback(
    (sheetIndex: number, columnIndex: number, patch: Partial<ColumnDraft>) => {
      setSheets((previous) =>
        previous.map((sheet, currentSheetIndex) =>
          currentSheetIndex !== sheetIndex
            ? sheet
            : {
                ...sheet,
                columns: sheet.columns.map((column, currentColumnIndex) =>
                  currentColumnIndex !== columnIndex
                    ? column
                    : {
                        ...column,
                        ...patch,
                      },
                ),
              },
        ),
      );
    },
    [],
  );

  async function handleUpload() {
    if (!file || !preUploadedDocId) {
      setError("Please select a file.");
      return;
    }

    if (!storageLocation) {
      setError("Storage location unavailable.");
      return;
    }

    if (!templateName.trim()) {
      setError("Please enter a template name.");
      return;
    }

    if (includedSheets.length === 0) {
      setError("No sheets are included. Toggle at least one sheet on.");
      return;
    }

    if (!templateCode) {
      setError("Template name must produce a valid template code.");
      return;
    }

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
        sheets: includedSheets.map((sheet, sheetIndex) => ({
          sheet: {
            code: normalizeKey(sheet.name),
            name: sheet.name,
            display_name: sheet.name,
            required: true,
            sheet_order: sheetIndex + 1,
            header_row: sheet.headerRow,
            start_row: sheet.startRow,
            allow_extra_columns: true,
            allow_duplicate_headers: false,
            configuration: {},
          },
          columns: sheet.columns.map((column) => ({
            column_key: column.column_key,
            column_name: column.column_name,
            display_name: column.column_name,
            data_type: column.data_type,
            required: column.required,
            is_unique: false,
            column_order: column.column_order,
            default_value: null,
            allowed_values: [],
            aliases: [],
            configuration: {
              filterable: column.filterable,
            },
          })),
        })),
      }).unwrap();

      onClose();
    } catch (caughtError: unknown) {
      const errorData =
        typeof caughtError === "object" && caughtError !== null && "data" in caughtError
          ? (
              caughtError as {
                data?: {
                  message?: string;
                  code?: string;
                };
              }
            ).data
          : null;

      let message = errorData?.message ?? "Upload failed. Please try again.";

      if (errorData?.code === "TEMPLATE_CODE_EXISTS") {
        message = `A template with code "${templateCode}" already exists. Try a different name, or edit the existing template from the Templates list.`;
      }

      setError(message);
    }
  }

  const isSubmitDisabled =
    isSubmitting ||
    !file ||
    !preUploadedDocId ||
    !storageLocation ||
    !templateName.trim() ||
    !templateCode ||
    includedSheets.length === 0;

  return (
    <div style={{ maxWidth: 760 }}>
      <div
        style={{
          marginBottom: "1.5rem",
        }}
      >
        <h2
          style={{
            margin: 0,
            marginBottom: "0.5rem",
          }}
        >
          Upload Template
        </h2>

        <p
          style={{
            margin: 0,
            color: "#6f6f6f",
          }}
        >
          Upload an Excel file to define a reusable template. Set which columns are required so
          uploads against this template can be validated.
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

          <FormGroup legendText="Template file">
            <FileUploaderDropContainer
              labelText="Drag and drop an Excel file (.xlsx / .xls) here, or click to browse"
              accept={[".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isSubmitting}
            />
          </FormGroup>

          <div>
            <TextInput
              id="template-name"
              labelText="Template name"
              helperText='Give this template a unique name, for example "GHSC PSM" or "NMS Stock Report".'
              placeholder="e.g. GHSC PSM"
              value={templateName}
              onChange={(event) => {
                setTemplateName(event.target.value);
                setError(null);
              }}
              disabled={isSubmitting}
              invalid={Boolean(templateName.trim()) && !templateCode}
              invalidText="Enter a name containing letters or numbers."
            />

            {templateName.trim() && templateCode && (
              <div
                style={{
                  marginTop: "0.4rem",
                  display: "flex",
                  alignItems: "center",
                  gap: "0.5rem",
                }}
              >
                <span
                  style={{
                    fontSize: "0.75rem",
                    color: "#6f6f6f",
                  }}
                >
                  Code:
                </span>

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
            onChange={(event) => setTemplateDescription(event.target.value)}
            disabled={isSubmitting}
          />

          {file && (
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.75rem",
                flexWrap: "wrap",
              }}
            >
              <Tag type="cyan">Template file</Tag>

              <span
                style={{
                  fontWeight: 500,
                }}
              >
                {file.name}
              </span>

              <span
                style={{
                  color: "#6f6f6f",
                }}
              >
                {formatFileSize(file.size)}
              </span>

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
              description={isUploadingFile ? "Uploading file…" : "Detecting columns from file…"}
            />
          )}

          {!isDetecting && sheets.length > 0 && (
            <div>
              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "baseline",
                  marginBottom: "0.75rem",
                  gap: "1rem",
                }}
              >
                <p
                  style={{
                    margin: 0,
                    fontWeight: 600,
                  }}
                >
                  Columns detected —{" "}
                  <span
                    style={{
                      color: "#6f6f6f",
                      fontWeight: 400,
                    }}
                  >
                    toggle <strong>Required</strong> for columns that must be present when data is
                    uploaded
                  </span>
                </p>

                <span
                  style={{
                    color: "#6f6f6f",
                    fontSize: "0.875rem",
                    whiteSpace: "nowrap",
                  }}
                >
                  {requiredCount} of {totalColumns} required
                </span>
              </div>

              {sheets.map((sheet, sheetIndex) => (
                <div
                  key={`${sheet.name}-${sheetIndex}`}
                  style={{
                    marginBottom: "1.5rem",
                    opacity: sheet.excluded ? 0.6 : 1,
                  }}
                >
                  {sheets.length > 1 && (
                    <div
                      style={{
                        display: "flex",
                        alignItems: "center",
                        gap: "0.75rem",
                        marginBottom: "0.5rem",
                        flexWrap: "wrap",
                      }}
                    >
                      <p
                        style={{
                          margin: 0,
                          fontSize: "0.875rem",
                          color: sheet.excluded ? "#a8a8a8" : "#6f6f6f",
                          fontWeight: 600,
                        }}
                      >
                        Sheet: {sheet.name}
                      </p>

                      <Toggle
                        id={`sheet-include-${sheetIndex}`}
                        labelText="Include sheet"
                        labelA="Excluded"
                        labelB="Included"
                        size="sm"
                        toggled={!sheet.excluded}
                        onToggle={() => toggleSheetExcluded(sheetIndex)}
                        disabled={isSubmitting}
                      />
                    </div>
                  )}

                  {!sheet.excluded && (
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
                          }}
                        >
                          <th
                            style={{
                              textAlign: "left",
                              padding: "0.5rem 0.75rem",
                              fontWeight: 600,
                              width: "25%",
                              verticalAlign: "middle",
                            }}
                          >
                            Column name
                          </th>

                          <th
                            style={{
                              textAlign: "left",
                              padding: "0.5rem 0.75rem",
                              fontWeight: 600,
                              width: "30%",
                              verticalAlign: "middle",
                            }}
                          >
                            Column key
                            <span
                              style={{
                                fontWeight: 400,
                                color: "#6f6f6f",
                                fontSize: "0.75rem",
                                marginLeft: "0.35rem",
                              }}
                            >
                              (editable)
                            </span>
                          </th>

                          <th
                            style={{
                              textAlign: "left",
                              padding: "0.5rem 0.75rem",
                              fontWeight: 600,
                              width: "20%",
                              verticalAlign: "middle",
                            }}
                          >
                            Data type
                          </th>

                          <th
                            style={{
                              textAlign: "left",
                              padding: "0.5rem 0.75rem",
                              fontWeight: 600,
                              width: "12%",
                              verticalAlign: "middle",
                            }}
                          >
                            Required
                          </th>

                          <th
                            style={{
                              textAlign: "left",
                              padding: "0.5rem 0.75rem",
                              fontWeight: 600,
                              width: "13%",
                              verticalAlign: "middle",
                            }}
                          >
                            Filterable
                          </th>
                        </tr>
                      </thead>

                      <tbody>
                        {sheet.columns.map((column, columnIndex) => (
                          <tr
                            key={`${sheetIndex}-${columnIndex}`}
                            style={{
                              borderBottom: "1px solid #f4f4f4",
                              backgroundColor: column.required ? "#f0f7ff" : undefined,
                            }}
                          >
                            <td
                              style={{
                                padding: "0.5rem 0.75rem",
                                fontWeight: column.required ? 600 : 400,
                                verticalAlign: "middle",
                              }}
                            >
                              {column.column_name}
                            </td>

                            <td
                              style={{
                                padding: "0.25rem 0.75rem",
                                verticalAlign: "middle",
                              }}
                            >
                              <TextInput
                                id={`key-${sheetIndex}-${columnIndex}`}
                                labelText="Column key"
                                hideLabel
                                size="sm"
                                value={column.column_key}
                                onChange={(event) =>
                                  updateColumn(sheetIndex, columnIndex, {
                                    column_key: event.target.value
                                      .trim()
                                      .toLowerCase()
                                      .replace(/[^a-z0-9_]/g, ""),
                                  })
                                }
                                disabled={isSubmitting}
                                style={{
                                  fontFamily: "monospace",
                                }}
                              />
                            </td>

                            <td
                              style={{
                                padding: "0.25rem 0.75rem",
                                verticalAlign: "middle",
                              }}
                            >
                              <Select
                                id={`dtype-${sheetIndex}-${columnIndex}`}
                                labelText="Data type"
                                hideLabel
                                size="sm"
                                value={column.data_type}
                                onChange={(event) =>
                                  updateColumn(sheetIndex, columnIndex, {
                                    data_type: event.target.value,
                                  })
                                }
                                disabled={isSubmitting}
                              >
                                {DATA_TYPE_OPTIONS.map((option) => (
                                  <SelectItem
                                    key={option.value}
                                    value={option.value}
                                    text={option.label}
                                  />
                                ))}
                              </Select>
                            </td>

                            <td
                              style={{
                                padding: "0.5rem 0.75rem",
                                verticalAlign: "middle",
                              }}
                            >
                              <Toggle
                                id={`req-${sheetIndex}-${columnIndex}`}
                                labelText="Required"
                                hideLabel
                                size="sm"
                                toggled={column.required}
                                onToggle={(checked) =>
                                  updateColumn(sheetIndex, columnIndex, {
                                    required: checked,
                                  })
                                }
                                disabled={isSubmitting}
                              />
                            </td>

                            <td
                              style={{
                                padding: "0.5rem 0.75rem",
                                verticalAlign: "middle",
                              }}
                            >
                              <Toggle
                                id={`filter-${sheetIndex}-${columnIndex}`}
                                labelText="Filterable"
                                hideLabel
                                size="sm"
                                toggled={column.filterable}
                                onToggle={(checked) =>
                                  updateColumn(sheetIndex, columnIndex, {
                                    filterable: checked,
                                  })
                                }
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

          <Select
            id="storage-location"
            labelText="Storage location"
            value={storageLocation}
            onChange={() => undefined}
            disabled
          >
            {localLocation ? (
              <SelectItem value={localLocation.id} text={localLocation.name} />
            ) : (
              <SelectItem value="" text={isLocationsLoading ? "Loading…" : "No local storage"} />
            )}
          </Select>

          <div
            style={{
              display: "flex",
              justifyContent: "flex-end",
              alignItems: "center",
              gap: "1rem",
              marginTop: "0.5rem",
            }}
          >
            {isCreatingStructure && <InlineLoading description="Saving template structure…" />}

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
}

export const UploadTemplateModal: React.FC<UploadTemplateModalProps> = ({ onClose }) => {
  return (
    <PermissionGuard
      permission={PERMISSIONS.documentTemplatesWrite}
      fallback={
        <div style={{ maxWidth: 760 }}>
          <InlineNotification
            kind="warning"
            title="Access denied"
            subtitle="You do not have permission to create document templates."
            lowContrast
            hideCloseButton
          />

          <div
            style={{
              display: "flex",
              justifyContent: "flex-end",
              marginTop: "1.5rem",
            }}
          >
            <Button kind="secondary" onClick={onClose}>
              Close
            </Button>
          </div>
        </div>
      }
    >
      <UploadTemplateModalContent onClose={onClose} />
    </PermissionGuard>
  );
};
