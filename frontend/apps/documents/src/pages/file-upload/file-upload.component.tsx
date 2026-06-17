import { useMemo, useState, type SyntheticEvent } from "react";
import Papa from "papaparse";
import * as XLSX from "xlsx";
import "./file-upload.scss";
import {
  Button,
  FileUploaderDropContainer,
  FileUploaderItem,
  FormItem,
  InlineLoading,
  InlineNotification,
  Select,
  SelectItem,
  Stack,
  Tag,
} from "@carbon/react";
import type { FileUploaderItemProps } from "@carbon/react";
import { useCreateDocumentMutation, useListStorageLocationsQuery } from "@moh-sso/api";
import {
  ACCEPTED_EXTENSIONS,
  formatFileSize,
  getExtension,
  getSuggestedProcessType,
  isAcceptedFile,
  validateProcessTypeAgainstFile,
} from "@moh-sso/utils";
import { DOCUMENT_PROCESS_TYPE_OPTIONS, type DocumentProcessType } from "@moh-sso/types";

type Row = Record<string, unknown>;
type FileItemStatus = NonNullable<FileUploaderItemProps["status"]>;

export function getPreviewHeaders(rows: Row[]): string[] {
  return Object.keys(rows[0] || {});
}

async function parseCsvFile(file: File): Promise<Row[]> {
  return new Promise((resolve, reject) => {
    Papa.parse<Row>(file, {
      header: true,
      skipEmptyLines: true,
      complete: (results) => {
        const rows = (results.data || []).filter((row) =>
          Object.values(row).some((value) => String(value ?? "").trim() !== ""),
        );

        if (!rows.length) {
          reject(new Error("CSV file contains no rows."));
          return;
        }

        resolve(rows);
      },
      error: (error) => reject(error),
    });
  });
}

async function parseExcelFile(file: File): Promise<Row[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onerror = () => reject(new Error("Failed to read Excel file."));

    reader.onload = (event) => {
      try {
        const binary = event?.target?.result;
        if (!binary) {
          throw new Error("Invalid Excel file.");
        }

        const workbook = XLSX.read(binary, { type: "binary" });
        const firstSheetName = workbook.SheetNames[0];

        if (!firstSheetName) {
          throw new Error("Excel file has no sheets.");
        }

        const firstSheet = workbook.Sheets[firstSheetName];
        const rows = XLSX.utils
          .sheet_to_json<Row>(firstSheet, { defval: "" })
          .filter((row) => Object.values(row).some((value) => String(value ?? "").trim() !== ""));

        if (!rows.length) {
          throw new Error("Excel sheet contains no rows.");
        }

        resolve(rows);
      } catch (error) {
        reject(error);
      }
    };

    reader.readAsBinaryString(file);
  });
}

async function parseFile(file: File): Promise<Row[]> {
  const extension = getExtension(file.name);

  if (extension === ".csv") {
    return parseCsvFile(file);
  }

  return parseExcelFile(file);
}

