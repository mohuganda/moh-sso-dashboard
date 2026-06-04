import { Tile, Stack } from "@carbon/react";

import type { CreateAnnouncementRequest } from "@/store/types/announcements.types";
import { useCreateAnnouncementMutation } from "@/store/api/announcement.api";
import { AnnouncementForm } from "./announcements-form.component";

export function ManageAnnouncementsPanel() {
  const [createAnnouncement, { isLoading: isSubmitting }] = useCreateAnnouncementMutation();

  const handleSubmit = async (payload: CreateAnnouncementRequest) => {
    await createAnnouncement(payload).unwrap();
  };

  return (
    <Tile style={{ padding: "1.5rem", maxWidth: "880px" }}>
      <Stack gap={6}>
        <div>
          <h3 style={{ margin: 0 }}>Create Announcement</h3>
          <p style={{ marginTop: "0.5rem", color: "#6f6f6f" }}>
            Publish announcements, alerts, and important updates for dashboard users.
          </p>
        </div>

        <AnnouncementForm mode="create" isSubmitting={isSubmitting} onSubmit={handleSubmit} />
      </Stack>
    </Tile>
  );
}
