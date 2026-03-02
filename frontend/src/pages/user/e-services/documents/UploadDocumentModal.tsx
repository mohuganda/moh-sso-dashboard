import { useState, useMemo } from "react";
import {
  Button,
  FileUploaderDropContainer,
  Form,
  FormGroup,
  Select,
  SelectItem,
  InlineNotification,
  Stack,
  Loading,
  Tag,
} from "@carbon/react";
import {
  useCreateDocumentMutation,
  useListStorageLocationsQuery,
} from "../../../../store/api/document.api";

type UploadDocumentModalProps = {
  onClose: () => void;
};

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [storageLocation, setStorageLocation] = useState("");
  const [error, setError] = useState<string | null>(null);

  const [createDocument, { isLoading: isUploading }] = useCreateDocumentMutation();

  const { data: locations = [], isLoading: isLocationsLoading } = useListStorageLocationsQuery();

  const activeLocations = useMemo(() => locations.filter((loc) => loc.is_active), [locations]);

  const handleFileChange = (event: any) => {
    const selected = event?.target?.files?.[0];
    if (!selected) return;

    const allowedTypes = [
      "text/csv",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ];

    const allowedExtensions = [".csv", ".xlsx"];

    const hasValidType = allowedTypes.includes(selected.type);
    const hasValidExtension = allowedExtensions.some((ext) =>
      selected.name.toLowerCase().endsWith(ext),
    );

    if (!hasValidType && !hasValidExtension) {
      setError("Only CSV or Excel (.xlsx) files are allowed.");
      return;
    }

    setError(null);
    setFile(selected);
  };

  const handleUpload = async () => {
    if (!file) {
      setError("Please select a file.");
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

  return (
    <div style={{ maxWidth: 700 }}>
      <h2 style={{ marginBottom: "1rem" }}>Upload Document</h2>

      <Form>
        <Stack gap={6}>
          {error && (
            <InlineNotification kind="error" title="Upload Error" subtitle={error} lowContrast />
          )}

          <FormGroup legendText="File">
            <FileUploaderDropContainer
              labelText="Drag and drop file here or click to upload"
              accept={[".csv", ".xlsx"]}
              multiple={false}
              onAddFiles={handleFileChange}
            />
          </FormGroup>

          {isLocationsLoading ? (
            <Loading description="Loading storage locations..." />
          ) : (
            <Select
              id="storage-location"
              labelText="Storage Location"
              value={storageLocation}
              onChange={(e) => setStorageLocation(e.target.value)}
              disabled={activeLocations.length === 0}
            >
              <SelectItem value="" text="Select location" />

              {activeLocations.map((loc) => (
                <SelectItem
                  key={loc.id}
                  value={loc.id}
                  text={`${loc.name} (${loc.provider.toUpperCase()})`}
                />
              ))}
            </Select>
          )}

          {file && <Tag type="blue">Selected: {file.name}</Tag>}

          <Button onClick={handleUpload} disabled={isUploading || isLocationsLoading}>
            {isUploading ? "Uploading..." : "Upload Document"}
          </Button>
        </Stack>
      </Form>
    </div>
  );
};
