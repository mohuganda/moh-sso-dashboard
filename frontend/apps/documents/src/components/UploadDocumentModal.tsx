import { useCallback, useEffect, useMemo, useState, type SyntheticEvent } from "react";
import {
  Button,
  DatePicker,
  DatePickerInput,
  FileUploaderDropContainer,
  Form,
  FormGroup,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Stack,
  Tag,
} from "@carbon/react";
import { TrashCan } from "@carbon/react/icons";

import { PermissionGuard, PERMISSIONS } from "@moh-sso/auth";
import {
  formatFileSize,
  getSuggestedProcessType,
  isAcceptedFile,
  isPdfFile,
  requiresProcessing,
  validateProcessTypeAgainstFile,
} from "@moh-sso/utils";

import {
  useCreateDocumentMutation,
  useGetTemplateHasDataQuery,
  useGetTemplateStructureQuery,
  useListActiveTemplatesQuery,
  useListDocumentsQuery,
  useListStorageLocationsQuery,
} from "../api";
import "./documents-components.scss";
import { DOCUMENT_PROCESS_TYPE_OPTIONS, type DocumentProcessType } from "../types";

const CSV_HEADER_READ_BYTES = 64 * 1024;

function getFileExtension(fileName: string): string {
  const index = fileName.lastIndexOf(".");

  return index >= 0 ? fileName.slice(index).toLowerCase() : "";
}

function normalizeForCompare(value: string): string {
  return value.trim().toLowerCase().replace(/\s+/g, "_");
}

async function extractFileHeaders(file: File): Promise<string[]> {
  const extension = getFileExtension(file.name);

  if (extension === ".csv") {
    const { default: Papa } = await import("papaparse");

    const text = await file.slice(0, CSV_HEADER_READ_BYTES).text();

    return new Promise((resolve, reject) => {
      Papa.parse<Record<string, unknown>>(text, {
        header: true,
        skipEmptyLines: true,
        preview: 1,
        complete: (results) => {
          resolve(
            (results.meta.fields ?? [])
              .map((header) => String(header ?? "").trim())
              .filter(Boolean),
          );
        },
        error: reject,
      });
    });
  }

  if (extension === ".xlsx" || extension === ".xls") {
    const XLSX = await import("xlsx");
    const buffer = await file.arrayBuffer();

    const workbook = XLSX.read(buffer, {
      type: "array",
      sheetRows: 10,
    });

    const visibleSheetNames = workbook.SheetNames.filter((_, index) => {
      const metadata = workbook.Workbook?.Sheets?.[index];

      return !metadata?.Hidden;
    });

    const allHeaders: string[] = [];

    for (const sheetName of visibleSheetNames) {
      const worksheet = workbook.Sheets[sheetName];

      if (!worksheet) {
        continue;
      }

      const rows = XLSX.utils.sheet_to_json<(string | number | null)[]>(worksheet, {
        header: 1,
        defval: "",
        blankrows: false,
      });

      let headerRowIndex = 0;

      for (let rowIndex = 0; rowIndex < rows.length; rowIndex += 1) {
        const nonEmptyCells = rows[rowIndex].filter((cell) => String(cell ?? "").trim() !== "");

        if (nonEmptyCells.length >= 4) {
          headerRowIndex = rowIndex;
          break;
        }
      }

      const headerRow = Array.isArray(rows[headerRowIndex]) ? rows[headerRowIndex] : [];

      allHeaders.push(...headerRow.map((header) => String(header ?? "").trim()).filter(Boolean));
    }

    return allHeaders;
  }

  return [];
}

export type UploadDocumentModalProps = {
  onClose: () => void;
};

type ColumnValidationResult = {
  required: string[];
  present: string[];
  missing: string[];
};

