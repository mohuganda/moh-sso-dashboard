import { useRef, useState } from "react";
import {
  Button,
  Checkbox,
  InlineLoading,
  Stack,
  TextInput,
} from "@carbon/react";
import { Add, Download, TrashCan } from "@carbon/react/icons";

import {
  buildAnnouncementAttachmentDownloadUrl,
  useDeleteAnnouncementAttachmentMutation,
  useListAnnouncementAttachmentsQuery,
  useUpdateAnnouncementAttachmentMutation,
  useUploadAnnouncementAttachmentMutation,
} from "../api";
import { useToast } from "@moh-sso/ui";
import "./announcements.components.scss";

type AnnouncementAttachmentsProps = {
  announcementId: string;
};

const allowedExtensions = new Set(["pdf", "doc", "docx", "xls", "xlsx", "csv", "png", "jpg", "jpeg"]);
const maxFileBytes = 10 * 1024 * 1024;

function formatFileSize(bytes: number) {
  if (!Number.isFinite(bytes) || bytes <= 0) return "0 B";
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function getExtension(fileName: string) {
  return fileName.split(".").pop()?.toLowerCase() ?? "";
}

export function AnnouncementAttachments({ announcementId }: AnnouncementAttachmentsProps) {
  const toast = useToast();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [includeInEmail, setIncludeInEmail] = useState(true);
  const [sortOrder, setSortOrder] = useState(0);

  const {
    data: attachments = [],
    isFetching,
    refetch,
  } = useListAnnouncementAttachmentsQuery(announcementId);
  const [uploadAttachment, uploadState] = useUploadAnnouncementAttachmentMutation();
  const [updateAttachment] = useUpdateAnnouncementAttachmentMutation();
  const [deleteAttachment, deleteState] = useDeleteAnnouncementAttachmentMutation();

  const handleUpload = async (files: FileList | null) => {
    const file = files?.[0];
    if (!file) return;

    if (file.size > maxFileBytes) {
      toast.error("File too large", "Attachments must be 10 MB or smaller.");
      return;
    }

    if (!allowedExtensions.has(getExtension(file.name))) {
      toast.error("Unsupported file", "Use PDF, Office, CSV, PNG, JPG, or JPEG files.");
      return;
    }

    try {
      await uploadAttachment({
        announcementId,
        file,
        include_in_email: includeInEmail,
        sort_order: sortOrder,
      }).unwrap();
      toast.success("Attachment uploaded", "The file was added to this announcement.");
      setSortOrder((value) => value + 1);
      if (inputRef.current) {
        inputRef.current.value = "";
      }
      refetch();
    } catch {
      toast.error("Upload failed", "Unable to upload the announcement attachment.");
    }
  };

  const handleToggleInclude = async (attachmentId: string, nextValue: boolean) => {
    try {
      await updateAttachment({
        announcementId,
        attachmentId,
        include_in_email: nextValue,
      }).unwrap();
    } catch {
      toast.error("Update failed", "Unable to update attachment email settings.");
    }
  };

  const handleDelete = async (attachmentId: string) => {
    try {
      await deleteAttachment({ announcementId, attachmentId }).unwrap();
      toast.success("Attachment removed", "The attachment was removed from this announcement.");
      refetch();
    } catch {
      toast.error("Delete failed", "Unable to remove attachment.");
    }
  };

  return (
    <section aria-labelledby="announcement-attachments-title">
      <Stack gap={4}>
        <div>
          <h4 id="announcement-attachments-title" className="announcement-attachments__title">
            Attachments
          </h4>
        </div>

        <div className="announcement-attachments__upload-row">
          <TextInput
            id="announcement-attachment-file"
            ref={inputRef}
            type="file"
            labelText="File"
            accept=".pdf,.doc,.docx,.xls,.xlsx,.csv,.png,.jpg,.jpeg"
            onChange={(event) => handleUpload(event.target.files)}
          />
          <Button
            kind="tertiary"
            renderIcon={Add}
            disabled={uploadState.isLoading}
            onClick={() => inputRef.current?.click()}
          >
            Select
          </Button>
        </div>

        <Checkbox
          id="announcement-attachment-include-email"
          labelText="Include new uploads in announcement email"
          checked={includeInEmail}
          onChange={(_, data) => setIncludeInEmail(Boolean(data.checked))}
        />

        {(isFetching || uploadState.isLoading || deleteState.isLoading) && (
          <InlineLoading description="Updating attachments..." />
        )}

        <div className="announcement-attachments__list">
          {attachments.length === 0 ? (
            <p className="announcement-attachments__empty">No attachments uploaded.</p>
          ) : (
            attachments.map((attachment) => (
              <div key={attachment.id} className="announcement-attachments__item">
                <div className="announcement-attachments__file">
                  <strong className="announcement-attachments__file-name">
                    {attachment.original_file_name}
                  </strong>
                  <span className="announcement-attachments__file-meta">
                    {formatFileSize(attachment.file_size)}
                  </span>
                  <Checkbox
                    id={`announcement-attachment-email-${attachment.id}`}
                    labelText="Include in email"
                    checked={attachment.include_in_email}
                    onChange={(_, data) =>
                      handleToggleInclude(attachment.id, Boolean(data.checked))
                    }
                  />
                </div>
                <Button
                  hasIconOnly
                  kind="ghost"
                  iconDescription="Download attachment"
                  renderIcon={Download}
                  onClick={() =>
                    window.open(
                      buildAnnouncementAttachmentDownloadUrl(announcementId, attachment.id),
                      "_blank",
                      "noopener,noreferrer",
                    )
                  }
                />
                <Button
                  hasIconOnly
                  kind="danger--ghost"
                  iconDescription="Remove attachment"
                  renderIcon={TrashCan}
                  onClick={() => handleDelete(attachment.id)}
                />
              </div>
            ))
          )}
        </div>
      </Stack>
    </section>
  );
}
