import { useMemo, useState, type SyntheticEvent } from "react";
import Papa from "papaparse";
import * as XLSX from "xlsx";

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

import { useCreateDocumentMutation, useListStorageLocationsQuery } from "../../api";
import { PERMISSIONS, PermissionGuard } from "@moh-sso/auth";
import {
  ACCEPTED_EXTENSIONS,
  formatFileSize,
  getExtension,
  getSuggestedProcessType,
  isAcceptedFile,
  validateProcessTypeAgainstFile,
} from "@moh-sso/utils";
import { DOCUMENT_PROCESS_TYPE_OPTIONS, type DocumentProcessType } from "../../types";

import "./file-upload.scss";

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

      error: (error) => {
        reject(error);
      },
    });
  });
}

async function parseExcelFile(file: File): Promise<Row[]> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();

    reader.onerror = () => {
      reject(new Error("Failed to read Excel file."));
    };

    reader.onload = (event) => {
      try {
        const binary = event?.target?.result;

        if (!binary) {
          throw new Error("Invalid Excel file.");
        }

        const workbook = XLSX.read(binary, {
          type: "binary",
        });

        const firstSheetName = workbook.SheetNames[0];

        if (!firstSheetName) {
          throw new Error("Excel file has no sheets.");
        }

        const firstSheet = workbook.Sheets[firstSheetName];

        const rows = XLSX.utils
          .sheet_to_json<Row>(firstSheet, {
            defval: "",
          })
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

    if (!selectedFile) {
      return;
    }

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
    } catch (requestError: unknown) {
      const message =
        requestError instanceof Error
          ? requestError.message
          : "Failed to parse file. Please try again.";

      setFile(null);
      clearPreviewData();
      setProcessType("");
      setFileStatus("edit");
      setError(message);
    }
  };

  const handleFileSelection = (
    _event: SyntheticEvent<HTMLElement, Event>,
    {
      addedFiles,
    }: {
      addedFiles: File[];
    },
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
    } catch (requestError: unknown) {
      const message =
        typeof requestError === "object" &&
        requestError !== null &&
        "data" in requestError &&
        typeof (
          requestError as {
            data?: {
              message?: unknown;
            };
          }
        ).data?.message === "string"
          ? (
              requestError as {
                data?: {
                  message?: string;
                };
              }
            ).data?.message
          : "Upload failed. Please try again.";

      setError(message ?? "Upload failed. Please try again.");
    }
  };

  return (
    <PermissionGuard permission={PERMISSIONS.documentsWrite}>
      <>
        <FormItem className="uploader-form-item">
          <div className="file-upload__intro">
            <p className="cds--file--label">UPLOAD A FILE &amp; PREVIEW DATA</p>

            <p className="cds--label-description">
              Supported file types are .csv, .xlsx, and .xls.
            </p>
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
                disabled={isUploading || isLocationsLoading || !hasActiveLocations}
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
              <div className="file-upload__file-summary">
                <Tag type="blue">Selected file</Tag>

                <span>{file.name}</span>

                <span className="file-upload__file-meta">{formatFileSize(file.size)}</span>

                <span className="file-upload__file-meta">{parsedData.length} rows parsed</span>
              </div>
            )}

            <Select
              id="storage-location"
              labelText="Storage location"
              value={storageLocation}
              onChange={(event) => setStorageLocation(event.target.value)}
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
                onChange={(event) => setProcessType(event.target.value as DocumentProcessType)}
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
            <div className="file-upload-preview__header">
              <div>
                <p className="cds--file--label file-upload-preview__label">
                  PREVIEW: {file?.name}
                </p>

                <p className="cds--label-description">
                  Showing first {previewRows.length} rows of {parsedData.length} total rows.
                </p>
              </div>
            </div>

            <div className="file-upload-preview__table-scroll">
              <table className="file-upload-preview__table">
                <thead>
                  <tr>
                    {headers.map((header) => (
                      <th key={header}>
                        {header}
                      </th>
                    ))}
                  </tr>
                </thead>

                <tbody>
                  {previewRows.map((row, rowIndex) => (
                    <tr key={rowIndex}>
                      {headers.map((header) => (
                        <td key={`${rowIndex}-${header}`}>
                          {String(row[header] ?? "")}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="file-upload-btn-container">
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
    </PermissionGuard>
  );
};

export default FileUpload;