function UploadDocumentModalContent({ onClose }: UploadDocumentModalProps) {
  const [file, setFile] = useState<File | null>(null);

  const [fileHeaders, setFileHeaders] = useState<string[]>([]);

  const [storageLocation, setStorageLocation] = useState("");

  const [processType, setProcessType] = useState<DocumentProcessType | "">("");

  const [selectedTemplate, setSelectedTemplate] = useState("");

  const [reportDate, setReportDate] = useState("");

  const [reportDateTouched, setReportDateTouched] = useState(false);

  const [isReadingFile, setIsReadingFile] = useState(false);

  const [error, setError] = useState<string | null>(null);

  const [reuploadMode, setReuploadMode] = useState<"same" | "new" | null>(null);

  const todayString = useMemo(() => new Date().toISOString().split("T")[0], []);

  const [createDocument, { isLoading: isUploading }] = useCreateDocumentMutation();

  const { data: allDocuments = [] } = useListDocumentsQuery();

  const { data: savedTemplates = [], isLoading: isTemplatesLoading } =
    useListActiveTemplatesQuery();

  const { data: templateStructure, isFetching: isFetchingStructure } = useGetTemplateStructureQuery(
    selectedTemplate,
    {
      skip: !selectedTemplate,
    },
  );

  useGetTemplateHasDataQuery(selectedTemplate, {
    skip: !selectedTemplate || !processType,
  });

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

  const previousUpload = useMemo(() => {
    if (!file || !selectedTemplate) {
      return null;
    }

    const matches = allDocuments
      .filter(
        (document) =>
          document.original_filename === file.name &&
          document.metadata?.template_code === selectedTemplate,
      )
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

    return matches[0] ?? null;
  }, [file, selectedTemplate, allDocuments]);
  const previousReportDate = previousUpload?.metadata?.report_date ?? null;

  useEffect(() => {
    if (previousUpload && previousReportDate) {
      setReuploadMode("same");
      setReportDate(previousReportDate);
      setReportDateTouched(false);
      return;
    }

    if (!previousUpload) {
      setReuploadMode(null);
    }
  }, [previousUpload, previousReportDate]);

  const fileNeedsProcessing = useMemo(() => (file ? requiresProcessing(file) : false), [file]);

  const isPdf = useMemo(() => (file ? isPdfFile(file) : false), [file]);

  // CSV is the only format that can be imported without a template — it
  // needs no template-defined structure to parse (unlike Excel, which
  // relies on the template to know header row / column layout / types).
  // Mirrors the backend's isCSV check (mapper.go): MIME type OR extension,
  // not extension alone — a mismatch here means the frontend could decide
  // not to send process_type while the backend still expects one.
  const isCsvFile = useMemo(() => {
    if (!file) return false;
    return file.type?.toLowerCase().trim() === "text/csv" || getFileExtension(file.name) === ".csv";
  }, [file]);
  const canProcessAdHoc = fileNeedsProcessing && isCsvFile;

  const columnValidation = useMemo<ColumnValidationResult | null>(() => {
    if (!selectedTemplate || !templateStructure || !file || isPdf) {
      return null;
    }

    const requiredColumns = templateStructure.sheets.flatMap((sheet) =>
      sheet.columns.filter((column) => column.required).map((column) => column.column_name),
    );

    if (requiredColumns.length === 0) {
      return null;
    }

    const fileHeaderSet = new Set(fileHeaders.map(normalizeForCompare));

    const present: string[] = [];
    const missing: string[] = [];

    for (const columnName of requiredColumns) {
      if (fileHeaderSet.has(normalizeForCompare(columnName))) {
        present.push(columnName);
      } else {
        missing.push(columnName);
      }
    }

    return {
      required: requiredColumns,
      present,
      missing,
    };
  }, [selectedTemplate, templateStructure, file, isPdf, fileHeaders]);

  const isTemplateValid = columnValidation === null || columnValidation.missing.length === 0;

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

      if (!isAcceptedFile(selectedFile)) {
        setFile(null);
        setFileHeaders([]);
        setError("Only PDF, CSV, or Excel (.pdf, .csv, .xlsx, .xls) files are allowed.");
        return;
      }

      setError(null);
      setFile(selectedFile);
      setFileHeaders([]);

      if (requiresProcessing(selectedFile)) {
        setProcessType((current) => current || getSuggestedProcessType(selectedFile));
      } else {
        setProcessType("");
      }

      if (isPdfFile(selectedFile)) {
        return;
      }

      setIsReadingFile(true);

      try {
        const headers = await extractFileHeaders(selectedFile);

        setFileHeaders(headers);
      } catch {
        setError("Failed to read file headers.");
      } finally {
        setIsReadingFile(false);
      }
    },
    [],
  );

  const handleClearFile = useCallback(() => {
    setFile(null);
    setFileHeaders([]);
    setProcessType("");
    setSelectedTemplate("");
    setReportDate("");
    setReportDateTouched(false);
    setReuploadMode(null);
    setError(null);
  }, []);

  async function handleUpload() {
    if (!file) {
      setError("Please select a file.");
      return;
    }

    if (!storageLocation) {
      setError("Please select a storage location.");
      return;
    }

    if ((selectedTemplate || canProcessAdHoc) && fileNeedsProcessing && !processType) {
      setError("Please select a process type.");
      return;
    }

    if ((selectedTemplate || canProcessAdHoc) && fileNeedsProcessing && processType) {
      const validationError = validateProcessTypeAgainstFile(file, processType);

      if (validationError) {
        setError(validationError);
        return;
      }
    }

    if (!isTemplateValid) {
      setError(`Missing required columns: ${columnValidation?.missing.join(", ") ?? ""}`);
      return;
    }

    if (selectedTemplate && fileNeedsProcessing && !reportDate) {
      setReportDateTouched(true);
      setError("Please select a report date.");
      return;
    }

    if (reuploadMode === "new" && previousReportDate && reportDate === previousReportDate) {
      setReportDateTouched(true);
      setError(
        `Report date cannot be the same as the previous upload (${previousReportDate}). Choose a different date or select "Yes, same report date" to replace it.`,
      );
      return;
    }

    setError(null);

    const shouldReplace = Boolean(previousUpload && reuploadMode === "same");

    const metadata = selectedTemplate
      ? {
          template_code: selectedTemplate,
          ...(reportDate
            ? {
                report_date: reportDate,
              }
            : {}),
          ...(shouldReplace
            ? {
                replace_document_id: previousUpload!.id,
              }
            : {}),
        }
      : undefined;

    try {
      await createDocument({
        file,
        storageLocation,
        ...((selectedTemplate || canProcessAdHoc) && fileNeedsProcessing && processType
          ? { processType }
          : {}),
        ...(metadata ? { metadata } : {}),
      }).unwrap();

      onClose();
    } catch (caughtError: unknown) {
      const responseMessage =
        typeof caughtError === "object" &&
        caughtError !== null &&
        "data" in caughtError &&
        typeof (
          caughtError as {
            data?: { error?: { message?: unknown }; message?: unknown };
          }
        ).data?.error?.message === "string"
          ? (
              caughtError as {
                data?: { error?: { message?: string } };
              }
            ).data?.error?.message
          : typeof caughtError === "object" &&
              caughtError !== null &&
              "data" in caughtError &&
              typeof (caughtError as { data?: { message?: unknown } }).data?.message === "string"
            ? (caughtError as { data?: { message?: string } }).data?.message
            : null;

      setError(responseMessage ?? "Upload failed. Please try again.");
    }
  }

  const isSubmitDisabled =
    isUploading ||
    isReadingFile ||
    isFetchingStructure ||
    !file ||
    !storageLocation ||
    ((selectedTemplate || canProcessAdHoc) && fileNeedsProcessing && !processType) ||
    (selectedTemplate && fileNeedsProcessing && !reportDate) ||
    (reuploadMode === "new" && !!previousReportDate && reportDate === previousReportDate) ||
    !isTemplateValid;

  return (
    <div className="document-upload-modal">
      <div className="document-upload-modal__header">
        <h2>Upload Document</h2>

        <p>
          Upload a file for processing. Optionally select a template to validate the file&apos;s
          columns before uploading.
        </p>
      </div>

      <Form>
        <Stack gap={6}>
          {error && (
            <InlineNotification
              kind="error"
              title="Upload error"
              subtitle={error}
              lowContrast
              onCloseButtonClick={() => setError(null)}
            />
          )}

          <FormGroup legendText="Document file">
            <FileUploaderDropContainer
              labelText="Drag and drop a file here, or click to browse"
              accept={[".pdf", ".csv", ".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isUploading}
            />
          </FormGroup>

          {file && (
            <div className="document-upload-modal__file">
              <Tag type="blue">Selected</Tag>

              <span className="document-upload-modal__file-name">{file.name}</span>

              <span className="document-upload-modal__file-size">{formatFileSize(file.size)}</span>

              {isPdf && <Tag type="cool-gray">No processing required</Tag>}

              {isReadingFile && <InlineLoading description="Reading file headers..." />}

              {!isReadingFile && (
                <Button
                  kind="ghost"
                  size="sm"
                  renderIcon={TrashCan}
                  iconDescription="Remove file"
                  onClick={handleClearFile}
                  disabled={isUploading}
                >
                  Remove
                </Button>
              )}
            </div>
          )}

          {previousUpload && fileNeedsProcessing && (
            <div className="document-upload-modal__warning">
              <p className="document-upload-modal__warning-title">
                This filename was uploaded before
              </p>

              <p className="document-upload-modal__warning-text">
                Previous report date: <strong>{previousReportDate ?? "—"}</strong>
              </p>

              <div className="document-upload-modal__radio-list">
                <label className="document-upload-modal__radio">
                  <input
                    type="radio"
                    name="reupload-mode"
                    value="same"
                    checked={reuploadMode === "same" || reuploadMode === null}
                    onChange={() => {
                      setReuploadMode("same");

                      if (previousReportDate) {
                        setReportDate(previousReportDate);
                      }
                    }}
                  />

                  <span className="document-upload-modal__radio-text">
                    <strong>Yes, same report date</strong> — replace the previous upload
                    {previousReportDate && (
                      <span className="document-upload-modal__radio-note">
                        {" "}
                        ({previousReportDate})
                      </span>
                    )}
                  </span>
                </label>

                <label className="document-upload-modal__radio">
                  <input
                    type="radio"
                    name="reupload-mode"
                    value="new"
                    checked={reuploadMode === "new"}
                    onChange={() => {
                      setReuploadMode("new");
                      setReportDate("");
                      setReportDateTouched(false);
                    }}
                  />

                  <span className="document-upload-modal__radio-text">
                    <strong>No, different report date</strong> — add as a new upload
                  </span>
                </label>
              </div>
            </div>
          )}

          <Select
            id="saved-template"
            labelText="Validate against template (optional)"
            helperText={
              isPdf
                ? "Template validation is not applicable for PDF files."
                : "Select a template to check that all required columns are present."
            }
            value={selectedTemplate}
            onChange={(event) => {
              setSelectedTemplate(event.target.value);
              setError(null);
            }}
            disabled={!file || isUploading || isReadingFile || isPdf}
          >
            <SelectItem
              value=""
              text={isTemplatesLoading ? "Loading templates..." : "No template (optional)"}
            />

            {savedTemplates.map((template) => (
              <SelectItem
                key={template.id}
                value={template.code}
                text={`${template.name} (${template.code})`}
              />
            ))}
          </Select>

          {selectedTemplate && (
            <>
              {isFetchingStructure && <InlineLoading description="Loading template columns..." />}

              {!isFetchingStructure && templateStructure && columnValidation === null && (
                <InlineNotification
                  kind="info"
                  title="No required columns"
                  subtitle="This template has no required columns defined."
                  lowContrast
                />
              )}

              {!isFetchingStructure &&
                columnValidation !== null &&
                fileHeaders.length === 0 &&
                !isReadingFile && (
                  <InlineNotification
                    kind="warning"
                    title="Waiting for file"
                    subtitle="Select a file to validate its columns against the template."
                    lowContrast
                  />
                )}

              {!isFetchingStructure && columnValidation !== null && fileHeaders.length > 0 && (
                <div>
                  {columnValidation.missing.length > 0 ? (
                    <InlineNotification
                      kind="error"
                      title="Required columns missing"
                      subtitle={`The file is missing: ${columnValidation.missing.join(", ")}`}
                      lowContrast
                    />
                  ) : (
                    <InlineNotification
                      kind="success"
                      title="All required columns present"
                      subtitle={`${columnValidation.present.length} required column${
                        columnValidation.present.length !== 1 ? "s" : ""
                      } validated successfully.`}
                      lowContrast
                    />
                  )}

                  <div className="document-upload-modal__section">
                    <p className="document-upload-modal__section-title">Required columns</p>

                    <div className="document-upload-modal__chip-list">
                      {columnValidation.required.map((columnName) => {
                        const isPresent = columnValidation.present.includes(columnName);

                        return (
                          <span
                            key={columnName}
                            className={
                              isPresent
                                ? "document-upload-modal__validation-chip document-upload-modal__validation-chip--present"
                                : "document-upload-modal__validation-chip"
                            }
                          >
                            {isPresent ? "✓" : "✗"} {columnName}
                          </span>
                        );
                      })}
                    </div>
                  </div>
                </div>
              )}
            </>
          )}

          {selectedTemplate && fileNeedsProcessing && (
            <DatePicker
              datePickerType="single"
              dateFormat="Y-m-d"
              maxDate={todayString}
              value={reportDate}
              onChange={(dates) => {
                const selectedDate = dates[0];

                setReportDateTouched(true);

                if (selectedDate) {
                  const localDate = new Date(
                    selectedDate.getTime() - selectedDate.getTimezoneOffset() * 60_000,
                  )
                    .toISOString()
                    .split("T")[0];

                  setReportDate(localDate);
                } else {
                  setReportDate("");
                }
              }}
            >
              <DatePickerInput
                id="report-date"
                labelText="Report date *"
                helperText={
                  reuploadMode === "same"
                    ? "Locked to the previous upload's report date."
                    : "Select the date this report covers. Cannot be a future date."
                }
                placeholder="YYYY-MM-DD"
                disabled={isUploading || reuploadMode === "same"}
                invalid={
                  (reportDateTouched && !reportDate) ||
                  (reuploadMode === "new" && !!reportDate && reportDate === previousReportDate)
                }
                invalidText={
                  reuploadMode === "new" && reportDate === previousReportDate
                    ? `Must differ from the previous report date (${previousReportDate}).`
                    : "Report date is required."
                }
              />
            </DatePicker>
          )}

          {selectedTemplate && fileNeedsProcessing && (
            <Select
              id="process-type"
              labelText="Process type"
              value={processType}
              onChange={(event) => setProcessType(event.target.value as DocumentProcessType | "")}
              disabled={isUploading}
            >
              <SelectItem value="" text="Select process type" />

              {DOCUMENT_PROCESS_TYPE_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value} text={option.label} />
              ))}
            </Select>
          )}

          {file && fileNeedsProcessing && !selectedTemplate && canProcessAdHoc && (
            <InlineNotification
              kind="info"
              title="Imported without a template"
              subtitle="No template selected, so column headers won't be validated. The file's rows will still be imported and available in the data preview."
              lowContrast
              hideCloseButton
            />
          )}

          {file && fileNeedsProcessing && !selectedTemplate && !canProcessAdHoc && (
            <InlineNotification
              kind="info"
              title="Stored without processing"
              subtitle="Without a template, this file will be stored as a completed document. Select a template to validate and import its data."
              lowContrast
              hideCloseButton
            />
          )}

          {!fileNeedsProcessing && file && (
            <InlineNotification
              kind="info"
              title="No processing job"
              subtitle="This file will be stored without creating a processing job."
              lowContrast
            />
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
            {isUploading && <InlineLoading description="Uploading document..." />}

            <Button kind="secondary" onClick={onClose} disabled={isUploading}>
              Cancel
            </Button>

            <Button onClick={() => void handleUpload()} disabled={isSubmitDisabled}>
              Upload Document
            </Button>
          </div>
        </Stack>
      </Form>
    </div>
  );
}

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  return (
    <PermissionGuard
      permission={PERMISSIONS.documentsWrite}
      fallback={
        <div className="document-upload-modal">
          <InlineNotification
            kind="warning"
            title="Access denied"
            subtitle="You do not have permission to upload documents."
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
      <UploadDocumentModalContent onClose={onClose} />
    </PermissionGuard>
  );
};
