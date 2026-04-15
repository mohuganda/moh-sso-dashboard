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

type UploadDocumentModalProps = {
  onClose: () => void;
};

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [storageLocation, setStorageLocation] = useState("");
  const [processType, setProcessType] = useState<DocumentProcessType | "">("");
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

  const processSelectedFiles = async (addedFiles: File[]) => {
    const selectedFile = addedFiles[0];
    if (!selectedFile) return;

    if (!isAcceptedFile(selectedFile)) {
      setFile(null);
      setProcessType("");
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

  const isSubmitDisabled =
    isUploading ||
    isLocationsLoading ||
    !file ||
    !storageLocation ||
    (fileNeedsProcessing && !processType) ||
    !hasStorageLocations;

  return (
    <div style={{ maxWidth: 720 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Document</h2>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload a PDF, CSV, or Excel file, select where it should be stored, and choose a process
          type only when the file requires processing.
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
              disabled={isUploading}
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

          {fileNeedsProcessing ? (
            <Select
              id="process-type"
              labelText="Process type"
              value={processType}
              onChange={(e) => handleProcessTypeChange(e.target.value)}
              disabled={isUploading}
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
            disabled={isLocationsLoading || !hasStorageLocations || isUploading}
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
            {isUploading && <InlineLoading description="Uploading document..." />}

            <Button kind="secondary" onClick={onClose} disabled={isUploading}>
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
