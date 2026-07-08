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
  InlineLoading,
} from "@carbon/react";
import { useMemo, useState, type SyntheticEvent } from "react";
import { useCreateImportBatchMutation } from "../api";
import "./surveillance.scss";

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

const ALLOWED_EXTENSIONS = [".csv", ".xlsx", ".xls"];

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
  const [createImportBatch, { isLoading: isUploading }] = useCreateImportBatchMutation();

  const isFormValid = useMemo(() => {
    return Boolean(file && surveillanceFileType);
  }, [file, surveillanceFileType]);

  const processSelectedFiles = async (addedFiles: File[]) => {
    const selectedFile = addedFiles[0];
    if (!selectedFile) return;

    if (!isValidDocument(selectedFile)) {
      setFile(null);
      setError("Only CSV or Excel (.csv, .xlsx, .xls) files are allowed.");
      return;
    }

    setError(null);
    setFile(selectedFile);
  };

  const handleFileChange = (
    _event: SyntheticEvent<HTMLElement, Event>,
    { addedFiles }: { addedFiles: File[] },
  ): void => {
    void processSelectedFiles(addedFiles);
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

      await createImportBatch({
        source_name: file.name,
        file_name: file.name,
        dataset_type: surveillanceFileType,
        status: "pending",
      }).unwrap();

      onClose();
    } catch (err: unknown) {
      const message: string =
        typeof err === "object" &&
        err !== null &&
        "data" in err &&
        typeof (err as { data?: { message?: unknown } }).data?.message === "string"
          ? ((err as { data?: { message?: string } }).data?.message ?? "Upload failed. Please try again.")
          : "Upload failed. Please try again.";

      setError(message);
    }
  };

  return (
    <div className="surveillance-upload">
      <div className="surveillance-upload__header">
        <h2>Upload Surveillance File</h2>
        <p>
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
              accept={[".csv", ".xlsx", ".xls"]}
              multiple={false}
              onAddFiles={handleFileChange}
              disabled={isUploading}
            />
          </FormGroup>

          {file && (
            <div className="surveillance-upload__selected-file">
              <Tag type="blue">Selected file</Tag>
              <span>{file.name}</span>
              <span className="surveillance-upload__file-size">{formatFileSize(file.size)}</span>
              <Button kind="ghost" size="sm" onClick={handleRemoveFile} disabled={isUploading}>
                Remove
              </Button>
            </div>
          )}

          <Select
            id="surveillance-file-type"
            labelText="Surveillance file type"
            value={surveillanceFileType}
            onChange={(e) => setSurveillanceFileType(e.target.value)}
            disabled={isUploading}
          >
            <SelectItem value="" text="Select a file type" />
            {SURVEILLANCE_FILE_TYPES.map((option) => (
              <SelectItem key={option.value} value={option.value} text={option.label} />
            ))}
          </Select>

          <div className="surveillance-upload__actions">
            {isUploading && <InlineLoading description="Uploading file..." />}

            <Button kind="secondary" onClick={onClose} disabled={isUploading}>
              Cancel
            </Button>

            <Button onClick={handleUpload} disabled={!isFormValid || isUploading}>
              Upload File
            </Button>
          </div>
        </Stack>
      </Form>
    </div>
  );
};
