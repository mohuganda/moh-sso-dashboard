import {
  Stack,
  FileUploaderDropContainer,
  FileUploaderItem,
  Button,
  Checkbox,
  Tag,
} from "@carbon/react";
import { useState } from "react";

import { FormInlineAlert } from "@moh-sso/ui";

type ImportOptions = {
  enabled: boolean;
  emailVerified: boolean;
  sendResetEmail: boolean;
};

export function ImportUsersPanel() {
  const [file, setFile] = useState<File | null>(null);

  const [options, setOptions] = useState<ImportOptions>({
    enabled: true,
    emailVerified: true,
    sendResetEmail: true,
  });

  const updateOption = (field: keyof ImportOptions, checked: boolean) => {
    setOptions((current) => ({ ...current, [field]: checked }));
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
          onChange={(_, { checked }) => {
            updateOption("enabled", checked);
          }}
        />

        <Checkbox
          id="emailVerified"
          labelText="Mark email as verified"
          checked={options.emailVerified}
          onChange={(_, { checked }) => {
            updateOption("emailVerified", checked);
          }}
        />

        <Checkbox
          id="sendResetEmail"
          labelText="Send password reset email"
          checked={options.sendResetEmail}
          onChange={(_, { checked }) => {
            updateOption("sendResetEmail", checked);
          }}
        />
      </Stack>

      <FormInlineAlert
        title="User import is not connected yet"
        subtitle="The frontend no longer posts to the old hardcoded import URL because the matching backend users import route is not registered."
      />

      {/* -----------------------------
       * Actions
       * ----------------------------- */}
      <Button kind="primary" disabled>
        Start import
      </Button>
    </Stack>
  );
}
