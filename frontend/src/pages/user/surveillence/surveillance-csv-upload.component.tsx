import {
  Stack,
  InlineNotification,
  FormGroup,
  FileUploaderDropContainer,
  Tag,
  Select,
  SelectItem,
  Button,
  Form,
} from "@carbon/react";
import { useMemo, useState } from "react";

type UploadCSVModalProps = {
  onClose: () => void;
};

type SurveillanceFileTypeOption = {
  value: string;
  label: string;
};

const ALLOWED_MIME_TYPES = [
  "text/csv",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  "application/vnd.ms-excel",
];

const ALLOWED_EXTENSIONS = [".csv", ".xlsx"];

const SURVEILLANCE_FILE_TYPES: SurveillanceFileTypeOption[] = [
  { value: "facility_weekly_metrics", label: "Facility Weekly Metrics" },
  { value: "district_weekly_status", label: "District Weekly Status" },
  { value: "region_weekly_status", label: "Region Weekly Status" },
  { value: "national_weekly_status", label: "National Weekly Status" },
];

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

export const UploadCSVModal: React.FC<UploadCSVModalProps> = ({ onClose }) => {
  const [file, setFile] = useState<File | null>(null);
  const [surveillanceFileType, setSurveillanceFileType] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  const isFormValid = useMemo(() => {
    return Boolean(file && surveillanceFileType);
  }, [file, surveillanceFileType]);

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

  const handleRemoveFile = () => {
    setFile(null);
    setError(null);
  };

  const handleUpload = async () => {
    if (!file) {
      setError("Please select a file to upload.");
      return;
    }

    if (!surveillanceFileType) {
      setError("Please select a surveillance file type.");
      return;
    }

    try {
      setError(null);
      setIsUploading(true);

      const formData = new FormData();
      formData.append("file", file);
      formData.append("surveillanceFileType", surveillanceFileType);

      // TODO: replace with your actual API mutation/call
      // await uploadSurveillanceFile(formData).unwrap();

      onClose();
    } catch (err: any) {
      setError(err?.data?.message || "Upload failed. Please try again.");
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div style={{ maxWidth: 720 }}>
      <div style={{ marginBottom: "1.5rem" }}>
        <h2 style={{ margin: 0, marginBottom: "0.5rem" }}>Upload Surveillance File</h2>
        <p style={{ margin: 0, color: "#6f6f6f" }}>
          Upload a CSV or Excel file and choose the surveillance dataset it belongs to.
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
              labelText="Drag and drop a CSV or Excel file here, or click to browse"
              accept={[".csv", ".xlsx"]}
              multiple={false}
              onAddFiles={handleFileChange}
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
              <Button kind="ghost" size="sm" onClick={handleRemoveFile}>
                Remove
              </Button>
            </div>
          )}

          <Select
            id="surveillance-file-type"
            labelText="Surveillance file type"
            value={surveillanceFileType}
            onChange={(e) => setSurveillanceFileType(e.target.value)}
          >
            <SelectItem value="" text="Select a file type" />
            {SURVEILLANCE_FILE_TYPES.map((option) => (
              <SelectItem key={option.value} value={option.value} text={option.label} />
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
            <Button kind="secondary" onClick={onClose} disabled={isUploading}>
              Cancel
            </Button>
            <Button onClick={handleUpload} disabled={!isFormValid || isUploading}>
              {isUploading ? "Uploading..." : "Upload File"}
            </Button>
          </div>
        </Stack>
      </Form>
    </div>
  );
};
