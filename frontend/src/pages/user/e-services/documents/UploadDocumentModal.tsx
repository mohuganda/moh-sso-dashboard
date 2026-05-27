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
import * as XLSX from "xlsx";
import Papa from "papaparse";

import {
  useCreateDocumentMutation,
  useListStorageLocationsQuery,
} from "../../../../store/api/document.api";

import {
  useCreateTemplateStructureMutation,
  useListActiveTemplatesQuery,
} from "../../../../store/api/document_template.api";

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

type UploadMode = "document" | "template";

type UploadDocumentModalProps = {
  onClose: () => void;
};

type TemplateValidationConfig = {
  label: string;
  allowedExtensions: string[];
  requiredHeaders: string[];
};

type DetectedSheet = {
  name: string;
  headers: string[];
};

const CSV_HEADER_READ_BYTES = 64 * 1024;

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

function normalizeTemplateCode(value: string): string {
  return value
    .trim()
    .toUpperCase()
    .replace(/\s+/g, "_")
    .replace(/[^A-Z0-9_]/g, "");
}

function getStringArrayFromConfig(
  configuration: Record<string, unknown> | null | undefined,
  keys: string[],
): string[] {
  if (!configuration) return [];

  for (const key of keys) {
    const value = configuration[key];

    if (Array.isArray(value)) {
      return value.map(String).filter(Boolean);
    }
  }

  return [];
}

async function extractHeadersFromFile(file: File): Promise<string[]> {
  const extension = getFileExtension(file.name);

  if (extension === ".csv") {
    const text = await file.slice(0, CSV_HEADER_READ_BYTES).text();

    return new Promise((resolve, reject) => {
      Papa.parse<Record<string, unknown>>(text, {
        header: true,
        skipEmptyLines: true,
        preview: 1,
        complete: (results) => {
          resolve((results.meta.fields ?? []).map(normalizeHeader).filter(Boolean));
        },
        error: reject,
      });
    });
  }

  if (extension === ".xlsx" || extension === ".xls") {
    const sheets = await extractSheetsFromFile(file);
    return sheets[0]?.headers ?? [];
  }

  return [];
}

async function extractSheetsFromFile(file: File): Promise<DetectedSheet[]> {
  const extension = getFileExtension(file.name);

  if (extension === ".csv") {
    const text = await file.slice(0, CSV_HEADER_READ_BYTES).text();

    const headers = await new Promise<string[]>((resolve, reject) => {
      Papa.parse<Record<string, unknown>>(text, {
        header: true,
        skipEmptyLines: true,
        preview: 1,
        complete: (results) => {
          resolve((results.meta.fields ?? []).map(normalizeHeader).filter(Boolean));
        },
        error: reject,
      });
    });

    return [
      {
        name: "CSV Data",
        headers,
      },
    ];
  }

  if (extension === ".xlsx" || extension === ".xls") {
    const buffer = await file.arrayBuffer();

    const workbook = XLSX.read(buffer, {
      type: "array",
      sheetRows: 1,
    });

    return workbook.SheetNames.map((sheetName) => {
      const worksheet = workbook.Sheets[sheetName];

      const rows = XLSX.utils.sheet_to_json<(string | number | null)[]>(worksheet, {
        header: 1,
        defval: "",
        blankrows: false,
      });

      const firstRow = Array.isArray(rows[0]) ? rows[0] : [];

      return {
        name: sheetName,
        headers: firstRow.map(normalizeHeader).filter(Boolean),
      };
    }).filter((sheet) => sheet.headers.length > 0);
  }

  return [];
}

