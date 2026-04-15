import { useMemo, useState, type SyntheticEvent } from "react";
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
} from "@carbon/react";
import * as XLSX from "xlsx";
import Papa from "papaparse";

import {
  useCreateDocumentMutation,
  useListStorageLocationsQuery,
} from "../../../../store/api/document.api";
import {
  DOCUMENT_PROCESS_TYPE_OPTIONS,
  type DocumentProcessType,
} from "../../../../store/types/documents.types";
import {
  formatFileSize,
  getSuggestedProcessType,
  isAcceptedFile,
  isPdfFile,
  requiresProcessing,
  validateProcessTypeAgainstFile,
} from "../../../../utils/utils";
import {
  type StandardTemplate,
  type StandardTemplateKey,
  STANDARD_TEMPLATES,
} from "../../../../lib/constants/document-templates";

type UploadDocumentModalProps = {
  onClose: () => void;
};

function getFileExtension(fileName: string): string {
  const index = fileName.lastIndexOf(".");
  return index >= 0 ? fileName.slice(index).toLowerCase() : "";
}

function normalizeHeader(value: unknown): string {
  return String(value ?? "")
    .trim()
    .toLowerCase()
    .replace(/\s+/g, "_");
}

async function extractHeadersFromFile(file: File): Promise<string[]> {
  const extension = getFileExtension(file.name);

  if (extension === ".csv") {
    const text = await file.text();

    return new Promise((resolve, reject) => {
      Papa.parse<Record<string, unknown>>(text, {
        header: true,
        skipEmptyLines: true,
        preview: 1,
        complete: (results) => {
          const headers = (results.meta.fields ?? []).map(normalizeHeader);
          resolve(headers);
        },
        error: (error) => reject(error),
      });
    });
  }

  if (extension === ".xlsx" || extension === ".xls") {
    const buffer = await file.arrayBuffer();
    const workbook = XLSX.read(buffer, { type: "array" });
    const firstSheetName = workbook.SheetNames[0];

    if (!firstSheetName) {
      return [];
    }

    const worksheet = workbook.Sheets[firstSheetName];
    const rows = XLSX.utils.sheet_to_json<(string | number | null)[]>(worksheet, {
      header: 1,
      defval: "",
      blankrows: false,
    });

    const firstRow = Array.isArray(rows[0]) ? rows[0] : [];
    return firstRow.map(normalizeHeader).filter(Boolean);
  }

  return [];
}

