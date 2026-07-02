import { Button, InlineLoading } from "@carbon/react";

import type { EmailOutboxItem } from "../types";

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
    <div
      style={{
        display: "flex",
        gap: 8,
        alignItems: "center",
        justifyContent: "space-between",
        paddingBlockEnd: 12,
      }}
    >
      <strong>{emails.length} selected</strong>

      <div
        style={{
          display: "flex",
          gap: 8,
          alignItems: "center",
          flexWrap: "wrap",
        }}
      >
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
