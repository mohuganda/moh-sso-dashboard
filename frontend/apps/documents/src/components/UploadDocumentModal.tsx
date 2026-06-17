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
import { TrashCan } from "@carbon/icons-react";

import {
  useCreateDocumentMutation,
  useListDocumentsQuery,
  useListStorageLocationsQuery,
  useListActiveTemplatesQuery,
  useGetTemplateStructureQuery,
  useGetTemplateHasDataQuery,
} from "@moh-sso/api";
import {
  DOCUMENT_PROCESS_TYPE_OPTIONS,
  type DocumentProcessType,
} from "@moh-sso/types";
import {
  formatFileSize,
  getSuggestedProcessType,
  isAcceptedFile,
  isPdfFile,
  requiresProcessing,
  validateProcessTypeAgainstFile,
} from "@moh-sso/utils";

// ── Helpers ───────────────────────────────────────────────────────────────────

const CSV_HEADER_READ_BYTES = 64 * 1024;

function getFileExtension(fileName: string): string {
  const idx = fileName.lastIndexOf(".");
  return idx >= 0 ? fileName.slice(idx).toLowerCase() : "";
}

function normalizeForCompare(value: string): string {
  return value.trim().toLowerCase().replace(/\s+/g, "_");
}

async function extractFileHeaders(file: File): Promise<string[]> {
  const ext = getFileExtension(file.name);

  if (ext === ".csv") {
    const { default: Papa } = await import("papaparse");
    const text = await file.slice(0, CSV_HEADER_READ_BYTES).text();
    return new Promise((resolve, reject) => {
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
  }

  if (ext === ".xlsx" || ext === ".xls") {
    const XLSX = await import("xlsx");
    const buffer = await file.arrayBuffer();
    const workbook = XLSX.read(buffer, { type: "array", sheetRows: 10 });

    const visibleNames = workbook.SheetNames.filter((_, idx) => {
      const meta = workbook.Workbook?.Sheets?.[idx];
      return !meta?.Hidden;
    });

    const allHeaders: string[] = [];
    for (const sheetName of visibleNames) {
      const ws = workbook.Sheets[sheetName];
      const rows = XLSX.utils.sheet_to_json<(string | number | null)[]>(ws, {
        header: 1,
        defval: "",
        blankrows: false,
      });

      // Find the first row with at least 4 non-empty cells (avoids picking up legend/key rows)
      let headerRowIdx = 0;
      for (let i = 0; i < rows.length; i++) {
        if (rows[i].filter((cell) => String(cell ?? "").trim() !== "").length >= 4) {
          headerRowIdx = i;
          break;
        }
      }

      const headerRow = Array.isArray(rows[headerRowIdx]) ? rows[headerRowIdx] : [];
      allHeaders.push(...headerRow.map((h) => String(h ?? "").trim()).filter(Boolean));
    }
    return allHeaders;
  }

  return [];
}

// ── Types ─────────────────────────────────────────────────────────────────────

export type UploadDocumentModalProps = {
  onClose: () => void;
};

type ColumnValidationResult = {
  required: string[];
  present: string[];
  missing: string[];
};

// ── Component ─────────────────────────────────────────────────────────────────

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [fileHeaders, setFileHeaders] = useState<string[]>([]);
  const [storageLocation, setStorageLocation] = useState("");
  const [processType, setProcessType] = useState<DocumentProcessType | "">("");
  const [selectedTemplate, setSelectedTemplate] = useState("");
  const [reportDate, setReportDate] = useState<string>("");
  const [reportDateTouched, setReportDateTouched] = useState(false);
  const [isReadingFile, setIsReadingFile] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Re-upload state
  const [reuploadMode, setReuploadMode] = useState<"same" | "new" | null>(null);

  const todayStr = useMemo(() => new Date().toISOString().split("T")[0], []);

  const [createDocument, { isLoading: isUploading }] = useCreateDocumentMutation();

  const { data: allDocuments = [] } = useListDocumentsQuery();

  const { data: savedTemplates = [], isLoading: isTemplatesLoading } =
    useListActiveTemplatesQuery();

  const { data: templateStructure, isFetching: isFetchingStructure } =
    useGetTemplateStructureQuery(selectedTemplate, { skip: !selectedTemplate });

  const { data: hasDataResult } = useGetTemplateHasDataQuery(selectedTemplate, {
    skip: !selectedTemplate || !processType,
  });

  const {
    data: locations = [],
    isLoading: isLocationsLoading,
    isError: isLocationsError,
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

  // Detect a previous upload of the same filename (most recent match)
  const previousUpload = useMemo(() => {
    if (!file) return null;
    const matches = allDocuments
      .filter((d) => d.original_filename === file.name && d.metadata?.template_code)
      .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
    return matches[0] ?? null;
  }, [file, allDocuments]);

  const previousReportDate = previousUpload?.metadata?.report_date ?? null;

  // When a previous upload is detected, default to "same" mode and pre-fill the date
  useEffect(() => {
    if (previousUpload && previousReportDate) {
      setReuploadMode("same");
      setReportDate(previousReportDate);
      setReportDateTouched(false);
    } else if (!previousUpload) {
      setReuploadMode(null);
    }
  }, [previousUpload?.id, previousReportDate]);

  const fileNeedsProcessing = useMemo(() => (file ? requiresProcessing(file) : false), [file]);
  const isPdf = useMemo(() => (file ? isPdfFile(file) : false), [file]);

  // ── Required-column validation ────────────────────────────────────────────

  const columnValidation = useMemo<ColumnValidationResult | null>(() => {
    if (!selectedTemplate || !templateStructure || !file || isPdf) return null;

    const requiredColumns = templateStructure.sheets.flatMap((sheet) =>
      sheet.columns.filter((col) => col.required).map((col) => col.column_name),
    );

    if (requiredColumns.length === 0) return null;

    const fileHeaderSet = new Set(fileHeaders.map(normalizeForCompare));

    const present: string[] = [];
    const missing: string[] = [];

    for (const col of requiredColumns) {
      if (fileHeaderSet.has(normalizeForCompare(col))) {
        present.push(col);
      } else {
        missing.push(col);
      }
    }

    return { required: requiredColumns, present, missing };
  }, [selectedTemplate, templateStructure, file, isPdf, fileHeaders]);

  const isTemplateValid = columnValidation === null || columnValidation.missing.length === 0;

  // ── File selection ────────────────────────────────────────────────────────

  const handleFileChange = useCallback(
    async (_event: SyntheticEvent<HTMLElement, Event>, { addedFiles }: { addedFiles: File[] }) => {
      const selected = addedFiles[0];
      if (!selected) return;

      if (!isAcceptedFile(selected)) {
        setFile(null);
        setFileHeaders([]);
        setError("Only PDF, CSV, or Excel (.pdf, .csv, .xlsx, .xls) files are allowed.");
        return;
      }

      setError(null);
      setFile(selected);
      setFileHeaders([]);

      if (requiresProcessing(selected)) {
        setProcessType((cur) => cur || getSuggestedProcessType(selected));
      } else {
        setProcessType("");
      }

      if (isPdfFile(selected)) return;

      setIsReadingFile(true);
      try {
        const headers = await extractFileHeaders(selected);
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

  // ── Submit ────────────────────────────────────────────────────────────────

  const handleUpload = async () => {
    if (!file) return setError("Please select a file.");
    if (!storageLocation) return setError("Please select a storage location.");

    if (fileNeedsProcessing && !processType) return setError("Please select a process type.");

    if (fileNeedsProcessing && processType) {
      const err = validateProcessTypeAgainstFile(file, processType);
      if (err) return setError(err);
    }

    if (!isTemplateValid) {
      return setError(
        `Missing required columns: ${columnValidation!.missing.join(", ")}`,
      );
    }

    if (selectedTemplate && fileNeedsProcessing && !reportDate) {
      setReportDateTouched(true);
      return setError("Please select a report date.");
    }

    if (reuploadMode === "new" && previousReportDate && reportDate === previousReportDate) {
      setReportDateTouched(true);
      return setError(`Report date cannot be the same as the previous upload (${previousReportDate}). Choose a different date or select "Yes, same report date" to replace it.`);
    }

    setError(null);

    const isReplace = !!(previousUpload && reuploadMode === "same");

    const metadata = selectedTemplate
      ? {
          template_code: selectedTemplate,
          ...(reportDate ? { report_date: reportDate } : {}),
          ...(isReplace ? { replace_document_id: previousUpload!.id } : {}),
        }
      : undefined;

    try {
      await createDocument({
        file,
        storageLocation,
        ...(fileNeedsProcessing && processType ? { processType } : {}),
        ...(metadata ? { metadata } : {}),
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
    isUploading ||
    isReadingFile ||
    isFetchingStructure ||
    !file ||
    !storageLocation ||
    (fileNeedsProcessing && !processType) ||
    (selectedTemplate && fileNeedsProcessing && !reportDate) ||
    (reuploadMode === "new" && !!previousReportDate && reportDate === previousReportDate) ||
    !isTemplateValid;

  // ── Render ────────────────────────────────────────────────────────────────

  return (
    <div style={{ maxWidth: 720 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Document</h2>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload a file for processing. Optionally select a template to validate the file's columns
          before uploading.
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

          {/* ── File drop ── */}
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
            <div style={{ display: "flex", alignItems: "center", gap: "0.75rem", flexWrap: "wrap" }}>
              <Tag type="blue">Selected</Tag>
              <span style={{ fontWeight: 500 }}>{file.name}</span>
              <span style={{ color: "#6f6f6f" }}>{formatFileSize(file.size)}</span>
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

          {/* ── Re-upload prompt ── */}
          {previousUpload && fileNeedsProcessing && (
            <div style={{
              border: "1px solid #f1c21b", borderRadius: 4,
              background: "#fdf6dd", padding: "0.875rem 1rem",
            }}>
              <p style={{ margin: "0 0 0.6rem", fontWeight: 600, fontSize: "0.875rem", color: "#161616" }}>
                This filename was uploaded before
              </p>
              <p style={{ margin: "0 0 0.75rem", fontSize: "0.8125rem", color: "#525252" }}>
                Previous report date: <strong>{previousReportDate ?? "—"}</strong>
              </p>
              <div style={{ display: "flex", flexDirection: "column", gap: "0.4rem" }}>
                <label style={{ display: "flex", alignItems: "flex-start", gap: "0.5rem", cursor: "pointer" }}>
                  <input
                    type="radio"
                    name="reupload-mode"
                    value="same"
                    checked={reuploadMode === "same" || reuploadMode === null}
                    onChange={() => {
                      setReuploadMode("same");
                      if (previousReportDate) setReportDate(previousReportDate);
                    }}
                    style={{ marginTop: 3 }}
                  />
                  <span style={{ fontSize: "0.8125rem", color: "#161616" }}>
                    <strong>Yes, same report date</strong> — replace the previous upload
                    {previousReportDate && (
                      <span style={{ color: "#525252" }}> ({previousReportDate})</span>
                    )}
                  </span>
                </label>
                <label style={{ display: "flex", alignItems: "flex-start", gap: "0.5rem", cursor: "pointer" }}>
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
                    style={{ marginTop: 3 }}
                  />
                  <span style={{ fontSize: "0.8125rem", color: "#161616" }}>
                    <strong>No, different report date</strong> — add as a new upload
                  </span>
                </label>
              </div>
            </div>
          )}

          {/* ── Template selector ── */}
          <Select
            id="saved-template"
            labelText="Validate against template (optional)"
            helperText={
              isPdf
                ? "Template validation is not applicable for PDF files."
                : "Select a template to check that all required columns are present."
            }
            value={selectedTemplate}
            onChange={(e) => {
              setSelectedTemplate(e.target.value);
              setError(null);
            }}
            disabled={!file || isUploading || isReadingFile || isPdf}
          >
            <SelectItem
              value=""
              text={isTemplatesLoading ? "Loading templates..." : "No template (optional)"}
            />
            {savedTemplates.map((t) => (
              <SelectItem key={t.id} value={t.code} text={`${t.name} (${t.code})`} />
            ))}
          </Select>

          {/* ── Column validation feedback ── */}
          {selectedTemplate && (
            <>
              {isFetchingStructure && (
                <InlineLoading description="Loading template columns..." />
              )}

              {!isFetchingStructure && templateStructure && columnValidation === null && (
                <InlineNotification
                  kind="info"
                  title="No required columns"
                  subtitle="This template has no required columns defined."
                  lowContrast
                />
              )}

              {!isFetchingStructure && columnValidation !== null && fileHeaders.length === 0 && !isReadingFile && (
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
                      subtitle={`${columnValidation.present.length} required column${columnValidation.present.length !== 1 ? "s" : ""} validated successfully.`}
                      lowContrast
                    />
                  )}

                  {/* Required column checklist */}
                  <div style={{ marginTop: "0.75rem" }}>
                    <p style={{ margin: "0 0 0.5rem", fontSize: "0.875rem", fontWeight: 600 }}>
                      Required columns
                    </p>
                    <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
                      {columnValidation.required.map((col) => {
                        const isPresent = columnValidation.present.includes(col);
                        return (
                          <span
                            key={col}
                            style={{
                              display: "inline-flex",
                              alignItems: "center",
                              gap: "0.25rem",
                              padding: "0.2rem 0.6rem",
                              borderRadius: "1rem",
                              fontSize: "0.8rem",
                              fontFamily: "monospace",
                              backgroundColor: isPresent ? "#defbe6" : "#fff1f1",
                              color: isPresent ? "#198038" : "#da1e28",
                              border: `1px solid ${isPresent ? "#a7f0ba" : "#ffd7d9"}`,
                            }}
                          >
                            {isPresent ? "✓" : "✗"} {col}
                          </span>
                        );
                      })}
                    </div>
                  </div>
                </div>
              )}
            </>
          )}

          {/* ── Report date ── */}
          {selectedTemplate && fileNeedsProcessing && (
            <DatePicker
              datePickerType="single"
              dateFormat="Y-m-d"
              maxDate={todayStr}
              value={reportDate}
              onChange={(dates) => {
                const d = dates[0];
                setReportDateTouched(true);
                if (d) {
                  const iso = new Date(d.getTime() - d.getTimezoneOffset() * 60000)
                    .toISOString()
                    .split("T")[0];
                  setReportDate(iso);
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

          {/* ── Process type ── */}
          {fileNeedsProcessing && (
            <Select
              id="process-type"
              labelText="Process type"
              value={processType}
              onChange={(e) => setProcessType(e.target.value as DocumentProcessType | "")}
              disabled={isUploading}
            >
              <SelectItem value="" text="Select process type" />
              {DOCUMENT_PROCESS_TYPE_OPTIONS.map((opt) => (
                <SelectItem key={opt.value} value={opt.value} text={opt.label} />
              ))}
            </Select>
          )}

          {!fileNeedsProcessing && file && (
            <InlineNotification
              kind="info"
              title="No processing job"
              subtitle="This file will be stored without creating a processing job."
              lowContrast
            />
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
};