const FileUpload = () => {
  const [file, setFile] = useState<File | null>(null);
  const [fileStatus, setFileStatus] = useState<FileItemStatus>("edit");
  const [parsedData, setParsedData] = useState<Row[]>([]);
  const [headers, setHeaders] = useState<string[]>([]);
  const [storageLocation, setStorageLocation] = useState("");
  const [processType, setProcessType] = useState<DocumentProcessType | "">("");
  const [error, setError] = useState<string | null>(null);

  const {
    data: locations = [],
    isLoading: isLocationsLoading,
    isError: isLocationsError,
  } = useListStorageLocationsQuery();

  const [createDocument, { isLoading: isUploading }] = useCreateDocumentMutation();

  const activeLocations = useMemo(
    () => locations.filter((location) => location.is_active),
    [locations],
  );

  const hasActiveLocations = activeLocations.length > 0;
  const previewRows = useMemo(() => parsedData.slice(0, 10), [parsedData]);

  const resetUploader = () => {
    setFile(null);
    setFileStatus("edit");
    setParsedData([]);
    setHeaders([]);
    setStorageLocation("");
    setProcessType("");
    setError(null);
  };

  const clearPreviewData = () => {
    setParsedData([]);
    setHeaders([]);
  };

  const processSelectedFiles = async (addedFiles: File[]) => {
    const selectedFile = addedFiles[0];
    if (!selectedFile) return;

    setError(null);
    setFileStatus("uploading");

    if (!isAcceptedFile(selectedFile)) {
      setFile(null);
      clearPreviewData();
      setProcessType("");
      setFileStatus("edit");
      setError("Please upload a CSV or Excel file (.csv, .xlsx, .xls).");
      return;
    }

    try {
      const rows = await parseFile(selectedFile);

      setFile(selectedFile);
      setParsedData(rows);
      setHeaders(getPreviewHeaders(rows));
      setProcessType((current) => current || getSuggestedProcessType(selectedFile));
      setFileStatus("complete");

      window.setTimeout(() => {
        setFileStatus("edit");
      }, 1200);
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : "Failed to parse file. Please try again.";

      setFile(null);
      clearPreviewData();
      setProcessType("");
      setFileStatus("edit");
      setError(message);
    }
  };

  const handleFileSelection = (
    _event: SyntheticEvent<HTMLElement, Event>,
    { addedFiles }: { addedFiles: File[] },
  ): void => {
    void processSelectedFiles(addedFiles);
  };

  const handleUploadToBackend = async () => {
    if (!file) {
      setError("Please select a file first.");
      return;
    }

    if (!storageLocation) {
      setError("Please select a storage location.");
      return;
    }

    if (!processType) {
      setError("Please select a process type.");
      return;
    }

    const validationError = validateProcessTypeAgainstFile(file, processType);
    if (validationError) {
      setError(validationError);
      return;
    }

    try {
      setError(null);

      await createDocument({
        file,
        storageLocation,
        processType,
      }).unwrap();
      resetUploader();
    } catch (err: unknown) {
      const message =
        typeof err === "object" &&
        err !== null &&
        "data" in err &&
        typeof (err as { data?: { message?: unknown } }).data?.message === "string"
          ? (err as { data?: { message?: string } }).data?.message
          : "Upload failed. Please try again.";

      setError(message!);
    }
  };

  return (
    <>
      <FormItem className="uploader-form-item">
        <div style={{ marginBottom: "1rem" }}>
          <p className="cds--file--label">UPLOAD A FILE &amp; PREVIEW DATA</p>
          <p className="cds--label-description">Supported file types are .csv, .xlsx, and .xls.</p>
        </div>

        <Stack gap={5}>
          {error && (
            <InlineNotification
              kind="error"
              title="Upload error"
              subtitle={error}
              lowContrast
              onCloseButtonClick={() => setError(null)}
            />
          )}

          {!isLocationsLoading && !hasActiveLocations && (
            <InlineNotification
              kind="warning"
              title="No active storage locations"
              subtitle="Activate or configure a storage location before uploading a file."
              lowContrast
            />
          )}

          {!file ? (
            <FileUploaderDropContainer
              accept={[...ACCEPTED_EXTENSIONS]}
              labelText="Drag and drop a file here or click to upload"
              onAddFiles={handleFileSelection}
              disabled={isUploading}
            />
          ) : (
            <FileUploaderItem
              name={file.name}
              size="lg"
              status={fileStatus}
              iconDescription="Remove file"
              onDelete={resetUploader}
            />
          )}

          {file && (
            <div
              style={{
                display: "flex",
                gap: "0.75rem",
                alignItems: "center",
                flexWrap: "wrap",
              }}
            >
              <Tag type="blue">Selected file</Tag>
              <span>{file.name}</span>
              <span style={{ color: "#6f6f6f" }}>{formatFileSize(file.size)}</span>
              <span style={{ color: "#6f6f6f" }}>{parsedData.length} rows parsed</span>
            </div>
          )}

          <Select
            id="storage-location"
            labelText="Storage location"
            value={storageLocation}
            onChange={(e) => setStorageLocation(e.target.value)}
            disabled={isLocationsLoading || !hasActiveLocations || isUploading}
          >
            <SelectItem
              value=""
              text={
                isLocationsLoading
                  ? "Loading locations..."
                  : hasActiveLocations
                    ? "Select storage location"
                    : "No active locations available"
              }
            />
            {activeLocations.map((location) => (
              <SelectItem
                key={location.id}
                value={location.id}
                text={`${location.name} (${location.provider.toUpperCase()})`}
              />
            ))}
          </Select>

          {file && (
            <Select
              id="document-process-type"
              labelText="Process type"
              value={processType}
              onChange={(e) => setProcessType(e.target.value as DocumentProcessType)}
              disabled={isUploading}
            >
              <SelectItem value="" text="Select process type" />
              {DOCUMENT_PROCESS_TYPE_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value} text={option.label} />
              ))}
            </Select>
          )}

          {isLocationsError && (
            <InlineNotification
              kind="error"
              title="Storage locations error"
              subtitle="Failed to load storage locations."
              lowContrast
            />
          )}
        </Stack>
      </FormItem>

      {parsedData.length > 0 && (
        <div className="preview-container">
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "flex-start",
              gap: "1rem",
              flexWrap: "wrap",
              marginBottom: "1rem",
            }}
          >
            <div>
              <p className="cds--file--label" style={{ marginBottom: "0.5rem" }}>
                PREVIEW: {file?.name}
              </p>
              <p className="cds--label-description">
                Showing first {previewRows.length} rows of {parsedData.length} total rows.
              </p>
            </div>
          </div>

          <div
            style={{
              overflowX: "auto",
              maxHeight: 400,
              border: "1px solid #e0e0e0",
              borderRadius: 4,
              background: "#fff",
            }}
          >
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead
                style={{
                  backgroundColor: "#f4f4f4",
                  position: "sticky",
                  top: 0,
                  zIndex: 1,
                }}
              >
                <tr>
                  {headers.map((header) => (
                    <th
                      key={header}
                      style={{
                        borderBottom: "1px solid #e0e0e0",
                        padding: 8,
                        textAlign: "left",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {header}
                    </th>
                  ))}
                </tr>
              </thead>

              <tbody>
                {previewRows.map((row, rowIndex) => (
                  <tr key={rowIndex}>
                    {headers.map((header) => (
                      <td
                        key={`${rowIndex}-${header}`}
                        style={{
                          borderBottom: "1px solid #f0f0f0",
                          padding: 8,
                          verticalAlign: "top",
                        }}
                      >
                        {String(row[header] ?? "")}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div
            className="file-upload-btn-container"
            style={{
              display: "flex",
              justifyContent: "flex-end",
              alignItems: "center",
              gap: "1rem",
              marginTop: "1rem",
              flexWrap: "wrap",
            }}
          >
            {isUploading && <InlineLoading description="Uploading to backend..." />}

            <Button kind="secondary" onClick={resetUploader} disabled={isUploading}>
              Cancel
            </Button>

            <Button
              kind="primary"
              onClick={handleUploadToBackend}
              disabled={
                isUploading ||
                isLocationsLoading ||
                !storageLocation ||
                !processType ||
                !file ||
                !hasActiveLocations
              }
            >
              Upload to Backend
            </Button>
          </div>
        </div>
      )}
    </>
  );
};

export default FileUpload;
