import { OverflowMenu, OverflowMenuItem } from "@carbon/react";

import type { Announcement } from "../types";

type AnnouncementActionsMenuProps = {
  announcement: Announcement;
  onEdit: () => void;
  onPublish: () => void;
  onMoveToDraft: () => void;
  onTogglePin: () => void;
  onArchive: () => void;
  onDelete: () => void;
};

export function AnnouncementActionsMenu({
  announcement,
  onEdit,
  onPublish,
  onMoveToDraft,
  onTogglePin,
  onArchive,
  onDelete,
}: AnnouncementActionsMenuProps) {
  return (
    <OverflowMenu size="sm" flipped>
      <OverflowMenuItem itemText="Edit" onClick={onEdit} />

      {announcement.status !== "PUBLISHED" && (
        <OverflowMenuItem itemText="Publish now" onClick={onPublish} />
      )}

      {announcement.status !== "DRAFT" && (
        <OverflowMenuItem itemText="Move to draft" onClick={onMoveToDraft} />
      )}

      <OverflowMenuItem itemText={announcement.is_pinned ? "Unpin" : "Pin"} onClick={onTogglePin} />

      {announcement.status !== "ARCHIVED" && (
        <OverflowMenuItem itemText="Archive" onClick={onArchive} />
      )}

      <OverflowMenuItem itemText="Delete" onClick={onDelete} isDelete hasDivider />
    </OverflowMenu>
  );
}
