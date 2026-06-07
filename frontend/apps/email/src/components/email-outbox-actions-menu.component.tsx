import { OverflowMenu, OverflowMenuItem } from "@carbon/react";

import type { EmailOutboxItem } from "@moh-sso/types";

type EmailOutboxActionsMenuProps = {
  email: EmailOutboxItem;
  isMutating?: boolean;
  onView: () => void;
  onRetry: () => void;
  onDelete: () => void;
};

export function EmailOutboxActionsMenu({
  email,
  isMutating = false,
  onView,
  onRetry,
  onDelete,
}: EmailOutboxActionsMenuProps) {
  const canRetry = email.status === "FAILED" || email.status === "RETRY";

  return (
    <OverflowMenu size="sm" flipped>
      <OverflowMenuItem itemText="View details" onClick={onView} />

      {canRetry && <OverflowMenuItem itemText="Retry" onClick={onRetry} disabled={isMutating} />}

      <OverflowMenuItem
        itemText="Delete"
        onClick={onDelete}
        disabled={isMutating}
        isDelete
        hasDivider
      />
    </OverflowMenu>
  );
}
