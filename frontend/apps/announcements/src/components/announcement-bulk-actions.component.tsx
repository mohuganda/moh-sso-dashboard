import { Button, InlineLoading } from "@carbon/react";

import type { Announcement } from "../types";
import "./announcements.components.scss";

type BulkAction = "publish" | "draft" | "archive" | "pin" | "unpin" | "delete" | null;

type AnnouncementBulkActionsProps = {
  announcements: Announcement[];
  loadingAction: BulkAction;

  onPublish: () => void;
  onMoveToDraft: () => void;
  onArchive: () => void;
  onPin: () => void;
  onUnpin: () => void;
  onDelete: () => void;
};

export function AnnouncementBulkActions({
  announcements,
  loadingAction,
  onPublish,
  onMoveToDraft,
  onArchive,
  onPin,
  onUnpin,
  onDelete,
}: AnnouncementBulkActionsProps) {
  if (announcements.length === 0) return null;

  const isLoading = Boolean(loadingAction);

  return (
    <div className="announcement-bulk-actions">
      <strong>{announcements.length} selected</strong>

      <div className="announcement-bulk-actions__buttons">
        {isLoading && <InlineLoading description="Processing…" />}

        <Button size="sm" kind="secondary" disabled={isLoading} onClick={onPublish}>
          Publish
        </Button>

        <Button size="sm" kind="ghost" disabled={isLoading} onClick={onMoveToDraft}>
          Move to draft
        </Button>

        <Button size="sm" kind="ghost" disabled={isLoading} onClick={onPin}>
          Pin
        </Button>

        <Button size="sm" kind="ghost" disabled={isLoading} onClick={onUnpin}>
          Unpin
        </Button>

        <Button size="sm" kind="ghost" disabled={isLoading} onClick={onArchive}>
          Archive
        </Button>

        <Button size="sm" kind="danger--tertiary" disabled={isLoading} onClick={onDelete}>
          Delete
        </Button>
      </div>
    </div>
  );
}
