import { Link, Stack, Tag, Tile } from "@carbon/react";
import { Attachment, Link as LinkIcon } from "@carbon/react/icons";

import type {
  Announcement,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
} from "../types";

import { useCreateAnnouncementMutation, useUpdateAnnouncementMutation } from "../api";

import { AnnouncementForm } from "./announcements-form.component";
import { AnnouncementAttachments } from "./announcement-attachments.component";
import { useToast } from "@moh-sso/ui";

type ManageAnnouncementsPanelProps = {
  mode: "create" | "edit";
  announcement?: Announcement | null;
  onSuccess?: () => void;
  onCancel?: () => void;
};

export function ManageAnnouncementsPanel({
  mode,
  announcement,
  onSuccess,
  onCancel,
}: ManageAnnouncementsPanelProps) {
  const toast = useToast();

  const [createAnnouncement, createState] = useCreateAnnouncementMutation();
  const [updateAnnouncement, updateState] = useUpdateAnnouncementMutation();

  const isSubmitting = createState.isLoading || updateState.isLoading;
  const attachmentCount = announcement?.attachment_count ?? announcement?.attachments?.length ?? 0;
  const linkUrl = announcement?.link_url?.trim();
  const linkLabel = announcement?.link_label?.trim() || "Open related link";

  const handleSubmit = async (payload: CreateAnnouncementRequest | UpdateAnnouncementRequest) => {
    try {
      if (mode === "create") {
        await createAnnouncement(payload as CreateAnnouncementRequest).unwrap();

        toast.success({
          title: "Announcement created",
          subtitle: "The announcement was created successfully.",
        });
      } else {
        if (!announcement) return;

        await updateAnnouncement({
          id: announcement.id,
          body: payload as UpdateAnnouncementRequest,
        }).unwrap();

        toast.success({
          title: "Announcement updated",
          subtitle: "The announcement was updated successfully.",
        });
      }

      onSuccess?.();
    } catch {
      toast.error({
        title: mode === "create" ? "Create failed" : "Update failed",
        subtitle:
          mode === "create" ? "Failed to create announcement." : "Failed to update announcement.",
      });
    }
  };

  return (
    <Tile style={{ padding: "1.5rem" }}>
      <Stack gap={6}>
        <div>
          <h3 style={{ margin: 0 }}>
            {mode === "create" ? "Create Announcement" : "Edit Announcement"}
          </h3>

          <p style={{ marginTop: "0.5rem", color: "#6f6f6f" }}>
            {mode === "create"
              ? "Publish announcements, alerts, and important updates for dashboard users."
              : "Update the announcement details, audience, status, schedule, and visibility."}
          </p>
        </div>

        <AnnouncementForm
          mode={mode}
          initialValues={announcement}
          isSubmitting={isSubmitting}
          onSubmit={handleSubmit}
          onCancel={onCancel}
        />

        {mode === "edit" && (linkUrl || attachmentCount > 0) ? (
          <section
            aria-label="Announcement publishing assets"
            style={{
              display: "grid",
              gap: "0.75rem",
              padding: "1rem",
              border: "1px solid var(--cds-border-subtle-01, #e0e0e0)",
              background: "var(--cds-layer, #ffffff)",
            }}
          >
            <h4 style={{ margin: 0 }}>Publishing assets</h4>

            <div style={{ display: "flex", flexWrap: "wrap", gap: "0.5rem" }}>
              {linkUrl && (
                <Tag type="blue" size="sm">
                  <span style={{ display: "inline-flex", alignItems: "center", gap: 4 }}>
                    <LinkIcon size={12} />
                    Related link
                  </span>
                </Tag>
              )}

              {attachmentCount > 0 && (
                <Tag type="cyan" size="sm">
                  <span style={{ display: "inline-flex", alignItems: "center", gap: 4 }}>
                    <Attachment size={12} />
                    {attachmentCount === 1 ? "1 attachment" : `${attachmentCount} attachments`}
                  </span>
                </Tag>
              )}
            </div>

            {linkUrl && (
              <Link href={linkUrl} target="_blank" rel="noopener noreferrer">
                {linkLabel}
              </Link>
            )}
          </section>
        ) : null}

        {mode === "edit" && announcement?.id ? (
          <AnnouncementAttachments announcementId={announcement.id} />
        ) : null}
      </Stack>
    </Tile>
  );
}