async function validateFileAgainstTemplate(
  file: File,
  template: StandardTemplate,
): Promise<string | null> {
  const extension = getFileExtension(file.name);

  if (!template.allowedExtensions.includes(extension)) {
    return `${template.label} only supports ${template.allowedExtensions.join(", ")} files.`;
  }

  const headers = await extractHeadersFromFile(file);

  if (!headers.length) {
    return "Could not read file headers. Please check the file format and try again.";
  }

  const missingHeaders = template.requiredHeaders.filter(
    (requiredHeader) => !headers.includes(normalizeHeader(requiredHeader)),
  );

  if (missingHeaders.length > 0) {
    return `The uploaded file does not match the selected template. Missing required columns: ${missingHeaders.join(", ")}.`;
  }

  return null;
}

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [storageLocation, setStorageLocation] = useState("");
  const [processType, setProcessType] = useState<DocumentProcessType | "">("");
  const [selectedTemplate, setSelectedTemplate] = useState<StandardTemplateKey>("");
  const [templateValidationMessage, setTemplateValidationMessage] = useState<string | null>(null);
  const [isValidatingTemplate, setIsValidatingTemplate] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [createDocument, { isLoading: isUploading }] = useCreateDocumentMutation();

  const {
    data: locations = [],
    isLoading: isLocationsLoading,
    isError: isLocationsError,
  } = useListStorageLocationsQuery();

  const activeLocations = useMemo(
    () => locations.filter((location) => location.is_active),
    [locations],
  );

  const hasStorageLocations = activeLocations.length > 0;
  const fileNeedsProcessing = file ? requiresProcessing(file) : false;
  const isPdf = file ? isPdfFile(file) : false;

  const selectedTemplateConfig = useMemo(
    () => STANDARD_TEMPLATES.find((item) => item.key === selectedTemplate),
    [selectedTemplate],
  );

  const processSelectedFiles = async (addedFiles: File[]) => {
    const selectedFile = addedFiles[0];
    if (!selectedFile) return;

    setTemplateValidationMessage(null);

    if (!isAcceptedFile(selectedFile)) {
      setFile(null);
      setProcessType("");
      setSelectedTemplate("");
      setError("Only PDF, CSV, or Excel (.pdf, .csv, .xlsx, .xls) files are allowed.");
      return;
    }

    setError(null);
    setFile(selectedFile);

    if (requiresProcessing(selectedFile)) {
      setProcessType((current) => current || getSuggestedProcessType(selectedFile));
    } else {
      setProcessType("");
    }
  };

  const handleFileChange = (
    _event: SyntheticEvent<HTMLElement, Event>,
    { addedFiles }: { addedFiles: File[] },
  ): void => {
    void processSelectedFiles(addedFiles);
  };

  const handleProcessTypeChange = (value: string) => {
    setProcessType(value as DocumentProcessType | "");
  };

  const handleTemplateChange = async (templateKey: string) => {
    const nextTemplate = templateKey as StandardTemplateKey;

    setSelectedTemplate(nextTemplate);
    setTemplateValidationMessage(null);
    setError(null);

    const template = STANDARD_TEMPLATES.find((item) => item.key === nextTemplate);

    if (template?.suggestedProcessType && file && requiresProcessing(file)) {
      setProcessType((current) => current || template.suggestedProcessType!);
    }

    if (!file || !template) return;
    if (isPdfFile(file)) return;

    try {
      setIsValidatingTemplate(true);
      const validationError = await validateFileAgainstTemplate(file, template);

      if (validationError) {
        setTemplateValidationMessage(validationError);
      } else {
        setTemplateValidationMessage(`File matches the ${template.label}.`);
      }
    } catch {
      setTemplateValidationMessage("Failed to validate file against the selected template.");
    } finally {
      setIsValidatingTemplate(false);
    }
  };

  const handleUpload = async () => {
    if (!file) {
      setError("Please select a file to upload.");
      return;
    }

    if (!storageLocation) {
      setError("Please select a storage location.");
      return;
    }

    if (requiresProcessing(file) && !processType) {
      setError("Please select a process type.");
      return;
    }

    if (requiresProcessing(file) && processType) {
      const validationError = validateProcessTypeAgainstFile(file, processType);
      if (validationError) {
        setError(validationError);
        return;
      }
    }

    if (selectedTemplateConfig && !isPdfFile(file)) {
      try {
        setIsValidatingTemplate(true);
        const templateError = await validateFileAgainstTemplate(file, selectedTemplateConfig);
        if (templateError) {
          setError(templateError);
          return;
        }
      } catch {
        setError("Failed to validate the uploaded file against the selected template.");
        return;
      } finally {
        setIsValidatingTemplate(false);
      }
    }

    try {
      setError(null);

      await createDocument({
        file,
        storageLocation,
        ...(requiresProcessing(file) && processType ? { processType } : {}),
      }).unwrap();

      onClose();
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

  const isTemplateInvalid =
    Boolean(selectedTemplateConfig) &&
    Boolean(templateValidationMessage) &&
    !templateValidationMessage?.startsWith("File matches");

  const isSubmitDisabled =
    isUploading ||
    isLocationsLoading ||
    isValidatingTemplate ||
    !file ||
    !storageLocation ||
    (fileNeedsProcessing && !processType) ||
    !hasStorageLocations ||
    isTemplateInvalid;

  return (
    <div style={{ maxWidth: 720 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Document</h2>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload a PDF, CSV, or Excel file, select where it should be stored, choose a process type,
          and optionally validate it against a standard template before upload.
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

          {!isLocationsLoading && !hasStorageLocations && (
            <InlineNotification
              kind="warning"
              title="No active storage locations"
              subtitle="Activate or configure a storage location before uploading documents."
              lowContrast
            />
          )}

          {isLocationsError && (
            <InlineNotification
              kind="error"
              title="Storage locations error"
              subtitle="Failed to load storage locations."
              lowContrast
            />
          )}

          <FormGroup legendText="Document file">
            <FileUploaderDropContainer
              labelText="Drag and drop a file here or click to browse"
              accept={[".pdf", ".csv", ".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isUploading || isValidatingTemplate}
            />
          </FormGroup>

          {file && (
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "0.75rem",
                flexWrap: "wrap",
              }}
            >
              <Tag type="blue">Selected file</Tag>
              <span>{file.name}</span>
              <span style={{ color: "#6f6f6f" }}>{formatFileSize(file.size)}</span>
              {isPdf ? (
                <Tag type="cool-gray">No processing required</Tag>
              ) : (
                <Tag type="purple">Processing required</Tag>
              )}
            </div>
          )}

          <Select
            id="standard-template"
            labelText="Standard template"
            value={selectedTemplate}
            onChange={(e) => void handleTemplateChange(e.target.value)}
            disabled={!file || isUploading || isValidatingTemplate || isPdf}
          >
            <SelectItem
              value=""
              text={isPdf ? "Templates not applicable for PDF" : "Select template (optional)"}
            />
            {STANDARD_TEMPLATES.map((template) => (
              <SelectItem key={template.key} value={template.key} text={template.label} />
            ))}
          </Select>

          {selectedTemplateConfig && (
            <div>
              <div style={{ marginBottom: "0.5rem", color: "#6f6f6f" }}>
                Required columns: {selectedTemplateConfig.requiredHeaders.join(", ")}
              </div>

              {isValidatingTemplate && (
                <InlineLoading description="Validating file against template..." />
              )}

              {!isValidatingTemplate && templateValidationMessage && (
                <InlineNotification
                  kind={isTemplateInvalid ? "error" : "success"}
                  title={
                    isTemplateInvalid
                      ? "Template validation failed"
                      : "Template validation successful"
                  }
                  subtitle={templateValidationMessage}
                  lowContrast
                  onCloseButtonClick={() => setTemplateValidationMessage(null)}
                />
              )}
            </div>
          )}

          {fileNeedsProcessing ? (
            <Select
              id="process-type"
              labelText="Process type"
              value={processType}
              onChange={(e) => handleProcessTypeChange(e.target.value)}
              disabled={isUploading || isValidatingTemplate}
            >
              <SelectItem value="" text="Select process type" />
              {DOCUMENT_PROCESS_TYPE_OPTIONS.map((option) => (
                <SelectItem key={option.value} value={option.value} text={option.label} />
              ))}
            </Select>
          ) : (
            file && (
              <InlineNotification
                kind="info"
                title="PDF upload"
                subtitle="This file will be uploaded and stored without any processing job."
                lowContrast
              />
            )
          )}

          <Select
            id="storage-location"
            labelText="Storage location"
            value={storageLocation}
            onChange={(e) => setStorageLocation(e.target.value)}
            disabled={
              isLocationsLoading || !hasStorageLocations || isUploading || isValidatingTemplate
            }
          >
            <SelectItem
              value=""
              text={
                isLocationsLoading
                  ? "Loading locations..."
                  : hasStorageLocations
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

          <div
            style={{
              display: "flex",
              justifyContent: "flex-end",
              alignItems: "center",
              gap: "1rem",
              marginTop: "0.5rem",
              flexWrap: "wrap",
            }}
          >
            {(isUploading || isValidatingTemplate) && (
              <InlineLoading
                description={isUploading ? "Uploading document..." : "Validating template..."}
              />
            )}

            <Button
              kind="secondary"
              onClick={onClose}
              disabled={isUploading || isValidatingTemplate}
            >
              Cancel
            </Button>

            <Button onClick={handleUpload} disabled={isSubmitDisabled}>
              Upload Document
            </Button>
          </div>
        </Stack>
      </Form>
    </div>
  );
};