function validateHeadersAgainstTemplate(
  headers: string[],
  template: TemplateValidationConfig,
  fileName: string,
): string | null {
  const extension = getFileExtension(fileName);

  if (template.allowedExtensions.length > 0 && !template.allowedExtensions.includes(extension)) {
    return `${template.label} only supports ${template.allowedExtensions.join(", ")} files.`;
  }

  if (!headers.length) {
    return "Could not read file headers. Please check the file format and try again.";
  }

  if (!template.requiredHeaders.length) {
    return null;
  }

  const headerSet = new Set(headers);

  const missingHeaders = template.requiredHeaders.filter(
    (requiredHeader) => !headerSet.has(normalizeHeader(requiredHeader)),
  );

  if (missingHeaders.length > 0) {
    return `The uploaded file does not match the selected template. Missing required columns: ${missingHeaders.join(", ")}.`;
  }

  return null;
}

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  const [uploadMode, setUploadMode] = useState<UploadMode>("document");
  const [file, setFile] = useState<File | null>(null);
  const [detectedHeaders, setDetectedHeaders] = useState<string[]>([]);
  const [detectedSheets, setDetectedSheets] = useState<DetectedSheet[]>([]);
  const [storageLocation, setStorageLocation] = useState("");
  const [processType, setProcessType] = useState<DocumentProcessType | "">("");
  const [selectedTemplate, setSelectedTemplate] = useState("");
  const [templateValidationMessage, setTemplateValidationMessage] = useState<string | null>(null);
  const [isValidatingTemplate, setIsValidatingTemplate] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [templateCode, setTemplateCode] = useState("");
  const [templateName, setTemplateName] = useState("");
  const [templateDescription, setTemplateDescription] = useState("");

  const [createDocument, { isLoading: isUploadingDocument }] = useCreateDocumentMutation();

  const [createTemplateStructure, { isLoading: isUploadingTemplate }] =
    useCreateTemplateStructureMutation();

  const { data: savedTemplatesResponse, isLoading: isTemplatesLoading } =
    useListActiveTemplatesQuery();

  // eslint-disable-next-line react-hooks/exhaustive-deps
  const savedTemplates = Array.isArray(savedTemplatesResponse) ? savedTemplatesResponse : [];

  const isUploading = isUploadingDocument || isUploadingTemplate;

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

  const fileNeedsProcessing = useMemo(() => (file ? requiresProcessing(file) : false), [file]);

  const isPdf = useMemo(() => (file ? isPdfFile(file) : false), [file]);

  const normalizedTemplateCode = useMemo(() => normalizeTemplateCode(templateCode), [templateCode]);

  const templateCodeExists = useMemo(
    () => savedTemplates.some((template) => template.code === normalizedTemplateCode),
    [savedTemplates, normalizedTemplateCode],
  );

  const selectedTemplateConfig = useMemo<TemplateValidationConfig | null>(() => {
    const template = savedTemplates.find((item) => item.code === selectedTemplate);

    if (!template) return null;

    const allowedExtensions = getStringArrayFromConfig(template.configuration, [
      "allowed_extensions",
      "allowedExtensions",
    ]);

    return {
      label: template.name,
      allowedExtensions,
      requiredHeaders: [],
    };
  }, [savedTemplates, selectedTemplate]);

  const isTemplateInvalid =
    Boolean(selectedTemplateConfig) &&
    Boolean(templateValidationMessage) &&
    !templateValidationMessage?.startsWith("File matches");

  const validateSelectedTemplate = useCallback(
    async (nextFile: File, template: TemplateValidationConfig, headersFromState?: string[]) => {
      if (isPdfFile(nextFile)) {
        return;
      }

      try {
        setIsValidatingTemplate(true);

        const headers =
          headersFromState && headersFromState.length > 0
            ? headersFromState
            : await extractHeadersFromFile(nextFile);

        setDetectedHeaders(headers);

        const validationError = validateHeadersAgainstTemplate(headers, template, nextFile.name);

        setTemplateValidationMessage(
          validationError ? validationError : `File matches the ${template.label}.`,
        );
      } catch {
        setTemplateValidationMessage("Failed to validate file against the selected template.");
      } finally {
        setIsValidatingTemplate(false);
      }
    },
    [],
  );

  const handleUploadModeChange = useCallback((checked: boolean) => {
    setUploadMode(checked ? "template" : "document");
    setError(null);
    setTemplateValidationMessage(null);
    setSelectedTemplate("");
  }, []);

  const processSelectedFiles = useCallback(
    async (addedFiles: File[]) => {
      const selectedFile = addedFiles[0];

      if (!selectedFile) return;

      setTemplateValidationMessage(null);
      setDetectedHeaders([]);
      setDetectedSheets([]);

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

      if (isPdfFile(selectedFile)) {
        return;
      }

      if (uploadMode === "template" || selectedTemplateConfig) {
        try {
          setIsValidatingTemplate(true);

          const sheets = await extractSheetsFromFile(selectedFile);
          const headers = sheets.flatMap((sheet) => sheet.headers);

          setDetectedSheets(sheets);
          setDetectedHeaders(headers);

          if (uploadMode === "template") {
            setTemplateValidationMessage(
              headers.length ? `Detected headers: ${headers.join(", ")}` : "No headers detected.",
            );
          }

          if (selectedTemplateConfig && uploadMode === "document") {
            const validationError = validateHeadersAgainstTemplate(
              headers,
              selectedTemplateConfig,
              selectedFile.name,
            );

            setTemplateValidationMessage(
              validationError
                ? validationError
                : `File matches the ${selectedTemplateConfig.label}.`,
            );
          }
        } catch {
          setTemplateValidationMessage("Failed to read file headers.");
        } finally {
          setIsValidatingTemplate(false);
        }
      }
    },
    [selectedTemplateConfig, uploadMode],
  );

  const handleFileChange = (
    _event: SyntheticEvent<HTMLElement, Event>,
    { addedFiles }: { addedFiles: File[] },
  ): void => {
    void processSelectedFiles(addedFiles);
  };

  const handleTemplateChange = async (templateCodeValue: string) => {
    setSelectedTemplate(templateCodeValue);
    setTemplateValidationMessage(null);
    setError(null);

    const template = savedTemplates.find((item) => item.code === templateCodeValue);

    if (!template) return;

    const config: TemplateValidationConfig = {
      label: template.name,
      requiredHeaders: [],
      allowedExtensions: getStringArrayFromConfig(template.configuration, [
        "allowed_extensions",
        "allowedExtensions",
      ]),
    };

    if (!file || isPdfFile(file)) return;

    await validateSelectedTemplate(file, config, detectedHeaders);
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

    if (uploadMode === "template" && !templateName.trim()) {
      setError("Please enter a template name.");
      return;
    }

    if (uploadMode === "template" && !normalizedTemplateCode) {
      setError("Please enter a template code.");
      return;
    }

    if (uploadMode === "template" && templateCodeExists) {
      setError(
        `Template code "${normalizedTemplateCode}" already exists. Please use another code.`,
      );
      return;
    }

    if (uploadMode === "document" && requiresProcessing(file) && !processType) {
      setError("Please select a process type.");
      return;
    }

    if (uploadMode === "document" && requiresProcessing(file) && processType) {
      const validationError = validateProcessTypeAgainstFile(file, processType);

      if (validationError) {
        setError(validationError);
        return;
      }
    }

    if (uploadMode === "document" && selectedTemplateConfig && !isPdfFile(file)) {
      const headers = detectedHeaders.length ? detectedHeaders : await extractHeadersFromFile(file);
      const templateError = validateHeadersAgainstTemplate(
        headers,
        selectedTemplateConfig,
        file.name,
      );

      if (templateError) {
        setError(templateError);
        return;
      }
    }

    try {
      setError(null);

      if (uploadMode === "template") {
        const sheets = !isPdfFile(file)
          ? detectedSheets.length
            ? detectedSheets
            : await extractSheetsFromFile(file)
          : [];

        const headers = sheets.flatMap((sheet) => sheet.headers);

        const uploadedDocument = await createDocument({
          file,
          storageLocation,
          isTemplate: true,
          ...(requiresProcessing(file) && processType ? { processType } : {}),
        }).unwrap();

        const payload = {
          template: {
            document_id: uploadedDocument.id,
            code: normalizedTemplateCode,
            name: templateName.trim(),
            description: templateDescription.trim(),
            file_type: getFileExtension(file.name).replace(".", ""),
            configuration: {
              source_file_name: file.name,
              allowed_extensions: [getFileExtension(file.name)],
              required_headers: headers,
              process_type: null,
            },
          },
          sheets: sheets.map((sheet, sheetIndex) => ({
            sheet: {
              code: normalizeHeader(sheet.name),
              name: sheet.name,
              display_name: sheet.name,
              required: true,
              sheet_order: sheetIndex + 1,
              header_row: 1,
              start_row: 2,
              allow_extra_columns: true,
              allow_duplicate_headers: false,
              configuration: {},
            },
            columns: sheet.headers.map((header, index) => ({
              column_key: header,
              column_name: header,
              display_name: header,
              data_type: "STRING",
              required: false,
              is_unique: false,
              column_order: index + 1,
              default_value: null,
              allowed_values: [],
              aliases: [],
              configuration: {},
            })),
          })),
        };

        const createdTemplate = await createTemplateStructure(payload).unwrap();

        console.log("[UploadDocumentModal] template structure created", createdTemplate);
      } else {
        await createDocument({
          file,
          storageLocation,
          ...(requiresProcessing(file) && processType ? { processType } : {}),
          ...(selectedTemplate ? { metadata: { template_code: selectedTemplate } } : {}),
        }).unwrap();
      }

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

  const isSubmitDisabled =
    isUploading ||
    isLocationsLoading ||
    isValidatingTemplate ||
    !file ||
    !storageLocation ||
    (uploadMode === "document" && fileNeedsProcessing && !processType) ||
    (uploadMode === "template" &&
      (!templateName.trim() || !normalizedTemplateCode || templateCodeExists)) ||
    !hasStorageLocations ||
    isTemplateInvalid;

  return (
    <div style={{ maxWidth: 720 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Document</h2>

        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload a document for processing or upload a reusable template for validating future
          files.
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

          {uploadMode === "template" && templateCodeExists && (
            <InlineNotification
              kind="warning"
              title="Template code already exists"
              subtitle={`Template code "${normalizedTemplateCode}" is already in use.`}
              lowContrast
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

          <Toggle
            id="upload-mode"
            labelText="Upload as template"
            toggled={uploadMode === "template"}
            onToggle={handleUploadModeChange}
            disabled={isUploading || isValidatingTemplate}
          />

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
                <Tag type={uploadMode === "template" ? "cyan" : "purple"}>
                  {uploadMode === "template" ? "Template only" : "Processing required"}
                </Tag>
              )}
            </div>
          )}

          {uploadMode === "template" && (
            <>
              <TextInput
                id="template-code"
                labelText="Template code"
                placeholder="e.g NMS_STOCK_REPORT"
                value={templateCode}
                onChange={(e) => setTemplateCode(normalizeTemplateCode(e.target.value))}
                invalid={templateCodeExists}
                invalidText={`Template code "${normalizedTemplateCode}" already exists.`}
                disabled={isUploading || isValidatingTemplate}
              />

              <TextInput
                id="template-name"
                labelText="Template name"
                placeholder="e.g HIV Viral Load Template"
                value={templateName}
                onChange={(e) => setTemplateName(e.target.value)}
                disabled={isUploading || isValidatingTemplate}
              />

              <TextArea
                id="template-description"
                labelText="Template description"
                placeholder="Describe what this template is used for"
                value={templateDescription}
                onChange={(e) => setTemplateDescription(e.target.value)}
                disabled={isUploading || isValidatingTemplate}
              />
            </>
          )}

          <Select
            id="standard-template"
            labelText="Saved template"
            value={selectedTemplate}
            onChange={(e) => void handleTemplateChange(e.target.value)}
            disabled={
              !file || isUploading || isValidatingTemplate || isPdf || uploadMode === "template"
            }
          >
            <SelectItem
              value=""
              text={
                isTemplatesLoading
                  ? "Loading saved templates..."
                  : isPdf
                    ? "Templates not applicable for PDF"
                    : "Select saved template (optional)"
              }
            />

            {savedTemplates.map((template) => (
              <SelectItem
                key={template.id}
                value={template.code}
                text={`${template.name} (${template.code})`}
              />
            ))}
          </Select>

          {selectedTemplateConfig && (
            <div>
              <div style={{ marginBottom: "0.5rem", color: "#6f6f6f" }}>
                Required columns:{" "}
                {selectedTemplateConfig.requiredHeaders.length
                  ? selectedTemplateConfig.requiredHeaders.join(", ")
                  : "No required columns configured"}
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

          {uploadMode === "template" && !selectedTemplateConfig && templateValidationMessage && (
            <InlineNotification
              kind={detectedHeaders.length ? "success" : "warning"}
              title={detectedHeaders.length ? "Headers detected" : "Template warning"}
              subtitle={templateValidationMessage}
              lowContrast
              onCloseButtonClick={() => setTemplateValidationMessage(null)}
            />
          )}

          {fileNeedsProcessing && uploadMode === "document" ? (
            <Select
              id="process-type"
              labelText="Process type"
              value={processType}
              onChange={(e) => setProcessType(e.target.value as DocumentProcessType | "")}
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
                title={uploadMode === "template" ? "Template upload" : "PDF upload"}
                subtitle={
                  uploadMode === "template"
                    ? "This file will be stored as a template and will not create a processing job."
                    : "This file will be uploaded and stored without any processing job."
                }
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
                description={
                  isUploading
                    ? uploadMode === "template"
                      ? "Uploading template..."
                      : "Uploading document..."
                    : "Reading file headers..."
                }
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
              {uploadMode === "template" ? "Upload Template" : "Upload Document"}
            </Button>
          </div>
        </Stack>
      </Form>
    </div>
  );
};
