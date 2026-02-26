import { useState } from "react";
import {
  Button,
  FileUploaderDropContainer,
  Form,
  FormGroup,
  Select,
  SelectItem,
  InlineNotification,
  Stack,
} from "@carbon/react";
import { useCreateDocumentMutation } from "../../../../store/api/document.api";

type UploadDocumentModalProps = {
  onClose: () => void;
};

export const UploadDocumentModal: React.FC<UploadDocumentModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [storageLocation, setStorageLocation] = useState("");
  const [error, setError] = useState<string | null>(null);

  const [createDocument, { isLoading }] = useCreateDocumentMutation();

  const handleFileChange = (event: any) => {
    const selected = event?.target?.files?.[0];
    if (!selected) return;

    const allowedTypes = [
      "text/csv",
      "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    ];

    if (!allowedTypes.includes(selected.type)) {
      setError("Only CSV or Excel files are allowed.");
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

          <Select
            id="storage-location"
            labelText="Storage Location"
            value={storageLocation}
            onChange={(e) => setStorageLocation(e.target.value)}
          >
            <SelectItem value="" text="Select location" />
            <SelectItem value="550e8400-e29b-41d4-a716-446655440000" text="Default Storage" />
            <SelectItem value="550e8400-e29b-41d4-a716-446655440111" text="Archive Storage" />
          </Select>

          <Button onClick={handleUpload} disabled={isLoading}>
            {isLoading ? "Uploading..." : "Upload Document"}
          </Button>
        </Stack>
      </Form>
    </div>
  );
};
