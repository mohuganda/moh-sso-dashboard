import {
  Stack,
  FileUploaderDropContainer,
  FileUploaderItem,
  Button,
  InlineLoading,
  Checkbox,
  Tag,
} from "@carbon/react";
import { useState } from "react";

type ImportOptions = {
  enabled: boolean;
  emailVerified: boolean;
  sendResetEmail: boolean;
};

export function ImportUsersPanel() {
  const [file, setFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [options] = useState<ImportOptions>({
    enabled: true,
    emailVerified: true,
    sendResetEmail: true,
  });

  const handleUpload = async () => {
    if (!file) return;

    setUploading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("enabled", String(options.enabled));
      formData.append("emailVerified", String(options.emailVerified));
      formData.append("sendResetEmail", String(options.sendResetEmail));

      await fetch("/api/v1/admin/users/import", {
        method: "POST",
        credentials: "include",
        body: formData,
      });
    } catch (err: any) {
      setError("Failed to start import job");
    } finally {
      setUploading(false);
    }
  };

  return (
    <Stack gap={5}>
      {/* -----------------------------
       * CSV Requirements
       * ----------------------------- */}
      <Stack gap={2}>
        <strong>CSV format</strong>
        <Stack orientation="horizontal" gap={2}>
          <Tag size="sm" type="cool-gray">
            username
          </Tag>
          <Tag size="sm" type="cool-gray">
            email
          </Tag>
          <Tag size="sm" type="cool-gray">
            first_name
          </Tag>
          <Tag size="sm" type="cool-gray">
            last_name
          </Tag>
          <Tag size="sm" type="cool-gray">
            roles
          </Tag>
        </Stack>
      </Stack>

      {/* -----------------------------
       * File Upload
       * ----------------------------- */}
      <FileUploaderDropContainer
        labelText="Drag and drop CSV file here or click to upload"
        accept={[".csv"]}
        multiple={false}
        onAddFiles={(_event, { addedFiles }: { addedFiles: File[] }) => {
          const file = addedFiles?.[0];
          if (file) setFile(file);
        }}
      />

      {file && (
        <FileUploaderItem
          name={file.name}
          status="edit"
          onDelete={() => {
            setFile(null);
          }}
        />
      )}

      {/* -----------------------------
       * Import Options
       * ----------------------------- */}
      <Stack gap={3}>
        <Checkbox
          id="enabled"
          labelText="Enable users after import"
          checked={options.enabled}
          onChange={() => {}}
        />

        <Checkbox
          id="emailVerified"
          labelText="Mark email as verified"
          checked={options.emailVerified}
          onChange={() => {}}
        />

        <Checkbox
          id="sendResetEmail"
          labelText="Send password reset email"
          checked={options.sendResetEmail}
          onChange={() => {}}
        />
      </Stack>

      {/* -----------------------------
       * Error
       * ----------------------------- */}
      {error && <p style={{ color: "var(--cds-text-error)" }}>{error}</p>}

      {/* -----------------------------
       * Actions
       * ----------------------------- */}
      <Button kind="primary" disabled={!file || uploading} onClick={handleUpload}>
        {uploading ? <InlineLoading description="Starting import…" /> : "Start import"}
      </Button>
    </Stack>
  );
}
