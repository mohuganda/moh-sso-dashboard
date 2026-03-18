import { useMemo, useState } from "react";
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

type UploadDocumentModalProps = {
  onClose: () => void;
};

const ALLOWED_MIME_TYPES = [
  "text/csv",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
];

const ALLOWED_EXTENSIONS = [".csv", ".xlsx"];

function isValidDocument(file: File) {
  const fileName = file.name.toLowerCase();
  const hasValidType = ALLOWED_MIME_TYPES.includes(file.type);
  const hasValidExtension = ALLOWED_EXTENSIONS.some((ext) => fileName.endsWith(ext));

  return hasValidType || hasValidExtension;
}

function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${Math.round(size / 1024)} KB`;
  return `${(size / (1024 * 1024)).toFixed(2)} MB`;
}

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [storageLocation, setStorageLocation] = useState("");
  const [error, setError] = useState<string | null>(null);

  const [createDocument, { isLoading: isUploading }] = useCreateDocumentMutation();

  const { data: locations = [], isLoading: isLocationsLoading } = useListStorageLocationsQuery();

  const activeLocations = useMemo(
    () => locations.filter((location) => location.is_active),
    [locations],
  );

  const hasStorageLocations = activeLocations.length > 0;

  const handleFileChange = (
    _event: React.DragEvent<HTMLElement>,
    { addedFiles }: { addedFiles: File[] },
  ) => {
    const selectedFile = addedFiles?.[0];
    if (!selectedFile) return;

    if (!isValidDocument(selectedFile)) {
      setFile(null);
      setError("Only CSV or Excel (.xlsx) files are allowed.");
      return;
    }

    setError(null);
    setFile(selectedFile);
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

    try {
      setError(null);

      await createDocument({
        file,
        storageLocation,
      }).unwrap();

      onClose();
    } catch (err: any) {
      setError(err?.data?.message || "Upload failed. Please try again.");
    }
  };

  const isSubmitDisabled =
    isUploading || isLocationsLoading || !file || !storageLocation || !hasStorageLocations;

  return (
    <div style={{ maxWidth: 720 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Document</h2>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload a CSV or Excel file and choose where it should be stored for processing.
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

          <FormGroup legendText="Document file">
            <FileUploaderDropContainer
              labelText="Drag and drop a file here or click to browse"
              accept={[".csv", ".xlsx"]}
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
            </div>
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
