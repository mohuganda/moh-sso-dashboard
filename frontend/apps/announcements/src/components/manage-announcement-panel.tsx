import { Tile, Stack } from "@carbon/react";

import type {
  Announcement,
  CreateAnnouncementRequest,
  UpdateAnnouncementRequest,
} from "@moh-sso/types";

import { useCreateAnnouncementMutation, useUpdateAnnouncementMutation } from "@moh-sso/api";

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

        {mode === "edit" && announcement?.id ? (
          <AnnouncementAttachments announcementId={announcement.id} />
        ) : null}
      </Stack>
    </Tile>
  );
}
