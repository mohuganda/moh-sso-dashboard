import { Button, InlineLoading } from "@carbon/react";

import type { EmailOutboxItem } from "../types";
import "../pages/email-outbox.scss";

type BulkAction = "retry" | "delete" | null;

type EmailOutboxBulkActionsProps = {
  emails: EmailOutboxItem[];
  loadingAction: BulkAction;

  onRetry: () => void;
  onDelete: () => void;
};

export function EmailOutboxBulkActions({
  emails,
  loadingAction,
  onRetry,
  onDelete,
}: EmailOutboxBulkActionsProps) {
  if (emails.length === 0) return null;

  const isLoading = Boolean(loadingAction);

  return (
    <div className="email-outbox-bulk-actions">
      <strong>{emails.length} selected</strong>

      <div className="email-outbox-bulk-actions__buttons">
        {isLoading && <InlineLoading description="Processing…" />}

        <Button size="sm" kind="secondary" disabled={isLoading} onClick={onRetry}>
          Retry
        </Button>

        <Button size="sm" kind="danger--tertiary" disabled={isLoading} onClick={onDelete}>
          Delete
        </Button>
      </div>
    </div>
  );
}
