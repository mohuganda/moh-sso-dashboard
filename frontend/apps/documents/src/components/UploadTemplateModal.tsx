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
  useDeleteDocumentMutation,
  useScanDocumentStructureMutation,
  useListStorageLocationsQuery,
} from "../api";
import "./documents-components.scss";

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
  const [isDetecting, setIsDetecting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [createDocument, { isLoading: isUploadingFile }] = useCreateDocumentMutation();

  const [createTemplateStructure, { isLoading: isCreatingStructure }] =
    useCreateTemplateStructureMutation();

  const [scanStructure] = useScanDocumentStructureMutation();

  const [deleteDocument] = useDeleteDocumentMutation();

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
      setIsDetecting(true);

      // Structure detection uses the transient scan endpoint — it parses the
      // file without creating a document record, so browsing/dropping a file
      // (or retrying after a bad file) never leaves an orphaned upload behind.
      // The real document is only created on final "Save Template" submit.
      try {
        const result = await scanStructure(selectedFile).unwrap();

        if (result.sheets.length === 0) {
          setError(
            "No usable sheets found. Make sure visible sheets have at least 4 column headers.",
          );
          setFile(null);
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
      } finally {
        setIsDetecting(false);
      }
    },
    [storageLocation, scanStructure],
  );

  const handleClearFile = useCallback(() => {
    setFile(null);
    setSheets([]);
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
    if (!file) {
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

    let uploadedDocId: string | undefined;

    try {
      // The document backing this template is only created here, at the
      // point the user has actually committed to saving — never as a side
      // effect of browsing/dropping a file during structure detection.
      const uploaded = await createDocument({
        file,
        storageLocation,
        isTemplate: true,
      }).unwrap();

      uploadedDocId = uploaded.id;

      await createTemplateStructure({
        template: {
          document_id: uploaded.id,
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

      handleClearFile();
      onClose();
    } catch (caughtError: unknown) {
      // If the template file was uploaded but saving its structure failed
      // (e.g. duplicate code), don't leave the orphaned document behind.
      if (uploadedDocId) {
        void deleteDocument(uploadedDocId);
      }

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
    !storageLocation ||
    !templateName.trim() ||
    !templateCode ||
    includedSheets.length === 0;

  return (
    <div className="document-upload-modal document-upload-modal--wide">
      <div className="document-upload-modal__header">
        <h2>Upload Template</h2>

        <p>
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
              <div className="document-upload-modal__code-preview">
                <span className="document-upload-modal__code-label">Code:</span>

                <code className="document-upload-modal__code">{templateCode}</code>
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
            <div className="document-upload-modal__file">
              <Tag type="cyan">Template file</Tag>

              <span className="document-upload-modal__file-name">{file.name}</span>

              <span className="document-upload-modal__file-size">{formatFileSize(file.size)}</span>

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
              description="Detecting columns from file…"
            />
          )}

          {!isDetecting && sheets.length > 0 && (
            <div>
              <div className="document-upload-modal__detected-header">
                <p className="document-upload-modal__detected-title">
                  Columns detected —{" "}
                  <span className="document-upload-modal__detected-note">
                    toggle <strong>Required</strong> for columns that must be present when data is
                    uploaded
                  </span>
                </p>

                <span className="document-upload-modal__detected-count">
                  {requiredCount} of {totalColumns} required
                </span>
              </div>

              {sheets.map((sheet, sheetIndex) => (
                <div
                  key={`${sheet.name}-${sheetIndex}`}
                  className={
                    sheet.excluded
                      ? "document-upload-modal__sheet document-upload-modal__sheet--excluded"
                      : "document-upload-modal__sheet"
                  }
                >
                  {sheets.length > 1 && (
                    <div className="document-upload-modal__sheet-header">
                      <p
                        className={
                          sheet.excluded
                            ? "document-upload-modal__sheet-title document-upload-modal__sheet-title--excluded"
                            : "document-upload-modal__sheet-title"
                        }
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
                    <table className="document-upload-modal__columns-table">
                      <thead>
                        <tr>
                          <th>Column name</th>

                          <th>
                            Column key
                            <span className="document-upload-modal__header-note">(editable)</span>
                          </th>

                          <th>Data type</th>

                          <th>Required</th>

                          <th>Filterable</th>
                        </tr>
                      </thead>

                      <tbody>
                        {sheet.columns.map((column, columnIndex) => (
                          <tr
                            key={`${sheetIndex}-${columnIndex}`}
                            className={
                              column.required ? "document-upload-modal__required-row" : undefined
                            }
                          >
                            <td
                              className={
                                column.required ? "document-upload-modal__required-name" : undefined
                              }
                            >
                              {column.column_name}
                            </td>

                            <td className="document-upload-modal__compact-cell">
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
                                className="document-upload-modal__mono-input"
                              />
                            </td>

                            <td className="document-upload-modal__compact-cell">
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

                            <td>
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

                            <td>
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

          <div className="document-upload-modal__actions">
            {(isUploadingFile || isCreatingStructure) && (
              <InlineLoading
                description={isUploadingFile ? "Uploading file…" : "Saving template structure…"}
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
}

export const UploadTemplateModal: React.FC<UploadTemplateModalProps> = ({ onClose }) => {
  return (
    <PermissionGuard
      permission={PERMISSIONS.documentTemplatesWrite}
      fallback={
        <div className="document-upload-modal document-upload-modal--wide">
          <InlineNotification
            kind="warning"
            title="Access denied"
            subtitle="You do not have permission to create document templates."
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
      <UploadTemplateModalContent onClose={onClose} />
    </PermissionGuard>
  );
};
